package chatbot

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// ConversationAgent handles natural conversation with users
type ConversationAgent struct {
	endpointID string
	registry   *MCPRegistry
}

// NewConversationAgent creates a new conversation agent
func NewConversationAgent(endpointID string, registry *MCPRegistry) *ConversationAgent {
	return &ConversationAgent{
		endpointID: endpointID,
		registry:   registry,
	}
}

// GetCapabilitiesDescription returns human-friendly description of what the agent can do
func (ca *ConversationAgent) GetCapabilitiesDescription(endpointID string) string {
	config, exists := ca.registry.GetEndpointConfig(endpointID)
	if !exists {
		return "I can help you with video editing tasks."
	}

	capabilities := []string{}
	
	// Translate MCP functions to human-friendly descriptions
	for _, function := range config.Functions {
		switch function.Name {
		case "reorder_highlights":
			capabilities = append(capabilities, "📝 **Reorder highlights** - I can rearrange your highlights for better narrative flow and engagement")
		case "analyze_highlights":
			capabilities = append(capabilities, "🔍 **Analyze content** - I can analyze your highlights for themes, structure, and improvement opportunities")
		case "get_current_order":
			capabilities = append(capabilities, "📋 **Review current order** - I can show you how your highlights are currently arranged")
		case "apply_ai_suggestion":
			capabilities = append(capabilities, "💡 **Apply suggestions** - I can apply previously generated optimization suggestions")
		case "reset_to_original":
			capabilities = append(capabilities, "🔄 **Reset order** - I can restore highlights to their original chronological order")
		default:
			// For future functions, use the description from MCP
			capabilities = append(capabilities, fmt.Sprintf("⚡ **%s** - %s", function.Name, function.Description))
		}
	}
	
	if len(capabilities) == 0 {
		return "I can help you with your video editing needs. What would you like to work on?"
	}
	
	return fmt.Sprintf("Here's what I can help you with:\n\n%s\n\nWhat would you like to do?", strings.Join(capabilities, "\n"))
}

