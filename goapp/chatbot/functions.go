package chatbot

import (
	"encoding/json"
	"fmt"
)

// registerFunctions registers all available functions for the LLM
func (s *ChatbotService) registerFunctions() {
	// Register highlight reordering functions
	s.functionRegistry["reorder_highlights"] = s.executeReorderHighlights
	s.functionRegistry["get_current_order"] = s.executeGetCurrentOrder
	s.functionRegistry["analyze_highlights"] = s.executeAnalyzeHighlights
	s.functionRegistry["apply_ai_suggestion"] = s.executeApplyAISuggestion
	s.functionRegistry["reset_to_original"] = s.executeResetToOriginal
	
	// Define function schemas for LLM
	s.functionDefs = []FunctionDefinition{
		{
			Name:        "reorder_highlights",
			Description: "Reorder video highlights with optional section titles",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"new_order": map[string]interface{}{
						"type":        "array",
						"description": "Array of highlight IDs and section objects in the desired order",
						"items": map[string]interface{}{
							"oneOf": []map[string]interface{}{
								{"type": "string", "description": "Highlight ID"},
								{
									"type": "object",
									"properties": map[string]interface{}{
										"type":  map[string]interface{}{"type": "string", "enum": []string{"N"}},
										"title": map[string]interface{}{"type": "string", "description": "Section title"},
									},
									"required": []string{"type"},
								},
							},
						},
					},
					"reason": map[string]interface{}{
						"type":        "string", 
						"description": "Brief explanation of why this order works better",
					},
				},
				"required": []string{"new_order"},
			},
		},
		{
			Name:        "get_current_order",
			Description: "Get the current highlight order for the project",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "analyze_highlights",
			Description: "Analyze highlights for content, themes, and structure recommendations",
			Parameters: map[string]interface{}{
				"type":       "object", 
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "apply_ai_suggestion",
			Description: "Apply a previously generated AI reorder suggestion",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "reset_to_original",
			Description: "Reset highlights to their original order",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

// buildToolDefinitions converts function definitions to OpenRouter tool format
func (s *ChatbotService) buildToolDefinitions() []map[string]interface{} {
	var tools []map[string]interface{}
	
	for _, funcDef := range s.functionDefs {
		tool := map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        funcDef.Name,
				"description": funcDef.Description,
				"parameters":  funcDef.Parameters,
			},
		}
		tools = append(tools, tool)
	}
	
	return tools
}

// executeFunctionCall executes a function call from the LLM
func (s *ChatbotService) executeFunctionCall(toolCall map[string]interface{}, projectID int) FunctionExecutionResult {
	functionInfo, ok := toolCall["function"].(map[string]interface{})
	if !ok {
		return FunctionExecutionResult{
			Success: false,
			Error:   "Invalid function call format",
		}
	}
	
	functionName, ok := functionInfo["name"].(string)
	if !ok {
		return FunctionExecutionResult{
			Success: false,
			Error:   "Function name not found",
		}
	}
	
	// Parse arguments
	var args map[string]interface{}
	if argsStr, ok := functionInfo["arguments"].(string); ok {
		if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
			return FunctionExecutionResult{
				FunctionName: functionName,
				Success:      false,
				Error:        fmt.Sprintf("Failed to parse function arguments: %v", err),
			}
		}
	}
	
	// Execute function
	executor, exists := s.functionRegistry[functionName]
	if !exists {
		return FunctionExecutionResult{
			FunctionName: functionName,
			Success:      false,
			Error:        "Function not found",
		}
	}
	
	result, err := executor(args, projectID, s)
	if err != nil {
		return FunctionExecutionResult{
			FunctionName: functionName,
			Success:      false,
			Error:        err.Error(),
		}
	}
	
	return FunctionExecutionResult{
		FunctionName: functionName,
		Success:      true,
		Result:       result,
		Message:      "Function executed successfully",
	}
}

// Function executors

// executeReorderHighlights reorders highlights based on the provided order
func (s *ChatbotService) executeReorderHighlights(args map[string]interface{}, projectID int, service *ChatbotService) (interface{}, error) {
	newOrderInterface, ok := args["new_order"]
	if !ok {
		return nil, fmt.Errorf("new_order parameter is required")
	}
	
	// Convert interface{} to []interface{}
	newOrderSlice, ok := newOrderInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("new_order must be an array")
	}
	
	// Return the new order for the caller to apply
	// Note: The actual database update will be handled by the app layer
	
	reason := ""
	if reasonInterface, ok := args["reason"]; ok {
		if reasonStr, ok := reasonInterface.(string); ok {
			reason = reasonStr
		}
	}
	
	return map[string]interface{}{
		"success":   true,
		"message":   "Highlight order prepared",
		"reason":    reason,
		"count":     len(newOrderSlice),
		"new_order": newOrderSlice,
		"apply_required": true,
	}, nil
}

