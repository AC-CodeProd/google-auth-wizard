# Makefile pour Google Auth Wizard
# Variables
BINARY_NAME=google-auth-wizard
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT?=$(shell git rev-parse HEAD)
BUILD_TIME?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build directories
BUILD_DIR=build
DIST_DIR=dist

# Couleurs pour l'affichage
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: help build clean test coverage lint fmt vet deps update-deps install run dev release docker-build docker-run setup

## help: Affiche cette aide
help:
	@echo "$(BLUE)Google Auth Wizard - Makefile$(NC)"
	@echo ""
	@echo "$(YELLOW)Commandes disponibles:$(NC)"
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ 🔨 Build et installation

## build: Compile le projet
build: clean
	@echo "$(BLUE)🔨 Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(GREEN)✅ Build terminé: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## build-all: Compile pour toutes les plateformes
build-all: clean
	@echo "$(BLUE)🔨 Building for all platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	
	# Linux AMD64
	@echo "$(YELLOW)Building for Linux AMD64...$(NC)"
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	
	# Linux ARM64
	@echo "$(YELLOW)Building for Linux ARM64...$(NC)"
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .
	
	# macOS AMD64
	@echo "$(YELLOW)Building for macOS AMD64...$(NC)"
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	
	# macOS ARM64 (Apple Silicon)
	@echo "$(YELLOW)Building for macOS ARM64...$(NC)"
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .
	
	# Windows AMD64
	@echo "$(YELLOW)Building for Windows AMD64...$(NC)"
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	
	@echo "$(GREEN)✅ Multi-platform build completed in $(DIST_DIR)/$(NC)"

## install: Installe le binaire dans $GOPATH/bin
install: build
	@echo "$(BLUE)📦 Installing $(BINARY_NAME)...$(NC)"
	$(GOCMD) install $(LDFLAGS) .
	@echo "$(GREEN)✅ Installation terminée$(NC)"

##@ 🧪 Tests et qualité

## test: Lance tous les tests
test:
	@echo "$(BLUE)🧪 Running tests...$(NC)"
	$(GOTEST) -v -race ./...
	@echo "$(GREEN)✅ Tests completed$(NC)"

## test-short: Lance les tests courts uniquement
test-short:
	@echo "$(BLUE)🧪 Running short tests...$(NC)"
	$(GOTEST) -short -v ./...

## test-coverage: Lance les tests avec couverture
coverage:
	@echo "$(BLUE)🧪 Running tests with coverage...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOTEST) -v -race -coverprofile=$(BUILD_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "$(GREEN)✅ Coverage report generated: $(BUILD_DIR)/coverage.html$(NC)"

## bench: Lance les benchmarks
bench:
	@echo "$(BLUE)⚡ Running benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...

##@ 🔍 Code Quality

## lint: Lance golangci-lint
lint:
	@echo "$(BLUE)🔍 Running linter...$(NC)"
	@command -v golangci-lint >/dev/null 2>&1 || { echo "$(RED)❌ golangci-lint not installed. Run: make setup$(NC)"; exit 1; }
	golangci-lint run ./...
	@echo "$(GREEN)✅ Linting completed$(NC)"

## fmt: Formate le code
fmt:
	@echo "$(BLUE)🎨 Formatting code...$(NC)"
	$(GOFMT) ./...
	@echo "$(GREEN)✅ Code formatted$(NC)"

## vet: Lance go vet
vet:
	@echo "$(BLUE)🔍 Running go vet...$(NC)"
	$(GOVET) ./...
	@echo "$(GREEN)✅ Vet completed$(NC)"

## check: Lance fmt, vet et lint
check: fmt vet lint
	@echo "$(GREEN)✅ All quality checks passed$(NC)"

##@ 📦 Dépendances

## deps: Télécharge les dépendances
deps:
	@echo "$(BLUE)📦 Downloading dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) verify
	@echo "$(GREEN)✅ Dependencies downloaded$(NC)"

## update-deps: Met à jour les dépendances
update-deps:
	@echo "$(BLUE)🔄 Updating dependencies...$(NC)"
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "$(GREEN)✅ Dependencies updated$(NC)"

