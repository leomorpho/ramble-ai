package ai

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// TextProcessingRequest represents a request for text-based AI processing
type TextProcessingRequest struct {
	SystemPrompt string                 `json:"system_prompt"`
	UserPrompt   string                 `json:"user_prompt"`
	Model        string                 `json:"model"`
	TaskType     string                 `json:"task_type"` // "suggest_highlights", "reorder", "improve_silences", "chat"
	Context      map[string]interface{} `json:"context,omitempty"`
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

// AudioProcessingRequest represents a request for audio transcription
type AudioProcessingRequest struct {
	AudioData string `json:"audio_data"` // Base64 encoded audio file
	Filename  string `json:"filename"`   // Original filename for extension detection
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

// OpenAITranscriptionResponse represents the response from OpenAI transcription API
type OpenAITranscriptionResponse struct {
	Task     string    `json:"task"`
	Language string    `json:"language"`
	Duration float64   `json:"duration"`
	Text     string    `json:"text"`
	Segments []Segment `json:"segments"`
	Words    []Word    `json:"words"`
}

// ProcessTextHandler handles text processing requests
func ProcessTextHandler(e *core.RequestEvent, app core.App) error {
	// Validate API key
	apiKey := extractBearerToken(e.Request.Header.Get("Authorization"))
	if apiKey == "" {
		return e.JSON(401, map[string]string{"error": "Missing or invalid API key"})
	}

	// Check API key validity and get user
	user, err := validateAPIKey(app, apiKey)
	if err != nil {
		return e.JSON(401, map[string]string{"error": "Invalid API key"})
	}

	// Check user's subscription status (placeholder - implement based on your subscription model)
	if !isUserSubscribed(user) {
		return e.JSON(403, map[string]string{"error": "Active subscription required"})
	}

	// Parse request body
	var request TextProcessingRequest
	if err := e.BindBody(&request); err != nil {
		return e.JSON(400, map[string]string{"error": "Invalid request format"})
	}

	// Validate required fields
	if request.UserPrompt == "" {
		return e.JSON(400, map[string]string{"error": "user_prompt is required"})
	}

	// Set default model if not provided
	if request.Model == "" {
		request.Model = "anthropic/claude-3.5-sonnet"
	}

	// Proxy request to OpenRouter
	result, err := proxyToOpenRouter(&request)
	if err != nil {
		return e.JSON(500, map[string]string{"error": fmt.Sprintf("AI processing failed: %v", err)})
	}

	// Log usage (optional)
	logAIUsage(app, user.Id, request.TaskType, request.Model, 0) // TokensUsed not available in raw response

	return e.JSON(200, result)
}

// GenerateAPIKeyHandler generates a new API key for authenticated users
func GenerateAPIKeyHandler(e *core.RequestEvent, app core.App) error {
	// Get authenticated user
	user := e.Auth
	if user == nil {
		return e.JSON(401, map[string]string{"error": "Authentication required"})
	}

	// Generate API key
	apiKey := generateAPIKey()
	keyHash := hashAPIKey(apiKey)

	// Create API key record
	apiKeyCollection, err := app.FindCollectionByNameOrId("api_keys")
	if err != nil {
		return e.JSON(500, map[string]string{"error": "Failed to find API keys collection"})
	}

	record := core.NewRecord(apiKeyCollection)
	record.Set("key_hash", keyHash)
	record.Set("user_id", user.Id)
	record.Set("active", true)
	record.Set("name", fmt.Sprintf("API Key - %s", time.Now().Format("2006-01-02 15:04")))

	if err := app.Save(record); err != nil {
		return e.JSON(500, map[string]string{"error": "Failed to save API key"})
	}

	return e.JSON(200, map[string]string{
		"api_key": apiKey,
		"message": "API key generated successfully",
	})
}

// Helper functions

func extractBearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

func hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

func generateAPIKey() string {
	// Generate a secure random API key (simplified for demo)
	hash := sha256.Sum256([]byte(fmt.Sprintf("ramble-ai-%d", time.Now().UnixNano())))
	return "ra-" + hex.EncodeToString(hash[:])[:32]
}

func validateAPIKey(app core.App, apiKey string) (*core.Record, error) {
	keyHash := hashAPIKey(apiKey)
	
	// Find API key record
	apiKeyRecord, err := app.FindFirstRecordByFilter("api_keys", "key_hash = {:hash} && active = true", map[string]interface{}{
		"hash": keyHash,
	})
	if err != nil {
		return nil, fmt.Errorf("API key not found or inactive")
	}

	// Get user record
	userRecord, err := app.FindRecordById("users", apiKeyRecord.GetString("user_id"))
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return userRecord, nil
}

func isUserSubscribed(user *core.Record) bool {
	// Placeholder - implement your subscription logic
	// This could check a subscriptions collection or Stripe status
	return true // For now, allow all users
}

func proxyToOpenRouter(request *TextProcessingRequest) (*OpenRouterResponse, error) {
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

	// TODO: Get OpenRouter API key from environment or settings
	// For now, this would need to be configured
	openRouterAPIKey := getOpenRouterAPIKey()
	if openRouterAPIKey == "" {
		return nil, fmt.Errorf("OpenRouter API key not configured")
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+openRouterAPIKey)
	req.Header.Set("Content-Type", "application/json")

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
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

func getOpenRouterAPIKey() string {
	// Get OpenRouter API key from environment
	return os.Getenv("OPENROUTER_API_KEY")
}

func logAIUsage(app core.App, userID, taskType, model string, tokensUsed int) {
	// Optional: Log AI usage for analytics and billing
	// This could be implemented to track usage per user
}

// ProcessAudioHandler handles audio transcription requests
func ProcessAudioHandler(e *core.RequestEvent, app core.App) error {
	// Validate API key
	apiKey := extractBearerToken(e.Request.Header.Get("Authorization"))
	if apiKey == "" {
		return e.JSON(401, map[string]string{"error": "Missing or invalid API key"})
	}

	// Check API key validity and get user
	user, err := validateAPIKey(app, apiKey)
	if err != nil {
		return e.JSON(401, map[string]string{"error": "Invalid API key"})
	}

	// Check user's subscription status
	if !isUserSubscribed(user) {
		return e.JSON(403, map[string]string{"error": "Active subscription required"})
	}

	// Parse request body
	var request AudioProcessingRequest
	if err := e.BindBody(&request); err != nil {
		return e.JSON(400, map[string]string{"error": "Invalid request format"})
	}

	// Validate required fields
	if request.AudioData == "" {
		return e.JSON(400, map[string]string{"error": "audio_data is required"})
	}

	// Decode base64 audio data
	audioBytes, err := base64.StdEncoding.DecodeString(request.AudioData)
	if err != nil {
		return e.JSON(400, map[string]string{"error": "Invalid base64 audio data"})
	}

	// Create temporary file for audio
	tempDir := os.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "audio-*.wav")
	if err != nil {
		return e.JSON(500, map[string]string{"error": "Failed to create temporary file"})
	}
	defer os.Remove(tempFile.Name()) // Clean up temp file
	defer tempFile.Close()

	// Write audio data to temp file
	if _, err := tempFile.Write(audioBytes); err != nil {
		return e.JSON(500, map[string]string{"error": "Failed to write audio data"})
	}

	// Call OpenAI Whisper API
	result, err := callOpenAIWhisper(tempFile.Name(), request.Filename)
	if err != nil {
		return e.JSON(500, map[string]string{"error": fmt.Sprintf("Transcription failed: %v", err)})
	}

	// Log usage (optional)
	logAIUsage(app, user.Id, "transcription", "whisper-1", 0)

	return e.JSON(200, result)
}

// callOpenAIWhisper calls OpenAI's Whisper API for audio transcription
func callOpenAIWhisper(audioPath string, originalFilename string) (*AudioProcessingResult, error) {
	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	// Open the audio file
	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file field
	part, err := writer.CreateFormFile("file", filepath.Base(originalFilename))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}

	// Add model field
	if err := writer.WriteField("model", "whisper-1"); err != nil {
		return nil, fmt.Errorf("failed to write model field: %w", err)
	}

	// Add response format for verbose JSON with timestamps
	if err := writer.WriteField("response_format", "verbose_json"); err != nil {
		return nil, fmt.Errorf("failed to write response_format field: %w", err)
	}

	// Add timestamp granularities for word-level timestamps
	if err := writer.WriteField("timestamp_granularities[]", "word"); err != nil {
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

	// Make request
	client := &http.Client{Timeout: 60 * time.Second} // Longer timeout for audio processing
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

	// Parse response
	var transcriptionResp OpenAITranscriptionResponse
	if err := json.Unmarshal(body, &transcriptionResp); err != nil {
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