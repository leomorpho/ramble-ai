package realtime

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockResponseWriter implements http.ResponseWriter and http.Flusher for testing
type MockResponseWriter struct {
	*httptest.ResponseRecorder
}

func (m *MockResponseWriter) Flush() {
	// No-op for testing
}

// NonFlushingWriter doesn't implement http.Flusher
type NonFlushingWriter struct{}

func (n *NonFlushingWriter) Header() http.Header {
	return make(http.Header)
}

func (n *NonFlushingWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (n *NonFlushingWriter) WriteHeader(statusCode int) {}

func TestNewClient(t *testing.T) {
	t.Run("successful client creation", func(t *testing.T) {
		mockWriter := &MockResponseWriter{ResponseRecorder: httptest.NewRecorder()}
		client := NewClient("client123", "project456", mockWriter)

		require.NotNil(t, client)
		assert.Equal(t, "client123", client.ID)
		assert.Equal(t, "project456", client.ProjectID)
		assert.Equal(t, mockWriter, client.Writer)
		assert.NotNil(t, client.Flusher)
		assert.NotNil(t, client.Done)
		assert.True(t, time.Since(client.LastPing) < time.Second)
	})

	t.Run("fails with non-flusher writer", func(t *testing.T) {
		// Use a writer that doesn't implement Flusher
		writer := &NonFlushingWriter{}
		client := NewClient("client123", "project456", writer)

		assert.Nil(t, client)
	})
}

func TestClient_Send(t *testing.T) {
	mockWriter := &MockResponseWriter{ResponseRecorder: httptest.NewRecorder()}
	client := NewClient("client123", "project456", mockWriter)
	require.NotNil(t, client)

	event := NewEvent(EventHighlightsUpdated, "project456", map[string]string{"test": "data"})

	err := client.Send(event)
	assert.NoError(t, err)

	// Check that data was written to the response
	body := mockWriter.Body.String()
	assert.Contains(t, body, "data: ")
	assert.Contains(t, body, "highlights_updated")
	assert.Contains(t, body, "project456")
}

func TestClient_SendPing(t *testing.T) {
	mockWriter := &MockResponseWriter{ResponseRecorder: httptest.NewRecorder()}
	client := NewClient("client123", "project456", mockWriter)
	require.NotNil(t, client)

	oldPing := client.LastPing
	time.Sleep(time.Millisecond) // Ensure different timestamp

	err := client.SendPing()
	assert.NoError(t, err)

	// Check that ping data was written
	body := mockWriter.Body.String()
	assert.Contains(t, body, ": ping")

	// Check that LastPing was updated
	assert.True(t, client.LastPing.After(oldPing))
}

func TestClient_Close(t *testing.T) {
	mockWriter := &MockResponseWriter{ResponseRecorder: httptest.NewRecorder()}
	client := NewClient("client123", "project456", mockWriter)
	require.NotNil(t, client)

	// Test closing once
	go client.Close() // Close in goroutine to avoid blocking

	// Verify that we can read from Done channel
	select {
	case <-client.Done:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected to receive from Done channel")
	}

	// Test closing twice (should not block)
	client.Close() // This should not block even if channel is closed
}

func TestNewSSEManager(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Shutdown()

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.clients)
	assert.NotNil(t, manager.projectClients)
	assert.NotNil(t, manager.eventChan)
	assert.NotNil(t, manager.addClientChan)
	assert.NotNil(t, manager.removeClientChan)
	assert.NotNil(t, manager.ctx)
	assert.NotNil(t, manager.cancel)

	// Initial state should be empty
	assert.Equal(t, 0, manager.GetClientCount())
}

func TestSSEManager_AddClient(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Shutdown()

	mockWriter := &MockResponseWriter{ResponseRecorder: httptest.NewRecorder()}
	client := NewClient("client123", "project456", mockWriter)
	require.NotNil(t, client)

	manager.AddClient(client)

	// Give some time for the goroutine to process
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, manager.GetClientCount())
	assert.Equal(t, 1, manager.GetProjectClientCount("project456"))
	assert.Equal(t, 0, manager.GetProjectClientCount("nonexistent"))

	// Check that connection confirmation was sent
	body := mockWriter.Body.String()
	assert.Contains(t, body, "connected")
	assert.Contains(t, body, "client123")
}

