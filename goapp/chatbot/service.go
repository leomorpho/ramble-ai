package chatbot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"MYAPP/ent"
	"MYAPP/ent/chatmessage"
	"MYAPP/ent/chatsession"
	"MYAPP/goapp/highlights"
	"MYAPP/goapp/realtime"
)

// NewChatbotService creates a new chatbot service
func NewChatbotService(client *ent.Client, ctx context.Context, updateOrderFunc UpdateOrderFunc) *ChatbotService {
	s := &ChatbotService{
		client:                  client,
		ctx:                     ctx,
		functionRegistry:        make(map[string]FunctionExecutor),
		highlightService:        highlights.NewHighlightService(client, ctx),
		aiService:               highlights.NewAIService(client, ctx),
		updateOrderFunc:         updateOrderFunc,
		mcpRegistry:             NewMCPRegistry(),
		conversationFlowManager: NewConversationFlowManager(),
	}

	// Register available functions (legacy - will be replaced by MCP registry)
	s.registerFunctions()

	return s
}

// findOrCreateSession finds an existing session or creates a new one
func (s *ChatbotService) findOrCreateSession(projectID int, endpointID, sessionID string) (*ent.ChatSession, error) {
	// Try to find existing session by session ID if provided
	if sessionID != "" {
		session, err := s.client.ChatSession.
			Query().
			Where(chatsession.SessionID(sessionID)).
			Only(s.ctx)
		if err == nil {
			return session, nil
		}
		if !ent.IsNotFound(err) {
			return nil, fmt.Errorf("failed to query session by ID: %w", err)
		}
	}

	// Try to find existing session by project/endpoint
	session, err := s.client.ChatSession.
		Query().
		Where(
			chatsession.ProjectID(projectID),
			chatsession.EndpointID(endpointID),
		).
		Only(s.ctx)
	if err == nil {
		return session, nil
	}
	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("failed to query session by project/endpoint: %w", err)
	}

	// Create new session
	newSessionID := sessionID
	if newSessionID == "" {
		newSessionID = fmt.Sprintf("session_%d_%s_%d", projectID, endpointID, time.Now().Unix())
	}

	return s.client.ChatSession.
		Create().
		SetSessionID(newSessionID).
		SetProjectID(projectID).
		SetEndpointID(endpointID).
		Save(s.ctx)
}

// persistMessage saves a message to the database
func (s *ChatbotService) persistMessage(session *ent.ChatSession, messageID, role, content, hiddenContext, model string) error {
	msgCreate := s.client.ChatMessage.
		Create().
		SetMessageID(messageID).
		SetSessionID(session.ID).
		SetRole(chatmessage.Role(role)).
		SetContent(content)

	if hiddenContext != "" {
		msgCreate = msgCreate.SetHiddenContext(hiddenContext)
	}

	if model != "" {
		msgCreate = msgCreate.SetModel(model)
	}

	_, err := msgCreate.Save(s.ctx)
	return err
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

	// Find or create session for persistence
	session, err := s.findOrCreateSession(req.ProjectID, req.EndpointID, req.SessionID)
	if err != nil {
		log.Printf("Failed to find or create session: %v", err)
		return &ChatResponse{
			SessionID: req.SessionID,
			MessageID: messageID,
			Success:   false,
			Error:     "Failed to initialize chat session",
		}, nil
	}

	// Update selected model if it's different from what's saved
	if req.Model != "" && req.Model != session.SelectedModel {
		session, err = session.Update().
			SetSelectedModel(req.Model).
			Save(s.ctx)
		if err != nil {
			log.Printf("Failed to update selected model: %v", err)
			// Continue without failing - this is a non-critical error
		}
	}

	// Generate user message ID and persist user message
	userMessageID := fmt.Sprintf("user_%d", time.Now().UnixNano())
	err = s.persistMessage(session, userMessageID, "user", req.Message, "", "")
	if err != nil {
		log.Printf("Failed to persist user message: %v", err)
		// Continue without failing - this is a non-critical error
	}

	// Broadcast user message
	userMessage := ChatMessage{
		ID:        userMessageID,
		Role:      "user",
		Content:   req.Message,
		Timestamp: time.Now(),
	}

	projectIDStr := strconv.Itoa(req.ProjectID)
	manager := realtime.GetManager()
	manager.BroadcastChatMessageAdded(projectIDStr, req.EndpointID, session.SessionID, userMessage)

	var response *ChatResponse
	var responseErr error

	// Check if endpoint supports MCP actions using registry
	supportsActions := s.mcpRegistry.SupportsActions(req.EndpointID)

	// Use MCP-based action flow if endpoint supports it
	if supportsActions {
		response, responseErr = s.sendMessageWithMCPActions(req, messageID, getAPIKey, session)
	} else {
		// Otherwise, send regular chat message
		response, responseErr = s.sendRegularMessage(req, messageID, getAPIKey, session)
	}

	// Persist and broadcast AI response if successful
	if responseErr == nil && response.Success && response.Message != "" {
		// Persist AI message with model info
		persistErr := s.persistMessage(session, response.MessageID, "assistant", response.Message, "", response.Model)
		if persistErr != nil {
			log.Printf("Failed to persist AI message: %v", persistErr)
			// Continue without failing - this is a non-critical error
		}

		aiMessage := ChatMessage{
			ID:        response.MessageID,
			Role:      "assistant",
			Content:   response.Message,
			Timestamp: time.Now(),
		}

		manager.BroadcastChatMessageAdded(projectIDStr, req.EndpointID, session.SessionID, aiMessage)
	}

	return response, responseErr
}

