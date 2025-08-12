# RambleAI - Video Editor

[![CI](https://github.com/leomorpho/vidking-wails/actions/workflows/ci.yml/badge.svg)](https://github.com/leomorpho/vidking-wails/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/leomorpho/vidking-wails)](https://goreportcard.com/report/github.com/leomorpho/vidking-wails)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A powerful desktop video editor built with [Wails](https://wails.io), Go, and SvelteKit.

## Features

- 🎬 Video clip management and organization
- ✂️ Highlight extraction with timestamp precision  
- 🤖 AI-powered highlight suggestions
- 📝 Automatic transcription with word-level timestamps
- 🎯 Drag-to-resize highlight editing
- 📤 Export highlights as individual clips or stitched compilations
- 🌓 Dark/Light theme support
- ⚡ Native desktop performance

## Getting Started

### Prerequisites

- Go 1.22+
- Node.js 20+
- FFmpeg (for video processing)
- Wails CLI

### Installation

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Clone the repository
git clone https://github.com/leomorpho/vidking-wails.git
cd vidking-wails/RambleAI

# Install dependencies
make setup

# Run in development mode
make dev
```

### Building

```bash
# Build for production
make build

# Build from scratch (clean build)
make full-build
```

## Testing

The project has comprehensive test coverage with multiple testing approaches:

```bash
# Run all tests
make test

# Run specific test suites
make test-go          # Go tests only
make test-frontend    # Frontend tests only
make test-coverage    # Generate coverage report

# Watch mode for frontend development
make test-watch
```

For detailed testing documentation, see [docs/TESTING.md](docs/TESTING.md).

## Development

### Project Structure

```
RambleAI/
├── app.go              # Main application logic
├── goapp/              # Go backend modules
│   ├── exports/        # Video export functionality
│   ├── highlights/     # Highlight management
│   ├── projects/       # Project management
│   └── settings/       # Application settings
├── frontend/           # SvelteKit frontend
│   ├── src/
│   │   ├── lib/        # Components and utilities
│   │   └── routes/     # Application pages
│   └── static/         # Static assets
├── ent/                # Database ORM (Ent)
└── Makefile           # Build and development commands
```

### Key Technologies

- **Backend**: Go, Ent ORM, SQLite
- **Frontend**: SvelteKit (Svelte 5), Tailwind CSS, shadcn-svelte
- **Build**: Wails, Vite, Earthly
- **Testing**: Vitest, Go testing, Earthly CI

### Database Management

```bash
# Create new entity
make new-entity name=EntityName

# Generate Ent code
make generate

# Reset database (WARNING: deletes all data)
make db-reset
```

## CI/CD

The project uses GitHub Actions with Earthly for continuous integration:

- ✅ Automated testing on every push
- 🔍 Code linting and formatting checks
- 📊 Test coverage reporting
- 🏗️ Multi-platform builds (Linux, macOS, Windows)

### Running CI Locally with Earthly

```bash
# Install Earthly
brew install earthly/earthly/earthly

# Run full CI pipeline locally
earthly +ci

# Run specific targets
earthly +test
earthly +lint
earthly +build-frontend
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes and add tests
4. Ensure all tests pass (`make test`)
5. Commit your changes (commits are validated with conventional format)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Git Hooks

The project uses Lefthook for automated pre-commit and pre-push checks:

```bash
# Install git hooks
lefthook install
```

This ensures:
- Tests pass before commits
- Code is properly formatted
- Commit messages follow conventional format

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Wails](https://wails.io) - Desktop application framework
- [Ent](https://entgo.io) - Entity framework for Go
- [SvelteKit](https://kit.svelte.dev) - Frontend framework
- [shadcn-svelte](https://shadcn-svelte.com) - UI components
- [Earthly](https://earthly.dev) - Build automation

## Support

For bugs and feature requests, please [open an issue](https://github.com/leomorpho/vidking-wails/issues).

For questions and discussions, use [GitHub Discussions](https://github.com/leomorpho/vidking-wails/discussions).