// BuildConversationSystemPrompt creates a system prompt for conversation-first interaction
func (ca *ConversationAgent) BuildConversationSystemPrompt(endpointID string) string {
	capabilities := ca.GetCapabilitiesDescription(endpointID)
	
	return fmt.Sprintf(`You are an expert YouTube creator and video editing assistant. You work with HIGHLIGHTS which are selected text excerpts from scripts, not video clips.

CONTEXT ABOUT HIGHLIGHTS:
- Highlights are text snippets from video scripts
- They represent the most engaging parts of content
- You can move any highlight to any position - complete freedom to reorganize
- Your goal is to arrange them into logical SECTIONS for maximum YouTube success

%s

SECTIONING STRATEGY (default approach):
- **Organize highlights into logical sections with descriptive titles**
- **Section flow**: Hook/Intro → Content Sections → Conclusion
- **Section titles**: Should be engaging and descriptive (e.g., "The Problem", "The Solution", "Why This Works")
- **Content grouping**: Group related highlights together within sections
- **Engagement flow**: Each section should build on the previous one to maintain attention

CORE YOUTUBE PRINCIPLES (always apply):
- Start with strong hook section to grab attention in first 3 seconds
- Organize content into logical, flowing sections
- Build narrative tension and engagement across sections
- End with powerful conclusion section for high note finish
- Optimize for audience retention through logical progression

BEHAVIOR GUIDELINES:
1. **Understand user intent**: Use your expertise to interpret what the user actually wants, not just keywords
2. **Think through the request**: Consider the full context and primary vs secondary goals
3. **Be decisive**: Once you understand the intent, proceed with confidence
4. **Use best judgment**: Apply YouTube best practices unless user specifies otherwise
5. **Default assumptions**: Create sections with titles, start fresh, optimize for engagement and logical flow

INTENT UNDERSTANDING APPROACH:
- **Read the full request carefully** - don't just look for keywords
- **Identify the primary goal** - what does the user mainly want to accomplish?
- **Note secondary goals** - what additional context or methods did they mention?
- **Understand the reasoning** - why does the user want this?
- **Extract important context** - any specific requirements or preferences?

EXAMPLES OF INTENT UNDERSTANDING:
- "Please analyze my highlights and reorder them" → PRIMARY: reorder, SECONDARY: provide analysis reasoning
- "Can you optimize the flow?" → PRIMARY: reorder for better flow, SECONDARY: optimization focus
- "Just analyze my content" → PRIMARY: analyze, SECONDARY: none
- "Organize this better with good sections" → PRIMARY: reorder, SECONDARY: focus on sectioning

INTERACTION APPROACH:
1. **Ask ONE question at a time** - gather context incrementally for natural conversation
2. **For REORDER requests**: First ask about current order preference if not specified
3. **After gathering context**: Explain plan and ask for final confirmation
4. **For ANALYZE-only requests**: Proceed immediately (no DB changes)

CONVERSATION FLOW:
- Ask the MOST IMPORTANT missing question first
- Wait for user response before asking next question
- Only ask for final confirmation when you have enough context
- Keep each response focused on ONE question or confirmation

EXAMPLE CLARIFICATION QUESTIONS (ask ONE per response):
- "Would you like me to use your current highlight order as a starting point, or start completely fresh?"
- "Should I prioritize engagement hooks or narrative flow more strongly?"
- "Are there any specific sections or themes you want me to focus on?"

FINAL CONFIRMATION FORMAT:
"I'll [specific plan based on gathered context]. This will modify your highlight order in the database. Should I proceed?"

ONLY AFTER USER CONFIRMS, respond with JSON:
` + "`" + `json
{
  "conversation_summary": {
    "intent": "reorder|improve_hook|improve_conclusion|analyze",
    "userWantsCurrentOrder": true/false,
    "optimizationGoals": ["engagement", "flow", "retention"],
    "specificRequests": ["any specific user requests"],
    "userContext": "important context from user message",
    "confirmed": true
  }
}
` + "`" + `

EXAMPLES:

User: "Please analyze my highlights and reorder them for maximum engagement and narrative flow"
You: "Would you like me to use your current highlight order as a starting point, or start completely fresh?"

User: "Start fresh"
You: "Got it! Should I prioritize engagement hooks or narrative flow more strongly, or balance them equally?"

User: "Balance them equally"
You: "Perfect! I'll create a fresh organization that balances engagement and narrative flow. I'll reorganize your highlights into logical sections with engaging titles like 'Hook', 'The Problem', 'The Solution', etc., optimizing for YouTube retention and story progression. This will reorder your highlights in the database. Should I proceed?"

User: "Yes, go ahead"
You: ` + "`" + `json
{
  "conversation_summary": {
    "intent": "reorder",
    "userWantsCurrentOrder": false,
    "optimizationGoals": ["engagement", "narrative flow"],
    "specificRequests": ["balance engagement and flow equally", "provide analytical reasoning"],
    "userContext": "User wants fresh reordering balancing engagement and narrative flow",
    "confirmed": true
  }
}
` + "`" + `

User: "Just analyze my content structure"
You: ` + "`" + `json
{
  "conversation_summary": {
    "intent": "analyze",
    "userWantsCurrentOrder": false,
    "optimizationGoals": ["content analysis"],
    "specificRequests": ["analyze structure only", "no modifications"],
    "userContext": "User specifically wants analysis only, indicated by 'just analyze'",
    "confirmed": true
  }
}
` + "`" + `

User: "Can you improve my current highlight order for better flow?"
You: "I'll improve your current highlight order by reorganizing it for better narrative flow while keeping your existing structure as the foundation. I'll move highlights around and add section titles to create smoother transitions and better storytelling progression. This will modify your highlight order in the database. Should I proceed?"

User: "Yes"
You: ` + "`" + `json
{
  "conversation_summary": {
    "intent": "reorder",
    "userWantsCurrentOrder": true,
    "optimizationGoals": ["flow", "improvement"],
    "specificRequests": ["use current order as starting point"],
    "userContext": "User wants to improve their existing order rather than start fresh",
    "confirmed": true
  }
}
` + "`" + `

REMEMBER: Be conversational and ask clarifying questions when needed. Only output JSON after the user has confirmed they want to proceed with DB changes!`, capabilities)
}

