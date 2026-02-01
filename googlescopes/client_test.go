package googlescopes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client.baseURL != "https://developers.google.com/oauthplayground" {
		t.Errorf("Expected default baseURL, got %s", client.baseURL)
	}

	if client.scopeEndpoint != "getScopes" {
		t.Errorf("Expected default scopeEndpoint, got %s", client.scopeEndpoint)
	}

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", client.httpClient.Timeout)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	client := NewClient(
		WithTimeout(60*time.Second),
		WithBaseURL("https://custom.example.com"),
		WithScopeEndpoint("customScopes"),
	)

	if client.baseURL != "https://custom.example.com" {
		t.Errorf("Expected custom baseURL, got %s", client.baseURL)
	}

	if client.scopeEndpoint != "customScopes" {
		t.Errorf("Expected custom scopeEndpoint, got %s", client.scopeEndpoint)
	}

	if client.httpClient.Timeout != 60*time.Second {
		t.Errorf("Expected custom timeout 60s, got %v", client.httpClient.Timeout)
	}
}

func TestFetchScopes_Success(t *testing.T) {
	mockResponse := getScopesResponse{
		Success: true,
		Apis: map[string]apiInfoResponse{
			"Gmail API": {
				IconURL: "https://example.com/gmail.png",
				Scopes: []map[string]scopeResponse{
					{"https://www.googleapis.com/auth/gmail.readonly": {Description: "Read email messages"}},
					{"https://www.googleapis.com/auth/gmail.send": {Description: "Send email messages"}},
				},
			},
			"Drive API": {
				IconURL: "https://example.com/drive.png",
				Scopes: []map[string]scopeResponse{
					{"https://www.googleapis.com/auth/drive": {Description: "Full access to Drive"}},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/getScopes" {
			t.Errorf("Expected path /getScopes, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithScopeEndpoint("getScopes"),
	)

	services, err := client.FetchScopes()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if services.GetServiceCount() != 2 {
		t.Errorf("Expected 2 services, got %d", services.GetServiceCount())
	}

	if !services.HasService("Gmail API") {
		t.Error("Expected Gmail API service to exist")
	}

	if !services.HasService("Drive API") {
		t.Error("Expected Drive API service to exist")
	}

	gmailScopes, exists := services.GetScopesForService("Gmail API")
	if !exists {
		t.Error("Expected Gmail API scopes to exist")
	}

	if len(gmailScopes) != 2 {
		t.Errorf("Expected 2 Gmail scopes, got %d", len(gmailScopes))
	}
}

func TestFetchScopes_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	_, err := client.FetchScopes()
	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestFetchScopes_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	_, err := client.FetchScopes()
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestGoogleServices_Methods(t *testing.T) {
	services := GoogleServices{
		"Gmail API": []Scope{
			{URL: "https://www.googleapis.com/auth/gmail.readonly", Description: "Read email"},
			{URL: "https://www.googleapis.com/auth/gmail.send", Description: "Send email"},
		},
		"Drive API": []Scope{
			{URL: "https://www.googleapis.com/auth/drive", Description: "Drive access"},
		},
	}

	if services.GetServiceCount() != 2 {
		t.Errorf("Expected 2 services, got %d", services.GetServiceCount())
	}

	if services.GetTotalScopeCount() != 3 {
		t.Errorf("Expected 3 total scopes, got %d", services.GetTotalScopeCount())
	}

	if !services.HasService("Gmail API") {
		t.Error("Expected Gmail API to exist")
	}

	if services.HasService("Nonexistent API") {
		t.Error("Expected Nonexistent API to not exist")
	}

	if services.IsEmpty() {
		t.Error("Expected services to not be empty")
	}

	emptyServices := GoogleServices{}
	if !emptyServices.IsEmpty() {
		t.Error("Expected empty services to be empty")
	}

	scope, serviceName, found := services.GetScopeByURL("https://www.googleapis.com/auth/gmail.readonly")
	if !found {
		t.Error("Expected to find scope by URL")
	}

	if serviceName != "Gmail API" {
		t.Errorf("Expected service name 'Gmail API', got '%s'", serviceName)
	}

	if scope.Description != "Read email" {
		t.Errorf("Expected description 'Read email', got '%s'", scope.Description)
	}
}

func TestFetchScopesWithContext_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte(`{"success": true, "apis": {}}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.FetchScopesWithContext(ctx)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}