// sendMessageWithFunctions handles LLM requests with function calling enabled
func (s *ChatbotService) sendMessageWithFunctions(req ChatRequest, messageID string, getAPIKey func() (string, error), session *ent.ChatSession) (*ChatResponse, error) {
	// Get API key
	apiKey, err := getAPIKey()
	if err != nil || apiKey == "" {
		return &ChatResponse{
			SessionID: session.SessionID,
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
			SessionID: session.SessionID,
			MessageID: messageID,
			Success:   false,
			Error:     fmt.Sprintf("LLM API call failed: %v", err),
		}, nil
	}

	// Process LLM response and execute any function calls
	response := &ChatResponse{
		SessionID: session.SessionID,
		MessageID: messageID,
		Model:     req.Model,
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
func (s *ChatbotService) sendRegularMessage(req ChatRequest, messageID string, getAPIKey func() (string, error), session *ent.ChatSession) (*ChatResponse, error) {
	// Skip getting chat history since we're not using it
	// This improves performance by avoiding database queries
	/*
		history, err := s.GetChatHistory(req.ProjectID, req.EndpointID)
		if err != nil {
			log.Printf("Failed to get chat history: %v", err)
			// Continue with empty history
			history = &ChatHistoryResponse{
				SessionID: session.SessionID,
				Messages:  []ChatMessage{},
			}
		}
	*/

	// Build messages for OpenRouter
	messages := []map[string]interface{}{}

	// Add system message based on endpoint
	systemMessage := s.getSystemPromptForEndpoint(req.EndpointID, req.ContextData)
	messages = append(messages, map[string]interface{}{
		"role":    "system",
		"content": systemMessage,
	})

	// Skip conversation history - only send current message
	// This improves performance and reduces token usage
	// Uncomment below to include history if needed
	/*
		historyLimit := 10
		startIdx := 0
		if len(history.Messages) > historyLimit {
			startIdx = len(history.Messages) - historyLimit
		}

		for i := startIdx; i < len(history.Messages); i++ {
			msg := history.Messages[i]
			messages = append(messages, map[string]interface{}{
				"role":    msg.Role,
				"content": msg.Content,
			})
		}
	*/

	// Add current user message
	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": req.Message,
	})

	// Create OpenRouter request
	openRouterReq := map[string]interface{}{
		"model":    req.Model,
		"messages": messages,
	}

	// Get API key
	apiKey, err := getAPIKey()
	if err != nil || apiKey == "" {
		return &ChatResponse{
			SessionID: session.SessionID,
			MessageID: messageID,
			Success:   false,
			Error:     "OpenRouter API key not configured",
		}, nil
	}

	// Call OpenRouter API
	aiResponse, err := s.callOpenRouterAPI(apiKey, openRouterReq)
	if err != nil {
		return &ChatResponse{
			SessionID: session.SessionID,
			MessageID: messageID,
			Success:   false,
			Error:     fmt.Sprintf("AI request failed: %v", err),
		}, nil
	}

	// Extract response
	response := &ChatResponse{
		SessionID: session.SessionID,
		MessageID: messageID,
		Model:     req.Model,
		Success:   true,
	}

	if content, ok := aiResponse["content"].(string); ok {
		response.Message = content
	} else {
		response.Message = "I couldn't generate a proper response. Please try again."
		response.Success = false
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

// getSystemPromptForEndpoint returns appropriate system prompt based on the chatbot endpoint
func (s *ChatbotService) getSystemPromptForEndpoint(endpointID string, contextData map[string]interface{}) string {
	switch endpointID {
	case "highlight_ordering":
		basePrompt := `You are an expert video editor assistant specializing in highlight organization and narrative flow. 
Your role is to help users create engaging video content by:
- Analyzing highlight sequences for optimal viewer engagement
- Suggesting narrative structures and pacing
- Providing creative insights on content flow
- Offering specific reordering recommendations when asked

Be concise, practical, and focused on maximizing viewer retention and emotional impact.`

		// Add highlight context if available
		if highlightsData, ok := contextData["highlights"].(map[string]interface{}); ok {
			if highlightMap, ok := highlightsData["highlightMap"].(map[string]interface{}); ok {
				basePrompt += "\n\nCURRENT PROJECT HIGHLIGHTS:\n\n"

				// Add highlight information
				for id, textInterface := range highlightMap {
					if text, ok := textInterface.(string); ok {
						// Truncate long text for the prompt
						displayText := text
						if len(displayText) > 100 {
							displayText = displayText[:100] + "..."
						}
						basePrompt += fmt.Sprintf("- %s: \"%s\"\n", id, displayText)
					}
				}

				// Add current order information
				if currentOrder, ok := highlightsData["currentOrder"].([]interface{}); ok {
					basePrompt += "\nCURRENT ORDER:\n"
					for i, item := range currentOrder {
						switch v := item.(type) {
						case string:
							basePrompt += fmt.Sprintf("%d. %s\n", i+1, v)
						case map[string]interface{}:
							if title, ok := v["title"].(string); ok {
								basePrompt += fmt.Sprintf("%d. [SECTION: %s]\n", i+1, title)
							}
						}
					}
				}

				if totalHighlights, ok := highlightsData["totalHighlights"].(float64); ok {
					basePrompt += fmt.Sprintf("\nTotal highlights in project: %.0f\n", totalHighlights)
				}

				basePrompt += "\nYou can now provide specific reordering recommendations based on this content. Use the reorder_highlights function if you want to suggest a new arrangement."
			}
		}

		return basePrompt

	case "highlight_suggestions":
		return `You are an AI video analysis assistant that helps identify the most engaging moments in video content.
Your expertise includes:
- Recognizing high-energy moments and emotional peaks
- Identifying key narrative beats and turning points
- Suggesting optimal clip lengths and timing
- Understanding audience engagement patterns

Provide specific, actionable suggestions for highlight selection.`

	case "export_assistance":
		return `You are a video export and optimization specialist. 
Help users with:
- Choosing appropriate export formats and codecs
- Optimizing quality vs file size trade-offs
- Platform-specific requirements (YouTube, TikTok, Instagram, etc.)
- Troubleshooting export issues
- Batch processing strategies

Be technical when needed but explain concepts clearly.`

	default:
		return `You are a helpful AI assistant for video editing and content creation.
Provide clear, concise, and practical advice to help users create better video content.
Focus on being helpful, accurate, and easy to understand.`
	}
}

// GetChatHistory retrieves chat history for a project/endpoint
func (s *ChatbotService) GetChatHistory(projectID int, endpointID string) (*ChatHistoryResponse, error) {
	return s.GetChatHistoryWithLimit(projectID, endpointID, 0) // 0 = no limit
}

// GetChatHistoryWithLimit retrieves chat history with optional message limit for performance
func (s *ChatbotService) GetChatHistoryWithLimit(projectID int, endpointID string, limit int) (*ChatHistoryResponse, error) {
	// Find existing chat session for this project/endpoint
	session, err := s.client.ChatSession.
		Query().
		Where(
			chatsession.ProjectID(projectID),
			chatsession.EndpointID(endpointID),
		).
		WithMessages(func(q *ent.ChatMessageQuery) {
			q.Order(ent.Asc("timestamp"))
			if limit > 0 {
				// Get most recent messages by ordering desc first, limiting, then we'll reverse
				q.Order(ent.Desc("timestamp")).Limit(limit)
			}
		}).
		Only(s.ctx)

	var sessionID string
	var messages []ChatMessage
	var selectedModel string

	if err != nil {
		if ent.IsNotFound(err) {
			// Session doesn't exist yet, return empty history with generated session ID
			sessionID = fmt.Sprintf("session_%d_%s_%d", projectID, endpointID, time.Now().Unix())
			messages = []ChatMessage{}
			// session is nil, so no selected model
		} else {
			return nil, fmt.Errorf("failed to query chat session: %w", err)
		}
	} else {
		// Session exists, convert messages
		sessionID = session.SessionID
		selectedModel = session.SelectedModel
		entMessages := session.Edges.Messages

		// If we used a limit, messages are in desc order, so reverse them
		if limit > 0 && len(entMessages) > 1 {
			for i := 0; i < len(entMessages)/2; i++ {
				j := len(entMessages) - 1 - i
				entMessages[i], entMessages[j] = entMessages[j], entMessages[i]
			}
		}

		messages = make([]ChatMessage, len(entMessages))
		for i, msg := range entMessages {
			messages[i] = ChatMessage{
				ID:        msg.MessageID,
				Role:      string(msg.Role),
				Content:   msg.Content,
				Timestamp: msg.Timestamp,
			}
		}
	}

	response := &ChatHistoryResponse{
		SessionID: sessionID,
		Messages:  messages,
	}

	// Include selected model if available
	if selectedModel != "" {
		response.SelectedModel = selectedModel
	}

	return response, nil
}

// SaveModelSelection saves the selected model for a chat session
func (s *ChatbotService) SaveModelSelection(projectID int, endpointID string, model string) error {
	// Find or create session
	session, err := s.findOrCreateSession(projectID, endpointID, "")
	if err != nil {
		return fmt.Errorf("failed to find or create session: %w", err)
	}

	// Update the selected model
	_, err = session.Update().
		SetSelectedModel(model).
		Save(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to update selected model: %w", err)
	}

	return nil
}

// ClearChatHistory clears chat history for a project/endpoint
func (s *ChatbotService) ClearChatHistory(projectID int, endpointID string) error {
	// Find existing chat session for this project/endpoint
	session, err := s.client.ChatSession.
		Query().
		Where(
			chatsession.ProjectID(projectID),
			chatsession.EndpointID(endpointID),
		).
		Only(s.ctx)

	var sessionID string

	if err != nil {
		if ent.IsNotFound(err) {
			// No session exists, nothing to clear but generate session ID for broadcast
			sessionID = fmt.Sprintf("session_%d_%s_%d", projectID, endpointID, time.Now().Unix())
		} else {
			return fmt.Errorf("failed to query chat session: %w", err)
		}
	} else {
		// Session exists, delete all messages
		sessionID = session.SessionID

		_, err = s.client.ChatMessage.
			Delete().
			Where(chatmessage.SessionID(session.ID)).
			Exec(s.ctx)

		if err != nil {
			return fmt.Errorf("failed to delete chat messages: %w", err)
		}

		// Clear conversation flow for this session
		s.conversationFlowManager.ClearFlow(sessionID)
	}

	// Broadcast chat history cleared event
	projectIDStr := strconv.Itoa(projectID)

	manager := realtime.GetManager()
	manager.BroadcastChatHistoryCleared(projectIDStr, endpointID, sessionID)

	return nil
}

// sendMessageWithMCPActions handles LLM requests using the conversational flow system
func (s *ChatbotService) sendMessageWithMCPActions(req ChatRequest, messageID string, getAPIKey func() (string, error), session *ent.ChatSession) (*ChatResponse, error) {
	// Get or create conversation flow for this session
	flow := s.conversationFlowManager.GetOrCreateFlow(session.SessionID)

	// Create progress broadcaster
	broadcaster := NewProgressBroadcaster(req.ProjectID, req.EndpointID, session.SessionID)

	// Determine which phase we're in
	if flow.Phase == PhaseConversation {
		return s.handleConversationPhase(req, messageID, getAPIKey, session, flow, broadcaster)
	} else {
		return s.handleExecutionPhase(req, messageID, getAPIKey, session, flow, broadcaster)
	}
}

// handleConversationPhase processes messages in conversation mode
func (s *ChatbotService) handleConversationPhase(req ChatRequest, messageID string, getAPIKey func() (string, error), session *ent.ChatSession, flow *ConversationFlow, broadcaster *ProgressBroadcaster) (*ChatResponse, error) {
	broadcaster.UpdateProgress("conversation", "Understanding your request...")

	// Create conversation agent
	conversationAgent := NewConversationAgent(req.EndpointID, s.mcpRegistry)

	// Process conversation
	result, err := conversationAgent.ProcessConversation(req.Message, flow, getAPIKey, s, req.ProjectID)
	if err != nil {
		return &ChatResponse{
			SessionID: session.SessionID,
			MessageID: messageID,
			Success:   false,
			Error:     fmt.Sprintf("Conversation processing failed: %v", err),
		}, nil
	}

	response := &ChatResponse{
		SessionID: session.SessionID,
		MessageID: messageID,
		Model:     "anthropic/claude-sonnet-4",
		Success:   true,
		Message:   result.Response,
	}

	// If we got a confirmed conversation summary, prepare for execution
	if result.HasConversationSummary {
		flow.AddContext("conversation_summary", result.ConversationSummary)
		flow.MoveToExecution()
		s.conversationFlowManager.UpdateFlow(session.SessionID, flow)

		broadcaster.UpdateProgress("summary_confirmed", "User intent confirmed, preparing execution...")

		// Immediately execute using the conversation summary
		return s.handleExecutionPhaseWithSummary(req, messageID, getAPIKey, session, flow, broadcaster, result.ConversationSummary)
	}

	return response, nil
}

// handleExecutionPhaseWithSummary processes execution using conversation summary
func (s *ChatbotService) handleExecutionPhaseWithSummary(req ChatRequest, messageID string, getAPIKey func() (string, error), session *ent.ChatSession, flow *ConversationFlow, broadcaster *ProgressBroadcaster, summary *ConversationSummary) (*ChatResponse, error) {
	// Step 1: Prepare complete prompt using Go preparer
	broadcaster.UpdateProgress("preparing", "Gathering data and building execution prompt...")

	completePrompt, err := s.PrepareExecutorPrompt(summary, req.ProjectID)
	if err != nil {
		return &ChatResponse{
			SessionID: session.SessionID,
			MessageID: messageID,
			Success:   false,
			Error:     fmt.Sprintf("Failed to prepare execution prompt: %v", err),
		}, nil
	}

	// Step 2: Execute using structured approach with complete prompt
	broadcaster.UpdateProgress("processing", s.getProgressMessageForIntent(summary.Intent))

	// Get API key
	apiKey, err := getAPIKey()
	if err != nil || apiKey == "" {
		return &ChatResponse{
			SessionID: session.SessionID,
			MessageID: messageID,
			Success:   false,
			Error:     "OpenRouter API key not configured",
		}, nil
	}

	// Create OpenRouter request with complete prompt
	openRouterReq := map[string]interface{}{
		"model": "anthropic/claude-sonnet-4",
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": completePrompt,
			},
		},
		"temperature": 0.3, // Lower temperature for precise, structured output
		"max_tokens":  4000,
	}

	// Call OpenRouter API
	aiResponse, err := s.callOpenRouterAPI(apiKey, openRouterReq)
	if err != nil {
		broadcaster.UpdateProgress("error", "AI processing failed")
		return &ChatResponse{
			SessionID: session.SessionID,
			MessageID: messageID,
			Success:   false,
			Error:     fmt.Sprintf("AI API call failed: %v", err),
		}, nil
	}

	// Extract content from response
	content, ok := aiResponse["content"].(string)
	if !ok {
		broadcaster.UpdateProgress("error", "Invalid AI response format")
		return &ChatResponse{
			SessionID: session.SessionID,
			MessageID: messageID,
			Success:   false,
			Error:     "AI response missing content",
		}, nil
	}

	broadcaster.UpdateProgress("parsing", "Parsing structured response...")

	// Parse structured output
	structuredOutput, err := ParseStructuredExecutionOutput(content)
	if err != nil {
		broadcaster.UpdateProgress("error", "Failed to parse AI response")
		return &ChatResponse{
			SessionID: session.SessionID,
			MessageID: messageID,
			Success:   false,
			Error:     fmt.Sprintf("Failed to parse structured output: %v", err),
		}, nil
	}

	broadcaster.UpdateProgress("applying", "Applying changes to your project...")

	// Step 3: Apply results using existing update function
	if structuredOutput.Success && summary.Intent != "analyze" {
		err = s.updateOrderFunc(req.ProjectID, structuredOutput.NewOrder)
		if err != nil {
			broadcaster.UpdateProgress("error", "Failed to apply changes")
			return &ChatResponse{
				SessionID: session.SessionID,
				MessageID: messageID,
				Success:   false,
				Error:     fmt.Sprintf("Failed to apply changes: %v", err),
			}, nil
		}
	}

	broadcaster.UpdateProgress("completed", "Successfully completed your request!")

	// Build response
	response := &ChatResponse{
		SessionID:  session.SessionID,
		MessageID:  messageID,
		Model:      "anthropic/claude-sonnet-4",
		Success:    structuredOutput.Success,
		HasActions: summary.Intent != "analyze",
	}

	if structuredOutput.Success {
		response.Message = s.generateSummaryFromStructuredOutput(summary, structuredOutput)
		if structuredOutput.Reasoning != "" {
			response.ActionSummary = structuredOutput.Reasoning
		}
	} else {
		response.Error = structuredOutput.Error
	}

	// Reset flow for next conversation
	flow.Reset()
	s.conversationFlowManager.UpdateFlow(session.SessionID, flow)

	return response, nil
}

