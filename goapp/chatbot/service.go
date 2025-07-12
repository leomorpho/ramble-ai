package chatbot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"MYAPP/ent"
	"MYAPP/goapp/highlights"
)

// NewChatbotService creates a new chatbot service
func NewChatbotService(client *ent.Client, ctx context.Context, updateOrderFunc UpdateOrderFunc) *ChatbotService {
	s := &ChatbotService{
		client:           client,
		ctx:              ctx,
		functionRegistry: make(map[string]FunctionExecutor),
		highlightService: highlights.NewHighlightService(client, ctx),
		aiService:        highlights.NewAIService(client, ctx),
		updateOrderFunc:  updateOrderFunc,
	}
	
	// Register available functions
	s.registerFunctions()
	
	return s
}

// SendMessage handles sending a message and getting an AI response
func (s *ChatbotService) SendMessage(req ChatRequest, getAPIKey func() (string, error)) (*ChatResponse, error) {
	// Generate message ID
	messageID := fmt.Sprintf("msg_%d", time.Now().UnixNano())
	
	// Validate required fields
	if req.ProjectID == 0 {
		return &ChatResponse{
			SessionID: req.SessionID,
			MessageID: messageID,
			Success:   false,
			Error:     "Project ID is required",
		}, nil
	}
	
	// If function calling is enabled and we're in reorder mode, call LLM with functions
	if req.EnableFunctionCalls && req.Mode == "reorder" {
		return s.sendMessageWithFunctions(req, messageID, getAPIKey)
	}
	
	// Otherwise, send regular chat message
	return s.sendRegularMessage(req, messageID)
}

// sendMessageWithFunctions handles LLM requests with function calling enabled  
func (s *ChatbotService) sendMessageWithFunctions(req ChatRequest, messageID string, getAPIKey func() (string, error)) (*ChatResponse, error) {
	// Get API key
	apiKey, err := getAPIKey()
	if err != nil || apiKey == "" {
		return &ChatResponse{
			SessionID: req.SessionID,
			MessageID: messageID,
			Success:   false,
			Error:     "OpenRouter API key not configured",
		}, nil
	}
	
	// Build context for reorder mode
	context, err := s.buildReorderContext(req.ProjectID)
	if err != nil {
		log.Printf("Failed to build reorder context: %v", err)
		context = "Current highlights context unavailable."
	}
	
	// Build system prompt for reorder mode
	systemPrompt := fmt.Sprintf(`You are an expert video editor assistant helping to organize highlight segments for maximum engagement and flow.

Current project context:
%s

You have access to the following functions to help manage highlights:
- reorder_highlights: Reorder highlights with optional section titles
- get_current_order: Get the current highlight arrangement
- analyze_highlights: Analyze content for themes and structure
- apply_ai_suggestion: Apply a previously generated AI suggestion
- reset_to_original: Reset to the original highlight order

When suggesting reorders, consider:
- Content flow and narrative structure
- Balancing section lengths by text content
- Creating engaging hooks and transitions
- Maintaining viewer engagement throughout

Always explain your reasoning when reordering highlights.`, context)
	
	// Create OpenRouter request with function calling
	openRouterReq := map[string]interface{}{
		"model": req.Model,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user", 
				"content": req.Message,
			},
		},
		"tools": s.buildToolDefinitions(),
	}
	
	// Call OpenRouter API
	llmResponse, err := s.callOpenRouterAPI(apiKey, openRouterReq)
	if err != nil {
		return &ChatResponse{
			SessionID: req.SessionID,
			MessageID: messageID,
			Success:   false,
			Error:     fmt.Sprintf("LLM API call failed: %v", err),
		}, nil
	}
	
	// Process LLM response and execute any function calls
	response := &ChatResponse{
		SessionID: req.SessionID,
		MessageID: messageID,
		Success:   true,
	}
	
	// Extract message and tool calls from LLM response
	if content, ok := llmResponse["content"].(string); ok {
		response.Message = content
	}
	
	// Process function calls if present
	if toolCalls, ok := llmResponse["tool_calls"].([]interface{}); ok {
		var functionResults []FunctionExecutionResult
		
		for _, toolCall := range toolCalls {
			if toolCallMap, ok := toolCall.(map[string]interface{}); ok {
				result := s.executeFunctionCall(toolCallMap, req.ProjectID)
				functionResults = append(functionResults, result)
				
				// If the function result requires applying an order, do it now
				if result.Success && result.Result != nil {
					if resultMap, ok := result.Result.(map[string]interface{}); ok {
						if applyRequired, ok := resultMap["apply_required"].(bool); ok && applyRequired {
							if newOrder, ok := resultMap["new_order"].([]interface{}); ok && s.updateOrderFunc != nil {
								if err := s.updateOrderFunc(req.ProjectID, newOrder); err != nil {
									result.Success = false
									result.Error = fmt.Sprintf("Failed to apply order: %v", err)
								} else {
									// Update the result message to indicate successful application
									result.Message = "Function executed and order applied successfully"
									if resultMap["message"] != nil {
										resultMap["message"] = fmt.Sprintf("%s (Applied to database)", resultMap["message"])
									}
								}
							}
						}
					}
				}
			}
		}
		
		response.FunctionResults = functionResults
	}
	
	return response, nil
}

