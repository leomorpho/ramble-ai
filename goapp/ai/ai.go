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
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"MYAPP/ent"
	"MYAPP/ent/schema"
	"MYAPP/ent/settings"
)

// Word represents a word with timing information from transcription
type Word struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// Segment represents a segment of transcribed audio
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

// WhisperResponse represents the response from OpenAI Whisper API
type WhisperResponse struct {
	Task     string    `json:"task"`
	Language string    `json:"language"`
	Duration float64   `json:"duration"`
	Text     string    `json:"text"`
	Segments []Segment `json:"segments"`
	Words    []Word    `json:"words"`
}

// TranscriptionResponse represents the response returned to the frontend
type TranscriptionResponse struct {
	Success       bool    `json:"success"`
	Message       string  `json:"message"`
	Transcription string  `json:"transcription,omitempty"`
	Words         []Word  `json:"words,omitempty"`
	Language      string  `json:"language,omitempty"`
	Duration      float64 `json:"duration,omitempty"`
}

// TestOpenAIApiKeyResponse represents the response from testing OpenAI API key
type TestOpenAIApiKeyResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
	Model   string `json:"model,omitempty"`
}

// TestOpenRouterApiKeyResponse represents the response from testing OpenRouter API key
type TestOpenRouterApiKeyResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
	Model   string `json:"model,omitempty"`
}

// TranscriptionService provides transcription functionality
type TranscriptionService struct {
	client *ent.Client
	ctx    context.Context
}

// NewTranscriptionService creates a new transcription service
func NewTranscriptionService(client *ent.Client, ctx context.Context) *TranscriptionService {
	return &TranscriptionService{
		client: client,
		ctx:    ctx,
	}
}

// TranscribeVideoClip transcribes audio from a video clip using OpenAI Whisper
func (s *TranscriptionService) TranscribeVideoClip(clipID int) (*TranscriptionResponse, error) {
	// Get the video clip
	clip, err := s.client.VideoClip.Get(s.ctx, clipID)
	if err != nil {
		return &TranscriptionResponse{
			Success: false,
			Message: "Video clip not found",
		}, nil
	}

	// Check if file exists
	if _, err := os.Stat(clip.FilePath); os.IsNotExist(err) {
		return &TranscriptionResponse{
			Success: false,
			Message: "Video file not found",
		}, nil
	}

	// Get OpenAI API key
	apiKey, err := s.getOpenAIApiKey()
	if err != nil || apiKey == "" {
		return &TranscriptionResponse{
			Success: false,
			Message: "OpenAI API key not configured",
		}, nil
	}

	// Extract audio from video
	audioPath, err := s.extractAudio(clip.FilePath)
	if err != nil {
		return &TranscriptionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to extract audio: %v", err),
		}, nil
	}
	defer os.Remove(audioPath) // Clean up temporary audio file

	// Transcribe audio using OpenAI Whisper
	whisperResponse, err := s.transcribeAudio(audioPath, apiKey)
	if err != nil {
		return &TranscriptionResponse{
			Success: false,
			Message: fmt.Sprintf("Transcription failed: %v", err),
		}, nil
	}

	// Convert Word structs for storage
	var wordsForStorage []schema.Word
	for _, w := range whisperResponse.Words {
		wordsForStorage = append(wordsForStorage, schema.Word{
			Word:  w.Word,
			Start: w.Start,
			End:   w.End,
		})
	}

	// Save transcription to database
	_, err = s.client.VideoClip.
		UpdateOneID(clipID).
		SetTranscription(whisperResponse.Text).
		SetTranscriptionWords(wordsForStorage).
		SetTranscriptionLanguage(whisperResponse.Language).
		SetTranscriptionDuration(whisperResponse.Duration).
		Save(s.ctx)

	if err != nil {
		return &TranscriptionResponse{
			Success: false,
			Message: "Failed to save transcription",
		}, nil
	}

	return &TranscriptionResponse{
		Success:       true,
		Message:       "Transcription completed successfully",
		Transcription: whisperResponse.Text,
		Words:         whisperResponse.Words,
		Language:      whisperResponse.Language,
		Duration:      whisperResponse.Duration,
	}, nil
}

