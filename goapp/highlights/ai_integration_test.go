package highlights

import (
	"testing"

	"ramble-ai/ent/schema"
	"github.com/stretchr/testify/assert"
)

// TestAIHighlightSuggestionsEndToEnd simulates the complete flow from LLM response to final output
func TestAIHighlightSuggestionsEndToEnd(t *testing.T) {
	// Setup test data
	transcriptWords := []schema.Word{
		{Word: "What", Start: 422.620, End: 422.780},
		{Word: "is", Start: 422.780, End: 422.860},
		{Word: "your", Start: 422.860, End: 422.980},
		{Word: "experience", Start: 422.980, End: 423.420},
		{Word: "with", Start: 423.420, End: 423.540},
		{Word: "toxic", Start: 423.540, End: 423.820},
		{Word: "shame", Start: 423.820, End: 424.120},
		{Word: "How", Start: 424.120, End: 424.260},
		{Word: "has", Start: 424.260, End: 424.380},
		{Word: "it", Start: 424.380, End: 424.460},
		{Word: "affected", Start: 424.460, End: 424.860},
		{Word: "your", Start: 424.860, End: 424.980},
		{Word: "life", Start: 424.980, End: 425.220},
		{Word: "and", Start: 425.220, End: 425.340},
		{Word: "what", Start: 425.340, End: 425.460},
		{Word: "have", Start: 425.460, End: 425.580},
		{Word: "you", Start: 425.580, End: 425.700},
		{Word: "done", Start: 425.700, End: 425.940},
		{Word: "to", Start: 425.940, End: 426.060},
		{Word: "successfully", Start: 426.060, End: 426.580},
		{Word: "perhaps", Start: 426.580, End: 426.940},
		{Word: "revert", Start: 426.940, End: 427.280},
		{Word: "it", Start: 427.280, End: 431.420},
	}

	// Simulate AI service
	aiService := &AIService{
		highlightService: &HighlightService{},
	}

	t.Run("Complete flow from LLM response to text extraction", func(t *testing.T) {
		// Step 1: Simulate LLM response
		llmResponse := `[{"start": 0, "end": 7}, {"start": 7, "end": 13}, {"start": 14, "end": 22}]`

		// Step 2: Parse AI response
		suggestions, err := aiService.parseAIHighlightSuggestionsResponse(llmResponse, transcriptWords)
		assert.NoError(t, err)
		assert.Len(t, suggestions, 3)

		// Verify parsed suggestions
		assert.Equal(t, "suggestion_0_7", suggestions[0].ID)
		assert.Equal(t, "What is your experience with toxic shame", suggestions[0].Text)
		assert.Equal(t, 0, suggestions[0].Start)
		assert.Equal(t, 7, suggestions[0].End)

		assert.Equal(t, "suggestion_7_13", suggestions[1].ID)
		assert.Equal(t, "How has it affected your life", suggestions[1].Text)
		assert.Equal(t, 7, suggestions[1].Start)
		assert.Equal(t, 13, suggestions[1].End)

		assert.Equal(t, "suggestion_14_22", suggestions[2].ID)
		assert.Equal(t, "what have you done to successfully perhaps revert", suggestions[2].Text)
		assert.Equal(t, 14, suggestions[2].Start)
		assert.Equal(t, 22, suggestions[2].End)

		// Step 3: Simulate saving to database (convert to time coordinates)
		service := &HighlightService{}
		savedHighlights := []schema.Highlight{}

		for _, suggestion := range suggestions {
			startTime := service.wordIndexToTime(suggestion.Start, transcriptWords)
			endTime := service.wordIndexToTime(suggestion.End, transcriptWords)
			if suggestion.End > 0 && suggestion.End <= len(transcriptWords) {
				endTime = transcriptWords[suggestion.End-1].End
			}

			highlight := schema.Highlight{
				ID:      suggestion.ID,
				Start:   startTime,
				End:     endTime,
				ColorID: suggestion.ColorID,
			}
			savedHighlights = append(savedHighlights, highlight)
		}

		// Verify saved time coordinates
		assert.Equal(t, 422.620, savedHighlights[0].Start)
		assert.Equal(t, 424.120, savedHighlights[0].End) // End of "shame" (word 6)
		assert.Equal(t, 424.120, savedHighlights[1].Start)
		assert.Equal(t, 425.220, savedHighlights[1].End) // End of "life" (word 12)
		assert.Equal(t, 425.340, savedHighlights[2].Start)
		assert.Equal(t, 427.280, savedHighlights[2].End) // End of "revert" (word 21)

		// Step 4: Simulate loading from database (convert back to word indices)
		loadedSuggestions := []HighlightSuggestion{}
		for _, h := range savedHighlights {
			startIndex := service.timeToWordIndex(h.Start, transcriptWords)
			endIndex := service.timeToWordIndexForEnd(h.End, transcriptWords)

			text := service.extractTextFromWordRange(transcriptWords, startIndex, endIndex)

			suggestion := HighlightSuggestion{
				ID:      h.ID,
				Start:   startIndex,
				End:     endIndex,
				Text:    text,
				ColorID: h.ColorID,
			}
			loadedSuggestions = append(loadedSuggestions, suggestion)
		}

		// Verify round-trip conversion maintains correct data
		assert.Len(t, loadedSuggestions, 3)

		// First suggestion
		assert.Equal(t, "suggestion_0_7", loadedSuggestions[0].ID)
		assert.Equal(t, 0, loadedSuggestions[0].Start)
		assert.Equal(t, 7, loadedSuggestions[0].End)
		assert.Equal(t, "What is your experience with toxic shame", loadedSuggestions[0].Text)

		// Second suggestion
		assert.Equal(t, "suggestion_7_13", loadedSuggestions[1].ID)
		assert.Equal(t, 7, loadedSuggestions[1].Start)
		assert.Equal(t, 13, loadedSuggestions[1].End)
		assert.Equal(t, "How has it affected your life", loadedSuggestions[1].Text)

		// Third suggestion
		assert.Equal(t, "suggestion_14_22", loadedSuggestions[2].ID)
		assert.Equal(t, 14, loadedSuggestions[2].Start)
		assert.Equal(t, 22, loadedSuggestions[2].End)
		assert.Equal(t, "what have you done to successfully perhaps revert", loadedSuggestions[2].Text)
	})

	t.Run("Edge cases with word boundaries", func(t *testing.T) {
		// Test case where highlights end exactly at word boundaries
		llmResponse := `[{"start": 361, "end": 382}, {"start": 415, "end": 445}]`

		// Create a different set of words for this test
		edgeCaseWords := make([]schema.Word, 500)
		for i := 0; i < 500; i++ {
			edgeCaseWords[i] = schema.Word{
				Word:  "word" + string(rune(i)),
				Start: float64(i) * 0.5,
				End:   float64(i)*0.5 + 0.4,
			}
		}

		// Parse response
		suggestions, err := aiService.parseAIHighlightSuggestionsResponse(llmResponse, edgeCaseWords)
		assert.NoError(t, err)
		assert.Len(t, suggestions, 2)

		// Simulate save and load cycle
		service := &HighlightService{}
		for _, suggestion := range suggestions {
			// Save (convert to time)
			startTime := service.wordIndexToTime(suggestion.Start, edgeCaseWords)
			endTime := service.wordIndexToTime(suggestion.End, edgeCaseWords)
			if suggestion.End > 0 && suggestion.End <= len(edgeCaseWords) {
				endTime = edgeCaseWords[suggestion.End-1].End
			}

			// Load (convert back to indices)
			recoveredStart := service.timeToWordIndex(startTime, edgeCaseWords)
			recoveredEnd := service.timeToWordIndexForEnd(endTime, edgeCaseWords)

			// Verify round-trip
			assert.Equal(t, suggestion.Start, recoveredStart, "Start index should match after round-trip")
			assert.Equal(t, suggestion.End, recoveredEnd, "End index should match after round-trip")

			// Verify text extraction
			originalText := service.extractTextFromWordRange(edgeCaseWords, suggestion.Start, suggestion.End)
			recoveredText := service.extractTextFromWordRange(edgeCaseWords, recoveredStart, recoveredEnd)
			assert.Equal(t, originalText, recoveredText, "Text should match after round-trip")
		}
	})

	t.Run("Overlap detection with real coordinates", func(t *testing.T) {
		// Create existing highlights
		existingHighlights := []schema.Highlight{
			{ID: "existing1", Start: 423.0, End: 424.0},
			{ID: "existing2", Start: 426.0, End: 427.0},
		}

		// LLM suggests some overlapping and non-overlapping highlights
		llmResponse := `[
			{"start": 0, "end": 3},
			{"start": 2, "end": 5},
			{"start": 10, "end": 15},
			{"start": 18, "end": 22}
		]`

		suggestions, err := aiService.parseAIHighlightSuggestionsResponse(llmResponse, transcriptWords)
		assert.NoError(t, err)

		// Filter overlapping suggestions
		validSuggestions := aiService.filterValidHighlightSuggestions(suggestions, existingHighlights, transcriptWords)

		// Should filter out suggestions that overlap with existing highlights
		assert.Less(t, len(validSuggestions), len(suggestions))

		// Verify no overlaps remain
		service := &HighlightService{}
		for _, suggestion := range validSuggestions {
			suggestionStartTime := service.wordIndexToTime(suggestion.Start, transcriptWords)
			suggestionEndTime := service.wordIndexToTime(suggestion.End, transcriptWords)
			if suggestion.End > 0 && suggestion.End <= len(transcriptWords) {
				suggestionEndTime = transcriptWords[suggestion.End-1].End
			}

			for _, existing := range existingHighlights {
				// Verify no overlap
				hasOverlap := suggestionStartTime < existing.End && suggestionEndTime > existing.Start
				assert.False(t, hasOverlap, "Valid suggestion should not overlap with existing highlights")
			}
		}
	})
}

