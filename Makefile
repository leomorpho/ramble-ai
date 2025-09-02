# Makefile for RambleAI

# Load environment variables from .env files if they exist
-include .env
-include .env.wails
export

# Variables
APP_NAME = RambleAI
DB_FILE = ~/Library/Application\ Support/RambleAI/database.db

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Available commands:"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk -F':.*?## ' '{printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# === DEVELOPMENT COMMANDS ===

.PHONY: dev
dev: ## Start development server with local PocketBase backend
	@echo "🛠️  Starting Wails app with local PocketBase development configuration..."
	@echo "   - DEVELOPMENT mode: ENABLED (triggers auto-seeding)"
	@echo "   - Backend URL: http://localhost:8090"
	@echo "   - Development API key: ra-dev-12345678901234567890123456789012"
	@echo ""
	@echo "⚠️  Make sure PocketBase is running: make pb"
	@echo ""
	DEVELOPMENT=true \
	USE_REMOTE_AI_BACKEND=true \
	REMOTE_AI_BACKEND_URL=http://localhost:8090 \
	RAMBLE_FRONTEND_URL=http://localhost:8090 \
	wails dev

.PHONY: dev-prod
dev-prod: ## Start development server with production-like configuration for testing
	@echo "🚀 Starting Wails app with production-like configuration..."
	@echo "   - Remote backend URL: https://api.ramble.goosebyteshq.com"
	@echo "   - Frontend URL: https://ramble.goosebyteshq.com"
	@echo ""
	@echo "⚠️  Note: Requires internet connection to reach remote services"
	@echo ""
	USE_REMOTE_AI_BACKEND=true \
	REMOTE_AI_BACKEND_URL=https://api.ramble.goosebyteshq.com \
	RAMBLE_FRONTEND_URL=https://ramble.goosebyteshq.com \
	wails dev

# === POCKETBASE BACKEND ===

.PHONY: pb
pb: ## Start PocketBase backend + SvelteKit (use NUKE=1 to delete database first)
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
	@echo "   Development API Key: ra-dev-12345678901234567890123456789012"
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

.PHONY: stripe
stripe: ## Start Stripe webhook forwarding (run in separate terminal)
	@echo "💳 Starting Stripe webhook forwarding..."
	@if ! command -v stripe >/dev/null 2>&1; then \
		echo "⚠️  Stripe CLI not found. Install it from: https://stripe.com/docs/stripe-cli"; \
		echo "   On macOS: brew install stripe/stripe-cli/stripe"; \
		exit 1; \
	fi
	@echo "🔗 Forwarding webhooks to: http://127.0.0.1:8090/api/webhooks/stripe"
	@echo "📝 Make sure PocketBase backend is running on port 8090"
	@echo ""
	stripe listen --forward-to=127.0.0.1:8090/api/webhooks/stripe

.PHONY: kill-pb
kill-pb: ## Safely kill PocketBase processes (NEVER touches Firefox/OrbStack)
	@echo "🛑 Safely killing PocketBase processes..."
	@echo "🔍 Checking what will be killed:"
	@ps aux | grep "go run.*main.go serve" | grep -v grep || echo "   No Go PocketBase processes found"
	@ps aux | grep "pocketbase.*serve" | grep -v grep || echo "   No binary PocketBase processes found"
	@ps aux | grep -E "(make.*pb|modd)" | grep -v grep || echo "   No parent make/modd processes found"
	@echo ""
	@echo "🔪 Killing parent make pb processes..."
	@pkill -f "make.*pb" && echo "✅ Killed make pb processes" || echo "ℹ️  No make pb processes to kill"
	@echo "🔪 Killing modd processes (auto-restart managers)..."
	@pkill -f "modd" && echo "✅ Killed modd processes" || echo "ℹ️  No modd processes to kill"
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

# === BUILD COMMANDS ===

.PHONY: build
build: ## Build the application for production
	wails build -tags production

.PHONY: build-prod
build-prod: ## Build obfuscated production binary for macOS (FFmpeg auto-downloads)
	@echo "🔒 Building obfuscated production binary for macOS..."
	@echo "ℹ️  FFmpeg will be auto-downloaded on first run if needed"
	@echo "Excluding Atlas SQL packages from obfuscation..."
	GOGARBLE="*,!ariga.io/atlas/..." wails build -tags production -obfuscated -garbleargs "-literals -tiny -seed=random" -platform=darwin/amd64
	@echo "✅ Production build complete!"

.PHONY: test-prod
test-prod: ## Build and run production version locally for testing (FFmpeg auto-downloads)
	@echo "🔨 Building production version for local testing..."
	@echo "ℹ️  FFmpeg will be auto-downloaded on first run if needed"
	wails build -tags production
	@echo ""
	@echo "✅ Build complete! Running production build locally..."
	@echo "   Binary location: ./build/bin/RambleAI.app/Contents/MacOS/RambleAI"
	@echo ""
	./build/bin/RambleAI.app/Contents/MacOS/RambleAI

# === SIGNING & DISTRIBUTION ===

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

.PHONY: create-dmg
create-dmg: ## Create a DMG installer
	@echo "📦 Creating DMG installer..."
	@if [ ! -d "build/bin/RambleAI.app" ]; then \
		echo "Error: App bundle not found. Run 'make build' or 'make build-prod' first"; \
		exit 1; \
	fi
	@mkdir -p build/dmg
	@cp -R build/bin/RambleAI.app build/dmg/
	@ln -sf /Applications build/dmg/Applications
	hdiutil create -volname "RambleAI" -srcfolder build/dmg -ov -format UDZO build/RambleAI.dmg
	@rm -rf build/dmg
	@echo "✅ DMG created: build/RambleAI.dmg"

.PHONY: release-local
release-local: build-prod create-dmg ## Local build without signing (FFmpeg auto-downloads)
	@echo "🚀 Local macOS build complete!"
	@echo "📦 Unsigned DMG: build/RambleAI.dmg"
	@echo "⚠️  DMG is unsigned - use signing commands for signed version"
	@echo "ℹ️  FFmpeg will be auto-downloaded on first run if needed"

# === TESTING ===

.PHONY: test
test: ## Run all tests (Go + Frontend)
	@echo "🧪 Running Go tests..."
	go test $$(go list ./... | grep -v "/ent") -short
	@echo "🧪 Running Frontend tests..."
	cd frontend && npm test

.PHONY: test-go
test-go: ## Run Go tests only
	go test $$(go list ./... | grep -v "/ent")

# === FFMPEG (AUTO-DOWNLOAD) ===
# FFmpeg is now auto-downloaded by the app on first run
# No need to manually download binaries anymore

.PHONY: ffmpeg-clean
ffmpeg-clean: ## Clean auto-downloaded FFmpeg binaries
	@echo "🧹 Cleaning auto-downloaded FFmpeg binaries..."
	@rm -rf ~/Library/Application\ Support/RambleAI/binaries
	@rm -rf binaries  # Development mode binaries
	@echo "✅ FFmpeg binaries cleaned!"
	@echo "ℹ️  FFmpeg will be re-downloaded on next app run"

# === CLEANUP ===

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf build/
	rm -rf frontend/build/
	rm -rf frontend/.svelte-kit/