// sendRegularMessage handles regular chat without function calling
func (s *ChatbotService) sendRegularMessage(req ChatRequest, messageID string) (*ChatResponse, error) {
	// For now, return a simple response
	response := &ChatResponse{
		SessionID: req.SessionID,
		MessageID: messageID,
		Message:   fmt.Sprintf("Regular chat response to: %s", req.Message),
		Success:   true,
	}
	
	return response, nil
}

// buildReorderContext builds context about current highlights for the LLM
func (s *ChatbotService) buildReorderContext(projectID int) (string, error) {
	// Get current order
	currentOrder, err := s.highlightService.GetProjectHighlightOrderWithTitles(projectID)
	if err != nil {
		return "", err
	}
	
	// Get highlight summaries
	projectHighlights, err := s.highlightService.GetProjectHighlights(projectID)
	if err != nil {
		return "", err
	}
	
	// Build context string
	var contextBuilder strings.Builder
	contextBuilder.WriteString(fmt.Sprintf("Total highlights: %d\n\n", len(projectHighlights)))
	
	contextBuilder.WriteString("Current highlight order:\n")
	for i, item := range currentOrder {
		switch v := item.(type) {
		case string:
			// Find highlight content
			for _, ph := range projectHighlights {
				for _, h := range ph.Highlights {
					if h.ID == v {
						text := h.Text
						if len(text) > 80 {
							text = text[:80] + "..."
						}
						contextBuilder.WriteString(fmt.Sprintf("%d. %s: \"%s\"\n", i+1, h.ID, text))
						break
					}
				}
			}
		case map[string]interface{}:
			if title, ok := v["title"].(string); ok {
				contextBuilder.WriteString(fmt.Sprintf("%d. [SECTION: %s]\n", i+1, title))
			}
		}
	}
	
	return contextBuilder.String(), nil
}

// GetChatHistory retrieves chat history for a project/endpoint
func (s *ChatbotService) GetChatHistory(projectID int, endpointID string) (*ChatHistoryResponse, error) {
	// TODO: Implement database retrieval
	// For now, return empty history
	return &ChatHistoryResponse{
		SessionID: fmt.Sprintf("session_%d_%s", projectID, endpointID),
		Messages:  []ChatMessage{},
	}, nil
}

// ClearChatHistory clears chat history for a project/endpoint
func (s *ChatbotService) ClearChatHistory(projectID int, endpointID string) error {
	// TODO: Implement database clearing
	return nil
}

// generatePlaceholderResponse creates a placeholder AI response for testing
func (s *ChatbotService) generatePlaceholderResponse(endpointID, userMessage string) string {
	switch endpointID {
	case "highlight_ordering":
		return fmt.Sprintf("I can help you organize your highlights! You mentioned: \"%s\". Here are some suggestions for better highlight ordering:\n\n1. Start with a strong hook\n2. Build narrative tension\n3. Include emotional peaks\n4. End with a satisfying conclusion\n\nWould you like me to analyze your current highlight order?", userMessage)
	case "highlight_suggestions":
		return fmt.Sprintf("Great question about highlight suggestions: \"%s\". I can help you find the most engaging moments in your video content. Let me analyze the patterns and suggest optimal highlight points.", userMessage)
	case "export_assistance":
		return fmt.Sprintf("I'd be happy to help with your export needs: \"%s\". I can guide you through different export formats, quality settings, and optimization strategies for various platforms.", userMessage)
	default:
		return fmt.Sprintf("Thanks for your message: \"%s\". I'm here to help with video editing, highlight organization, and content optimization. What would you like to work on?", userMessage)
	}
}