func TestSSEManager_RemoveClient(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Shutdown()

	mockWriter := &MockResponseWriter{ResponseRecorder: httptest.NewRecorder()}
	client := NewClient("client123", "project456", mockWriter)
	require.NotNil(t, client)

	// Add client first
	manager.AddClient(client)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, manager.GetClientCount())

	// Remove client
	manager.RemoveClient("client123")
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 0, manager.GetClientCount())
	assert.Equal(t, 0, manager.GetProjectClientCount("project456"))
}

func TestSSEManager_BroadcastToProject(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Shutdown()

	// Add two clients for the same project
	mockWriter1 := &MockResponseWriter{ResponseRecorder: httptest.NewRecorder()}
	client1 := NewClient("client1", "project456", mockWriter1)
	require.NotNil(t, client1)

	mockWriter2 := &MockResponseWriter{ResponseRecorder: httptest.NewRecorder()}
	client2 := NewClient("client2", "project456", mockWriter2)
	require.NotNil(t, client2)

	manager.AddClient(client1)
	manager.AddClient(client2)
	time.Sleep(10 * time.Millisecond)

	// Clear any connection confirmation messages
	mockWriter1.Body.Reset()
	mockWriter2.Body.Reset()

	// Broadcast an event
	event := NewEvent(EventHighlightsUpdated, "project456", map[string]string{"test": "broadcast"})
	manager.BroadcastToProject("project456", event)

	// Give time for the broadcast to process
	time.Sleep(10 * time.Millisecond)

	// Check that both clients received the event
	body1 := mockWriter1.Body.String()
	body2 := mockWriter2.Body.String()

	assert.Contains(t, body1, "highlights_updated")
	assert.Contains(t, body1, "broadcast")
	assert.Contains(t, body2, "highlights_updated")
	assert.Contains(t, body2, "broadcast")
}

func TestSSEManager_ClientCounts(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Shutdown()

	assert.Equal(t, 0, manager.GetClientCount())
	assert.Equal(t, 0, manager.GetProjectClientCount("project1"))

	// Add clients for different projects
	mockWriter1 := &MockResponseWriter{ResponseRecorder: httptest.NewRecorder()}
	client1 := NewClient("client1", "project1", mockWriter1)
	require.NotNil(t, client1)

	mockWriter2 := &MockResponseWriter{ResponseRecorder: httptest.NewRecorder()}
	client2 := NewClient("client2", "project1", mockWriter2)
	require.NotNil(t, client2)

	mockWriter3 := &MockResponseWriter{ResponseRecorder: httptest.NewRecorder()}
	client3 := NewClient("client3", "project2", mockWriter3)
	require.NotNil(t, client3)

	manager.AddClient(client1)
	manager.AddClient(client2)
	manager.AddClient(client3)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 3, manager.GetClientCount())
	assert.Equal(t, 2, manager.GetProjectClientCount("project1"))
	assert.Equal(t, 1, manager.GetProjectClientCount("project2"))
}

func TestSSEManager_PingClients(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Shutdown()

	mockWriter := &MockResponseWriter{ResponseRecorder: httptest.NewRecorder()}
	client := NewClient("client123", "project456", mockWriter)
	require.NotNil(t, client)

	manager.AddClient(client)
	time.Sleep(10 * time.Millisecond)

	// Clear connection confirmation
	mockWriter.Body.Reset()

	// Manually trigger ping (since we can't wait 30 seconds)
	manager.sendPingsToAllClients()

	// Check that ping was sent
	body := mockWriter.Body.String()
	assert.Contains(t, body, ": ping")
}

