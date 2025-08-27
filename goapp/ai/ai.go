package ai

import (
	"context"
	"fmt"

	"ramble-ai/ent"
	"ramble-ai/ent/settings"
)

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

// getOpenAIApiKey retrieves the OpenAI API key from settings
func (s *ApiKeyService) getOpenAIApiKey() (string, error) {
	return s.getSetting("openai_api_key")
}

// getOpenRouterApiKey retrieves the OpenRouter API key from settings
func (s *ApiKeyService) getOpenRouterApiKey() (string, error) {
	return s.getSetting("openrouter_api_key")
}

// getUseRemoteAIBackend retrieves the remote AI backend toggle setting
func (s *ApiKeyService) getUseRemoteAIBackend() (bool, error) {
	value, err := s.getSetting("use_remote_ai_backend")
	if err != nil {
		return false, err
	}
	if value == "" {
		return false, nil // default to false
	}
	return value == "true", nil
}

// getRemoteAIBackendURL retrieves the remote AI backend URL setting
func (s *ApiKeyService) getRemoteAIBackendURL() (string, error) {
	return s.getSetting("remote_ai_backend_url")
}

// getRambleAIApiKey retrieves the Ramble AI API key setting
func (s *ApiKeyService) getRambleAIApiKey() (string, error) {
	return s.getSetting("ramble_ai_api_key")
}
