package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	AwsRegion                            string   `yaml:"aws_region"`
	AwsClientTimeoutSeconds              int64    `yaml:"aws_client_timeout_seconds"`
	MinimumIterationDurationMilliseconds int64    `yaml:"minimum_iteration_duration_milliseconds"`
	HealthyThreshold                     int      `yaml:"healthy_threshold"`
	UnhealthyThreshold                   int      `yaml:"unhealthy_threshold"`
	SendgridApiKey                       string   `yaml:"sendgrid_api_key"`
	QueueUrls                            []string `yaml:"queue_urls"`
}

func (c Config) validate() error {
	if c.AwsRegion == "" {
		return fmt.Errorf("aws_region is missing")
	}

	if c.SendgridApiKey == "" {
		return fmt.Errorf("sendgrid_api_key is missing")
	}

	if len(c.QueueUrls) == 0 {
		return fmt.Errorf("queue_urls must contain at least one value")
	}

	return nil
}

func NewConfig(filePath string) (*Config, error) {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err = yaml.Unmarshal(contents, &config); err != nil {
		return nil, err
	}
	if err = config.validate(); err != nil {
		return nil, err
	}

	return &config, nil
}
