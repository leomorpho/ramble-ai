package chatbot

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProgressBroadcaster(t *testing.T) {
	broadcaster := NewProgressBroadcaster(123, "endpoint1", "session1")

	assert.Equal(t, "123", broadcaster.projectID)
	assert.Equal(t, "endpoint1", broadcaster.endpointID)
	assert.Equal(t, "session1", broadcaster.sessionID)
}

func TestProgressBroadcaster_UpdateProgress(t *testing.T) {
	broadcaster := NewProgressBroadcaster(123, "endpoint1", "session1")

	// This should not panic (realtime manager will handle the actual broadcast)
	broadcaster.UpdateProgress("processing", "Analyzing highlights...")
	broadcaster.UpdateProgress("reordering", "Reordering based on AI suggestions...")
	broadcaster.UpdateProgress("complete", "Complete!")
}

func TestNewConversationFlowManager(t *testing.T) {
	manager := NewConversationFlowManager()

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.flows)
}

func TestConversationFlowManager_GetOrCreateFlow(t *testing.T) {
	manager := NewConversationFlowManager()
	sessionID := "test-session-123"

	t.Run("creates new flow", func(t *testing.T) {
		flow := manager.GetOrCreateFlow(sessionID)

		assert.NotNil(t, flow)
		assert.Equal(t, sessionID, flow.SessionID)
		assert.Equal(t, PhaseConversation, flow.Phase)
		assert.Nil(t, flow.Intent)
		assert.NotNil(t, flow.Context)
	})

	t.Run("returns existing flow", func(t *testing.T) {
		flow1 := manager.GetOrCreateFlow(sessionID)
		intent := &UserIntent{Action: "reorder_highlights", Confirmed: true}
		flow1.SetIntent(intent)
		flow1.Phase = PhaseExecution

		flow2 := manager.GetOrCreateFlow(sessionID)

		assert.Same(t, flow1, flow2)
		assert.Equal(t, "reorder_highlights", flow2.Intent.Action)
		assert.Equal(t, PhaseExecution, flow2.Phase)
	})
}

func TestConversationFlowManager_UpdateFlow(t *testing.T) {
	manager := NewConversationFlowManager()
	sessionID := "test-session-123"

	flow := manager.GetOrCreateFlow(sessionID)
	originalPhase := flow.Phase

	// Modify the flow
	intent := &UserIntent{Action: "analyze_highlights", Confirmed: true}
	flow.SetIntent(intent)
	flow.Phase = PhaseExecution

	// Update the flow in the manager
	manager.UpdateFlow(sessionID, flow)

	updatedFlow := manager.GetOrCreateFlow(sessionID)
	assert.Equal(t, "analyze_highlights", updatedFlow.Intent.Action)
	assert.Equal(t, PhaseExecution, updatedFlow.Phase)
	assert.True(t, updatedFlow.Intent.Confirmed)
	assert.NotEqual(t, originalPhase, updatedFlow.Phase)
}

func TestConversationFlowManager_ClearFlow(t *testing.T) {
	manager := NewConversationFlowManager()
	sessionID := "test-session-123"

	// Set up flow with some state
	flow := manager.GetOrCreateFlow(sessionID)
	intent := &UserIntent{Action: "reorder_highlights", Confirmed: true}
	flow.SetIntent(intent)
	flow.Phase = PhaseExecution
	flow.AddContext("test", "data")

	// Clear the flow
	manager.ClearFlow(sessionID)

	// Get flow again - should be a new one
	clearedFlow := manager.GetOrCreateFlow(sessionID)
	assert.Equal(t, PhaseConversation, clearedFlow.Phase)
	assert.Nil(t, clearedFlow.Intent)
	assert.NotSame(t, flow, clearedFlow) // Should be a different instance
}

