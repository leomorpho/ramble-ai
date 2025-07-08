# Testing Guide

This document describes the testing infrastructure for MYAPP.

## Table of Contents
- [Running Tests Locally](#running-tests-locally)
- [Using Make Commands](#using-make-commands)
- [Using Earthly](#using-earthly)
- [CI/CD Pipeline](#cicd-pipeline)
- [Git Hooks](#git-hooks)
- [Test Coverage](#test-coverage)

## Running Tests Locally

### Quick Start
```bash
# Run all tests
make test

# Run only Go tests
make test-go

# Run only frontend tests  
make test-frontend

# Run tests with coverage
make test-coverage
```

### Using Make Commands

Available test commands:
- `make test` - Run all tests (Go + Frontend)
- `make test-go` - Run Go tests only
- `make test-go-short` - Run Go tests in short mode
- `make test-go-verbose` - Run Go tests with verbose output
- `make test-frontend` - Run Frontend tests only
- `make test-frontend-ui` - Run Frontend tests with UI
- `make test-frontend-run` - Run Frontend tests once (CI mode)
- `make test-verbose` - Run all tests with verbose output
- `make test-coverage` - Run tests with coverage report
- `make test-watch` - Run frontend tests in watch mode

## Using Earthly

Earthly provides reproducible builds and tests across all environments.

### Install Earthly
```bash
# macOS
brew install earthly/earthly/earthly

# Linux
sudo /bin/sh -c 'wget https://github.com/earthly/earthly/releases/latest/download/earthly-linux-amd64 -O /usr/local/bin/earthly && chmod +x /usr/local/bin/earthly'
```

### Run Tests with Earthly
```bash
# Run all tests
earthly +test

# Run Go tests only
earthly +test-go

# Run frontend tests only
earthly +test-frontend

# Run full CI pipeline
earthly +ci

# Run linting
earthly +lint
```

## CI/CD Pipeline

The project uses GitHub Actions with Earthly for continuous integration.

### Pipeline Stages
1. **Linting** - Code formatting and style checks
2. **Testing** - All unit and integration tests
3. **Coverage** - Test coverage reporting
4. **Building** - Verify the application builds

### Supported Environments
- Ubuntu (primary)
- macOS
- Windows
- Go versions: 1.22, 1.23

## Git Hooks

The project uses Lefthook for git hooks automation.

### Pre-commit Hooks
- Runs tests for changed files only
- Executes in ~0.3 seconds
- Only runs relevant tests (Go or Frontend)

### Pre-push Hooks  
- Runs full test suite
- Ensures all tests pass before pushing

### Setup Git Hooks
```bash
# Install lefthook
brew install lefthook  # macOS
# or
npm install -g @evilmartians/lefthook  # All platforms

# Install hooks
lefthook install
```

## Test Coverage

### Viewing Coverage
```bash
# Generate coverage report
make test-coverage

# Open HTML coverage report
open coverage.html
```

### Current Coverage
- **goapp/exports**: ~73%
- **goapp/highlights**: ~31%
- **Frontend**: 95 tests passing

### Coverage Goals
- Maintain >70% coverage for critical packages
- All new features must include tests
- Bug fixes should include regression tests

## Test Structure

### Go Tests
```
goapp/
├── exports/
│   ├── exports_test.go           # Main tests
│   ├── exports_errors_test.go    # Error handling tests
│   ├── exports_individual_test.go # Individual export tests
│   ├── exports_stitched_test.go  # Stitched export tests
│   └── test_helpers.go           # Test utilities
└── highlights/
    ├── highlights_test.go        # Main tests
    ├── highlights_edge_cases_test.go
    └── ai_test.go               # AI integration tests
```

### Frontend Tests
```
frontend/src/lib/components/
├── TextHighlighter.behavior.test.js    # Business logic tests
├── TextHighlighter.drag.test.js        # Drag interaction tests
├── TextHighlighter.integration.test.js # Integration tests
└── TextHighlighter.utils.test.js       # Utility function tests
```

## Writing Tests

### Go Test Guidelines
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Use `testing.Short()` for long-running tests
- Always clean up resources with `defer`

### Frontend Test Guidelines
- Use Vitest with jsdom environment
- Test components in isolation
- Mock Wails runtime calls
- Focus on user interactions

## Troubleshooting

### Common Issues

**Tests hang on git push**
- Fixed by using `npm run test:run` instead of `npm test`
- Check lefthook.yml configuration

**Database errors in Go tests**
- Each test uses unique in-memory database
- Ensures proper schema migration

**Color system test failures**
- Tests now use CSS variables (e.g., `var(--highlight-1)`)
- Updated from hardcoded hex values

### Debug Commands
```bash
# Run specific Go test
go test -v -run TestName ./path/to/package

# Run specific frontend test
cd frontend && npm test -- TestName

# Check what Earthly will run
earthly --dry-run +test
```