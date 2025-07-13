package chatbot

import (
	"fmt"
	"log"
	"strings"
	"unicode/utf8"
)

// TokenCounter provides token counting utilities for different models
type TokenCounter struct {
	modelLimits map[string]int
}

// NewTokenCounter creates a new token counter with model-specific limits
func NewTokenCounter() *TokenCounter {
	return &TokenCounter{
		modelLimits: map[string]int{
			"anthropic/claude-sonnet-4":     200000, // Claude Sonnet 4 context window
			"anthropic/claude-3.5-sonnet":   200000, // Claude 3.5 Sonnet context window
			"anthropic/claude-3-haiku":      200000, // Claude 3 Haiku context window
			"openai/gpt-4o":                 128000, // GPT-4o context window
			"openai/gpt-4o-mini":           128000, // GPT-4o mini context window
			"openai/gpt-4-turbo":           128000, // GPT-4 Turbo context window
			"default":                       32000,  // Conservative default
		},
	}
}

// GetModelLimit returns the context window limit for a given model
func (tc *TokenCounter) GetModelLimit(model string) int {
	if limit, exists := tc.modelLimits[model]; exists {
		return limit
	}
	return tc.modelLimits["default"]
}

// EstimateTokens provides a rough estimate of token count for text
// This is a simplified approximation - for production use, consider integrating
// with model-specific tokenizers
func (tc *TokenCounter) EstimateTokens(text string) int {
	if text == "" {
		return 0
	}
	
	// Rough approximation: 1 token ≈ 4 characters for English text
	// This varies by model and language, but provides a reasonable estimate
	charCount := utf8.RuneCountInString(text)
	return (charCount + 3) / 4 // Round up division
}

// EstimateMessageTokens estimates tokens for a message including role overhead
func (tc *TokenCounter) EstimateMessageTokens(role, content string) int {
	// Add some overhead for message structure (role, formatting, etc.)
	roleOverhead := 5
	contentTokens := tc.EstimateTokens(content)
	return contentTokens + roleOverhead
}

// EstimateMessagesTokens calculates total tokens for a slice of messages
func (tc *TokenCounter) EstimateMessagesTokens(messages []map[string]interface{}) int {
	totalTokens := 0
	for _, msg := range messages {
		role, _ := msg["role"].(string)
		content, _ := msg["content"].(string)
		totalTokens += tc.EstimateMessageTokens(role, content)
	}
	return totalTokens
}

// ContextManager handles intelligent context preservation and trimming
type ContextManager struct {
	tokenCounter *TokenCounter
}

// NewContextManager creates a new context manager
func NewContextManager() *ContextManager {
	return &ContextManager{
		tokenCounter: NewTokenCounter(),
	}
}

// ContextWindow represents a manageable conversation context
type ContextWindow struct {
	SystemPrompt    string                   `json:"systemPrompt"`
	Messages        []map[string]interface{} `json:"messages"`
	TotalTokens     int                      `json:"totalTokens"`
	TrimmedMessages int                      `json:"trimmedMessages"`
	Summary         string                   `json:"summary,omitempty"`
}

// BuildContextWindow creates an optimized context window for a model
func (cm *ContextManager) BuildContextWindow(
	model string,
	systemPrompt string,
	chatHistory *ChatHistoryResponse,
	currentMessage string,
	reserveTokens int, // Reserve tokens for response
) (*ContextWindow, error) {

	limit := cm.tokenCounter.GetModelLimit(model)
	maxContextTokens := limit - reserveTokens
	
	// Start with system prompt
	messages := []map[string]interface{}{
		{
			"role":    "system",
			"content": systemPrompt,
		},
	}
	
	systemTokens := cm.tokenCounter.EstimateMessageTokens("system", systemPrompt)
	currentTokens := systemTokens
	
	// Add current message tokens to planning
	currentMsgTokens := cm.tokenCounter.EstimateMessageTokens("user", currentMessage)
	
	// Calculate available tokens for history
	availableForHistory := maxContextTokens - systemTokens - currentMsgTokens
	
	var historyMessages []map[string]interface{}
	var trimmedCount int
	var summary string
	
	if chatHistory != nil && len(chatHistory.Messages) > 0 {
		historyMessages, trimmedCount, summary = cm.trimHistory(
			chatHistory.Messages,
			availableForHistory,
		)
	}
	
	// Add history messages
	messages = append(messages, historyMessages...)
	currentTokens += cm.tokenCounter.EstimateMessagesTokens(historyMessages)
	
	// Add current message
	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": currentMessage,
	})
	currentTokens += currentMsgTokens
	
	return &ContextWindow{
		SystemPrompt:    systemPrompt,
		Messages:        messages,
		TotalTokens:     currentTokens,
		TrimmedMessages: trimmedCount,
		Summary:         summary,
	}, nil
}

