package chatbot

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"MYAPP/goapp/realtime"
)

// ConversationPhase represents the current phase of conversation
type ConversationPhase string

const (
	PhaseConversation ConversationPhase = "conversation"
	PhaseExecution    ConversationPhase = "execution"
)

// ConversationSummary represents the result of a conversation with the user
type ConversationSummary struct {
	Intent                string   `json:"intent"`                // "reorder", "improve_hook", "improve_conclusion", "analyze"
	UserWantsCurrentOrder bool     `json:"userWantsCurrentOrder"` // Whether to include current order as starting point
	OptimizationGoals     []string `json:"optimizationGoals"`     // User's optimization goals (engagement, flow, etc.)
	SpecificRequests      []string `json:"specificRequests"`      // Specific things user asked for
	UserContext           string   `json:"userContext"`           // Important context from user's message
	Confirmed             bool     `json:"confirmed"`             // User has confirmed they want to proceed
}

// UserIntent represents what the user wants to do (LEGACY - keeping for backward compatibility)
type UserIntent struct {
	Action          string                 `json:"action"`           // "reorder", "analyze", "reset", etc.
	Confirmed       bool                   `json:"confirmed"`        // Has user confirmed this action?
	Parameters      map[string]interface{} `json:"parameters"`       // Action-specific parameters
	UserPreferences map[string]interface{} `json:"userPreferences"`  // User choices (e.g., use current order)
	Description     string                 `json:"description"`      // Human description of what will happen
	PrimaryGoal     string                 `json:"primaryGoal"`      // What user mainly wants
	SecondaryGoals  []string               `json:"secondaryGoals"`   // Additional things they mentioned
	Reasoning       string                 `json:"reasoning"`        // LLM's understanding of why user wants this
	Context         string                 `json:"context"`          // Important context from user message
}

// ConversationFlow manages the state of a conversation
type ConversationFlow struct {
	Phase     ConversationPhase      `json:"phase"`
	Intent    *UserIntent            `json:"intent,omitempty"`
	Context   map[string]interface{} `json:"context"`
	SessionID string                 `json:"sessionId"`
}

// ExecutionProgress represents progress during execution
type ExecutionProgress struct {
	Step    string `json:"step"`    // Brief step identifier
	Message string `json:"message"` // Human-readable progress update
}

// ProgressBroadcaster handles real-time progress updates
type ProgressBroadcaster struct {
	manager    *realtime.Manager
	projectID  string
	endpointID string
	sessionID  string
}

// NewProgressBroadcaster creates a new progress broadcaster
func NewProgressBroadcaster(projectID int, endpointID, sessionID string) *ProgressBroadcaster {
	return &ProgressBroadcaster{
		manager:    realtime.GetManager(),
		projectID:  strconv.Itoa(projectID),
		endpointID: endpointID,
		sessionID:  sessionID,
	}
}

// UpdateProgress broadcasts a progress update to the frontend
func (p *ProgressBroadcaster) UpdateProgress(step, message string) {
	if p.manager != nil {
		p.manager.BroadcastChatProgress(p.projectID, p.endpointID, p.sessionID, message)
		log.Printf("Progress [%s]: %s", step, message)
	}
}

// ConversationFlowManager manages the conversation flow lifecycle
type ConversationFlowManager struct {
	flows map[string]*ConversationFlow // sessionID -> flow
}

// NewConversationFlowManager creates a new conversation flow manager
func NewConversationFlowManager() *ConversationFlowManager {
	return &ConversationFlowManager{
		flows: make(map[string]*ConversationFlow),
	}
}

// GetOrCreateFlow gets or creates a conversation flow for a session
func (cfm *ConversationFlowManager) GetOrCreateFlow(sessionID string) *ConversationFlow {
	if flow, exists := cfm.flows[sessionID]; exists {
		return flow
	}
	
	flow := &ConversationFlow{
		Phase:     PhaseConversation,
		Context:   make(map[string]interface{}),
		SessionID: sessionID,
	}
	
	cfm.flows[sessionID] = flow
	return flow
}

// UpdateFlow updates a conversation flow
func (cfm *ConversationFlowManager) UpdateFlow(sessionID string, flow *ConversationFlow) {
	cfm.flows[sessionID] = flow
}

// ClearFlow removes a conversation flow (when session is cleared)
func (cfm *ConversationFlowManager) ClearFlow(sessionID string) {
	delete(cfm.flows, sessionID)
}

// IsIntentConfirmed checks if the user has confirmed their intent
func (flow *ConversationFlow) IsIntentConfirmed() bool {
	return flow.Intent != nil && flow.Intent.Confirmed
}

// ShouldExecute determines if we should move to execution phase
func (flow *ConversationFlow) ShouldExecute() bool {
	return flow.Phase == PhaseConversation && flow.IsIntentConfirmed()
}

// MoveToExecution transitions to execution phase
func (flow *ConversationFlow) MoveToExecution() {
	flow.Phase = PhaseExecution
}

// Reset resets the conversation flow back to conversation phase
func (flow *ConversationFlow) Reset() {
	flow.Phase = PhaseConversation
	flow.Intent = nil
	// Keep context for continuity
}

// SetIntent sets the user intent
func (flow *ConversationFlow) SetIntent(intent *UserIntent) {
	flow.Intent = intent
}

// AddContext adds context information
func (flow *ConversationFlow) AddContext(key string, value interface{}) {
	if flow.Context == nil {
		flow.Context = make(map[string]interface{})
	}
	flow.Context[key] = value
}

// GetContext retrieves context information
func (flow *ConversationFlow) GetContext(key string) (interface{}, bool) {
	if flow.Context == nil {
		return nil, false
	}
	value, exists := flow.Context[key]
	return value, exists
}

// ToJSON serializes the conversation flow to JSON for debugging
func (flow *ConversationFlow) ToJSON() string {
	data, err := json.MarshalIndent(flow, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error serializing flow: %v", err)
	}
	return string(data)
}

// ParseUserIntent attempts to parse user intent from a JSON string
func ParseUserIntent(intentJSON string) (*UserIntent, error) {
	var intent UserIntent
	err := json.Unmarshal([]byte(intentJSON), &intent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user intent: %w", err)
	}
	return &intent, nil
}

// ValidateUserIntent validates that a user intent is complete and valid
func ValidateUserIntent(intent *UserIntent) error {
	if intent == nil {
		return fmt.Errorf("intent cannot be nil")
	}
	
	if intent.Action == "" {
		return fmt.Errorf("intent action cannot be empty")
	}
	
	if !intent.Confirmed {
		return fmt.Errorf("intent must be confirmed before execution")
	}
	
	return nil
}