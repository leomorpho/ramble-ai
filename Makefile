# Makefile for RambleAI - Video Editor with Ent ORM

# Load environment variables from .env files if they exist
-include .env
-include .env.wails
export

# Variables
APP_NAME = RambleAI
ENT_DIR = ./ent
DB_FILE = ~/Library/Application\ Support/RambleAI/database.db

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Available commands:"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk -F':.*?## ' '{printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
.PHONY: dev
dev: ## Start development server with hot reload
	wails dev


.PHONY: pb
pb: ## Smart PocketBase command - setup if needed, then start backend + SvelteKit (use NUKE=1 to delete database first)
	@echo "🚀 Starting PocketBase development environment..."
	@echo ""
	@echo "⚠️  Make sure you have set your API keys in pb-be/pb/.env:"
	@echo "   OPENROUTER_API_KEY=your-openrouter-key"
	@echo "   OPENAI_API_KEY=your-openai-key"
	@echo ""
	@if [ ! -f pb-be/pb/.env ]; then \
		echo "❌ Error: pb-be/pb/.env file not found!"; \
		echo "   Please create it with your API keys first."; \
		echo "   You can copy from pb-be/pb/.env.example"; \
		exit 1; \
	fi
	@echo "🔍 Checking if setup is needed..."
	@# Check if SvelteKit dependencies are installed
	@if [ ! -d pb-be/sk/node_modules ]; then \
		echo "📦 First run detected - setting up dependencies..."; \
		echo "   This may take a few minutes on first run..."; \
		echo ""; \
		echo "⏳ Installing SvelteKit frontend dependencies..."; \
		cd pb-be/sk && npm install --progress=true; \
		echo "✅ Frontend dependencies installed!"; \
		echo ""; \
		echo "⏳ Installing Go backend dependencies..."; \
		cd pb-be/pb && go mod tidy; \
		echo "✅ Backend dependencies installed!"; \
		echo ""; \
		echo "⏳ Building PocketBase backend..."; \
		cd pb-be/pb && go build -o pocketbase; \
		echo "✅ Backend built successfully!"; \
		echo ""; \
		echo "🎉 Setup complete!"; \
	else \
		echo "✅ Dependencies already installed, skipping setup..."; \
	fi
	@echo ""
	@echo "🎯 Starting services..."
	@echo "   📧 Email Testing: http://localhost:8025 (Mailpit)"
	@echo "   🔧 PocketBase Backend: http://localhost:8090"
	@echo "   🌐 SvelteKit Frontend: http://localhost:5174"
	@echo "   Admin UI: http://localhost:8090/_/"
	@echo "   API Endpoints:"
	@echo "     POST /api/ai/process-text"
	@echo "     POST /api/ai/process-audio" 
	@echo "     POST /api/generate-api-key"
	@echo "     GET /api/banners"
	@echo "     GET /api/banners/authenticated"
	@echo "   Development API Key: ra-dev-12345678901234567890123456789012"
	@echo "   (Auto-seeded for development - use in your Wails app)"
	@echo ""
	@echo "🧑‍💼 Admin User (for PocketBase frontend):"
	@echo "   Email: alice@test.com"
	@echo "   Password: password"
	@echo ""
	@if [ "$(NUKE)" = "1" ]; then \
		echo "💥 Nuking PocketBase database first..."; \
		cd pb-be && $(MAKE) nuke-db; \
	fi
	cd pb-be && $(MAKE) dev

.PHONY: pb-setup
pb-setup: ## Setup PocketBase backend dependencies (usually not needed - pb does this automatically)
	@echo "📦 Setting up PocketBase backend dependencies..."
	@cd pb-be && $(MAKE) setup
	@echo "✅ Backend setup complete!"


.PHONY: dev-remote
dev-remote: ## Start Wails app with remote AI backend enabled (configure .env.wails)
	@echo "🎯 Starting Wails app with remote AI backend enabled..."
	@echo "   Configuration loaded from .env.wails file"
	@echo "   Connecting to PocketBase at: $${REMOTE_AI_BACKEND_URL:-http://localhost:8090}"
	@echo "   Development API Key: ra-dev-12345678901234567890123456789012"
	@echo ""
	@echo "💡 Make sure:"
	@echo "   1. PocketBase backend is running: make backend"
	@echo "   2. .env.wails has USE_REMOTE_AI_BACKEND=true"
	@echo ""
	USE_REMOTE_AI_BACKEND=true wails dev


