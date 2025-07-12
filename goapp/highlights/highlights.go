package highlights

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"MYAPP/ent"
	"MYAPP/ent/project"
	"MYAPP/ent/schema"
	"MYAPP/ent/settings"
	"MYAPP/ent/videoclip"
	"MYAPP/goapp/realtime"
)

// Highlight represents a highlighted text region with timestamps
type Highlight struct {
	ID    string  `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Color string  `json:"color"`
}

// HighlightWithText represents a highlight with its text content
type HighlightWithText struct {
	ID         string  `json:"id"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Color      string  `json:"color"`
	Text       string  `json:"text"`
	StartIndex int     `json:"startIndex"` // Word index where highlight starts
	EndIndex   int     `json:"endIndex"`   // Word index where highlight ends
}

// ProjectHighlight represents a video clip with its highlights
type ProjectHighlight struct {
	VideoClipID   int                 `json:"videoClipId"`
	VideoClipName string              `json:"videoClipName"`
	FilePath      string              `json:"filePath"`
	Duration      float64             `json:"duration"`
	Highlights    []HighlightWithText `json:"highlights"`
}

// HighlightSegment represents a highlight segment for export operations
type HighlightSegment struct {
	ID            string  `json:"id"`
	VideoPath     string  `json:"videoPath"`
	Start         float64 `json:"start"`
	End           float64 `json:"end"`
	Color         string  `json:"color"`
	Text          string  `json:"text"`
	VideoClipID   int     `json:"videoClipId"`
	VideoClipName string  `json:"videoClipName"`
}

// ProjectHighlightAISettings represents AI settings for highlight suggestions
type ProjectHighlightAISettings struct {
	AIModel  string `json:"aiModel"`
	AIPrompt string `json:"aiPrompt"`
}

// HighlightSuggestion represents an AI-generated highlight suggestion
type HighlightSuggestion struct {
	ID    string `json:"id"`
	Start int    `json:"start"`
	End   int    `json:"end"`
	Text  string `json:"text"`
	Color string `json:"color"`
}

// HighlightService provides highlight management functionality
type HighlightService struct {
	client *ent.Client
	ctx    context.Context
}

// NewHighlightService creates a new highlight service
func NewHighlightService(client *ent.Client, ctx context.Context) *HighlightService {
	return &HighlightService{
		client: client,
		ctx:    ctx,
	}
}

// GetSuggestedHighlights retrieves saved suggested highlights for a video
func (s *HighlightService) GetSuggestedHighlights(videoID int) ([]HighlightSuggestion, error) {
	// Get the video clip with suggested highlights
	clip, err := s.client.VideoClip.Get(s.ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video clip: %w", err)
	}

	// Debug log database highlights
	log.Printf("=== GET SUGGESTED HIGHLIGHTS FROM DATABASE ===")
	log.Printf("VideoID: %d", videoID)
	log.Printf("Number of stored highlights: %d", len(clip.SuggestedHighlights))
	for i, h := range clip.SuggestedHighlights {
		log.Printf("  Stored highlight %d: ID=%s, Start=%.3f, End=%.3f, Color=%s", i+1, h.ID, h.Start, h.End, h.Color)
	}
	log.Printf("===============================================")

	var suggestions []HighlightSuggestion
	for i, h := range clip.SuggestedHighlights {
		// Convert time-based highlight to word index for text extraction
		startIndex := s.timeToWordIndex(h.Start, clip.TranscriptionWords)
		endIndex := s.timeToWordIndexForEnd(h.End, clip.TranscriptionWords)

		// Debug log conversion
		log.Printf("  Converting highlight %d: Time(%.3f-%.3f) -> WordIndex(%d-%d)", i+1, h.Start, h.End, startIndex, endIndex)
		
		// Additional debug: show what words are at these times
		if startIndex < len(clip.TranscriptionWords) {
			log.Printf("    Start: Time %.3f -> Word[%d]='%s' (%.3f-%.3f)", h.Start, startIndex, 
				clip.TranscriptionWords[startIndex].Word, 
				clip.TranscriptionWords[startIndex].Start, 
				clip.TranscriptionWords[startIndex].End)
		}
		if endIndex < len(clip.TranscriptionWords) {
			log.Printf("    End: Time %.3f -> Word[%d]='%s' (%.3f-%.3f)", h.End, endIndex,
				clip.TranscriptionWords[endIndex].Word,
				clip.TranscriptionWords[endIndex].Start,
				clip.TranscriptionWords[endIndex].End)
		}
		if endIndex > 0 && endIndex-1 < len(clip.TranscriptionWords) {
			log.Printf("    Previous word: Word[%d]='%s' (%.3f-%.3f)", endIndex-1,
				clip.TranscriptionWords[endIndex-1].Word,
				clip.TranscriptionWords[endIndex-1].Start,
				clip.TranscriptionWords[endIndex-1].End)
		}

		// Extract text from the transcript
		text := s.extractTextFromWordRange(clip.TranscriptionWords, startIndex, endIndex)

		// Debug log extracted text
		log.Printf("  Extracted text %d: '%s'", i+1, text)

		suggestion := HighlightSuggestion{
			ID:    h.ID,
			Start: startIndex,
			End:   endIndex,
			Text:  text,
			Color: h.Color,
		}
		suggestions = append(suggestions, suggestion)
	}

	// Debug log final suggestions being returned
	log.Printf("=== FINAL SUGGESTIONS BEING RETURNED ===")
	log.Printf("Returning %d suggestions:", len(suggestions))
	for i, s := range suggestions {
		log.Printf("  Final suggestion %d: ID=%s, Start=%d, End=%d, Text=%s", i+1, s.ID, s.Start, s.End, s.Text)
	}
	log.Printf("========================================")

	return suggestions, nil
}