func TestConversationFlow_Methods(t *testing.T) {
	manager := NewConversationFlowManager()
	sessionID := "test-session-123"

	flow := manager.GetOrCreateFlow(sessionID)

	t.Run("initial state", func(t *testing.T) {
		assert.False(t, flow.IsIntentConfirmed())
		assert.False(t, flow.ShouldExecute())
	})

	t.Run("after setting confirmed intent", func(t *testing.T) {
		intent := &UserIntent{Action: "reorder_highlights", Confirmed: true}
		flow.SetIntent(intent)
		assert.True(t, flow.IsIntentConfirmed())
		assert.True(t, flow.ShouldExecute()) // Should execute when confirmed and in conversation phase
	})

	t.Run("after moving to execution", func(t *testing.T) {
		flow.MoveToExecution()
		assert.Equal(t, PhaseExecution, flow.Phase)
		assert.True(t, flow.IsIntentConfirmed())
		assert.False(t, flow.ShouldExecute()) // Should not execute when already in execution phase
	})

	t.Run("after reset", func(t *testing.T) {
		flow.Reset()
		assert.Equal(t, PhaseConversation, flow.Phase)
		assert.Nil(t, flow.Intent)
		assert.False(t, flow.IsIntentConfirmed())
		assert.False(t, flow.ShouldExecute())
	})
}

func TestConversationFlow_ContextManagement(t *testing.T) {
	manager := NewConversationFlowManager()
	sessionID := "test-session-123"

	flow := manager.GetOrCreateFlow(sessionID)

	t.Run("add and get context", func(t *testing.T) {
		highlights := []interface{}{
			map[string]interface{}{"id": "h1", "start": 10.0},
			map[string]interface{}{"id": "h2", "start": 20.0},
		}

		flow.AddContext("highlights", highlights)
		flow.AddContext("metadata", "test metadata")

		retrievedHighlights, exists := flow.GetContext("highlights")
		assert.True(t, exists)
		assert.Equal(t, highlights, retrievedHighlights)

		retrievedMetadata, exists := flow.GetContext("metadata")
		assert.True(t, exists)
		assert.Equal(t, "test metadata", retrievedMetadata)

		// Non-existent key should return false
		_, exists = flow.GetContext("nonexistent")
		assert.False(t, exists)
	})

	t.Run("overwrite context", func(t *testing.T) {
		flow.AddContext("test", "original")
		retrievedValue, exists := flow.GetContext("test")
		assert.True(t, exists)
		assert.Equal(t, "original", retrievedValue)

		flow.AddContext("test", "updated")
		retrievedValue, exists = flow.GetContext("test")
		assert.True(t, exists)
		assert.Equal(t, "updated", retrievedValue)
	})
}

func TestConversationFlow_ToJSON(t *testing.T) {
	manager := NewConversationFlowManager()
	sessionID := "test-session-123"

	flow := manager.GetOrCreateFlow(sessionID)
	intent := &UserIntent{
		Action:    "reorder_highlights",
		Confirmed: true,
		Parameters: map[string]interface{}{
			"order": "chronological",
		},
	}
	flow.SetIntent(intent)
	flow.Phase = PhaseExecution
	flow.AddContext("highlights", []string{"h1", "h2", "h3"})
	flow.AddContext("order", "chronological")

	jsonData := flow.ToJSON()
	assert.NotEmpty(t, jsonData)

	// Parse JSON to verify structure
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &parsed)
	require.NoError(t, err)

	assert.Equal(t, sessionID, parsed["sessionId"])
	assert.Equal(t, string(PhaseExecution), parsed["phase"])

	intent_data, ok := parsed["intent"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "reorder_highlights", intent_data["action"])
	assert.Equal(t, true, intent_data["confirmed"])

	context, ok := parsed["context"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, context, "highlights")
	assert.Contains(t, context, "order")
}

