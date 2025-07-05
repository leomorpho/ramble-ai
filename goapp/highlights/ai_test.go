package highlights

import (
	"testing"

	"MYAPP/ent/schema"
	"github.com/stretchr/testify/assert"
)

// TestParseAIHighlightSuggestionsResponse tests the AI response parsing logic
func TestParseAIHighlightSuggestionsResponse(t *testing.T) {
	service := &AIService{
		highlightService: &HighlightService{},
	}

	transcriptWords := []schema.Word{
		{Word: "Hello", Start: 0.0, End: 0.5},
		{Word: "world", Start: 0.5, End: 1.0},
		{Word: "this", Start: 1.0, End: 1.3},
		{Word: "is", Start: 1.3, End: 1.5},
		{Word: "a", Start: 1.5, End: 1.6},
		{Word: "test", Start: 1.6, End: 2.0},
		{Word: "transcript", Start: 2.0, End: 2.8},
		{Word: "with", Start: 2.8, End: 3.0},
		{Word: "multiple", Start: 3.0, End: 3.5},
		{Word: "words", Start: 3.5, End: 4.0},
	}

	tests := []struct {
		name        string
		aiResponse  string
		expected    []HighlightSuggestion
		shouldError bool
	}{
		{
			name:       "Valid JSON array",
			aiResponse: `[{"start": 0, "end": 2}, {"start": 3, "end": 6}]`,
			expected: []HighlightSuggestion{
				{
					ID:    "suggestion_0_2",
					Start: 0,
					End:   2,
					Text:  "Hello world", // Exclusive end: words 0 and 1
					Color: "#ffeb3b",
				},
				{
					ID:    "suggestion_3_6",
					Start: 3,
					End:   6,
					Text:  "is a test", // Exclusive end: words 3, 4, and 5
					Color: "#81c784",
				},
			},
			shouldError: false,
		},
		{
			name:       "JSON with extra text",
			aiResponse: `Here are the suggestions: [{"start": 1, "end": 3}] That's all!`,
			expected: []HighlightSuggestion{
				{
					ID:    "suggestion_1_3",
					Start: 1,
					End:   3,
					Text:  "world this", // Exclusive end: words 1 and 2
					Color: "#ffeb3b",
				},
			},
			shouldError: false,
		},
		{
			name:        "Invalid JSON",
			aiResponse:  "This is not JSON",
			expected:    nil,
			shouldError: true,
		},
		{
			name:       "Empty array",
			aiResponse: "[]",
			expected:   []HighlightSuggestion{},
			shouldError: false,
		},
		{
			name:       "Invalid indices - start > end",
			aiResponse: `[{"start": 5, "end": 3}]`,
			expected:   []HighlightSuggestion{}, // Should skip invalid suggestion
			shouldError: false,
		},
		{
			name:       "Invalid indices - out of bounds",
			aiResponse: `[{"start": 0, "end": 100}]`,
			expected:   []HighlightSuggestion{}, // Should skip invalid suggestion
			shouldError: false,
		},
		{
			name:       "Single word highlight",
			aiResponse: `[{"start": 2, "end": 3}]`,
			expected: []HighlightSuggestion{
				{
					ID:    "suggestion_2_3",
					Start: 2,
					End:   3,
					Text:  "this", // Just word 2
					Color: "#ffeb3b",
				},
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions, err := service.parseAIHighlightSuggestionsResponse(tt.aiResponse, transcriptWords)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expected), len(suggestions))

				for i, expected := range tt.expected {
					assert.Equal(t, expected.ID, suggestions[i].ID)
					assert.Equal(t, expected.Start, suggestions[i].Start)
					assert.Equal(t, expected.End, suggestions[i].End)
					assert.Equal(t, expected.Text, suggestions[i].Text)
					assert.Equal(t, expected.Color, suggestions[i].Color)
				}
			}
		})
	}
}

