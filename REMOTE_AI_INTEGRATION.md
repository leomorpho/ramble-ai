# Remote AI Backend Integration

## Overview
RambleAI can now use a PocketBase backend server to handle all LLM API calls instead of making them directly from the desktop app. This allows for centralized API key management, usage tracking, and subscription control.

## Architecture

```
┌─────────────────┐          ┌──────────────────┐          ┌─────────────┐
│  RambleAI App   │  ──────> │ PocketBase Backend│  ──────> │  AI APIs    │
│   (Wails)       │   HTTP   │  (localhost:8090) │   HTTP   │ (OpenRouter,│
│                 │          │                  │          │  OpenAI)    │
└─────────────────┘          └──────────────────┘          └─────────────┘
     ↑                              ↑
     │                              │
     └── Ramble AI API Key ────────┘
```

## Features

### Two API Endpoints

1. **Text Processing** (`/api/ai/process-text`)
   - Used for: Highlight suggestions, reordering, chat features
   - Calls: OpenRouter API (Claude, GPT, etc.)
   - Request: SystemPrompt, UserPrompt, Model, TaskType
   - Response: Content, TaskType, TokensUsed

2. **Audio Processing** (`/api/ai/process-audio`)
   - Used for: Video transcription
   - Calls: OpenAI Whisper API
   - Request: Base64 encoded audio data, filename
   - Response: Transcript, Duration, Language, Words, Segments

### Authentication
- Uses Ramble AI API keys (format: `ra-xxxxx`)
- Keys are hashed and stored in PocketBase `api_keys` collection
- Bearer token authentication on all AI endpoints

## Configuration

### Environment Variables

#### For PocketBase Backend (`pb-be/pb/.env`):
```bash
OPENROUTER_API_KEY=your-openrouter-api-key
OPENAI_API_KEY=your-openai-api-key
```

#### For RambleAI App:
```bash
USE_REMOTE_AI_BACKEND=true              # Enable remote backend mode
REMOTE_AI_BACKEND_URL=http://localhost:8090  # Backend URL
```

### User Configuration
Users must enter their Ramble AI API key in Settings → Remote AI Backend

## Running Locally

### Terminal 1: Start PocketBase Backend
```bash
cd pb-be/pb
go run main.go serve --dev --http 0.0.0.0:8090
```

### Terminal 2: Start RambleAI with Remote Backend
```bash
USE_REMOTE_AI_BACKEND=true \
REMOTE_AI_BACKEND_URL=http://localhost:8090 \
wails dev
```

## Implementation Details

### Service Factory Pattern
- `AIServiceFactory` checks `USE_REMOTE_AI_BACKEND` environment variable
- If `true`: Creates `RemoteAIService` that calls PocketBase
- If `false`: Creates `LocalAIService` that calls AI APIs directly

### Files Modified
1. **PocketBase Backend:**
   - `/pb-be/pb/internal/ai/endpoints.go` - Added ProcessAudioHandler
   - `/pb-be/pb/main.go` - Registered audio endpoint route

2. **RambleAI App:**
   - `/goapp/ai/factory.go` - Added env var support
   - `/goapp/ai/remote.go` - Implemented ProcessAudio method
   - `/goapp/ai/local.go` - Local AI service implementation
   - `/goapp/ai/interface.go` - Common interface

3. **Frontend:**
   - `/frontend/src/lib/components/settings/RemoteAIConfig.svelte` - UI for API key

## Security Considerations
- API keys are never exposed to frontend
- All AI requests authenticated with Ramble AI keys
- Audio files sent as base64 (consider multipart for large files)
- User subscription status checked on each request

## Testing
Run `./test-remote-ai.sh` for setup instructions and testing guide.