// ClearSuggestedHighlights removes all suggested highlights for a video
func (s *HighlightService) ClearSuggestedHighlights(videoID int) error {
	// Get clip with project information for broadcasting
	clip, err := s.client.VideoClip.
		Query().
		Where(videoclip.ID(videoID)).
		WithProject().
		Only(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to get video clip: %w", err)
	}

	_, err = s.client.VideoClip.
		UpdateOneID(videoID).
		ClearSuggestedHighlights().
		Save(s.ctx)

	if err != nil {
		return fmt.Errorf("failed to clear suggested highlights: %w", err)
	}

	// Broadcast real-time deletion event for suggested highlights
	if clip.Edges.Project != nil {
		projectIDStr := strconv.Itoa(clip.Edges.Project.ID)
		manager := realtime.GetManager()
		// Use a special ID to indicate all suggested highlights were cleared
		manager.BroadcastHighlightsDelete(projectIDStr, []string{"suggested_highlights_cleared"})
	}

	return nil
}

// DeleteSuggestedHighlight removes a specific suggested highlight from a video
func (s *HighlightService) DeleteSuggestedHighlight(videoID int, suggestionID string) error {
	// Get the current video clip with its suggested highlights and project
	clip, err := s.client.VideoClip.
		Query().
		Where(videoclip.ID(videoID)).
		WithProject().
		Only(s.ctx)

	if err != nil {
		return fmt.Errorf("failed to get video clip: %w", err)
	}

	// Filter out the suggested highlight to delete
	var updatedSuggestions []schema.Highlight
	for _, suggestion := range clip.SuggestedHighlights {
		if suggestion.ID != suggestionID {
			updatedSuggestions = append(updatedSuggestions, suggestion)
		}
	}

	// Update the video clip with the filtered suggested highlights
	_, err = s.client.VideoClip.
		UpdateOneID(videoID).
		SetSuggestedHighlights(updatedSuggestions).
		Save(s.ctx)

	if err != nil {
		return fmt.Errorf("failed to update video clip suggested highlights: %w", err)
	}

	// Broadcast real-time deletion event for suggested highlights
	if clip.Edges.Project != nil {
		projectIDStr := strconv.Itoa(clip.Edges.Project.ID)
		manager := realtime.GetManager()
		// Prefix suggested highlight IDs to differentiate from regular highlights
		manager.BroadcastHighlightsDelete(projectIDStr, []string{"suggested_" + suggestionID})
	}

	return nil
}

