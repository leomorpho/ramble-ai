package ai

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"

	"ramble-ai/ent/enttest"

	_ "github.com/mattn/go-sqlite3"
)

// getTestMaxWhisperFileSize returns a hardcoded 25MB for consistent testing
// This ensures tests are not affected by environment variables
func getTestMaxWhisperFileSize() int64 {
	return 25 * 1024 * 1024 // Always 25MB for tests
}

func TestAnalyzeAudioFile(t *testing.T) {
	// Save original environment variable and ensure test uses default 25MB limit
	originalMaxSize := os.Getenv("WHISPER_MAX_FILE_SIZE")
	os.Unsetenv("WHISPER_MAX_FILE_SIZE")
	resetWhisperConfigLog() // Reset sync.Once so it will re-evaluate with new env
	defer func() {
		if originalMaxSize != "" {
			os.Setenv("WHISPER_MAX_FILE_SIZE", originalMaxSize)
		}
		resetWhisperConfigLog() // Reset again to restore original behavior
	}()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	service := NewCoreAIService(client, context.Background())

	tests := []struct {
		name           string
		fileSize       int64
		expectedChunks bool
		expectedCount  int
	}{
		{
			name:           "Small file under 25MB",
			fileSize:       10 * 1024 * 1024, // 10MB
			expectedChunks: false,
			expectedCount:  0,
		},
		{
			name:           "Exactly 25MB",
			fileSize:       getTestMaxWhisperFileSize(),
			expectedChunks: false,
			expectedCount:  0,
		},
		{
			name:           "Just over 25MB",
			fileSize:       getTestMaxWhisperFileSize() + 1,
			expectedChunks: true,
			expectedCount:  2, // Minimum 2 chunks
		},
		{
			name:           "50MB file",
			fileSize:       50 * 1024 * 1024,
			expectedChunks: true,
			expectedCount:  3, // ~28 minutes estimated
		},
		{
			name:           "100MB file",
			fileSize:       100 * 1024 * 1024,
			expectedChunks: true,
			expectedCount:  6, // ~56 minutes estimated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file with specified size
			tempFile, err := os.CreateTemp("", "test_audio_*.mp3")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())

			// Write specified amount of data
			if _, err := tempFile.Seek(tt.fileSize-1, 0); err != nil {
				t.Fatalf("Failed to seek in temp file: %v", err)
			}
			if _, err := tempFile.Write([]byte{0}); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tempFile.Close()

			chunkInfo, err := service.analyzeAudioFile(tempFile.Name())
			if err != nil {
				t.Fatalf("analyzeAudioFile failed: %v", err)
			}

			if chunkInfo.NeedsChunking != tt.expectedChunks {
				t.Errorf("Expected NeedsChunking = %v, got %v", tt.expectedChunks, chunkInfo.NeedsChunking)
			}

			if chunkInfo.FileSizeBytes != tt.fileSize {
				t.Errorf("Expected FileSizeBytes = %d, got %d", tt.fileSize, chunkInfo.FileSizeBytes)
			}

			if tt.expectedChunks {
				if chunkInfo.ChunkCount < 2 {
					t.Errorf("Expected at least 2 chunks for large file, got %d", chunkInfo.ChunkCount)
				}
				// Allow some variance in chunk count estimation
				if math.Abs(float64(chunkInfo.ChunkCount-tt.expectedCount)) > 2 {
					t.Errorf("Expected chunk count around %d, got %d", tt.expectedCount, chunkInfo.ChunkCount)
				}
			}
		})
	}
}

