package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	maxNumberOfMessagesPerReceive = 10
	maxNumberOfMessagesPerQueue   = 120000
)

type MessageBody struct {
	Email *Email `json:"email"`
}

type Email struct {
	FromEmail string `json:"fromEmail"`
	FromName  string `json:"fromName"`
	ToEmail   string `json:"toEmail"`
	ToName    string `json:"toName"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}

func (e Email) From() string {
	if e.FromName == "" {
		return e.FromEmail
	}
	return fmt.Sprintf("%s <%s>", e.FromName, e.FromEmail)
}

func (e Email) To() string {
	if e.ToName == "" {
		return e.ToEmail
	}
	return fmt.Sprintf("%s <%s>", e.ToName, e.ToEmail)
}

type Message struct {
	Message  *sqs.Message
	QueueUrl string
}

func NewMessage(message *sqs.Message, queueUrl string) *Message {
	return &Message{
		Message:  message,
		QueueUrl: queueUrl,
	}
}

type Worker interface {
	Send(emails []string, failures chan<- int)
}

type worker struct {
	isHealthy             bool
	consecHealthyChecks   int
	consecUnhealthyChecks int
}

func (w *worker) UpdateHealthStatus(workerName string, failures int) {
	if failures > 0 {
		if w.isHealthy {
			w.consecUnhealthyChecks++
		} else {
			w.consecHealthyChecks = 0
		}
	} else {
		if w.isHealthy {
			w.consecUnhealthyChecks = 0
		} else {
			w.consecHealthyChecks++
		}
	}

	if w.isHealthy && w.consecUnhealthyChecks > config.UnhealthyThreshold {
		log.Printf("[INFO] Setting worker %s health status to UNHEALTHY", workerName)
		w.isHealthy = false
	}
	if !w.isHealthy && w.consecHealthyChecks > config.HealthyThreshold {
		log.Printf("[INFO] Setting worker %s health status to HEALTHY", workerName)
		w.isHealthy = true
	}
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		SendgridWorker: &SendgridWorker{worker: &worker{isHealthy: true}},
		SESWorker:      &SESWorker{worker: &worker{isHealthy: true}},
	}
}

type Pipeline struct {
	SendgridWorker *SendgridWorker
	SESWorker      *SESWorker
}

func Read() []*Message {
	t := time.Now()
	messages := make([]*Message, 0)
	var wg sync.WaitGroup

	for _, queueUrl := range config.QueueUrls {
		resp, err := sqsClient.GetQueueAttributes(&sqs.GetQueueAttributesInput{
			AttributeNames: []*string{aws.String(sqs.QueueAttributeNameApproximateNumberOfMessages)},
			QueueUrl:       &queueUrl,
		})
		if err != nil {
			log.Print("[ERROR] error retrieving queue attributes: %v", err.Error())
			continue
		}

		messageCountStr := resp.Attributes[sqs.QueueAttributeNameApproximateNumberOfMessages]
		messageCount, err := strconv.ParseInt(*messageCountStr, 10, 64)
		if err != nil {
			log.Print("[ERROR] error converting ApproximateNumberOfMessages to int")
			continue
		}

		// thresholding the maximum number of inflight messages
		if messageCount > maxNumberOfMessagesPerQueue {
			messageCount = maxNumberOfMessagesPerQueue
		}

		messagesC := make(chan *sqs.Message)
		readersCount := (messageCount / maxNumberOfMessagesPerReceive) + 1
		log.Printf("[INFO] Using %d readers to receive messages from queue (%s)", readersCount, queueUrl)
		for i := int64(0); i < readersCount; i++ {
			wg.Add(1)
			go read(queueUrl, messagesC, &wg)
		}

		go func() {
			for message := range messagesC {
				messages = append(messages, NewMessage(message, queueUrl))
			}
		}()
		wg.Wait()
		close(messagesC)
	}

	took := time.Since(t)
	log.Printf("[INFO] Took %v to read %d messages", took, len(messages))
	return messages
}

func read(queueUrl string, messagesC chan<- *sqs.Message, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            &queueUrl,
		MaxNumberOfMessages: aws.Int64(10),
	})
	if err != nil {
		log.Print("[ERROR] error retrieving messages: %v", err.Error())
		return
	}

	for _, message := range resp.Messages {
		messagesC <- message
	}
}

func messageToEmail(message *sqs.Message) (*Email, error) {
	var messageBody MessageBody
	body := []byte(*message.Body)
	if err := json.Unmarshal(body, &messageBody); err != nil {
		log.Print("[ERROR] Could not convert SQS message body to email: ", err.Error())
		return nil, err
	}

	return messageBody.Email, nil
}

func (p *Pipeline) Run() error {
	sendgridFailures := make(chan int)
	sesFailures := make(chan int)
	for {
		t := time.Now()
		log.Print("[INFO] Reading messages from queue(s)")
		messages := Read()
		p.run(messages, sendgridFailures, sesFailures)

		minIterDuration := time.Duration(config.MinimumIterationDurationMilliseconds) * time.Millisecond
		took := time.Since(t)
		log.Printf("[INFO] Pipeline iteration took %v to execute", took)
		if took < minIterDuration {
			sleepDuration := minIterDuration - took
			log.Printf("[INFO] Pipeline sleeping for %v", sleepDuration)
			time.Sleep(sleepDuration)
		}
	}

	return fmt.Errorf("Execution stopped unexpectedly")
}

func (p *Pipeline) run(messages []*Message, sendgridFailures, sesFailures chan int) {
	// Break early if no messages read
	if len(messages) == 0 {
		return
	}

	splitPoint := p.calculateSplitPoint(len(messages))
	sendgridMessages := messages[splitPoint:]
	sesMessages := messages[:splitPoint]
	runningWorkers := 0

	if len(sendgridMessages) > 0 {
		log.Printf("[INFO] Sendgrid dispatched with %d messages", len(sendgridMessages))
		runningWorkers++
		go p.SendgridWorker.Send(sendgridMessages, sendgridFailures)
	}
	if len(sesMessages) > 0 {
		log.Printf("[INFO] SES dispatched with %d messages", len(sesMessages))
		runningWorkers++
		go p.SESWorker.Send(sesMessages, sesFailures)
	}

	for i := 0; i < runningWorkers; i++ {
		select {
		case failures := <-sendgridFailures:
			log.Printf("[INFO] Sendgrid finished with %d failures", failures)
			p.SendgridWorker.UpdateHealthStatus("Sendgrid", failures)
		case failures := <-sesFailures:
			log.Printf("[INFO] SES finished with %d failures", failures)
			p.SESWorker.UpdateHealthStatus("SES", failures)
		}
	}
}

func (p *Pipeline) calculateSplitPoint(size int) int {
	// when only 1 message needs to be sent, prioritize a healthy worker - if any
	if size == 1 {
		if p.SendgridWorker.isHealthy {
			return 0
		} else {
			return 1
		}
	}

	switch {
	case p.SendgridWorker.isHealthy && p.SESWorker.isHealthy ||
		!p.SendgridWorker.isHealthy && !p.SESWorker.isHealthy:
		return size / 2
	case p.SendgridWorker.isHealthy: // Sendgrid healthy but SES not healthy
		return 1
	default: // SES healthy but Sendgrid not healthy
		return size - 1
	}
}

func deleteFromQueue(message *Message) error {
	_, err := sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &message.QueueUrl,
		ReceiptHandle: message.Message.ReceiptHandle,
	})
	if err != nil {
		log.Print("[ERROR] Could not delete message from queue: %v", err.Error())
		return err
	}
	return nil
}

func returnToQueue(message *Message) error {
	_, err := sqsClient.ChangeMessageVisibility(&sqs.ChangeMessageVisibilityInput{
		QueueUrl:          &message.QueueUrl,
		ReceiptHandle:     message.Message.ReceiptHandle,
		VisibilityTimeout: aws.Int64(0),
	})
	if err != nil {
		log.Print("[ERROR] Could not set visibility timeout to zero: %v", err.Error())
		return err
	}
	return nil
}
