package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfigWithDefaults_FileExists(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	configContent := `
server:
  defaultPort: 9000
  maxPortTries: 5
  serverTimeout: 2m0s

oauth:
  callbackPath: /test-callback
  oauthPlaygroundURL: https://test.example.com
  scopeEndpoint: testScopes
  scopeTimeout: 30s

terminal:
  height: 15
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cfg := LoadConfigWithDefaults(configFile)

	if cfg.Server.DefaultPort != 9000 {
		t.Errorf("Expected DefaultPort 9000, got %d", cfg.Server.DefaultPort)
	}

	if cfg.Server.MaxPortTries != 5 {
		t.Errorf("Expected MaxPortTries 5, got %d", cfg.Server.MaxPortTries)
	}

	if cfg.Server.ServerTimeout != 2*time.Minute {
		t.Errorf("Expected ServerTimeout 2m, got %v", cfg.Server.ServerTimeout)
	}

	if cfg.OAuth.CallbackPath != "/test-callback" {
		t.Errorf("Expected CallbackPath '/test-callback', got %s", cfg.OAuth.CallbackPath)
	}

	if cfg.Terminal.Height != 15 {
		t.Errorf("Expected Terminal Height 15, got %d", cfg.Terminal.Height)
	}
}

func TestLoadConfigWithDefaults_FileNotExists(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "nonexistent_config.yaml")
	cfg := LoadConfigWithDefaults(configFile)

	defaultCfg := GetDefaultConfig()

	if cfg.Server.DefaultPort != defaultCfg.Server.DefaultPort {
		t.Errorf("Expected default DefaultPort %d, got %d", defaultCfg.Server.DefaultPort, cfg.Server.DefaultPort)
	}

	if cfg.OAuth.CallbackPath != defaultCfg.OAuth.CallbackPath {
		t.Errorf("Expected default CallbackPath %s, got %s", defaultCfg.OAuth.CallbackPath, cfg.OAuth.CallbackPath)
	}
}

func TestValidateConfig_Valid(t *testing.T) {
	cfg := GetDefaultConfig()

	err := ValidateConfig(cfg)
	if err != nil {
		t.Errorf("Valid config should not return error, got: %v", err)
	}
}

func TestValidateConfig_InvalidPort(t *testing.T) {
	cfg := GetDefaultConfig()
	cfg.Server.DefaultPort = -1

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("Expected error for invalid port, got nil")
	}

	if !contains(err.Error(), "invalid default port") {
		t.Errorf("Expected error about invalid port, got: %v", err)
	}
}

func TestValidateConfig_InvalidMaxPortTries(t *testing.T) {
	cfg := GetDefaultConfig()
	cfg.Server.MaxPortTries = 0

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("Expected error for invalid maxPortTries, got nil")
	}
}

func TestValidateConfig_EmptyCallbackPath(t *testing.T) {
	cfg := GetDefaultConfig()
	cfg.OAuth.CallbackPath = ""

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("Expected error for empty callback path, got nil")
	}
}

func TestConfigExists(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "existing_config.yaml")

	err := os.WriteFile(configFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if !ConfigExists(configFile) {
		t.Error("ConfigExists should return true for existing file")
	}

	if ConfigExists("nonexistent_file.yaml") {
		t.Error("ConfigExists should return false for nonexistent file")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || s[len(s)-len(substr):] == substr || s[:len(substr)] == substr || containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestApplyEnvironmentOverrides(t *testing.T) {
	originalEnvs := map[string]string{
		"GOOGLE_AUTH_WIZARD_PORT":            os.Getenv("GOOGLE_AUTH_WIZARD_PORT"),
		"GOOGLE_AUTH_WIZARD_MAX_PORT_TRIES":  os.Getenv("GOOGLE_AUTH_WIZARD_MAX_PORT_TRIES"),
		"GOOGLE_AUTH_WIZARD_SERVER_TIMEOUT":  os.Getenv("GOOGLE_AUTH_WIZARD_SERVER_TIMEOUT"),
		"GOOGLE_AUTH_WIZARD_CALLBACK_PATH":   os.Getenv("GOOGLE_AUTH_WIZARD_CALLBACK_PATH"),
		"GOOGLE_AUTH_WIZARD_PLAYGROUND_URL":  os.Getenv("GOOGLE_AUTH_WIZARD_PLAYGROUND_URL"),
		"GOOGLE_AUTH_WIZARD_SCOPE_ENDPOINT":  os.Getenv("GOOGLE_AUTH_WIZARD_SCOPE_ENDPOINT"),
		"GOOGLE_AUTH_WIZARD_SCOPE_TIMEOUT":   os.Getenv("GOOGLE_AUTH_WIZARD_SCOPE_TIMEOUT"),
		"GOOGLE_AUTH_WIZARD_TERMINAL_HEIGHT": os.Getenv("GOOGLE_AUTH_WIZARD_TERMINAL_HEIGHT"),
	}

	defer func() {
		for key, value := range originalEnvs {
			if value == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, value)
			}
		}
	}()

	_ = os.Setenv("GOOGLE_AUTH_WIZARD_PORT", "9090")
	_ = os.Setenv("GOOGLE_AUTH_WIZARD_MAX_PORT_TRIES", "15")
	_ = os.Setenv("GOOGLE_AUTH_WIZARD_SERVER_TIMEOUT", "10m")
	_ = os.Setenv("GOOGLE_AUTH_WIZARD_CALLBACK_PATH", "/custom-callback")
	_ = os.Setenv("GOOGLE_AUTH_WIZARD_PLAYGROUND_URL", "https://custom.example.com")
	_ = os.Setenv("GOOGLE_AUTH_WIZARD_SCOPE_ENDPOINT", "customScopes")
	os.Setenv("GOOGLE_AUTH_WIZARD_SCOPE_TIMEOUT", "2m")
	os.Setenv("GOOGLE_AUTH_WIZARD_TERMINAL_HEIGHT", "25")

	config := GetDefaultConfig()

	config = applyEnvironmentOverrides(config)

	if config.Server.DefaultPort != 9090 {
		t.Errorf("Expected DefaultPort 9090, got %d", config.Server.DefaultPort)
	}

	if config.Server.MaxPortTries != 15 {
		t.Errorf("Expected MaxPortTries 15, got %d", config.Server.MaxPortTries)
	}

	if config.Server.ServerTimeout != 10*time.Minute {
		t.Errorf("Expected ServerTimeout 10m, got %v", config.Server.ServerTimeout)
	}

	if config.OAuth.CallbackPath != "/custom-callback" {
		t.Errorf("Expected CallbackPath '/custom-callback', got %s", config.OAuth.CallbackPath)
	}

	if config.OAuth.OAuthPlaygroundURL != "https://custom.example.com" {
		t.Errorf("Expected OAuthPlaygroundURL 'https://custom.example.com', got %s", config.OAuth.OAuthPlaygroundURL)
	}

	if config.OAuth.ScopeEndpoint != "customScopes" {
		t.Errorf("Expected ScopeEndpoint 'customScopes', got %s", config.OAuth.ScopeEndpoint)
	}

	if config.OAuth.ScopeTimeout != 2*time.Minute {
		t.Errorf("Expected ScopeTimeout 2m, got %v", config.OAuth.ScopeTimeout)
	}

	if config.Terminal.Height != 25 {
		t.Errorf("Expected Terminal Height 25, got %d", config.Terminal.Height)
	}
}

func TestApplyEnvironmentOverrides_InvalidValues(t *testing.T) {
	originalPort := os.Getenv("GOOGLE_AUTH_WIZARD_PORT")
	defer func() {
		if originalPort == "" {
			_ = os.Unsetenv("GOOGLE_AUTH_WIZARD_PORT")
		} else {
			_ = os.Setenv("GOOGLE_AUTH_WIZARD_PORT", originalPort)
		}
	}()

	os.Setenv("GOOGLE_AUTH_WIZARD_PORT", "invalid")

	config := GetDefaultConfig()
	originalPortValue := config.Server.DefaultPort

	config = applyEnvironmentOverrides(config)

	if config.Server.DefaultPort != originalPortValue {
		t.Errorf("Expected port to remain unchanged for invalid value, got %d", config.Server.DefaultPort)
	}
}
