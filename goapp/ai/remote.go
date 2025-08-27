package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"ramble-ai/ent"
)

// RemoteAIService handles AI processing through remote backend
type RemoteAIService struct {
	client     *ent.Client
	ctx        context.Context
	httpClient *http.Client
	backendURL string
	apiKey     string
}

// NewRemoteAIService creates a new remote AI service
func NewRemoteAIService(client *ent.Client, ctx context.Context, backendURL, apiKey string) *RemoteAIService {
	return &RemoteAIService{
		client:     client,
		ctx:        ctx,
		backendURL: backendURL,
		apiKey:     apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Longer timeout for AI processing
		},
	}
}

// ProcessText handles text processing requests via remote backend
func (s *RemoteAIService) ProcessText(request *TextProcessingRequest) (*TextProcessingResult, error) {
	if s.backendURL == "" {
		return nil, fmt.Errorf("backend URL not configured")
	}
	if s.apiKey == "" {
		return nil, fmt.Errorf("API key not configured")
	}

	// Build the full URL
	url := s.backendURL + "/api/ai/process-text"

	// Marshal request
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
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

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if json.Unmarshal(body, &errorResp) == nil {
			if errorMsg, ok := errorResp["error"].(string); ok {
				return nil, fmt.Errorf("remote AI service error: %s", errorMsg)
			}
		}
		return nil, fmt.Errorf("remote AI service error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result TextProcessingResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// ProcessAudio handles audio processing requests via remote backend
func (s *RemoteAIService) ProcessAudio(audioFile string) (*AudioProcessingResult, error) {
	if s.backendURL == "" {
		return nil, fmt.Errorf("backend URL not configured")
	}
	if s.apiKey == "" {
		return nil, fmt.Errorf("API key not configured")
	}

	// Read the audio file
	audioData, err := os.ReadFile(audioFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %w", err)
	}

	// Create request with base64 encoded audio
	request := map[string]string{
		"audio_data": base64.StdEncoding.EncodeToString(audioData),
		"filename":   filepath.Base(audioFile),
	}

	// Build the full URL
	url := s.backendURL + "/api/ai/process-audio"

	// Marshal request
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Make request with longer timeout for audio processing
	client := &http.Client{
		Timeout: 120 * time.Second, // 2 minutes for larger audio files
	}
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

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if json.Unmarshal(body, &errorResp) == nil {
			if errorMsg, ok := errorResp["error"].(string); ok {
				return nil, fmt.Errorf("remote AI service error: %s", errorMsg)
			}
		}
		return nil, fmt.Errorf("remote AI service error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result AudioProcessingResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}