package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
)

type SESWorker struct {
	*worker
}

func (w *SESWorker) Send(messages []*Message, failures chan<- int) {
	emailStatus := make(chan bool)
	for _, message := range messages {
		go w.send(message, emailStatus)
	}

	var failedEmails int
	for i := 0; i < len(messages); i++ {
		success := <-emailStatus
		if !success {
			failedEmails++
		}
	}
	close(emailStatus)

	failures <- failedEmails
}

func (w *SESWorker) send(message *Message, status chan<- bool) {
	email, err := messageToEmail(message.Message)
	if err != nil {
		log.Print("[ERROR] Could not convert sqs message to email: %v", err.Error())
		deleteFromQueue(message)
		status <- true
		// TODO push message to dead letter queue

		return
	}

	_, err = sesClient.SendEmail(&ses.SendEmailInput{
		Source: aws.String(email.From()),
		Destination: &ses.Destination{
			ToAddresses: []*string{aws.String(email.To())},
		},
		Message: &ses.Message{
			Subject: &ses.Content{Data: &email.Subject},
			Body: &ses.Body{
				Text: &ses.Content{Data: &email.Body},
			},
		},
	})

	if err != nil {
		log.Print("[ERROR] SES: Could not send email: %v", err.Error())
		returnToQueue(message)
		status <- false
	}
}