// TestSaveSuggestedHighlights tests the saving logic with exclusive end indices
func TestSaveSuggestedHighlights(t *testing.T) {
	// This test demonstrates the expected behavior of the saving logic
	// In a real test, we would use mocks or a test database

	transcriptWords := []schema.Word{
		{Word: "Hello", Start: 0.0, End: 0.5},
		{Word: "world", Start: 0.5, End: 1.0},
		{Word: "this", Start: 1.0, End: 1.3},
		{Word: "is", Start: 1.3, End: 1.5},
		{Word: "a", Start: 1.5, End: 1.6},
		{Word: "test", Start: 1.6, End: 2.0},
	}

	suggestions := []HighlightSuggestion{
		{
			ID:    "suggestion_0_2",
			Start: 0,
			End:   2,   // Exclusive: includes words 0 and 1
			Text:  "Hello world",
			Color: "#ffeb3b",
		},
		{
			ID:    "suggestion_3_5",
			Start: 3,
			End:   5,   // Exclusive: includes words 3 and 4
			Text:  "is a",
			Color: "#81c784",
		},
	}

	// Test the conversion logic that would happen in saveSuggestedHighlights
	service := &HighlightService{}
	
	for _, suggestion := range suggestions {
		startTime := service.wordIndexToTime(suggestion.Start, transcriptWords)
		endTime := service.wordIndexToTime(suggestion.End, transcriptWords)
		if suggestion.End > 0 && suggestion.End <= len(transcriptWords) {
			endTime = transcriptWords[suggestion.End-1].End
		}

		// Verify the conversion
		if suggestion.ID == "suggestion_0_2" {
			assert.Equal(t, 0.0, startTime, "Start time for first suggestion")
			assert.Equal(t, 1.0, endTime, "End time for first suggestion (word[1].End)")
		} else if suggestion.ID == "suggestion_3_5" {
			assert.Equal(t, 1.3, startTime, "Start time for second suggestion")
			assert.Equal(t, 1.6, endTime, "End time for second suggestion (word[4].End)")
		}
	}
}

// TestFilterValidHighlightSuggestions tests the overlap detection logic
func TestFilterValidHighlightSuggestions(t *testing.T) {
	service := &AIService{
		highlightService: &HighlightService{},
	}

	transcriptWords := []schema.Word{
		{Word: "Word0", Start: 0.0, End: 1.0},
		{Word: "Word1", Start: 1.0, End: 2.0},
		{Word: "Word2", Start: 2.0, End: 3.0},
		{Word: "Word3", Start: 3.0, End: 4.0},
		{Word: "Word4", Start: 4.0, End: 5.0},
		{Word: "Word5", Start: 5.0, End: 6.0},
		{Word: "Word6", Start: 6.0, End: 7.0},
		{Word: "Word7", Start: 7.0, End: 8.0},
		{Word: "Word8", Start: 8.0, End: 9.0},
		{Word: "Word9", Start: 9.0, End: 10.0},
	}

	existingHighlights := []schema.Highlight{
		{ID: "existing1", Start: 2.5, End: 4.5}, // Covers parts of words 2, 3, 4
		{ID: "existing2", Start: 7.0, End: 8.0}, // Covers word 7
	}

	suggestions := []HighlightSuggestion{
		{ID: "s1", Start: 0, End: 2, Text: "Word0 Word1"},     // No overlap
		{ID: "s2", Start: 2, End: 4, Text: "Word2 Word3"},     // Overlaps with existing1
		{ID: "s3", Start: 5, End: 7, Text: "Word5 Word6"},     // No overlap
		{ID: "s4", Start: 6, End: 8, Text: "Word6 Word7"},     // Overlaps with existing2
		{ID: "s5", Start: 8, End: 10, Text: "Word8 Word9"},    // No overlap
		{ID: "s6", Start: 1, End: 3, Text: "Word1 Word2"},     // Overlaps with existing1
	}

	validSuggestions := service.filterValidHighlightSuggestions(suggestions, existingHighlights, transcriptWords)

	// Should only include non-overlapping suggestions
	assert.Equal(t, 3, len(validSuggestions))
	
	expectedIDs := map[string]bool{"s1": true, "s3": true, "s5": true}
	for _, vs := range validSuggestions {
		assert.True(t, expectedIDs[vs.ID], "Unexpected suggestion ID: %s", vs.ID)
	}
}

