package realtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetManager(t *testing.T) {
	// Reset global manager for testing
	globalManager = nil
	once = sync.Once{}

	manager1 := GetManager()
	manager2 := GetManager()

	// Should return the same instance (singleton)
	assert.Same(t, manager1, manager2)
	assert.NotNil(t, manager1.sseManager)
	
	// Clean up
	manager1.Shutdown()
}

func TestManager_SetContext(t *testing.T) {
	// Reset global manager for testing
	globalManager = nil
	once = sync.Once{}

	manager := GetManager()
	defer manager.Shutdown()

	ctx := context.Background()
	manager.SetContext(ctx)

	// Can't directly test if context is set due to private field,
	// but we can test that it doesn't panic
	assert.NotNil(t, manager)
}

func TestManager_HandleSSEConnection(t *testing.T) {
	// Reset global manager for testing
	globalManager = nil
	once = sync.Once{}

	manager := GetManager()
	defer manager.Shutdown()

	t.Run("missing project ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/sse", nil)
		rr := httptest.NewRecorder()

		manager.HandleSSEConnection(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Project ID is required")
	})

	t.Run("invalid project ID format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/sse?projectId=invalid", nil)
		rr := httptest.NewRecorder()

		manager.HandleSSEConnection(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid project ID format")
	})

	t.Run("valid project ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/sse?projectId=123", nil)
		
		// Create a custom ResponseRecorder that implements Flusher
		rr := &FlushableRecorder{ResponseRecorder: httptest.NewRecorder()}

		// Use a context that can be cancelled
		ctx, cancel := context.WithCancel(context.Background())
		req = req.WithContext(ctx)

		done := make(chan bool)
		go func() {
			defer func() { done <- true }()
			manager.HandleSSEConnection(rr, req)
		}()

		// Give some time for headers to be set
		time.Sleep(10 * time.Millisecond)

		// Cancel the request context to simulate client disconnect
		cancel()

		// Wait for handler to complete
		<-done

		// Check SSE headers were set
		assert.Equal(t, "text/event-stream", rr.Header().Get("Content-Type"))
		assert.Equal(t, "no-cache", rr.Header().Get("Cache-Control"))
		assert.Equal(t, "keep-alive", rr.Header().Get("Connection"))
		assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	})
}

// FlushableRecorder implements both http.ResponseWriter and http.Flusher
type FlushableRecorder struct {
	*httptest.ResponseRecorder
}

func (f *FlushableRecorder) Flush() {
	// No-op for testing
}

func TestManager_BroadcastHighlightsUpdate(t *testing.T) {
	// Reset global manager for testing
	globalManager = nil
	once = sync.Once{}

	manager := GetManager()
	defer manager.Shutdown()

	// Don't set context to avoid Wails runtime issues in tests
	highlights := map[string]interface{}{
		"id":    "123",
		"start": 10.0,
		"end":   20.0,
	}

	// This should not panic and should create the event
	manager.BroadcastHighlightsUpdate("123", highlights)

	// Add a brief sleep to allow any goroutines to process
	time.Sleep(10 * time.Millisecond)
}

func TestManager_BroadcastHighlightsDelete(t *testing.T) {
	// Reset global manager for testing
	globalManager = nil
	once = sync.Once{}

	manager := GetManager()
	defer manager.Shutdown()

	highlightIDs := []string{"123", "456"}

	manager.BroadcastHighlightsDelete("123", highlightIDs)
	time.Sleep(10 * time.Millisecond)
}

func TestManager_BroadcastHighlightsReorder(t *testing.T) {
	// Reset global manager for testing
	globalManager = nil
	once = sync.Once{}

	manager := GetManager()
	defer manager.Shutdown()

	newOrder := []interface{}{"3", "1", "2"}

	manager.BroadcastHighlightsReorder("123", newOrder)
	time.Sleep(10 * time.Millisecond)
}

func TestManager_BroadcastProjectUpdate(t *testing.T) {
	// Reset global manager for testing
	globalManager = nil
	once = sync.Once{}

	manager := GetManager()
	defer manager.Shutdown()

	project := map[string]interface{}{
		"id":   123,
		"name": "Test Project",
	}

	manager.BroadcastProjectUpdate("123", project)
	time.Sleep(10 * time.Millisecond)
}

func TestManager_BroadcastChatMessageAdded(t *testing.T) {
	// Reset global manager for testing
	globalManager = nil
	once = sync.Once{}

	manager := GetManager()
	defer manager.Shutdown()

	message := map[string]interface{}{
		"id":      "msg123",
		"content": "Hello world",
	}

	manager.BroadcastChatMessageAdded("123", "endpoint1", "session1", message)
	time.Sleep(10 * time.Millisecond)
}

