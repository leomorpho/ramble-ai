package highlights

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"ramble-ai/ent"
	"ramble-ai/ent/schema"
)

// Test file focusing on edge cases, error conditions, and boundary scenarios

func TestHighlightService_extractHighlightText_EdgeCases(t *testing.T) {
	service := &HighlightService{}

	tests := []struct {
		name         string
		highlight    schema.Highlight
		words        []schema.Word
		fullText     string
		expectedText string
		description  string
	}{
		{
			name: "Highlight with zero duration",
			highlight: schema.Highlight{
				ID:      "h1",
				Start:   1.0,
				End:     1.0,
				ColorID: 3,
			},
			words: []schema.Word{
				{Word: "Hello", Start: 0.0, End: 0.5},
				{Word: "world", Start: 0.5, End: 1.0},
				{Word: "test", Start: 1.0, End: 1.5},
			},
			fullText:     "Hello world test",
			expectedText: "",
			description:  "Should return empty string for zero duration highlight",
		},
		{
			name: "Highlight before transcript starts",
			highlight: schema.Highlight{
				ID:      "h2",
				Start:   -1.0,
				End:     0.0,
				ColorID: 2,
			},
			words: []schema.Word{
				{Word: "Hello", Start: 0.0, End: 0.5},
				{Word: "world", Start: 0.5, End: 1.0},
			},
			fullText:     "Hello world",
			expectedText: "",
			description:  "Should return empty string for highlight before transcript",
		},
		{
			name: "Highlight after transcript ends",
			highlight: schema.Highlight{
				ID:      "h3",
				Start:   10.0,
				End:     11.0,
				ColorID: 3,
			},
			words: []schema.Word{
				{Word: "Hello", Start: 0.0, End: 0.5},
				{Word: "world", Start: 0.5, End: 1.0},
			},
			fullText:     "Hello world",
			expectedText: "",
			description:  "Should return empty string for highlight after transcript",
		},
		{
			name: "Highlight with fractional word overlap",
			highlight: schema.Highlight{
				ID:      "h4",
				Start:   0.25,
				End:     0.75,
				ColorID: 1,
			},
			words: []schema.Word{
				{Word: "Hello", Start: 0.0, End: 0.5},
				{Word: "world", Start: 0.5, End: 1.0},
				{Word: "test", Start: 1.0, End: 1.5},
			},
			fullText:     "Hello world test",
			expectedText: "llo wo",
			description:  "Should use fallback text extraction when no words match exactly",
		},
		{
			name: "Words with identical timestamps",
			highlight: schema.Highlight{
				ID:      "h5",
				Start:   1.0,
				End:     1.0,
				ColorID: 3,
			},
			words: []schema.Word{
				{Word: "Hello", Start: 1.0, End: 1.0},
				{Word: "world", Start: 1.0, End: 1.0},
				{Word: "test", Start: 1.0, End: 1.0},
			},
			fullText:     "Hello world test",
			expectedText: "Hello world test",
			description:  "Should handle words with identical timestamps",
		},
		{
			name: "Empty words with non-empty full text",
			highlight: schema.Highlight{
				ID:      "h6",
				Start:   0.0,
				End:     1.0,
				ColorID: 5,
			},
			words:        []schema.Word{},
			fullText:     "Hello world test",
			expectedText: "",
			description:  "Should return empty string when no words but full text exists",
		},
		{
			name: "Single character words",
			highlight: schema.Highlight{
				ID:      "h7",
				Start:   0.0,
				End:     1.0,
				ColorID: 2,
			},
			words: []schema.Word{
				{Word: "I", Start: 0.0, End: 0.2},
				{Word: "a", Start: 0.2, End: 0.4},
				{Word: "m", Start: 0.4, End: 0.6},
				{Word: "ok", Start: 0.6, End: 1.0},
			},
			fullText:     "I a m ok",
			expectedText: "I a m ok",
			description:  "Should handle single character words",
		},
		{
			name: "Words with special characters",
			highlight: schema.Highlight{
				ID:      "h8",
				Start:   0.0,
				End:     2.0,
				ColorID: 4,
			},
			words: []schema.Word{
				{Word: "Hello,", Start: 0.0, End: 0.5},
				{Word: "world!", Start: 0.5, End: 1.0},
				{Word: "How're", Start: 1.0, End: 1.5},
				{Word: "you?", Start: 1.5, End: 2.0},
			},
			fullText:     "Hello, world! How're you?",
			expectedText: "Hello, world! How're you?",
			description:  "Should handle words with punctuation and contractions",
		},
		{
			name: "Highlight with microsecond precision",
			highlight: schema.Highlight{
				ID:      "h9",
				Start:   0.001,
				End:     0.999,
				ColorID: 2,
			},
			words: []schema.Word{
				{Word: "Hello", Start: 0.0, End: 0.5},
				{Word: "world", Start: 0.5, End: 1.0},
			},
			fullText:     "Hello world",
			expectedText: "Hello worl",
			description:  "Should handle microsecond precision in timing",
		},
		{
			name: "Overlapping word timestamps",
			highlight: schema.Highlight{
				ID:      "h10",
				Start:   0.0,
				End:     1.0,
				ColorID: 4,
			},
			words: []schema.Word{
				{Word: "Hello", Start: 0.0, End: 0.7},
				{Word: "world", Start: 0.3, End: 1.0}, // Overlapping with "Hello"
			},
			fullText:     "Hello world",
			expectedText: "Hello world",
			description:  "Should handle overlapping word timestamps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.extractHighlightText(tt.highlight, tt.words, tt.fullText)
			assert.Equal(t, tt.expectedText, result, tt.description)
		})
	}
}

