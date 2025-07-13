package chatbot

import (
	"fmt"
	"log"
	"strings"
)

// ExecutionAgent handles precise MCP function calling with progress updates
type ExecutionAgent struct {
	endpointID string
	registry   *MCPRegistry
}

// NewExecutionAgent creates a new execution agent
func NewExecutionAgent(endpointID string, registry *MCPRegistry) *ExecutionAgent {
	return &ExecutionAgent{
		endpointID: endpointID,
		registry:   registry,
	}
}

// ExecuteIntentStructured executes a confirmed user intent using structured JSON approach
func (ea *ExecutionAgent) ExecuteIntentStructured(intent *UserIntent, projectID int, chatService *ChatbotService, broadcaster *ProgressBroadcaster, getAPIKey func() (string, error)) (*ExecutionResult, error) {
	// Validate intent
	if err := ValidateUserIntent(intent); err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Invalid intent: %v", err),
		}, nil
	}
	
	broadcaster.UpdateProgress("initializing", "Preparing structured execution...")
	
	// Get API key
	apiKey, err := getAPIKey()
	if err != nil || apiKey == "" {
		return &ExecutionResult{
			Success: false,
			Error:   "OpenRouter API key not configured",
		}, nil
	}
	
	// Build structured input
	structuredInput, err := ea.buildStructuredInput(intent, projectID, chatService)
	if err != nil {
		broadcaster.UpdateProgress("error", "Failed to prepare structured input")
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to build structured input: %v", err),
		}, nil
	}
	
	broadcaster.UpdateProgress("analyzing", "Processing your request with structured approach...")
	
	// Build execution prompt using structured templates
	executionPrompt, err := BuildStructuredExecutionPrompt(structuredInput)
	if err != nil {
		broadcaster.UpdateProgress("error", "Failed to build execution prompt")
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to build prompt: %v", err),
		}, nil
	}
	
	// Create OpenRouter request WITHOUT tools - just pure text response
	openRouterReq := map[string]interface{}{
		"model": "anthropic/claude-sonnet-4",
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": executionPrompt,
			},
		},
		"temperature": 0.3, // Lower temperature for precise, structured output
		"max_tokens":  4000,
	}
	
	broadcaster.UpdateProgress("processing", ea.getProgressMessageForAction(intent.Action))
	
	// Call OpenRouter API
	aiResponse, err := chatService.callOpenRouterAPI(apiKey, openRouterReq)
	if err != nil {
		broadcaster.UpdateProgress("error", "AI processing failed")
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("AI API call failed: %v", err),
		}, nil
	}
	
	// Extract content from response
	content, ok := aiResponse["content"].(string)
	if !ok {
		broadcaster.UpdateProgress("error", "Invalid AI response format")
		return &ExecutionResult{
			Success: false,
			Error:   "AI response missing content",
		}, nil
	}
	
	broadcaster.UpdateProgress("parsing", "Parsing structured response...")
	
	// Parse structured output
	structuredOutput, err := ParseStructuredExecutionOutput(content)
	if err != nil {
		broadcaster.UpdateProgress("error", "Failed to parse AI response")
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse structured output: %v", err),
		}, nil
	}
	
	// Validate the output
	originalHighlightCount := len(structuredInput.HighlightMap)
	err = ValidateStructuredOutput(structuredOutput, originalHighlightCount)
	if err != nil {
		broadcaster.UpdateProgress("error", "Invalid response from AI")
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Output validation failed: %v", err),
		}, nil
	}
	
	broadcaster.UpdateProgress("applying", "Applying changes to your project...")
	
	// Apply the changes using the existing update function
	if chatService.updateOrderFunc != nil {
		err = chatService.updateOrderFunc(projectID, structuredOutput.NewOrder)
		if err != nil {
			broadcaster.UpdateProgress("error", "Failed to apply changes")
			return &ExecutionResult{
				Success: false,
				Error:   fmt.Sprintf("Failed to apply changes: %v", err),
			}, nil
		}
	}
	
	broadcaster.UpdateProgress("completed", "Successfully completed your request!")
	
	// Create successful result
	result := &ExecutionResult{
		Success:     true,
		Summary:     ea.generateSuccessSummaryFromStructured(intent, structuredOutput),
		Intent:      intent,
		AIReasoning: structuredOutput.Reasoning,
	}
	
	return result, nil
}

