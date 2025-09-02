package ai

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"ramble-ai/ent"
	"ramble-ai/goapp"
)

// CoreAIService provides centralized AI processing functionality
type CoreAIService struct {
	client     *ent.Client
	ctx        context.Context
	httpClient *http.Client
}

// TextProcessingRequest represents a request for text-based AI processing
type TextProcessingRequest struct {
	SystemPrompt string                 `json:"system_prompt"`
	UserPrompt   string                 `json:"user_prompt"`
	Model        string                 `json:"model"`
	TaskType     string                 `json:"task_type"` // "suggest_highlights", "reorder", "improve_silences", "chat"
	Context      map[string]interface{} `json:"context,omitempty"`
}

// AudioProcessingResult represents the result of audio processing
type AudioProcessingResult struct {
	Transcript string    `json:"transcript"`
	Duration   float64   `json:"duration,omitempty"`
	Language   string    `json:"language,omitempty"`
	Words      []Word    `json:"words,omitempty"`
	Segments   []Segment `json:"segments,omitempty"`
}

// Word represents a word with timestamps
type Word struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// Segment represents a segment with timestamps  
type Segment struct {
	ID               int     `json:"id"`
	Seek             int     `json:"seek"`
	Start            float64 `json:"start"`
	End              float64 `json:"end"`
	Text             string  `json:"text"`
	Tokens           []int   `json:"tokens"`
	Temperature      float64 `json:"temperature"`
	AvgLogprob       float64 `json:"avg_logprob"`
	CompressionRatio float64 `json:"compression_ratio"`
	NoSpeechProb     float64 `json:"no_speech_prob"`
	Words            []Word  `json:"words"`
}

// TextProcessingResult represents the result of text processing
type TextProcessingResult struct {
	Content    string      `json:"content"`
	TaskType   string      `json:"task_type"`
	Structured interface{} `json:"structured,omitempty"`
	TokensUsed int         `json:"tokens_used,omitempty"`
}

// OpenRouterRequest represents the request format for OpenRouter API
type OpenRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenRouterResponse represents the response from OpenRouter API
type OpenRouterResponse struct {
	Choices []Choice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// Choice represents a response choice
type Choice struct {
	Message Message `json:"message"`
}

// Audio chunking constants
const (
	DefaultMaxWhisperFileSize = 25 * 1024 * 1024 // 25MB default limit for Whisper API
	DefaultChunkDurationSec   = 600               // 10 minutes per chunk
	DefaultOverlapSec         = 30                // 30 seconds overlap between chunks
)

var whisperConfigLogOnce sync.Once

// resetWhisperConfigLog resets the sync.Once for testing purposes
// This function should only be used in tests
func resetWhisperConfigLog() {
	whisperConfigLogOnce = sync.Once{}
}

// GetMaxWhisperFileSize returns the maximum file size for Whisper API
// Can be overridden with WHISPER_MAX_FILE_SIZE environment variable for testing
func GetMaxWhisperFileSize() int64 {
	var maxSize int64
	var source string
	
	if maxSizeStr := os.Getenv("WHISPER_MAX_FILE_SIZE"); maxSizeStr != "" {
		if parsedSize, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
			maxSize = parsedSize
			source = "environment variable"
		} else {
			log.Printf("Warning: Invalid WHISPER_MAX_FILE_SIZE value '%s', using default", maxSizeStr)
			maxSize = DefaultMaxWhisperFileSize
			source = "default (invalid env var)"
		}
	} else {
		maxSize = DefaultMaxWhisperFileSize
		source = "default"
	}
	
	// Log the configuration once
	whisperConfigLogOnce.Do(func() {
		sizeMB := float64(maxSize) / (1024 * 1024)
		log.Printf("[WHISPER_CONFIG] Max file size: %d bytes (%.1f MB) - source: %s", maxSize, sizeMB, source)
	})
	
	return maxSize
}

// ChunkResult represents the result of processing a single audio chunk
type ChunkResult struct {
	ChunkIndex   int                    `json:"chunk_index"`   // Index of this chunk (0, 1, 2, ...)
	StartOffset  float64                `json:"start_offset"`  // Absolute start time of chunk in original audio (seconds)
	EndOffset    float64                `json:"end_offset"`    // Absolute end time of chunk in original audio (seconds)
	OverlapStart float64                `json:"overlap_start"` // Start of overlap region in this chunk (seconds)
	Result       *AudioProcessingResult `json:"result"`        // Transcription result for this chunk
}