func TestHighlightService_timeToWordIndex_BoundaryConditions(t *testing.T) {
	service := &HighlightService{}

	tests := []struct {
		name          string
		timeSeconds   float64
		words         []schema.Word
		expectedIndex int
		description   string
	}{
		{
			name:          "Empty words array",
			timeSeconds:   1.0,
			words:         []schema.Word{},
			expectedIndex: -1,
			description:   "Should return -1 for empty words array",
		},
		{
			name:        "Time exactly at word boundary",
			timeSeconds: 1.0,
			words: []schema.Word{
				{Word: "Hello", Start: 0.0, End: 0.5},
				{Word: "world", Start: 1.0, End: 1.5},
			},
			expectedIndex: 1,
			description:   "Should return correct index for exact word boundary",
		},
		{
			name:        "Time before first word",
			timeSeconds: -1.0,
			words: []schema.Word{
				{Word: "Hello", Start: 0.0, End: 0.5},
				{Word: "world", Start: 1.0, End: 1.5},
			},
			expectedIndex: 0,
			description:   "Should return first word index for negative time",
		},
		{
			name:        "Time after last word",
			timeSeconds: 10.0,
			words: []schema.Word{
				{Word: "Hello", Start: 0.0, End: 0.5},
				{Word: "world", Start: 1.0, End: 1.5},
			},
			expectedIndex: 1,
			description:   "Should return last word index for time after transcript",
		},
		{
			name:        "Words with same start time",
			timeSeconds: 1.0,
			words: []schema.Word{
				{Word: "Hello", Start: 1.0, End: 1.2},
				{Word: "world", Start: 1.0, End: 1.3},
				{Word: "test", Start: 1.0, End: 1.4},
			},
			expectedIndex: 0,
			description:   "Should return first word when multiple words have same start time",
		},
		{
			name:        "Very small time differences",
			timeSeconds: 1.0001,
			words: []schema.Word{
				{Word: "Hello", Start: 0.0, End: 0.5},
				{Word: "world", Start: 1.0, End: 1.5},
				{Word: "test", Start: 1.0002, End: 1.6},
			},
			expectedIndex: 1,
			description:   "Should return word that contains the time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.timeToWordIndex(tt.timeSeconds, tt.words)
			assert.Equal(t, tt.expectedIndex, result, tt.description)
		})
	}
}