.PHONY: pb-only
pb-only: ## Start PocketBase backend server only (no SvelteKit frontend)
	@echo "🔧 Starting PocketBase backend server only..."
	@if [ ! -f pb-be/pb/.env ]; then \
		echo "❌ Error: pb-be/pb/.env file not found!"; \
		echo "   Please create it with your API keys first."; \
		exit 1; \
	fi
	@echo "🎯 Starting PocketBase backend..."
	@echo "   Admin UI: http://localhost:8090/_/"
	@echo "   Development API Key: ra-dev-12345678901234567890123456789012"
	cd pb-be && $(MAKE) dev-backend


.PHONY: pb-stop
pb-stop: ## Stop the PocketBase backend if running in background
	@if [ -f pb-backend.pid ]; then \
		echo "🛑 Stopping PocketBase backend..."; \
		kill $$(cat pb-backend.pid) 2>/dev/null && echo "✅ PocketBase backend stopped" || echo "⚠️  Backend process not found"; \
		rm -f pb-backend.pid pb-backend.log; \
	else \
		echo "ℹ️  No background PocketBase backend running"; \
	fi

.PHONY: kill-pb
kill-pb: ## Safely kill PocketBase processes (NEVER touches Firefox/OrbStack)
	@echo "🛑 Safely killing PocketBase processes..."
	@echo "🔍 Checking what will be killed:"
	@ps aux | grep "go run.*main.go serve" | grep -v grep || echo "   No Go PocketBase processes found"
	@ps aux | grep "pocketbase.*serve" | grep -v grep || echo "   No binary PocketBase processes found"
	@echo ""
	@echo "🔪 Killing PocketBase Go processes..."
	@pkill -f "go run.*main.go serve" && echo "✅ Killed Go PocketBase processes" || echo "ℹ️  No Go processes to kill"
	@echo "🔪 Killing PocketBase binary processes..."
	@pkill -f "pocketbase.*serve" && echo "✅ Killed binary PocketBase processes" || echo "ℹ️  No binary processes to kill"
	@echo "🔪 Killing any lingering child processes..."
	@ps aux | grep "main.*serve" | grep -v grep | awk '{print $$2}' | xargs -r kill -9 && echo "✅ Killed lingering processes" || echo "ℹ️  No lingering processes"
	@echo ""
	@echo "🔍 Verifying port 8090 is clear..."
	@if lsof -i :8090 | grep -v "COMMAND" | grep -q .; then \
		echo "⚠️  Warning: Other processes still using port 8090:"; \
		lsof -i :8090; \
		echo "   These are NOT PocketBase processes - leaving them alone"; \
	else \
		echo "✅ Port 8090 is clear"; \
	fi

.PHONY: be
be: ## Start PocketBase backend (use NUKE=1 to delete database first)
	@if [ ! -f pb-be/pb/.env ]; then \
		echo "❌ Error: pb-be/pb/.env file not found!"; \
		echo "   Please create it with your API keys first."; \
		echo "   You can copy from pb-be/pb/.env.example"; \
		exit 1; \
	fi
	@cd pb-be && $(MAKE) be NUKE=$(NUKE)

.PHONY: nuke-db
nuke-db: ## Delete PocketBase database completely
	@cd pb-be && $(MAKE) nuke-db

.PHONY: build
build: ## Build the application for production
	wails build -tags production

.PHONY: build-obfuscated
build-obfuscated: ## Build with obfuscation for code protection (requires Go 1.24+)
	@echo "🔒 Building with obfuscation for code protection..."
	@echo "Note: Requires Go 1.24+ for obfuscation support"
	@echo "Excluding Atlas SQL packages from obfuscation..."
	GOGARBLE="*,!ariga.io/atlas/..." wails build -tags production -obfuscated -garbleargs "-literals -tiny -seed=random"

.PHONY: build-all-platforms
build-all-platforms: ffmpeg-binaries ## Build for all platforms with embedded FFmpeg
	@echo "🚀 Building for all platforms with embedded FFmpeg..."
	wails build -tags production -platform=windows/amd64,darwin/amd64,linux/amd64
	@echo "✅ Multi-platform build complete!"

