# Makefile for Klipd Clipboard Manager
# macOS clipboard manager built with Wails (Go + React)

# Variables
APP_NAME := klipd
BUILD_DIR := build/bin
FRONTEND_DIR := frontend
GO_FILES := $(shell find . -name "*.go" -not -path "./frontend/*" -not -path "./build/*")
FRONTEND_FILES := $(shell find frontend/src -name "*.tsx" -o -name "*.ts" -o -name "*.css")

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Default target
.PHONY: help
help: ## Show this help message
	@echo "$(BLUE)Klipd Clipboard Manager - Development Commands$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

# Development
.PHONY: dev
dev: ## Start development server with hot reload
	@echo "$(BLUE)Starting Wails development server...$(NC)"
	wails dev

.PHONY: build
build: clean ## Build production app
	@echo "$(BLUE)Building production app...$(NC)"
	wails build

.PHONY: build-debug
build-debug: clean ## Build debug version with console
	@echo "$(BLUE)Building debug version...$(NC)"
	wails build -debug

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build directory...$(NC)"
	rm -rf $(BUILD_DIR)
	rm -rf frontend/dist
	rm -rf frontend/node_modules/.vite

# Testing
.PHONY: test
test: ## Run all Go tests
	@echo "$(BLUE)Running all tests...$(NC)"
	go test ./... -v

.PHONY: test-short
test-short: ## Run tests without verbose output
	@echo "$(BLUE)Running tests (short)...$(NC)"
	go test ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	@echo "$(BLUE)Running tests with verbose output...$(NC)"
	go test ./... -v -count=1

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo "$(BLUE)Running tests with race detector...$(NC)"
	go test ./... -race

.PHONY: test-bench
test-bench: ## Run benchmark tests
	@echo "$(BLUE)Running benchmark tests...$(NC)"
	go test ./... -bench=. -benchmem

.PHONY: test-watch
test-watch: ## Watch files and run tests on changes
	@echo "$(BLUE)Watching for changes and running tests...$(NC)"
	@which fswatch > /dev/null || (echo "$(RED)fswatch not found. Install with: brew install fswatch$(NC)" && exit 1)
	fswatch -o . --exclude='.*' --include='\.go$$' | xargs -n1 -I{} make test-short

# Package-specific tests
.PHONY: test-config
test-config: ## Test config package only
	@echo "$(BLUE)Testing config package...$(NC)"
	go test ./config -v

.PHONY: test-database
test-database: ## Test database package only
	@echo "$(BLUE)Testing database package...$(NC)"
	go test ./database -v

.PHONY: test-models
test-models: ## Test models package only
	@echo "$(BLUE)Testing models package...$(NC)"
	go test ./models -v

.PHONY: test-services
test-services: ## Test services package only
	@echo "$(BLUE)Testing services package...$(NC)"
	go test ./services -v

# Code Quality
.PHONY: lint
lint: ## Run golangci-lint
	@echo "$(BLUE)Running linter...$(NC)"
	@which golangci-lint > /dev/null || (echo "$(RED)golangci-lint not found. Install with: brew install golangci-lint$(NC)" && exit 1)
	golangci-lint run

.PHONY: fmt
fmt: ## Format Go code
	@echo "$(BLUE)Formatting Go code...$(NC)"
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	go vet ./...

.PHONY: mod-tidy
mod-tidy: ## Tidy go modules
	@echo "$(BLUE)Tidying go modules...$(NC)"
	go mod tidy

.PHONY: mod-download
mod-download: ## Download go modules
	@echo "$(BLUE)Downloading go modules...$(NC)"
	go mod download

# Frontend
.PHONY: frontend-install
frontend-install: ## Install frontend dependencies
	@echo "$(BLUE)Installing frontend dependencies...$(NC)"
	cd $(FRONTEND_DIR) && npm install

.PHONY: frontend-build
frontend-build: ## Build frontend for production
	@echo "$(BLUE)Building frontend...$(NC)"
	cd $(FRONTEND_DIR) && npm run build

.PHONY: frontend-dev
frontend-dev: ## Start frontend development server
	@echo "$(BLUE)Starting frontend development server...$(NC)"
	cd $(FRONTEND_DIR) && npm run dev

.PHONY: frontend-lint
frontend-lint: ## Lint frontend code
	@echo "$(BLUE)Linting frontend code...$(NC)"
	cd $(FRONTEND_DIR) && npm run lint

.PHONY: frontend-type-check
frontend-type-check: ## Run TypeScript type checking
	@echo "$(BLUE)Running TypeScript type checking...$(NC)"
	cd $(FRONTEND_DIR) && npx tsc --noEmit

# Database
.PHONY: db-reset
db-reset: ## Reset database (delete klipd.db)
	@echo "$(YELLOW)Resetting database...$(NC)"
	rm -f klipd.db
	@echo "$(GREEN)Database reset complete$(NC)"

