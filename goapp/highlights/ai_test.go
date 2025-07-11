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

// TestFlattenConsecutiveNewlines tests the consecutive newline flattening logic
func TestFlattenConsecutiveNewlines(t *testing.T) {
	service := &AIService{
		highlightService: &HighlightService{},
	}

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "No newlines",
			input:    []string{"id1", "id2", "id3"},
			expected: []string{"id1", "id2", "id3"},
		},
		{
			name:     "Single newline",
			input:    []string{"id1", "N", "id2"},
			expected: []string{"id1", "N", "id2"},
		},
		{
			name:     "Two consecutive newlines",
			input:    []string{"id1", "N", "N", "id2"},
			expected: []string{"id1", "N", "id2"},
		},
		{
			name:     "Three consecutive newlines",
			input:    []string{"id1", "N", "N", "N", "id2"},
			expected: []string{"id1", "N", "id2"},
		},
		{
			name:     "Multiple separated newlines",
			input:    []string{"id1", "N", "id2", "N", "id3"},
			expected: []string{"id1", "N", "id2", "N", "id3"},
		},
		{
			name:     "Consecutive newlines at start",
			input:    []string{"N", "N", "id1", "id2"},
			expected: []string{"N", "id1", "id2"},
		},
		{
			name:     "Consecutive newlines at end",
			input:    []string{"id1", "id2", "N", "N"},
			expected: []string{"id1", "id2", "N"},
		},
		{
			name:     "Mixed consecutive newlines",
			input:    []string{"id1", "N", "N", "id2", "N", "N", "N", "id3", "N", "id4"},
			expected: []string{"id1", "N", "id2", "N", "id3", "N", "id4"},
		},
		{
			name:     "All newlines",
			input:    []string{"N", "N", "N", "N"},
			expected: []string{"N"},
		},
		{
			name:     "Empty array",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "Single element",
			input:    []string{"id1"},
			expected: []string{"id1"},
		},
		{
			name:     "Single newline",
			input:    []string{"N"},
			expected: []string{"N"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.flattenConsecutiveNewlines(tt.input)
			
			assert.Equal(t, len(tt.expected), len(result))
			for i, expected := range tt.expected {
				assert.Equal(t, expected, result[i])
			}
		})
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

// TestParseAIReorderingResponse tests the AI reordering response parsing with mixed types
func TestParseAIReorderingResponse(t *testing.T) {
	service := &AIService{
		highlightService: &HighlightService{},
	}

	tests := []struct {
		name     string
		input    string
		expected []interface{}
		wantErr  bool
	}{
		{
			name:     "Valid response with N characters",
			input:    `["id1", "id2", "N", "id3", "id4", "N", "id5"]`,
			expected: []interface{}{"id1", "id2", "N", "id3", "id4", "N", "id5"},
			wantErr:  false,
		},
		{
			name:     "Valid response without N characters",
			input:    `["id1", "id2", "id3", "id4", "id5"]`,
			expected: []interface{}{"id1", "id2", "id3", "id4", "id5"},
			wantErr:  false,
		},
		{
			name:     "Response with section objects",
			input:    `["id1", {"type":"N","title":"Section 1"}, "id2", {"type":"N","title":"Section 2"}, "id3"]`,
			expected: []interface{}{"id1", map[string]interface{}{"type": "N", "title": "Section 1"}, "id2", map[string]interface{}{"type": "N", "title": "Section 2"}, "id3"},
			wantErr:  false,
		},
		{
			name:     "Mixed strings, N markers, and section objects",
			input:    `["highlight_1", "N", "highlight_2", {"type":"N","title":"Main Content"}, "highlight_3"]`,
			expected: []interface{}{"highlight_1", "N", "highlight_2", map[string]interface{}{"type": "N", "title": "Main Content"}, "highlight_3"},
			wantErr:  false,
		},
		{
			name:     "Response with markdown formatting",
			input:    "```json\n[\"id1\", \"id2\", \"N\", \"id3\"]\n```",
			expected: []interface{}{"id1", "id2", "N", "id3"},
			wantErr:  false,
		},
		{
			name:     "Response with extra text",
			input:    `Here is the reordered array: ["id1", "N", "id2"] - this should work well.`,
			expected: []interface{}{"id1", "N", "id2"},
			wantErr:  false,
		},
		{
			name:    "Invalid JSON",
			input:   `["id1", "id2", "N", "id3"`,
			wantErr: true,
		},
		{
			name:    "No JSON array found",
			input:   `This is just text without any JSON array.`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.parseAIReorderingResponse(tt.input)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result), "Result length mismatch")

			for i, expected := range tt.expected {
				switch expectedVal := expected.(type) {
				case string:
					actualVal, ok := result[i].(string)
					assert.True(t, ok, "Expected string at index %d, got %T", i, result[i])
					assert.Equal(t, expectedVal, actualVal, "String mismatch at index %d", i)
				case map[string]interface{}:
					actualVal, ok := result[i].(map[string]interface{})
					assert.True(t, ok, "Expected map at index %d, got %T", i, result[i])
					assert.Equal(t, expectedVal["type"], actualVal["type"], "Type field mismatch at index %d", i)
					if title, hasTitle := expectedVal["title"]; hasTitle {
						assert.Equal(t, title, actualVal["title"], "Title field mismatch at index %d", i)
					}
				default:
					assert.Equal(t, expected, result[i], "Value mismatch at index %d", i)
				}
			}
		})
	}
}