// trimHistory intelligently trims chat history to fit within token limits
func (cm *ContextManager) trimHistory(
	messages []ChatMessage,
	maxTokens int,
) ([]map[string]interface{}, int, string) {
	
	if len(messages) == 0 {
		return []map[string]interface{}{}, 0, ""
	}
	
	// Convert to API format and estimate tokens
	var apiMessages []map[string]interface{}
	var tokenCounts []int
	
	for _, msg := range messages {
		apiMsg := map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}
		tokens := cm.tokenCounter.EstimateMessageTokens(msg.Role, msg.Content)
		
		apiMessages = append(apiMessages, apiMsg)
		tokenCounts = append(tokenCounts, tokens)
	}
	
	// Use sliding window approach: keep recent messages that fit
	totalTokens := 0
	includeFromIndex := len(apiMessages)
	
	// Work backwards from most recent messages
	for i := len(apiMessages) - 1; i >= 0; i-- {
		if totalTokens + tokenCounts[i] <= maxTokens {
			totalTokens += tokenCounts[i]
			includeFromIndex = i
		} else {
			break
		}
	}
	
	trimmedCount := includeFromIndex
	var summary string
	
	// If we trimmed messages, create a summary of what was removed
	if trimmedCount > 0 {
		summary = cm.createConversationSummary(messages[:trimmedCount])
		
		// Add summary as a system message if we have space
		summaryTokens := cm.tokenCounter.EstimateMessageTokens("system", summary)
		if summaryTokens <= maxTokens - totalTokens {
			summaryMsg := map[string]interface{}{
				"role":    "system", 
				"content": fmt.Sprintf("Previous conversation summary: %s", summary),
			}
			result := []map[string]interface{}{summaryMsg}
			result = append(result, apiMessages[includeFromIndex:]...)
			return result, trimmedCount, summary
		}
	}
	
	return apiMessages[includeFromIndex:], trimmedCount, summary
}

// createConversationSummary creates a concise summary of conversation messages
func (cm *ContextManager) createConversationSummary(messages []ChatMessage) string {
	if len(messages) == 0 {
		return ""
	}
	
	var summaryParts []string
	var userRequests []string
	var assistantActions []string
	
	for _, msg := range messages {
		content := strings.TrimSpace(msg.Content)
		if content == "" {
			continue
		}
		
		// Truncate very long messages for summary
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		
		if msg.Role == "user" {
			userRequests = append(userRequests, content)
		} else if msg.Role == "assistant" {
			// Skip JSON responses in summaries
			if !strings.Contains(content, "conversation_summary") {
				assistantActions = append(assistantActions, content)
			}
		}
	}
	
	if len(userRequests) > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("User requests: %s", 
			strings.Join(userRequests, "; ")))
	}
	
	if len(assistantActions) > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("Assistant responses: %s", 
			strings.Join(assistantActions, "; ")))
	}
	
	if len(summaryParts) == 0 {
		return fmt.Sprintf("Previous conversation with %d messages", len(messages))
	}
	
	return strings.Join(summaryParts, ". ")
}

// GetOptimalHistoryLimit calculates optimal message limit for a model
func (cm *ContextManager) GetOptimalHistoryLimit(model string, systemPromptTokens int) int {
	limit := cm.tokenCounter.GetModelLimit(model)
	
	// Reserve 25% for response and buffer, 10% for system prompt overhead
	availableForHistory := int(float64(limit) * 0.65)
	
	// Estimate messages: average 50 tokens per message (rough estimate)
	averageTokensPerMessage := 50
	optimalMessageLimit := availableForHistory / averageTokensPerMessage
	
	// Cap at reasonable limits
	if optimalMessageLimit > 200 {
		optimalMessageLimit = 200 // Don't retrieve more than 200 messages
	}
	if optimalMessageLimit < 10 {
		optimalMessageLimit = 10 // Always try to get at least 10 recent messages
	}
	
	return optimalMessageLimit
}

// LogContextUsage logs context window usage for monitoring
func (cm *ContextManager) LogContextUsage(model string, window *ContextWindow) {
	limit := cm.tokenCounter.GetModelLimit(model)
	usagePercent := float64(window.TotalTokens) / float64(limit) * 100
	
	log.Printf("Context usage for %s: %d/%d tokens (%.1f%%), trimmed %d messages", 
		model, window.TotalTokens, limit, usagePercent, window.TrimmedMessages)
	
	if window.TrimmedMessages > 0 {
		log.Printf("Conversation summary: %s", window.Summary)
	}
}