// buildStructuredInput creates structured input for the execution agent
func (ea *ExecutionAgent) buildStructuredInput(intent *UserIntent, projectID int, chatService *ChatbotService) (*StructuredExecutionInput, error) {
	// Get project highlights
	projectHighlights, err := chatService.highlightService.GetProjectHighlights(projectID)
	if err != nil {
		return nil, err
	}
	
	// Build highlight map
	highlightMap := make(map[string]string)
	for _, ph := range projectHighlights {
		for _, h := range ph.Highlights {
			highlightMap[h.ID] = h.Text
		}
	}
	
	// Get current order if needed
	var currentOrder []interface{}
	useCurrentOrder := false
	if use, ok := intent.Parameters["use_current_order"].(bool); ok && use {
		useCurrentOrder = true
		currentOrder, err = chatService.highlightService.GetProjectHighlightOrderWithTitles(projectID)
		if err != nil {
			log.Printf("Failed to get current order: %v", err)
			// Continue without current order
			useCurrentOrder = false
		}
	}
	
	// Determine intent name for structured execution
	intentName := intent.Action
	if intentName == "" {
		intentName = "reorder" // default
	}
	
	// Build user goals
	userGoals := []string{intent.PrimaryGoal}
	userGoals = append(userGoals, intent.SecondaryGoals...)
	
	// Build additional context
	additionalContext := intent.Context
	if intent.Reasoning != "" {
		additionalContext += ". LLM Understanding: " + intent.Reasoning
	}
	
	structuredInput := &StructuredExecutionInput{
		Intent:            intentName,
		HighlightMap:      highlightMap,
		CurrentOrder:      currentOrder,
		UseCurrentOrder:   useCurrentOrder,
		UserGoals:         userGoals,
		AdditionalContext: additionalContext,
	}
	
	return structuredInput, nil
}

// generateSuccessSummaryFromStructured creates summary from structured output
func (ea *ExecutionAgent) generateSuccessSummaryFromStructured(intent *UserIntent, output *StructuredExecutionOutput) string {
	var summaryBuilder strings.Builder
	
	summaryBuilder.WriteString("✅ **Success!** ")
	
	switch intent.Action {
	case "reorder":
		summaryBuilder.WriteString(fmt.Sprintf("Reorganized your highlights into %d sections for better engagement and flow.", output.SectionCount))
		
	case "improve_hook":
		summaryBuilder.WriteString("Improved your opening section for stronger hook and better viewer retention.")
		
	case "improve_conclusion":
		summaryBuilder.WriteString("Enhanced your conclusion for more powerful ending and better viewer satisfaction.")
		
	case "analyze":
		summaryBuilder.WriteString("Completed detailed analysis of your content structure and flow.")
		
	default:
		summaryBuilder.WriteString(fmt.Sprintf("Completed %s operation successfully.", intent.Action))
	}
	
	// Add key changes
	if len(output.Changes) > 0 {
		summaryBuilder.WriteString(fmt.Sprintf("\n\n**Key Changes:** %s", strings.Join(output.Changes, ", ")))
	}
	
	// Add reasoning if available
	if output.Reasoning != "" {
		summaryBuilder.WriteString(fmt.Sprintf("\n\n**Reasoning:** %s", output.Reasoning))
	}
	
	return summaryBuilder.String()
}

