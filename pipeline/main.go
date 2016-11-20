package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

var (
	sqsClient sqsiface.SQSAPI

	configFilePath = "config.yaml"
	config         *Config
)

func fatal(v ...interface{}) {
	log.Fatal(v)
	os.Exit(1)
}

func parseFlags() {
	flag.StringVar(&configFilePath, "config", configFilePath, "path to config file (defaults to ./config.yaml)")
	flag.Parse()
}

func init() {
	// parse -config flag (if any)
	parseFlags()

	// read config
	var err error
	config, err = NewConfig(configFilePath)
	if err != nil {
		fatal("Could not initialize config:", err)
	}

	// initialize sqs client
	awsConfig := aws.NewConfig().
		WithHTTPClient(&http.Client{Timeout: time.Duration(config.AwsClientTimeoutSeconds) * time.Second})

	awsSession := session.New(awsConfig)
	sqsClient = sqs.New(awsSession)
}

func main() {
	pipeline := NewPipeline()

	pipeline.Run()
}
