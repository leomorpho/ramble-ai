package realtime

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEvent(t *testing.T) {
	eventType := EventHighlightsUpdated
	projectID := "123"
	data := map[string]string{"test": "data"}

	event := NewEvent(eventType, projectID, data)

	assert.Equal(t, eventType, event.Type)
	assert.Equal(t, projectID, event.ProjectID)
	assert.Equal(t, data, event.Data)
	assert.True(t, time.Since(event.Timestamp) < time.Second)
}

func TestEvent_ToSSE(t *testing.T) {
	t.Run("successful conversion", func(t *testing.T) {
		event := &Event{
			Type:      EventHighlightsUpdated,
			ProjectID: "123",
			Data:      map[string]string{"test": "data"},
			Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		}

		sseData, err := event.ToSSE()

		require.NoError(t, err)
		assert.Contains(t, sseData, "data: ")
		assert.True(t, strings.HasSuffix(sseData, "\n\n"))

		// Parse the JSON to verify it's valid
		jsonStart := len("data: ")
		jsonEnd := len(sseData) - 2 // Remove \n\n
		jsonData := sseData[jsonStart:jsonEnd]

		var parsedEvent Event
		err = json.Unmarshal([]byte(jsonData), &parsedEvent)
		require.NoError(t, err)
		assert.Equal(t, event.Type, parsedEvent.Type)
		assert.Equal(t, event.ProjectID, parsedEvent.ProjectID)
	})

	t.Run("invalid data", func(t *testing.T) {
		event := &Event{
			Type:      EventHighlightsUpdated,
			ProjectID: "123",
			Data:      make(chan int), // Unmarshalable data
			Timestamp: time.Now(),
		}

		_, err := event.ToSSE()
		assert.Error(t, err)
	})
}

func TestEventTypes(t *testing.T) {
	// Test that all event types are defined
	eventTypes := []EventType{
		EventHighlightsUpdated,
		EventHighlightsDeleted,
		EventHighlightsReordered,
		EventProjectUpdated,
		EventChatMessageAdded,
		EventChatHistoryCleared,
		EventChatSessionUpdated,
		EventChatProgress,
		EventConnected,
		EventDisconnected,
	}

	for _, eventType := range eventTypes {
		assert.NotEmpty(t, string(eventType))
	}
}

func TestHighlightsUpdateData(t *testing.T) {
	data := &HighlightsUpdateData{
		Highlights: map[string]interface{}{"id": "123"},
		UpdatedBy:  "user1",
	}

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	var unmarshaled HighlightsUpdateData
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, data.UpdatedBy, unmarshaled.UpdatedBy)
}

func TestHighlightsDeleteData(t *testing.T) {
	data := &HighlightsDeleteData{
		HighlightIDs: []string{"123", "456"},
		DeletedBy:    "user1",
	}

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	var unmarshaled HighlightsDeleteData
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, data.HighlightIDs, unmarshaled.HighlightIDs)
	assert.Equal(t, data.DeletedBy, unmarshaled.DeletedBy)
}

func TestHighlightsReorderData(t *testing.T) {
	data := &HighlightsReorderData{
		NewOrder:    []interface{}{"1", "2", "3"},
		ReorderedBy: "user1",
	}

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	var unmarshaled HighlightsReorderData
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, len(data.NewOrder), len(unmarshaled.NewOrder))
	assert.Equal(t, data.ReorderedBy, unmarshaled.ReorderedBy)
}

func TestProjectUpdateData(t *testing.T) {
	project := map[string]interface{}{
		"id":   123,
		"name": "Test Project",
	}

	data := &ProjectUpdateData{
		Project:   project,
		UpdatedBy: "user1",
	}

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	var unmarshaled ProjectUpdateData
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, data.UpdatedBy, unmarshaled.UpdatedBy)
	assert.NotNil(t, unmarshaled.Project)
}

func TestChatMessageAddedData(t *testing.T) {
	message := map[string]interface{}{
		"id":      "msg123",
		"content": "Hello world",
	}

	data := &ChatMessageAddedData{
		EndpointID: "endpoint1",
		SessionID:  "session1",
		Message:    message,
		AddedBy:    "user1",
	}

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	var unmarshaled ChatMessageAddedData
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, data.EndpointID, unmarshaled.EndpointID)
	assert.Equal(t, data.SessionID, unmarshaled.SessionID)
	assert.Equal(t, data.AddedBy, unmarshaled.AddedBy)
	assert.NotNil(t, unmarshaled.Message)
}

func TestChatHistoryClearedData(t *testing.T) {
	data := &ChatHistoryClearedData{
		EndpointID: "endpoint1",
		SessionID:  "session1",
		ClearedBy:  "user1",
	}

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	var unmarshaled ChatHistoryClearedData
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, data.EndpointID, unmarshaled.EndpointID)
	assert.Equal(t, data.SessionID, unmarshaled.SessionID)
	assert.Equal(t, data.ClearedBy, unmarshaled.ClearedBy)
}

func TestChatSessionUpdatedData(t *testing.T) {
	messages := []interface{}{
		map[string]interface{}{"id": "1", "content": "Hello"},
		map[string]interface{}{"id": "2", "content": "World"},
	}

	data := &ChatSessionUpdatedData{
		EndpointID: "endpoint1",
		SessionID:  "session1",
		Messages:   messages,
		UpdatedBy:  "user1",
	}

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	var unmarshaled ChatSessionUpdatedData
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, data.EndpointID, unmarshaled.EndpointID)
	assert.Equal(t, data.SessionID, unmarshaled.SessionID)
	assert.Equal(t, len(data.Messages), len(unmarshaled.Messages))
	assert.Equal(t, data.UpdatedBy, unmarshaled.UpdatedBy)
}

func TestChatProgressData(t *testing.T) {
	data := &ChatProgressData{
		EndpointID: "endpoint1",
		SessionID:  "session1",
		Message:    "Processing...",
	}

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	var unmarshaled ChatProgressData
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, data.EndpointID, unmarshaled.EndpointID)
	assert.Equal(t, data.SessionID, unmarshaled.SessionID)
	assert.Equal(t, data.Message, unmarshaled.Message)
}