// LEGACY METHOD - keeping for backward compatibility but marking as deprecated
// ExecuteIntent executes a confirmed user intent using MCP functions
func (ea *ExecutionAgent) ExecuteIntent(intent *UserIntent, projectID int, chatService *ChatbotService, broadcaster *ProgressBroadcaster, getAPIKey func() (string, error)) (*ExecutionResult, error) {
	// Validate intent
	if err := ValidateUserIntent(intent); err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Invalid intent: %v", err),
		}, nil
	}
	
	broadcaster.UpdateProgress("initializing", "Preparing to execute your request...")
	
	// Get API key
	apiKey, err := getAPIKey()
	if err != nil || apiKey == "" {
		return &ExecutionResult{
			Success: false,
			Error:   "OpenRouter API key not configured",
		}, nil
	}
	
	// Build execution context based on user preferences
	context, err := ea.buildExecutionContext(intent, projectID, chatService)
	if err != nil {
		broadcaster.UpdateProgress("error", "Failed to prepare execution context")
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to build context: %v", err),
		}, nil
	}
	
	broadcaster.UpdateProgress("analyzing", "Analyzing your request and available options...")
	
	// Build system prompt for execution
	systemPrompt := ea.buildExecutionSystemPrompt(intent, context)
	
	// Get MCP functions for this endpoint
	tools, err := ea.registry.GetFunctionsForEndpoint(ea.endpointID)
	if err != nil {
		broadcaster.UpdateProgress("error", "Failed to load available functions")
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to get functions: %v", err),
		}, nil
	}
	
	// Create execution prompt based on intent
	executionPrompt := ea.buildExecutionPrompt(intent)
	
	// Create OpenRouter request with MCP functions
	openRouterReq := map[string]interface{}{
		"model": "anthropic/claude-sonnet-4",
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": executionPrompt,
			},
		},
		"tools":       tools,
		"tool_choice": "auto", // Let the AI decide when to call functions
		"temperature": 0.3,    // Lower temperature for precise execution
		"max_tokens":  4000,
	}
	
	broadcaster.UpdateProgress("processing", ea.getProgressMessageForAction(intent.Action))
	
	// Call OpenRouter API
	aiResponse, err := chatService.callOpenRouterAPI(apiKey, openRouterReq)
	if err != nil {
		broadcaster.UpdateProgress("error", "AI processing failed")
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("AI API call failed: %v", err),
		}, nil
	}
	
	// Process function calls if present
	var functionResults []FunctionExecutionResult
	var actionsPerformed []string
	
	if toolCalls, ok := aiResponse["tool_calls"].([]interface{}); ok && len(toolCalls) > 0 {
		broadcaster.UpdateProgress("executing", "Applying changes to your project...")
		
		for _, toolCallInterface := range toolCalls {
			if toolCall, ok := toolCallInterface.(map[string]interface{}); ok {
				result := chatService.executeMCPFunctionCall(toolCall, projectID, ea.endpointID)
				functionResults = append(functionResults, result)
				
				if result.Success {
					actionsPerformed = append(actionsPerformed, result.FunctionName)
				}
			}
		}
	}
	
	// Generate result
	result := &ExecutionResult{
		Success:          len(functionResults) > 0 && allSuccessful(functionResults),
		FunctionResults:  functionResults,
		ActionsPerformed: actionsPerformed,
		Intent:           intent,
	}
	
	if result.Success {
		broadcaster.UpdateProgress("completed", "Successfully completed your request!")
		result.Summary = ea.generateSuccessSummary(intent, functionResults)
	} else {
		broadcaster.UpdateProgress("error", "Some actions failed to complete")
		result.Error = ea.generateErrorSummary(functionResults)
	}
	
	// Extract AI reasoning if available
	if content, ok := aiResponse["content"].(string); ok {
		result.AIReasoning = content
	}
	
	return result, nil
}

// buildExecutionContext builds context based on user preferences
func (ea *ExecutionAgent) buildExecutionContext(intent *UserIntent, projectID int, chatService *ChatbotService) (string, error) {
	// Check if user wants current order included
	useCurrentOrder := false
	if use, ok := intent.Parameters["use_current_order"].(bool); ok {
		useCurrentOrder = use
	}
	
	var contextBuilder strings.Builder
	
	// Always include highlight reference
	contextBuilder.WriteString("Available highlights for your project:\n\n")
	
	// Get project highlights
	projectHighlights, err := chatService.highlightService.GetProjectHighlights(projectID)
	if err != nil {
		return "", err
	}
	
	// Build highlight map
	allIDs := []string{}
	for _, ph := range projectHighlights {
		for _, h := range ph.Highlights {
			text := h.Text
			if len(text) > 150 {
				text = text[:150] + "..."
			}
			contextBuilder.WriteString(fmt.Sprintf("- %s: \"%s\"\n", h.ID, text))
			allIDs = append(allIDs, h.ID)
		}
	}
	
	contextBuilder.WriteString(fmt.Sprintf("\nTotal: %d highlights\n", len(allIDs)))
	
	// Conditionally include current order
	if useCurrentOrder {
		contextBuilder.WriteString("\nCurrent highlight order (to use as starting point):\n")
		currentOrder, err := chatService.highlightService.GetProjectHighlightOrderWithTitles(projectID)
		if err != nil {
			log.Printf("Failed to get current order: %v", err)
			contextBuilder.WriteString("Current order unavailable.\n")
		} else {
			for i, item := range currentOrder {
				switch v := item.(type) {
				case string:
					contextBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, v))
				case map[string]interface{}:
					if title, ok := v["title"].(string); ok {
						contextBuilder.WriteString(fmt.Sprintf("%d. [SECTION] %s\n", i+1, title))
					} else {
						contextBuilder.WriteString(fmt.Sprintf("%d. [SECTION]\n", i+1))
					}
				}
			}
		}
	} else {
		contextBuilder.WriteString("\nUser prefers to start fresh (not using current order).\n")
	}
	
	return contextBuilder.String(), nil
}

