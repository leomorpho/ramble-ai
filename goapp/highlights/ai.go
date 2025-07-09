package highlights

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"MYAPP/ent"
	"MYAPP/ent/project"
	"MYAPP/ent/schema"
	"MYAPP/ent/videoclip"
)

// OpenRouterRequest represents the request format for OpenRouter API
type OpenRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenRouterResponse represents the response from OpenRouter API
type OpenRouterResponse struct {
	Choices []Choice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// Choice represents a response choice
type Choice struct {
	Message Message `json:"message"`
}

// ProjectAISettings represents AI settings for a project
type ProjectAISettings struct {
	AIModel  string `json:"aiModel"`
	AIPrompt string `json:"aiPrompt"`
}

// ProjectAISuggestion represents an AI suggestion for a project
type ProjectAISuggestion struct {
	Order     []string  `json:"order"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"createdAt"`
}

// AIService provides AI-powered highlight functionality
type AIService struct {
	client *ent.Client
	ctx    context.Context
	highlightService *HighlightService
}

// NewAIService creates a new AI service
func NewAIService(client *ent.Client, ctx context.Context) *AIService {
	return &AIService{
		client: client,
		ctx:    ctx,
		highlightService: NewHighlightService(client, ctx),
	}
}

// SuggestHighlightsWithAI generates AI-powered highlight suggestions for a video
func (s *AIService) SuggestHighlightsWithAI(projectID int, videoID int, customPrompt string, getAPIKey func() (string, error)) ([]HighlightSuggestion, error) {
	// Get OpenRouter API key
	apiKey, err := getAPIKey()
	if err != nil || apiKey == "" {
		return nil, fmt.Errorf("OpenRouter API key not configured")
	}

	// Get project AI settings
	aiSettings, err := s.highlightService.GetProjectHighlightAISettings(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project highlight AI settings: %w", err)
	}

	// Use custom prompt if provided, otherwise use project's saved prompt
	prompt := customPrompt
	if prompt == "" {
		prompt = aiSettings.AIPrompt
	}

	// Default prompt if none set
	if prompt == "" {
		prompt = `You are an expert content analyst. Analyze this transcript and suggest meaningful highlight segments that would be valuable for viewers.

Consider:
- Key quotes or important statements
- Actionable advice or insights
- Emotional or engaging moments
- Clear, complete thoughts or phrases
- Natural sentence boundaries

Avoid overlapping with existing highlights and ensure segments are coherent and meaningful.`
	}

	// Get video with transcription and existing highlights
	video, err := s.client.VideoClip.
		Query().
		Where(videoclip.ID(videoID)).
		Only(s.ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	if video.Transcription == "" {
		return nil, fmt.Errorf("video has no transcription")
	}

	// Prepare transcript words with indices
	transcriptWords := video.TranscriptionWords
	if len(transcriptWords) == 0 {
		// Fallback: split transcription into words
		words := strings.Fields(video.Transcription)
		transcriptWords = make([]schema.Word, len(words))
		for i, word := range words {
			transcriptWords[i] = schema.Word{
				Word:  word,
				Start: 0, // No timing info available
				End:   0,
			}
		}
	}

	// Get existing highlights to avoid overlaps
	existingHighlights := video.Highlights

	// Debug log existing highlights
	log.Printf("SuggestHighlightsWithAI: Video ID %d has %d existing highlights", videoID, len(existingHighlights))
	for i, h := range existingHighlights {
		log.Printf("  Existing highlight %d: %s (%.3f-%.3f)", i, h.ID, h.Start, h.End)
	}

	// Call AI to get suggestions
	suggestions, err := s.callOpenRouterForHighlightSuggestions(apiKey, aiSettings.AIModel, transcriptWords, existingHighlights, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI highlight suggestions: %w", err)
	}

	// Debug log raw AI suggestions
	log.Printf("SuggestHighlightsWithAI: AI returned %d raw suggestions", len(suggestions))
	for i, suggestion := range suggestions {
		startTime := s.highlightService.WordIndexToTime(suggestion.Start, transcriptWords)
		endTime := s.highlightService.WordIndexToTime(suggestion.End, transcriptWords)
		if suggestion.End < len(transcriptWords) {
			endTime = transcriptWords[suggestion.End].End
		}
		log.Printf("  Raw suggestion %d: %s [%d-%d] (%.3f-%.3f) '%s'", i, suggestion.ID, suggestion.Start, suggestion.End, startTime, endTime, suggestion.Text)
	}

	// Filter out overlapping suggestions
	validSuggestions := s.filterValidHighlightSuggestions(suggestions, existingHighlights, transcriptWords)

	// Save suggestions to database
	err = s.saveSuggestedHighlights(videoID, validSuggestions, transcriptWords)
	if err != nil {
		log.Printf("Failed to save suggested highlights to database: %v", err)
		// Don't fail the request if saving fails, just log the error
	}

	return validSuggestions, nil
}

// callOpenRouterForHighlightSuggestions calls OpenRouter API to get highlight suggestions
func (s *AIService) callOpenRouterForHighlightSuggestions(apiKey string, model string, transcriptWords []schema.Word, existingHighlights []schema.Highlight, customPrompt string) ([]HighlightSuggestion, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// Build the prompt for AI highlight suggestions
	prompt := s.buildHighlightSuggestionsPrompt(transcriptWords, existingHighlights, customPrompt)

	// Debug log prompt
	log.Printf("=== AI HIGHLIGHT SUGGESTIONS PROMPT ===")
	log.Printf("Model: %s", model)
	log.Printf("Prompt length: %d characters", len(prompt))
	log.Printf("Prompt content: %s", prompt)
	log.Printf("============================================")

	// Create request payload
	requestData := OpenRouterRequest{
		Model: model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(requestData)
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
	req.Header.Set("X-Title", "Video Highlight Suggestions")

	// Make the request
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenRouter API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var openRouterResp OpenRouterResponse
	err = json.Unmarshal(body, &openRouterResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if openRouterResp.Error != nil {
		return nil, fmt.Errorf("OpenRouter API error: %s", openRouterResp.Error.Message)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices received from AI")
	}

	// Extract the highlight suggestions from the AI response
	aiResponse := openRouterResp.Choices[0].Message.Content

	// Debug log response
	log.Printf("=== AI HIGHLIGHT SUGGESTIONS RESPONSE ===")
	log.Printf("Response length: %d characters", len(aiResponse))
	log.Printf("Response content: %s", aiResponse)
	log.Printf("==============================================")

	suggestions, err := s.parseAIHighlightSuggestionsResponse(aiResponse, transcriptWords)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI highlight suggestions response: %w", err)
	}

	// Debug log parsed suggestions
	log.Printf("=== PARSED AI SUGGESTIONS ===")
	log.Printf("Parsed %d suggestions:", len(suggestions))
	for i, suggestion := range suggestions {
		log.Printf("  %d. ID: %s, Start: %d, End: %d, Text: %s", i+1, suggestion.ID, suggestion.Start, suggestion.End, suggestion.Text)
	}
	log.Printf("==============================")

	return suggestions, nil
}

// buildHighlightSuggestionsPrompt creates a prompt for the AI to suggest highlights
func (s *AIService) buildHighlightSuggestionsPrompt(transcriptWords []schema.Word, existingHighlights []schema.Highlight, customPrompt string) string {
	var prompt strings.Builder

	// Add custom prompt
	prompt.WriteString(customPrompt)
	prompt.WriteString("\n\n")

	// Add transcript as indexed words
	prompt.WriteString("TRANSCRIPT (as indexed word pairs):\n")
	for i, word := range transcriptWords {
		prompt.WriteString(fmt.Sprintf("[%d, \"%s\"]", i, word.Word))
		if i < len(transcriptWords)-1 {
			prompt.WriteString(", ")
		}
		if (i+1)%10 == 0 {
			prompt.WriteString("\n")
		}
	}
	prompt.WriteString("\n\n")

	// Add existing highlights context
	if len(existingHighlights) > 0 {
		prompt.WriteString("EXISTING HIGHLIGHTS (do not overlap with these):\n")
		for _, highlight := range existingHighlights {
			// Convert highlight times to word indices (approximate)
			startIdx := s.highlightService.TimeToWordIndex(highlight.Start, transcriptWords)
			endIdx := s.highlightService.TimeToWordIndex(highlight.End, transcriptWords)
			prompt.WriteString(fmt.Sprintf("[%d, %d] ", startIdx, endIdx))
		}
		prompt.WriteString("\n\n")
	}

	prompt.WriteString("TASK: Return suggested highlight segments as word index ranges in JSON format.\n")
	prompt.WriteString("Format: [{\"start\": 5, \"end\": 12}, {\"start\": 25, \"end\": 35}]\n")
	prompt.WriteString("Only return the JSON array, no other text.")

	return prompt.String()
}

// parseAIHighlightSuggestionsResponse parses the AI response to extract highlight suggestions
func (s *AIService) parseAIHighlightSuggestionsResponse(aiResponse string, transcriptWords []schema.Word) ([]HighlightSuggestion, error) {
	// Extract JSON from response (in case AI adds extra text)
	jsonStart := strings.Index(aiResponse, "[")
	jsonEnd := strings.LastIndex(aiResponse, "]")

	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no valid JSON array found in AI response")
	}

	jsonStr := aiResponse[jsonStart : jsonEnd+1]

	// Parse JSON
	var rawSuggestions []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	}

	err := json.Unmarshal([]byte(jsonStr), &rawSuggestions)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI suggestions JSON: %w", err)
	}

	// Convert to HighlightSuggestion structs
	var suggestions []HighlightSuggestion
	baseColors := []string{"#ffeb3b", "#81c784", "#64b5f6", "#ff8a65", "#f06292"}

	for i, raw := range rawSuggestions {
		// Validate indices
		if raw.Start < 0 || raw.End >= len(transcriptWords) || raw.Start > raw.End {
			continue // Skip invalid suggestions
		}

		// Extract text
		var textParts []string
		for j := raw.Start; j < raw.End; j++ {
			textParts = append(textParts, transcriptWords[j].Word)
		}
		text := strings.Join(textParts, " ")

		suggestion := HighlightSuggestion{
			ID:    fmt.Sprintf("suggestion_%d_%d", raw.Start, raw.End),
			Start: raw.Start,
			End:   raw.End,
			Text:  text,
			Color: baseColors[i%len(baseColors)],
		}
		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

// filterValidHighlightSuggestions removes suggestions that overlap with existing highlights
func (s *AIService) filterValidHighlightSuggestions(suggestions []HighlightSuggestion, existingHighlights []schema.Highlight, transcriptWords []schema.Word) []HighlightSuggestion {
	var validSuggestions []HighlightSuggestion

	for _, suggestion := range suggestions {
		hasOverlap := false

		// Get the time range for the suggestion
		suggestionStartTime := s.highlightService.WordIndexToTime(suggestion.Start, transcriptWords)
		suggestionEndTime := s.highlightService.WordIndexToTime(suggestion.End, transcriptWords)

		// For the end time, use the end of the last word (End is exclusive, so use End-1)
		if suggestion.End > 0 && suggestion.End <= len(transcriptWords) {
			suggestionEndTime = transcriptWords[suggestion.End-1].End
		}

		// Check for overlap with existing highlights using time-based comparison
		for _, existing := range existingHighlights {
			// Check for ANY intersection between the two time ranges
			// A highlight overlaps if:
			// 1. It starts before the existing ends AND
			// 2. It ends after the existing starts
			if suggestionStartTime < existing.End && suggestionEndTime > existing.Start {
				hasOverlap = true
				log.Printf("Dropping suggested highlight [%d-%d] (%.2f-%.2f) due to overlap with existing highlight (%.2f-%.2f)",
					suggestion.Start, suggestion.End, suggestionStartTime, suggestionEndTime, existing.Start, existing.End)
				break
			}
		}

		// Also check for overlap with other suggestions that we've already validated
		if !hasOverlap {
			for _, validSuggestion := range validSuggestions {
				validStartTime := s.highlightService.WordIndexToTime(validSuggestion.Start, transcriptWords)
				validEndTime := s.highlightService.WordIndexToTime(validSuggestion.End, transcriptWords)

				if validSuggestion.End > 0 && validSuggestion.End <= len(transcriptWords) {
					validEndTime = transcriptWords[validSuggestion.End-1].End
				}

				if suggestionStartTime < validEndTime && suggestionEndTime > validStartTime {
					hasOverlap = true
					log.Printf("Dropping suggested highlight [%d-%d] (%.2f-%.2f) due to overlap with another suggestion (%.2f-%.2f)",
						suggestion.Start, suggestion.End, suggestionStartTime, suggestionEndTime, validStartTime, validEndTime)
					break
				}
			}
		}

		if !hasOverlap {
			validSuggestions = append(validSuggestions, suggestion)
		}
	}

	log.Printf("Filtered %d suggestions down to %d valid suggestions (removed %d overlapping)",
		len(suggestions), len(validSuggestions), len(suggestions)-len(validSuggestions))

	return validSuggestions
}

// saveSuggestedHighlights saves suggested highlights to the database
func (s *AIService) saveSuggestedHighlights(videoID int, suggestions []HighlightSuggestion, transcriptWords []schema.Word) error {
	// Debug log suggestions before saving
	log.Printf("=== SAVING SUGGESTED HIGHLIGHTS ===")
	log.Printf("VideoID: %d", videoID)
	log.Printf("Number of suggestions to save: %d", len(suggestions))
	for i, suggestion := range suggestions {
		log.Printf("  Input suggestion %d: ID=%s, Start=%d, End=%d, Text=%s", i+1, suggestion.ID, suggestion.Start, suggestion.End, suggestion.Text)
	}
	log.Printf("=====================================")

	// Convert suggestions to schema.Highlight format with time-based coordinates
	var highlights []schema.Highlight
	for _, suggestion := range suggestions {
		startTime := s.highlightService.WordIndexToTime(suggestion.Start, transcriptWords)
		endTime := s.highlightService.WordIndexToTime(suggestion.End, transcriptWords)

		// For the end time, use the end of the last word (End is exclusive, so use End-1)
		if suggestion.End > 0 && suggestion.End <= len(transcriptWords) {
			endTime = transcriptWords[suggestion.End-1].End
		}

		highlight := schema.Highlight{
			ID:    suggestion.ID,
			Start: startTime,
			End:   endTime,
			Color: suggestion.Color,
		}
		highlights = append(highlights, highlight)
	}

	// Debug log converted highlights before database save
	log.Printf("=== CONVERTED HIGHLIGHTS FOR DATABASE ===")
	for i, highlight := range highlights {
		log.Printf("  Converted highlight %d: ID=%s, Start=%.3f, End=%.3f, Color=%s", i+1, highlight.ID, highlight.Start, highlight.End, highlight.Color)
	}
	log.Printf("=========================================")

	_, err := s.client.VideoClip.
		UpdateOneID(videoID).
		SetSuggestedHighlights(highlights).
		Save(s.ctx)

	if err != nil {
		return fmt.Errorf("failed to save suggested highlights: %w", err)
	}

	return nil
}

// GetProjectAISettings gets the AI settings for a specific project
func (s *AIService) GetProjectAISettings(projectID int) (*ProjectAISettings, error) {
	project, err := s.client.Project.
		Query().
		Where(project.ID(projectID)).
		Only(s.ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	aiModel := project.AiModel
	if aiModel == "" {
		aiModel = "anthropic/claude-3.5-haiku-20241022"
	}

	aiPrompt := project.AiPrompt

	return &ProjectAISettings{
		AIModel:  aiModel,
		AIPrompt: aiPrompt,
	}, nil
}

// SaveProjectAISettings saves the AI settings for a specific project
func (s *AIService) SaveProjectAISettings(projectID int, settings ProjectAISettings) error {
	_, err := s.client.Project.
		UpdateOneID(projectID).
		SetAiModel(settings.AIModel).
		SetAiPrompt(settings.AIPrompt).
		Save(s.ctx)

	if err != nil {
		return fmt.Errorf("failed to save project AI settings: %w", err)
	}

	return nil
}

// saveAISuggestion saves the AI suggestion to the database (internal helper)
func (s *AIService) saveAISuggestion(projectID int, reorderedIDs []string, model string) error {
	_, err := s.client.Project.
		UpdateOneID(projectID).
		SetAiSuggestionOrder(reorderedIDs).
		SetAiSuggestionModel(model).
		SetAiSuggestionCreatedAt(time.Now()).
		Save(s.ctx)

	if err != nil {
		return fmt.Errorf("failed to save AI suggestion: %w", err)
	}

	return nil
}

// GetProjectAISuggestion retrieves cached AI suggestion for a project
func (s *AIService) GetProjectAISuggestion(projectID int) (*ProjectAISuggestion, error) {
	project, err := s.client.Project.
		Query().
		Where(project.ID(projectID)).
		Only(s.ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Check if there's a cached AI suggestion
	if project.AiSuggestionOrder == nil {
		return nil, nil // No cached suggestion
	}

	return &ProjectAISuggestion{
		Order:     project.AiSuggestionOrder,
		Model:     project.AiSuggestionModel,
		CreatedAt: project.AiSuggestionCreatedAt,
	}, nil
}

// ReorderHighlightsWithAI uses OpenRouter API to intelligently reorder highlights
func (s *AIService) ReorderHighlightsWithAI(projectID int, customPrompt string, getAPIKey func() (string, error), getProjectHighlights func(int) ([]ProjectHighlight, error)) ([]string, error) {
	// Get OpenRouter API key
	apiKey, err := getAPIKey()
	if err != nil || apiKey == "" {
		return nil, fmt.Errorf("OpenRouter API key not configured")
	}

	// Get project AI settings
	aiSettings, err := s.GetProjectAISettings(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project AI settings: %w", err)
	}

	// Use custom prompt if provided, otherwise use project's saved prompt
	prompt := customPrompt
	if prompt == "" {
		prompt = aiSettings.AIPrompt
	}

	// Get current highlight order to preserve existing newlines
	currentOrder, err := s.highlightService.GetProjectHighlightOrder(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current highlight order: %w", err)
	}

	// Get all project highlights
	projectHighlights, err := getProjectHighlights(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project highlights: %w", err)
	}

	if len(projectHighlights) == 0 {
		return []string{}, nil
	}

	// Create a minimal map of ID to highlight text for AI processing
	highlightMap := make(map[string]string)
	var highlightIDs []string

	for _, ph := range projectHighlights {
		for _, highlight := range ph.Highlights {
			highlightMap[highlight.ID] = highlight.Text
			highlightIDs = append(highlightIDs, highlight.ID)
		}
	}

	if len(highlightMap) == 0 {
		return []string{}, nil
	}

	// Call OpenRouter API to get AI reordering with newline support
	reorderedIDs, err := s.callOpenRouterForReordering(apiKey, aiSettings.AIModel, highlightMap, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI reordering: %w", err)
	}

	// Debug log before deduplication
	log.Printf("=== BEFORE DEDUPLICATION ===")
	log.Printf("Reordered IDs count: %d", len(reorderedIDs))
	for i, id := range reorderedIDs {
		log.Printf("  %d. %s", i+1, id)
	}
	log.Printf("=============================")

	// Clean up duplicates while preserving order and newlines
	reorderedIDs = s.deduplicateHighlightIDs(reorderedIDs, highlightIDs)

	// Debug log after deduplication
	log.Printf("=== AFTER DEDUPLICATION ===")
	log.Printf("Final reordered IDs count: %d", len(reorderedIDs))
	for i, id := range reorderedIDs {
		log.Printf("  %d. %s", i+1, id)
	}
	log.Printf("=============================")

	// Flatten consecutive 'N' characters before returning
	reorderedIDs = s.flattenConsecutiveNewlines(reorderedIDs)

	// Debug log final return value
	log.Printf("=== FINAL RETURN VALUE (after flattening) ===")
	log.Printf("Returning %d items to frontend:", len(reorderedIDs))
	for i, id := range reorderedIDs {
		if id == "N" {
			log.Printf("  %d. NEWLINE CHARACTER", i+1)
		} else {
			log.Printf("  %d. %s", i+1, id)
		}
	}
	log.Printf("==============================================")

	// Validate that all highlight IDs are present in the reordered list (excluding "N" characters)
	// Handle duplicates by creating a set of unique IDs
	actualHighlightIDSet := make(map[string]bool)
	for _, id := range reorderedIDs {
		if id != "N" {
			actualHighlightIDSet[id] = true
		}
	}

	// Convert to slice for counting
	actualHighlightIDs := make([]string, 0, len(actualHighlightIDSet))
	for id := range actualHighlightIDSet {
		actualHighlightIDs = append(actualHighlightIDs, id)
	}

	if len(actualHighlightIDs) != len(highlightIDs) {
		log.Printf("AI reordering returned %d unique highlight IDs but expected %d", len(actualHighlightIDs), len(highlightIDs))
		// Fallback to original order if counts don't match
		return currentOrder, nil
	}

	// Validate that all original IDs are present
	originalIDSet := make(map[string]bool)
	for _, id := range highlightIDs {
		originalIDSet[id] = true
	}

	for _, id := range actualHighlightIDs {
		if !originalIDSet[id] {
			log.Printf("AI reordering returned unknown ID: %s", id)
			// Fallback to original order if unknown IDs are present
			return currentOrder, nil
		}
	}

	// Check for missing IDs
	for _, id := range highlightIDs {
		if !actualHighlightIDSet[id] {
			log.Printf("AI reordering is missing required ID: %s", id)
			// Fallback to original order if IDs are missing
			return currentOrder, nil
		}
	}

	// Save AI suggestion to database
	err = s.saveAISuggestion(projectID, reorderedIDs, aiSettings.AIModel)
	if err != nil {
		log.Printf("Failed to save AI suggestion to database: %v", err)
		// Don't fail the request if saving fails, just log the error
	}

	return reorderedIDs, nil
}

// callOpenRouterForReordering calls the OpenRouter API to get intelligent highlight reordering
func (s *AIService) callOpenRouterForReordering(apiKey string, model string, highlightMap map[string]string, customPrompt string) ([]string, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 60 * time.Second, // AI requests can take longer
	}

	// Build the prompt for AI reordering
	prompt := s.buildReorderingPrompt(highlightMap, customPrompt)

	// Debug log prompt
	log.Printf("=== AI REORDERING PROMPT ===")
	log.Printf("Model: %s", model)
	log.Printf("Highlight count: %d", len(highlightMap))
	log.Printf("Prompt length: %d characters", len(prompt))
	log.Printf("Prompt content: %s", prompt)
	log.Printf("Contains newline instructions: %v", strings.Contains(prompt, "N\" characters"))
	log.Printf("==============================")

	// Create request payload
	requestData := OpenRouterRequest{
		Model: model, // Use the project-specific model
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(requestData)
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
	req.Header.Set("HTTP-Referer", "https://github.com/yourusername/video-app") // Required by OpenRouter
	req.Header.Set("X-Title", "Video Highlight Reordering")                     // Optional but recommended

	// Make the request
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenRouter API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var openRouterResp OpenRouterResponse
	err = json.Unmarshal(body, &openRouterResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if openRouterResp.Error != nil {
		return nil, fmt.Errorf("OpenRouter API error: %s", openRouterResp.Error.Message)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices received from AI")
	}

	// Extract the reordered IDs from the AI response
	aiResponse := openRouterResp.Choices[0].Message.Content

	// Debug log response
	log.Printf("=== AI REORDERING RESPONSE ===")
	log.Printf("Response length: %d characters", len(aiResponse))
	log.Printf("Response content: %s", aiResponse)
	log.Printf("================================")

	reorderedIDs, err := s.parseAIReorderingResponse(aiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Debug log parsed reordering
	log.Printf("=== PARSED AI REORDERING ===")
	log.Printf("Parsed %d reordered IDs:", len(reorderedIDs))
	for i, id := range reorderedIDs {
		log.Printf("  %d. %s", i+1, id)
	}
	log.Printf("==============================")

	return reorderedIDs, nil
}

// buildReorderingPrompt creates a prompt for the AI to reorder highlights intelligently
func (s *AIService) buildReorderingPrompt(highlightMap map[string]string, customPrompt string) string {
	// Use default YouTube expert prompt if no custom prompt provided
	var basePrompt string
	if customPrompt != "" {
		basePrompt = customPrompt
	} else {
		basePrompt = `You are an expert YouTuber and content creator with millions of subscribers, known for creating highly engaging videos that maximize viewer retention and satisfaction. Your task is to reorder these video highlight segments to create the highest quality video possible.

Reorder these segments using your expertise in:
- Hook creation and audience retention
- Storytelling and narrative structure
- Pacing and rhythm for maximum engagement
- Building emotional connections with viewers
- Creating viral-worthy content flow
- Strategic placement of key moments

Feel free to completely restructure the order - move any segment to any position if it will improve video quality and viewer experience.`
	}

	prompt := basePrompt + `

Here are the video highlight segments:

`

	// Convert map to sorted slice for consistent ordering in prompt
	type highlightEntry struct {
		id   string
		text string
	}
	var entries []highlightEntry
	for id, text := range highlightMap {
		entries = append(entries, highlightEntry{id: id, text: text})
	}

	// Sort entries by ID for consistent ordering
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].id < entries[j].id
	})

	for i, entry := range entries {
		prompt += fmt.Sprintf("%d. ID: %s\n", i+1, entry.id)
		prompt += fmt.Sprintf("   Content: %s\n\n", entry.text)
	}

	// Always append the newline instructions regardless of custom prompt
	prompt += `

Analyze these segments and reorder them to create the highest quality video possible for maximum viewer engagement and retention.

IMPORTANT: 
1. Respond with ONLY a JSON array containing the highlight IDs in the new order
2. You can add section breaks by including "N" characters in the array to separate different sections
3. Place "N" characters strategically to create logical groups or chapters in the video
4. Each "N" represents a visual break/newline in the final video timeline
5. Use each highlight ID exactly once - do not duplicate any IDs
6. Include ALL provided highlight IDs in your response (no IDs should be missing)
7. Do not include any explanation, reasoning, or additional text

Example format: ["id1", "id2", "N", "id3", "id4", "N", "id5"]`

	return prompt
}

// parseAIReorderingResponse extracts the reordered highlight IDs from the AI response
func (s *AIService) parseAIReorderingResponse(response string) ([]string, error) {
	// Clean the response - remove any markdown formatting
	cleanResponse := strings.TrimSpace(response)
	cleanResponse = strings.Trim(cleanResponse, "`")
	if strings.HasPrefix(cleanResponse, "json") {
		cleanResponse = strings.TrimPrefix(cleanResponse, "json")
		cleanResponse = strings.TrimSpace(cleanResponse)
	}

	// Try to parse as JSON array
	var reorderedIDs []string
	err := json.Unmarshal([]byte(cleanResponse), &reorderedIDs)
	if err != nil {
		// If direct parsing fails, try to extract JSON from the response
		// Look for JSON array pattern
		jsonStart := strings.Index(cleanResponse, "[")
		jsonEnd := strings.LastIndex(cleanResponse, "]")

		if jsonStart >= 0 && jsonEnd > jsonStart {
			jsonPart := cleanResponse[jsonStart : jsonEnd+1]
			err = json.Unmarshal([]byte(jsonPart), &reorderedIDs)
			if err != nil {
				return nil, fmt.Errorf("failed to parse JSON array from AI response: %w", err)
			}
		} else {
			return nil, fmt.Errorf("no valid JSON array found in AI response")
		}
	}

	// Validate that "N" characters are properly formatted
	for i, id := range reorderedIDs {
		if id != "N" && !strings.HasPrefix(id, "highlight_") {
			log.Printf("Warning: unexpected ID format at position %d: %s", i, id)
		}
	}

	return reorderedIDs, nil
}

// deduplicateHighlightIDs removes duplicate highlight IDs while preserving order and newlines
func (s *AIService) deduplicateHighlightIDs(reorderedIDs []string, originalIDs []string) []string {
	// Create a set of original IDs for validation
	originalIDSet := make(map[string]bool)
	for _, id := range originalIDs {
		originalIDSet[id] = true
	}

	// Track which IDs we've seen to avoid duplicates
	seenIDs := make(map[string]bool)
	result := make([]string, 0)

	for _, id := range reorderedIDs {
		if id == "N" {
			// Always include newline characters
			result = append(result, id)
		} else if originalIDSet[id] && !seenIDs[id] {
			// Only include original IDs that we haven't seen before
			result = append(result, id)
			seenIDs[id] = true
		}
		// Skip duplicates and unknown IDs
	}

	// Add any missing original IDs at the end
	for _, id := range originalIDs {
		if !seenIDs[id] {
			result = append(result, id)
		}
	}

	log.Printf("Deduplication: input %d items, output %d items", len(reorderedIDs), len(result))
	return result
}

// ImproveHighlightSilencesWithAI uses AI to suggest improved timings for highlights with natural silence buffers
func (s *AIService) ImproveHighlightSilencesWithAI(projectID int, getAPIKey func() (string, error)) ([]ProjectHighlight, error) {
	// Get OpenRouter API key
	apiKey, err := getAPIKey()
	if err != nil || apiKey == "" {
		return nil, fmt.Errorf("OpenRouter API key not configured")
	}

	// Get project AI settings
	aiSettings, err := s.GetProjectAISettings(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project AI settings: %w", err)
	}

	// Get current highlight order to preserve newlines
	currentOrder, err := s.highlightService.GetProjectHighlightOrder(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current highlight order: %w", err)
	}

	// Get all project highlights with their transcription words
	projectHighlights, err := s.highlightService.GetProjectHighlights(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project highlights: %w", err)
	}

	if len(projectHighlights) == 0 {
		return []ProjectHighlight{}, nil
	}

	// Get all video clips with transcription words for boundary calculation
	clips, err := s.client.VideoClip.
		Query().
		Where(videoclip.HasProjectWith(project.IDEQ(projectID))).
		All(s.ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get video clips: %w", err)
	}

	// Create map of video clip ID to transcription words
	clipWordsMap := make(map[int][]schema.Word)
	for _, clip := range clips {
		clipWordsMap[clip.ID] = clip.TranscriptionWords
	}

	// Process each highlight to get word boundaries and improved timings
	var improvedHighlights []ProjectHighlight
	for _, ph := range projectHighlights {
		transcriptWords, exists := clipWordsMap[ph.VideoClipID]
		if !exists || len(transcriptWords) == 0 {
			// Skip if no transcript words available
			improvedHighlights = append(improvedHighlights, ph)
			continue
		}

		// Get improved highlights for this video
		improvedVideoHighlights, err := s.improveVideoHighlights(apiKey, aiSettings.AIModel, ph, transcriptWords)
		if err != nil {
			log.Printf("Failed to improve highlights for video %d: %v", ph.VideoClipID, err)
			// Keep original on error
			improvedHighlights = append(improvedHighlights, ph)
			continue
		}

		improvedHighlights = append(improvedHighlights, improvedVideoHighlights)
	}

	// Save AI silence improvements to database cache
	err = s.saveAISilenceImprovements(projectID, improvedHighlights, aiSettings.AIModel)
	if err != nil {
		log.Printf("Failed to save AI silence improvements to database: %v", err)
		// Don't fail the request if saving fails, just log the error
	}

	log.Printf("ImproveHighlightSilencesWithAI: Current highlight order preserved (contains %d newlines)", 
		len(currentOrder) - countHighlightIds(currentOrder))

	return improvedHighlights, nil
}

// countHighlightIds counts the number of actual highlight IDs (excluding "N" characters)
func countHighlightIds(order []string) int {
	count := 0
	for _, id := range order {
		if id != "N" {
			count++
		}
	}
	return count
}

// improveVideoHighlights improves highlights for a single video
func (s *AIService) improveVideoHighlights(apiKey string, model string, videoHighlights ProjectHighlight, transcriptWords []schema.Word) (ProjectHighlight, error) {
	if len(videoHighlights.Highlights) == 0 {
		return videoHighlights, nil
	}

	// Prepare highlight boundaries for AI
	var boundaries []struct {
		ID            string  `json:"id"`
		Text          string  `json:"text"`
		CurrentStart  float64 `json:"currentStart"`
		CurrentEnd    float64 `json:"currentEnd"`
		PrevWordEnd   float64 `json:"prevWordEnd"`
		NextWordStart float64 `json:"nextWordStart"`
	}

	for _, h := range videoHighlights.Highlights {
		// Find word indices for current highlight boundaries
		startIdx := s.highlightService.TimeToWordIndex(h.Start, transcriptWords)
		endIdx := s.highlightService.TimeToWordIndex(h.End, transcriptWords)

		// Get previous word end time
		prevWordEnd := float64(0)
		if startIdx > 0 {
			prevWordEnd = transcriptWords[startIdx-1].End
		}

		// Get next word start time
		nextWordStart := h.End // Default to current end
		if endIdx < len(transcriptWords)-1 {
			// Find the next word after the highlight
			for i := endIdx; i < len(transcriptWords); i++ {
				if transcriptWords[i].Start > h.End {
					nextWordStart = transcriptWords[i].Start
					break
				}
			}
		} else if len(transcriptWords) > 0 {
			// If at the end, use video duration as boundary
			nextWordStart = transcriptWords[len(transcriptWords)-1].End + 0.5
		}

		boundaries = append(boundaries, struct {
			ID            string  `json:"id"`
			Text          string  `json:"text"`
			CurrentStart  float64 `json:"currentStart"`
			CurrentEnd    float64 `json:"currentEnd"`
			PrevWordEnd   float64 `json:"prevWordEnd"`
			NextWordStart float64 `json:"nextWordStart"`
		}{
			ID:            h.ID,
			Text:          h.Text,
			CurrentStart:  h.Start,
			CurrentEnd:    h.End,
			PrevWordEnd:   prevWordEnd,
			NextWordStart: nextWordStart,
		})
	}

	// Call AI to get improved timings
	improvedTimings, err := s.callOpenRouterForSilenceImprovement(apiKey, model, boundaries)
	if err != nil {
		return videoHighlights, fmt.Errorf("failed to get AI silence improvements: %w", err)
	}

	// Apply improved timings to highlights
	improved := ProjectHighlight{
		VideoClipID:   videoHighlights.VideoClipID,
		VideoClipName: videoHighlights.VideoClipName,
		FilePath:      videoHighlights.FilePath,
		Duration:      videoHighlights.Duration,
		Highlights:    make([]HighlightWithText, len(videoHighlights.Highlights)),
	}

	// Create a map for quick lookup of improved timings
	timingMap := make(map[string]struct {
		Start float64
		End   float64
	})
	for _, timing := range improvedTimings {
		timingMap[timing.ID] = struct {
			Start float64
			End   float64
		}{Start: timing.Start, End: timing.End}
	}

	// Apply improvements
	for i, h := range videoHighlights.Highlights {
		if timing, exists := timingMap[h.ID]; exists {
			improved.Highlights[i] = HighlightWithText{
				ID:    h.ID,
				Start: timing.Start,
				End:   timing.End,
				Color: h.Color,
				Text:  h.Text,
			}
		} else {
			// Keep original if no improvement found
			improved.Highlights[i] = h
		}
	}

	return improved, nil
}

// callOpenRouterForSilenceImprovement calls AI to improve highlight timings
func (s *AIService) callOpenRouterForSilenceImprovement(apiKey string, model string, boundaries []struct {
	ID            string  `json:"id"`
	Text          string  `json:"text"`
	CurrentStart  float64 `json:"currentStart"`
	CurrentEnd    float64 `json:"currentEnd"`
	PrevWordEnd   float64 `json:"prevWordEnd"`
	NextWordStart float64 `json:"nextWordStart"`
}) ([]struct {
	ID    string  `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}, error) {
	// Create HTTP client
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// Build prompt
	prompt := s.buildSilenceImprovementPrompt(boundaries)

	// Create request
	requestData := OpenRouterRequest{
		Model: model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	// Log the full request
	log.Printf("=== AI SILENCE IMPROVEMENT LLM REQUEST ===")
	log.Printf("Model: %s", model)
	log.Printf("User Message: %s", prompt)
	log.Printf("===========================================")

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/yourusername/video-app")
	req.Header.Set("X-Title", "Video Highlight Silence Improvement")

	// Make request
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenRouter API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var openRouterResp OpenRouterResponse
	err = json.Unmarshal(body, &openRouterResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if openRouterResp.Error != nil {
		return nil, fmt.Errorf("OpenRouter API error: %s", openRouterResp.Error.Message)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices received from AI")
	}

	// Parse AI response
	aiResponse := openRouterResp.Choices[0].Message.Content
	
	// Log the full response
	log.Printf("=== AI SILENCE IMPROVEMENT LLM RESPONSE ===")
	log.Printf("Model: %s", model)
	log.Printf("Assistant Message: %s", aiResponse)
	log.Printf("Response Length: %d characters", len(aiResponse))
	log.Printf("============================================")
	
	return s.parseAISilenceImprovementResponse(aiResponse)
}

// buildSilenceImprovementPrompt creates the prompt for AI silence improvement
func (s *AIService) buildSilenceImprovementPrompt(boundaries []struct {
	ID            string  `json:"id"`
	Text          string  `json:"text"`
	CurrentStart  float64 `json:"currentStart"`
	CurrentEnd    float64 `json:"currentEnd"`
	PrevWordEnd   float64 `json:"prevWordEnd"`
	NextWordStart float64 `json:"nextWordStart"`
}) string {
	prompt := `You are an expert video editor specializing in creating natural-sounding speech cuts. Your task is to improve highlight timings by including appropriate silence buffers that make the speech flow naturally.

For each highlight, you're given:
- The current start/end times
- The text content
- The end time of the word before the highlight starts
- The start time of the word after the highlight ends

Adjust the start and end times to include natural pauses while staying within the given boundaries. Consider:
- Include slight pauses before sentences or thoughts (100-300ms)
- Include natural breathing room after sentences (200-500ms)
- For questions, include the pause before the answer
- For dramatic statements, include the build-up pause
- Never cut into the middle of words
- Prefer to include complete breaths and natural speech rhythms

Here are the highlights to improve:

`

	// Add highlight data
	for i, b := range boundaries {
		prompt += fmt.Sprintf("%d. Highlight ID: %s\n", i+1, b.ID)
		prompt += fmt.Sprintf("   Text: \"%s\"\n", b.Text)
		prompt += fmt.Sprintf("   Current timing: %.3f - %.3f seconds\n", b.CurrentStart, b.CurrentEnd)
		prompt += fmt.Sprintf("   Available range: %.3f - %.3f seconds\n", b.PrevWordEnd, b.NextWordStart)
		prompt += fmt.Sprintf("   Maximum buffer: %.3fms before, %.3fms after\n\n",
			(b.CurrentStart-b.PrevWordEnd)*1000,
			(b.NextWordStart-b.CurrentEnd)*1000)
	}

	prompt += `
Return a JSON array with improved timings. Each object should have:
- "id": The highlight ID
- "start": The improved start time (must be >= prevWordEnd)
- "end": The improved end time (must be <= nextWordStart)

Only include highlights where you recommend changes. Format:
[{"id": "highlight_1", "start": 1.234, "end": 5.678}, ...]`

	return prompt
}

// parseAISilenceImprovementResponse parses the AI response for improved timings
func (s *AIService) parseAISilenceImprovementResponse(response string) ([]struct {
	ID    string  `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}, error) {
	// Extract JSON from response
	jsonStart := strings.Index(response, "[")
	jsonEnd := strings.LastIndex(response, "]")

	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no valid JSON array found in AI response")
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	var improvements []struct {
		ID    string  `json:"id"`
		Start float64 `json:"start"`
		End   float64 `json:"end"`
	}

	err := json.Unmarshal([]byte(jsonStr), &improvements)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI improvements JSON: %w", err)
	}

	return improvements, nil
}

// saveAISilenceImprovements saves the AI silence improvements to the database cache
func (s *AIService) saveAISilenceImprovements(projectID int, improvements []ProjectHighlight, model string) error {
	log.Printf("=== SAVING AI SILENCE IMPROVEMENTS TO CACHE ===")
	log.Printf("Project ID: %d", projectID)
	log.Printf("Number of improvements: %d", len(improvements))
	log.Printf("Model: %s", model)
	
	// Convert improvements to JSON-serializable format
	improvementsData := make([]map[string]interface{}, 0, len(improvements))
	
	for _, ph := range improvements {
		videoData := map[string]interface{}{
			"videoClipId":   ph.VideoClipID,
			"videoClipName": ph.VideoClipName,
			"filePath":      ph.FilePath,
			"duration":      ph.Duration,
			"highlights":    make([]map[string]interface{}, 0, len(ph.Highlights)),
		}
		
		for _, h := range ph.Highlights {
			highlightData := map[string]interface{}{
				"id":    h.ID,
				"start": h.Start,
				"end":   h.End,
				"color": h.Color,
				"text":  h.Text,
			}
			videoData["highlights"] = append(videoData["highlights"].([]map[string]interface{}), highlightData)
		}
		
		improvementsData = append(improvementsData, videoData)
	}

	_, err := s.client.Project.
		UpdateOneID(projectID).
		SetAiSilenceImprovements(improvementsData).
		SetAiSilenceModel(model).
		SetAiSilenceCreatedAt(time.Now()).
		Save(s.ctx)

	if err != nil {
		log.Printf("ERROR saving AI silence improvements: %v", err)
		return fmt.Errorf("failed to save AI silence improvements: %w", err)
	}

	log.Printf("Successfully saved AI silence improvements to cache")
	return nil
}

// GetProjectAISilenceImprovements retrieves cached AI silence improvements for a project
func (s *AIService) GetProjectAISilenceImprovements(projectID int) ([]ProjectHighlight, time.Time, string, error) {
	log.Printf("=== LOADING AI SILENCE IMPROVEMENTS FROM CACHE ===")
	log.Printf("Project ID: %d", projectID)
	
	project, err := s.client.Project.
		Query().
		Where(project.ID(projectID)).
		Only(s.ctx)

	if err != nil {
		log.Printf("ERROR getting project: %v", err)
		return nil, time.Time{}, "", fmt.Errorf("failed to get project: %w", err)
	}

	log.Printf("Found project. AiSilenceImprovements is nil: %v", project.AiSilenceImprovements == nil)
	if project.AiSilenceImprovements != nil {
		log.Printf("AiSilenceImprovements length: %d", len(project.AiSilenceImprovements))
	}
	log.Printf("AiSilenceModel: %s", project.AiSilenceModel)
	log.Printf("AiSilenceCreatedAt: %v", project.AiSilenceCreatedAt)

	// Check if there's a cached AI silence improvement
	if project.AiSilenceImprovements == nil || len(project.AiSilenceImprovements) == 0 {
		log.Printf("No cached AI silence improvements found")
		return nil, time.Time{}, "", nil // No cached improvements
	}

	// Convert from JSON format back to ProjectHighlight structs using JSON marshaling for reliability
	var improvements []ProjectHighlight
	
	// First, convert the data back to JSON and then unmarshal it properly
	jsonBytes, err := json.Marshal(project.AiSilenceImprovements)
	if err != nil {
		log.Printf("Error marshaling cached improvements: %v", err)
		return nil, time.Time{}, "", fmt.Errorf("failed to marshal cached improvements: %w", err)
	}
	
	log.Printf("Cached improvements JSON: %s", string(jsonBytes))
	
	// Unmarshal into our struct
	err = json.Unmarshal(jsonBytes, &improvements)
	if err != nil {
		log.Printf("Error unmarshaling cached improvements: %v", err)
		return nil, time.Time{}, "", fmt.Errorf("failed to unmarshal cached improvements: %w", err)
	}
	
	log.Printf("Successfully loaded %d cached improvements", len(improvements))
	
	return improvements, project.AiSilenceCreatedAt, project.AiSilenceModel, nil
}

// ClearAISilenceImprovementsCache clears the cached AI silence improvements for a project
func ClearAISilenceImprovementsCache(ctx context.Context, client *ent.Client, projectID int) error {
	_, err := client.Project.
		UpdateOneID(projectID).
		ClearAiSilenceImprovements().
		ClearAiSilenceModel().
		ClearAiSilenceCreatedAt().
		Save(ctx)
	
	if err != nil {
		return fmt.Errorf("failed to clear AI silence improvements cache: %w", err)
	}
	
	return nil
}

// flattenConsecutiveNewlines removes consecutive 'N' characters from the array
func (s *AIService) flattenConsecutiveNewlines(ids []string) []string {
	if len(ids) <= 1 {
		return ids
	}

	var result []string
	lastWasNewline := false

	for _, id := range ids {
		if id == "N" {
			if !lastWasNewline {
				result = append(result, id)
				lastWasNewline = true
			}
			// Skip consecutive newlines
		} else {
			result = append(result, id)
			lastWasNewline = false
		}
	}

	return result
}