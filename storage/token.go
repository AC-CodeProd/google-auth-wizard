package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
)

type TokenStorage struct {
	filepath string
}

type StoredToken struct {
	Token     *oauth2.Token `json:"token"`
	Scopes    []string      `json:"scopes"`
	SavedAt   time.Time     `json:"saved_at"`
	ExpiresAt time.Time     `json:"expires_at"`
}

func NewTokenStorage(filepath string) *TokenStorage {
	return &TokenStorage{
		filepath: filepath,
	}
}

func GetDefaultTokenPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".google-auth-wizard-token.json"
	}
	return filepath.Join(homeDir, ".google-auth-wizard", "token.json")
}

func (ts *TokenStorage) Save(token *oauth2.Token, scopes []string) error {
	dir := filepath.Dir(ts.filepath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	storedToken := StoredToken{
		Token:     token,
		Scopes:    scopes,
		SavedAt:   time.Now(),
		ExpiresAt: token.Expiry,
	}

	data, err := json.MarshalIndent(storedToken, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	if err := os.WriteFile(ts.filepath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

func (ts *TokenStorage) Load() (*StoredToken, error) {
	data, err := os.ReadFile(ts.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("token file not found: %s", ts.filepath)
		}
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var storedToken StoredToken
	if err := json.Unmarshal(data, &storedToken); err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}

	return &storedToken, nil
}

func (ts *TokenStorage) Exists() bool {
	_, err := os.Stat(ts.filepath)
	return !os.IsNotExist(err)
}

func (ts *TokenStorage) Delete() error {
	if !ts.Exists() {
		return nil
	}

	if err := os.Remove(ts.filepath); err != nil {
		return fmt.Errorf("failed to delete token file: %w", err)
	}

	return nil
}

func (st *StoredToken) IsValid() bool {
	if st.Token == nil {
		return false
	}

	if !st.Token.Expiry.IsZero() && time.Now().Add(5*time.Minute).After(st.Token.Expiry) {
		return false
	}

	return true
}

func (st *StoredToken) HasScopes(requiredScopes []string) bool {
	if len(requiredScopes) == 0 {
		return true
	}

	scopeMap := make(map[string]bool)
	for _, scope := range st.Scopes {
		scopeMap[scope] = true
	}

	for _, required := range requiredScopes {
		if !scopeMap[required] {
			return false
		}
	}

	return true
}

func (st *StoredToken) GetSummary() string {
	if st.Token == nil {
		return "Invalid token"
	}

	status := "Valid"
	if !st.IsValid() {
		status = "Expired"
	}

	return fmt.Sprintf("Token saved: %s | Status: %s | Scopes: %d | Expires: %s",
		st.SavedAt.Format("2006-01-02 15:04:05"),
		status,
		len(st.Scopes),
		st.ExpiresAt.Format("2006-01-02 15:04:05"))
}