.PHONY: build-darwin-obfuscated
build-darwin-obfuscated: ffmpeg-binaries ## Build obfuscated binary for Darwin (macOS) amd64
	@echo "🔒🍎 Building obfuscated binary for Darwin (macOS) amd64..."
	@echo "Excluding Atlas SQL packages from obfuscation..."
	GOGARBLE="*,!ariga.io/atlas/..." wails build -tags production -obfuscated -garbleargs "-literals -tiny -seed=random" -platform=darwin/amd64
	@echo "✅ Darwin obfuscated build complete!"

.PHONY: build-windows-obfuscated
build-windows-obfuscated: ## Build obfuscated binary for Windows amd64 (REQUIRES WINDOWS HOST)
	@echo "❌ Windows obfuscated builds are not supported on macOS due to cross-compilation limitations."
	@echo "   This requires building on a Windows machine or using CI/CD with Windows runners."
	@echo "   Use 'make build-darwin-obfuscated' for macOS builds instead."

# Windows-specific build commands (to be run on Windows)
.PHONY: build-windows-on-windows
build-windows-on-windows: ffmpeg-binaries ## Build Windows binary on Windows host
	@echo "🪟 Building Windows binary..."
	wails build -tags production -platform=windows/amd64
	@echo "✅ Windows build complete!"

.PHONY: build-windows-installer
build-windows-installer: ## Build Windows NSIS installer (REQUIRES WINDOWS HOST)
	@echo "📦 Building Windows installer..."
	wails build -tags production -platform=windows/amd64 -nsis
	@echo "✅ Windows installer created!"

.PHONY: build-windows-obfuscated-on-windows
build-windows-obfuscated-on-windows: ffmpeg-binaries ## Build obfuscated Windows binary on Windows host
	@echo "🔒🪟 Building obfuscated Windows binary..."
	@echo "Excluding Atlas SQL packages from obfuscation..."
	set GOGARBLE=*,!ariga.io/atlas/... && wails build -tags production -obfuscated -garbleargs "-literals -tiny -seed=random" -platform=windows/amd64
	@echo "✅ Windows obfuscated build complete!"

# Helper function to detect and set APPLE_DEVELOPER_ID
define detect_developer_id
	@if [ -z "$$APPLE_DEVELOPER_ID" ]; then \
		echo "🔍 Auto-detecting Developer ID..."; \
		DEVELOPER_ID=$$(security find-identity -v -p codesigning | grep "Developer ID Application" | head -1 | awk -F'"' '{print $$2}'); \
		if [ -n "$$DEVELOPER_ID" ]; then \
			echo "✅ Found: $$DEVELOPER_ID"; \
			export APPLE_DEVELOPER_ID="$$DEVELOPER_ID"; \
		else \
			echo "❌ No Developer ID Application certificate found."; \
			echo "💡 Install a Developer ID certificate or set APPLE_DEVELOPER_ID manually."; \
			exit 1; \
		fi; \
	else \
		echo "✅ Using: $$APPLE_DEVELOPER_ID"; \
	fi
endef

# Simple local signing using script/sign (recommended)
.PHONY: sign
sign: ## Sign the RambleAI.app locally (requires Developer ID Application certificate)
	@echo "🔐 Signing RambleAI.app locally..."
	@if [ ! -f build/bin/RambleAI.app/Contents/MacOS/RambleAI ]; then \
		echo "❌ RambleAI.app not found. Run 'make build' first."; \
		exit 1; \
	fi
	$(call detect_developer_id)
	@APPLE_DEVELOPER_ID="$${APPLE_DEVELOPER_ID:-$$(security find-identity -v -p codesigning | grep 'Developer ID Application' | head -1 | awk -F'\"' '{print $$2}')}" ./script/sign build/bin/RambleAI.app

.PHONY: sign-zip
sign-zip: ## Create and sign a zip archive for distribution  
	@echo "📦 Creating signed zip archive..."
	@if [ ! -d build/bin/RambleAI.app ]; then \
		echo "❌ RambleAI.app not found. Run 'make build' first."; \
		exit 1; \
	fi
	$(call detect_developer_id)
	@cd build/bin && zip -r ../RambleAI-macos.zip RambleAI.app
	@APPLE_DEVELOPER_ID="$${APPLE_DEVELOPER_ID:-$$(security find-identity -v -p codesigning | grep 'Developer ID Application' | head -1 | awk -F'\"' '{print $$2}')}" ./script/sign build/RambleAI-macos.zip

