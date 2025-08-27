.PHONY: help dev dev-backend dev-frontend storybook build test lint format clean install deps check

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Development commands
dev: mailpit-up ## Start full development environment (backend + frontend + mailpit)
	@echo "🚀 Starting full development environment..."
	@echo "📧 Mailpit (email testing): http://localhost:8025"
	@echo "🔧 PocketBase (backend): http://localhost:8090"
	@echo "🌐 SvelteKit (frontend): http://localhost:5174"
	@make -j2 dev-backend dev-frontend

dev-backend: ## Start PocketBase backend in watch mode
	@echo "Starting PocketBase backend..."
	@cd pb && modd

dev-frontend: ## Start SvelteKit frontend in development mode
	@echo "Starting SvelteKit frontend..."
	@cd sk && npm run dev

storybook: ## Start Storybook development server
	@echo "Starting Storybook..."
	@cd sk && npm run storybook

# Build commands
build: ## Build the entire project
	@echo "Building frontend..."
	@cd sk && npm run build
	@echo "Building backend..."
	@cd pb && go build
	@echo "Build complete!"

build-frontend: ## Build only the frontend
	@cd sk && npm run build

build-backend: ## Build only the backend
	@cd pb && go build

# Testing commands
test: ## Run all tests
	@cd sk && npm run test

test-unit: ## Run unit tests only
	@cd sk && npm run test:unit

test-e2e: ## Run e2e tests only
	@cd sk && npm run test:e2e

# Code quality commands
lint: ## Run linting
	@cd sk && npm run lint

format: ## Format code
	@cd sk && npm run format

check: ## Run type checking
	@cd sk && npm run check

# Setup commands
setup: ## Complete project setup (deps + git hooks + build)
	@echo "🚀 Setting up PocketBase Starter Kit..."
	@make deps
	@echo "🪝 Setting up git pre-commit hook..."
	@if [ ! -f .git/hooks/pre-commit ]; then \
		cp scripts/pre-commit .git/hooks/pre-commit 2>/dev/null || \
		echo '#!/bin/bash\necho "🧪 Running pre-commit tests..."\ncd sk\nif ! npm run test:unit -- --run; then\n    echo "❌ Unit tests failed. Commit aborted."\n    exit 1\nfi\nif ! npm run test:e2e; then\n    echo "❌ E2E tests failed. Commit aborted."\n    exit 1\nfi\necho "✅ All tests passed! Proceeding with commit."' > .git/hooks/pre-commit; \
	fi
	@chmod +x .git/hooks/pre-commit
	@echo "🔧 Building PocketBase backend..."
	@cd pb && go build
	@echo "✅ Setup complete! Run 'make dev' to start development."

install: deps ## Install all dependencies

deps: ## Install dependencies for both frontend and backend
	@echo "Installing frontend dependencies..."
	@cd sk && npm install
	@echo "Installing backend dependencies..."
	@cd pb && go mod tidy
	@echo "Dependencies installed!"

# Production commands
preview: build-frontend ## Preview production build
	@cd sk && npm run preview

serve: build-frontend ## Serve the app with PocketBase
	@cd pb && ./pocketbase serve --http 0.0.0.0:8090

# Cleanup commands
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf sk/build sk/.svelte-kit sk/node_modules/.vite
	@rm -f pb/pocketbase pb/pb_data/logs.db*
	@echo "Clean complete!"

clean-deps: ## Remove all dependencies
	@echo "Removing dependencies..."
	@rm -rf sk/node_modules sk/pnpm-lock.yaml
	@cd pb && go clean -modcache
	@echo "Dependencies removed!"

# Email testing with Mailpit
mailpit: ## Start Mailpit email testing server (foreground)
	@echo "📧 Starting Mailpit email testing server..."
	@echo "📧 Web UI: http://localhost:8025"
	@echo "📧 SMTP: localhost:1025"
	docker-compose up mailpit

mailpit-up: ## Start Mailpit in background
	@echo "📧 Checking Mailpit status..."
	@if curl -s http://localhost:8025 > /dev/null 2>&1; then \
		echo "📧 Mailpit already running at http://localhost:8025"; \
	else \
		echo "📧 Starting Mailpit in background..."; \
		docker-compose up -d mailpit 2>/dev/null || \
		docker run -d --name mailpit-shared -p 8025:8025 -p 1025:1025 axllent/mailpit 2>/dev/null || \
		echo "⚠️  Could not start Mailpit (port may be in use by another service)"; \
	fi
	@echo "📧 Mailpit should be available at: http://localhost:8025"

mailpit-down: ## Stop Mailpit
	@echo "📧 Stopping Mailpit..."
	@docker-compose down mailpit 2>/dev/null || echo "📧 No docker-compose Mailpit found"
	@docker stop mailpit-shared 2>/dev/null && docker rm mailpit-shared 2>/dev/null || echo "📧 No shared Mailpit container found"
	@echo "📧 Mailpit stopped (if it was running from this project)"

mailpit-logs: ## Show Mailpit logs
	@echo "📧 Showing Mailpit logs..."
	@docker-compose logs -f mailpit 2>/dev/null || \
	docker logs -f mailpit-shared 2>/dev/null || \
	echo "📧 No Mailpit logs found (check if Mailpit is running)"

mailpit-status: ## Check Mailpit status
	@echo "📧 Checking Mailpit status..."
	@if curl -s http://localhost:8025 > /dev/null 2>&1; then \
		echo "✅ Mailpit is running at http://localhost:8025"; \
		echo "📧 SMTP available at localhost:1025"; \
	else \
		echo "❌ Mailpit is not accessible at http://localhost:8025"; \
		echo "💡 Run 'make mailpit-up' to start it"; \
	fi

# Utility commands
logs: ## Show PocketBase logs
	@cd pb && tail -f pb_data/logs.db

backup: ## Create a backup of PocketBase data
	@cd pb && cp -r pb_data pb_data_backup_$(shell date +%Y%m%d_%H%M%S)
	@echo "Backup created!"