// TestProcessReorderedItems tests the processing logic for reordered items
func TestProcessReorderedItems(t *testing.T) {
	service := &AIService{
		highlightService: &HighlightService{},
	}

	tests := []struct {
		name        string
		reorderedItems []interface{}
		originalIDs []string
		expected    []interface{}
	}{
		{
			name:        "No duplicates with N characters",
			reorderedItems: []interface{}{"id1", "id2", "N", "id3", "id4"},
			originalIDs: []string{"id1", "id2", "id3", "id4"},
			expected:    []interface{}{"id1", "id2", "N", "id3", "id4"},
		},
		{
			name:        "Duplicates with N characters",
			reorderedItems: []interface{}{"id1", "id2", "N", "id1", "id3", "N", "id2", "id4"},
			originalIDs: []string{"id1", "id2", "id3", "id4"},
			expected:    []interface{}{"id1", "id2", "N", "id3", "N", "id4"},
		},
		{
			name:        "Missing ID gets added at end",
			reorderedItems: []interface{}{"id1", "N", "id3"},
			originalIDs: []string{"id1", "id2", "id3"},
			expected:    []interface{}{"id1", "N", "id3", "id2"},
		},
		{
			name:        "Unknown IDs get filtered out",
			reorderedItems: []interface{}{"id1", "unknown", "N", "id2", "another_unknown"},
			originalIDs: []string{"id1", "id2"},
			expected:    []interface{}{"id1", "N", "id2"},
		},
		{
			name:        "Complex case with duplicates and unknowns",
			reorderedItems: []interface{}{"id1", "id2", "N", "id1", "unknown", "id3", "N", "id2", "id4"},
			originalIDs: []string{"id1", "id2", "id3", "id4"},
			expected:    []interface{}{"id1", "id2", "N", "id3", "N", "id4"},
		},
		{
			name:        "Section objects are preserved",
			reorderedItems: []interface{}{"id1", map[string]interface{}{"type": "N", "title": "Section 1"}, "id2", "N", "id3"},
			originalIDs: []string{"id1", "id2", "id3"},
			expected:    []interface{}{"id1", map[string]interface{}{"type": "N", "title": "Section 1"}, "id2", "N", "id3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.processReorderedItems(tt.reorderedItems, tt.originalIDs)
			
			assert.Equal(t, len(tt.expected), len(result))

			for i, expected := range tt.expected {
				switch expectedVal := expected.(type) {
				case string:
					actualVal, ok := result[i].(string)
					assert.True(t, ok, "Expected string at index %d, got %T", i, result[i])
					assert.Equal(t, expectedVal, actualVal, "String mismatch at index %d", i)
				case map[string]interface{}:
					actualVal, ok := result[i].(map[string]interface{})
					assert.True(t, ok, "Expected map at index %d, got %T", i, result[i])
					assert.Equal(t, expectedVal["type"], actualVal["type"], "Type field mismatch at index %d", i)
					if title, hasTitle := expectedVal["title"]; hasTitle {
						assert.Equal(t, title, actualVal["title"], "Title field mismatch at index %d", i)
					}
				default:
					assert.Equal(t, expected, result[i], "Value mismatch at index %d", i)
				}
			}
		})
	}
}

