package auth

import (
	"context"
	"fmt"
	"google-auth-wizard/config"
	"google-auth-wizard/logger"
	"google-auth-wizard/utils"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	DEFAULT_STATE_TOKEN = "state-token"

	SUCCESS_HTML_TEMPLATE = `<!DOCTYPE html>
<html>
<head>
    <title>Authorization Successful</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
        .success { color: green; font-size: 24px; margin-bottom: 20px; }
        .info { color: #666; }
    </style>
</head>
<body>
    <div class="success">✅ Authorization Successful!</div>
    <div class="info">You can close this window and return to the terminal.</div>
</body>
</html>`

	MISSING_AUTH_CODE_MSG        = "missing authorization code"
	CODE_EXCHANGE_FAILED_MSG     = "code exchange failed"
	DEFAULT_SERVER_STARTUP_DELAY = 100 * time.Millisecond
)

func CreateOAuthConfig(credentials []byte, selectedScopes []string) (*oauth2.Config, error) {
	config, err := google.ConfigFromJSON(credentials, selectedScopes...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth config: %w", err)
	}
	return config, nil
}

func GetTokenFromLocalServer(cfg *config.Config, config *oauth2.Config) (*oauth2.Token, error) {
	port, err := utils.FindAvailablePort(cfg.Server.DefaultPort, cfg.Server.MaxPortTries)
	if err != nil {
		return nil, fmt.Errorf("error finding available port: %w", err)
	}

	config.RedirectURL = fmt.Sprintf("http://localhost:%d%s", port, cfg.OAuth.CallbackPath)

	logger.Info("Using port %d for OAuth callback", port)
	logger.Debug("Redirect URL: %s", config.RedirectURL)

	tokenChan := make(chan *oauth2.Token, 1)
	errChan := make(chan error, 1)

	serverAddr := fmt.Sprintf(":%d", port)
	srv := &http.Server{Addr: serverAddr}

	http.HandleFunc(cfg.OAuth.CallbackPath, createCallbackHandler(config, tokenChan, errChan))

	go func() {
		logger.Debug("Starting OAuth callback server on port %d...", port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server error: %v", err)
		}
	}()

	time.Sleep(DEFAULT_SERVER_STARTUP_DELAY)

	authURL := config.AuthCodeURL(DEFAULT_STATE_TOKEN, oauth2.AccessTypeOffline)
	logger.Info("Opening browser to: %s", authURL)

	if err := utils.OpenBrowser(authURL); err != nil {
		logger.Info("Unable to open browser automatically. Please open manually: %s", authURL)
	}

	var token *oauth2.Token
	select {
	case token = <-tokenChan:
		logger.Info("Authorization successful!")
	case err := <-errChan:
		return nil, fmt.Errorf("authorization error: %w", err)
	case <-time.After(cfg.Server.ServerTimeout):
		return nil, fmt.Errorf("timeout: authorization not received within %v", cfg.Server.ServerTimeout)
	}

	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Debug("Error shutting down server: %v", err)
	}

	return token, nil
}

func createCallbackHandler(config *oauth2.Config, tokenChan chan<- *oauth2.Token, errChan chan<- error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("%s", MISSING_AUTH_CODE_MSG)
			http.Error(w, "Missing authorization code", http.StatusBadRequest)
			return
		}

		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			errChan <- fmt.Errorf("%s: %v", CODE_EXCHANGE_FAILED_MSG, err)
			http.Error(w, "Code exchange failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, SUCCESS_HTML_TEMPLATE)

		tokenChan <- token
	}
}
