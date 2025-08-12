# Makefile for MYAPP - Video Editor with Ent ORM

# Load environment variables from .env file if it exists
-include .env
export

# Variables
APP_NAME = MYAPP
ENT_DIR = ./ent
DB_FILE = ~/Library/Application\ Support/MYAPP/database.db

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Available commands:"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk -F':.*?## ' '{printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
.PHONY: dev
dev: ## Start development server with hot reload
	wails dev

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
		signtool sign /f "$(PFX_FILE)" /p "$(PFX_PASSWORD)" /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 build/bin/MYAPP.exe; \
	else \
		signtool sign /sha1 "$(CERT_THUMBPRINT)" /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 build/bin/MYAPP.exe; \
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

# Testing
.PHONY: test
test: ## Run all tests (Go + Frontend)
	@echo "🧪 Running Go tests..."
	go test ./... -short
	@echo "🧪 Running Frontend tests..."
	cd frontend && npm test

.PHONY: test-go
test-go: ## Run Go tests only
	go test ./...

.PHONY: test-go-short
test-go-short: ## Run Go tests in short mode
	go test ./... -short

.PHONY: test-go-verbose
test-go-verbose: ## Run Go tests with verbose output
	go test -v ./...

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
	go test -v ./...
	@echo "🧪 Running Frontend tests..."
	cd frontend && npm test

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "🧪 Running Go tests with coverage..."
	go test ./... -short -coverprofile=coverage.out
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

# Code signing and distribution
.PHONY: sign-app
sign-app: ## Sign the macOS app (requires Apple Developer Certificate)
	@echo "🔏 Signing macOS app..."
	@if [ ! -d "build/bin/MYAPP.app" ]; then \
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
		codesign --force --options runtime --sign "$$IDENTITY" build/bin/MYAPP.app/Contents/MacOS/MYAPP; \
		codesign --force --options runtime --sign "$$IDENTITY" build/bin/MYAPP.app; \
		echo "✅ App signed successfully"; \
	else \
		echo "❌ No Developer ID Application certificate found"; \
		exit 1; \
	fi

.PHONY: verify-signature
verify-signature: ## Verify the app signature
	@echo "🔍 Verifying app signature..."
	codesign --verify --verbose build/bin/MYAPP.app
	spctl --assess --verbose build/bin/MYAPP.app
	@echo "✅ Signature verification complete"

.PHONY: create-dmg
create-dmg: ## Create a DMG installer
	@echo "📦 Creating DMG installer..."
	@if [ ! -d "build/bin/MYAPP.app" ]; then \
		echo "Error: App bundle not found. Run 'make build-darwin-obfuscated' first"; \
		exit 1; \
	fi
	@mkdir -p build/dmg
	@cp -R build/bin/MYAPP.app build/dmg/
	@ln -sf /Applications build/dmg/Applications
	hdiutil create -volname "MYAPP" -srcfolder build/dmg -ov -format UDZO build/MYAPP.dmg
	@rm -rf build/dmg
	@echo "✅ DMG created: build/MYAPP.dmg"

.PHONY: sign-dmg
sign-dmg: ## Sign the DMG (requires Apple Developer Certificate)
	@echo "🔏 Signing DMG..."
	@if [ ! -f "build/MYAPP.dmg" ]; then \
		echo "Error: DMG not found. Run 'make create-dmg' first"; \
		exit 1; \
	fi
	@IDENTITY=$$(security find-identity -v -p codesigning | grep "Developer ID Application" | head -1 | awk -F'"' '{print $$2}'); \
	if [ -n "$$IDENTITY" ]; then \
		echo "Signing DMG with identity: $$IDENTITY"; \
		codesign --force --sign "$$IDENTITY" build/MYAPP.dmg; \
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
	@if [ ! -f "build/MYAPP.dmg" ]; then \
		echo "Error: DMG not found. Run 'make create-dmg' and 'make sign-dmg' first"; \
		exit 1; \
	fi
	xcrun notarytool submit build/MYAPP.dmg --apple-id "$(APPLE_ID)" --password "$(APPLE_ID_PASSWORD)" --team-id "$(TEAM_ID)" --wait
	xcrun stapler staple build/MYAPP.dmg
	@echo "✅ App notarized and stapled successfully"

.PHONY: release-darwin
release-darwin: build-darwin-obfuscated sign-app create-dmg sign-dmg notarize-app ## Complete macOS release build with signing and notarization
	@echo "🚀 macOS release build complete!"
	@echo "📦 Signed and notarized DMG: build/MYAPP.dmg"
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
	@echo "📦 Unsigned DMG: build/MYAPP.dmg"
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