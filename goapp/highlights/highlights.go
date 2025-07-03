package highlights

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"MYAPP/ent"
	"MYAPP/ent/schema"
	"MYAPP/ent/settings"
	"MYAPP/ent/videoclip"
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
	ID    string  `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Color string  `json:"color"`
	Text  string  `json:"text"`
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

	var suggestions []HighlightSuggestion
	for _, h := range clip.SuggestedHighlights {
		// Convert time-based highlight to word index for text extraction
		startIndex := s.timeToWordIndex(h.Start, clip.TranscriptionWords)
		endIndex := s.timeToWordIndex(h.End, clip.TranscriptionWords)
		
		// Extract text from the transcript
		text := s.extractTextFromWordRange(clip.TranscriptionWords, startIndex, endIndex)
		
		suggestions = append(suggestions, HighlightSuggestion{
			ID:    h.ID,
			Start: startIndex,
			End:   endIndex,
			Text:  text,
			Color: h.Color,
		})
	}

	return suggestions, nil
}

// ClearSuggestedHighlights removes all suggested highlights for a video
func (s *HighlightService) ClearSuggestedHighlights(videoID int) error {
	_, err := s.client.VideoClip.
		UpdateOneID(videoID).
		ClearSuggestedHighlights().
		Save(s.ctx)
	
	if err != nil {
		return fmt.Errorf("failed to clear suggested highlights: %w", err)
	}
	
	return nil
}

// DeleteHighlight removes a specific highlight from a video clip by highlight ID
func (s *HighlightService) DeleteHighlight(clipID int, highlightID string) error {
	// Get the current video clip with its highlights
	clip, err := s.client.VideoClip.
		Query().
		Where(videoclip.ID(clipID)).
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

	return nil
}

// GetProjectHighlights gets all highlights for a project with their associated video clips
func (s *HighlightService) GetProjectHighlights(projectID int) ([]ProjectHighlight, error) {
	// Get all video clips for the project with their highlights
	clips, err := s.client.VideoClip.
		Query().
		Where(videoclip.HasProjectWith()).
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

			// Extract text for the highlight if transcription exists
			if clip.Transcription != "" && len(clip.TranscriptionWords) > 0 {
				hwt.Text = s.extractHighlightText(h, clip.TranscriptionWords, clip.Transcription)
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
	settingKey := fmt.Sprintf("project_%d_highlight_order", projectID)

	setting, err := s.client.Settings.
		Query().
		Where(settings.Key(settingKey)).
		First(s.ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// No custom order exists
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to get highlight order: %w", err)
	}

	var order []string
	err = json.Unmarshal([]byte(setting.Value), &order)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal highlight order: %w", err)
	}

	return order, nil
}

// GetProjectHighlightsForExport gets highlights formatted for export
func (s *HighlightService) GetProjectHighlightsForExport(projectID int) ([]HighlightSegment, error) {
	// Get all video clips for the project with their highlights
	clips, err := s.client.VideoClip.
		Query().
		Where(videoclip.HasProjectWith()).
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
	for i, word := range transcriptWords {
		if word.Start >= timeSeconds {
			return i
		}
	}
	return len(transcriptWords) - 1
}

// TimeToWordIndex converts time in seconds to approximate word index (public method)
func (s *HighlightService) TimeToWordIndex(timeSeconds float64, transcriptWords []schema.Word) int {
	return s.timeToWordIndex(timeSeconds, transcriptWords)
}

// WordIndexToTime converts word index to time in seconds (public method)
func (s *HighlightService) WordIndexToTime(wordIndex int, transcriptWords []schema.Word) float64 {
	return s.wordIndexToTime(wordIndex, transcriptWords)
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
	if startIndex < 0 || endIndex >= len(words) || startIndex > endIndex {
		return ""
	}

	var textParts []string
	for i := startIndex; i <= endIndex; i++ {
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