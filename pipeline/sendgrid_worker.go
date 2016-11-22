package main

import (
	"log"
	"net/http"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

const (
	sendgridEndpoint = "/v3/mail/send"
	sendgridUrl      = "https://api.sendgrid.com"
	sendgridMethod   = "POST"
)

type SendgridWorker struct {
	*worker
}

func (w *SendgridWorker) Send(messages []*Message, failures chan<- int) {
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

func (w *SendgridWorker) send(message *Message, status chan<- bool) {
	email, err := messageToEmail(message.Message)
	if err != nil {
		log.Print("[ERROR] Could not convert sqs message to email: ", err.Error())
		deleteFromQueue(message)
		status <- true
		// TODO push message to dead letter queue

		return
	}

	from := mail.NewEmail(email.FromName, email.FromEmail)
	to := mail.NewEmail(email.ToName, email.ToEmail)
	content := mail.NewContent("text/plain", email.Body)
	m := mail.NewV3MailInit(from, email.Subject, to, content)

	request := sendgrid.GetRequest(config.SendgridApiKey, sendgridEndpoint, sendgridUrl)
	request.Method = sendgridMethod
	request.Body = mail.GetRequestBody(m)
	resp, err := sendgrid.API(request)
	if err != nil || resp.StatusCode != http.StatusAccepted {
		log.Printf(
			"[ERROR] Sendgrid: Could not send email: status code=%v, body=%v, headers=%v, err=%v",
			resp.StatusCode,
			resp.Body,
			resp.Headers,
			err,
		)
		returnToQueue(message)
		status <- false
		return
	}

	deleteFromQueue(message)
	status <- true
}
