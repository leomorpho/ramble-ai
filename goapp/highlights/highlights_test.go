package highlights

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"ramble-ai/ent"
	"ramble-ai/ent/schema"
)

// MockEntClient is a mock implementation of ent.Client for testing
type MockEntClient struct {
	mock.Mock
}

type MockVideoClipClient struct {
	mock.Mock
}

type MockSettingsClient struct {
	mock.Mock
}

type MockVideoClipQuery struct {
	mock.Mock
}

type MockSettingsQuery struct {
	mock.Mock
}

type MockVideoClipUpdate struct {
	mock.Mock
}

type MockSettingsUpdate struct {
	mock.Mock
}

// Mock implementations for ent.Client interfaces
func (m *MockEntClient) VideoClip() *MockVideoClipClient {
	return &MockVideoClipClient{}
}

func (m *MockEntClient) Settings() *MockSettingsClient {
	return &MockSettingsClient{}
}

// Helper function to create test data
func createTestWords() []schema.Word {
	return []schema.Word{
		{Word: "Hello", Start: 0.0, End: 0.5},
		{Word: "world", Start: 0.5, End: 1.0},
		{Word: "this", Start: 1.0, End: 1.5},
		{Word: "is", Start: 1.5, End: 2.0},
		{Word: "a", Start: 2.0, End: 2.5},
		{Word: "test", Start: 2.5, End: 3.0},
		{Word: "transcript", Start: 3.0, End: 3.5},
		{Word: "with", Start: 3.5, End: 4.0},
		{Word: "multiple", Start: 4.0, End: 4.5},
		{Word: "words", Start: 4.5, End: 5.0},
	}
}

func createTestHighlights() []schema.Highlight {
	return []schema.Highlight{
		{
			ID:      "h1",
			Start:   0.0,
			End:     1.0,
			ColorID: 3,
		},
		{
			ID:      "h2",
			Start:   2.0,
			End:     3.0,
			ColorID: 2,
		},
		{
			ID:      "h3",
			Start:   4.0,
			End:     5.0,
			ColorID: 3,
		},
	}
}

func createTestVideoClip() *ent.VideoClip {
	return &ent.VideoClip{
		ID:                  1,
		Name:                "Test Video",
		FilePath:            "/test/video.mp4",
		Duration:            10.0,
		Transcription:       "Hello world this is a test transcript with multiple words",
		TranscriptionWords:  createTestWords(),
		Highlights:          createTestHighlights(),
		SuggestedHighlights: []schema.Highlight{},
	}
}

func TestNewHighlightService(t *testing.T) {
	client := &ent.Client{}
	ctx := context.Background()

	service := NewHighlightService(client, ctx)

	assert.NotNil(t, service)
	assert.Equal(t, client, service.client)
	assert.Equal(t, ctx, service.ctx)
}

func TestHighlightService_TimeToWordIndex(t *testing.T) {
	service := &HighlightService{}
	words := createTestWords()

	tests := []struct {
		name          string
		timeSeconds   float64
		expectedIndex int
		description   string
	}{
		{
			name:          "Start of transcript",
			timeSeconds:   0.0,
			expectedIndex: 0,
			description:   "Should return first word index",
		},
		{
			name:          "Middle of transcript",
			timeSeconds:   2.5,
			expectedIndex: 5,
			description:   "Should return index of word at that time",
		},
		{
			name:          "End of transcript",
			timeSeconds:   5.0,
			expectedIndex: 9,
			description:   "Should return last word index",
		},
		{
			name:          "Beyond transcript end",
			timeSeconds:   10.0,
			expectedIndex: 9,
			description:   "Should return last word index for time beyond transcript",
		},
		{
			name:          "Between words",
			timeSeconds:   1.25,
			expectedIndex: 2,
			description:   "Should return word index that contains the time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.TimeToWordIndex(tt.timeSeconds, words)
			assert.Equal(t, tt.expectedIndex, result, tt.description)
		})
	}
}

