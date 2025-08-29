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
  - Uses **Svelte 5** with runes mode (modern syntax)
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
```bash
# Run all tests
cd frontend && npm test

# Run tests with UI
cd frontend && npm run test:ui

# Run tests once (CI mode)
cd frontend && npm run test:run
```

The project uses **Vitest** with jsdom environment for testing. Test files should be named `*.test.js` or `*.spec.js` and placed alongside the code they test.

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

## UI Components & Styling

### Component Organization

- **shadcn-svelte**: Always use https://shadcn-svelte.com/docs/components components first for UI elements
- **shadcn-svelte components** are located in `frontend/src/lib/components/ui/`
- Import shadcn components from `$lib/components/ui/component-name`
- **NEVER add custom components to `frontend/src/lib/components/ui/`** - this directory is reserved exclusively for shadcn-svelte components
- **Custom components** should be placed in `frontend/src/lib/components/` (without the `ui/` subdirectory)

### Styling Guidelines

- **Use Tailwind CSS** for all styling - no custom CSS unless absolutely necessary
- **Theme-Aware Classes**: Always use CSS custom properties/variables that adapt to theme
  - Use `bg-background`, `text-foreground`, `border-border` etc.
  - **NEVER use `dark:` classes** - use universal classes that work with both themes
  - Use `text-muted-foreground` for secondary text
  - Use `bg-card`, `bg-secondary`, `bg-primary` for surfaces
  - Use `border-input` for form elements
- **Color System**: Stick to the predefined color tokens that support theme switching
- **Responsive Design**: Use responsive utilities (`md:`, `lg:`, etc.) for different screen sizes

### Svelte 5 Syntax (Runes Mode)

**IMPORTANT**: This project uses Svelte 5 with runes mode enabled. You MUST use Svelte 5 syntax exclusively:

- **State**: Use `let variable = $state(value)` instead of `let variable = value`
- **Reactive**: Use `let derived = $derived(expression)` instead of `$: derived = expression`
- **Effects**: Use `$effect(() => { ... })` instead of `$: { ... }`
- **Props**: Use `let { prop } = $props()` instead of `export let prop`
- **Events**: Use `onclick={handler}` instead of `on:click={handler}`
- **Class/Style**: Use `class={condition ? 'class' : ''}` instead of `class:name={condition}`
- **Binding**: `bind:value={variable}` remains the same

**NEVER use legacy Svelte syntax** like `$:`, `export let`, `on:click`, `class:name` - the build will fail in runes mode.

## Important Notes

- Frontend builds to `frontend/build/` directory (not `dist/`)
- Main.go embeds `all:frontend/build` - ensure this matches build output
- Use `pnpm` for package management instead of `npm`
- Static adapter disables SSR for proper Wails integration
- ALWAYS use the pocketbase javascript SDK when in a js/ts fe, except if it cannot be used for security reasons or custom endpoint or user asks you to not to
- when starting pocketbase for test on port 8090 or other, always kill it once you're done
## PocketBase Server Management

### ⚠️ CRITICAL: Safe PocketBase Process Killing

**NEVER use `lsof -ti:8090 | xargs kill -9`** - This kills ALL processes on port 8090 including Firefox and OrbStack!

**ONLY use these safe commands to kill PocketBase:**

```bash
# Method 1: Use Makefile command (RECOMMENDED - handles auto-restart processes)
make kill-pb

# Method 2: Kill specific Go processes
pkill -f "go run.*main.go serve"

# Method 3: More specific grep-based approach  
ps aux | grep "go run.*main.go serve" | grep -v grep | awk '{print $2}' | xargs -r kill -9

# Method 4: If using compiled PocketBase binary
pkill -f "pocketbase.*serve"

# Method 5: Kill parent processes too (for auto-restart scenarios)
pkill -f "go run.*main.go serve" && pkill -f "main.*serve"
```

**Before killing, verify what will be killed:**
```bash
# Check what processes will be affected BEFORE killing
ps aux | grep "go run.*main.go serve" | grep -v grep
```

### PocketBase Development Guidelines

- **Assume PocketBase is running** on port 8090 before starting tasks
- **Only kill/restart** when absolutely necessary for backend changes  
- **Ask before killing** if unsure whether restart is needed
- **Always verify** what processes will be killed before running kill commands
- **Use KillBash tool** when working with background processes started by Claude Code