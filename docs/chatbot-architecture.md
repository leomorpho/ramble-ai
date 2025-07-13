# Chatbot AI Architecture Documentation

## Overview

This document describes the sophisticated AI chatbot architecture implemented in the Go backend. The system uses a modern three-stage conversational flow with intelligent context management and function calling capabilities.

## Architecture Components

### High-Level Flow

```
User Message → Conversation Agent → Go Preparer → Executor Agent → Result
     ↑               ↓                    ↓             ↓          ↓
Database ←——— Message Persistence ←——— MCP Functions ←————————————┘
```

## Three-Stage Conversational Flow

### Stage 1: Conversation Agent (`conversation_agent.go`)

**Purpose**: Natural language conversation and intent understanding

**Responsibilities**:
- Engage in natural conversation with users
- Ask clarifying questions one at a time
- Understand user intent through LLM reasoning (not keyword matching)
- Wait for user confirmation before database-changing operations
- Output structured `ConversationSummary` when ready to proceed

**Key Features**:
- **Interactive Flow**: Asks questions incrementally for natural conversation
- **Context Preservation**: Includes full chat history with intelligent trimming
- **Intent Understanding**: Uses YouTube expertise to interpret user requests
- **Confirmation Required**: Explains what will happen before proceeding

**Input**: User message + conversation history
**Output**: Either conversational response OR `ConversationSummary` JSON

```go
type ConversationSummary struct {
    Intent                string   `json:"intent"`                // "reorder", "analyze", etc.
    UserWantsCurrentOrder bool     `json:"userWantsCurrentOrder"` 
    OptimizationGoals     []string `json:"optimizationGoals"`     
    SpecificRequests      []string `json:"specificRequests"`      
    UserContext           string   `json:"userContext"`           
    Confirmed             bool     `json:"confirmed"`             
}
```

### Stage 2: Go Preparer (`go_preparer.go`)

**Purpose**: Data gathering and prompt construction

**Responsibilities**:
- Call MCP functions to gather necessary data
- Build complete, context-rich prompts for the executor
- Append all required information (highlights, current order, etc.)
- Create intent-specific templates and output format requirements

**Key Features**:
- **Pure Go Implementation**: No LLM calls, just data operations
- **MCP Function Calls**: Transparent to user, gathers all needed context
- **Intent-Specific Templates**: Different prompts for reorder vs analyze
- **Complete Context**: Provides everything executor needs in one prompt

**Functions**:
```go
func (s *ChatbotService) PrepareExecutorPrompt(summary *ConversationSummary, projectID int) (string, error)
```

### Stage 3: Executor Agent (`executor_agent.go`)

**Purpose**: Task execution with structured output

**Responsibilities**:
- Receive complete prompt from preparer
- Execute the specific task (reorder, analyze, etc.)
- Return structured JSON results
- No MCP function calls - everything provided in prompt

**Key Features**:
- **Structured Input/Output**: Predefined JSON formats
- **Task-Specific**: Different behavior for different intents
- **No Function Calling**: Simplified execution model
- **Consistent Results**: Reliable JSON output format

## Context Management System

### Context Manager (`context_manager.go`)

**Purpose**: Intelligent conversation context preservation

**Key Components**:

#### 1. Token Counting
```go
type TokenCounter struct {
    modelLimits map[string]int // Model-specific context windows
}
```

**Supported Models**:
- Claude Sonnet 4: 200,000 tokens
- Claude 3.5 Sonnet: 200,000 tokens
- GPT-4o: 128,000 tokens
- Default: 32,000 tokens

#### 2. Context Window Management
```go
type ContextWindow struct {
    SystemPrompt    string                   
    Messages        []map[string]interface{} 
    TotalTokens     int                      
    TrimmedMessages int                      
    Summary         string                   
}
```

#### 3. Intelligent Context Trimming

**Sliding Window Approach**:
1. Keep system prompt (always)
2. Reserve tokens for response
3. Include recent messages that fit
4. Summarize older messages that don't fit
5. Add summary as system message if space allows

**Context Prioritization**:
- System prompt: Highest priority
- Recent messages: High priority  
- Current user message: Required
- Older messages: Summarized if space limited

## Data Flow and Persistence

### Message Persistence

**Storage**: PostgreSQL via Ent ORM
**Tables**: 
- `chat_sessions` - Session metadata
- `chat_messages` - Individual messages

**Persistence Flow**:
1. Message received → Create/find session
2. Process through three-stage flow
3. Persist user message and assistant response
4. Update session metadata

### Chat History Retrieval

```go
func (s *ChatbotService) GetChatHistory(projectID int, endpointID string) (*ChatHistoryResponse, error)
```