// TestRealWorldScenario tests with actual data that caused issues
func TestRealWorldScenario(t *testing.T) {
	// This is based on the actual problematic data from the logs
	transcriptWords := []schema.Word{
		{Word: "I", Start: 288.04, End: 288.10},
		{Word: "can", Start: 288.10, End: 288.26},
		{Word: "see", Start: 288.26, End: 288.46},
		{Word: "enormous", Start: 288.46, End: 288.94},
		{Word: "changes", Start: 288.94, End: 289.40},
		{Word: "in", Start: 289.40, End: 289.48},
		{Word: "my", Start: 289.48, End: 289.62},
		{Word: "behavior", Start: 289.62, End: 290.08},
		{Word: "and", Start: 290.08, End: 290.20},
		{Word: "in", Start: 290.20, End: 290.28},
		{Word: "what", Start: 290.28, End: 290.44},
		{Word: "I", Start: 290.44, End: 290.52},
		{Word: "do", Start: 290.52, End: 290.68},
		{Word: "now", Start: 290.68, End: 290.94},
		{Word: "and", Start: 290.94, End: 291.06},
		{Word: "how", Start: 291.06, End: 291.22},
		{Word: "I", Start: 291.22, End: 291.30},
		{Word: "react", Start: 291.30, End: 291.64},
		{Word: "to", Start: 291.64, End: 291.74},
		{Word: "things", Start: 291.74, End: 292.06},
		{Word: "simply", Start: 292.06, End: 292.42},
		{Word: "because", Start: 292.42, End: 292.74},
		{Word: "I've", Start: 292.74, End: 292.94},
		{Word: "become", Start: 292.94, End: 293.24},
		{Word: "aware", Start: 293.24, End: 293.52},
		{Word: "of", Start: 293.52, End: 293.62},
		{Word: "the", Start: 293.62, End: 293.72},
		{Word: "shame", Start: 293.72, End: 294.02},
		{Word: "which", Start: 294.02, End: 294.18},
		{Word: "is", Start: 294.18, End: 294.30},
		{Word: "remarkable", Start: 294.30, End: 294.86},
		{Word: "I", Start: 294.86, End: 294.94}, // This is the extra word that was being included
	}

	// Adjust indices to start from 415
	adjustedWords := make([]schema.Word, 446)
	for i := 0; i < 415; i++ {
		adjustedWords[i] = schema.Word{Word: "placeholder" + string(rune(i)), Start: float64(i) * 0.5, End: float64(i)*0.5 + 0.4}
	}
	copy(adjustedWords[415:], transcriptWords)

	aiService := &AIService{
		highlightService: &HighlightService{},
	}

	// LLM response suggesting words 415-445
	llmResponse := `[{"start": 415, "end": 445}]`

	suggestions, err := aiService.parseAIHighlightSuggestionsResponse(llmResponse, adjustedWords)
	assert.NoError(t, err)
	assert.Len(t, suggestions, 1)

	suggestion := suggestions[0]
	assert.Equal(t, 415, suggestion.Start)
	assert.Equal(t, 445, suggestion.End)

	// The text should NOT include the word at index 445
	expectedText := "I can see enormous changes in my behavior and in what I do now and how I react to things simply because I've become aware of the shame which is"
	assert.Equal(t, expectedText, suggestion.Text)
	assert.NotContains(t, suggestion.Text, "remarkable I", "Should not include the word at index 445")

	// Simulate save/load cycle
	service := &HighlightService{}

	// Save
	startTime := service.wordIndexToTime(suggestion.Start, adjustedWords)
	endTime := service.wordIndexToTime(suggestion.End, adjustedWords)
	if suggestion.End > 0 && suggestion.End <= len(adjustedWords) {
		endTime = adjustedWords[suggestion.End-1].End
	}

	// Load
	recoveredStart := service.timeToWordIndex(startTime, adjustedWords)
	recoveredEnd := service.timeToWordIndexForEnd(endTime, adjustedWords)

	assert.Equal(t, 415, recoveredStart)
	assert.Equal(t, 445, recoveredEnd)

	// Extract text again
	recoveredText := service.extractTextFromWordRange(adjustedWords, recoveredStart, recoveredEnd)
	assert.Equal(t, expectedText, recoveredText)
	assert.NotContains(t, recoveredText, "remarkable I", "Recovered text should not include extra words")
}