func TestHighlightService_ApplyHighlightOrder_EdgeCases(t *testing.T) {
	service := &HighlightService{}

	tests := []struct {
		name          string
		segments      []HighlightSegment
		order         []string
		expectedOrder []string
		description   string
	}{
		{
			name:          "Empty segments array",
			segments:      []HighlightSegment{},
			order:         []string{"h1", "h2"},
			expectedOrder: nil,
			description:   "Should return empty array for empty segments",
		},
		{
			name: "Order with all non-existent IDs",
			segments: []HighlightSegment{
				{ID: "h1", VideoClipName: "Video1"},
				{ID: "h2", VideoClipName: "Video2"},
			},
			order:         []string{"h3", "h4", "h5"},
			expectedOrder: []string{"h1", "h2"},
			description:   "Should maintain stable sort when all order IDs don't exist",
		},
		{
			name: "Duplicate segments",
			segments: []HighlightSegment{
				{ID: "h1", VideoClipName: "Video1"},
				{ID: "h1", VideoClipName: "Video1_duplicate"},
				{ID: "h2", VideoClipName: "Video2"},
			},
			order:         []string{"h2", "h1"},
			expectedOrder: []string{"h2", "h1", "h1"},
			description:   "Should handle duplicate segment IDs",
		},
		{
			name: "Single segment",
			segments: []HighlightSegment{
				{ID: "h1", VideoClipName: "Video1"},
			},
			order:         []string{"h1"},
			expectedOrder: []string{"h1"},
			description:   "Should handle single segment correctly",
		},
		{
			name: "Order longer than segments",
			segments: []HighlightSegment{
				{ID: "h1", VideoClipName: "Video1"},
				{ID: "h2", VideoClipName: "Video2"},
			},
			order:         []string{"h2", "h1", "h3", "h4", "h5"},
			expectedOrder: []string{"h2", "h1"},
			description:   "Should handle order longer than available segments",
		},
		{
			name: "Mixed case IDs",
			segments: []HighlightSegment{
				{ID: "H1", VideoClipName: "Video1"},
				{ID: "h2", VideoClipName: "Video2"},
			},
			order:         []string{"h1", "H1", "h2"},
			expectedOrder: []string{"H1", "h2"},
			description:   "Should handle case-sensitive ID matching",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ApplyHighlightOrder(tt.segments, tt.order)

			var resultIDs []string
			if len(result) > 0 {
				for _, segment := range result {
					resultIDs = append(resultIDs, segment.ID)
				}
			}

			assert.Equal(t, tt.expectedOrder, resultIDs, tt.description)
		})
	}
}

func TestHighlightService_extractTextFromWordRange_InvalidRanges(t *testing.T) {
	service := &HighlightService{}
	words := []schema.Word{
		{Word: "Hello", Start: 0.0, End: 0.5},
		{Word: "world", Start: 0.5, End: 1.0},
		{Word: "test", Start: 1.0, End: 1.5},
	}

	tests := []struct {
		name         string
		startIndex   int
		endIndex     int
		expectedText string
		description  string
	}{
		{
			name:         "Start index equals end index",
			startIndex:   1,
			endIndex:     1,
			expectedText: "",
			description:  "Should return empty string when start equals end (exclusive end)",
		},
		{
			name:         "Start index greater than end index",
			startIndex:   2,
			endIndex:     0,
			expectedText: "",
			description:  "Should return empty string when start > end",
		},
		{
			name:         "Very large negative start index",
			startIndex:   -1000,
			endIndex:     1,
			expectedText: "",
			description:  "Should return empty string for very large negative start",
		},
		{
			name:         "Very large positive end index",
			startIndex:   0,
			endIndex:     1000,
			expectedText: "",
			description:  "Should return empty string for very large positive end",
		},
		{
			name:         "Start index at array length",
			startIndex:   3,
			endIndex:     3,
			expectedText: "",
			description:  "Should return empty string when start index equals array length",
		},
		{
			name:         "End index at array length",
			startIndex:   0,
			endIndex:     3,
			expectedText: "Hello world test",
			description:  "Should extract all words when end index equals array length (exclusive end)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.extractTextFromWordRange(words, tt.startIndex, tt.endIndex)
			assert.Equal(t, tt.expectedText, result, tt.description)
		})
	}
}

func TestHighlightService_StructCreation_EdgeCases(t *testing.T) {
	// Test struct creation with nil/empty values
	t.Run("Highlight with empty values", func(t *testing.T) {
		highlight := Highlight{
			ID:      "",
			Start:   0.0,
			End:     0.0,
			ColorID: 0,
		}
		assert.Equal(t, "", highlight.ID)
		assert.Equal(t, 0.0, highlight.Start)
		assert.Equal(t, 0.0, highlight.End)
		assert.Equal(t, 0, highlight.ColorID)
	})

	t.Run("HighlightWithText with empty text", func(t *testing.T) {
		highlight := HighlightWithText{
			ID:      "h1",
			Start:   1.0,
			End:     2.0,
			ColorID: 3,
			Text:    "",
		}
		assert.Equal(t, "", highlight.Text)
	})

	t.Run("ProjectHighlight with empty highlights", func(t *testing.T) {
		projectHighlight := ProjectHighlight{
			VideoClipID:   1,
			VideoClipName: "Test",
			FilePath:      "/test.mp4",
			Duration:      10.0,
			Highlights:    []HighlightWithText{},
		}
		assert.Empty(t, projectHighlight.Highlights)
		assert.NotNil(t, projectHighlight.Highlights)
	})

	t.Run("HighlightSegment with negative values", func(t *testing.T) {
		segment := HighlightSegment{
			ID:            "h1",
			VideoPath:     "/test.mp4",
			Start:         -1.0,
			End:           -0.5,
			ColorID:       3,
			Text:          "Test",
			VideoClipID:   -1,
			VideoClipName: "Test",
		}
		assert.Equal(t, -1.0, segment.Start)
		assert.Equal(t, -0.5, segment.End)
		assert.Equal(t, -1, segment.VideoClipID)
	})

	t.Run("HighlightSuggestion with negative indices", func(t *testing.T) {
		suggestion := HighlightSuggestion{
			ID:      "s1",
			Start:   -1,
			End:     -1,
			Text:    "Test",
			ColorID: 3,
		}
		assert.Equal(t, -1, suggestion.Start)
		assert.Equal(t, -1, suggestion.End)
	})
}