// extractAudio extracts audio from a video file using ffmpeg
func (s *TranscriptionService) extractAudio(videoPath string) (string, error) {
	// Create temp directory for audio files
	tempDir := "temp_audio"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate unique audio filename
	hash := md5.Sum([]byte(videoPath + fmt.Sprintf("%d", time.Now().UnixNano())))
	audioFilename := hex.EncodeToString(hash[:]) + ".mp3"
	audioPath := filepath.Join(tempDir, audioFilename)

	log.Printf("[TRANSCRIPTION] Extracting audio from: %s to: %s", videoPath, audioPath)

	// Use ffmpeg to extract audio
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vn",            // No video
		"-acodec", "mp3", // Audio codec
		"-ar", "16000",   // Sample rate (16kHz for Whisper)
		"-ac", "1",       // Mono channel
		"-b:a", "64k",    // Bitrate
		"-y",             // Overwrite output file
		audioPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[TRANSCRIPTION] ffmpeg error: %v, output: %s", err, string(output))
		return "", fmt.Errorf("ffmpeg failed: %w", err)
	}

	log.Printf("[TRANSCRIPTION] Audio extracted successfully: %s", audioPath)
	return audioPath, nil
}

// transcribeAudio transcribes audio using OpenAI Whisper API
func (s *TranscriptionService) transcribeAudio(audioPath, apiKey string) (*WhisperResponse, error) {
	// Create HTTP client with longer timeout for transcription
	client := &http.Client{
		Timeout: 120 * time.Second, // 2 minutes for transcription
	}

	// Open audio file
	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file field
	fileWriter, err := writer.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}

	// Add model field
	err = writer.WriteField("model", "whisper-1")
	if err != nil {
		return nil, fmt.Errorf("failed to add model field: %w", err)
	}

	// Add response format field for verbose JSON with timestamps
	err = writer.WriteField("response_format", "verbose_json")
	if err != nil {
		return nil, fmt.Errorf("failed to add response format field: %w", err)
	}

	// Add timestamp granularities for word-level timestamps
	err = writer.WriteField("timestamp_granularities[]", "word")
	if err != nil {
		return nil, fmt.Errorf("failed to add timestamp granularities field: %w", err)
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

	log.Printf("[TRANSCRIPTION] Sending audio to OpenAI Whisper API")

	// Make request
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
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var whisperResponse WhisperResponse
	err = json.Unmarshal(body, &whisperResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transcription response: %w", err)
	}

	log.Printf("[TRANSCRIPTION] Transcription completed, text length: %d characters, words: %d",
		len(whisperResponse.Text), len(whisperResponse.Words))

	return &whisperResponse, nil
}

// getOpenAIApiKey retrieves the OpenAI API key from settings
func (s *TranscriptionService) getOpenAIApiKey() (string, error) {
	return s.getSetting("openai_api_key")
}

// getSetting retrieves a setting value by key
func (s *TranscriptionService) getSetting(key string) (string, error) {
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

// TestOpenAIApiKey tests the validity of the stored OpenAI API key
func (s *TranscriptionService) TestOpenAIApiKey() (*TestOpenAIApiKeyResponse, error) {
	// Get the stored API key
	apiKey, err := s.getOpenAIApiKey()
	if err != nil {
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: "Failed to retrieve API key from database",
		}, nil
	}

	if apiKey == "" {
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: "No API key found. Please set your OpenAI API key first.",
		}, nil
	}

	// Test the API key with a simple request to the models endpoint
	return s.testOpenAIConnection(apiKey)
}

// TestOpenRouterApiKey tests the validity of the stored OpenRouter API key
func (s *TranscriptionService) TestOpenRouterApiKey() (*TestOpenRouterApiKeyResponse, error) {
	// Get the stored API key
	apiKey, err := s.getOpenRouterApiKey()
	if err != nil {
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: "Failed to retrieve API key from database",
		}, nil
	}

	if apiKey == "" {
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: "No API key found. Please set your OpenRouter API key first.",
		}, nil
	}

	// Test the API key with a simple request to the models endpoint
	return s.testOpenRouterConnection(apiKey)
}