.PHONY: build-and-sign
build-and-sign: ## Build and sign in one command (auto-detects certificate)
	@echo "🚀 Building and signing RambleAI..."
	$(call detect_developer_id)
	@$(MAKE) build
	@APPLE_DEVELOPER_ID="$${APPLE_DEVELOPER_ID:-$$(security find-identity -v -p codesigning | grep 'Developer ID Application' | head -1 | awk -F'\"' '{print $$2}')}" $(MAKE) sign

.PHONY: check-signing
check-signing: ## Check available Developer ID certificates
	@echo "🔍 Checking for Developer ID Application certificates..."
	@CERTS=$$(security find-identity -v -p codesigning | grep "Developer ID Application" || echo ""); \
	if [ -n "$$CERTS" ]; then \
		echo "✅ Found Developer ID certificates:"; \
		echo "$$CERTS"; \
		FIRST_CERT=$$(echo "$$CERTS" | head -1 | awk -F'"' '{print $$2}'); \
		echo ""; \
		echo "🎯 Will use: $$FIRST_CERT"; \
		echo "💡 To override, set: export APPLE_DEVELOPER_ID=\"Your Certificate Name\""; \
	else \
		echo "❌ No Developer ID Application certificates found."; \
		echo ""; \
		echo "To fix this:"; \
		echo "1. Get a Developer ID certificate from Apple Developer Portal"; \
		echo "2. Download and install it in Keychain Access"; \
		echo "3. Run 'make check-signing' again"; \
	fi

.PHONY: sign-windows-exe
sign-windows-exe: ## Sign Windows executable (requires code signing certificate)
	@echo "🔏 Signing Windows executable..."
	@echo "Note: Requires Windows SDK signtool and a code signing certificate"
	@if [ -z "$(CERT_THUMBPRINT)" ]; then \
		echo "Error: CERT_THUMBPRINT not set"; \
		echo "Usage: make sign-windows-exe CERT_THUMBPRINT=your-cert-thumbprint"; \
		echo "Or with PFX: make sign-windows-exe PFX_FILE=cert.pfx PFX_PASSWORD=password"; \
		exit 1; \
	fi
	@if [ -n "$(PFX_FILE)" ]; then \
		signtool sign /f "$(PFX_FILE)" /p "$(PFX_PASSWORD)" /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 build/bin/RambleAI.exe; \
	else \
		signtool sign /sha1 "$(CERT_THUMBPRINT)" /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 build/bin/RambleAI.exe; \
	fi
	@echo "✅ Windows executable signed!"

