package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	invalidContent = "InvalidMessageContents"
)

var (
	emailRegexp = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
)

type ResponseError struct {
	Errors map[string]string `json:"errors"`
}

func NewResponseError(errors map[string]string) *ResponseError {
	return &ResponseError{errors}
}

func NewBaseResponseError(errorMsg string) *ResponseError {
	return &ResponseError{map[string]string{"base": errorMsg}}
}

type SendEmailRequest struct {
	Email Email `json:"email"`
}

type Email struct {
	FromEmail string `json:"fromEmail"`
	FromName  string `json:"fromName"`
	ToEmail   string `json:"toEmail"`
	ToName    string `json:"toName"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}

type SendEmailResponse struct {
	MessageId string `json:"messageId"`
}

func (e Email) Validate() (bool, *ResponseError) {
	errors := make(map[string]string)

	if valid, errorMsg := validateEmail("From email", e.FromEmail); !valid {
		errors["fromEmail"] = errorMsg
	}
	if valid, errorMsg := validateEmail("To email", e.ToEmail); !valid {
		errors["toEmail"] = errorMsg
	}

	if len(errors) > 0 {
		return false, NewResponseError(errors)
	}
	return true, nil
}

func validateEmail(fieldName, email string) (bool, string) {
	if email == "" {
		return false, fieldName + " is required"
	}
	if !emailRegexp.MatchString(email) {
		return false, fieldName + " is not a valid email"
	}
	return true, ""
}

func respondWithError(w http.ResponseWriter, respErr *ResponseError, httpStatus int) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	errorBytes, err := json.Marshal(respErr)
	if err != nil {
		// This should never happen
		panic("Could not marshal error response: " + err.Error())
	}

	w.Write(errorBytes)
}

func SendEmailHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, config.MaxBodySizeBytes))
	if err != nil {
		respondWithError(w, NewBaseResponseError("Could not read body"), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var request SendEmailRequest
	if err := json.Unmarshal(body, &request); err != nil {
		respondWithError(w, NewBaseResponseError("Could not decode JSON body"), http.StatusBadRequest)
		return
	}
	email := request.Email
	if valid, respErr := email.Validate(); !valid {
		respondWithError(w, respErr, http.StatusUnprocessableEntity)
		return
	}

	// get a random queue url from config
	queueUrl := config.QueueUrls[rand.Intn(len(config.QueueUrls))]

	bodyStr := string(body)
	resp, err := sqsClient.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    &queueUrl,
		MessageBody: &bodyStr,
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == invalidContent {
			respondWithError(
				w,
				NewBaseResponseError("Body contains characters outside the allowed set"),
				http.StatusBadRequest,
			)
			return
		} else {
			respondWithError(w, NewBaseResponseError("Service unavailable"), http.StatusServiceUnavailable)
			return
		}
	}

	response := &SendEmailResponse{MessageId: *resp.MessageId}
	respBytes, err := json.Marshal(response)
	if err != nil {
		// This should never happen
		panic("Could not marshal response: " + err.Error())
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}
