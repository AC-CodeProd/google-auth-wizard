package main

import (
	"fmt"
	"google-auth-wizard/auth"
	"google-auth-wizard/config"
	"google-auth-wizard/googlescopes"
	"google-auth-wizard/logger"
	"google-auth-wizard/storage"
	"google-auth-wizard/terminal"
	"google-auth-wizard/utils"
	"os"
	"sort"

	"github.com/goforj/godump"
	"golang.org/x/oauth2"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	logger.Debug("Starting Google Auth Wizard")

	cfg := config.LoadConfigWithDefaults("config.yaml")
	logger.Debug("Configuration loaded: %+v", cfg)

	filename := utils.ParseFlags()
	logger.Debug("Using credentials file: %s", filename)

	credentials := utils.ReadCredentials(filename)
	googleServices, err := fetchGoogleScopes(cfg)
	if err != nil {
		return fmt.Errorf("failed to fetch Google scopes: %w", err)
	}

	terminal := createTerminal(cfg)
	items := convertToTerminalItems(googleServices)

	logger.Debug("Created %d terminal items for user selection", len(items))
	logger.Info("Starting scope selection interface...")

	selectedScopes, err := terminal.Run("Select Google Scopes OAuth 2.0", items)
	if err != nil {
		return fmt.Errorf("terminal error: %w", err)
	}

	logger.Debug("User selected %d scopes", len(selectedScopes))

	if terminal.HasBeenValidated() {
		printSelectedScopes(selectedScopes)
		if len(selectedScopes) == 0 {
			return fmt.Errorf("no OAuth scopes selected. Please run the application again and select at least one scope to proceed with authentication")
		}

		logger.Info("Creating OAuth configuration...")
		config, err := auth.CreateOAuthConfig(credentials, selectedScopes)
		if err != nil {
			return fmt.Errorf("failed to create OAuth config: %w", err)
		}

		logger.Info("Starting OAuth flow...")

		tokenStorage := storage.NewTokenStorage(storage.GetDefaultTokenPath())
		forceNew := os.Getenv("GOOGLE_AUTH_WIZARD_FORCE_NEW") == "true"

		var token *oauth2.Token
		if !forceNew && tokenStorage.Exists() {
			logger.Debug("Found existing token file, checking validity...")
			if storedToken, err := tokenStorage.Load(); err == nil {
				if storedToken.IsValid() && storedToken.HasScopes(selectedScopes) {
					logger.Info("Using existing valid token")
					token = storedToken.Token
				} else {
					logger.Debug("Stored token is invalid or missing required scopes")
				}
			} else {
				logger.Debug("Failed to load stored token: %v", err)
			}
		} else if forceNew {
			logger.Debug("Force new token requested, ignoring saved tokens")
		}

		if token == nil {
			logger.Info("Obtaining new OAuth token...")
			newToken, err := auth.GetTokenFromLocalServer(cfg, config)
			if err != nil {
				return fmt.Errorf("failed to get OAuth token: %w", err)
			}
			token = newToken

			if err := tokenStorage.Save(token, selectedScopes); err != nil {
				logger.Error("Failed to save token: %v", err)
			} else {
				logger.Info("Token saved to %s", storage.GetDefaultTokenPath())
			}
		}

		logger.Info("OAuth token received successfully!")
		if logger.IsDebug() {
			godump.Dump(token)
		} else {
			logger.Print("Access token: %s\n", token.AccessToken[:10]+"...")
		}
	}
	return nil
}

func fetchGoogleScopes(cfg *config.Config) (*googlescopes.GoogleServices, error) {
	logger.Debug("Fetching Google scopes from %s", cfg.OAuth.OAuthPlaygroundURL)

	client := googlescopes.NewClient(
		googlescopes.WithTimeout(cfg.OAuth.ScopeTimeout),
		googlescopes.WithBaseURL(cfg.OAuth.OAuthPlaygroundURL),
		googlescopes.WithScopeEndpoint(cfg.OAuth.ScopeEndpoint),
	)

	googleServices, err := client.FetchScopes()
	if err != nil {
		return nil, fmt.Errorf("error fetching scopes: %w", err)
	}

	logger.Debug("Fetched %d Google services with %d total scopes",
		googleServices.GetServiceCount(), googleServices.GetTotalScopeCount())

	return googleServices, nil
}

func createTerminal(cfg *config.Config) *terminal.Terminal {
	return terminal.New(
		terminal.WithListHeight(cfg.Terminal.Height),
		terminal.WithTitleStyle(terminal.DefaultTitleStyle().
			Foreground(terminal.Color("#FFFFFF")).
			Bold(true),
		),
		terminal.WithSelectedItemStyle(terminal.DefaultSelectedItemStyle().
			Foreground(terminal.Color("#3498DB")).
			Bold(true),
		),
	)
}

func printSelectedScopes(selectedScopes []string) {
	fmt.Printf("\nSelected scopes (%d):\n", len(selectedScopes))
	for _, scope := range selectedScopes {
		fmt.Printf("- %s\n", scope)
	}
}

func convertToTerminalItems(services *googlescopes.GoogleServices) []terminal.Item {
	var items []terminal.Item

	for serviceName, scopes := range *services {
		if len(scopes) == 0 {
			continue
		}

		children := make([]terminal.Item, len(scopes))
		for i, scope := range scopes {
			children[i] = terminal.Item{
				Title:       scope.URL,
				Description: scope.Description,
				Value:       scope.URL,
				IsHeader:    false,
			}
		}

		items = append(items, terminal.Item{
			Title:       serviceName,
			Description: fmt.Sprintf("%d scopes available", len(scopes)),
			Value:       serviceName,
			IsHeader:    true,
			Children:    children,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Title < items[j].Title
	})

	return items
}