func TestAdjustWordTimestamps(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	service := NewCoreAIService(client, context.Background())

	originalWords := []Word{
		{Word: "Hello", Start: 0.0, End: 0.5},
		{Word: "world", Start: 0.5, End: 1.0},
		{Word: "this", Start: 1.0, End: 1.5},
		{Word: "is", Start: 1.5, End: 2.0},
		{Word: "test", Start: 2.0, End: 2.5},
	}

	tests := []struct {
		name        string
		startOffset float64
		expected    []Word
	}{
		{
			name:        "No offset (chunk 0)",
			startOffset: 0.0,
			expected:    originalWords, // Should be unchanged
		},
		{
			name:        "570 second offset (chunk 1)",
			startOffset: 570.0, // 9.5 minutes
			expected: []Word{
				{Word: "Hello", Start: 570.0, End: 570.5},
				{Word: "world", Start: 570.5, End: 571.0},
				{Word: "this", Start: 571.0, End: 571.5},
				{Word: "is", Start: 571.5, End: 572.0},
				{Word: "test", Start: 572.0, End: 572.5},
			},
		},
		{
			name:        "1170 second offset (chunk 2)",
			startOffset: 1170.0, // 19.5 minutes
			expected: []Word{
				{Word: "Hello", Start: 1170.0, End: 1170.5},
				{Word: "world", Start: 1170.5, End: 1171.0},
				{Word: "this", Start: 1171.0, End: 1171.5},
				{Word: "is", Start: 1171.5, End: 1172.0},
				{Word: "test", Start: 1172.0, End: 1172.5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adjusted := service.adjustWordTimestamps(originalWords, tt.startOffset)

			if len(adjusted) != len(tt.expected) {
				t.Fatalf("Expected %d words, got %d", len(tt.expected), len(adjusted))
			}

			for i, word := range adjusted {
				expected := tt.expected[i]
				if word.Word != expected.Word {
					t.Errorf("Word %d: expected text %q, got %q", i, expected.Word, word.Word)
				}
				if math.Abs(word.Start-expected.Start) > 0.001 {
					t.Errorf("Word %d: expected start %.3f, got %.3f", i, expected.Start, word.Start)
				}
				if math.Abs(word.End-expected.End) > 0.001 {
					t.Errorf("Word %d: expected end %.3f, got %.3f", i, expected.End, word.End)
				}
			}
		})
	}
}

func TestAdjustSegmentTimestamps(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	service := NewCoreAIService(client, context.Background())

	originalSegments := []Segment{
		{
			ID:    1,
			Start: 0.0,
			End:   2.5,
			Text:  "Hello world this is test",
			Words: []Word{
				{Word: "Hello", Start: 0.0, End: 0.5},
				{Word: "world", Start: 0.5, End: 1.0},
				{Word: "this", Start: 1.0, End: 1.5},
				{Word: "is", Start: 1.5, End: 2.0},
				{Word: "test", Start: 2.0, End: 2.5},
			},
		},
	}

	startOffset := 570.0 // 9.5 minutes
	adjusted := service.adjustSegmentTimestamps(originalSegments, startOffset)

	if len(adjusted) != 1 {
		t.Fatalf("Expected 1 segment, got %d", len(adjusted))
	}

	segment := adjusted[0]
	if math.Abs(segment.Start-570.0) > 0.001 {
		t.Errorf("Expected segment start 570.0, got %.3f", segment.Start)
	}
	if math.Abs(segment.End-572.5) > 0.001 {
		t.Errorf("Expected segment end 572.5, got %.3f", segment.End)
	}

	// Check that words within segment are also adjusted
	if len(segment.Words) != 5 {
		t.Fatalf("Expected 5 words in segment, got %d", len(segment.Words))
	}

	firstWord := segment.Words[0]
	if math.Abs(firstWord.Start-570.0) > 0.001 {
		t.Errorf("Expected first word start 570.0, got %.3f", firstWord.Start)
	}

	lastWord := segment.Words[4]
	if math.Abs(lastWord.End-572.5) > 0.001 {
		t.Errorf("Expected last word end 572.5, got %.3f", lastWord.End)
	}
}

func TestRemoveOverlapWords(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	service := NewCoreAIService(client, context.Background())

	// Simulate chunk 1 words (already adjusted with startOffset = 570)
	currentWords := []Word{
		// These should be removed (in overlap region 570-600)
		{Word: "overlap1", Start: 575.0, End: 580.0},
		{Word: "overlap2", Start: 585.0, End: 590.0},
		{Word: "overlap3", Start: 595.0, End: 599.0},
		// These should be kept (after overlap)
		{Word: "keep1", Start: 600.0, End: 605.0},
		{Word: "keep2", Start: 605.0, End: 610.0},
	}

	prevChunk := &ChunkResult{
		ChunkIndex:  0,
		StartOffset: 0,
		EndOffset:   600.0, // 10 minutes
	}

	currentChunk := &ChunkResult{
		ChunkIndex:  1,
		StartOffset: 570.0, // 9.5 minutes (570s = 9.5 * 60)
		EndOffset:   1200.0, // 20 minutes
	}

	result := service.removeOverlapWords(currentWords, prevChunk, currentChunk)

	// Should keep only the 2 words that start at or after 600.0
	if len(result) != 2 {
		t.Errorf("Expected 2 words after deduplication, got %d", len(result))
		for i, word := range result {
			t.Logf("Kept word %d: %s (%.1f-%.1f)", i, word.Word, word.Start, word.End)
		}
	}

	// Check that we kept the right words
	if len(result) >= 1 && result[0].Word != "keep1" {
		t.Errorf("Expected first kept word to be 'keep1', got %q", result[0].Word)
	}
	if len(result) >= 2 && result[1].Word != "keep2" {
		t.Errorf("Expected second kept word to be 'keep2', got %q", result[1].Word)
	}
}

func TestMergeChunkResults(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	service := NewCoreAIService(client, context.Background())

	// Create mock chunk results
	chunkResults := []*ChunkResult{
		{
			ChunkIndex:   0,
			StartOffset:  0,
			EndOffset:    600, // 10 minutes
			OverlapStart: 570, // Last 30 seconds
			Result: &AudioProcessingResult{
				Transcript: "This is the first chunk of audio",
				Language:   "en",
				Duration:   600.0,
				Words: []Word{
					{Word: "This", Start: 0.0, End: 0.5},
					{Word: "is", Start: 0.5, End: 1.0},
					{Word: "first", Start: 1.0, End: 1.5},
					{Word: "chunk", Start: 570.0, End: 575.0}, // In overlap region
				},
				Segments: []Segment{
					{
						ID:    1,
						Start: 0.0,
						End:   600.0,
						Text:  "This is the first chunk of audio",
						Words: []Word{
							{Word: "This", Start: 0.0, End: 0.5},
							{Word: "is", Start: 0.5, End: 1.0},
							{Word: "first", Start: 1.0, End: 1.5},
							{Word: "chunk", Start: 570.0, End: 575.0},
						},
					},
				},
			},
		},
		{
			ChunkIndex:   1,
			StartOffset:  570, // Starts at 9.5 minutes
			EndOffset:    1170, // Ends at 19.5 minutes
			OverlapStart: 570,  // 30 seconds overlap with previous
			Result: &AudioProcessingResult{
				Transcript: "continuation of the audio",
				Language:   "en",
				Duration:   600.0,
				Words: []Word{
					// These would be removed during deduplication (overlap region)
					{Word: "chunk", Start: 0.0, End: 5.0},    // Will be adjusted to 570-575 (overlaps)
					{Word: "continuation", Start: 30.0, End: 35.0}, // Will be adjusted to 600-605 (kept)
					{Word: "audio", Start: 35.0, End: 40.0},        // Will be adjusted to 605-610 (kept)
				},
			},
		},
	}

	result, err := service.mergeChunkResults(chunkResults)
	if err != nil {
		t.Fatalf("mergeChunkResults failed: %v", err)
	}

	// Check basic properties
	if result.Language != "en" {
		t.Errorf("Expected language 'en', got %q", result.Language)
	}

	if len(result.Words) == 0 {
		t.Error("Expected some words in merged result")
	}

	// Check that timestamps are properly ordered
	for i := 1; i < len(result.Words); i++ {
		if result.Words[i].Start < result.Words[i-1].Start {
			t.Errorf("Words not properly ordered: word %d start %.3f < word %d start %.3f",
				i, result.Words[i].Start, i-1, result.Words[i-1].Start)
		}
	}

	// Check that we have a reasonable total duration
	if result.Duration < 600 { // Should be at least as long as first chunk
		t.Errorf("Expected duration >= 600, got %.1f", result.Duration)
	}

	// Check that transcript is not empty
	if len(result.Transcript) == 0 {
		t.Error("Expected non-empty transcript")
	}

	t.Logf("Merged result: %d words, %.1fs duration, transcript: %q",
		len(result.Words), result.Duration, result.Transcript[:min(50, len(result.Transcript))])
}

func TestTimestampContinuity(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	service := NewCoreAIService(client, context.Background())

	// Test that chunk boundaries maintain timestamp continuity
	chunk0Words := []Word{
		{Word: "end", Start: 595.0, End: 600.0}, // End of chunk 0
	}

	chunk1Words := []Word{
		{Word: "start", Start: 0.0, End: 5.0}, // Start of chunk 1 (before adjustment)
	}

	// Adjust chunk 1 words
	adjustedChunk1 := service.adjustWordTimestamps(chunk1Words, 570.0)

	// The gap should be reasonable (within overlap region)
	lastWordChunk0 := chunk0Words[len(chunk0Words)-1]
	firstWordChunk1 := adjustedChunk1[0]

	gap := firstWordChunk1.Start - lastWordChunk0.End
	if gap < -35.0 || gap > 5.0 { // Allow for some overlap but not too much
		t.Errorf("Unexpected gap between chunks: %.1fs (from %.1f to %.1f)",
			gap, lastWordChunk0.End, firstWordChunk1.Start)
	}

	t.Logf("Timestamp continuity: chunk0 ends %.1fs, chunk1 starts %.1fs, gap %.1fs",
		lastWordChunk0.End, firstWordChunk1.Start, gap)
}

func TestEdgeCases(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	service := NewCoreAIService(client, context.Background())

	t.Run("Empty chunk results", func(t *testing.T) {
		_, err := service.mergeChunkResults([]*ChunkResult{})
		if err == nil {
			t.Error("Expected error for empty chunk results")
		}
	})

	t.Run("Single chunk result", func(t *testing.T) {
		singleChunk := []*ChunkResult{
			{
				ChunkIndex:  0,
				StartOffset: 0,
				EndOffset:   600,
				Result: &AudioProcessingResult{
					Transcript: "Single chunk",
					Language:   "en",
					Duration:   600.0,
					Words:      []Word{{Word: "Single", Start: 0, End: 1}},
				},
			},
		}

		result, err := service.mergeChunkResults(singleChunk)
		if err != nil {
			t.Fatalf("Failed to merge single chunk: %v", err)
		}

		if result.Transcript != "Single chunk" {
			t.Errorf("Expected transcript 'Single chunk', got %q", result.Transcript)
		}
	})

	t.Run("Empty words in chunk", func(t *testing.T) {
		emptyWordsChunk := []*ChunkResult{
			{
				ChunkIndex:  0,
				StartOffset: 0,
				EndOffset:   600,
				Result: &AudioProcessingResult{
					Transcript: "No words",
					Language:   "en",
					Duration:   600.0,
					Words:      []Word{}, // Empty
				},
			},
		}

		result, err := service.mergeChunkResults(emptyWordsChunk)
		if err != nil {
			t.Fatalf("Failed to merge chunk with empty words: %v", err)
		}

		if len(result.Words) != 0 {
			t.Errorf("Expected 0 words, got %d", len(result.Words))
		}
	})
}

// Helper function for min (Go 1.18+ has this built-in)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Benchmark tests for performance
func BenchmarkAdjustWordTimestamps(b *testing.B) {
	client := enttest.Open(b, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	service := NewCoreAIService(client, context.Background())

	// Create a large set of words to adjust
	words := make([]Word, 1000)
	for i := 0; i < 1000; i++ {
		words[i] = Word{
			Word:  fmt.Sprintf("word%d", i),
			Start: float64(i),
			End:   float64(i) + 0.5,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.adjustWordTimestamps(words, 570.0)
	}
}

func BenchmarkRemoveOverlapWords(b *testing.B) {
	client := enttest.Open(b, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	service := NewCoreAIService(client, context.Background())

	// Create words spanning the overlap region
	words := make([]Word, 100)
	for i := 0; i < 100; i++ {
		words[i] = Word{
			Word:  fmt.Sprintf("word%d", i),
			Start: float64(590 + i), // Some in overlap, some after
			End:   float64(590 + i + 1),
		}
	}

	prevChunk := &ChunkResult{EndOffset: 600.0}
	currentChunk := &ChunkResult{ChunkIndex: 1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.removeOverlapWords(words, prevChunk, currentChunk)
	}
}