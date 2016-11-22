package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gomail/awsmock"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ApiSuite struct {
	suite.Suite
}

func TestApiSuite(t *testing.T) {
	suite.Run(t, new(ApiSuite))
}

func (s *ApiSuite) TestSendEmail() {
	target := "/email/send"
	method := "POST"
	var standardBodySize int64 = 204800
	stdQueueUrl := "https://sqs.us-east-1.amazonaws.com/111111111111/gomail-mails"
	stdMessageId := "123e4567-e89b-12d3-a456-426655440000"
	testCases := []struct {
		Case              string
		Body              string
		ConfigMaxBodySize int64
		AwsErr            awserr.Error

		ExpectedStatusCode int
		ExpectedResponse   string
	}{
		{
			Case:               "Valid request",
			Body:               `{"email":{"fromEmail":"from@example.com","fromName":"From Name","toEmail":"to@example.com","toName":"To Name","subject":"Test subject","body":"Test body"}}`,
			ConfigMaxBodySize:  standardBodySize,
			ExpectedStatusCode: http.StatusOK,
			ExpectedResponse:   fmt.Sprintf(`{"messageId":"%v"}`, stdMessageId),
		},
		{
			Case:               "No body",
			Body:               "",
			ConfigMaxBodySize:  standardBodySize,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedResponse:   `{"errors":{"base":"Could not decode JSON body"}}`,
		},
		{
			Case:               "Invalid body",
			Body:               "this is not json",
			ConfigMaxBodySize:  standardBodySize,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedResponse:   `{"errors":{"base":"Could not decode JSON body"}}`,
		},
		{
			Case:               "Body size too big",
			Body:               `{"email":{"fromEmail":"from@example.com","fromName":"From Name","toName":"To Name","subject":"Test subject","body":"Test body"}}`,
			ConfigMaxBodySize:  10,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedResponse:   `{"errors":{"base":"Could not decode JSON body"}}`,
		},
		{
			Case:               "Invalid character in body",
			Body:               fmt.Sprintf(`{"email":{"fromEmail":"from@example.com","fromName":"From Name","toEmail":"to@example.com","toName":"To Name","subject":"Test subject","body":"Test body %v"}}`, '\uFFFE'),
			ConfigMaxBodySize:  standardBodySize,
			AwsErr:             awsmock.NewMockAwsErr(invalidContent, "The message contains characters outside the allowed set."),
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedResponse:   `{"errors":{"base":"Body contains characters outside the allowed set"}}`,
		},
		{
			Case:               "Service unavailable",
			Body:               `{"email":{"fromEmail":"from@example.com","fromName":"From Name","toEmail":"to@example.com","toName":"To Name","subject":"Test subject","body":"Test body"}}`,
			ConfigMaxBodySize:  standardBodySize,
			AwsErr:             awsmock.NewMockAwsErr("ServiceUnavailable", "The service is currently unavailable"),
			ExpectedStatusCode: http.StatusServiceUnavailable,
			ExpectedResponse:   `{"errors":{"base":"Service unavailable"}}`,
		},
		{
			Case:               "Invalid from email",
			Body:               `{"email":{"fromEmail":"nonValidEmail","fromName":"From Name","toEmail":"to@example.com","toName":"To Name","subject":"Test subject","body":"Test body"}}`,
			ConfigMaxBodySize:  standardBodySize,
			ExpectedStatusCode: http.StatusUnprocessableEntity,
			ExpectedResponse:   `{"errors":{"fromEmail":"From email is not a valid email"}}`,
		},
		{
			Case:               "Missing to email",
			Body:               `{"email":{"fromEmail":"from@example.com","fromName":"From Name","toName":"To Name","subject":"Test subject","body":"Test body"}}`,
			ConfigMaxBodySize:  standardBodySize,
			ExpectedStatusCode: http.StatusUnprocessableEntity,
			ExpectedResponse:   `{"errors":{"toEmail":"To email is required"}}`,
		},
		{
			Case:               "Missing email body",
			Body:               `{"email":{"fromEmail":"from@example.com","fromName":"From Name","toEmail":"to@example.com","toName":"To Name","subject":"Test subject","body":""}}`,
			ConfigMaxBodySize:  standardBodySize,
			ExpectedStatusCode: http.StatusUnprocessableEntity,
			ExpectedResponse:   `{"errors":{"body":"Body is required"}}`,
		},
	}

	for _, testCase := range testCases {
		sqsClient = awsmock.MockSQSSendEmail(stdQueueUrl, testCase.Body, stdMessageId, testCase.AwsErr)
		config = &Config{
			MaxBodySizeBytes: testCase.ConfigMaxBodySize,
			QueueUrls:        []string{stdQueueUrl},
		}
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(method, target, strings.NewReader(testCase.Body))
		SendEmailHandler(recorder, req)

		assert.Equal(s.T(), testCase.ExpectedStatusCode, recorder.Code)
		assert.Equal(s.T(), testCase.ExpectedResponse, recorder.Body.String())
	}
}
