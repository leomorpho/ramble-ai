# Makefile for MYAPP - Video Editor with Ent ORM

# Variables
APP_NAME = MYAPP
ENT_DIR = ./ent
DB_FILE = database.db

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
.PHONY: dev
dev: ## Start development server with hot reload
	wails dev

.PHONY: build
build: ## Build the application for production
	wails build

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf build/
	rm -rf frontend/build/
	rm -rf frontend/.svelte-kit/

# Database and Ent ORM
.PHONY: new-entity
new-entity: ## Create a new Ent entity (usage: make new-entity name=EntityName)
	@if [ -z "$(name)" ]; then \
		echo "Error: Please provide entity name (e.g., make new-entity name=User)"; \
		exit 1; \
	fi
	go run entgo.io/ent/cmd/ent new $(name)

.PHONY: generate
generate: ## Generate Ent code from schema definitions
	go generate ./ent

.PHONY: migrate-create
migrate-create: ## Create a new migration file (usage: make migrate-create name=migration_name)
	@if [ -z "$(name)" ]; then \
		echo "Error: Please provide migration name (e.g., make migrate-create name=add_users_table)"; \
		exit 1; \
	fi
	go run -mod=mod entgo.io/ent/cmd/ent migrate hash

.PHONY: migrate-up
migrate-up: ## Apply all pending migrations
	go run main.go --migrate

.PHONY: migrate-down
migrate-down: ## Rollback the last migration (WARNING: may cause data loss)
	@echo "WARNING: This will rollback the last migration and may cause data loss!"
	@echo "Press Ctrl+C to cancel, or Enter to continue..."
	@read
	rm -f $(DB_FILE)
	@echo "Database reset. Run 'make migrate-up' to apply migrations."

.PHONY: schema-inspect
schema-inspect: ## Inspect current database schema
	@if [ -f $(DB_FILE) ]; then \
		sqlite3 $(DB_FILE) ".schema"; \
	else \
		echo "Database file $(DB_FILE) not found. Run 'make migrate-up' first."; \
	fi

.PHONY: db-reset
db-reset: ## Reset database (WARNING: deletes all data)
	@echo "WARNING: This will delete all data in the database!"
	@echo "Press Ctrl+C to cancel, or Enter to continue..."
	@read
	rm -f $(DB_FILE)
	@echo "Database deleted. Run 'make migrate-up' to recreate."

# Frontend
.PHONY: frontend-install
frontend-install: ## Install frontend dependencies
	cd frontend && pnpm install

.PHONY: frontend-dev
frontend-dev: ## Start frontend development server
	cd frontend && pnpm dev

.PHONY: frontend-build
frontend-build: ## Build frontend for production
	cd frontend && pnpm build

.PHONY: frontend-check
frontend-check: ## Type check frontend code
	cd frontend && pnpm check

.PHONY: frontend-check-watch
frontend-check-watch: ## Type check frontend code in watch mode
	cd frontend && pnpm check:watch

# Testing
.PHONY: test
test: ## Run all tests (Go + Frontend)
	@echo "🧪 Running Go tests..."
	go test -p 1 ./... -short
	@echo "🧪 Running Frontend tests..."
	cd frontend && npm test

.PHONY: test-go
test-go: ## Run Go tests only
	go test -p 1 ./...

.PHONY: test-go-short
test-go-short: ## Run Go tests in short mode
	go test -p 1 ./... -short

.PHONY: test-go-verbose
test-go-verbose: ## Run Go tests with verbose output
	go test -p 1 -v ./...

.PHONY: test-frontend
test-frontend: ## Run Frontend tests only
	cd frontend && npm test

.PHONY: test-frontend-ui
test-frontend-ui: ## Run Frontend tests with UI
	cd frontend && npm run test:ui

.PHONY: test-frontend-run
test-frontend-run: ## Run Frontend tests once (CI mode)
	cd frontend && npm run test:run

.PHONY: test-verbose
test-verbose: ## Run all tests with verbose output
	@echo "🧪 Running Go tests (verbose)..."
	go test -p 1 -v ./...
	@echo "🧪 Running Frontend tests..."
	cd frontend && npm test

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "🧪 Running Go tests with coverage..."
	go test -p 1 ./... -short -coverprofile=coverage.out
	@if [ -f coverage.out ]; then \
		go tool cover -html=coverage.out -o coverage.html; \
		echo "📊 Coverage report generated: coverage.html"; \
	fi
	@echo "🧪 Running Frontend tests..."
	cd frontend && npm test

.PHONY: test-watch
test-watch: ## Run tests in watch mode
	@echo "Starting test watchers..."
	@echo "Frontend tests will watch for changes..."
	cd frontend && npm run test:watch

# Maintenance
.PHONY: deps
deps: ## Update Go dependencies
	go mod tidy
	go mod download

.PHONY: format
format: ## Format Go code
	go fmt ./...

.PHONY: lint
lint: ## Run Go linter (requires golangci-lint)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it from https://golangci-lint.run/"; \
	fi

# Common workflows
.PHONY: setup
setup: deps frontend-install generate ## Set up the project for development
	@echo "Setup complete! Run 'make dev' to start development server."

.PHONY: full-build
full-build: clean frontend-install frontend-build build ## Clean build from scratch

# Database utilities
.PHONY: db-shell
db-shell: ## Open SQLite shell for the database
	@if [ -f $(DB_FILE) ]; then \
		sqlite3 $(DB_FILE); \
	else \
		echo "Database file $(DB_FILE) not found. Run 'make migrate-up' first."; \
	fi

.PHONY: db-backup
db-backup: ## Backup database to timestamped file
	@if [ -f $(DB_FILE) ]; then \
		cp $(DB_FILE) $(DB_FILE).backup.$$(date +%Y%m%d_%H%M%S); \
		echo "Database backed up to $(DB_FILE).backup.$$(date +%Y%m%d_%H%M%S)"; \
	else \
		echo "Database file $(DB_FILE) not found."; \
	fi