// ChunkInfo contains metadata about how an audio file should be chunked
type ChunkInfo struct {
	NeedsChunking   bool    `json:"needs_chunking"`
	FileSizeBytes   int64   `json:"file_size_bytes"`
	ChunkCount      int     `json:"chunk_count"`
	ChunkDurationSec int    `json:"chunk_duration_sec"`
	OverlapSec      int     `json:"overlap_sec"`
	TotalDuration   float64 `json:"total_duration,omitempty"` // If available from ffprobe
}

// OpenAITranscriptionResponse represents the response from OpenAI transcription API (verbose JSON)
type OpenAITranscriptionResponse struct {
	Task     string    `json:"task"`
	Language string    `json:"language"`
	Duration float64   `json:"duration"`
	Text     string    `json:"text"`
	Segments []Segment `json:"segments"`
	Words    []Word    `json:"words"`
}

// NewCoreAIService creates a new core AI service
func NewCoreAIService(client *ent.Client, ctx context.Context) *CoreAIService {
	return &CoreAIService{
		client: client,
		ctx:    ctx,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProcessAudio handles all audio processing tasks with automatic chunking for large files
func (s *CoreAIService) ProcessAudio(audioFile string, apiKey string) (*AudioProcessingResult, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not provided")
	}

	// Check if file exists
	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("audio file does not exist: %s", audioFile)
	}

	// Determine if chunking is needed
	chunkInfo, err := s.analyzeAudioFile(audioFile)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze audio file: %w", err)
	}

	log.Printf("[AUDIO_PROCESSING] File: %s, Size: %d bytes, Needs chunking: %v", 
		audioFile, chunkInfo.FileSizeBytes, chunkInfo.NeedsChunking)

	if !chunkInfo.NeedsChunking {
		// File is small enough, process normally
		return s.processSingleAudioFile(audioFile, apiKey)
	}

	// File is too large, use chunking approach
	return s.processAudioWithChunking(audioFile, apiKey, chunkInfo)
}

