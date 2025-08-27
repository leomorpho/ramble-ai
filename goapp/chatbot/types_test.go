package chatbot

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatMessage(t *testing.T) {
	now := time.Now()
	msg := ChatMessage{
		ID:        "msg123",
		Role:      "user",
		Content:   "Hello world",
		Timestamp: now,
		Hidden:    "secret context",
	}

	t.Run("JSON serialization excludes hidden field", func(t *testing.T) {
		data, err := json.Marshal(msg)
		require.NoError(t, err)

		// Parse JSON to verify hidden field is not included
		var parsed map[string]interface{}
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Equal(t, "msg123", parsed["id"])
		assert.Equal(t, "user", parsed["role"])
		assert.Equal(t, "Hello world", parsed["content"])
		_, hasHidden := parsed["hidden"]
		assert.False(t, hasHidden, "hidden field should not be serialized")
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonStr := `{
			"id": "msg456",
			"role": "assistant",
			"content": "Hi there!",
			"timestamp": "2024-01-01T12:00:00Z"
		}`

		var msg ChatMessage
		err := json.Unmarshal([]byte(jsonStr), &msg)
		require.NoError(t, err)

		assert.Equal(t, "msg456", msg.ID)
		assert.Equal(t, "assistant", msg.Role)
		assert.Equal(t, "Hi there!", msg.Content)
		assert.Empty(t, msg.Hidden) // Should be empty since not in JSON
	})
}

func TestChatSession(t *testing.T) {
	now := time.Now()
	session := ChatSession{
		ID:         "session123",
		SessionID:  "sess-abc-123",
		ProjectID:  456,
		EndpointID: "endpoint1",
		Messages: []ChatMessage{
			{
				ID:        "msg1",
				Role:      "user",
				Content:   "Hello",
				Timestamp: now,
			},
			{
				ID:        "msg2",
				Role:      "assistant",
				Content:   "Hi there!",
				Timestamp: now.Add(time.Second),
			},
		},
		CreatedAt: now,
		UpdatedAt: now.Add(time.Minute),
	}

	t.Run("JSON serialization", func(t *testing.T) {
		data, err := json.Marshal(session)
		require.NoError(t, err)

		var parsed map[string]interface{}
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Equal(t, "session123", parsed["id"])
		assert.Equal(t, "sess-abc-123", parsed["sessionId"])
		assert.Equal(t, float64(456), parsed["projectId"])
		assert.Equal(t, "endpoint1", parsed["endpointId"])

		messages, ok := parsed["messages"].([]interface{})
		require.True(t, ok)
		assert.Len(t, messages, 2)
	})
}

func TestChatRequest(t *testing.T) {
	req := ChatRequest{
		ProjectID:  123,
		EndpointID: "test-endpoint",
		Message:    "Hello AI",
		SessionID:  "session-123",
		ContextData: map[string]interface{}{
			"highlights": []string{"h1", "h2"},
			"metadata":   "test",
		},
		Model:               "gpt-4",
		EnableFunctionCalls: true,
		Mode:                "chat",
	}

	t.Run("JSON serialization and deserialization", func(t *testing.T) {
		data, err := json.Marshal(req)
		require.NoError(t, err)

		var parsed ChatRequest
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Equal(t, req.ProjectID, parsed.ProjectID)
		assert.Equal(t, req.EndpointID, parsed.EndpointID)
		assert.Equal(t, req.Message, parsed.Message)
		assert.Equal(t, req.SessionID, parsed.SessionID)
		assert.Equal(t, req.Model, parsed.Model)
		assert.Equal(t, req.EnableFunctionCalls, parsed.EnableFunctionCalls)
		assert.Equal(t, req.Mode, parsed.Mode)
		assert.NotNil(t, parsed.ContextData)
	})

	t.Run("optional fields", func(t *testing.T) {
		minimalReq := ChatRequest{
			ProjectID:  123,
			EndpointID: "test",
			Message:    "Hello",
			Model:      "gpt-4",
		}

		data, err := json.Marshal(minimalReq)
		require.NoError(t, err)

		var parsed ChatRequest
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Equal(t, 123, parsed.ProjectID)
		assert.Equal(t, "test", parsed.EndpointID)
		assert.Equal(t, "Hello", parsed.Message)
		assert.Equal(t, "gpt-4", parsed.Model)
		assert.Empty(t, parsed.SessionID)
		assert.False(t, parsed.EnableFunctionCalls)
		assert.Empty(t, parsed.Mode)
	})
}

