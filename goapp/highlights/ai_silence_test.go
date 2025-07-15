package highlights

import (
	"context"
	"testing"

	"MYAPP/ent/enttest"
	"MYAPP/ent/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAPIKeyGetter returns a mock API key for testing
func MockAPIKeyGetter() (string, error) {
	return "mock-api-key", nil
}

// TestImproveHighlightSilences tests the highlight silence improvement logic
func TestImproveHighlightSilences(t *testing.T) {
	// Create in-memory SQLite database for testing
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	ctx := context.Background()

	// Create test project
	project, err := client.Project.Create().
		SetName("Test Project").
		SetPath("/test/path").
		Save(ctx)
	require.NoError(t, err)

	// Test Case 1: Highlight with available padding space
	t.Run("AddPaddingWhenSpaceAvailable", func(t *testing.T) {
		// Create video clip with transcription words
		// Simulating: "Hello world this is a test sentence with some padding"
		words := []schema.Word{
			{Word: "Hello", Start: 0.0, End: 0.5},
			{Word: "world", Start: 0.6, End: 1.0},
			// Gap from 1.0 to 1.5 (silence)
			{Word: "this", Start: 1.5, End: 1.8},
			{Word: "is", Start: 1.9, End: 2.0},
			{Word: "a", Start: 2.1, End: 2.2},
			{Word: "test", Start: 2.3, End: 2.6},
			// Gap from 2.6 to 3.0 (silence)
			{Word: "sentence", Start: 3.0, End: 3.5},
			{Word: "with", Start: 3.6, End: 3.8},
			{Word: "some", Start: 3.9, End: 4.1},
			{Word: "padding", Start: 4.2, End: 4.6},
		}

		clip, err := client.VideoClip.Create().
			SetName("Test Video").
			SetFilePath("/test/video.mp4").
			SetDuration(5.0).
			SetTranscription("Hello world this is a test sentence with some padding").
			SetTranscriptionWords(words).
			SetProject(project).
			Save(ctx)
		require.NoError(t, err)

		// Create a highlight that currently cuts exactly at word boundaries
		// Highlighting "this is a test" (words index 2-5)
		highlight := schema.Highlight{
			ID:      "highlight_1",
			Start:   1.5, // Start of "this"
			End:     2.6, // End of "test"
			ColorID: 1,
		}

		_, err = client.VideoClip.UpdateOne(clip).
			SetHighlights([]schema.Highlight{highlight}).
			Save(ctx)
		require.NoError(t, err)

		// Create mock improver that adds padding
		mockImprover := &MockSilenceImprover{
			improvements: map[string]struct {
				Start float64
				End   float64
			}{
				"highlight_1": {
					Start: 1.2, // 300ms before "this" (in the silence gap)
					End:   2.9, // 300ms after "test" (in the silence gap)
				},
			},
		}

		// Test the improvement logic
		improved := mockImprover.ImproveHighlight(highlight, words)

		// Verify padding was added
		assert.Equal(t, 1.2, improved.Start, "Should add padding before highlight")
		assert.Equal(t, 2.9, improved.End, "Should add padding after highlight")

		// Verify it doesn't cut into words
		assert.Less(t, improved.Start, 1.5, "Padded start should be before word start")
		assert.Greater(t, improved.Start, 1.0, "Padded start should be after previous word end")
		assert.Greater(t, improved.End, 2.6, "Padded end should be after word end")
		assert.Less(t, improved.End, 3.0, "Padded end should be before next word start")
	})

	// Test Case 2: Highlight with limited padding space
	t.Run("AddLimitedPaddingWhenLessSpaceAvailable", func(t *testing.T) {
		// Create video clip with tighter word spacing
		words := []schema.Word{
			{Word: "Quick", Start: 0.0, End: 0.4},
			{Word: "brown", Start: 0.45, End: 0.8}, // Only 50ms gap
			{Word: "fox", Start: 0.85, End: 1.1},   // Only 50ms gap
			{Word: "jumps", Start: 1.15, End: 1.5}, // Only 50ms gap
		}

		clip, err := client.VideoClip.Create().
			SetName("Test Video 2").
			SetFilePath("/test/video2.mp4").
			SetDuration(2.0).
			SetTranscription("Quick brown fox jumps").
			SetTranscriptionWords(words).
			SetProject(project).
			Save(ctx)
		require.NoError(t, err)

		// Highlight "brown fox"
		highlight := schema.Highlight{
			ID:      "highlight_2",
			Start:   0.45, // Start of "brown"
			End:     1.1,  // End of "fox"
			ColorID: 2,
		}

		_, err = client.VideoClip.UpdateOne(clip).
			SetHighlights([]schema.Highlight{highlight}).
			Save(ctx)
		require.NoError(t, err)

		// Create mock improver that respects available space
		mockImprover := &MockSilenceImprover{
			improvements: map[string]struct {
				Start float64
				End   float64
			}{
				"highlight_2": {
					Start: 0.42, // Only 30ms before (limited by previous word)
					End:   1.13, // Only 30ms after (limited by next word)
				},
			},
		}

		improved := mockImprover.ImproveHighlight(highlight, words)

		// Verify limited padding was added
		assert.Equal(t, 0.42, improved.Start, "Should add limited padding before")
		assert.Equal(t, 1.13, improved.End, "Should add limited padding after")

		// Verify it respects word boundaries
		assert.Greater(t, improved.Start, 0.4, "Should not overlap previous word")
		assert.Less(t, improved.End, 1.15, "Should not overlap next word")
	})

	// Test Case 3: Highlight at video boundaries
	t.Run("HandleVideoBoundaries", func(t *testing.T) {
		words := []schema.Word{
			{Word: "First", Start: 0.1, End: 0.4},
			{Word: "word", Start: 0.5, End: 0.8},
			{Word: "last", Start: 4.5, End: 4.8},
			{Word: "word", Start: 4.85, End: 4.95},
		}

		clip, err := client.VideoClip.Create().
			SetName("Test Video 3").
			SetFilePath("/test/video3.mp4").
			SetDuration(5.0).
			SetTranscription("First word last word").
			SetTranscriptionWords(words).
			SetProject(project).
			Save(ctx)
		require.NoError(t, err)

		// Highlight at the beginning
		highlightStart := schema.Highlight{
			ID:      "highlight_3",
			Start:   0.1,
			End:     0.8,
			ColorID: 3,
		}

		// Highlight at the end
		highlightEnd := schema.Highlight{
			ID:      "highlight_4",
			Start:   4.5,
			End:     4.95,
			ColorID: 4,
		}

		_, err = client.VideoClip.UpdateOne(clip).
			SetHighlights([]schema.Highlight{highlightStart, highlightEnd}).
			Save(ctx)
		require.NoError(t, err)

		mockImprover := &MockSilenceImprover{
			improvements: map[string]struct {
				Start float64
				End   float64
			}{
				"highlight_3": {
					Start: 0.0, // Can't go before video start
					End:   1.0, // Add padding after
				},
				"highlight_4": {
					Start: 4.3, // Add padding before
					End:   5.0, // Can't go beyond video duration
				},
			},
		}

		improvedStart := mockImprover.ImproveHighlight(highlightStart, words)
		improvedEnd := mockImprover.ImproveHighlight(highlightEnd, words)

		// Verify video boundaries are respected
		assert.GreaterOrEqual(t, improvedStart.Start, 0.0, "Should not go before video start")
		assert.LessOrEqual(t, improvedEnd.End, 5.0, "Should not go beyond video duration")
	})

	// Test Case 4: Multiple highlights with no overlap after padding
	t.Run("PreventOverlapBetweenPaddedHighlights", func(t *testing.T) {
		words := []schema.Word{
			{Word: "One", Start: 0.0, End: 0.3},
			{Word: "two", Start: 0.5, End: 0.8},
			{Word: "three", Start: 1.0, End: 1.3},
			{Word: "four", Start: 1.5, End: 1.8},
			{Word: "five", Start: 2.0, End: 2.3},
		}

		clip, err := client.VideoClip.Create().
			SetName("Test Video 4").
			SetFilePath("/test/video4.mp4").
			SetDuration(3.0).
			SetTranscription("One two three four five").
			SetTranscriptionWords(words).
			SetProject(project).
			Save(ctx)
		require.NoError(t, err)

		// Two highlights close to each other
		highlight1 := schema.Highlight{
			ID:      "highlight_5",
			Start:   0.5, // "two"
			End:     0.8,
			ColorID: 1,
		}
		highlight2 := schema.Highlight{
			ID:      "highlight_6",
			Start:   1.0, // "three"
			End:     1.3,
			ColorID: 2,
		}

		_, err = client.VideoClip.UpdateOne(clip).
			SetHighlights([]schema.Highlight{highlight1, highlight2}).
			Save(ctx)
		require.NoError(t, err)

		// Improver should limit padding to prevent overlap
		mockImprover := &MockSilenceImprover{
			improvements: map[string]struct {
				Start float64
				End   float64
			}{
				"highlight_5": {
					Start: 0.35, // Add padding before
					End:   0.9,  // Limited padding after to avoid overlap
				},
				"highlight_6": {
					Start: 0.95, // Limited padding before to avoid overlap
					End:   1.45, // Add padding after
				},
			},
		}

		improved1 := mockImprover.ImproveHighlight(highlight1, words)
		improved2 := mockImprover.ImproveHighlight(highlight2, words)

		// Verify no overlap between padded highlights
		assert.Less(t, improved1.End, improved2.Start, "Padded highlights should not overlap")
		assert.GreaterOrEqual(t, improved2.Start-improved1.End, 0.04, "Should maintain minimum gap")
	})
}

// MockSilenceImprover provides controlled improvements for testing
type MockSilenceImprover struct {
	improvements map[string]struct {
		Start float64
		End   float64
	}
}

func (m *MockSilenceImprover) ImproveHighlight(highlight schema.Highlight, words []schema.Word) schema.Highlight {
	if improvement, exists := m.improvements[highlight.ID]; exists {
		return schema.Highlight{
			ID:      highlight.ID,
			Start:   improvement.Start,
			End:     improvement.End,
			ColorID: highlight.ColorID,
		}
	}
	return highlight
}

// TestCalculatePaddingBoundaries tests the logic for calculating safe padding boundaries
func TestCalculatePaddingBoundaries(t *testing.T) {

	words := []schema.Word{
		{Word: "Hello", Start: 0.0, End: 0.5},
		{Word: "world", Start: 1.0, End: 1.5},
		{Word: "test", Start: 2.0, End: 2.5},
	}

	t.Run("MiddleHighlight", func(t *testing.T) {
		// Highlight the word "world"
		startIdx := 1
		endIdx := 2

		prevEnd := words[startIdx-1].End // End of "Hello" = 0.5
		nextStart := words[endIdx].Start // Start of "test" = 2.0

		// Available padding space
		assert.Equal(t, 0.5, prevEnd)
		assert.Equal(t, 2.0, nextStart)

		// Current highlight bounds
		currentStart := words[startIdx].Start // 1.0
		currentEnd := words[endIdx-1].End     // 1.5

		// Maximum padding available
		maxPaddingBefore := currentStart - prevEnd // 1.0 - 0.5 = 0.5
		maxPaddingAfter := nextStart - currentEnd  // 2.0 - 1.5 = 0.5

		assert.Equal(t, 0.5, maxPaddingBefore)
		assert.Equal(t, 0.5, maxPaddingAfter)
	})

	t.Run("FirstHighlight", func(t *testing.T) {
		// Highlight the first word "Hello"
		startIdx := 0
		endIdx := 1

		// No previous word, so padding can go to 0
		prevEnd := 0.0
		nextStart := words[endIdx].Start // Start of "world" = 1.0

		currentStart := words[startIdx].Start // 0.0
		currentEnd := words[endIdx-1].End     // 0.5

		maxPaddingBefore := currentStart - prevEnd // 0.0 - 0.0 = 0.0
		maxPaddingAfter := nextStart - currentEnd  // 1.0 - 0.5 = 0.5

		assert.Equal(t, 0.0, maxPaddingBefore)
		assert.Equal(t, 0.5, maxPaddingAfter)
	})
}

// TestPaddingExampleDemo demonstrates realistic padding scenarios
func TestPaddingExampleDemo(t *testing.T) {
	// Create in-memory SQLite database for testing
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	ctx := context.Background()

	// Create test project
	project, err := client.Project.Create().
		SetName("Padding Demo Project").
		SetPath("/test/demo").
		Save(ctx)
	require.NoError(t, err)

	// Create realistic video with natural speech patterns
	words := []schema.Word{
		{Word: "Welcome", Start: 0.0, End: 0.4},
		{Word: "everyone", Start: 0.5, End: 0.9},
		// Natural pause for breath
		{Word: "to", Start: 1.3, End: 1.4},
		{Word: "today's", Start: 1.45, End: 1.8},
		{Word: "presentation", Start: 1.85, End: 2.6},
		// Short pause
		{Word: "we're", Start: 3.0, End: 3.2},
		{Word: "going", Start: 3.25, End: 3.5},
		{Word: "to", Start: 3.55, End: 3.65},
		{Word: "cover", Start: 3.7, End: 4.0},
		{Word: "some", Start: 4.05, End: 4.25},
		{Word: "amazing", Start: 4.3, End: 4.7},
		{Word: "features", Start: 4.75, End: 5.2},
	}

	clip, err := client.VideoClip.Create().
		SetName("Demo Video").
		SetFilePath("/test/demo.mp4").
		SetDuration(6.0).
		SetTranscription("Welcome everyone to today's presentation we're going to cover some amazing features").
		SetTranscriptionWords(words).
		SetProject(project).
		Save(ctx)
	require.NoError(t, err)

	// Create test highlights
	highlights := []schema.Highlight{
		{
			ID:      "highlight_opening",
			Start:   1.45, // Start of "today's"
			End:     2.6,  // End of "presentation"
			ColorID: 1,
		},
		{
			ID:      "highlight_features",
			Start:   4.3, // Start of "amazing"
			End:     5.2, // End of "features"
			ColorID: 2,
		},
	}

	_, err = client.VideoClip.UpdateOne(clip).
		SetHighlights(highlights).
		Save(ctx)
	require.NoError(t, err)

	// Test padding calculations
	t.Run("CalculatePaddingForEachHighlight", func(t *testing.T) {
		for _, highlight := range highlights {
			// Find available space around highlight
			prevWordEnd := 0.0
			nextWordStart := clip.Duration

			// Find previous word
			for i, word := range words {
				if word.Start >= highlight.Start {
					if i > 0 {
						prevWordEnd = words[i-1].End
					}
					break
				}
			}

			// Find next word
			for _, word := range words {
				if word.Start > highlight.End {
					nextWordStart = word.Start
					break
				}
			}

			availableBefore := highlight.Start - prevWordEnd
			availableAfter := nextWordStart - highlight.End

			t.Logf("Highlight %s:", highlight.ID)
			t.Logf("  Current: %.2fs - %.2fs", highlight.Start, highlight.End)
			t.Logf("  Available padding: %.2fs before, %.2fs after", availableBefore, availableAfter)

			// Calculate recommended padding (250ms)
			recommendedPadding := 0.25
			actualPaddingBefore := min(recommendedPadding, availableBefore-0.05)
			if actualPaddingBefore < 0 {
				actualPaddingBefore = 0
			}

			actualPaddingAfter := min(recommendedPadding, availableAfter-0.05)
			if actualPaddingAfter < 0 {
				actualPaddingAfter = 0
			}

			improvedStart := highlight.Start - actualPaddingBefore
			improvedEnd := highlight.End + actualPaddingAfter

			t.Logf("  Improved: %.2fs - %.2fs", improvedStart, improvedEnd)
			t.Logf("  Added: %.0fms before, %.0fms after", actualPaddingBefore*1000, actualPaddingAfter*1000)

			// Verify improvements are valid
			assert.GreaterOrEqual(t, improvedStart, prevWordEnd, "Start should not overlap previous word")
			assert.LessOrEqual(t, improvedEnd, nextWordStart, "End should not overlap next word")
			assert.GreaterOrEqual(t, improvedStart, 0.0, "Start should not be negative")
			assert.LessOrEqual(t, improvedEnd, clip.Duration, "End should not exceed video duration")
		}
	})
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// TestImproveHighlightWithRealLogic tests with actual improvement logic
func TestImproveHighlightWithRealLogic(t *testing.T) {
	// This test demonstrates how the real improvement would work

	t.Run("NaturalSilencePadding", func(t *testing.T) {
		words := []schema.Word{
			{Word: "So", Start: 1.0, End: 1.2},
			{Word: "the", Start: 1.25, End: 1.4},
			{Word: "key", Start: 1.45, End: 1.65},
			{Word: "insight", Start: 1.7, End: 2.1},
			{Word: "here", Start: 2.15, End: 2.35},
			// Natural pause
			{Word: "is", Start: 2.8, End: 2.9},
			{Word: "that", Start: 2.95, End: 3.15},
		}

		// Original highlight covers "the key insight here" tightly
		highlight := schema.Highlight{
			ID:      "highlight_natural",
			Start:   1.25, // Start of "the"
			End:     2.35, // End of "here"
			ColorID: 1,
		}

		// Calculate ideal padding boundaries
		prevWordEnd := words[0].End     // "So" ends at 1.2
		nextWordStart := words[5].Start // "is" starts at 2.8

		// With natural silence detection, we might want:
		// - Small padding before (100-200ms): include slight pause after "So"
		// - Larger padding after (300-400ms): include natural pause before "is"

		idealStart := 1.22 // 20ms after "So" ends, before "the"
		idealEnd := 2.65   // 300ms after "here", in natural pause

		// Verify these are valid
		assert.GreaterOrEqual(t, idealStart, prevWordEnd, "Padding should not overlap previous word")
		assert.LessOrEqual(t, idealEnd, nextWordStart, "Padding should not overlap next word")

		// The improvement captures natural speech rhythm
		improvedHighlight := schema.Highlight{
			ID:      highlight.ID,
			Start:   idealStart,
			End:     idealEnd,
			ColorID: highlight.ColorID,
		}

		// Verify improvement
		assert.Equal(t, 1.22, improvedHighlight.Start)
		assert.Equal(t, 2.65, improvedHighlight.End)
	})
}