// buildExecutionSystemPrompt creates system prompt for execution agent
func (ea *ExecutionAgent) buildExecutionSystemPrompt(intent *UserIntent, context string) string {
	return fmt.Sprintf(`You are a precise execution agent specializing in YouTube content optimization. Your job is to execute the user's confirmed intent using the available MCP functions.

CRITICAL FIRST STEP: Based on the Action field below, you MUST call the corresponding function:
- Action="reorder" → Call reorder_highlights function
- Action="analyze" → Call analyze_highlights function
- Action="reset" → Call reset_to_original function
- Action="get_current_order" → Call get_current_order function
- Action="apply_suggestion" → Call apply_ai_suggestion function

USER'S CONFIRMED INTENT:
Action: %s ← THIS DETERMINES WHICH FUNCTION TO CALL
Primary Goal: %s
Secondary Goals: %s
LLM Understanding: %s
User Context: %s
Description: %s
User Preferences: %s

PROJECT CONTEXT:
%s

EXECUTION GUIDELINES:
1. **Use the available functions** to fulfill the user's intent precisely based on the LLM understanding
2. **Focus on primary goal** - this is what the user mainly wants
3. **Incorporate secondary goals** - use these to enhance your approach and reasoning
4. **Include ALL highlight IDs** in any reordering operations
5. **Provide analytical reasoning** when secondary goals include analysis or reasoning
6. **Follow the LLM's understanding** of why the user wants this

CRITICAL: FUNCTION MAPPING (you MUST call the correct function):
- **Action: "reorder"** → CALL reorder_highlights function
- **Action: "analyze"** → CALL analyze_highlights function  
- **Action: "reset"** → CALL reset_to_original function
- **Action: "get_current_order"** → CALL get_current_order function
- **Action: "apply_suggestion"** → CALL apply_ai_suggestion function

IMPORTANT: The user's intent.Action determines which function to call. If Action="reorder", you MUST call reorder_highlights, NOT analyze_highlights.

SECTIONING REQUIREMENTS (when sectioning=true in preferences):
- **Create logical sections** with descriptive, engaging titles
- **Section flow**: Hook/Intro → Content Sections → Conclusion
- **Section titles**: Should be specific and engaging (e.g., "The Problem", "Why This Matters", "The Solution", "Key Benefits", "Final Thoughts")
- **Content grouping**: Group related highlights within each section
- **Engagement optimization**: Structure for maximum viewer retention

ANALYTICAL REASONING (when provide_reasoning=true in preferences):
- **Explain your decisions** - why you chose this specific order
- **Reference engagement principles** - how this improves viewer retention
- **Connect to user goals** - show how this achieves their primary and secondary goals
- **Provide narrative justification** - explain the story flow you created

SECTION OBJECT FORMAT:
When creating sections, use: {"type": "N", "title": "Section Title"}

EXAMPLE REORDER WITH SECTIONS AND REASONING:
[
  {"type": "N", "title": "Hook: The Big Problem"},
  "highlight_id_1",  // Most attention-grabbing statement
  "highlight_id_2",  // Problem amplification
  {"type": "N", "title": "Why This Matters"},
  "highlight_id_3",  // Personal relevance
  {"type": "N", "title": "The Solution"},
  "highlight_id_4",  // Main solution reveal
  "highlight_id_5",  // Solution benefits
  {"type": "N", "title": "Conclusion: Take Action"}
]

REMEMBER: You MUST call the function that corresponds to Action="%s":
- If Action="reorder" → Call reorder_highlights
- If Action="analyze" → Call analyze_highlights

Execute the user's confirmed intent now, incorporating their goals and providing reasoning when requested.`, 
		intent.Action,
		intent.Action,
		intent.PrimaryGoal,
		formatSecondaryGoals(intent.SecondaryGoals),
		intent.Reasoning,
		intent.Context,
		intent.Description,
		formatUserPreferences(intent.UserPreferences),
		context)
}

