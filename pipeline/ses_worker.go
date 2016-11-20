package main

import (
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SESWorker struct {
	*worker
}

func (w *SESWorker) Send(emails []*sqs.Message, failures chan<- int) {
	// TODO
}
