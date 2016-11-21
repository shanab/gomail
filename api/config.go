package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Port                    int      `yaml:"port"`
	MaxBodySizeBytes        int64    `yaml:"max_body_size_bytes"`
	AwsRegion               string   `yaml:"aws_region"`
	AwsClientTimeoutSeconds int64    `yaml:"aws_client_timeout_seconds"`
	AccessLogFilePath       string   `yaml:"access_log_file_path"`
	QueueUrls               []string `yaml:"queue_urls"`
}

func (c Config) validate() error {
	if c.Port <= 0 {
		return fmt.Errorf("port is either missing or invalid")
	}
	if c.MaxBodySizeBytes < 0 {
		return fmt.Errorf("max_body_size_bytes is invalid")
	}
	if c.AwsRegion == "" {
		return fmt.Errorf("aws_region is missing")
	}
	if c.AccessLogFilePath == "" {
		return fmt.Errorf("access_log_file_path is missing")
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