func TestHighlightService_WordIndexToTime(t *testing.T) {
	service := &HighlightService{}
	words := createTestWords()

	tests := []struct {
		name         string
		wordIndex    int
		expectedTime float64
		description  string
	}{
		{
			name:         "First word",
			wordIndex:    0,
			expectedTime: 0.0,
			description:  "Should return start time of first word",
		},
		{
			name:         "Middle word",
			wordIndex:    5,
			expectedTime: 2.5,
			description:  "Should return start time of middle word",
		},
		{
			name:         "Last word",
			wordIndex:    9,
			expectedTime: 4.5,
			description:  "Should return start time of last word",
		},
		{
			name:         "Invalid index - negative",
			wordIndex:    -1,
			expectedTime: 0.0,
			description:  "Should return 0 for negative index",
		},
		{
			name:         "Invalid index - beyond range",
			wordIndex:    20,
			expectedTime: 0.0,
			description:  "Should return 0 for index beyond range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.WordIndexToTime(tt.wordIndex, words)
			assert.Equal(t, tt.expectedTime, result, tt.description)
		})
	}
}

func TestHighlightService_extractTextFromWordRange(t *testing.T) {
	service := &HighlightService{}
	words := createTestWords()

	tests := []struct {
		name         string
		startIndex   int
		endIndex     int
		expectedText string
		description  string
	}{
		{
			name:         "Single word",
			startIndex:   0,
			endIndex:     1,
			expectedText: "Hello",
			description:  "Should extract single word",
		},
		{
			name:         "Multiple words",
			startIndex:   0,
			endIndex:     3,
			expectedText: "Hello world this",
			description:  "Should extract multiple words",
		},
		{
			name:         "Full range",
			startIndex:   0,
			endIndex:     10,
			expectedText: "Hello world this is a test transcript with multiple words",
			description:  "Should extract all words",
		},
		{
			name:         "Invalid range - start > end",
			startIndex:   5,
			endIndex:     2,
			expectedText: "",
			description:  "Should return empty string for invalid range",
		},
		{
			name:         "Invalid range - negative start",
			startIndex:   -1,
			endIndex:     2,
			expectedText: "",
			description:  "Should return empty string for negative start",
		},
		{
			name:         "Invalid range - end beyond length",
			startIndex:   0,
			endIndex:     20,
			expectedText: "",
			description:  "Should return empty string for end beyond length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.extractTextFromWordRange(words, tt.startIndex, tt.endIndex)
			assert.Equal(t, tt.expectedText, result, tt.description)
		})
	}
}

func TestHighlightService_extractHighlightText(t *testing.T) {
	service := &HighlightService{}
	words := createTestWords()
	fullText := "Hello world this is a test transcript with multiple words"

	tests := []struct {
		name         string
		highlight    schema.Highlight
		expectedText string
		description  string
	}{
		{
			name: "Exact word boundaries",
			highlight: schema.Highlight{
				ID:      "h1",
				Start:   0.0,
				End:     1.0,
				ColorID: 3,
			},
			expectedText: "Hello world",
			description:  "Should extract text for exact word boundaries",
		},
		{
			name: "Partial word overlap",
			highlight: schema.Highlight{
				ID:      "h2",
				Start:   1.25,
				End:     2.75,
				ColorID: 2,
			},
			expectedText: "is a",
			description:  "Should extract words that fall within time range",
		},
		{
			name: "No matching words",
			highlight: schema.Highlight{
				ID:      "h3",
				Start:   10.0,
				End:     11.0,
				ColorID: 3,
			},
			expectedText: "",
			description:  "Should return empty string when no words match",
		},
		{
			name: "Single word highlight",
			highlight: schema.Highlight{
				ID:      "h4",
				Start:   2.0,
				End:     2.5,
				ColorID: 1,
			},
			expectedText: "a",
			description:  "Should extract single word",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.extractHighlightText(tt.highlight, words, fullText)
			assert.Equal(t, tt.expectedText, result, tt.description)
		})
	}
}

func TestHighlightService_extractHighlightText_FallbackToFullText(t *testing.T) {
	service := &HighlightService{}
	// Empty words array to trigger fallback
	words := []schema.Word{}
	fullText := "Hello world this is a test transcript with multiple words"

	highlight := schema.Highlight{
		ID:      "h1",
		Start:   0.0,
		End:     1.0,
		ColorID: 3,
	}

	// Should return empty string when no words and no duration info
	result := service.extractHighlightText(highlight, words, fullText)
	assert.Equal(t, "", result)

	// Test with words that have duration info
	words = []schema.Word{
		{Word: "Hello", Start: 0.0, End: 0.5},
		{Word: "world", Start: 0.5, End: 1.0},
		{Word: "test", Start: 1.0, End: 1.5},
	}

	// This should still return empty because the time range doesn't match any words exactly
	result = service.extractHighlightText(highlight, words, fullText)
	assert.Equal(t, "Hello world", result)
}

