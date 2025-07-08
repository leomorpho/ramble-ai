package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"MYAPP/ent"
	"MYAPP/ent/settings"
)


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


// ApiKeyService provides API key testing functionality
type ApiKeyService struct {
	client *ent.Client
	ctx    context.Context
}

// NewApiKeyService creates a new API key service
func NewApiKeyService(client *ent.Client, ctx context.Context) *ApiKeyService {
	return &ApiKeyService{
		client: client,
		ctx:    ctx,
	}
}

// getSetting retrieves a setting value by key
func (s *ApiKeyService) getSetting(key string) (string, error) {
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
func (s *ApiKeyService) TestOpenAIApiKey() (*TestOpenAIApiKeyResponse, error) {
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
func (s *ApiKeyService) TestOpenRouterApiKey() (*TestOpenRouterApiKeyResponse, error) {
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
func (s *ApiKeyService) testOpenAIConnection(apiKey string) (*TestOpenAIApiKeyResponse, error) {
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
func (s *ApiKeyService) testOpenRouterConnection(apiKey string) (*TestOpenRouterApiKeyResponse, error) {
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

// getOpenAIApiKey retrieves the OpenAI API key from settings
func (s *ApiKeyService) getOpenAIApiKey() (string, error) {
	return s.getSetting("openai_api_key")
}

// getOpenRouterApiKey retrieves the OpenRouter API key from settings
func (s *ApiKeyService) getOpenRouterApiKey() (string, error) {
	return s.getSetting("openrouter_api_key")
}