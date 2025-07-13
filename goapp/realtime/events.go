package realtime

import (
	"encoding/json"
	"time"
)

// EventType represents the type of real-time event
type EventType string

const (
	// Highlight events
	EventHighlightsUpdated  EventType = "highlights_updated"
	EventHighlightsDeleted  EventType = "highlights_deleted"
	EventHighlightsReordered EventType = "highlights_reordered"
	
	// Project events
	EventProjectUpdated EventType = "project_updated"
	
	// Chatbot events
	EventChatMessageAdded   EventType = "chat_message_added"
	EventChatHistoryCleared EventType = "chat_history_cleared"
	EventChatSessionUpdated EventType = "chat_session_updated"
	EventChatProgress       EventType = "chat_progress"
	
	// Connection events
	EventConnected    EventType = "connected"
	EventDisconnected EventType = "disconnected"
)

// Event represents a real-time event message
type Event struct {
	Type      EventType   `json:"type"`
	ProjectID string      `json:"projectId"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewEvent creates a new event with current timestamp
func NewEvent(eventType EventType, projectID string, data interface{}) *Event {
	return &Event{
		Type:      eventType,
		ProjectID: projectID,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// ToSSE formats the event as Server-Sent Events format
func (e *Event) ToSSE() (string, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	
	return "data: " + string(data) + "\n\n", nil
}

// HighlightsUpdateData represents data for highlights update events
type HighlightsUpdateData struct {
	Highlights interface{} `json:"highlights"`
	UpdatedBy  string      `json:"updatedBy,omitempty"`
}

// HighlightsDeleteData represents data for highlights delete events
type HighlightsDeleteData struct {
	HighlightIDs []string `json:"highlightIds"`
	DeletedBy    string   `json:"deletedBy,omitempty"`
}

// HighlightsReorderData represents data for highlights reorder events
type HighlightsReorderData struct {
	NewOrder  []interface{} `json:"newOrder"`
	ReorderedBy string       `json:"reorderedBy,omitempty"`
}

// ProjectUpdateData represents data for project update events
type ProjectUpdateData struct {
	Project   interface{} `json:"project"`
	UpdatedBy string      `json:"updatedBy,omitempty"`
}

// ChatMessageAddedData represents data for chat message added events
type ChatMessageAddedData struct {
	EndpointID string      `json:"endpointId"`
	SessionID  string      `json:"sessionId"`
	Message    interface{} `json:"message"`
	AddedBy    string      `json:"addedBy,omitempty"`
}

// ChatHistoryClearedData represents data for chat history cleared events
type ChatHistoryClearedData struct {
	EndpointID string `json:"endpointId"`
	SessionID  string `json:"sessionId,omitempty"`
	ClearedBy  string `json:"clearedBy,omitempty"`
}

// ChatSessionUpdatedData represents data for chat session updated events
type ChatSessionUpdatedData struct {
	EndpointID string        `json:"endpointId"`
	SessionID  string        `json:"sessionId"`
	Messages   []interface{} `json:"messages,omitempty"`
	UpdatedBy  string        `json:"updatedBy,omitempty"`
}

// ChatProgressData represents data for chat progress events
type ChatProgressData struct {
	EndpointID string `json:"endpointId"`
	SessionID  string `json:"sessionId"`
	Message    string `json:"message"`
}