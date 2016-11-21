package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
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

func main() {
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

	router := mux.NewRouter()

	// enable access logging
	f, err := os.OpenFile(config.AccessLogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fatal("Could not open access log file for writing:", err)
	}
	defer f.Close()
	loggedRouter := handlers.LoggingHandler(f, router)

	router.HandleFunc("/email/send", SendEmailHandler).Methods("POST")

	l, err := net.Listen("tcp", ":"+strconv.Itoa(config.Port))
	if err != nil {
		fatal("Could not listen to port:", err)
	}

	corsAllowedHeaders := []string{"Content-Type"}
	corsAllowedMethods := []string{"POST"}
	corsAllowedOrigins := []string{"*"}
	listenerClosed := make(chan struct{})
	go func() {
		// start serving
		http.Serve(
			l,
			handlers.CORS(
				handlers.AllowedHeaders(corsAllowedHeaders),
				handlers.AllowedMethods(corsAllowedMethods),
				handlers.AllowedOrigins(corsAllowedOrigins),
			)(loggedRouter))
		// done serving, signal listener closed
		listenerClosed <- struct{}{}
	}()

	log.Printf("Server startup complete! Serving requests on port %v", config.Port)

	// setup signal handler and wait for signal
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, syscall.SIGINT)
	<-signalChannel

	// signal received, initiate shutdown
	log.Print("Shutdown signal received, closing listener")

	// close listener
	err = l.Close()
	if err != nil {
		fatal("Could not close listener:", err)
	}

	// wait for listener to close
	<-listenerClosed
	log.Print("Listener closed")
}
