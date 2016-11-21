package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigSuite struct {
	suite.Suite
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}

func (s *ConfigSuite) TestConfig() {
	testCases := []struct {
		Case          string
		FilePath      string
		ExpectedError error
	}{
		{
			Case:          "Valid config",
			FilePath:      "fixtures/config_valid.yaml",
			ExpectedError: nil,
		},
		{
			Case:          "Missing config file",
			FilePath:      "fixtures/missing_config.yaml",
			ExpectedError: fmt.Errorf("no such file or directory"),
		},
		{
			Case:          "Invalid port",
			FilePath:      "fixtures/config_invalid_port.yaml",
			ExpectedError: fmt.Errorf("port is either missing or invalid"),
		},
		{
			Case:          "Invalid max_body_size_bytes",
			FilePath:      "fixtures/config_invalid_max_body_size.yaml",
			ExpectedError: fmt.Errorf("max_body_size_bytes is invalid"),
		},
		{
			Case:          "Missing access_log_file_path",
			FilePath:      "fixtures/config_invalid_max_body_size.yaml",
			ExpectedError: fmt.Errorf("max_body_size_bytes is invalid"),
		},
		{
			Case:          "Empty queue_urls",
			FilePath:      "fixtures/config_empty_queue_urls.yaml",
			ExpectedError: fmt.Errorf("queue_urls must contain at least one value"),
		},
		{
			Case:          "Missing AWS Region",
			FilePath:      "fixtures/config_missing_aws_region.yaml",
			ExpectedError: fmt.Errorf("aws_region is missing"),
		},
	}

	for _, testCase := range testCases {
		config, err := NewConfig(testCase.FilePath)
		if testCase.ExpectedError != nil && assert.Error(s.T(), err) {
			assert.Nil(s.T(), config)
			assert.Contains(s.T(), err.Error(), testCase.ExpectedError.Error())
		} else {
			assert.NoError(s.T(), err)
			assert.NotNil(s.T(), config)
		}
	}
}
