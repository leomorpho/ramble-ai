package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"ramble-ai/ent"
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

// ProcessAudio handles all audio processing tasks (currently transcription)
func (s *CoreAIService) ProcessAudio(audioFile string, apiKey string) (*AudioProcessingResult, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not provided")
	}

	// Check if file exists
	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("audio file does not exist: %s", audioFile)
	}

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

// ProcessText handles all text-based AI processing tasks
func (s *CoreAIService) ProcessText(request *TextProcessingRequest, apiKey string) (*OpenRouterResponse, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenRouter API key not provided")
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