package chatbot

import (
	"fmt"
	"strings"
)

// PrepareExecutorPrompt builds a complete prompt for the executor by calling MCP functions and appending data
func (s *ChatbotService) PrepareExecutorPrompt(summary *ConversationSummary, projectID int) (string, error) {
	// Start with base template for the intent
	basePrompt := getExecutorTemplate(summary.Intent)

	// Always get project highlights
	projectHighlights, err := s.highlightService.GetProjectHighlights(projectID)
	if err != nil {
		return "", fmt.Errorf("failed to get project highlights: %w", err)
	}

	// Build highlight map section
	basePrompt += "\n\nAVAILABLE HIGHLIGHTS:\n"
	highlightCount := 0
	for _, ph := range projectHighlights {
		for _, h := range ph.Highlights {
			basePrompt += fmt.Sprintf("- %s: \"%s\"\n", h.ID, h.Text)
			highlightCount++
		}
	}
	basePrompt += fmt.Sprintf("\nTotal highlights: %d (ALL must be included in new order)\n", highlightCount)

	// Conditionally get current order if user wants it
	if summary.UserWantsCurrentOrder {
		currentOrder, err := s.highlightService.GetProjectHighlightOrderWithTitles(projectID)
		if err != nil {
			// Log error but continue without current order
			basePrompt += "\nCURRENT ORDER: (unavailable due to error)\n"
		} else {
			basePrompt += "\n\nCURRENT ORDER (use as starting point):\n"
			for i, item := range currentOrder {
				basePrompt += fmt.Sprintf("%d. %s\n", i+1, formatOrderItem(item))
			}
		}
	} else {
		basePrompt += "\n\nCURRENT ORDER: User prefers to start fresh (not using current order)\n"
	}

	// Add user goals and context
	if len(summary.OptimizationGoals) > 0 {
		basePrompt += fmt.Sprintf("\n\nUSER OPTIMIZATION GOALS: %s\n", strings.Join(summary.OptimizationGoals, ", "))
	}

	if len(summary.SpecificRequests) > 0 {
		basePrompt += fmt.Sprintf("\nSPECIFIC USER REQUESTS: %s\n", strings.Join(summary.SpecificRequests, ", "))
	}

	if summary.UserContext != "" {
		basePrompt += fmt.Sprintf("\nUSER CONTEXT: %s\n", summary.UserContext)
	}

	// Add output format requirements
	basePrompt += getOutputFormatRequirements(summary.Intent)

	return basePrompt, nil
}

// getExecutorTemplate returns the base template for a specific intent
func getExecutorTemplate(intent string) string {
	switch intent {
	case "reorder":
		return `You are a YouTube content optimization specialist. Your task is to REORDER highlights for maximum engagement.

REORDER INSTRUCTIONS:
- You can move ANY highlight to ANY position - complete freedom to reorganize
- Organize into logical sections with engaging titles using: {"type": "N", "title": "Section Title"}
- Section flow: Hook/Intro → Content Sections → Conclusion
- Group related highlights within sections
- Optimize for YouTube viewer retention and engagement
- Focus on creating strong narrative flow`

	case "improve_hook":
		return `You are a YouTube content optimization specialist. Your task is to IMPROVE THE HOOK by reordering highlights.

HOOK IMPROVEMENT INSTRUCTIONS:
- Focus on the first 1-3 highlights to create maximum impact opening
- Use the most attention-grabbing content first
- Create curiosity, urgency, or emotional connection
- Ensure first 3 seconds grab viewer attention
- You can reorder any highlights to create the best hook`

	case "improve_conclusion":
		return `You are a YouTube content optimization specialist. Your task is to IMPROVE THE CONCLUSION by reordering highlights.

CONCLUSION IMPROVEMENT INSTRUCTIONS:
- Focus on the last 1-3 highlights for maximum impact ending
- Use the most powerful, memorable content for the finish
- Create strong call-to-action or emotional payoff
- Leave viewers satisfied but wanting more
- You can reorder any highlights to create the best conclusion`

	case "analyze":
		return `You are a YouTube content optimization specialist. Your task is to ANALYZE the content structure.

ANALYSIS INSTRUCTIONS:
- Analyze the current structure and content themes
- Identify strengths and weaknesses in current flow
- Suggest potential improvements without making changes
- Do NOT reorder highlights - this is analysis only
- Keep current order intact in your response`

	case "improve_silences":
		return `You are a YouTube content optimization specialist. Your task is to SUGGEST silence improvements.

SILENCE IMPROVEMENT INSTRUCTIONS:
- Suggest that the user use the "Improve Silences" button in the UI
- Explain that this feature uses AI to add natural silence buffers around words
- This improves the timing and flow of highlights for better video editing
- Note that this action requires direct access to AI services
- Keep current order intact in your response`

	default:
		return `You are a YouTube content optimization specialist. Your task is to optimize highlight organization.`
	}
}

// getOutputFormatRequirements returns the output format requirements for a specific intent
func getOutputFormatRequirements(intent string) string {
	switch intent {
	case "analyze":
		return `

REQUIRED JSON OUTPUT FORMAT:
{
  "success": true,
  "newOrder": [exact same order as current - DO NOT CHANGE],
  "reasoning": "Detailed analysis of current structure, themes, flow, and potential improvements",
  "sectionCount": 0,
  "changes": ["Analysis only - no changes made"]
}

Return ONLY the JSON object above - no additional text.`

	case "improve_silences":
		return `

REQUIRED JSON OUTPUT FORMAT:
{
  "success": true,
  "newOrder": [exact same order as current - DO NOT CHANGE],
  "reasoning": "Explanation of how silence improvements work and instructions to use the UI button",
  "sectionCount": 0,
  "changes": ["Recommended using 'Improve Silences' button in UI for AI-powered timing improvements"],
  "action_required": "improve_silences_ui",
  "description": "Use the 'Improve Silences' button to add natural silence buffers around words for better highlight timing"
}

Return ONLY the JSON object above - no additional text.`

	default:
		return `

REQUIRED JSON OUTPUT FORMAT:
{
  "success": true,
  "newOrder": [
    {"type": "N", "title": "Hook: Engaging Title"},
    "highlight_id_1",
    "highlight_id_2",
    {"type": "N", "title": "Main Content"},
    "highlight_id_3",
    {"type": "N", "title": "Conclusion: Strong Finish"}
  ],
  "reasoning": "Detailed explanation of your reordering decisions and why this improves engagement",
  "sectionCount": 3,
  "changes": ["Created engaging hook", "Grouped related concepts", "Built narrative flow", "Added strong conclusion"]
}

CRITICAL: Include ALL highlight IDs from the available highlights list.
Return ONLY the JSON object above - no additional text.`
	}
}

// formatOrderItem formats an order item (highlight ID or section) for display
func formatOrderItem(item interface{}) string {
	switch v := item.(type) {
	case string:
		return v
	case map[string]interface{}:
		if title, ok := v["title"].(string); ok {
			return fmt.Sprintf("[SECTION] %s", title)
		} else {
			return "[SECTION]"
		}
	default:
		return fmt.Sprintf("%v", item)
	}
}
