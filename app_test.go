package main

import (
	"testing"

	"MYAPP/ent/schema"
)

func TestFilterValidHighlightSuggestions(t *testing.T) {
	app := &App{}

	// Create test transcript words
	transcriptWords := []schema.Word{
		{Word: "Hello", Start: 0.0, End: 0.5},
		{Word: "world", Start: 0.5, End: 1.0},
		{Word: "this", Start: 1.0, End: 1.3},
		{Word: "is", Start: 1.3, End: 1.5},
		{Word: "a", Start: 1.5, End: 1.6},
		{Word: "test", Start: 1.6, End: 2.0},
		{Word: "of", Start: 2.0, End: 2.2},
		{Word: "the", Start: 2.2, End: 2.4},
		{Word: "highlight", Start: 2.4, End: 3.0},
		{Word: "system", Start: 3.0, End: 3.5},
	}

	tests := []struct {
		name               string
		suggestions        []HighlightSuggestion
		existingHighlights []schema.Highlight
		expectedCount      int
		description        string
	}{
		{
			name: "No overlap - all suggestions should be kept",
			suggestions: []HighlightSuggestion{
				{ID: "s1", Start: 0, End: 1, Text: "Hello world"},    // 0.0-1.0
				{ID: "s2", Start: 4, End: 5, Text: "a test"},        // 1.5-2.0
			},
			existingHighlights: []schema.Highlight{
				{ID: "h1", Start: 2.0, End: 3.0}, // "of the highlight"
			},
			expectedCount: 2,
			description: "Suggestions don't overlap with existing highlight",
		},
		{
			name: "Complete overlap - suggestion encompasses existing highlight",
			suggestions: []HighlightSuggestion{
				{ID: "s1", Start: 0, End: 9, Text: "Hello world this is a test of the highlight system"}, // 0.0-3.5
			},
			existingHighlights: []schema.Highlight{
				{ID: "h1", Start: 1.0, End: 2.0}, // "this is a test"
			},
			expectedCount: 0,
			description: "Suggestion completely encompasses existing highlight",
		},
		{
			name: "Partial overlap at start",
			suggestions: []HighlightSuggestion{
				{ID: "s1", Start: 0, End: 3, Text: "Hello world this is"}, // 0.0-1.5
			},
			existingHighlights: []schema.Highlight{
				{ID: "h1", Start: 1.0, End: 2.0}, // "this is a test"
			},
			expectedCount: 0,
			description: "Suggestion overlaps with start of existing highlight",
		},
		{
			name: "Partial overlap at end",
			suggestions: []HighlightSuggestion{
				{ID: "s1", Start: 4, End: 9, Text: "a test of the highlight system"}, // 1.5-3.5
			},
			existingHighlights: []schema.Highlight{
				{ID: "h1", Start: 1.0, End: 2.0}, // "this is a test"
			},
			expectedCount: 0,
			description: "Suggestion overlaps with end of existing highlight",
		},
		{
			name: "Existing highlight encompasses suggestion",
			suggestions: []HighlightSuggestion{
				{ID: "s1", Start: 3, End: 4, Text: "is a"}, // 1.3-1.6
			},
			existingHighlights: []schema.Highlight{
				{ID: "h1", Start: 1.0, End: 2.0}, // "this is a test"
			},
			expectedCount: 0,
			description: "Existing highlight completely encompasses suggestion",
		},
		{
			name: "Multiple suggestions with some overlapping",
			suggestions: []HighlightSuggestion{
				{ID: "s1", Start: 0, End: 1, Text: "Hello world"},          // 0.0-1.0 (no overlap)
				{ID: "s2", Start: 2, End: 5, Text: "this is a test"},      // 1.0-2.0 (overlaps)
				{ID: "s3", Start: 6, End: 9, Text: "of the highlight"},    // 2.0-3.5 (no overlap)
			},
			existingHighlights: []schema.Highlight{
				{ID: "h1", Start: 1.2, End: 1.8}, // middle of "this is a test"
			},
			expectedCount: 2,
			description: "Only non-overlapping suggestions should be kept",
		},
		{
			name: "Suggestions overlapping with each other",
			suggestions: []HighlightSuggestion{
				{ID: "s1", Start: 0, End: 3, Text: "Hello world this is"},     // 0.0-1.5
				{ID: "s2", Start: 1, End: 5, Text: "world this is a test"},   // 0.5-2.0 (overlaps with s1)
				{ID: "s3", Start: 6, End: 9, Text: "of the highlight"},       // 2.0-3.5 (no overlap)
			},
			existingHighlights: []schema.Highlight{},
			expectedCount: 2, // s1 and s3 (s2 overlaps with s1)
			description: "Suggestions that overlap with each other should be filtered",
		},
		{
			name: "Edge case - exact boundary touch",
			suggestions: []HighlightSuggestion{
				{ID: "s1", Start: 0, End: 2, Text: "Hello world this"}, // 0.0-1.3
				{ID: "s2", Start: 5, End: 9, Text: "test of the highlight system"}, // 1.6-3.5
			},
			existingHighlights: []schema.Highlight{
				{ID: "h1", Start: 1.3, End: 1.6}, // "is a" - exactly touches s1 end and s2 start
			},
			expectedCount: 2,
			description: "Suggestions that exactly touch boundaries should be kept",
		},
		{
			name: "Real overlap scenario - suggestion contains existing",
			suggestions: []HighlightSuggestion{
				{ID: "s1", Start: 0, End: 5, Text: "Hello world this is a test"}, // 0.0-2.0 (encompasses existing)
			},
			existingHighlights: []schema.Highlight{
				{ID: "h1", Start: 1.0, End: 1.5}, // "this is" - inside suggestion
			},
			expectedCount: 0,
			description: "Suggestion that encompasses existing highlight should be dropped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.filterValidHighlightSuggestions(tt.suggestions, tt.existingHighlights, transcriptWords)
			
			if len(result) != tt.expectedCount {
				t.Errorf("%s: expected %d suggestions, got %d. %s", 
					tt.name, tt.expectedCount, len(result), tt.description)
				
				// Print details for debugging
				t.Logf("Existing highlights:")
				for _, h := range tt.existingHighlights {
					t.Logf("  %s: %.2f-%.2f", h.ID, h.Start, h.End)
				}
				
				t.Logf("Input suggestions:")
				for _, s := range tt.suggestions {
					start := app.wordIndexToTime(s.Start, transcriptWords)
					end := transcriptWords[s.End].End
					t.Logf("  %s: [%d-%d] %.2f-%.2f '%s'", s.ID, s.Start, s.End, start, end, s.Text)
				}
				
				t.Logf("Filtered suggestions:")
				for _, s := range result {
					start := app.wordIndexToTime(s.Start, transcriptWords)
					end := transcriptWords[s.End].End
					t.Logf("  %s: [%d-%d] %.2f-%.2f '%s'", s.ID, s.Start, s.End, start, end, s.Text)
				}
			}
		})
	}
}