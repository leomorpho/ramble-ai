package chatbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// callOpenRouterAPI makes the actual API call to OpenRouter
func (s *ChatbotService) callOpenRouterAPI(apiKey string, request map[string]interface{}) (map[string]interface{}, error) {
	// Convert request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/yourusername/video-app")
	req.Header.Set("X-Title", "Video Highlight Assistant")
	
	// Make request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	// Parse response
	var openRouterResp map[string]interface{}
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	// Check for errors
	if errorInfo, ok := openRouterResp["error"]; ok {
		return nil, fmt.Errorf("OpenRouter API error: %v", errorInfo)
	}
	
	// Extract first choice content and tool calls
	choices, ok := openRouterResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}
	
	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid choice format")
	}
	
	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid message format")
	}
	
	return message, nil
}