**Features**:
- Retrieves all messages for project/endpoint
- Ordered by timestamp
- Includes session metadata
- Used by context manager for intelligent trimming

## Function System (MCP Registry)

### MCP (Model Context Protocol) Functions

**Available Functions**:
- `reorder_highlights` - Reorder video highlights
- `get_current_order` - Get current highlight arrangement  
- `analyze_highlights` - Analyze content structure
- `apply_ai_suggestion` - Apply cached AI suggestions
- `reset_to_original` - Reset to original order

**Function Execution Flow**:
1. Preparer calls MCP functions to gather data
2. Results appended to executor prompt
3. Executor receives complete context
4. Service layer applies results to database

### Function Registry

```go
type FunctionExecutor func(args map[string]interface{}, projectID int, service *ChatbotService) (interface{}, error)
```

## Configuration and Models

### Model Configuration

**Conversation Agent**: `anthropic/claude-sonnet-4`
- Higher temperature (0.7) for natural conversation
- 2000 max tokens for responses

**Executor Agent**: Model specified in request
- Lower temperature for consistent results
- Structured JSON output required

### Endpoint Configuration

**Endpoints**: Each project can have multiple chatbot endpoints
**Session Management**: One session per project/endpoint combination
**Context Isolation**: Conversations isolated by endpoint

## Service Layer Integration

### Main Service (`service.go`)

**Core Methods**:
- `ProcessChatMessage()` - Main entry point
- `handleConversationPhase()` - Stage 1 processing
- `handleExecutionPhase()` - Stage 2+3 processing  
- `GetChatHistory()` - Retrieve conversation history
- `ClearChatHistory()` - Reset conversations

### Request/Response Types

```go
type ChatRequest struct {
    ProjectID  int    `json:"projectId"`
    EndpointID string `json:"endpointId"`  
    Message    string `json:"message"`
    Model      string `json:"model"`
    SessionID  string `json:"sessionId,omitempty"`
}

type ChatResponse struct {
    SessionID string `json:"sessionId"`
    MessageID string `json:"messageId"`
    Response  string `json:"response"`
    Success   bool   `json:"success"`
    Error     string `json:"error,omitempty"`
}
```

## Flow State Management

### Conversation Flow

```go
type ConversationFlow struct {
    Phase     ConversationPhase      // "conversation" or "execution"
    Context   map[string]interface{} // Flow-specific context
    SessionID string                 
}
```

**Phases**:
- `PhaseConversation` - Gathering user intent
- `PhaseExecution` - Executing confirmed intent

## Error Handling and Resilience

### Error Strategies

1. **API Failures**: Graceful degradation with user-friendly messages
2. **Context Overflow**: Intelligent trimming and summarization
3. **Function Failures**: Continue with available data
4. **Database Issues**: Maintain conversation state where possible

### Monitoring and Logging

**Context Usage Logging**:
```go
func (cm *ContextManager) LogContextUsage(model string, window *ContextWindow)
```

**Progress Broadcasting**: Real-time updates to frontend via WebSocket

## Security and Best Practices

### Input Validation
- Sanitize user input
- Validate JSON structures
- Check project/endpoint permissions

### API Key Management
- Secure API key storage
- Model-specific key validation
- Fallback handling for missing keys

### Rate Limiting
- Per-project rate limiting
- Context window size limits
- Response time monitoring

## Integration Examples

### Adding New Endpoints

1. Register endpoint in MCP registry
2. Define endpoint-specific functions
3. Create endpoint-specific prompts
4. Test conversation flow

### Adding New Models

1. Add model to `TokenCounter.modelLimits`
2. Test context window behavior
3. Adjust token estimates if needed
4. Update model selection logic

## Performance Considerations

### Context Optimization
- Efficient token counting
- Smart message trimming
- Conversation summarization
- Database query optimization

### Caching Strategies
- Session-level caching
- Function result caching  
- Model response caching
- Context window caching

## Debugging and Troubleshooting

### Common Issues

1. **Context Too Large**: Check token counting and trimming logic
2. **Intent Not Understood**: Review conversation agent prompts
3. **Function Failures**: Check MCP registry and function implementations
4. **Response Format**: Validate JSON output from executor

### Debug Tools

- Context usage logging
- Message persistence tracking
- Function execution logging
- Real-time progress monitoring

## Future Enhancements

### Planned Improvements

1. **Cross-Session Context**: Maintain context across session boundaries
2. **Project-Level Memory**: Share context between endpoints
3. **Advanced Summarization**: LLM-powered conversation summaries
4. **Context Relevance Scoring**: Smarter message prioritization
5. **Adaptive Context Windows**: Dynamic sizing based on conversation complexity

This architecture provides a robust, scalable foundation for sophisticated AI conversations while maintaining performance and user experience.