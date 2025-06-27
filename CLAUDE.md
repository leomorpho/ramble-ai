# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Wails v2 application using Go as the backend and SvelteKit with the static adapter as the frontend. Wails allows building desktop applications using web technologies, with Go providing the backend API and SvelteKit handling the UI.

## Architecture

- **Backend (Go)**: Located in root directory
  - `main.go`: Application entry point, embeds frontend assets from `frontend/build`
  - `app.go`: Application struct with methods exposed to frontend via Wails bindings
  - `wails.json`: Wails configuration file defining build commands and settings

- **Frontend (SvelteKit)**: Located in `frontend/` directory  
  - Uses `@sveltejs/adapter-static` to generate static files for embedding
  - Wails JS bindings available at `$lib/wailsjs/` (auto-generated)
  - Routes in `frontend/src/routes/`
  - Assets in `frontend/src/lib/assets/`

## Common Commands

### Development
```bash
# Start development server with hot reload
wails dev

# Frontend development (if needed separately)
cd frontend && pnpm dev
```

### Building
```bash
# Build for production
wails build

# Frontend build only
cd frontend && pnpm build
```

### Frontend Package Management
```bash
# Install dependencies (use pnpm by default)
cd frontend && pnpm install

# Type checking
cd frontend && pnpm check

# Type checking with watch mode
cd frontend && pnpm check:watch
```

### Testing
No test configuration found - add test scripts to `frontend/package.json` if needed.

## Key Configuration

- **Wails Config** (`wails.json`): 
  - Frontend install: `npm install` (consider updating to `pnpm install`)
  - Frontend build: `npm run build --base=./` 
  - Wails JS directory: `./frontend/src/lib`

- **SvelteKit Config** (`frontend/svelte.config.js`):
  - Uses static adapter with `index.html` fallback
  - Builds to `frontend/build/` for Wails embedding

- **Go Module**: Uses Wails v2.10.1 with Go 1.22.0+

## Frontend-Backend Communication

- Go methods in `app.go` are automatically exposed to frontend
- Access via `$lib/wailsjs/go/main/App.js` (auto-generated)
- Example: `Greet(name)` function returns a promise

## UI Components

- **shadcn-svelte**: Always use https://shadcn-svelte.com/docs/components components first for UI elements
- Components are located in `frontend/src/lib/components/ui/`
- Import from `$lib/components/ui/component-name`

## Important Notes

- Frontend builds to `frontend/build/` directory (not `dist/`)
- Main.go embeds `all:frontend/build` - ensure this matches build output
- Use `pnpm` for package management instead of `npm`
- Static adapter disables SSR for proper Wails integration