func TestManager_BroadcastChatHistoryCleared(t *testing.T) {
	// Reset global manager for testing
	globalManager = nil
	once = sync.Once{}

	manager := GetManager()
	defer manager.Shutdown()

	manager.BroadcastChatHistoryCleared("123", "endpoint1", "session1")
	time.Sleep(10 * time.Millisecond)
}

func TestManager_BroadcastChatSessionUpdated(t *testing.T) {
	// Reset global manager for testing
	globalManager = nil
	once = sync.Once{}

	manager := GetManager()
	defer manager.Shutdown()

	messages := []interface{}{
		map[string]interface{}{"id": "1", "content": "Hello"},
		map[string]interface{}{"id": "2", "content": "World"},
	}

	manager.BroadcastChatSessionUpdated("123", "endpoint1", "session1", messages)
	time.Sleep(10 * time.Millisecond)
}

func TestManager_BroadcastChatProgress(t *testing.T) {
	// Reset global manager for testing
	globalManager = nil
	once = sync.Once{}

	manager := GetManager()
	defer manager.Shutdown()

	manager.BroadcastChatProgress("123", "endpoint1", "session1", "Processing...")
	time.Sleep(10 * time.Millisecond)
}

func TestManager_GetStats(t *testing.T) {
	// Reset global manager for testing
	globalManager = nil
	once = sync.Once{}

	manager := GetManager()
	defer manager.Shutdown()

	stats := manager.GetStats()
	require.NotNil(t, stats)

	totalClients, exists := stats["totalClients"]
	assert.True(t, exists)
	assert.Equal(t, 0, totalClients) // Initially no clients
}

func TestManager_GetProjectStats(t *testing.T) {
	// Reset global manager for testing
	globalManager = nil
	once = sync.Once{}

	manager := GetManager()
	defer manager.Shutdown()

	stats := manager.GetProjectStats("123")
	require.NotNil(t, stats)

	projectClients, exists := stats["projectClients"]
	assert.True(t, exists)
	assert.Equal(t, 0, projectClients) // Initially no clients
}

func TestManager_Shutdown(t *testing.T) {
	// Reset global manager for testing
	globalManager = nil
	once = sync.Once{}

	manager := GetManager()
	
	assert.NotNil(t, manager.sseManager)
	
	manager.Shutdown()
	
	// After shutdown, the manager should still exist but internal state may be cleaned up
	assert.NotNil(t, manager)
}

func TestGenerateClientID(t *testing.T) {
	id1, err1 := generateClientID()
	id2, err2 := generateClientID()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2) // Should generate unique IDs
	assert.Equal(t, 32, len(id1)) // 16 bytes * 2 hex chars = 32 chars
	assert.Equal(t, 32, len(id2))
}

func TestProjectIDToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"string input", "123", "123"},
		{"int input", 123, "123"},
		{"int64 input", int64(123), "123"},
		{"float input", 123.456, "123.456"},
		{"bool input", true, "true"},
		{"nil input", nil, "<nil>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := projectIDToString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManager_IntegrationWithSSE(t *testing.T) {
	// Reset global manager for testing
	globalManager = nil
	once = sync.Once{}

	manager := GetManager()
	defer manager.Shutdown()

	// Create a mock SSE client
	mockWriter := &MockResponseWriter{ResponseRecorder: httptest.NewRecorder()}
	client := NewClient("client123", "project456", mockWriter)
	require.NotNil(t, client)

	// Add client to manager
	manager.sseManager.AddClient(client)
	time.Sleep(10 * time.Millisecond)

	// Clear connection confirmation messages
	mockWriter.Body.Reset()

	// Test broadcasting highlights update
	highlights := map[string]interface{}{"id": "h1", "start": 10.0}
	manager.BroadcastHighlightsUpdate("project456", highlights)
	time.Sleep(10 * time.Millisecond)

	// Verify the client received the event
	body := mockWriter.Body.String()
	assert.Contains(t, body, "highlights_updated")
	assert.Contains(t, body, "project456")
}

func TestManager_BroadcastsWithoutContext(t *testing.T) {
	// Reset global manager for testing
	globalManager = nil
	once = sync.Once{}

	manager := GetManager()
	defer manager.Shutdown()

	// Don't set context - should still work for SSE broadcasts
	highlights := map[string]interface{}{"id": "h1"}
	
	// These should not panic even without Wails context
	manager.BroadcastHighlightsUpdate("123", highlights)
	manager.BroadcastHighlightsDelete("123", []string{"h1"})
	manager.BroadcastChatMessageAdded("123", "e1", "s1", map[string]string{"msg": "test"})
	
	time.Sleep(10 * time.Millisecond)
}