// buildExecutionPrompt creates the execution prompt based on intent
func (ea *ExecutionAgent) buildExecutionPrompt(intent *UserIntent) string {
	basePrompt := fmt.Sprintf("Execute the user's primary goal: %s", intent.PrimaryGoal)
	
	if len(intent.SecondaryGoals) > 0 {
		basePrompt += fmt.Sprintf(". Also incorporate these secondary goals: %s", strings.Join(intent.SecondaryGoals, ", "))
	}
	
	if intent.Context != "" {
		basePrompt += fmt.Sprintf(". Important context: %s", intent.Context)
	}
	
	switch intent.Action {
	case "reorder":
		basePrompt += ". CALL reorder_highlights function to reorganize the highlights."
		if specific, ok := intent.Parameters["specific_request"].(string); ok {
			basePrompt += fmt.Sprintf(" Specific request: %s", specific)
		}
		return basePrompt
		
	case "analyze":
		basePrompt += ". CALL analyze_highlights function to analyze the content."
		return basePrompt
		
	case "reset":
		return "CALL reset_to_original function to reset the highlights to their original chronological order."
		
	case "get_current_order":
		return "CALL get_current_order function to get and display the current highlight order."
		
	case "apply_suggestion":
		return "CALL apply_ai_suggestion function to apply the previously generated AI suggestion."
		
	default:
		return basePrompt + fmt.Sprintf(". Action: %s", intent.Action)
	}
}

// getProgressMessageForAction returns appropriate progress message for action
func (ea *ExecutionAgent) getProgressMessageForAction(action string) string {
	switch action {
	case "reorder":
		return "Optimizing highlight arrangement..."
	case "analyze":
		return "Analyzing content structure and themes..."
	case "reset":
		return "Restoring original highlight order..."
	case "get_current_order":
		return "Retrieving current highlight arrangement..."
	case "apply_suggestion":
		return "Applying optimization suggestions..."
	default:
		return "Processing your request..."
	}
}

// generateSuccessSummary creates a human-readable summary of successful execution
func (ea *ExecutionAgent) generateSuccessSummary(intent *UserIntent, results []FunctionExecutionResult) string {
	var summaryBuilder strings.Builder
	
	summaryBuilder.WriteString("✅ **Success!** ")
	
	switch intent.Action {
	case "reorder":
		summaryBuilder.WriteString("Your highlights have been reordered for better flow.")
		if goal, ok := intent.UserPreferences["optimization_goal"].(string); ok {
			summaryBuilder.WriteString(fmt.Sprintf(" Optimized for: %s.", goal))
		}
		
	case "analyze":
		summaryBuilder.WriteString("Content analysis completed.")
		
	case "reset":
		summaryBuilder.WriteString("Highlights reset to original chronological order.")
		
	case "get_current_order":
		summaryBuilder.WriteString("Current highlight order retrieved.")
		
	case "apply_suggestion":
		summaryBuilder.WriteString("AI suggestions applied successfully.")
		
	default:
		summaryBuilder.WriteString(fmt.Sprintf("%s action completed.", intent.Action))
	}
	
	// Add reasoning from function results if available
	for _, result := range results {
		if result.Success && result.Result != nil {
			if resultMap, ok := result.Result.(map[string]interface{}); ok {
				if reason, ok := resultMap["reason"].(string); ok && reason != "" {
					summaryBuilder.WriteString(fmt.Sprintf("\n\n**Reasoning:** %s", reason))
					break
				}
			}
		}
	}
	
	return summaryBuilder.String()
}

// generateErrorSummary creates error summary from failed function results
func (ea *ExecutionAgent) generateErrorSummary(results []FunctionExecutionResult) string {
	var errors []string
	for _, result := range results {
		if !result.Success {
			errors = append(errors, result.Error)
		}
	}
	
	if len(errors) == 0 {
		return "Unknown error occurred during execution"
	}
	
	return fmt.Sprintf("Execution failed: %s", strings.Join(errors, "; "))
}

// formatUserPreferences formats user preferences for display
func formatUserPreferences(prefs map[string]interface{}) string {
	if len(prefs) == 0 {
		return "None specified"
	}
	
	var parts []string
	for key, value := range prefs {
		parts = append(parts, fmt.Sprintf("%s: %v", key, value))
	}
	
	return strings.Join(parts, ", ")
}

// formatSecondaryGoals formats secondary goals for display
func formatSecondaryGoals(goals []string) string {
	if len(goals) == 0 {
		return "None specified"
	}
	
	return strings.Join(goals, ", ")
}

// allSuccessful checks if all function results were successful
func allSuccessful(results []FunctionExecutionResult) bool {
	for _, result := range results {
		if !result.Success {
			return false
		}
	}
	return len(results) > 0
}

// ExecutionResult represents the result of an execution
type ExecutionResult struct {
	Success          bool                       `json:"success"`
	Summary          string                     `json:"summary,omitempty"`
	Error            string                     `json:"error,omitempty"`
	FunctionResults  []FunctionExecutionResult  `json:"functionResults,omitempty"`
	ActionsPerformed []string                   `json:"actionsPerformed,omitempty"`
	AIReasoning      string                     `json:"aiReasoning,omitempty"`
	Intent           *UserIntent               `json:"intent,omitempty"`
}