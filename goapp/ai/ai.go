package ai

import (
	"context"
	"fmt"

	"MYAPP/ent"
	"MYAPP/ent/settings"
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