# RambleAI - AI-Powered Video Preprocessing for Talking Head Content

[![Build and Release](https://github.com/leoaudibert/ramble-ai/actions/workflows/build.yml/badge.svg)](https://github.com/leoaudibert/ramble-ai/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/leoaudibert/ramble-ai)](https://goreportcard.com/report/github.com/leoaudibert/ramble-ai)

**Transform hours of raw talking head footage into polished scripts in minutes.**

RambleAI is an AI-powered desktop application that preprocesses talking head videos, automatically selecting the best clips and reordering them into coherent scripts. Built with [Wails](https://wails.io), Go, and SvelteKit, it saves content creators 60-80% of their editing time.

## Why RambleAI?

RambleAI is NOT a video editor - it's a preprocessing tool that works WITH your favorite editor (Premiere, Final Cut, DaVinci Resolve). It handles the tedious part of editing talking head content: finding the good takes, removing the rambling, and organizing everything into a coherent narrative.

### Key Benefits

- ⚡ **60-80% Time Savings**: What used to take 8 hours now takes 2-3 hours
- 🎯 **Smart Clip Selection**: AI identifies the best parts of your videos automatically
- 🧠 **AI Script Reordering**: Transform disorganized clips into coherent scripts
- 🔒 **100% Private**: All processing happens locally on your machine
- 🔄 **Works with Any Editor**: Export to your preferred video editing software

## Features

- 🎬 **Video Management**: Import and organize talking head footage
- ✂️ **Smart Highlight Extraction**: AI-powered clip selection with quality scoring  
- 🤖 **Flexible AI Integration**: Use any AI model via OpenRouter (bring your own API keys)
- 📝 **Accurate Transcription**: Word-level timestamps for precise editing
- 🎯 **Visual Timeline Editor**: Drag-to-resize highlight editing
- 📤 **Multiple Export Options**: Individual clips or stitched compilations
- 🌓 **Modern UI**: Dark/Light theme with native desktop performance
- 🚀 **Fast Processing**: Optimized for quick turnaround times

## Quick Start

### Download & Install
1. Download the latest release from [GitHub Releases](https://github.com/leoaudibert/ramble-ai/releases)
2. Install the application:
   - **macOS**: Open the .dmg and drag RambleAI to Applications
   - **Windows**: Run the installer
   - **Linux**: Extract and run the AppImage

### Your First Project
1. Launch RambleAI
2. Create a new project
3. Import your talking head video
4. Let AI analyze and transcribe
5. Review AI-suggested clips
6. Reorder clips into your script
7. Export to your video editor

## How It Works

### The RambleAI Workflow

1. **📤 Upload Footage**: Import your raw talking head video. AI analyzes speech patterns and content quality.

2. **🎯 Smart Selection**: AI identifies the best clips based on clarity, coherence, and narrative value.

3. **🧠 Script Reordering**: AI reorders selected clips into a coherent, engaging script using your choice of LLM.

4. **📦 Export & Handoff**: Export optimized scripts and clips ready for your favorite video editor.

## System Requirements

### Minimum Requirements
- macOS 10.15+ / Windows 10+ / Linux (Ubuntu 20.04+)
- 8GB RAM
- 2GB free disk space
- Internet connection (for AI features only)

### Recommended
- 16GB RAM for optimal performance
- SSD for faster video processing
- Dedicated GPU for smoother playback

## Getting Started

### Prerequisites for Development

- Go 1.22+
- Node.js 20+
- pnpm 9+
- FFmpeg (for video processing)
- Wails CLI

### Installation

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Clone the repository
git clone https://github.com/leoaudibert/ramble-ai.git
cd ramble-ai

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

### Local Development & Testing

To test the complete application locally with both PocketBase backend and Wails desktop app:

```bash
# Terminal 1: Start PocketBase backend (reset database)
cd pb-be
make pb NUKE=1

# Terminal 2: Start Wails desktop app  
make dev
```

This setup provides:
- ✅ **Clean Database**: `NUKE=1` resets PocketBase database completely
- ✅ **Consistent API Keys**: Both services seed the same development API key  
- ✅ **Remote Backend Mode**: Wails app connects to local PocketBase (localhost:8090)
- ✅ **Full Feature Testing**: Audio transcription, user management, subscriptions

**Development API Key**: Both services automatically use `ra-dev-12345678901234567890123456789012`

**Ports**:
- PocketBase Backend: `http://localhost:8090`
- SvelteKit Frontend: `http://localhost:5174` 
- Wails Desktop App: Development window

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
- 🔐 Automated macOS code signing and notarization

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

### macOS Code Signing and Notarization

The project automatically signs and notarizes macOS builds using GitHub Actions. This requires the following GitHub secrets/variables:

#### Required Secrets
Set these in your GitHub repository settings under Settings → Secrets and variables → Actions:

- `APPLE_DEVELOPER_CERTIFICATE_P12_BASE64`: Your Apple Developer certificate exported as .p12 and base64 encoded
- `APPLE_DEVELOPER_CERTIFICATE_PASSWORD`: Password for your .p12 certificate  
- `APPLE_ID_PASSWORD`: App-specific password for your Apple ID (generate at [appleid.apple.com](https://appleid.apple.com))

#### Required Variables (or Secrets)
Set these as either repository variables or secrets:

- `APPLE_ID`: Your Apple ID email address
- `APPLE_TEAM_ID`: Your Apple Developer Team ID (10-character string like `ABCDE12345`)

#### Creating the Certificate File
1. Export your Apple Developer certificate from Keychain Access as a .p12 file
2. Convert to base64: `base64 YourCertificate.p12 | pbcopy`
3. Paste the result into the `APPLE_DEVELOPER_CERTIFICATE_P12_BASE64` secret

#### Finding Your Team ID
Run this command on macOS to find your Team ID:
```bash
security find-identity -v -p codesigning
```

The signed and notarized app will be available for download as a build artifact and included in GitHub releases.

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

## AI Integration

### Your AI, Your Choice

RambleAI works with any AI model through OpenRouter, giving you complete flexibility:

- 🔑 **Bring Your Own API Keys**: No subscription lock-in, control your AI costs
- 🤖 **Any LLM Model**: Choose from GPT-4, Claude, Llama, or any OpenRouter-supported model
- ⚡ **Smart Defaults**: Optimized prompts work great out of the box
- 🎯 **Customizable**: Tweak AI behavior to match your content style

### Supported AI Providers
- OpenRouter (recommended - access to 100+ models)
- OpenAI (direct integration)
- Local LLMs (coming soon)

## Privacy & Security

🔒 **Your content stays private**: 
- All video processing happens locally on your machine
- No footage is uploaded to our servers
- AI features use your own API keys
- Complete control over your data

## License

License pending - please check back soon for licensing information.

## Acknowledgments

- [Wails](https://wails.io) - Desktop application framework
- [Ent](https://entgo.io) - Entity framework for Go
- [SvelteKit](https://kit.svelte.dev) - Frontend framework
- [shadcn-svelte](https://shadcn-svelte.com) - UI components
- [Earthly](https://earthly.dev) - Build automation

## Support

For bugs and feature requests, please [open an issue](https://github.com/leoaudibert/ramble-ai/issues).

For questions and discussions, use [GitHub Discussions](https://github.com/leoaudibert/ramble-ai/discussions).