// DeleteHighlight removes a specific highlight from a video clip by highlight ID
func (s *HighlightService) DeleteHighlight(clipID int, highlightID string) error {
	// Get the current video clip with its highlights and project
	clip, err := s.client.VideoClip.
		Query().
		Where(videoclip.ID(clipID)).
		WithProject().
		Only(s.ctx)

	if err != nil {
		return fmt.Errorf("failed to get video clip: %w", err)
	}

	// Filter out the highlight to delete
	var updatedHighlights []schema.Highlight
	for _, highlight := range clip.Highlights {
		if highlight.ID != highlightID {
			updatedHighlights = append(updatedHighlights, highlight)
		}
	}

	// Update the video clip with the filtered highlights
	_, err = s.client.VideoClip.
		UpdateOneID(clipID).
		SetHighlights(updatedHighlights).
		Save(s.ctx)

	if err != nil {
		return fmt.Errorf("failed to update video clip highlights: %w", err)
	}

	// Broadcast real-time deletion event
	if clip.Edges.Project != nil {
		projectIDStr := strconv.Itoa(clip.Edges.Project.ID)
		manager := realtime.GetManager()
		manager.BroadcastHighlightsDelete(projectIDStr, []string{highlightID})
	}

	return nil
}

// GetProjectHighlights gets all highlights for a project with their associated video clips
func (s *HighlightService) GetProjectHighlights(projectID int) ([]ProjectHighlight, error) {
	// Get all video clips for the project with their highlights
	clips, err := s.client.VideoClip.
		Query().
		Where(videoclip.HasProjectWith(project.IDEQ(projectID))).
		All(s.ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get video clips: %w", err)
	}

	var projectHighlights []ProjectHighlight
	for _, clip := range clips {
		// Skip clips without highlights
		if len(clip.Highlights) == 0 {
			continue
		}

		// Convert schema highlights to highlights with text
		var highlightsWithText []HighlightWithText
		for _, h := range clip.Highlights {
			hwt := HighlightWithText{
				ID:    h.ID,
				Start: h.Start,
				End:   h.End,
				Color: h.Color,
			}

			// Extract text and word indices for the highlight if transcription exists
			if clip.Transcription != "" && len(clip.TranscriptionWords) > 0 {
				hwt.Text = s.extractHighlightText(h, clip.TranscriptionWords, clip.Transcription)
				// Calculate word indices from timestamps
				hwt.StartIndex = s.findWordIndexByTimestamp(h.Start, clip.TranscriptionWords, true)
				hwt.EndIndex = s.findWordIndexByTimestamp(h.End, clip.TranscriptionWords, false)
			} else {
				hwt.StartIndex = -1
				hwt.EndIndex = -1
			}

			highlightsWithText = append(highlightsWithText, hwt)
		}

		projectHighlight := ProjectHighlight{
			VideoClipID:   clip.ID,
			VideoClipName: clip.Name,
			FilePath:      clip.FilePath,
			Duration:      clip.Duration,
			Highlights:    highlightsWithText,
		}

		projectHighlights = append(projectHighlights, projectHighlight)
	}

	return projectHighlights, nil
}

// GetProjectHighlightAISettings retrieves AI settings for highlight suggestions
func (s *HighlightService) GetProjectHighlightAISettings(projectID int) (*ProjectHighlightAISettings, error) {
	// Get AI model setting
	modelKey := fmt.Sprintf("project_%d_highlight_ai_model", projectID)
	model, err := s.getSetting(modelKey)
	if err != nil {
		model = "openai/gpt-4o-mini" // Default model
	}

	// Get AI prompt setting
	promptKey := fmt.Sprintf("project_%d_highlight_ai_prompt", projectID)
	prompt, err := s.getSetting(promptKey)
	if err != nil {
		prompt = "Analyze this transcript and suggest the most interesting highlights for a video compilation." // Default prompt
	}

	return &ProjectHighlightAISettings{
		AIModel:  model,
		AIPrompt: prompt,
	}, nil
}

// SaveProjectHighlightAISettings saves AI settings for highlight suggestions
func (s *HighlightService) SaveProjectHighlightAISettings(projectID int, settings ProjectHighlightAISettings) error {
	// Save AI model setting
	modelKey := fmt.Sprintf("project_%d_highlight_ai_model", projectID)
	if err := s.saveSetting(modelKey, settings.AIModel); err != nil {
		return fmt.Errorf("failed to save AI model setting: %w", err)
	}

	// Save AI prompt setting
	promptKey := fmt.Sprintf("project_%d_highlight_ai_prompt", projectID)
	if err := s.saveSetting(promptKey, settings.AIPrompt); err != nil {
		return fmt.Errorf("failed to save AI prompt setting: %w", err)
	}

	return nil
}