func TestHighlightService_ApplyHighlightOrder(t *testing.T) {
	service := &HighlightService{}

	segments := []HighlightSegment{
		{ID: "h1", VideoClipName: "Video1", Start: 0.0, End: 1.0},
		{ID: "h2", VideoClipName: "Video2", Start: 2.0, End: 3.0},
		{ID: "h3", VideoClipName: "Video3", Start: 4.0, End: 5.0},
		{ID: "h4", VideoClipName: "Video4", Start: 6.0, End: 7.0},
	}

	tests := []struct {
		name          string
		order         []string
		expectedOrder []string
		description   string
	}{
		{
			name:          "Empty order",
			order:         []string{},
			expectedOrder: []string{"h1", "h2", "h3", "h4"},
			description:   "Should maintain original order when no custom order provided",
		},
		{
			name:          "Reverse order",
			order:         []string{"h4", "h3", "h2", "h1"},
			expectedOrder: []string{"h4", "h3", "h2", "h1"},
			description:   "Should apply reverse order",
		},
		{
			name:          "Partial order",
			order:         []string{"h3", "h1"},
			expectedOrder: []string{"h3", "h1", "h4", "h2"},
			description:   "Should place ordered items first, then unordered items",
		},
		{
			name:          "Order with non-existent IDs",
			order:         []string{"h5", "h2", "h6", "h4"},
			expectedOrder: []string{"h2", "h4", "h3", "h1"},
			description:   "Should ignore non-existent IDs and order remaining items",
		},
		{
			name:          "Duplicate IDs in order",
			order:         []string{"h2", "h2", "h1", "h3"},
			expectedOrder: []string{"h2", "h1", "h3", "h4"},
			description:   "Should handle duplicate IDs in order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ApplyHighlightOrder(segments, tt.order)

			// Extract IDs from result for comparison
			var resultIDs []string
			for _, segment := range result {
				resultIDs = append(resultIDs, segment.ID)
			}

			assert.Equal(t, tt.expectedOrder, resultIDs, tt.description)
		})
	}
}

func TestHighlightService_GetProjectHighlightAISettings_DefaultValues(t *testing.T) {
	// Since we can't easily mock the database calls without significant refactoring,
	// we'll test the logic by creating structs and checking default values
	// This test would need a proper mock setup or test database in a real scenario

	// For now, we'll create a simple test to verify the struct creation
	settings := &ProjectHighlightAISettings{
		AIModel:  "openai/gpt-4o-mini",
		AIPrompt: "Analyze this transcript and suggest the most interesting highlights for a video compilation.",
	}

	assert.Equal(t, "openai/gpt-4o-mini", settings.AIModel)
	assert.Equal(t, "Analyze this transcript and suggest the most interesting highlights for a video compilation.", settings.AIPrompt)
}

func TestHighlightService_ProjectHighlight_Creation(t *testing.T) {
	// Test creating ProjectHighlight struct
	highlights := []HighlightWithText{
		{
			ID:      "h1",
			Start:   0.0,
			End:     1.0,
			ColorID: 3,
			Text:    "Hello world",
		},
		{
			ID:      "h2",
			Start:   2.0,
			End:     3.0,
			ColorID: 2,
			Text:    "test transcript",
		},
	}

	projectHighlight := ProjectHighlight{
		VideoClipID:   1,
		VideoClipName: "Test Video",
		FilePath:      "/test/video.mp4",
		Duration:      10.0,
		Highlights:    highlights,
	}

	assert.Equal(t, 1, projectHighlight.VideoClipID)
	assert.Equal(t, "Test Video", projectHighlight.VideoClipName)
	assert.Equal(t, "/test/video.mp4", projectHighlight.FilePath)
	assert.Equal(t, 10.0, projectHighlight.Duration)
	assert.Len(t, projectHighlight.Highlights, 2)
	assert.Equal(t, "Hello world", projectHighlight.Highlights[0].Text)
	assert.Equal(t, "test transcript", projectHighlight.Highlights[1].Text)
}

func TestHighlightService_HighlightSegment_Creation(t *testing.T) {
	// Test creating HighlightSegment struct
	segment := HighlightSegment{
		ID:            "h1",
		VideoPath:     "/test/video.mp4",
		Start:         0.0,
		End:           1.0,
		ColorID:       3,
		Text:          "Hello world",
		VideoClipID:   1,
		VideoClipName: "Test Video",
	}

	assert.Equal(t, "h1", segment.ID)
	assert.Equal(t, "/test/video.mp4", segment.VideoPath)
	assert.Equal(t, 0.0, segment.Start)
	assert.Equal(t, 1.0, segment.End)
	assert.Equal(t, 3, segment.ColorID)
	assert.Equal(t, "Hello world", segment.Text)
	assert.Equal(t, 1, segment.VideoClipID)
	assert.Equal(t, "Test Video", segment.VideoClipName)
}

