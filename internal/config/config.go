// Package config provides internal configuration management for the application
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// NATSConfig represents NATS-specific configuration options
type NATSConfig struct {
	URL            string `json:"url"`
	Username       string `json:"username,omitempty"`
	Password       string `json:"password,omitempty"`
	Token          string `json:"token,omitempty"`
	AllowReconnect bool   `json:"allowReconnect"`
	MaxReconnect   int    `json:"maxReconnect"`
	ReconnectWait  int    `json:"reconnectWait"` // in seconds
}

// AppConfig represents the application configuration
type AppConfig struct {
	Environment string     `json:"environment"` // dev, test, prod
	LogLevel    string     `json:"logLevel"`
	NATS        NATSConfig `json:"nats"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *AppConfig {
	return &AppConfig{
		Environment: "dev",
		LogLevel:    "info",
		NATS: NATSConfig{
			URL:            "nats://localhost:4222",
			AllowReconnect: true,
			MaxReconnect:   10,
			ReconnectWait:  5,
		},
	}
}

// LoadConfig loads configuration from the specified file path
func LoadConfig(configPath string) (*AppConfig, error) {
	// Start with default config
	config := DefaultConfig()

	// If no config path is provided, return default
	if configPath == "" {
		return config, nil
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the config
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variables overrides
	applyEnvironmentOverrides(config)

	return config, nil
}

// applyEnvironmentOverrides applies configuration overrides from environment variables
func applyEnvironmentOverrides(config *AppConfig) {
	// Override environment if specified
	if env := os.Getenv("APP_ENV"); env != "" {
		config.Environment = env
	}

	// Override log level if specified
	if logLevel := os.Getenv("APP_LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	// Override NATS URL if specified
	if natsURL := os.Getenv("NATS_URL"); natsURL != "" {
		config.NATS.URL = natsURL
	}

	// Override NATS credentials if specified
	if natsUser := os.Getenv("NATS_USER"); natsUser != "" {
		config.NATS.Username = natsUser
	}

	if natsPass := os.Getenv("NATS_PASS"); natsPass != "" {
		config.NATS.Password = natsPass
	}

	if natsToken := os.Getenv("NATS_TOKEN"); natsToken != "" {
		config.NATS.Token = natsToken
	}
}

// SaveConfig saves the configuration to the specified file path
func SaveConfig(config *AppConfig, configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
