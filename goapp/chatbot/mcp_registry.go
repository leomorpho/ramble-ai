package chatbot

import (
	"fmt"
	"log"
)

// ContextBuilder interface for building endpoint-specific context
type ContextBuilder interface {
	BuildContext(projectID int, service *ChatbotService) (string, error)
	GetContextDescription() string
}

// MCPFunction represents a function that can be called by the LLM
type MCPFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Executor    FunctionExecutor       `json:"-"` // Not serialized
}

// EndpointMCPConfig defines the MCP configuration for a specific endpoint
type EndpointMCPConfig struct {
	EndpointID        string         `json:"endpointId"`
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	ContextBuilder    ContextBuilder `json:"-"` // Not serialized
	Functions         []MCPFunction  `json:"functions"`
	SystemPrompt      string         `json:"systemPrompt"`
	RequiresFunctions bool           `json:"requiresFunctions"`
	DefaultModel      string         `json:"defaultModel"`
}

// MCPRegistry manages MCP configurations for all endpoints
type MCPRegistry struct {
	configs map[string]*EndpointMCPConfig
}

// NewMCPRegistry creates a new MCP registry
func NewMCPRegistry() *MCPRegistry {
	registry := &MCPRegistry{
		configs: make(map[string]*EndpointMCPConfig),
	}

	// Register all endpoint configurations
	registry.registerAllEndpoints()

	return registry
}

// registerAllEndpoints registers MCP configurations for all supported endpoints
func (r *MCPRegistry) registerAllEndpoints() {
	// Register highlight ordering endpoint
	r.RegisterEndpoint(&EndpointMCPConfig{
		EndpointID:        "highlight_ordering",
		Name:              "Highlight Ordering Assistant",
		Description:       "Help with organizing and reordering highlights for better flow",
		ContextBuilder:    &HighlightOrderingContextBuilder{},
		RequiresFunctions: true,
		DefaultModel:      "anthropic/claude-sonnet-4",
		SystemPrompt: `You are an expert video editor assistant specializing in highlight organization. 

When a user asks to reorder highlights:

1. FIRST ask if they want to use the current highlight order as a starting point or create a completely new arrangement
2. If they want to modify the current order, reference the "Current highlight order" section provided in the context
3. If they want to start fresh, focus only on the available highlights and their content

WHEN READY TO REORDER:
- Identify ALL highlight IDs from the "Available highlights" section (they start with "highlight_")
- Create an optimal order considering narrative flow, engagement, and content themes
- Call reorder_highlights with the complete new_order array including ALL highlight IDs
- Provide a clear reason for your reordering decisions

Example reorder call format:
reorder_highlights({
  "new_order": ["highlight_1752254837026_c873lgzyc", "highlight_1752254847396_0hbn26kzn", {"type": "N", "title": "Section Title"}, "highlight_1752254873052_l5o8f0uhg"],
  "reason": "Created strong hook, built narrative tension, and ended with emotional payoff"
})

CRITICAL: Include ALL highlight IDs from the context in your new_order array. Missing any highlight ID will break the reordering.`,
		Functions: []MCPFunction{
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
				Executor: func(args map[string]interface{}, projectID int, service *ChatbotService) (interface{}, error) {
					return service.executeReorderHighlights(args, projectID, service)
				},
			},
			{
				Name:        "get_current_order",
				Description: "Get the current highlight order for the project",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
				Executor: func(args map[string]interface{}, projectID int, service *ChatbotService) (interface{}, error) {
					return service.executeGetCurrentOrder(args, projectID, service)
				},
			},
			{
				Name:        "analyze_highlights",
				Description: "Analyze highlights for content, themes, and structure recommendations",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
				Executor: func(args map[string]interface{}, projectID int, service *ChatbotService) (interface{}, error) {
					return service.executeAnalyzeHighlights(args, projectID, service)
				},
			},
			{
				Name:        "apply_ai_suggestion",
				Description: "Apply a previously generated AI reorder suggestion",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
				Executor: func(args map[string]interface{}, projectID int, service *ChatbotService) (interface{}, error) {
					return service.executeApplyAISuggestion(args, projectID, service)
				},
			},
			{
				Name:        "reset_to_original",
				Description: "Reset highlights to their original order",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
				Executor: func(args map[string]interface{}, projectID int, service *ChatbotService) (interface{}, error) {
					return service.executeResetToOriginal(args, projectID, service)
				},
			},
		},
	})

	// Register other endpoints with placeholder configs
	r.RegisterEndpoint(&EndpointMCPConfig{
		EndpointID:        "highlight_suggestions",
		Name:              "Highlight Suggestions Assistant",
		Description:       "Get AI suggestions for creating engaging highlights",
		ContextBuilder:    &GenericContextBuilder{},
		RequiresFunctions: false,
		DefaultModel:      "anthropic/claude-sonnet-4",
		SystemPrompt:      "You are an expert at identifying compelling moments in video content. Help suggest highlights that will engage viewers.",
		Functions:         []MCPFunction{}, // No MCP functions yet
	})

	r.RegisterEndpoint(&EndpointMCPConfig{
		EndpointID:        "content_analysis",
		Name:              "Content Analysis Assistant",
		Description:       "Analyze video content for insights and recommendations",
		ContextBuilder:    &GenericContextBuilder{},
		RequiresFunctions: false,
		DefaultModel:      "google/gemini-2.0-flash-001",
		SystemPrompt:      "You are a content analysis expert. Help analyze video content for themes, key messages, and audience engagement opportunities.",
		Functions:         []MCPFunction{}, // No MCP functions yet
	})

	r.RegisterEndpoint(&EndpointMCPConfig{
		EndpointID:        "export_optimization",
		Name:              "Export Optimization Assistant",
		Description:       "Optimize export settings and final video production",
		ContextBuilder:    &GenericContextBuilder{},
		RequiresFunctions: false,
		DefaultModel:      "anthropic/claude-3.5-haiku-20241022",
		SystemPrompt:      "You are a video production expert. Help optimize export settings and final video production for different platforms and audiences.",
		Functions:         []MCPFunction{}, // No MCP functions yet
	})
}

