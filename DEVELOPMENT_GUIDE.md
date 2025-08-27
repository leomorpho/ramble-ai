# RambleAI Development Guide

## Quick Start with Remote AI Backend

### Option 1: Full Stack Development (Recommended)
```bash
# 1. Add your API keys to pb-be/pb/.env
cp pb-be/pb/.env.example pb-be/pb/.env
# Edit .env with your actual API keys

# 2. Start both backend and frontend
make dev-with-backend
```

This automatically:
- ✅ Starts PocketBase backend on http://localhost:8090
- ✅ Starts Wails app with `USE_REMOTE_AI_BACKEND=true`
- ✅ Enables remote AI processing for all features
- ✅ Cleans up backend process when you exit

### Option 2: Backend Only (for API testing)
```bash
# Start just the PocketBase backend
make dev-backend-only
```

Access:
- **Admin UI**: http://localhost:8090/_/
- **API Endpoints**: http://localhost:8090/api/

### Option 3: Traditional Local Development
```bash
# Start Wails app only (uses local API keys)
make dev
```

## Environment Configuration

### Required API Keys
Add these to `pb-be/pb/.env`:
```bash
OPENROUTER_API_KEY=your-openrouter-key-here
OPENAI_API_KEY=your-openai-key-here
```

### Environment Variables for Remote Mode
```bash
USE_REMOTE_AI_BACKEND=true              # Enable remote backend
REMOTE_AI_BACKEND_URL=http://localhost:8090  # Backend URL (optional if default)
```

## How It Works

### Local Mode (Default)
```
RambleAI App ──────────────────> OpenRouter/OpenAI APIs
```

### Remote Mode (`USE_REMOTE_AI_BACKEND=true`)
```
RambleAI App ──> PocketBase Backend ──> OpenRouter/OpenAI APIs
                     ↑
                 Ramble AI API Key
                 (user configured)
```

## API Endpoints

### Text Processing
- **Endpoint**: `POST /api/ai/process-text`
- **Auth**: Bearer token (Ramble AI API key)
- **Used for**: Highlight suggestions, content reordering, chat

### Audio Processing  
- **Endpoint**: `POST /api/ai/process-audio`
- **Auth**: Bearer token (Ramble AI API key) 
- **Used for**: Video transcription via Whisper

### API Key Management
- **Generate**: `POST /api/generate-api-key` (requires user auth)
- **Format**: `ra-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`

## Development Workflow

1. **Setup API Keys**: Add your OpenRouter and OpenAI keys to `pb-be/pb/.env`
2. **Start Development**: Run `make dev-with-backend`
3. **Configure User API Key**: 
   - Open app → Settings → Remote AI Backend
   - Generate API key via PocketBase admin or API
   - Enter the `ra-xxx` key in the app
4. **Test Features**:
   - Highlight suggestions → uses text endpoint
   - Video transcription → uses audio endpoint

## Troubleshooting

### Backend Won't Start
- Check `pb-be/pb/.env` exists with valid API keys
- Ensure port 8090 is free: `lsof -ti:8090`

### API Key Issues
- Generate new key in PocketBase admin: http://localhost:8090/_/
- Format must be `ra-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`
- Check key is active in `api_keys` collection

### Remote AI Not Working
- Verify `USE_REMOTE_AI_BACKEND=true` is set
- Check backend URL in app settings matches PocketBase
- Check PocketBase logs for authentication errors

## Useful Commands

```bash
make dev-with-backend    # Full stack with remote AI
make dev-backend-only    # Backend only  
make stop-backend        # Stop background backend
make help                # Show all available commands
```

## File Structure

```
ramble-ai/
├── pb-be/                    # PocketBase backend
│   ├── pb/
│   │   ├── internal/ai/      # AI endpoints  
│   │   ├── main.go           # Server entry point
│   │   └── .env              # API keys (git ignored)
│   └── sk/                   # SvelteKit frontend (optional)
├── goapp/ai/                 # AI service factory
│   ├── factory.go            # Service selection logic
│   ├── remote.go             # Remote backend client
│   └── local.go              # Local API client
└── frontend/                 # Wails frontend
    └── src/lib/components/settings/
        └── RemoteAIConfig.svelte  # UI for backend config
```

Happy coding! 🚀