// generateSummaryFromStructuredOutput creates a human-readable summary
func (s *ChatbotService) generateSummaryFromStructuredOutput(summary *ConversationSummary, output *StructuredExecutionOutput) string {
	switch summary.Intent {
	case "reorder":
		return fmt.Sprintf("✅ **Success!** Reorganized your highlights into %d sections for better engagement and flow.\n\n**Key Changes:** %s\n\n**Reasoning:** %s",
			output.SectionCount,
			strings.Join(output.Changes, ", "),
			output.Reasoning)

	case "improve_hook":
		return fmt.Sprintf("✅ **Success!** Improved your opening section for stronger hook and better viewer retention.\n\n**Key Changes:** %s\n\n**Reasoning:** %s",
			strings.Join(output.Changes, ", "),
			output.Reasoning)

	case "improve_conclusion":
		return fmt.Sprintf("✅ **Success!** Enhanced your conclusion for more powerful ending and better viewer satisfaction.\n\n**Key Changes:** %s\n\n**Reasoning:** %s",
			strings.Join(output.Changes, ", "),
			output.Reasoning)

	case "analyze":
		return fmt.Sprintf("✅ **Analysis Complete!** Here are insights about your content structure:\n\n%s", output.Reasoning)

	default:
		return fmt.Sprintf("✅ **Success!** Completed %s operation.\n\n**Reasoning:** %s", summary.Intent, output.Reasoning)
	}
}