.PHONY: sign-windows-installer
sign-windows-installer: ## Sign Windows installer (requires code signing certificate)
	@echo "🔏 Signing Windows installer..."
	@if [ -z "$(CERT_THUMBPRINT)" ] && [ -z "$(PFX_FILE)" ]; then \
		echo "Error: Certificate not specified"; \
		echo "Usage: make sign-windows-installer CERT_THUMBPRINT=your-cert-thumbprint"; \
		echo "Or: make sign-windows-installer PFX_FILE=cert.pfx PFX_PASSWORD=password"; \
		exit 1; \
	fi
	@if [ -n "$(PFX_FILE)" ]; then \
		signtool sign /f "$(PFX_FILE)" /p "$(PFX_PASSWORD)" /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 build/bin/*-installer.exe; \
	else \
		signtool sign /sha1 "$(CERT_THUMBPRINT)" /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 build/bin/*-installer.exe; \
	fi
	@echo "✅ Windows installer signed!"

.PHONY: release-windows
release-windows: ## Complete Windows release build with installer (REQUIRES WINDOWS HOST)
	@echo "🚀 Building Windows release..."
	@echo "Note: This command must be run on a Windows machine"
	$(MAKE) build-windows-obfuscated-on-windows
	$(MAKE) build-windows-installer
	@if [ -n "$(CERT_THUMBPRINT)" ] || [ -n "$(PFX_FILE)" ]; then \
		$(MAKE) sign-windows-exe; \
		$(MAKE) sign-windows-installer; \
	else \
		echo "⚠️  Warning: No certificate provided, skipping signing"; \
	fi
	@echo "✅ Windows release build complete!"

.PHONY: build-linux-obfuscated
build-linux-obfuscated: ## Build obfuscated binary for Linux amd64 (REQUIRES LINUX HOST)
	@echo "❌ Linux obfuscated builds are not supported on macOS due to cross-compilation limitations."
	@echo "   This requires building on a Linux machine or using CI/CD with Linux runners."
	@echo "   Use 'make build-darwin-obfuscated' for macOS builds instead."

.PHONY: build-all-platforms-obfuscated
build-all-platforms-obfuscated: build-darwin-obfuscated ## Build obfuscated binaries for all supported platforms (macOS only on this host)
	@echo "✅ macOS obfuscated build complete!"
	@echo "ℹ️  Note: Windows and Linux obfuscated builds require their respective host platforms."

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
	@if [ -f "$(DB_FILE)" ]; then \
		sqlite3 "$(DB_FILE)" ".schema"; \
	else \
		echo "Database file $(DB_FILE) not found. Run the app first to create it."; \
	fi

.PHONY: db-reset
db-reset: ## Reset database (WARNING: deletes all data)
	@echo "WARNING: This will delete all data in the database!"
	@echo "Press Ctrl+C to cancel, or Enter to continue..."
	@read
	rm -f "$(DB_FILE)"
	@echo "Database deleted. Run the app to recreate it."

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

# Landing Page
.PHONY: landing-dev
landing-dev: ## Start landing page development server
	cd landing-page && pnpm dev

# Testing
.PHONY: test
test: ## Run all tests (Go + Frontend, excluding ent package)
	@echo "🧪 Running Go tests (excluding ent package)..."
	go test $$(go list ./... | grep -v "/ent") -short
	@echo "🧪 Running Frontend tests..."
	cd frontend && npm test

.PHONY: test-go
test-go: ## Run Go tests only (excluding ent package)
	go test $$(go list ./... | grep -v "/ent")

.PHONY: test-go-short
test-go-short: ## Run Go tests in short mode (excluding ent package)
	go test $$(go list ./... | grep -v "/ent") -short

.PHONY: test-go-verbose
test-go-verbose: ## Run Go tests with verbose output (excluding ent package)
	go test -v $$(go list ./... | grep -v "/ent")

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
test-verbose: ## Run all tests with verbose output (excluding ent package)
	@echo "🧪 Running Go tests (verbose, excluding ent package)..."
	go test -v $$(go list ./... | grep -v "/ent")
	@echo "🧪 Running Frontend tests..."
	cd frontend && npm test

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report (excluding ent package)
	@echo "🧪 Running Go tests with coverage (excluding ent package)..."
	@go test -short -coverprofile=coverage.out $$(go list ./... | grep -v "/ent")
	@if [ -f coverage.out ]; then \
		go tool cover -html=coverage.out -o coverage.html; \
		echo "📊 Coverage report generated: coverage.html"; \
		echo ""; \
		echo "📈 Coverage summary (excluding ent):"; \
		go tool cover -func=coverage.out | grep -E "total:" | awk '{print "Total coverage: " $$3}'; \
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

# Code signing and distribution
.PHONY: sign-app
sign-app: ## Sign the macOS app (requires Apple Developer Certificate)
	@echo "🔏 Signing macOS app..."
	@if [ ! -d "build/bin/RambleAI.app" ]; then \
		echo "Error: App bundle not found. Run 'make build-darwin-obfuscated' first"; \
		exit 1; \
	fi
	@if [ -n "$(APPLE_DEVELOPER_CERTIFICATE_P12_PATH)" ] && [ -f "$(APPLE_DEVELOPER_CERTIFICATE_P12_PATH)" ]; then \
		echo "Using certificate file: $(APPLE_DEVELOPER_CERTIFICATE_P12_PATH)"; \
		if [ -z "$(APPLE_DEVELOPER_CERTIFICATE_PASSWORD)" ]; then \
			echo "Error: APPLE_DEVELOPER_CERTIFICATE_PASSWORD required for P12 file"; \
			exit 1; \
		fi; \
		security import "$(APPLE_DEVELOPER_CERTIFICATE_P12_PATH)" -P "$(APPLE_DEVELOPER_CERTIFICATE_PASSWORD)" -T /usr/bin/codesign -T /usr/bin/security 2>/dev/null || true; \
	elif [ -n "$(APPLE_DEVELOPER_CERTIFICATE_P12_BASE64)" ]; then \
		echo "Using base64 encoded certificate..."; \
		if [ -z "$(APPLE_DEVELOPER_CERTIFICATE_PASSWORD)" ]; then \
			echo "Error: APPLE_DEVELOPER_CERTIFICATE_PASSWORD required"; \
			exit 1; \
		fi; \
		mkdir -p /tmp/cert; \
		echo "$(APPLE_DEVELOPER_CERTIFICATE_P12_BASE64)" | base64 --decode > /tmp/cert/cert.p12; \
		security import /tmp/cert/cert.p12 -P "$(APPLE_DEVELOPER_CERTIFICATE_PASSWORD)" -T /usr/bin/codesign -T /usr/bin/security 2>/dev/null || true; \
		rm -rf /tmp/cert; \
	else \
		echo "Error: No certificate provided. Set either:"; \
		echo "  APPLE_DEVELOPER_CERTIFICATE_P12_PATH (path to .p12 file)"; \
		echo "  or APPLE_DEVELOPER_CERTIFICATE_P12_BASE64 (base64 encoded)"; \
		exit 1; \
	fi
	@echo "Finding signing identity..."
	@IDENTITY=$$(security find-identity -v -p codesigning | grep "Developer ID Application" | head -1 | awk -F'"' '{print $$2}'); \
	if [ -n "$$IDENTITY" ]; then \
		echo "Signing with identity: $$IDENTITY"; \
		codesign --force --options runtime --sign "$$IDENTITY" build/bin/RambleAI.app/Contents/MacOS/RambleAI; \
		codesign --force --options runtime --sign "$$IDENTITY" build/bin/RambleAI.app; \
		echo "✅ App signed successfully"; \
	else \
		echo "❌ No Developer ID Application certificate found"; \
		exit 1; \
	fi

.PHONY: verify-signature
verify-signature: ## Verify the app signature
	@echo "🔍 Verifying app signature..."
	codesign --verify --verbose build/bin/RambleAI.app
	spctl --assess --verbose build/bin/RambleAI.app
	@echo "✅ Signature verification complete"

.PHONY: create-dmg
create-dmg: ## Create a DMG installer
	@echo "📦 Creating DMG installer..."
	@if [ ! -d "build/bin/RambleAI.app" ]; then \
		echo "Error: App bundle not found. Run 'make build-darwin-obfuscated' first"; \
		exit 1; \
	fi
	@mkdir -p build/dmg
	@cp -R build/bin/RambleAI.app build/dmg/
	@ln -sf /Applications build/dmg/Applications
	hdiutil create -volname "RambleAI" -srcfolder build/dmg -ov -format UDZO build/RambleAI.dmg
	@rm -rf build/dmg
	@echo "✅ DMG created: build/RambleAI.dmg"

.PHONY: sign-dmg
sign-dmg: ## Sign the DMG (requires Apple Developer Certificate)
	@echo "🔏 Signing DMG..."
	@if [ ! -f "build/RambleAI.dmg" ]; then \
		echo "Error: DMG not found. Run 'make create-dmg' first"; \
		exit 1; \
	fi
	@IDENTITY=$$(security find-identity -v -p codesigning | grep "Developer ID Application" | head -1 | awk -F'"' '{print $$2}'); \
	if [ -n "$$IDENTITY" ]; then \
		echo "Signing DMG with identity: $$IDENTITY"; \
		codesign --force --sign "$$IDENTITY" build/RambleAI.dmg; \
		echo "✅ DMG signed successfully"; \
	else \
		echo "❌ No Developer ID Application certificate found"; \
		exit 1; \
	fi

.PHONY: notarize-app
notarize-app: ## Notarize the app with Apple (requires Apple ID credentials)
	@echo "🍎 Notarizing app with Apple..."
	@if [ -z "$(APPLE_ID)" ] || [ -z "$(APPLE_ID_PASSWORD)" ] || [ -z "$(TEAM_ID)" ]; then \
		echo "Error: Apple ID credentials not set"; \
		echo "Required environment variables:"; \
		echo "  APPLE_ID=your-apple-id@email.com"; \
		echo "  APPLE_ID_PASSWORD=your-app-specific-password"; \
		echo "  TEAM_ID=your-team-id"; \
		exit 1; \
	fi
	@if [ ! -f "build/RambleAI.dmg" ]; then \
		echo "Error: DMG not found. Run 'make create-dmg' and 'make sign-dmg' first"; \
		exit 1; \
	fi
	xcrun notarytool submit build/RambleAI.dmg --apple-id "$(APPLE_ID)" --password "$(APPLE_ID_PASSWORD)" --team-id "$(TEAM_ID)" --wait
	xcrun stapler staple build/RambleAI.dmg
	@echo "✅ App notarized and stapled successfully"

.PHONY: release-darwin
release-darwin: build-darwin-obfuscated sign-app create-dmg sign-dmg notarize-app ## Complete macOS release build with signing and notarization
	@echo "🚀 macOS release build complete!"
	@echo "📦 Signed and notarized DMG: build/RambleAI.dmg"
	@echo ""
	@echo "Required environment variables for signing and notarization:"
	@echo "  APPLE_DEVELOPER_CERTIFICATE_P12_BASE64='base64-encoded-p12-cert'"
	@echo "  APPLE_DEVELOPER_CERTIFICATE_PASSWORD='p12-password'"
	@echo "  APPLE_ID='your-apple-id@email.com'"
	@echo "  APPLE_ID_PASSWORD='your-app-specific-password'"
	@echo "  TEAM_ID='your-team-id'"
	@echo ""
	@echo "💡 Tip: These are automatically loaded from .env file if it exists"

.PHONY: release-local
release-local: ffmpeg-binaries build-darwin-obfuscated create-dmg ## Local build without signing (for testing)
	@echo "🚀 Local macOS build complete!"
	@echo "📦 Unsigned DMG: build/RambleAI.dmg"
	@echo "⚠️  DMG is unsigned - use 'make release-darwin' for signed version"

# FFmpeg binaries
.PHONY: ffmpeg-binaries
ffmpeg-binaries: ## Download FFmpeg binaries for all platforms
	@echo "📦 Downloading FFmpeg binaries..."
	@mkdir -p binaries/static
	@echo "Downloading Windows binary..."
	@curl -L -o binaries/static/ffmpeg-windows.zip https://github.com/ffbinaries/ffbinaries-prebuilt/releases/download/v6.1/ffmpeg-6.1-win-64.zip
	@echo "Downloading macOS binary..."
	@curl -L -o binaries/static/ffmpeg-macos.zip https://github.com/ffbinaries/ffbinaries-prebuilt/releases/download/v6.1/ffmpeg-6.1-macos-64.zip
	@echo "Downloading Linux binary..."
	@curl -L -o binaries/static/ffmpeg-linux.zip https://github.com/ffbinaries/ffbinaries-prebuilt/releases/download/v6.1/ffmpeg-6.1-linux-64.zip
	@echo "Extracting binaries..."
	@cd binaries/static && unzip -o ffmpeg-windows.zip && mv ffmpeg.exe ffmpeg-windows-amd64.exe
	@cd binaries/static && unzip -o ffmpeg-macos.zip && mv ffmpeg ffmpeg-darwin-amd64
	@cd binaries/static && unzip -o ffmpeg-linux.zip && mv ffmpeg ffmpeg-linux-amd64
	@echo "Cleaning up zip files..."
	@rm -f binaries/static/*.zip
	@echo "✅ FFmpeg binaries downloaded and extracted!"

.PHONY: ffmpeg-clean
ffmpeg-clean: ## Clean downloaded FFmpeg binaries
	@echo "🧹 Cleaning FFmpeg binaries..."
	@rm -rf binaries/static
	@echo "✅ FFmpeg binaries cleaned!"

# Database utilities
.PHONY: db-shell
db-shell: ## Open SQLite shell for the database
	@if [ -f "$(DB_FILE)" ]; then \
		sqlite3 "$(DB_FILE)"; \
	else \
		echo "Database file $(DB_FILE) not found. Run the app first to create it."; \
	fi

.PHONY: db-backup
db-backup: ## Backup database to timestamped file
	@if [ -f "$(DB_FILE)" ]; then \
		cp "$(DB_FILE)" "$(DB_FILE).backup.$$(date +%Y%m%d_%H%M%S)"; \
		echo "Database backed up to $(DB_FILE).backup.$$(date +%Y%m%d_%H%M%S)"; \
	else \
		echo "Database file $(DB_FILE) not found."; \
	fi