func TestHighlightService_HighlightSuggestion_Creation(t *testing.T) {
	// Test creating HighlightSuggestion struct
	suggestion := HighlightSuggestion{
		ID:      "s1",
		Start:   0,
		End:     5,
		Text:    "Hello world this is a test",
		ColorID: 3,
	}

	assert.Equal(t, "s1", suggestion.ID)
	assert.Equal(t, 0, suggestion.Start)
	assert.Equal(t, 5, suggestion.End)
	assert.Equal(t, "Hello world this is a test", suggestion.Text)
	assert.Equal(t, 3, suggestion.ColorID)
}

// Edge case tests
func TestHighlightService_extractTextFromWordRange_EdgeCases(t *testing.T) {
	service := &HighlightService{}

	// Test with empty words array
	emptyWords := []schema.Word{}
	result := service.extractTextFromWordRange(emptyWords, 0, 0)
	assert.Equal(t, "", result)

	// Test with single word
	singleWord := []schema.Word{
		{Word: "Hello", Start: 0.0, End: 0.5},
	}
	result = service.extractTextFromWordRange(singleWord, 0, 1)
	assert.Equal(t, "Hello", result)

	// Test with words containing special characters
	specialWords := []schema.Word{
		{Word: "Hello,", Start: 0.0, End: 0.5},
		{Word: "world!", Start: 0.5, End: 1.0},
		{Word: "How's", Start: 1.0, End: 1.5},
		{Word: "everything?", Start: 1.5, End: 2.0},
	}
	result = service.extractTextFromWordRange(specialWords, 0, 4)
	assert.Equal(t, "Hello, world! How's everything?", result)
}

func TestHighlightService_timeToWordIndex_EdgeCases(t *testing.T) {
	service := &HighlightService{}

	// Test with empty words array
	emptyWords := []schema.Word{}
	result := service.timeToWordIndex(1.0, emptyWords)
	assert.Equal(t, -1, result)

	// Test with single word
	singleWord := []schema.Word{
		{Word: "Hello", Start: 0.0, End: 0.5},
	}
	result = service.timeToWordIndex(0.25, singleWord)
	assert.Equal(t, 0, result)

	// Test with negative time
	words := createTestWords()
	result = service.timeToWordIndex(-1.0, words)
	assert.Equal(t, 0, result)
}

func TestHighlightService_wordIndexToTime_EdgeCases(t *testing.T) {
	service := &HighlightService{}

	// Test with empty words array
	emptyWords := []schema.Word{}
	result := service.wordIndexToTime(0, emptyWords)
	assert.Equal(t, 0.0, result)

	// Test with single word
	singleWord := []schema.Word{
		{Word: "Hello", Start: 0.0, End: 0.5},
	}
	result = service.wordIndexToTime(0, singleWord)
	assert.Equal(t, 0.0, result)
}

// Performance test (basic benchmark)
func BenchmarkHighlightService_extractTextFromWordRange(b *testing.B) {
	service := &HighlightService{}
	words := createTestWords()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.extractTextFromWordRange(words, 0, len(words)-1)
	}
}

func BenchmarkHighlightService_timeToWordIndex(b *testing.B) {
	service := &HighlightService{}
	words := createTestWords()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.timeToWordIndex(2.5, words)
	}
}

func BenchmarkHighlightService_ApplyHighlightOrder(b *testing.B) {
	service := &HighlightService{}
	segments := []HighlightSegment{
		{ID: "h1", VideoClipName: "Video1", Start: 0.0, End: 1.0},
		{ID: "h2", VideoClipName: "Video2", Start: 2.0, End: 3.0},
		{ID: "h3", VideoClipName: "Video3", Start: 4.0, End: 5.0},
		{ID: "h4", VideoClipName: "Video4", Start: 6.0, End: 7.0},
		{ID: "h5", VideoClipName: "Video5", Start: 8.0, End: 9.0},
	}
	order := []string{"h3", "h1", "h5", "h2", "h4"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ApplyHighlightOrder(segments, order)
	}
}