// getProgressMessageForIntent returns appropriate progress message for intent
func (s *ChatbotService) getProgressMessageForIntent(intent string) string {
	switch intent {
	case "reorder":
		return "Optimizing highlight arrangement..."
	case "improve_hook":
		return "Enhancing opening section..."
	case "improve_conclusion":
		return "Strengthening conclusion..."
	case "analyze":
		return "Analyzing content structure..."
	default:
		return "Processing your request..."
	}
}

// LEGACY METHOD - keeping for backward compatibility
// handleExecutionPhase processes messages in execution mode
func (s *ChatbotService) handleExecutionPhase(req ChatRequest, messageID string, getAPIKey func() (string, error), session *ent.ChatSession, flow *ConversationFlow, broadcaster *ProgressBroadcaster) (*ChatResponse, error) {
	// Legacy implementation kept for backward compatibility
	// This method is no longer used - replaced by handleExecutionPhaseWithSummary
	return &ChatResponse{
		SessionID: session.SessionID,
		MessageID: req.SessionID,
		Success:   false,
		Error:     "Legacy execution method - use conversation summary flow instead",
	}, nil
}

// executeMCPFunctionCall executes a function call using the MCP registry
func (s *ChatbotService) executeMCPFunctionCall(toolCall map[string]interface{}, projectID int, endpointID string) FunctionExecutionResult {
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
		if argsStr != "" && argsStr != "{}" {
			if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
				return FunctionExecutionResult{
					FunctionName: functionName,
					Success:      false,
					Error:        fmt.Sprintf("Failed to parse function arguments: %v", err),
				}
			}
		}
	}

	// Initialize args if nil
	if args == nil {
		args = make(map[string]interface{})
	}

	// Execute using MCP registry
	result, err := s.mcpRegistry.ExecuteFunction(endpointID, functionName, args, projectID, s)
	if err != nil {
		log.Printf("MCP function execution failed: %v", err)
	}

	return result
}