// GetProjectHighlightOrder retrieves the custom highlight order for a project
func (s *HighlightService) GetProjectHighlightOrder(projectID int) ([]string, error) {
	project, err := s.client.Project.
		Query().
		Where(project.ID(projectID)).
		First(s.ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("project not found: %d", projectID)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Return the highlight order from the project schema
	if project.HighlightOrder == nil {
		return []string{}, nil
	}

	// Convert []interface{} to []string
	result := make([]string, 0, len(project.HighlightOrder))
	for _, item := range project.HighlightOrder {
		switch v := item.(type) {
		case string:
			result = append(result, v)
		case map[string]interface{}:
			// Convert newline objects back to simple "N" 
			if typeVal, ok := v["type"].(string); ok && typeVal == "N" {
				result = append(result, "N")
			}
		default:
			// Handle other types as strings
			result = append(result, fmt.Sprintf("%v", v))
		}
	}

	return result, nil
}

// GetProjectHighlightOrderWithTitles retrieves the highlight order with rich newline objects
func (s *HighlightService) GetProjectHighlightOrderWithTitles(projectID int) ([]interface{}, error) {
	project, err := s.client.Project.
		Query().
		Where(project.ID(projectID)).
		First(s.ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("project not found: %d", projectID)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Return the highlight order as-is, which can contain both strings and objects
	if project.HighlightOrder == nil {
		return []interface{}{}, nil
	}

	return project.HighlightOrder, nil
}

// GetProjectHighlightsForExport gets highlights formatted for export
func (s *HighlightService) GetProjectHighlightsForExport(projectID int) ([]HighlightSegment, error) {
	// Get all video clips for the project with their highlights
	clips, err := s.client.VideoClip.
		Query().
		Where(videoclip.HasProjectWith(project.IDEQ(projectID))).
		All(s.ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get video clips: %w", err)
	}

	var segments []HighlightSegment
	for _, clip := range clips {
		for _, highlight := range clip.Highlights {
			text := s.extractHighlightText(highlight, clip.TranscriptionWords, clip.Transcription)

			segments = append(segments, HighlightSegment{
				ID:            highlight.ID,
				VideoPath:     clip.FilePath,
				Start:         highlight.Start,
				End:           highlight.End,
				Color:         highlight.Color,
				Text:          text,
				VideoClipID:   clip.ID,
				VideoClipName: clip.Name,
			})
		}
	}

	return segments, nil
}

// ApplyHighlightOrder applies custom ordering to highlight segments
func (s *HighlightService) ApplyHighlightOrder(segments []HighlightSegment, order []string) []HighlightSegment {
	if len(order) == 0 {
		return segments
	}

	// Create a map for quick lookup
	orderMap := make(map[string]int)
	for i, id := range order {
		orderMap[id] = i
	}

	// Sort segments based on the custom order
	sort.Slice(segments, func(i, j int) bool {
		posI, foundI := orderMap[segments[i].ID]
		posJ, foundJ := orderMap[segments[j].ID]

		// If both are in the order, sort by position
		if foundI && foundJ {
			return posI < posJ
		}

		// If only one is in the order, it comes first
		if foundI && !foundJ {
			return true
		}
		if !foundI && foundJ {
			return false
		}

		// If neither is in the order, maintain original order
		return i < j
	})

	return segments
}

// Helper functions

// extractHighlightText extracts text content from a highlight using transcript words
func (s *HighlightService) extractHighlightText(highlight schema.Highlight, words []schema.Word, fullText string) string {
	if len(words) == 0 {
		return ""
	}

	// Find words that fall within the highlight time range
	var highlightWords []string
	for _, word := range words {
		if word.Start >= highlight.Start && word.End <= highlight.End {
			highlightWords = append(highlightWords, word.Word)
		}
	}

	if len(highlightWords) > 0 {
		return strings.Join(highlightWords, " ")
	}

	// Fallback: try to extract from full text based on time approximation
	if fullText != "" {
		totalDuration := float64(0)
		if len(words) > 0 {
			totalDuration = words[len(words)-1].End
		}

		if totalDuration > 0 {
			startRatio := highlight.Start / totalDuration
			endRatio := highlight.End / totalDuration

			startChar := int(startRatio * float64(len(fullText)))
			endChar := int(endRatio * float64(len(fullText)))

			if startChar >= 0 && endChar <= len(fullText) && startChar < endChar {
				return fullText[startChar:endChar]
			}
		}
	}

	return ""
}

// timeToWordIndex converts time in seconds to approximate word index
func (s *HighlightService) timeToWordIndex(timeSeconds float64, transcriptWords []schema.Word) int {
	// Special case: if time is 0 or negative, return 0
	if timeSeconds <= 0 {
		return 0
	}
	
	// First pass: find word that starts at exactly this time
	for i, word := range transcriptWords {
		if timeSeconds == word.Start {
			return i
		}
	}
	
	// Second pass: find word that contains this time
	for i, word := range transcriptWords {
		if timeSeconds > word.Start && timeSeconds <= word.End {
			return i
		}
		// If the time is before this word starts, return this index
		if word.Start > timeSeconds {
			return i
		}
	}
	return len(transcriptWords) - 1
}

// timeToWordIndexForEnd converts end time back to word index (inverse of saving logic)
func (s *HighlightService) timeToWordIndexForEnd(timeSeconds float64, transcriptWords []schema.Word) int {
	// This is the inverse of: endTime = transcriptWords[suggestion.End-1].End
	// We need to find the word whose end time matches (or is closest to) the given time,
	// then return the index + 1 to get back to the exclusive end index
	
	// Find the last word whose end time is <= the given time
	for i := len(transcriptWords) - 1; i >= 0; i-- {
		if transcriptWords[i].End <= timeSeconds + 0.001 { // Small epsilon for floating point comparison
			// The saved time was from word i, so the exclusive end index is i+1
			return i + 1
		}
	}
	return 0
}

// TimeToWordIndex converts time in seconds to approximate word index (public method)
func (s *HighlightService) TimeToWordIndex(timeSeconds float64, transcriptWords []schema.Word) int {
	return s.timeToWordIndex(timeSeconds, transcriptWords)
}

// WordIndexToTime converts word index to time in seconds (public method)
func (s *HighlightService) WordIndexToTime(wordIndex int, transcriptWords []schema.Word) float64 {
	return s.wordIndexToTime(wordIndex, transcriptWords)
}

// findWordIndexByTimestamp finds the word index for a given timestamp
// isStart: true for finding start index, false for end index
func (s *HighlightService) findWordIndexByTimestamp(timestamp float64, words []schema.Word, isStart bool) int {
	if len(words) == 0 {
		return -1
	}

	// For start timestamp, find first word that overlaps
	if isStart {
		for i, word := range words {
			if word.End >= timestamp {
				return i
			}
		}
		return len(words) - 1
	}

	// For end timestamp, find last word that overlaps
	for i := len(words) - 1; i >= 0; i-- {
		if words[i].Start <= timestamp {
			return i
		}
	}
	return 0
}

// wordIndexToTime converts word index to time in seconds
func (s *HighlightService) wordIndexToTime(wordIndex int, transcriptWords []schema.Word) float64 {
	if wordIndex >= len(transcriptWords) || wordIndex < 0 {
		return 0
	}
	return transcriptWords[wordIndex].Start
}

// extractTextFromWordRange extracts text from a range of words
func (s *HighlightService) extractTextFromWordRange(words []schema.Word, startIndex, endIndex int) string {
	if startIndex < 0 || endIndex > len(words) || startIndex > endIndex {
		return ""
	}

	var textParts []string
	for i := startIndex; i < endIndex; i++ {
		textParts = append(textParts, words[i].Word)
	}

	return strings.Join(textParts, " ")
}

// getSetting retrieves a setting value by key
func (s *HighlightService) getSetting(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("setting key cannot be empty")
	}

	setting, err := s.client.Settings.
		Query().
		Where(settings.Key(key)).
		Only(s.ctx)

	if err != nil {
		// Return empty string if setting doesn't exist
		return "", nil
	}

	return setting.Value, nil
}

// saveSetting saves a setting value
func (s *HighlightService) saveSetting(key, value string) error {
	// Check if setting exists
	existing, err := s.client.Settings.
		Query().
		Where(settings.Key(key)).
		Only(s.ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// Create new setting
			_, err = s.client.Settings.
				Create().
				SetKey(key).
				SetValue(value).
				Save(s.ctx)
			return err
		}
		return fmt.Errorf("failed to query setting: %w", err)
	}

	// Update existing setting
	_, err = s.client.Settings.
		UpdateOne(existing).
		SetValue(value).
		Save(s.ctx)

	return err
}
