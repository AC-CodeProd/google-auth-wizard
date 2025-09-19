package auth

import (
	"google-auth-wizard/config"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestCreateOAuthConfig_Success(t *testing.T) {
	credentials := []byte(`{
		"installed": {
			"client_id": "test-client-id.apps.googleusercontent.com",
			"project_id": "test-project",
			"auth_uri": "https://accounts.google.com/o/oauth2/auth",
			"token_uri": "https://oauth2.googleapis.com/token",
			"client_secret": "test-client-secret",
			"redirect_uris": ["urn:ietf:wg:oauth:2.0:oob","http://localhost"]
		}
	}`)

	scopes := []string{"https://www.googleapis.com/auth/gmail.readonly"}

	config, err := CreateOAuthConfig(credentials, scopes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if config.ClientID != "test-client-id.apps.googleusercontent.com" {
		t.Errorf("Expected client ID to be set correctly, got %s", config.ClientID)
	}

	if len(config.Scopes) != 1 || config.Scopes[0] != "https://www.googleapis.com/auth/gmail.readonly" {
		t.Errorf("Expected scopes to be set correctly, got %v", config.Scopes)
	}
}

func TestCreateOAuthConfig_InvalidJSON(t *testing.T) {
	invalidCredentials := []byte(`invalid json`)
	scopes := []string{"https://www.googleapis.com/auth/gmail.readonly"}

	_, err := CreateOAuthConfig(invalidCredentials, scopes)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestGetTokenFromLocalServer_Success(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			DefaultPort   int           `yaml:"defaultPort"`
			MaxPortTries  int           `yaml:"maxPortTries"`
			ServerTimeout time.Duration `yaml:"serverTimeout"`
		}{
			DefaultPort:   8080,
			MaxPortTries:  10,
			ServerTimeout: 5 * time.Second, // Court pour les tests
		},
		OAuth: struct {
			CallbackPath       string        `yaml:"callbackPath"`
			OAuthPlaygroundURL string        `yaml:"oauthPlaygroundURL"`
			ScopeEndpoint      string        `yaml:"scopeEndpoint"`
			ScopeTimeout       time.Duration `yaml:"scopeTimeout"`
		}{
			CallbackPath: "/callback",
		},
	}

	oauthConfig := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
		Scopes: []string{"https://www.googleapis.com/auth/gmail.readonly"},
	}

	_ = cfg
	_ = oauthConfig
}

func TestCreateCallbackHandler(t *testing.T) {
	// Créer une configuration OAuth de test
	oauthConfig := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	tokenChan := make(chan *oauth2.Token, 1)
	errChan := make(chan error, 1)

	handler := createCallbackHandler(oauthConfig, tokenChan, errChan)

	t.Run("Missing Code", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/callback", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		select {
		case err := <-errChan:
			if !strings.Contains(err.Error(), "missing authorization code") {
				t.Errorf("Expected missing code error, got %v", err)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected error to be sent to errChan")
		}
	})

	t.Run("With Code but Invalid Exchange", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/callback?code=test-code", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}

		select {
		case err := <-errChan:
			if !strings.Contains(err.Error(), "code exchange failed") {
				t.Errorf("Expected code exchange error, got %v", err)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected error to be sent to errChan")
		}
	})
}

func TestConstants(t *testing.T) {
	if DEFAULT_STATE_TOKEN == "" {
		t.Error("DEFAULT_STATE_TOKEN should not be empty")
	}

	if SUCCESS_HTML_TEMPLATE == "" {
		t.Error("SUCCESS_HTML_TEMPLATE should not be empty")
	}

	if !strings.Contains(SUCCESS_HTML_TEMPLATE, "Authorization Successful") {
		t.Error("SUCCESS_HTML_TEMPLATE should contain success message")
	}

	if MISSING_AUTH_CODE_MSG == "" {
		t.Error("MISSING_AUTH_CODE_MSG should not be empty")
	}

	if CODE_EXCHANGE_FAILED_MSG == "" {
		t.Error("CODE_EXCHANGE_FAILED_MSG should not be empty")
	}

	if DEFAULT_SERVER_STARTUP_DELAY <= 0 {
		t.Error("DEFAULT_SERVER_STARTUP_DELAY should be positive")
	}
}