func TestHighlightService_NilContext(t *testing.T) {
	client := &ent.Client{}
	service := NewHighlightService(client, nil)

	assert.NotNil(t, service)
	assert.Equal(t, client, service.client)
	assert.Nil(t, service.ctx)
}

func TestHighlightService_NilClient(t *testing.T) {
	ctx := context.Background()
	service := NewHighlightService(nil, ctx)

	assert.NotNil(t, service)
	assert.Nil(t, service.client)
	assert.Equal(t, ctx, service.ctx)
}

func TestHighlightService_LargeDatasets(t *testing.T) {
	service := &HighlightService{}

	// Test with large number of words
	var largeWords []schema.Word
	for i := 0; i < 1000; i++ {
		largeWords = append(largeWords, schema.Word{
			Word:  fmt.Sprintf("word%d", i),
			Start: float64(i) * 0.1,
			End:   float64(i)*0.1 + 0.1,
		})
	}

	// Test time to word index with large dataset
	result := service.timeToWordIndex(50.0, largeWords)
	assert.Equal(t, 500, result)

	// Test word index to time with large dataset
	time := service.wordIndexToTime(500, largeWords)
	assert.Equal(t, 50.0, time)

	// Test extracting text from large range
	text := service.extractTextFromWordRange(largeWords, 0, 99)
	assert.NotEmpty(t, text)
	assert.Contains(t, text, "word0")
	assert.Contains(t, text, "word98") // End is exclusive, so 99 extracts up to word98

	// Test with large number of segments
	var largeSegments []HighlightSegment
	for i := 0; i < 1000; i++ {
		largeSegments = append(largeSegments, HighlightSegment{
			ID:            fmt.Sprintf("h%d", i),
			VideoClipName: fmt.Sprintf("Video%d", i),
		})
	}

	// Test applying order to large dataset
	order := []string{"h500", "h100", "h900"}
	orderedSegments := service.ApplyHighlightOrder(largeSegments, order)
	assert.Len(t, orderedSegments, 1000)
	assert.Equal(t, "h500", orderedSegments[0].ID)
	assert.Equal(t, "h100", orderedSegments[1].ID)
	assert.Equal(t, "h900", orderedSegments[2].ID)
}

func TestHighlightService_UnicodeAndSpecialCharacters(t *testing.T) {
	service := &HighlightService{}

	// Test with unicode characters
	unicodeWords := []schema.Word{
		{Word: "Hello", Start: 0.0, End: 0.5},
		{Word: "世界", Start: 0.5, End: 1.0},
		{Word: "🌍", Start: 1.0, End: 1.5},
		{Word: "café", Start: 1.5, End: 2.0},
		{Word: "naïve", Start: 2.0, End: 2.5},
	}

	text := service.extractTextFromWordRange(unicodeWords, 0, 5)
	assert.Equal(t, "Hello 世界 🌍 café naïve", text)

	// Test with very long words
	longWords := []schema.Word{
		{Word: "supercalifragilisticexpialidocious", Start: 0.0, End: 1.0},
		{Word: "pneumonoultramicroscopicsilicovolcanoconiosis", Start: 1.0, End: 2.0},
	}

	text = service.extractTextFromWordRange(longWords, 0, 2)
	assert.Contains(t, text, "supercalifragilisticexpialidocious")
	assert.Contains(t, text, "pneumonoultramicroscopicsilicovolcanoconiosis")

	// Test with special characters and whitespace
	specialWords := []schema.Word{
		{Word: "  ", Start: 0.0, End: 0.1},
		{Word: "\t", Start: 0.1, End: 0.2},
		{Word: "\n", Start: 0.2, End: 0.3},
		{Word: "normal", Start: 0.3, End: 0.4},
	}

	text = service.extractTextFromWordRange(specialWords, 0, 4)
	assert.Contains(t, text, "normal")
}