## tidy: Nettoie go.mod et go.sum
tidy:
	@echo "$(BLUE)🧹 Tidying modules...$(NC)"
	$(GOMOD) tidy
	@echo "$(GREEN)✅ Modules tidied$(NC)"

##@ 🚀 Exécution et développement

## run: Lance l'application
run: build
	@echo "$(BLUE)🚀 Running $(BINARY_NAME)...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME)

## dev: Lance l'application en mode développement avec logs debug
dev: build
	@echo "$(BLUE)🚀 Running $(BINARY_NAME) in development mode...$(NC)"
	GOOGLE_AUTH_LOG_LEVEL=debug ./$(BUILD_DIR)/$(BINARY_NAME)

## watch: Lance l'application et la recompile automatiquement (nécessite entr)
watch:
	@echo "$(BLUE)👀 Watching for changes...$(NC)"
	@command -v entr >/dev/null 2>&1 || { echo "$(RED)❌ entr not installed. Install with: apt-get install entr$(NC)"; exit 1; }
	find . -name "*.go" | entr -r make run

##@ 🐳 Docker

## docker-build: Construit l'image Docker
docker-build:
	@echo "$(BLUE)🐳 Building Docker image...$(NC)"
	docker build -t $(BINARY_NAME):latest -t $(BINARY_NAME):$(VERSION) .
	@echo "$(GREEN)✅ Docker image built: $(BINARY_NAME):latest$(NC)"

## docker-run: Lance le conteneur Docker
docker-run: docker-build
	@echo "$(BLUE)🐳 Running Docker container...$(NC)"
	docker run --rm -it \
		-v $(PWD)/config.yaml:/app/config.yaml:ro \
		-v $(PWD)/tokens:/app/tokens \
		-p 8080:8080 \
		$(BINARY_NAME):latest

##@ 🧹 Nettoyage

## clean: Nettoie les fichiers de build
clean:
	@echo "$(BLUE)🧹 Cleaning...$(NC)"
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@echo "$(GREEN)✅ Cleaned$(NC)"

## clean-cache: Nettoie le cache de Go
clean-cache:
	@echo "$(BLUE)🧹 Cleaning Go cache...$(NC)"
	$(GOCMD) clean -cache -modcache -testcache
	@echo "$(GREEN)✅ Cache cleaned$(NC)"

##@ 🛠️  Setup et outils

## setup: Installe les outils de développement
setup:
	@echo "$(BLUE)🛠️  Installing development tools...$(NC)"
	
	# golangci-lint
	@echo "$(YELLOW)Installing golangci-lint...$(NC)"
	@command -v golangci-lint >/dev/null 2>&1 || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin latest
	
	# entr pour le watch mode
	@echo "$(YELLOW)Check if entr is available...$(NC)"
	@command -v entr >/dev/null 2>&1 || echo "$(YELLOW)⚠️  Install entr for watch mode: sudo apt-get install entr$(NC)"
	
	@echo "$(GREEN)✅ Development tools setup completed$(NC)"

## version: Affiche les informations de version
version:
	@echo "$(BLUE)📋 Version Information:$(NC)"
	@echo "  Binary: $(BINARY_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Commit: $(COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Go Version: $(shell $(GOCMD) version)"

##@ 📊 Informations

## info: Affiche les informations du projet
info:
	@echo "$(BLUE)📊 Project Information:$(NC)"
	@echo "  Name: $(BINARY_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Go Version: $(shell $(GOCMD) version)"
	@echo "  Build Dir: $(BUILD_DIR)"
	@echo "  Dist Dir: $(DIST_DIR)"
	@echo ""
	@echo "$(BLUE)📁 Project Structure:$(NC)"
	@find . -name "*.go" -not -path "./vendor/*" | head -10
	@echo ""
	@echo "$(BLUE)📦 Dependencies:$(NC)"
	@$(GOMOD) graph | head -5

# Valeur par défaut
.DEFAULT_GOAL := help