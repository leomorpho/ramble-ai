package chatbot

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	
	"ramble-ai/goapp/ai"
)

// callOpenRouterAPI makes the actual API call to OpenRouter
func (s *ChatbotService) callOpenRouterAPI(apiKey string, request map[string]interface{}) (map[string]interface{}, error) {
	log.Println("🤖 [LLM REQUEST] Starting OpenRouter API call")

	// Log request details (without sensitive data)
	model := "unknown"
	if m, ok := request["model"].(string); ok {
		model = m
	}

	messageCount := 0
	if messages, ok := request["messages"].([]map[string]interface{}); ok {
		messageCount = len(messages)
		log.Printf("🤖 [LLM REQUEST] Model: %s, Messages: %d", model, messageCount)

		// Log each message with full content
		for i, msg := range messages {
			role := "unknown"
			if r, ok := msg["role"].(string); ok {
				role = r
			}

			content := "empty"
			if c, ok := msg["content"].(string); ok {
				content = c
			}

			// Determine message type for logging
			messageType := "USER MESSAGE"
			if role == "assistant" {
				messageType = "LLM RESPONSE"
			} else if role == "system" {
				messageType = "SYSTEM MESSAGE"
			}

			log.Printf("\n========== 🤖 [%s] Message %d (%s) ==========\n%s\n========== END MESSAGE ==========\n",
				messageType, i+1, role, content)
		}
	}

	// Log function tools if present
	if tools, ok := request["tools"].([]map[string]interface{}); ok {
		log.Printf("🤖 [LLM REQUEST] Functions available: %d", len(tools))
		for i, tool := range tools {
			if funcInfo, ok := tool["function"].(map[string]interface{}); ok {
				if name, ok := funcInfo["name"].(string); ok {
					log.Printf("🤖 [LLM REQUEST] Function %d: %s", i+1, name)
				}
			}
		}
	}

	// Log tool choice if specified
	if toolChoice, ok := request["tool_choice"]; ok {
		log.Printf("🤖 [LLM REQUEST] Tool choice: %v", toolChoice)
	}

	// Update model for CoreAI service (model is already extracted above)
	if model == "unknown" || model == "" {
		model = "anthropic/claude-3.5-sonnet" // Default model
	}

	// Validate that the request can be marshaled (for compatibility with existing tests)
	_, err := json.Marshal(request)
	if err != nil {
		log.Printf("🤖 [LLM ERROR] Failed to marshal request: %v", err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// For chatbot, we need to construct messages differently
	// The request already contains the messages in the right format
	messages, ok := request["messages"].([]map[string]interface{})
	if !ok {
		log.Printf("🤖 [LLM ERROR] Invalid messages format in request")
		return nil, fmt.Errorf("invalid messages format in request")
	}

	// Build system and user prompts from messages
	var systemPrompt, userPrompt string
	for _, msg := range messages {
		role, _ := msg["role"].(string)
		content, _ := msg["content"].(string)
		
		if role == "system" {
			if systemPrompt != "" {
				systemPrompt += "\n\n" + content
			} else {
				systemPrompt = content
			}
		} else if role == "user" {
			if userPrompt != "" {
				userPrompt += "\n\nUser: " + content
			} else {
				userPrompt = content
			}
		} else if role == "assistant" {
			userPrompt += "\n\nAssistant: " + content
		}
	}

	// Use CoreAI service for the API call (keeping old implementation for chatbot)
	coreAI := ai.NewCoreAIService(s.client, s.ctx)
	
	aiRequest := &ai.TextProcessingRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Model:        model,
		TaskType:     "chat",
		Context:      map[string]interface{}{"originalRequest": request},
	}

	log.Printf("🤖 [LLM REQUEST] Using CoreAI service, Model: %s", model)
	
	// Log partial API key for debugging
	partialKey := apiKey
	if len(apiKey) > 20 {
		partialKey = apiKey[:4] + "..."
	}
	log.Printf("🤖 [LLM REQUEST] Using API key: %s", partialKey)

	startTime := time.Now()
	rawResult, err := coreAI.ProcessText(aiRequest, apiKey)
	if err != nil {
		duration := time.Since(startTime)
		log.Printf("🤖 [LLM ERROR] CoreAI request failed after %.2f seconds: %v", duration.Seconds(), err)
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	// Parse the raw OpenRouter response to get structured result
	result, err := ai.ParseTextResponse(rawResult, aiRequest.TaskType)
	if err != nil {
		duration := time.Since(startTime)
		log.Printf("🤖 [LLM ERROR] Failed to parse AI response after %.2f seconds: %v", duration.Seconds(), err)
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	duration := time.Since(startTime)
	log.Printf("🤖 [LLM RESPONSE] Received response in %.2f seconds", duration.Seconds())

	// Convert result back to OpenRouter format for compatibility
	openRouterResp := map[string]interface{}{
		"choices": []interface{}{
			map[string]interface{}{
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": result.Content,
				},
			},
		},
	}

	log.Printf("\n========== 🤖 [LLM RAW RESPONSE] ==========\n%s\n========== END RAW RESPONSE ==========\n", result.Content)

	// Log usage information if available
	if usage, ok := openRouterResp["usage"].(map[string]interface{}); ok {
		log.Printf("🤖 [LLM RESPONSE] Token usage: %v", usage)
	}

	// Extract first choice content and tool calls
	choices, ok := openRouterResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		log.Printf("🤖 [LLM ERROR] No choices in response")
		return nil, fmt.Errorf("no choices in response")
	}

	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		log.Printf("🤖 [LLM ERROR] Invalid choice format")
		return nil, fmt.Errorf("invalid choice format")
	}

	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		log.Printf("🤖 [LLM ERROR] Invalid message format")
		return nil, fmt.Errorf("invalid message format")
	}

	// Log response content
	if content, ok := message["content"].(string); ok && content != "" {
		log.Printf("\n========== 🤖 [LLM RESPONSE CONTENT] ==========\n%s\n========== END RESPONSE CONTENT ==========\n", content)
	} else {
		log.Printf("🤖 [LLM RESPONSE] Content: (empty or null)")
	}

	// Log function calls if present
	if toolCalls, ok := message["tool_calls"].([]interface{}); ok {
		log.Printf("🤖 [LLM RESPONSE] Function calls: %d", len(toolCalls))
		for i, toolCall := range toolCalls {
			if tcMap, ok := toolCall.(map[string]interface{}); ok {
				if function, ok := tcMap["function"].(map[string]interface{}); ok {
					if name, ok := function["name"].(string); ok {
						log.Printf("🤖 [LLM RESPONSE] Function call %d: %s", i+1, name)
						if arguments, ok := function["arguments"].(string); ok {
							// Log full arguments for debugging
							log.Printf("🤖 [LLM RESPONSE] Function %s arguments (full): %s", name, arguments)
						}
					}
				}
			}
		}
	}

	log.Println("🤖 [LLM RESPONSE] OpenRouter API call completed successfully")
	return message, nil
}