// TestAIReorderingWithNCharacters tests the complete flow from LLM response to final output
func TestAIReorderingWithNCharacters(t *testing.T) {
	service := &AIService{
		highlightService: &HighlightService{},
	}

	// Test the complete flow from LLM response to final output
	tests := []struct {
		name         string
		llmResponse  string
		originalIDs  []string
		expectedFinal []interface{}
	}{
		{
			name:         "Perfect LLM response with N characters",
			llmResponse:  `["highlight_1", "highlight_2", "N", "highlight_3", "highlight_4"]`,
			originalIDs:  []string{"highlight_1", "highlight_2", "highlight_3", "highlight_4"},
			expectedFinal: []interface{}{"highlight_1", "highlight_2", "N", "highlight_3", "highlight_4"},
		},
		{
			name:         "LLM response with duplicates",
			llmResponse:  `["highlight_1", "highlight_2", "N", "highlight_1", "highlight_3", "N", "highlight_2", "highlight_4"]`,
			originalIDs:  []string{"highlight_1", "highlight_2", "highlight_3", "highlight_4"},
			expectedFinal: []interface{}{"highlight_1", "highlight_2", "N", "highlight_3", "N", "highlight_4"},
		},
		{
			name:         "LLM response with missing ID",
			llmResponse:  `["highlight_1", "N", "highlight_3"]`,
			originalIDs:  []string{"highlight_1", "highlight_2", "highlight_3"},
			expectedFinal: []interface{}{"highlight_1", "N", "highlight_3", "highlight_2"},
		},
		{
			name:         "LLM response with unknown and duplicate IDs",
			llmResponse:  `["highlight_1", "unknown_id", "N", "highlight_1", "highlight_2", "N", "highlight_3"]`,
			originalIDs:  []string{"highlight_1", "highlight_2", "highlight_3"},
			expectedFinal: []interface{}{"highlight_1", "N", "highlight_2", "N", "highlight_3"},
		},
		{
			name:         "Real-world example from logs",
			llmResponse:  `["highlight_1752086566479_yioxdt6gz", "highlight_1752086557450_lr1swkjaj", "N", "highlight_1752086566479_yioxdt6gz", "highlight_1752086564341_wmna40inb", "highlight_1752086565468_egogda0qe", "N", "highlight_1752086562788_zo9mju0n2", "highlight_1752086566479_yioxdt6gz", "highlight_1752086567716_54jz3puhk"]`,
			originalIDs:  []string{"highlight_1752086566479_yioxdt6gz", "highlight_1752086557450_lr1swkjaj", "highlight_1752086564341_wmna40inb", "highlight_1752086565468_egogda0qe", "highlight_1752086562788_zo9mju0n2", "highlight_1752086567716_54jz3puhk"},
			expectedFinal: []interface{}{"highlight_1752086566479_yioxdt6gz", "highlight_1752086557450_lr1swkjaj", "N", "highlight_1752086564341_wmna40inb", "highlight_1752086565468_egogda0qe", "N", "highlight_1752086562788_zo9mju0n2", "highlight_1752086567716_54jz3puhk"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the LLM response
			parsed, err := service.parseAIReorderingResponse(tt.llmResponse)
			assert.NoError(t, err)

			// Process the items
			result := service.processReorderedItems(parsed, tt.originalIDs)

			// Check the final result
			assert.Equal(t, len(tt.expectedFinal), len(result))

			for i, expected := range tt.expectedFinal {
				switch expectedVal := expected.(type) {
				case string:
					actualVal, ok := result[i].(string)
					assert.True(t, ok, "Expected string at index %d, got %T", i, result[i])
					assert.Equal(t, expectedVal, actualVal, "String mismatch at index %d", i)
				case map[string]interface{}:
					actualVal, ok := result[i].(map[string]interface{})
					assert.True(t, ok, "Expected map at index %d, got %T", i, result[i])
					assert.Equal(t, expectedVal["type"], actualVal["type"], "Type field mismatch at index %d", i)
					if title, hasTitle := expectedVal["title"]; hasTitle {
						assert.Equal(t, title, actualVal["title"], "Title field mismatch at index %d", i)
					}
				default:
					assert.Equal(t, expected, result[i], "Value mismatch at index %d", i)
				}
			}

			// Verify that all original IDs are present exactly once
			idCount := make(map[string]int)
			for _, item := range result {
				if str, ok := item.(string); ok && str != "N" {
					idCount[str]++
				}
			}

			for _, originalID := range tt.originalIDs {
				assert.Equal(t, 1, idCount[originalID], "Expected original ID %s to appear exactly once", originalID)
			}
		})
	}
}