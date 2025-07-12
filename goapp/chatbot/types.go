package chatbot

import (
	"context"
	"time"
	"MYAPP/ent"
	"MYAPP/goapp/highlights"
)

// ChatMessage represents a single message in a conversation
type ChatMessage struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // "user", "assistant", "system", "error"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Hidden    string    `json:"-"` // Hidden context not sent to frontend
}

// ChatSession represents a conversation session
type ChatSession struct {
	ID        string        `json:"id"`
	SessionID string        `json:"sessionId"`
	ProjectID int           `json:"projectId"`
	EndpointID string       `json:"endpointId"`
	Messages  []ChatMessage `json:"messages"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

// ChatRequest represents a request to send a message
type ChatRequest struct {
	ProjectID          int                    `json:"projectId"`
	EndpointID         string                 `json:"endpointId"`
	Message            string                 `json:"message"`
	SessionID          string                 `json:"sessionId,omitempty"`
	ContextData        map[string]interface{} `json:"contextData"`
	Model              string                 `json:"model"`
	EnableFunctionCalls bool                   `json:"enableFunctionCalls,omitempty"`
	Mode               string                 `json:"mode,omitempty"` // "chat" or "reorder"
}

// ChatResponse represents the response from sending a message
type ChatResponse struct {
	SessionID         string                   `json:"sessionId"`
	MessageID         string                   `json:"messageId"`
	Message           string                   `json:"message"`
	Model             string                   `json:"model,omitempty"`
	Success           bool                     `json:"success"`
	Error             string                   `json:"error,omitempty"`
	FunctionResults   []FunctionExecutionResult `json:"functionResults,omitempty"`
	ActionsAvailable  []string                 `json:"actionsAvailable,omitempty"`
}

// FunctionExecutionResult represents the result of executing a function
type FunctionExecutionResult struct {
	FunctionName string      `json:"functionName"`
	Success      bool        `json:"success"`
	Result       interface{} `json:"result,omitempty"`
	Error        string      `json:"error,omitempty"`
	Message      string      `json:"message,omitempty"`
}

// ChatHistoryResponse represents chat history for a project/endpoint
type ChatHistoryResponse struct {
	SessionID string        `json:"sessionId"`
	Messages  []ChatMessage `json:"messages"`
}

// FunctionDefinition represents a function that can be called by the LLM
type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// FunctionExecutor represents a function that can be executed
type FunctionExecutor func(args map[string]interface{}, projectID int, service *ChatbotService) (interface{}, error)

// UpdateOrderFunc is a function type for updating highlight order
type UpdateOrderFunc func(projectID int, order []interface{}) error

// ChatbotService provides chatbot functionality
type ChatbotService struct {
	client           *ent.Client
	ctx              context.Context
	functionRegistry map[string]FunctionExecutor
	functionDefs     []FunctionDefinition
	highlightService *highlights.HighlightService
	aiService        *highlights.AIService
	updateOrderFunc  UpdateOrderFunc
}