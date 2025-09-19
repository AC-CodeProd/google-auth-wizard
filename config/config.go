package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		DefaultPort   int           `yaml:"defaultPort"`
		MaxPortTries  int           `yaml:"maxPortTries"`
		ServerTimeout time.Duration `yaml:"serverTimeout"`
	} `yaml:"server"`

	OAuth struct {
		CallbackPath       string        `yaml:"callbackPath"`
		OAuthPlaygroundURL string        `yaml:"oauthPlaygroundURL"`
		ScopeEndpoint      string        `yaml:"scopeEndpoint"`
		ScopeTimeout       time.Duration `yaml:"scopeTimeout"`
	} `yaml:"oauth"`

	Terminal struct {
		Height int `yaml:"height"`
	} `yaml:"terminal"`
}

var GlobalConfig *Config

func LoadConfig(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("error parsing config file: %v", err)
	}

	GlobalConfig = config
	return nil
}

func LoadConfigWithDefaults(filename string) *Config {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("Config file '%s' not found, creating default configuration...\n", filename)
		if err := CreateDefaultConfigFile(filename); err != nil {
			log.Printf("Warning: Could not create config file (%v), using in-memory defaults\n", err)
			config := GetDefaultConfig()
			return applyEnvironmentOverrides(config)
		}
		log.Printf("Default config file created at '%s'\n", filename)
	}

	if err := LoadConfig(filename); err != nil {
		log.Printf("Warning: Could not load config file (%v), using defaults\n", err)
		config := GetDefaultConfig()
		return applyEnvironmentOverrides(config)
	}
	return applyEnvironmentOverrides(GlobalConfig)
}

func GetDefaultConfig() *Config {
	return &Config{
		Server: struct {
			DefaultPort   int           `yaml:"defaultPort"`
			MaxPortTries  int           `yaml:"maxPortTries"`
			ServerTimeout time.Duration `yaml:"serverTimeout"`
		}{
			DefaultPort:   8080,
			MaxPortTries:  10,
			ServerTimeout: 5 * time.Minute,
		},
		OAuth: struct {
			CallbackPath       string        `yaml:"callbackPath"`
			OAuthPlaygroundURL string        `yaml:"oauthPlaygroundURL"`
			ScopeEndpoint      string        `yaml:"scopeEndpoint"`
			ScopeTimeout       time.Duration `yaml:"scopeTimeout"`
		}{
			CallbackPath:       "/callback",
			OAuthPlaygroundURL: "https://developers.google.com/oauthplayground",
			ScopeEndpoint:      "getScopes",
			ScopeTimeout:       60 * time.Second,
		},
		Terminal: struct {
			Height int `yaml:"height"`
		}{
			Height: 20,
		},
	}
}

func CreateDefaultConfigFile(filename string) error {
	defaultConfig := GetDefaultConfig()

	_, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("error marshaling default config: %v", err)
	}

	configWithComments := `# Google Auth Wizard Configuration
# This file contains the configuration settings for the Google Auth Wizard application

server:
  # Default port for the OAuth callback server
  defaultPort: 8080
  
  # Maximum number of ports to try if the default port is busy
  maxPortTries: 10
  
  # Timeout for the OAuth callback server (format: 5m, 300s, etc.)
  serverTimeout: 5m0s

oauth:
  # OAuth callback path
  callbackPath: /callback
  
  # Google OAuth playground URL for fetching scopes
  oauthPlaygroundURL: https://developers.google.com/oauthplayground
  
  # Endpoint for fetching scopes
  scopeEndpoint: getScopes
  
  # Timeout for scope fetching requests (format: 60s, 1m, etc.)
  scopeTimeout: 1m0s

terminal:
  # Terminal interface height (number of items to display)
  height: 20
`

	err = os.WriteFile(filename, []byte(configWithComments), 0644)
	if err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}

func ConfigExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func ValidateConfig(config *Config) error {
	if config.Server.DefaultPort <= 0 || config.Server.DefaultPort > 65535 {
		return fmt.Errorf("invalid default port: %d (must be between 1 and 65535)", config.Server.DefaultPort)
	}

	if config.Server.MaxPortTries <= 0 {
		return fmt.Errorf("invalid maxPortTries: %d (must be greater than 0)", config.Server.MaxPortTries)
	}

	if config.Server.ServerTimeout <= 0 {
		return fmt.Errorf("invalid serverTimeout: %v (must be greater than 0)", config.Server.ServerTimeout)
	}

	if config.OAuth.ScopeTimeout <= 0 {
		return fmt.Errorf("invalid scopeTimeout: %v (must be greater than 0)", config.OAuth.ScopeTimeout)
	}

	if config.OAuth.CallbackPath == "" {
		return fmt.Errorf("callbackPath cannot be empty")
	}

	if config.OAuth.OAuthPlaygroundURL == "" {
		return fmt.Errorf("oauthPlaygroundURL cannot be empty")
	}

	if config.OAuth.ScopeEndpoint == "" {
		return fmt.Errorf("scopeEndpoint cannot be empty")
	}

	return nil
}

func LoadConfigWithValidation(filename string) (*Config, error) {
	config := LoadConfigWithDefaults(filename)

	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %v", err)
	}

	return config, nil
}

func applyEnvironmentOverrides(config *Config) *Config {
	if port := os.Getenv("GOOGLE_AUTH_WIZARD_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil && p > 0 && p <= 65535 {
			config.Server.DefaultPort = p
		}
	}

	if maxTries := os.Getenv("GOOGLE_AUTH_WIZARD_MAX_PORT_TRIES"); maxTries != "" {
		if m, err := strconv.Atoi(maxTries); err == nil && m > 0 {
			config.Server.MaxPortTries = m
		}
	}

	if timeout := os.Getenv("GOOGLE_AUTH_WIZARD_SERVER_TIMEOUT"); timeout != "" {
		if t, err := time.ParseDuration(timeout); err == nil {
			config.Server.ServerTimeout = t
		}
	}

	if callbackPath := os.Getenv("GOOGLE_AUTH_WIZARD_CALLBACK_PATH"); callbackPath != "" {
		if strings.HasPrefix(callbackPath, "/") {
			config.OAuth.CallbackPath = callbackPath
		}
	}

	if playgroundURL := os.Getenv("GOOGLE_AUTH_WIZARD_PLAYGROUND_URL"); playgroundURL != "" {
		if strings.HasPrefix(playgroundURL, "http") {
			config.OAuth.OAuthPlaygroundURL = playgroundURL
		}
	}

	if scopeEndpoint := os.Getenv("GOOGLE_AUTH_WIZARD_SCOPE_ENDPOINT"); scopeEndpoint != "" {
		config.OAuth.ScopeEndpoint = scopeEndpoint
	}

	if scopeTimeout := os.Getenv("GOOGLE_AUTH_WIZARD_SCOPE_TIMEOUT"); scopeTimeout != "" {
		if t, err := time.ParseDuration(scopeTimeout); err == nil {
			config.OAuth.ScopeTimeout = t
		}
	}

	if height := os.Getenv("GOOGLE_AUTH_WIZARD_TERMINAL_HEIGHT"); height != "" {
		if h, err := strconv.Atoi(height); err == nil && h > 0 {
			config.Terminal.Height = h
		}
	}

	return config
}