func TestParseUserIntent(t *testing.T) {
	t.Run("valid intent JSON", func(t *testing.T) {
		intentJSON := `{
			"action": "reorder_highlights",
			"confirmed": true,
			"parameters": {"order": "chronological"},
			"description": "Reorder highlights chronologically"
		}`

		intent, err := ParseUserIntent(intentJSON)
		require.NoError(t, err)
		assert.Equal(t, "reorder_highlights", intent.Action)
		assert.True(t, intent.Confirmed)
		assert.Equal(t, "Reorder highlights chronologically", intent.Description)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		intentJSON := `{invalid json}`

		_, err := ParseUserIntent(intentJSON)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse user intent")
	})
}

func TestValidateUserIntent(t *testing.T) {
	t.Run("valid intent", func(t *testing.T) {
		intent := &UserIntent{
			Action:    "reorder_highlights",
			Confirmed: true,
		}

		err := ValidateUserIntent(intent)
		assert.NoError(t, err)
	})

	t.Run("nil intent", func(t *testing.T) {
		err := ValidateUserIntent(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "intent cannot be nil")
	})

	t.Run("empty action", func(t *testing.T) {
		intent := &UserIntent{
			Action:    "",
			Confirmed: true,
		}

		err := ValidateUserIntent(intent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "intent action cannot be empty")
	})

	t.Run("not confirmed", func(t *testing.T) {
		intent := &UserIntent{
			Action:    "reorder_highlights",
			Confirmed: false,
		}

		err := ValidateUserIntent(intent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "intent must be confirmed before execution")
	})
}

func TestConversationFlow_Integration(t *testing.T) {
	manager := NewConversationFlowManager()
	sessionID := "integration-test-session"

	t.Run("complete flow lifecycle", func(t *testing.T) {
		// Step 1: Initial state
		flow := manager.GetOrCreateFlow(sessionID)
		assert.Equal(t, PhaseConversation, flow.Phase)
		assert.False(t, flow.IsIntentConfirmed())

		// Step 2: Set intent
		intent := &UserIntent{
			Action:    "reorder_highlights",
			Confirmed: true,
		}
		flow.SetIntent(intent)
		assert.Equal(t, "reorder_highlights", flow.Intent.Action)
		assert.True(t, flow.IsIntentConfirmed())

		// Step 3: Add context data
		highlights := []interface{}{
			map[string]interface{}{"id": "h1", "text": "intro", "start": 0.0},
			map[string]interface{}{"id": "h2", "text": "main", "start": 30.0},
			map[string]interface{}{"id": "h3", "text": "outro", "start": 60.0},
		}
		flow.AddContext("highlights", highlights)
		flow.AddContext("reorderType", "importance")

		// Step 4: Move to execution
		assert.True(t, flow.ShouldExecute())
		flow.MoveToExecution()
		assert.Equal(t, PhaseExecution, flow.Phase)

		// Step 5: Verify context preserved
		retrievedHighlights, exists := flow.GetContext("highlights")
		assert.True(t, exists)
		assert.Equal(t, highlights, retrievedHighlights)

		// Step 6: Complete and reset
		flow.Reset()
		assert.Equal(t, PhaseConversation, flow.Phase)
		assert.Nil(t, flow.Intent)
		assert.False(t, flow.IsIntentConfirmed())
	})

	t.Run("JSON serialization throughout flow", func(t *testing.T) {
		flow := manager.GetOrCreateFlow("json-test-session")
		
		// Test at each phase
		phases := []ConversationPhase{PhaseConversation, PhaseExecution}
		
		for _, phase := range phases {
			flow.Phase = phase
			intent := &UserIntent{Action: "test_intent", Confirmed: true}
			flow.SetIntent(intent)
			flow.AddContext("phase", string(phase))
			
			jsonData := flow.ToJSON()
			assert.NotEmpty(t, jsonData)
			
			var parsed map[string]interface{}
			err := json.Unmarshal([]byte(jsonData), &parsed)
			require.NoError(t, err)
			assert.Equal(t, string(phase), parsed["phase"])
		}
	})
}