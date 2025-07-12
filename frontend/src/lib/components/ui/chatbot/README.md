# Chatbot Components

This directory contains the AI chatbot components for the video preparation application.

## Components

### `AIChatbot.svelte`
Main chatbot component that can be used in different configurations:
- **Floating**: A floating brain button that opens a chat sheet
- **Inline**: An inline button for replacing existing AI components
- **Sheet**: A sheet-only version without trigger button

### `ChatInterface.svelte`
The main chat interface that handles:
- Message history loading
- Message sending/receiving
- Settings panel
- Header with controls

### `MessageList.svelte`
Displays the conversation messages with auto-scrolling.

### `ChatMessage.svelte`
Individual message component with support for:
- User messages
- AI responses
- Error messages
- System messages
- Basic markdown rendering
- Copy to clipboard

### `MessageInput.svelte`
Auto-resizing textarea for user input with send button.

### `ChatSettings.svelte`
Settings panel for configuring AI model and other options.

## Configuration

See `frontend/src/lib/constants/chatbot.js` for:
- Endpoint definitions
- Model configurations
- Message types
- Position options

## Usage Examples

### Floating Brain Button
```svelte
<script>
  import { AIChatbot } from "$lib/components/ui/chatbot";
  import { CHATBOT_ENDPOINTS } from "$lib/constants/chatbot.js";
</script>

<AIChatbot 
  endpointId={CHATBOT_ENDPOINTS.HIGHLIGHT_ORDERING}
  projectId={currentProjectId}
  contextData={{ highlights, order }}
/>
```

### Inline Replacement for AISettings
```svelte
<AIChatbot 
  endpointId={CHATBOT_ENDPOINTS.HIGHLIGHT_SUGGESTIONS}
  projectId={currentProjectId}
  contextData={{ highlights, transcription }}
  position="inline"
  buttonText="AI Suggestions"
/>
```

## Backend Integration

The chatbot expects these Wails endpoints:
- `SendChatMessage(request)` - Send a message and get AI response
- `GetChatHistory(projectId, endpointId)` - Load conversation history  
- `ClearChatHistory(projectId, endpointId)` - Clear conversation

## Features

- ✅ Organized component structure
- ✅ Multiple positioning options (floating, inline, sheet-only)
- ✅ Configurable endpoints and contexts
- ✅ Message persistence with hidden context
- ✅ Auto-resizing input
- ✅ Copy to clipboard
- ✅ Model selection
- ✅ Error handling
- ✅ Loading states
- ✅ Responsive design
- ✅ Accessibility support
- ✅ **Svelte 5 Runes Mode Compatible** - Uses modern Svelte 5 patterns

## Important: Svelte 5 Compatibility

This chatbot system is built specifically for **Svelte 5 with runes mode**. Key patterns used:

- `$state()` for reactive state
- `$derived()` for computed values  
- `$effect()` for side effects
- `$props()` for component props
- Direct component prop passing (no `asChild`/`let:builder` patterns)

The components are **not compatible** with legacy Svelte patterns.