// testOpenAIConnection tests the OpenAI API connection with the given key
func (s *TranscriptionService) testOpenAIConnection(apiKey string) (*TestOpenAIApiKeyResponse, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request to list models (lightweight endpoint)
	req, err := http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
	if err != nil {
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: "Failed to create test request",
		}, nil
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: "Failed to connect to OpenAI API. Please check your internet connection.",
		}, nil
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: "Failed to read API response",
		}, nil
	}

	// Check response status
	switch resp.StatusCode {
	case http.StatusOK:
		// Parse response to get a model name
		var modelsResp struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}

		if err := json.Unmarshal(body, &modelsResp); err == nil && len(modelsResp.Data) > 0 {
			// Find Whisper model or use first available
			modelName := modelsResp.Data[0].ID
			for _, model := range modelsResp.Data {
				if strings.Contains(model.ID, "whisper") {
					modelName = model.ID
					break
				}
			}

			return &TestOpenAIApiKeyResponse{
				Valid:   true,
				Message: "API key is valid and working!",
				Model:   modelName,
			}, nil
		}

		return &TestOpenAIApiKeyResponse{
			Valid:   true,
			Message: "API key is valid and working!",
		}, nil

	case http.StatusUnauthorized:
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: "Invalid API key. Please check your OpenAI API key.",
		}, nil

	case http.StatusTooManyRequests:
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: "Rate limit exceeded. Please try again later.",
		}, nil

	case http.StatusForbidden:
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: "API key doesn't have sufficient permissions.",
		}, nil

	default:
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: fmt.Sprintf("API test failed with status %d: %s", resp.StatusCode, string(body)),
		}, nil
	}
}

// testOpenRouterConnection tests the OpenRouter API connection with the given key
func (s *TranscriptionService) testOpenRouterConnection(apiKey string) (*TestOpenRouterApiKeyResponse, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request to list models (lightweight endpoint)
	req, err := http.NewRequest("GET", "https://openrouter.ai/api/v1/models", nil)
	if err != nil {
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: "Failed to create test request",
		}, nil
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: "Failed to connect to OpenRouter API. Please check your internet connection.",
		}, nil
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: "Failed to read API response",
		}, nil
	}

	// Check response status
	switch resp.StatusCode {
	case http.StatusOK:
		// Parse response to get a model name
		var modelsResp struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}

		if err := json.Unmarshal(body, &modelsResp); err == nil && len(modelsResp.Data) > 0 {
			// Find a suitable model or use first available
			modelName := modelsResp.Data[0].ID
			for _, model := range modelsResp.Data {
				if strings.Contains(strings.ToLower(model.ID), "gpt") || strings.Contains(strings.ToLower(model.ID), "claude") {
					modelName = model.ID
					break
				}
			}

			return &TestOpenRouterApiKeyResponse{
				Valid:   true,
				Message: "API key is valid and working!",
				Model:   modelName,
			}, nil
		}

		return &TestOpenRouterApiKeyResponse{
			Valid:   true,
			Message: "API key is valid and working!",
		}, nil

	case http.StatusUnauthorized:
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: "Invalid API key. Please check your OpenRouter API key.",
		}, nil

	case http.StatusTooManyRequests:
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: "Rate limit exceeded. Please try again later.",
		}, nil

	case http.StatusForbidden:
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: "API key doesn't have sufficient permissions.",
		}, nil

	default:
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: fmt.Sprintf("API test failed with status %d: %s", resp.StatusCode, string(body)),
		}, nil
	}
}

// getOpenRouterApiKey retrieves the OpenRouter API key from settings
func (s *TranscriptionService) getOpenRouterApiKey() (string, error) {
	return s.getSetting("openrouter_api_key")
}