func TestSSEManager_Shutdown(t *testing.T) {
	manager := NewSSEManager()

	mockWriter := &MockResponseWriter{ResponseRecorder: httptest.NewRecorder()}
	client := NewClient("client123", "project456", mockWriter)
	require.NotNil(t, client)

	manager.AddClient(client)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, manager.GetClientCount())

	manager.Shutdown()
	time.Sleep(10 * time.Millisecond)

	// After shutdown, adding clients should not work
	mockWriter2 := &MockResponseWriter{ResponseRecorder: httptest.NewRecorder()}
	client2 := NewClient("client456", "project789", mockWriter2)
	require.NotNil(t, client2)

	manager.AddClient(client2)
	time.Sleep(10 * time.Millisecond)

	// Should still be 1 or 0 (depending on cleanup timing)
	// The important thing is it shouldn't increase
	assert.LessOrEqual(t, manager.GetClientCount(), 1)
}

func TestSSEManager_ConcurrentOperations(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Shutdown()

	var wg sync.WaitGroup
	numClients := 10

	// Add multiple clients concurrently
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mockWriter := &MockResponseWriter{ResponseRecorder: httptest.NewRecorder()}
			client := NewClient(
				"client"+strings.Repeat("0", 2-len(string(rune(id))))+string(rune(id)), 
				"project1", 
				mockWriter,
			)
			if client != nil {
				manager.AddClient(client)
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond) // Give time for all operations to complete

	// Should have all clients
	assert.Equal(t, numClients, manager.GetClientCount())

	// Broadcast to all clients concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			event := NewEvent(EventHighlightsUpdated, "project1", map[string]interface{}{
				"message": "concurrent_test",
				"id":      id,
			})
			manager.BroadcastToProject("project1", event)
		}(i)
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	// All clients should still be connected
	assert.Equal(t, numClients, manager.GetClientCount())
}

func TestSSEManager_FailedClientCleanup(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Shutdown()

	// Create a client that will fail to send - but only after initial connection
	mockWriter := &ConditionallyFailingResponseWriter{failAfterFirst: true}
	client := NewClient("failing_client", "project1", mockWriter)
	require.NotNil(t, client)

	manager.AddClient(client)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, manager.GetClientCount())

	// Try to broadcast - this should cause the client to be removed
	event := NewEvent(EventHighlightsUpdated, "project1", map[string]string{"test": "data"})
	manager.BroadcastToProject("project1", event)

	time.Sleep(20 * time.Millisecond) // Give time for cleanup

	// Failed client should be removed
	assert.Equal(t, 0, manager.GetClientCount())
}

// FailingResponseWriter simulates a failed connection
type FailingResponseWriter struct{}

func (f *FailingResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (f *FailingResponseWriter) Write([]byte) (int, error) {
	return 0, assert.AnError // Simulate write failure
}

func (f *FailingResponseWriter) WriteHeader(statusCode int) {}

func (f *FailingResponseWriter) Flush() {}

// ConditionallyFailingResponseWriter fails only after first successful write
type ConditionallyFailingResponseWriter struct {
	failAfterFirst bool
	hasWritten     bool
}

func (c *ConditionallyFailingResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (c *ConditionallyFailingResponseWriter) Write(data []byte) (int, error) {
	if c.failAfterFirst && c.hasWritten {
		return 0, assert.AnError // Fail on second write
	}
	c.hasWritten = true
	return len(data), nil // First write succeeds
}

func (c *ConditionallyFailingResponseWriter) WriteHeader(statusCode int) {}

func (c *ConditionallyFailingResponseWriter) Flush() {}

func TestSSEManager_StaleClientCleanup(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Shutdown()

	mockWriter := &MockResponseWriter{ResponseRecorder: httptest.NewRecorder()}
	client := NewClient("stale_client", "project1", mockWriter)
	require.NotNil(t, client)

	// Set LastPing to be very old
	client.LastPing = time.Now().Add(-10 * time.Minute)

	manager.AddClient(client)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, manager.GetClientCount())

	// Trigger ping cleanup
	manager.sendPingsToAllClients()
	time.Sleep(10 * time.Millisecond)

	// Stale client should be removed
	assert.Equal(t, 0, manager.GetClientCount())
}