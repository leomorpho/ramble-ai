package chatbot

import (
	"encoding/json"
	"fmt"
	"strings"
)

// StructuredExecutionInput defines the standardized input for execution agent
type StructuredExecutionInput struct {
	Intent               string                 `json:"intent"`               // "reorder", "improve_hook", "improve_conclusion", "analyze"
	HighlightMap         map[string]string      `json:"highlightMap"`         // highlight_id -> text
	CurrentOrder         []interface{}          `json:"currentOrder"`         // current order (if needed)
	UseCurrentOrder      bool                   `json:"useCurrentOrder"`      // whether to use current order as starting point
	UserGoals           []string               `json:"userGoals"`            // user's specific goals
	AdditionalContext   string                 `json:"additionalContext"`    // any additional context
}

// StructuredExecutionOutput defines the standardized output from execution agent
type StructuredExecutionOutput struct {
	Success       bool          `json:"success"`
	NewOrder      []interface{} `json:"newOrder"`      // array of highlight IDs and section objects
	Reasoning     string        `json:"reasoning"`     // explanation of decisions
	SectionCount  int           `json:"sectionCount"`  // number of sections created
	Changes       []string      `json:"changes"`       // list of key changes made
	Error         string        `json:"error,omitempty"` // error message if failed
}

// IntentTemplate defines the template for a specific intent
type IntentTemplate struct {
	IntentName    string
	Description   string
	Instructions  string
	OutputFormat  string
	Examples      string
}

// GetIntentTemplate returns the template for a specific intent
func GetIntentTemplate(intent string) *IntentTemplate {
	templates := map[string]*IntentTemplate{
		"reorder": {
			IntentName:   "reorder",
			Description:  "Reorder highlights for optimal engagement and narrative flow",
			Instructions: `REORDER INSTRUCTIONS:
- You can move ANY highlight to ANY position - complete freedom
- Organize into logical sections with engaging titles  
- Section flow: Hook/Intro → Content Sections → Conclusion
- Group related highlights within sections
- Optimize for YouTube viewer retention
- Use section objects: {"type": "N", "title": "Section Title"}
- Include ALL highlight IDs - missing any ID will break the system`,
			OutputFormat: `Return a JSON object with this EXACT structure:
{
  "success": true,
  "newOrder": [
    {"type": "N", "title": "Hook: Grab Attention"},
    "highlight_id_1",
    "highlight_id_2",
    {"type": "N", "title": "Main Content"}, 
    "highlight_id_3",
    {"type": "N", "title": "Conclusion: Strong Finish"}
  ],
  "reasoning": "Detailed explanation of your reordering decisions and why this improves engagement",
  "sectionCount": 3,
  "changes": ["Created engaging hook section", "Grouped related concepts", "Built narrative tension", "Added strong conclusion"]
}`,
			Examples: `EXAMPLE INPUT: 3 highlights about learning
EXAMPLE OUTPUT:
{
  "success": true,
  "newOrder": [
    {"type": "N", "title": "Hook: The Learning Problem"},
    "highlight_123",
    {"type": "N", "title": "The Solution"}, 
    "highlight_456",
    "highlight_789",
    {"type": "N", "title": "Take Action"}
  ],
  "reasoning": "Started with the problem statement to hook viewers, then provided the solution with supporting evidence, ending with a call to action for maximum engagement.",
  "sectionCount": 3,
  "changes": ["Problem-solution structure", "Logical flow", "Strong hook and conclusion"]
}`,
		},
		
		"improve_hook": {
			IntentName:   "improve_hook",
			Description:  "Improve the opening section to create a stronger hook",
			Instructions: `HOOK IMPROVEMENT INSTRUCTIONS:
- Focus on the first 1-3 highlights to create maximum impact
- Use the most attention-grabbing content first
- Create curiosity, urgency, or emotional connection
- Ensure first 3 seconds grab viewer attention
- You can reorder any highlights to create the best hook
- Include ALL highlight IDs in the output`,
			OutputFormat: `Return a JSON object with this EXACT structure:
{
  "success": true,
  "newOrder": [complete reordered list with improved hook],
  "reasoning": "Explanation of hook improvements and why this grabs attention better",
  "sectionCount": number,
  "changes": ["Moved strongest statement to start", "Created curiosity gap", "etc."]
}`,
			Examples: `Focus on creating the strongest possible opening while maintaining overall flow.`,
		},
		
		"improve_conclusion": {
			IntentName:   "improve_conclusion",
			Description:  "Improve the ending section for stronger finish",
			Instructions: `CONCLUSION IMPROVEMENT INSTRUCTIONS:
- Focus on the last 1-3 highlights for maximum impact ending
- Use the most powerful, memorable content for the finish
- Create strong call-to-action or emotional payoff
- Leave viewers satisfied but wanting more
- You can reorder any highlights to create the best conclusion
- Include ALL highlight IDs in the output`,
			OutputFormat: `Return a JSON object with this EXACT structure:
{
  "success": true,
  "newOrder": [complete reordered list with improved conclusion],
  "reasoning": "Explanation of conclusion improvements and why this creates stronger finish",
  "sectionCount": number,
  "changes": ["Moved most powerful statement to end", "Created satisfying payoff", "etc."]
}`,
			Examples: `Focus on creating the strongest possible ending while maintaining overall flow.`,
		},
		
		"analyze": {
			IntentName:   "analyze",
			Description:  "Analyze content structure and provide insights without reordering",
			Instructions: `ANALYSIS INSTRUCTIONS:
- Analyze the current structure and content themes
- Identify strengths and weaknesses in current flow
- Suggest potential improvements without making changes
- Do NOT reorder highlights - this is analysis only
- Keep current order intact`,
			OutputFormat: `Return a JSON object with this EXACT structure:
{
  "success": true,
  "newOrder": [exact same order as input - DO NOT CHANGE],
  "reasoning": "Detailed analysis of current structure, themes, flow, and potential improvements",
  "sectionCount": 0,
  "changes": ["Analysis only - no changes made"]
}`,
			Examples: `Provide insights and suggestions while keeping everything in the same order.`,
		},
	}
	
	return templates[intent]
}