// RegisterEndpoint registers an MCP configuration for an endpoint
func (r *MCPRegistry) RegisterEndpoint(config *EndpointMCPConfig) {
	if config == nil {
		log.Printf("Warning: Attempted to register nil endpoint config")
		return
	}

	if config.EndpointID == "" {
		log.Printf("Warning: Attempted to register endpoint with empty ID")
		return
	}

	r.configs[config.EndpointID] = config
	log.Printf("Registered MCP endpoint: %s with %d functions", config.EndpointID, len(config.Functions))
}

// GetEndpointConfig returns the MCP configuration for an endpoint
func (r *MCPRegistry) GetEndpointConfig(endpointID string) (*EndpointMCPConfig, bool) {
	config, exists := r.configs[endpointID]
	return config, exists
}

// GetAllEndpoints returns all registered endpoint IDs
func (r *MCPRegistry) GetAllEndpoints() []string {
	endpoints := make([]string, 0, len(r.configs))
	for endpointID := range r.configs {
		endpoints = append(endpoints, endpointID)
	}
	return endpoints
}

// BuildContextForEndpoint builds context for a specific endpoint
func (r *MCPRegistry) BuildContextForEndpoint(endpointID string, projectID int, service *ChatbotService) (string, error) {
	config, exists := r.GetEndpointConfig(endpointID)
	if !exists {
		return "", fmt.Errorf("endpoint %s not found in MCP registry", endpointID)
	}

	if config.ContextBuilder == nil {
		return "", fmt.Errorf("no context builder configured for endpoint %s", endpointID)
	}

	return config.ContextBuilder.BuildContext(projectID, service)
}

// GetFunctionsForEndpoint returns the MCP functions for an endpoint in OpenRouter tool format
func (r *MCPRegistry) GetFunctionsForEndpoint(endpointID string) ([]map[string]interface{}, error) {
	config, exists := r.GetEndpointConfig(endpointID)
	if !exists {
		return nil, fmt.Errorf("endpoint %s not found in MCP registry", endpointID)
	}

	var tools []map[string]interface{}
	for _, mcpFunc := range config.Functions {
		tool := map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        mcpFunc.Name,
				"description": mcpFunc.Description,
				"parameters":  mcpFunc.Parameters,
			},
		}
		tools = append(tools, tool)
	}

	return tools, nil
}

// ExecuteFunction executes an MCP function for an endpoint
func (r *MCPRegistry) ExecuteFunction(endpointID, functionName string, args map[string]interface{}, projectID int, service *ChatbotService) (FunctionExecutionResult, error) {
	config, exists := r.GetEndpointConfig(endpointID)
	if !exists {
		return FunctionExecutionResult{
			FunctionName: functionName,
			Success:      false,
			Error:        fmt.Sprintf("endpoint %s not found", endpointID),
		}, fmt.Errorf("endpoint %s not found in MCP registry", endpointID)
	}

	// Find the function
	for _, mcpFunc := range config.Functions {
		if mcpFunc.Name == functionName {
			if mcpFunc.Executor == nil {
				return FunctionExecutionResult{
					FunctionName: functionName,
					Success:      false,
					Error:        "function executor not configured",
				}, fmt.Errorf("function executor not configured for %s", functionName)
			}

			// Execute the function
			result, err := mcpFunc.Executor(args, projectID, service)
			if err != nil {
				return FunctionExecutionResult{
					FunctionName: functionName,
					Success:      false,
					Error:        err.Error(),
				}, err
			}

			return FunctionExecutionResult{
				FunctionName: functionName,
				Success:      true,
				Result:       result,
			}, nil
		}
	}

	return FunctionExecutionResult{
		FunctionName: functionName,
		Success:      false,
		Error:        "function not found",
	}, fmt.Errorf("function %s not found for endpoint %s", functionName, endpointID)
}

// SupportsActions returns whether an endpoint supports MCP function calls
func (r *MCPRegistry) SupportsActions(endpointID string) bool {
	config, exists := r.GetEndpointConfig(endpointID)
	if !exists {
		return false
	}
	return config.RequiresFunctions || len(config.Functions) > 0
}