// TestTimeToWordIndexRoundTrip tests the round-trip conversion between word indices and time
func TestTimeToWordIndexRoundTrip(t *testing.T) {
	service := &HighlightService{}

	transcriptWords := []schema.Word{
		{Word: "Word0", Start: 0.0, End: 1.0},
		{Word: "Word1", Start: 1.0, End: 2.0},
		{Word: "Word2", Start: 2.0, End: 3.0},
		{Word: "Word3", Start: 3.0, End: 4.0},
		{Word: "Word4", Start: 4.0, End: 5.0},
	}

	tests := []struct {
		name           string
		startIndex     int
		endIndex       int // Exclusive
		expectedText   string
	}{
		{
			name:         "Single word",
			startIndex:   1,
			endIndex:     2,
			expectedText: "Word1",
		},
		{
			name:         "Multiple words",
			startIndex:   1,
			endIndex:     4,
			expectedText: "Word1 Word2 Word3",
		},
		{
			name:         "All words",
			startIndex:   0,
			endIndex:     5,
			expectedText: "Word0 Word1 Word2 Word3 Word4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Extract text using exclusive end index
			text := service.extractTextFromWordRange(transcriptWords, tt.startIndex, tt.endIndex)
			assert.Equal(t, tt.expectedText, text)

			// Convert to time coordinates (as done in saveSuggestedHighlights)
			startTime := service.wordIndexToTime(tt.startIndex, transcriptWords)
			endTime := service.wordIndexToTime(tt.endIndex, transcriptWords)
			if tt.endIndex > 0 && tt.endIndex <= len(transcriptWords) {
				endTime = transcriptWords[tt.endIndex-1].End
			}

			// Debug logging
			t.Logf("Test %s: Start index %d -> time %f", tt.name, tt.startIndex, startTime)
			t.Logf("Test %s: End index %d -> time %f", tt.name, tt.endIndex, endTime)

			// Convert back to indices (as done in GetSuggestedHighlights)
			recoveredStartIndex := service.timeToWordIndex(startTime, transcriptWords)
			recoveredEndIndex := service.timeToWordIndexForEnd(endTime, transcriptWords)
			
			t.Logf("Test %s: Start time %f -> index %d", tt.name, startTime, recoveredStartIndex)
			t.Logf("Test %s: End time %f -> index %d", tt.name, endTime, recoveredEndIndex)

			// Verify round-trip conversion
			assert.Equal(t, tt.startIndex, recoveredStartIndex, "Start index mismatch")
			assert.Equal(t, tt.endIndex, recoveredEndIndex, "End index mismatch")

			// Verify text extraction still works
			recoveredText := service.extractTextFromWordRange(transcriptWords, recoveredStartIndex, recoveredEndIndex)
			assert.Equal(t, tt.expectedText, recoveredText, "Text mismatch after round-trip")
		})
	}
}

// TestTimeToWordIndexForEnd tests the special end time conversion function
func TestTimeToWordIndexForEnd(t *testing.T) {
	service := &HighlightService{}

	transcriptWords := []schema.Word{
		{Word: "Word0", Start: 0.0, End: 1.0},
		{Word: "Word1", Start: 1.0, End: 2.0},
		{Word: "Word2", Start: 2.0, End: 3.0},
		{Word: "Word3", Start: 3.0, End: 4.0},
		{Word: "Word4", Start: 4.0, End: 5.0},
	}

	tests := []struct {
		name          string
		endTime       float64
		expectedIndex int
	}{
		{
			name:          "End of first word",
			endTime:       1.0,
			expectedIndex: 1, // Exclusive index for word 0
		},
		{
			name:          "End of middle word",
			endTime:       3.0,
			expectedIndex: 3, // Exclusive index for words 0-2
		},
		{
			name:          "End of last word",
			endTime:       5.0,
			expectedIndex: 5, // Exclusive index for all words
		},
		{
			name:          "Time between words",
			endTime:       2.5,
			expectedIndex: 2, // Should return index of word 2
		},
		{
			name:          "Time before all words",
			endTime:       -1.0,
			expectedIndex: 0,
		},
		{
			name:          "Time with small epsilon",
			endTime:       1.0001,
			expectedIndex: 1, // Should still match word 0's end time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index := service.timeToWordIndexForEnd(tt.endTime, transcriptWords)
			assert.Equal(t, tt.expectedIndex, index)
		})
	}
}