// ProcessConversation processes a user message in conversation mode
func (ca *ConversationAgent) ProcessConversation(userMessage string, flow *ConversationFlow, getAPIKey func() (string, error), chatService *ChatbotService, projectID int) (*ConversationResult, error) {
	// Get API key
	apiKey, err := getAPIKey()
	if err != nil || apiKey == "" {
		return &ConversationResult{
			Response: "I'm sorry, but I'm having trouble connecting to my AI assistant. Please check your API configuration.",
			HasConversationSummary: false,
		}, nil
	}
	
	// Build conversation system prompt
	systemPrompt := ca.BuildConversationSystemPrompt(ca.endpointID)
	
	// Use context manager to determine optimal history retrieval
	contextManager := NewContextManager()
	model := "anthropic/claude-sonnet-4"
	systemPromptTokens := contextManager.tokenCounter.EstimateTokens(systemPrompt)
	historyLimit := contextManager.GetOptimalHistoryLimit(model, systemPromptTokens)
	
	// Get chat history with intelligent limit
	chatHistory, err := chatService.GetChatHistoryWithLimit(projectID, ca.endpointID, historyLimit)
	if err != nil {
		log.Printf("Failed to get chat history: %v", err)
		// Continue without history if we can't retrieve it
	}
	
	// Build optimized context window with the same context manager
	// Reserve tokens for response (2000) and some buffer (500)
	contextWindow, err := contextManager.BuildContextWindow(
		model,
		systemPrompt,
		chatHistory,
		userMessage,
		2500, // Reserve tokens for response
	)
	if err != nil {
		log.Printf("Failed to build context window: %v", err)
		// Fallback to simple approach
		contextWindow = &ContextWindow{
			Messages: []map[string]interface{}{
				{"role": "system", "content": systemPrompt},
				{"role": "user", "content": userMessage},
			},
		}
	}
	
	// Log context usage for monitoring
	contextManager.LogContextUsage(model, contextWindow)
	
	// Create OpenRouter request without any tools/functions
	openRouterReq := map[string]interface{}{
		"model":       model,
		"messages":    contextWindow.Messages,
		"temperature": 0.7, // Slightly higher temperature for more natural conversation
		"max_tokens":  2000,
	}
	
	// Call OpenRouter API
	aiResponse, err := chatService.callOpenRouterAPI(apiKey, openRouterReq)
	if err != nil {
		return &ConversationResult{
			Response: fmt.Sprintf("I'm having trouble processing your request: %v", err),
			HasConversationSummary: false,
		}, nil
	}
	
	// Extract response content
	content, ok := aiResponse["content"].(string)
	if !ok {
		return &ConversationResult{
			Response: "I'm sorry, I couldn't generate a proper response. Please try again.",
			HasConversationSummary: false,
		}, nil
	}
	
	// Try to parse conversation summary from response
	summary, hasSummary := ca.extractConversationSummaryFromResponse(content)
	
	result := &ConversationResult{
		Response:             content,
		HasConversationSummary: hasSummary,
		ConversationSummary:  summary,
	}
	
	return result, nil
}

// extractConversationSummaryFromResponse attempts to extract a conversation summary from the AI response
func (ca *ConversationAgent) extractConversationSummaryFromResponse(response string) (*ConversationSummary, bool) {
	// Look for JSON code blocks in the response
	lines := strings.Split(response, "\n")
	var jsonLines []string
	inCodeBlock := false
	
	for _, line := range lines {
		if strings.Contains(line, "```json") {
			inCodeBlock = true
			continue
		}
		if strings.Contains(line, "```") && inCodeBlock {
			break
		}
		if inCodeBlock {
			jsonLines = append(jsonLines, line)
		}
	}
	
	if len(jsonLines) == 0 {
		return nil, false
	}
	
	jsonStr := strings.Join(jsonLines, "\n")
	
	// Try to parse the JSON
	var responseData map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &responseData)
	if err != nil {
		log.Printf("Failed to parse conversation JSON: %v", err)
		return nil, false
	}
	
	// Check if this contains a conversation summary
	summaryData, ok := responseData["conversation_summary"].(map[string]interface{})
	if !ok {
		return nil, false
	}
	
	// Check if confirmed
	confirmed, ok := summaryData["confirmed"].(bool)
	if !ok || !confirmed {
		return nil, false
	}
	
	// Extract conversation summary fields
	intent, _ := summaryData["intent"].(string)
	userWantsCurrentOrder, _ := summaryData["userWantsCurrentOrder"].(bool)
	userContext, _ := summaryData["userContext"].(string)
	
	var optimizationGoals []string
	if goals, ok := summaryData["optimizationGoals"].([]interface{}); ok {
		for _, goal := range goals {
			if goalStr, ok := goal.(string); ok {
				optimizationGoals = append(optimizationGoals, goalStr)
			}
		}
	}
	
	var specificRequests []string
	if requests, ok := summaryData["specificRequests"].([]interface{}); ok {
		for _, request := range requests {
			if requestStr, ok := request.(string); ok {
				specificRequests = append(specificRequests, requestStr)
			}
		}
	}
	
	summary := &ConversationSummary{
		Intent:                intent,
		UserWantsCurrentOrder: userWantsCurrentOrder,
		OptimizationGoals:     optimizationGoals,
		SpecificRequests:      specificRequests,
		UserContext:           userContext,
		Confirmed:             confirmed,
	}
	
	return summary, true
}

// ConversationResult represents the result of a conversation interaction
type ConversationResult struct {
	Response               string               `json:"response"`
	HasConversationSummary bool                 `json:"hasConversationSummary"`
	ConversationSummary    *ConversationSummary `json:"conversationSummary,omitempty"`
}