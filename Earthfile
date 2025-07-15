VERSION 0.8

# Use a multi-stage approach for better caching
FROM golang:1.23-alpine
WORKDIR /app

# Install base system dependencies
deps:
    RUN apk add --no-cache \
        build-base \
        git \
        nodejs \
        npm \
        sqlite \
        sqlite-dev
    RUN npm install -g pnpm

# Install and cache Go dependencies
go-deps:
    FROM +deps
    COPY go.mod go.sum ./
    RUN go mod download
    SAVE ARTIFACT go.mod AS LOCAL go.mod.cached
    SAVE ARTIFACT go.sum AS LOCAL go.sum.cached

# Install and cache frontend dependencies  
frontend-deps:
    FROM +deps
    COPY frontend/package.json ./frontend/
    COPY --if-exists frontend/pnpm-lock.yaml ./frontend/
    WORKDIR /app/frontend
    RUN pnpm install --no-frozen-lockfile
    SAVE ARTIFACT node_modules AS LOCAL frontend/node_modules.cached

# Copy all source code
src:
    FROM +deps
    # Copy cached Go deps
    COPY +go-deps/go.mod +go-deps/go.sum ./
    RUN go mod download
    
    # Copy cached frontend deps
    COPY +frontend-deps/node_modules ./frontend/node_modules
    
    # Copy source files
    COPY . .
    
    # Build frontend for embedding
    WORKDIR /app/frontend
    RUN pnpm run build
    WORKDIR /app
    
    # Generate Ent code
    RUN go generate ./ent

# Run Go tests with coverage
test-go:
    FROM +src
    RUN --no-cache \
        go test $(go list ./... | grep -v '/exports') -v -short -race -coverprofile=coverage.out -covermode=atomic && \
        go tool cover -func=coverage.out
    SAVE ARTIFACT coverage.out AS LOCAL coverage.out

# Run frontend tests
test-frontend:
    FROM +src
    WORKDIR /app/frontend
    RUN --no-cache pnpm run test:run -- --reporter=verbose

# Run all tests
test:
    BUILD +test-go
    BUILD +test-frontend

# Build frontend for production
build-frontend:
    FROM +src
    WORKDIR /app/frontend
    RUN pnpm run build
    SAVE ARTIFACT build AS LOCAL frontend/build

# Build Go binary (without Wails for CI)
build-go:
    FROM +src
    # Build a test binary to ensure compilation works
    RUN go build -v ./...
    
# Full CI pipeline
ci:
    BUILD +test
    BUILD +build-frontend
    BUILD +build-go

# Development helpers
test-watch:
    FROM +src
    WORKDIR /app
    RUN --interactive --push \
        echo "Running tests in watch mode..." && \
        cd frontend && pnpm run test:watch