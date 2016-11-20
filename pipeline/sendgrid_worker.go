package main

import (
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SendgridWorker struct {
	*worker
}

func (w *SendgridWorker) Send(emails []*sqs.Message, failures chan<- int) {
	// TODO
}
