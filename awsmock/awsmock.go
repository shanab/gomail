package awsmock

import (
	"gomail/awsmock/mocks"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type mockAwsErr struct {
	code string
	err  string
}

func (e mockAwsErr) Error() string {
	return e.err
}

func (e mockAwsErr) Message() string {
	return e.err
}

func (e mockAwsErr) Code() string {
	return e.code
}

func (e mockAwsErr) OrigErr() error {
	return nil
}

func NewMockAwsErr(code, err string) awserr.Error {
	return &mockAwsErr{
		code: code,
		err:  err,
	}
}

func MockSQSSendEmail(queueUrl, messageBody, messageId string, awsErr awserr.Error) *mocks.SQSAPI {
	mockSQS := new(mocks.SQSAPI)

	if awsErr == nil {
		mockSQS.On("SendMessage", &sqs.SendMessageInput{
			QueueUrl:    &queueUrl,
			MessageBody: &messageBody,
		}).Return(&sqs.SendMessageOutput{MessageId: &messageId}, nil)
	} else {
		mockSQS.On("SendMessage", &sqs.SendMessageInput{
			QueueUrl:    &queueUrl,
			MessageBody: &messageBody,
		}).Return(nil, awsErr)
	}

	return mockSQS
}
