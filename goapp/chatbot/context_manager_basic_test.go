package chatbot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenCounter_Basic(t *testing.T) {
	counter := NewTokenCounter()

	t.Run("creates counter with default limits", func(t *testing.T) {
		assert.NotNil(t, counter)
		assert.NotNil(t, counter.modelLimits)
	})

	t.Run("gets model limits", func(t *testing.T) {
		assert.Equal(t, 200000, counter.GetModelLimit("anthropic/claude-sonnet-4"))
		assert.Equal(t, 128000, counter.GetModelLimit("openai/gpt-4o"))
		assert.Equal(t, 32000, counter.GetModelLimit("unknown-model"))
	})

	t.Run("estimates tokens", func(t *testing.T) {
		assert.Equal(t, 0, counter.EstimateTokens(""))
		assert.Equal(t, 2, counter.EstimateTokens("hello"))  // "hello" is 5 chars, (5+3)/4 = 2
		assert.Equal(t, 3, counter.EstimateTokens("hello world"))  // "hello world" is 11 chars, (11+3)/4 = 3
	})

	t.Run("estimates message tokens", func(t *testing.T) {
		tokens := counter.EstimateMessageTokens("user", "hello world")
		assert.Greater(t, tokens, 2) // Should be content + overhead
	})

	t.Run("estimates messages tokens", func(t *testing.T) {
		messages := []map[string]interface{}{
			{"role": "user", "content": "hello"},
			{"role": "assistant", "content": "hi there"},
		}
		tokens := counter.EstimateMessagesTokens(messages)
		assert.Greater(t, tokens, 5) // Should include all messages + overhead
	})
}

func TestNewContextManager_Basic(t *testing.T) {
	manager := NewContextManager()

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.tokenCounter)
}

func TestContextManager_GetOptimalHistoryLimit(t *testing.T) {
	manager := NewContextManager()

	limit := manager.GetOptimalHistoryLimit("anthropic/claude-sonnet-4", 1000)
	assert.Greater(t, limit, 50) // Should be reasonable
	assert.Less(t, limit, manager.tokenCounter.GetModelLimit("anthropic/claude-sonnet-4"))
}

func TestContextManager_LogContextUsage(t *testing.T) {
	manager := NewContextManager()

	// This should not panic
	window := &ContextWindow{
		SystemPrompt:    "Test system prompt",
		Messages:        []map[string]interface{}{{"role": "user", "content": "test"}},
		TotalTokens:     1500,
		TrimmedMessages: 5,
		Summary:         "Previous conversation summary",
	}

	manager.LogContextUsage("anthropic/claude-sonnet-4", window)
}