// BuildStructuredExecutionPrompt creates a complete prompt for the execution agent
func BuildStructuredExecutionPrompt(input *StructuredExecutionInput) (string, error) {
	template := GetIntentTemplate(input.Intent)
	if template == nil {
		return "", fmt.Errorf("unknown intent: %s", input.Intent)
	}
	
	// Convert input to JSON for display
	inputJSON, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal input: %w", err)
	}
	
	// Build the complete prompt
	prompt := fmt.Sprintf(`You are a YouTube content optimization specialist. You will receive structured input and must return structured JSON output.

TASK: %s
%s

INPUT DATA:
%s

INSTRUCTIONS:
%s

REQUIRED OUTPUT FORMAT:
%s

EXAMPLES:
%s

CRITICAL REQUIREMENTS:
1. Return ONLY the JSON object - no additional text
2. Include ALL highlight IDs from the input highlightMap
3. Ensure the JSON is valid and parseable
4. Follow the exact output format specified above

EXECUTE THE TASK NOW:`, 
		template.Description,
		template.IntentName,
		string(inputJSON),
		template.Instructions,
		template.OutputFormat,
		template.Examples)
	
	return prompt, nil
}

// ParseStructuredExecutionOutput parses the LLM response into structured output
func ParseStructuredExecutionOutput(response string) (*StructuredExecutionOutput, error) {
	// Clean the response to extract JSON
	response = strings.TrimSpace(response)
	
	// Look for JSON object in the response
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")
	
	if startIdx == -1 || endIdx == -1 {
		return nil, fmt.Errorf("no JSON object found in response")
	}
	
	jsonStr := response[startIdx : endIdx+1]
	
	var output StructuredExecutionOutput
	err := json.Unmarshal([]byte(jsonStr), &output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}
	
	return &output, nil
}

// ValidateStructuredOutput validates that the output is complete and correct
func ValidateStructuredOutput(output *StructuredExecutionOutput, originalHighlightCount int) error {
	if !output.Success {
		if output.Error != "" {
			return fmt.Errorf("execution failed: %s", output.Error)
		}
		return fmt.Errorf("execution failed with no error message")
	}
	
	if len(output.NewOrder) == 0 {
		return fmt.Errorf("new order is empty")
	}
	
	// Count highlight IDs in the new order
	highlightCount := 0
	for _, item := range output.NewOrder {
		if itemStr, ok := item.(string); ok && strings.HasPrefix(itemStr, "highlight_") {
			highlightCount++
		}
	}
	
	if highlightCount != originalHighlightCount {
		return fmt.Errorf("expected %d highlights, got %d", originalHighlightCount, highlightCount)
	}
	
	return nil
}