func TestChatResponse(t *testing.T) {
	resp := ChatResponse{
		SessionID: "session-123",
		MessageID: "msg-456",
		Message:   "Hello there!",
		Model:     "gpt-4",
		Success:   true,
		FunctionResults: []FunctionExecutionResult{
			{
				FunctionName: "test_function",
				Success:      true,
				Result:       "Function executed successfully",
			},
		},
	}

	t.Run("successful response serialization", func(t *testing.T) {
		data, err := json.Marshal(resp)
		require.NoError(t, err)

		var parsed ChatResponse
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Equal(t, resp.SessionID, parsed.SessionID)
		assert.Equal(t, resp.MessageID, parsed.MessageID)
		assert.Equal(t, resp.Message, parsed.Message)
		assert.Equal(t, resp.Model, parsed.Model)
		assert.True(t, parsed.Success)
		assert.Empty(t, parsed.Error)
		assert.Len(t, parsed.FunctionResults, 1)
		assert.Equal(t, "test_function", parsed.FunctionResults[0].FunctionName)
	})

	t.Run("error response", func(t *testing.T) {
		errorResp := ChatResponse{
			SessionID: "session-123",
			MessageID: "msg-456",
			Success:   false,
			Error:     "Something went wrong",
		}

		data, err := json.Marshal(errorResp)
		require.NoError(t, err)

		var parsed ChatResponse
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.False(t, parsed.Success)
		assert.Equal(t, "Something went wrong", parsed.Error)
		assert.Empty(t, parsed.FunctionResults)
	})
}

func TestFunctionExecutionResult(t *testing.T) {
	result := FunctionExecutionResult{
		FunctionName: "reorder_highlights",
		Success:      true,
		Result:       "Highlights reordered successfully",
		Error:        "",
	}

	t.Run("successful function result", func(t *testing.T) {
		data, err := json.Marshal(result)
		require.NoError(t, err)

		var parsed FunctionExecutionResult
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Equal(t, "reorder_highlights", parsed.FunctionName)
		assert.True(t, parsed.Success)
		assert.Equal(t, "Highlights reordered successfully", parsed.Result)
		assert.Empty(t, parsed.Error)
	})

	t.Run("failed function result", func(t *testing.T) {
		failedResult := FunctionExecutionResult{
			FunctionName: "invalid_function",
			Success:      false,
			Error:        "Function not found",
		}

		data, err := json.Marshal(failedResult)
		require.NoError(t, err)

		var parsed FunctionExecutionResult
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Equal(t, "invalid_function", parsed.FunctionName)
		assert.False(t, parsed.Success)
		assert.Equal(t, "Function not found", parsed.Error)
	})
}

func TestFunctionDefinition(t *testing.T) {
	funcDef := FunctionDefinition{
		Name:        "test_function",
		Description: "A test function",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param1": map[string]interface{}{
					"type":        "string",
					"description": "First parameter",
				},
				"param2": map[string]interface{}{
					"type":        "number",
					"description": "Second parameter",
				},
			},
			"required": []string{"param1"},
		},
	}

	t.Run("function definition serialization", func(t *testing.T) {
		data, err := json.Marshal(funcDef)
		require.NoError(t, err)

		var parsed FunctionDefinition
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Equal(t, "test_function", parsed.Name)
		assert.Equal(t, "A test function", parsed.Description)
		assert.NotNil(t, parsed.Parameters)

		// Verify parameters structure
		assert.Equal(t, "object", parsed.Parameters["type"])

		properties, ok := parsed.Parameters["properties"].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, properties, "param1")
		assert.Contains(t, properties, "param2")
	})
}

