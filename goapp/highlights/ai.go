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
	suggestions, err := s.parseAIHighlightSuggestionsResponse(aiResponse, transcriptWords)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI highlight suggestions response: %w", err)
	}

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
		for j := raw.Start; j <= raw.End; j++ {
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

		// For the end time, use the end of the last word
		if suggestion.End < len(transcriptWords) {
			suggestionEndTime = transcriptWords[suggestion.End].End
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

				if validSuggestion.End < len(transcriptWords) {
					validEndTime = transcriptWords[validSuggestion.End].End
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
	// Convert suggestions to schema.Highlight format with time-based coordinates
	var highlights []schema.Highlight
	for _, suggestion := range suggestions {
		startTime := s.highlightService.WordIndexToTime(suggestion.Start, transcriptWords)
		endTime := s.highlightService.WordIndexToTime(suggestion.End, transcriptWords)

		// For the end time, use the end of the last word
		if suggestion.End < len(transcriptWords) {
			endTime = transcriptWords[suggestion.End].End
		}

		highlight := schema.Highlight{
			ID:    suggestion.ID,
			Start: startTime,
			End:   endTime,
			Color: suggestion.Color,
		}
		highlights = append(highlights, highlight)
	}

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

	// Call OpenRouter API to get AI reordering
	reorderedIDs, err := s.callOpenRouterForReordering(apiKey, aiSettings.AIModel, highlightMap, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI reordering: %w", err)
	}

	// Validate that all IDs are present in the reordered list
	if len(reorderedIDs) != len(highlightIDs) {
		log.Printf("AI reordering returned %d IDs but expected %d", len(reorderedIDs), len(highlightIDs))
		// Fallback to original order if counts don't match
		return highlightIDs, nil
	}

	// Validate that all original IDs are present
	originalIDSet := make(map[string]bool)
	for _, id := range highlightIDs {
		originalIDSet[id] = true
	}

	for _, id := range reorderedIDs {
		if !originalIDSet[id] {
			log.Printf("AI reordering returned unknown ID: %s", id)
			// Fallback to original order if unknown IDs are present
			return highlightIDs, nil
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
	reorderedIDs, err := s.parseAIReorderingResponse(aiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

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

	prompt += `

Analyze these segments and reorder them to create the highest quality video possible for maximum viewer engagement and retention.

IMPORTANT: Respond with ONLY a JSON array containing the highlight IDs in the new order. Do not include any explanation, reasoning, or additional text.

Example format: ["id1", "id2", "id3", ...]`

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

	return reorderedIDs, nil
}