package realtime

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Manager is the central coordinator for real-time functionality
type Manager struct {
	sseManager *SSEManager
	ctx        context.Context
	mu         sync.RWMutex
}

var (
	globalManager *Manager
	once          sync.Once
)

// GetManager returns the singleton real-time manager
func GetManager() *Manager {
	once.Do(func() {
		globalManager = &Manager{
			sseManager: NewSSEManager(),
		}
		log.Println("Real-time manager initialized")
	})
	return globalManager
}

// SetContext sets the Wails context for broadcasting events
func (m *Manager) SetContext(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ctx = ctx
}

// HandleSSEConnection handles incoming SSE connection requests
func (m *Manager) HandleSSEConnection(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Extract project ID from URL path or query params
	projectID := r.URL.Query().Get("projectId")
	if projectID == "" {
		// Try to extract from path
		// Assuming URL format: /api/sse/highlights/{projectId}
		// You may need to adjust this based on your routing
		pathParts := r.URL.Path
		if len(pathParts) > 0 {
			// Extract project ID from path - this is a simplified approach
			// In a real router, you'd use proper path parameter extraction
			projectID = r.URL.Query().Get("projectId")
		}
	}

	if projectID == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	// Validate project ID format if needed
	if _, err := strconv.Atoi(projectID); err != nil {
		http.Error(w, "Invalid project ID format", http.StatusBadRequest)
		return
	}

	// Generate unique client ID
	clientID, err := generateClientID()
	if err != nil {
		http.Error(w, "Failed to generate client ID", http.StatusInternalServerError)
		return
	}

	// Create new client
	client := NewClient(clientID, projectID, w)
	if client == nil {
		http.Error(w, "Failed to create SSE client", http.StatusInternalServerError)
		return
	}

	// Add client to manager
	m.sseManager.AddClient(client)

	// Handle client disconnection
	notify := r.Context().Done()
	go func() {
		<-notify
		m.sseManager.RemoveClient(clientID)
	}()

	// Keep connection alive
	<-client.Done
}

// BroadcastHighlightsUpdate broadcasts a highlights update event
func (m *Manager) BroadcastHighlightsUpdate(projectID string, highlights interface{}) {
	data := &HighlightsUpdateData{
		Highlights: highlights,
	}
	
	event := NewEvent(EventHighlightsUpdated, projectID, data)
	
	// Broadcast via SSE for browser connections
	m.sseManager.BroadcastToProject(projectID, event)
	
	// Broadcast via Wails events for desktop app
	if m.ctx != nil {
		runtime.EventsEmit(m.ctx, string(EventHighlightsUpdated), event)
	}
	
	log.Printf("Broadcasted highlights update for project %s", projectID)
}

// BroadcastHighlightsDelete broadcasts a highlights delete event
func (m *Manager) BroadcastHighlightsDelete(projectID string, highlightIDs []string) {
	data := &HighlightsDeleteData{
		HighlightIDs: highlightIDs,
	}
	
	event := NewEvent(EventHighlightsDeleted, projectID, data)
	
	// Broadcast via SSE for browser connections
	m.sseManager.BroadcastToProject(projectID, event)
	
	// Broadcast via Wails events for desktop app
	if m.ctx != nil {
		runtime.EventsEmit(m.ctx, string(EventHighlightsDeleted), event)
	}
	
	log.Printf("Broadcasted highlights delete for project %s", projectID)
}

// BroadcastHighlightsReorder broadcasts a highlights reorder event
func (m *Manager) BroadcastHighlightsReorder(projectID string, newOrder []interface{}) {
	data := &HighlightsReorderData{
		NewOrder: newOrder,
	}
	
	event := NewEvent(EventHighlightsReordered, projectID, data)
	
	// Broadcast via SSE for browser connections
	m.sseManager.BroadcastToProject(projectID, event)
	
	// Broadcast via Wails events for desktop app
	if m.ctx != nil {
		runtime.EventsEmit(m.ctx, string(EventHighlightsReordered), event)
	}
	
	log.Printf("Broadcasted highlights reorder for project %s", projectID)
}

// BroadcastProjectUpdate broadcasts a project update event
func (m *Manager) BroadcastProjectUpdate(projectID string, project interface{}) {
	data := &ProjectUpdateData{
		Project: project,
	}
	
	event := NewEvent(EventProjectUpdated, projectID, data)
	m.sseManager.BroadcastToProject(projectID, event)
	
	log.Printf("Broadcasted project update for project %s", projectID)
}

// BroadcastChatMessageAdded broadcasts a chat message added event
func (m *Manager) BroadcastChatMessageAdded(projectID string, endpointID string, sessionID string, message interface{}) {
	data := &ChatMessageAddedData{
		EndpointID: endpointID,
		SessionID:  sessionID,
		Message:    message,
	}
	
	event := NewEvent(EventChatMessageAdded, projectID, data)
	
	// Broadcast via SSE for browser connections
	m.sseManager.BroadcastToProject(projectID, event)
	
	// Broadcast via Wails events for desktop app
	if m.ctx != nil {
		runtime.EventsEmit(m.ctx, string(EventChatMessageAdded), event)
	}
	
	log.Printf("Broadcasted chat message added for project %s, endpoint %s", projectID, endpointID)
}

// BroadcastChatHistoryCleared broadcasts a chat history cleared event
func (m *Manager) BroadcastChatHistoryCleared(projectID string, endpointID string, sessionID string) {
	data := &ChatHistoryClearedData{
		EndpointID: endpointID,
		SessionID:  sessionID,
	}
	
	event := NewEvent(EventChatHistoryCleared, projectID, data)
	
	// Broadcast via SSE for browser connections
	m.sseManager.BroadcastToProject(projectID, event)
	
	// Broadcast via Wails events for desktop app
	if m.ctx != nil {
		runtime.EventsEmit(m.ctx, string(EventChatHistoryCleared), event)
	}
	
	log.Printf("Broadcasted chat history cleared for project %s, endpoint %s", projectID, endpointID)
}

// BroadcastChatSessionUpdated broadcasts a chat session updated event
func (m *Manager) BroadcastChatSessionUpdated(projectID string, endpointID string, sessionID string, messages []interface{}) {
	data := &ChatSessionUpdatedData{
		EndpointID: endpointID,
		SessionID:  sessionID,
		Messages:   messages,
	}
	
	event := NewEvent(EventChatSessionUpdated, projectID, data)
	
	// Broadcast via SSE for browser connections
	m.sseManager.BroadcastToProject(projectID, event)
	
	// Broadcast via Wails events for desktop app
	if m.ctx != nil {
		runtime.EventsEmit(m.ctx, string(EventChatSessionUpdated), event)
	}
	
	log.Printf("Broadcasted chat session updated for project %s, endpoint %s", projectID, endpointID)
}

// GetStats returns statistics about connected clients
func (m *Manager) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"totalClients": m.sseManager.GetClientCount(),
	}
}

// GetProjectStats returns statistics for a specific project
func (m *Manager) GetProjectStats(projectID string) map[string]interface{} {
	return map[string]interface{}{
		"projectClients": m.sseManager.GetProjectClientCount(projectID),
	}
}

// Shutdown gracefully shuts down the real-time manager
func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.sseManager != nil {
		m.sseManager.Shutdown()
		log.Println("Real-time manager shut down")
	}
}

// generateClientID generates a unique client identifier
func generateClientID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Helper function to convert project ID to string safely
func projectIDToString(projectID interface{}) string {
	switch v := projectID.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}