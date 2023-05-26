package config

import (
	"fmt"

	"github.com/flowshot-io/x/pkg/config"
)

const (
	file = "commander.yaml"
)

type (
	Config struct {
		Services map[string]Service `json:"services"`
		Global   Global             `json:"global"`
	}

	Temporal struct {
		Host string `json:"host"`
	}

	Storage struct {
		ConnectionString string `json:"connectionString"`
	}

	Global struct {
		Temporal Temporal `json:"temporal"`
		Storage  Storage  `json:"storage"`
	}

	// Service contains the service specific config items
	Service struct {
	}
)

// LoadConfig Helper function for loading configuration
func LoadConfig(configDir string) (*Config, error) {
	conf := Config{}
	err := config.Load("", file, &conf)
	if err != nil {
		return nil, fmt.Errorf("config file corrupted: %w", err)
	}

	return &conf, nil
}

// Validate validates this config
func (c *Config) Validate() error {
	return nil
}