// generateActionSummary creates a human-readable summary of actions performed
func (s *ChatbotService) generateActionSummary(actionsPerformed []string, functionResults []FunctionExecutionResult, endpointID, originalMessage string) (string, error) {
	// Build a summary based on the actions performed
	var summaryBuilder strings.Builder

	summaryBuilder.WriteString("✅ **Actions Completed:**\n\n")

	for i, action := range actionsPerformed {
		switch action {
		case "reorder_highlights":
			summaryBuilder.WriteString(fmt.Sprintf("%d. **Reordered highlights** for better narrative flow\n", i+1))

			// Add details from function result if available
			if i < len(functionResults) && functionResults[i].Success {
				if result, ok := functionResults[i].Result.(map[string]interface{}); ok {
					if reason, ok := result["reason"].(string); ok && reason != "" {
						summaryBuilder.WriteString(fmt.Sprintf("   - *Reasoning:* %s\n", reason))
					}
					if order, ok := result["new_order"].([]interface{}); ok {
						summaryBuilder.WriteString(fmt.Sprintf("   - *New arrangement:* %d items reordered\n", len(order)))
					}
				}
			}

		case "analyze_highlights":
			summaryBuilder.WriteString(fmt.Sprintf("%d. **Analyzed highlight content** for themes and structure\n", i+1))

		case "get_current_order":
			summaryBuilder.WriteString(fmt.Sprintf("%d. **Retrieved current highlight order** for reference\n", i+1))

		case "apply_ai_suggestion":
			summaryBuilder.WriteString(fmt.Sprintf("%d. **Applied AI suggestion** to improve highlight order\n", i+1))

		case "reset_to_original":
			summaryBuilder.WriteString(fmt.Sprintf("%d. **Reset highlights** to original order\n", i+1))

		default:
			summaryBuilder.WriteString(fmt.Sprintf("%d. **Performed action:** %s\n", i+1, action))
		}
	}

	// Add contextual message based on endpoint
	switch endpointID {
	case "highlight_ordering":
		summaryBuilder.WriteString("\n💡 **Next Steps:** Review the new highlight order in your timeline. You can always undo changes or ask for further adjustments.")
	case "highlight_suggestions":
		summaryBuilder.WriteString("\n💡 **Next Steps:** Consider these suggestions when creating your highlight segments.")
	case "content_analysis":
		summaryBuilder.WriteString("\n💡 **Next Steps:** Use these insights to optimize your content strategy.")
	case "export_optimization":
		summaryBuilder.WriteString("\n💡 **Next Steps:** Apply these optimizations to your export settings.")
	}

	return summaryBuilder.String(), nil
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