.PHONY: db-backup
db-backup: ## Backup database with timestamp
	@echo "$(BLUE)Backing up database...$(NC)"
	@if [ -f klipd.db ]; then \
		cp klipd.db "klipd_backup_$$(date +%Y%m%d_%H%M%S).db"; \
		echo "$(GREEN)Database backed up$(NC)"; \
	else \
		echo "$(YELLOW)No database file found$(NC)"; \
	fi

# Installation and Setup
.PHONY: install-deps
install-deps: ## Install all dependencies (Go modules + frontend)
	@echo "$(BLUE)Installing all dependencies...$(NC)"
	$(MAKE) mod-download
	$(MAKE) frontend-install

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(NC)"
	@echo "Installing golangci-lint..."
	@which golangci-lint > /dev/null || brew install golangci-lint
	@echo "Installing fswatch..."
	@which fswatch > /dev/null || brew install fswatch
	@echo "$(GREEN)Tools installed$(NC)"

.PHONY: setup
setup: install-tools install-deps ## Complete project setup
	@echo "$(GREEN)Project setup complete!$(NC)"

# Release and Distribution
.PHONY: release
release: test lint build ## Build release version (with tests and linting)
	@echo "$(GREEN)Release build complete!$(NC)"

.PHONY: package
package: build ## Package the app for distribution
	@echo "$(BLUE)Packaging app...$(NC)"
	@if [ -d "$(BUILD_DIR)/$(APP_NAME).app" ]; then \
		cd $(BUILD_DIR) && zip -r $(APP_NAME)-$$(date +%Y%m%d).zip $(APP_NAME).app; \
		echo "$(GREEN)Package created: $(BUILD_DIR)/$(APP_NAME)-$$(date +%Y%m%d).zip$(NC)"; \
	else \
		echo "$(RED)App not found. Run 'make build' first.$(NC)"; \
		exit 1; \
	fi

# Maintenance
.PHONY: check
check: fmt vet lint test ## Run all checks (format, vet, lint, test)
	@echo "$(GREEN)All checks passed!$(NC)"

.PHONY: quick-check
quick-check: fmt vet test-short ## Quick checks without linting
	@echo "$(GREEN)Quick checks passed!$(NC)"

.PHONY: logs
logs: ## Show app logs (if running)
	@echo "$(BLUE)Showing app logs...$(NC)"
	@if pgrep -f "$(APP_NAME)" > /dev/null; then \
		echo "$(GREEN)App is running$(NC)"; \
		# Add log viewing command here if you implement logging to file \
	else \
		echo "$(YELLOW)App is not running$(NC)"; \
	fi

.PHONY: kill
kill: ## Kill running app instances
	@echo "$(YELLOW)Killing running app instances...$(NC)"
	@pkill -f "$(APP_NAME)" || echo "No running instances found"

# Information
.PHONY: info
info: ## Show project information
	@echo "$(BLUE)Project Information$(NC)"
	@echo "App Name: $(APP_NAME)"
	@echo "Build Directory: $(BUILD_DIR)"
	@echo "Frontend Directory: $(FRONTEND_DIR)"
	@echo ""
	@echo "$(BLUE)Go Information$(NC)"
	@go version
	@echo "Go files: $(words $(GO_FILES))"
	@echo ""
	@echo "$(BLUE)Frontend Information$(NC)"
	@cd $(FRONTEND_DIR) && node --version && npm --version
	@echo "Frontend files: $(words $(FRONTEND_FILES))"

.PHONY: deps-status
deps-status: ## Show dependency status
	@echo "$(BLUE)Go Dependencies$(NC)"
	@go list -m all
	@echo ""
	@echo "$(BLUE)Frontend Dependencies$(NC)"
	@cd $(FRONTEND_DIR) && npm list --depth=0

# Development shortcuts
.PHONY: run
run: build ## Build and run the app
	@echo "$(BLUE)Running app...$(NC)"
	./$(BUILD_DIR)/$(APP_NAME).app/Contents/MacOS/$(APP_NAME)

.PHONY: debug
debug: build-debug ## Build and run debug version
	@echo "$(BLUE)Running debug version...$(NC)"
	./$(BUILD_DIR)/$(APP_NAME).app/Contents/MacOS/$(APP_NAME)

# Git hooks
.PHONY: pre-commit
pre-commit: quick-check ## Run pre-commit checks
	@echo "$(GREEN)Pre-commit checks passed!$(NC)"

# Clean up coverage files
.PHONY: clean-coverage
clean-coverage: ## Clean coverage files
	@echo "$(YELLOW)Cleaning coverage files...$(NC)"
	rm -f coverage.out coverage.html

# Full clean
.PHONY: clean-all
clean-all: clean clean-coverage ## Clean everything
	@echo "$(YELLOW)Cleaning everything...$(NC)"
	rm -rf frontend/node_modules
	go clean -cache -modcache -testcache
