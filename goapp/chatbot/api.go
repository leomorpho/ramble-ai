package chatbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
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
		
		// Log each message with truncated content
		for i, msg := range messages {
			role := "unknown"
			if r, ok := msg["role"].(string); ok {
				role = r
			}
			
			content := "empty"
			if c, ok := msg["content"].(string); ok {
				// Truncate long content for readability
				if len(c) > 500 {
					content = c[:500] + "... [TRUNCATED]"
				} else {
					content = c
				}
			}
			
			log.Printf("🤖 [LLM REQUEST] Message %d (%s): %s", i+1, role, content)
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
	
	// Convert request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Printf("🤖 [LLM ERROR] Failed to marshal request: %v", err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	log.Printf("🤖 [LLM REQUEST] Request size: %d bytes", len(jsonData))
	
	// Create HTTP request
	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("🤖 [LLM ERROR] Failed to create HTTP request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/yourusername/video-app")
	req.Header.Set("X-Title", "Video Highlight Assistant")
	
	// Log partial API key for debugging (first 20 chars)
	partialKey := apiKey
	if len(apiKey) > 20 {
		partialKey = apiKey[:20] + "..."
	}
	log.Printf("🤖 [LLM REQUEST] Using API key: %s", partialKey)
	
	log.Println("🤖 [LLM REQUEST] Sending HTTP request to OpenRouter...")
	startTime := time.Now()
	
	// Make request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("🤖 [LLM ERROR] HTTP request failed after %.2f seconds: %v", time.Since(startTime).Seconds(), err)
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	duration := time.Since(startTime)
	log.Printf("🤖 [LLM RESPONSE] Received response in %.2f seconds, Status: %s", duration.Seconds(), resp.Status)
	
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("🤖 [LLM ERROR] Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	log.Printf("🤖 [LLM RESPONSE] Response size: %d bytes", len(body))
	
	// Parse response
	var openRouterResp map[string]interface{}
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		log.Printf("🤖 [LLM ERROR] Failed to parse JSON response: %v", err)
		log.Printf("🤖 [LLM ERROR] Raw response: %s", string(body))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	// Log the full raw response for debugging (temporarily)
	bodyStr := string(body)
	if len(bodyStr) > 1000 {
		log.Printf("🤖 [LLM DEBUG] Raw response (first 1000 chars): %s... [TRUNCATED]", bodyStr[:1000])
	} else {
		log.Printf("🤖 [LLM DEBUG] Raw response: %s", bodyStr)
	}
	
	// Check for errors
	if errorInfo, ok := openRouterResp["error"]; ok {
		log.Printf("🤖 [LLM ERROR] OpenRouter API error: %v", errorInfo)
		return nil, fmt.Errorf("OpenRouter API error: %v", errorInfo)
	}
	
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
		if len(content) > 500 {
			log.Printf("🤖 [LLM RESPONSE] Content (first 500 chars): %s... [TRUNCATED]", content[:500])
		} else {
			log.Printf("🤖 [LLM RESPONSE] Content: %s", content)
		}
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