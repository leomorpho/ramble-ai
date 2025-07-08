VERSION 0.8

# Use a multi-stage approach for better caching
FROM golang:1.22-alpine
WORKDIR /app

# Install base system dependencies
base:
    RUN apk add --no-cache \
        build-base \
        git \
        nodejs \
        npm \
        sqlite \
        sqlite-dev

# Install and cache Go dependencies
go-deps:
    FROM +base
    COPY go.mod go.sum ./
    RUN go mod download
    SAVE ARTIFACT go.mod AS LOCAL go.mod.cached
    SAVE ARTIFACT go.sum AS LOCAL go.sum.cached

# Install and cache frontend dependencies  
frontend-deps:
    FROM +base
    COPY frontend/package*.json ./frontend/
    WORKDIR /app/frontend
    RUN npm ci
    SAVE ARTIFACT node_modules AS LOCAL frontend/node_modules.cached

# Copy all source code
src:
    FROM +base
    # Copy cached Go deps
    COPY +go-deps/go.mod +go-deps/go.sum ./
    RUN go mod download
    
    # Copy cached frontend deps
    COPY +frontend-deps/node_modules ./frontend/node_modules
    
    # Copy source files
    COPY . .
    
    # Generate Ent code
    RUN go generate ./ent

# Run Go tests with coverage
test-go:
    FROM +src
    RUN --no-cache \
        go test ./... -v -short -race -coverprofile=coverage.out -covermode=atomic && \
        go tool cover -func=coverage.out
    SAVE ARTIFACT coverage.out AS LOCAL coverage.out

# Run frontend tests
test-frontend:
    FROM +src
    WORKDIR /app/frontend
    RUN --no-cache npm run test:run -- --reporter=verbose
    
# Run Go lint checks
lint-go:
    FROM +src
    RUN gofmt -l . | tee /tmp/gofmt.out && \
        test ! -s /tmp/gofmt.out || \
        (echo "Go files are not formatted:" && cat /tmp/gofmt.out && exit 1)
    
    # Install golangci-lint
    RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin v1.54.2
    RUN golangci-lint run ./...

# Run frontend type checking
check-frontend:
    FROM +src
    WORKDIR /app/frontend
    RUN npm run check

# Run all tests
test:
    BUILD +test-go
    BUILD +test-frontend

# Run all checks
lint:
    BUILD +lint-go
    BUILD +check-frontend

# Build frontend for production
build-frontend:
    FROM +src
    WORKDIR /app/frontend
    RUN npm run build
    SAVE ARTIFACT build AS LOCAL frontend/build

# Build Go binary (without Wails for CI)
build-go:
    FROM +src
    # Build a test binary to ensure compilation works
    RUN go build -v ./...
    
# Full CI pipeline
ci:
    BUILD +lint
    BUILD +test
    BUILD +build-frontend
    BUILD +build-go

# Development helpers
test-watch:
    FROM +src
    WORKDIR /app
    RUN --interactive --push \
        echo "Running tests in watch mode..." && \
        cd frontend && npm run test:watch