// analyzeAudioFile determines if an audio file needs chunking and calculates chunk info
func (s *CoreAIService) analyzeAudioFile(audioFile string) (*ChunkInfo, error) {
	// Get file size
	stat, err := os.Stat(audioFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	fileSizeBytes := stat.Size()
	needsChunking := fileSizeBytes > GetMaxWhisperFileSize()

	chunkInfo := &ChunkInfo{
		NeedsChunking:    needsChunking,
		FileSizeBytes:    fileSizeBytes,
		ChunkDurationSec: DefaultChunkDurationSec,
		OverlapSec:       DefaultOverlapSec,
	}

	if needsChunking {
		// Estimate number of chunks needed
		// Rough estimation: 10 minutes of 24k bitrate audio ≈ 18MB
		estimatedMinutes := float64(fileSizeBytes) / (1024 * 1024 * 1.8) // rough MB per minute at 24k bitrate
		chunksNeeded := int((estimatedMinutes / 10) + 0.5) // Round up
		if chunksNeeded < 2 {
			chunksNeeded = 2 // Minimum 2 chunks if chunking is needed
		}
		chunkInfo.ChunkCount = chunksNeeded

		log.Printf("[AUDIO_CHUNKING] File %s: %d bytes (%.2f MB), estimated %.1f minutes, will use %d chunks", 
			audioFile, fileSizeBytes, float64(fileSizeBytes)/(1024*1024), estimatedMinutes, chunksNeeded)
	}

	return chunkInfo, nil
}

// processSingleAudioFile handles normal processing for files ≤25MB
func (s *CoreAIService) processSingleAudioFile(audioFile string, apiKey string) (*AudioProcessingResult, error) {
	// Open the audio file
	file, err := os.Open(audioFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file field
	part, err := writer.CreateFormFile("file", audioFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}

	// Add model field
	err = writer.WriteField("model", "whisper-1")
	if err != nil {
		return nil, fmt.Errorf("failed to write model field: %w", err)
	}

	// Add response format field for verbose JSON with timestamps
	err = writer.WriteField("response_format", "verbose_json")
	if err != nil {
		return nil, fmt.Errorf("failed to write response_format field: %w", err)
	}

	// Add timestamp granularities for word-level timestamps
	err = writer.WriteField("timestamp_granularities[]", "word")
	if err != nil {
		return nil, fmt.Errorf("failed to write timestamp_granularities field: %w", err)
	}

	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Make request with longer timeout for transcription
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	// Parse response
	var transcriptionResp OpenAITranscriptionResponse
	err = json.Unmarshal(body, &transcriptionResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &AudioProcessingResult{
		Transcript: transcriptionResp.Text,
		Duration:   transcriptionResp.Duration,
		Language:   transcriptionResp.Language,
		Words:      transcriptionResp.Words,
		Segments:   transcriptionResp.Segments,
	}, nil
}

// processAudioWithChunking handles large audio files by splitting them into chunks
func (s *CoreAIService) processAudioWithChunking(audioFile string, apiKey string, chunkInfo *ChunkInfo) (*AudioProcessingResult, error) {
	log.Printf("[AUDIO_CHUNKING] Starting chunked processing for %s", audioFile)
	
	// Split audio into chunks
	chunkPaths, err := s.splitAudioIntoChunks(audioFile, chunkInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to split audio into chunks: %w", err)
	}
	
	// Ensure cleanup of temporary chunks
	defer s.cleanupChunks(chunkPaths)
	
	// Process all chunks in parallel
	chunkResults, err := s.processAudioChunks(chunkPaths, apiKey, chunkInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to process audio chunks: %w", err)
	}
	
	// Merge results with timestamp adjustment and overlap deduplication
	result, err := s.mergeChunkResults(chunkResults)
	if err != nil {
		return nil, fmt.Errorf("failed to merge chunk results: %w", err)
	}
	
	log.Printf("[AUDIO_CHUNKING] Successfully processed %d chunks, final transcript length: %d chars", 
		len(chunkResults), len(result.Transcript))
	
	return result, nil
}

// splitAudioIntoChunks splits an audio file into overlapping chunks using FFmpeg
func (s *CoreAIService) splitAudioIntoChunks(audioFile string, chunkInfo *ChunkInfo) ([]string, error) {
	// Create temp directory for chunks  
	tempDir := "temp_audio_chunks"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	var chunkPaths []string
	chunkDuration := float64(chunkInfo.ChunkDurationSec)
	overlap := float64(chunkInfo.OverlapSec)
	
	for i := 0; i < chunkInfo.ChunkCount; i++ {
		// Calculate start time for this chunk
		startTime := float64(i) * (chunkDuration - overlap)
		if i > 0 {
			// Adjust for overlap - chunks after first start earlier
			startTime = float64(i)*(chunkDuration) - overlap
		}
		
		// Generate unique chunk filename
		hash := md5.Sum([]byte(audioFile + fmt.Sprintf("_%d_%f", i, startTime)))
		chunkFilename := fmt.Sprintf("chunk_%d_%s.mp3", i, hex.EncodeToString(hash[:8]))
		chunkPath := filepath.Join(tempDir, chunkFilename)
		
		// Use ffmpeg-go library to extract audio chunk with optimized settings
		if err := goapp.ExtractAudioChunk(audioFile, startTime, chunkDuration, chunkPath); err != nil {
			s.cleanupChunks(chunkPaths) // Cleanup any successful chunks
			return nil, fmt.Errorf("failed to extract audio chunk %d: %w", i, err)
		}
		
		// Verify chunk was created and get size
		if stat, err := os.Stat(chunkPath); err == nil {
			sizeMB := float64(stat.Size()) / (1024 * 1024)
			log.Printf("[AUDIO_CHUNKING] Created chunk %d: %s (%.2f MB, start: %.1fs)", 
				i, chunkPath, sizeMB, startTime)
				
			if stat.Size() > GetMaxWhisperFileSize() {
				log.Printf("[AUDIO_CHUNKING] WARNING: Chunk %d is %d bytes (>25MB limit)", i, stat.Size())
			}
		}
		
		chunkPaths = append(chunkPaths, chunkPath)
	}
	
	return chunkPaths, nil
}

// cleanupChunks removes temporary chunk files
func (s *CoreAIService) cleanupChunks(chunkPaths []string) {
	for _, path := range chunkPaths {
		if err := os.Remove(path); err != nil {
			log.Printf("[AUDIO_CHUNKING] Warning: failed to cleanup chunk %s: %v", path, err)
		}
	}
	
	// Try to remove the temp directory if it's empty
	if len(chunkPaths) > 0 {
		tempDir := filepath.Dir(chunkPaths[0])
		if err := os.Remove(tempDir); err != nil {
			// It's OK if this fails - directory might not be empty
			log.Printf("[AUDIO_CHUNKING] Could not remove temp directory %s: %v", tempDir, err)
		}
	}
}

// processAudioChunks processes multiple audio chunks in parallel
func (s *CoreAIService) processAudioChunks(chunkPaths []string, apiKey string, chunkInfo *ChunkInfo) ([]*ChunkResult, error) {
	var wg sync.WaitGroup
	chunkResults := make([]*ChunkResult, len(chunkPaths))
	errors := make([]error, len(chunkPaths))
	
	// Process chunks in parallel
	for i, chunkPath := range chunkPaths {
		wg.Add(1)
		go func(index int, path string) {
			defer wg.Done()
			
			log.Printf("[AUDIO_CHUNKING] Processing chunk %d: %s", index, path)
			
			// Process this chunk
			result, err := s.processSingleAudioFile(path, apiKey)
			if err != nil {
				errors[index] = fmt.Errorf("chunk %d failed: %w", index, err)
				return
			}
			
			// Calculate chunk timing info
			chunkDuration := float64(chunkInfo.ChunkDurationSec)
			overlap := float64(chunkInfo.OverlapSec)
			
			var startOffset float64
			if index == 0 {
				startOffset = 0
			} else {
				// Each subsequent chunk starts at: chunk_index * (duration - overlap)  
				startOffset = float64(index) * (chunkDuration - overlap)
			}
			
			endOffset := startOffset + chunkDuration
			overlapStart := chunkDuration - overlap // Where overlap begins in this chunk
			
			chunkResults[index] = &ChunkResult{
				ChunkIndex:   index,
				StartOffset:  startOffset,
				EndOffset:    endOffset,
				OverlapStart: overlapStart,
				Result:       result,
			}
			
			log.Printf("[AUDIO_CHUNKING] Completed chunk %d: %.1fs-%.1fs, %d words", 
				index, startOffset, endOffset, len(result.Words))
			
		}(i, chunkPath)
	}
	
	wg.Wait()
	
	// Check for errors
	var errorList []string
	for i, err := range errors {
		if err != nil {
			errorList = append(errorList, fmt.Sprintf("chunk %d: %v", i, err))
		}
	}
	
	if len(errorList) > 0 {
		return nil, fmt.Errorf("chunk processing failed: %s", strings.Join(errorList, "; "))
	}
	
	// Filter out nil results and sort by chunk index
	var validResults []*ChunkResult
	for _, result := range chunkResults {
		if result != nil {
			validResults = append(validResults, result)
		}
	}
	
	sort.Slice(validResults, func(i, j int) bool {
		return validResults[i].ChunkIndex < validResults[j].ChunkIndex
	})
	
	return validResults, nil
}

// mergeChunkResults combines multiple chunk results into a single result with proper timestamp adjustment
func (s *CoreAIService) mergeChunkResults(chunkResults []*ChunkResult) (*AudioProcessingResult, error) {
	if len(chunkResults) == 0 {
		return nil, fmt.Errorf("no chunk results to merge")
	}
	
	if len(chunkResults) == 1 {
		// Single chunk, just return it
		return chunkResults[0].Result, nil
	}
	
	log.Printf("[AUDIO_CHUNKING] Merging %d chunks with timestamp adjustment and overlap deduplication", len(chunkResults))
	
	var allWords []Word
	var allSegments []Segment
	var combinedTranscript []string
	var totalDuration float64
	language := chunkResults[0].Result.Language // Assume same language across chunks
	
	for i, chunkResult := range chunkResults {
		chunk := chunkResult.Result
		startOffset := chunkResult.StartOffset
		
		// Adjust timestamps for all words in this chunk
		adjustedWords := s.adjustWordTimestamps(chunk.Words, startOffset)
		
		// Adjust timestamps for all segments in this chunk  
		adjustedSegments := s.adjustSegmentTimestamps(chunk.Segments, startOffset)
		
		if i == 0 {
			// First chunk: use all words and segments
			allWords = append(allWords, adjustedWords...)
			allSegments = append(allSegments, adjustedSegments...)
			combinedTranscript = append(combinedTranscript, chunk.Transcript)
		} else {
			// Subsequent chunks: remove overlap region before merging
			prevChunk := chunkResults[i-1]
			
			// Remove words that are in the overlap region between chunks
			deduplicatedWords := s.removeOverlapWords(adjustedWords, prevChunk, chunkResult)
			allWords = append(allWords, deduplicatedWords...)
			
			// Remove segments in overlap region  
			deduplicatedSegments := s.removeOverlapSegments(adjustedSegments, prevChunk, chunkResult)
			allSegments = append(allSegments, deduplicatedSegments...)
			
			// For transcript, we'll need to be more careful about overlap
			cleanTranscript := s.removeOverlapFromTranscript(chunk.Transcript, deduplicatedWords)
			combinedTranscript = append(combinedTranscript, cleanTranscript)
		}
		
		// Track the maximum end time
		if len(adjustedWords) > 0 {
			lastWord := adjustedWords[len(adjustedWords)-1]
			if lastWord.End > totalDuration {
				totalDuration = lastWord.End
			}
		}
	}
	
	// Sort all words by start time to ensure proper ordering
	sort.Slice(allWords, func(i, j int) bool {
		return allWords[i].Start < allWords[j].Start
	})
	
	// Sort all segments by start time
	sort.Slice(allSegments, func(i, j int) bool {
		return allSegments[i].Start < allSegments[j].Start
	})
	
	finalTranscript := strings.Join(combinedTranscript, " ")
	
	log.Printf("[AUDIO_CHUNKING] Merge complete: %d words, %d segments, %.1fs duration", 
		len(allWords), len(allSegments), totalDuration)
	
	return &AudioProcessingResult{
		Transcript: finalTranscript,
		Duration:   totalDuration,
		Language:   language,
		Words:      allWords,
		Segments:   allSegments,
	}, nil
}

// adjustWordTimestamps adds the start offset to all word timestamps in a chunk
func (s *CoreAIService) adjustWordTimestamps(words []Word, startOffset float64) []Word {
	adjustedWords := make([]Word, len(words))
	for i, word := range words {
		adjustedWords[i] = Word{
			Word:  word.Word,
			Start: word.Start + startOffset,
			End:   word.End + startOffset,
		}
	}
	return adjustedWords
}

// adjustSegmentTimestamps adds the start offset to all segment timestamps in a chunk
func (s *CoreAIService) adjustSegmentTimestamps(segments []Segment, startOffset float64) []Segment {
	adjustedSegments := make([]Segment, len(segments))
	for i, segment := range segments {
		// Adjust segment timestamps
		adjustedSegment := Segment{
			ID:               segment.ID,
			Seek:             segment.Seek,
			Start:            segment.Start + startOffset,
			End:              segment.End + startOffset,
			Text:             segment.Text,
			Tokens:           segment.Tokens,
			Temperature:      segment.Temperature,
			AvgLogprob:       segment.AvgLogprob,
			CompressionRatio: segment.CompressionRatio,
			NoSpeechProb:     segment.NoSpeechProb,
		}
		
		// Adjust word timestamps within the segment
		adjustedSegment.Words = s.adjustWordTimestamps(segment.Words, startOffset)
		adjustedSegments[i] = adjustedSegment
	}
	return adjustedSegments
}

// removeOverlapWords removes words from the current chunk that fall in the overlap region with the previous chunk
func (s *CoreAIService) removeOverlapWords(currentWords []Word, prevChunk *ChunkResult, currentChunk *ChunkResult) []Word {
	if len(currentWords) == 0 {
		return currentWords
	}
	
	// Overlap region: from prevChunk.EndOffset - overlap to prevChunk.EndOffset
	// In the current chunk's adjusted timestamps, this translates to:
	overlapStartTime := prevChunk.EndOffset - float64(DefaultOverlapSec)
	overlapEndTime := prevChunk.EndOffset
	
	var deduplicatedWords []Word
	for _, word := range currentWords {
		// Keep words that start after the overlap region
		if word.Start >= overlapEndTime {
			deduplicatedWords = append(deduplicatedWords, word)
		} else if word.Start < overlapStartTime {
			// Keep words that start before overlap region (shouldn't happen in normal cases)
			deduplicatedWords = append(deduplicatedWords, word)
		}
		// Skip words that fall within the overlap region
	}
	
	log.Printf("[AUDIO_CHUNKING] Chunk %d: removed %d overlapping words (%.1fs-%.1fs)", 
		currentChunk.ChunkIndex, len(currentWords)-len(deduplicatedWords), overlapStartTime, overlapEndTime)
	
	return deduplicatedWords
}

// removeOverlapSegments removes segments from the current chunk that fall in the overlap region with the previous chunk
func (s *CoreAIService) removeOverlapSegments(currentSegments []Segment, prevChunk *ChunkResult, currentChunk *ChunkResult) []Segment {
	if len(currentSegments) == 0 {
		return currentSegments
	}
	
	overlapStartTime := prevChunk.EndOffset - float64(DefaultOverlapSec)
	overlapEndTime := prevChunk.EndOffset
	
	var deduplicatedSegments []Segment
	for _, segment := range currentSegments {
		// Keep segments that start after the overlap region
		if segment.Start >= overlapEndTime {
			deduplicatedSegments = append(deduplicatedSegments, segment)
		} else if segment.End <= overlapStartTime {
			// Keep segments that end before overlap region
			deduplicatedSegments = append(deduplicatedSegments, segment)
		} else if segment.Start < overlapStartTime && segment.End > overlapEndTime {
			// Segment spans the overlap region - keep it but filter its words
			filteredSegment := segment
			filteredSegment.Words = s.filterWordsInTimeRange(segment.Words, overlapStartTime, overlapEndTime)
			deduplicatedSegments = append(deduplicatedSegments, filteredSegment)
		}
		// Skip segments that fall entirely within the overlap region
	}
	
	return deduplicatedSegments
}

// filterWordsInTimeRange removes words that fall within a specific time range
func (s *CoreAIService) filterWordsInTimeRange(words []Word, excludeStartTime, excludeEndTime float64) []Word {
	var filteredWords []Word
	for _, word := range words {
		// Keep words that are outside the exclusion time range
		if word.End <= excludeStartTime || word.Start >= excludeEndTime {
			filteredWords = append(filteredWords, word)
		}
	}
	return filteredWords
}

// removeOverlapFromTranscript attempts to clean transcript text based on deduplicated words
func (s *CoreAIService) removeOverlapFromTranscript(originalTranscript string, deduplicatedWords []Word) string {
	if len(deduplicatedWords) == 0 {
		return ""
	}
	
	// Simple approach: reconstruct transcript from deduplicated words
	var wordTexts []string
	for _, word := range deduplicatedWords {
		// Clean up word text (remove leading/trailing spaces and punctuation artifacts)
		cleanWord := strings.TrimSpace(word.Word)
		if cleanWord != "" {
			wordTexts = append(wordTexts, cleanWord)
		}
	}
	
	return strings.Join(wordTexts, " ")
}

// ProcessText handles all text-based AI processing tasks
func (s *CoreAIService) ProcessText(request *TextProcessingRequest, apiKey string) (*OpenRouterResponse, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenRouter API key not configured")
	}

	if request.Model == "" {
		request.Model = "anthropic/claude-3.5-sonnet" // Default model
	}

	// Build messages array
	messages := []Message{}

	// Add system message if provided
	if request.SystemPrompt != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: request.SystemPrompt,
		})
	}

	// Add user message
	messages = append(messages, Message{
		Role:    "user",
		Content: request.UserPrompt,
	})

	// Create OpenRouter request
	openRouterReq := OpenRouterRequest{
		Model:    request.Model,
		Messages: messages,
	}

	// Marshal request
	jsonData, err := json.Marshal(openRouterReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenRouter API error: %s", string(body))
	}

	// Parse response
	var openRouterResp OpenRouterResponse
	err = json.Unmarshal(body, &openRouterResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if openRouterResp.Error != nil {
		return nil, fmt.Errorf("OpenRouter API error: %s", openRouterResp.Error.Message)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenRouter API")
	}

	return &openRouterResp, nil
}