// executeGetCurrentOrder gets the current highlight order
func (s *ChatbotService) executeGetCurrentOrder(args map[string]interface{}, projectID int, service *ChatbotService) (interface{}, error) {
	currentOrder, err := s.highlightService.GetProjectHighlightOrderWithTitles(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current order: %w", err)
	}
	
	// Get highlight details for context
	projectHighlights, err := s.highlightService.GetProjectHighlights(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project highlights: %w", err)
	}
	
	// Create a summary for the LLM
	highlightSummary := make(map[string]string)
	for _, ph := range projectHighlights {
		for _, highlight := range ph.Highlights {
			// Truncate long text for summary
			text := highlight.Text
			if len(text) > 100 {
				text = text[:100] + "..."
			}
			highlightSummary[highlight.ID] = text
		}
	}
	
	return map[string]interface{}{
		"current_order":     currentOrder,
		"highlight_summary": highlightSummary,
		"total_highlights":  len(highlightSummary),
	}, nil
}

// executeAnalyzeHighlights analyzes highlights for content and structure
func (s *ChatbotService) executeAnalyzeHighlights(args map[string]interface{}, projectID int, service *ChatbotService) (interface{}, error) {
	projectHighlights, err := s.highlightService.GetProjectHighlights(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project highlights: %w", err)
	}
	
	if len(projectHighlights) == 0 {
		return map[string]interface{}{
			"total_highlights": 0,
			"message":          "No highlights found for analysis",
		}, nil
	}
	
	// Analyze content
	totalLength := 0
	var allTexts []string
	highlightCount := 0
	
	for _, ph := range projectHighlights {
		for _, highlight := range ph.Highlights {
			allTexts = append(allTexts, highlight.Text)
			totalLength += len(highlight.Text)
			highlightCount++
		}
	}
	
	avgLength := 0
	if highlightCount > 0 {
		avgLength = totalLength / highlightCount
	}
	
	return map[string]interface{}{
		"total_highlights":   highlightCount,
		"total_text_length":  totalLength,
		"average_length":     avgLength,
		"highlight_texts":    allTexts[:min(10, len(allTexts))], // First 10 for context
		"analysis_complete":  true,
	}, nil
}

// executeApplyAISuggestion applies a cached AI suggestion
func (s *ChatbotService) executeApplyAISuggestion(args map[string]interface{}, projectID int, service *ChatbotService) (interface{}, error) {
	// Get cached AI suggestion
	cachedSuggestion, err := s.aiService.GetProjectAISuggestion(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached AI suggestion: %w", err)
	}
	
	if cachedSuggestion == nil || len(cachedSuggestion.Order) == 0 {
		return nil, fmt.Errorf("no cached AI suggestion found")
	}
	
	// Return the cached suggestion for the caller to apply
	// Note: The actual database update will be handled by the app layer
	
	return map[string]interface{}{
		"success":      true,
		"message":      "Cached AI suggestion prepared",
		"model_used":   cachedSuggestion.Model,
		"created_at":   cachedSuggestion.CreatedAt,
		"new_order":    cachedSuggestion.Order,
		"apply_required": true,
	}, nil
}

// executeResetToOriginal resets highlights to original order
func (s *ChatbotService) executeResetToOriginal(args map[string]interface{}, projectID int, service *ChatbotService) (interface{}, error) {
	// Get all highlights and reset to their natural order (by creation time or ID)
	projectHighlights, err := s.highlightService.GetProjectHighlights(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project highlights: %w", err)
	}
	
	// Create original order (just highlight IDs in sequence)
	var originalOrder []interface{}
	for _, ph := range projectHighlights {
		for _, highlight := range ph.Highlights {
			originalOrder = append(originalOrder, highlight.ID)
		}
	}
	
	// Return the original order for the caller to apply
	// Note: The actual database update will be handled by the app layer
	
	return map[string]interface{}{
		"success":      true,
		"message":      "Original order prepared",
		"count":        len(originalOrder),
		"new_order":    originalOrder,
		"apply_required": true,
	}, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}