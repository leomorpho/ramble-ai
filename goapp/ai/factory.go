package ai

import (
	"context"
	"fmt"
	"os"

	"ramble-ai/ent"
	"ramble-ai/ent/settings"
)

// AIServiceFactory creates AI services based on configuration
type AIServiceFactory struct {
	client *ent.Client
	ctx    context.Context
}

// NewAIServiceFactory creates a new AI service factory
func NewAIServiceFactory(client *ent.Client, ctx context.Context) *AIServiceFactory {
	return &AIServiceFactory{
		client: client,
		ctx:    ctx,
	}
}

// CreateService creates an appropriate AI service based on current settings
func (f *AIServiceFactory) CreateService() (AIService, error) {
	// Check if remote backend is enabled
	useRemote, err := f.getUseRemoteAIBackend()
	if err != nil {
		return nil, fmt.Errorf("failed to get backend setting: %w", err)
	}

	if useRemote {
		return f.createRemoteService()
	}
	
	return f.createLocalService()
}

// createRemoteService creates a remote AI service
func (f *AIServiceFactory) createRemoteService() (AIService, error) {
	// Get remote settings - check environment variable first
	var backendURL string
	if envURL := os.Getenv("REMOTE_AI_BACKEND_URL"); envURL != "" {
		backendURL = envURL
	} else {
		var err error
		backendURL, err = f.getSetting("remote_ai_backend_url")
		if err != nil {
			return nil, fmt.Errorf("failed to get backend URL: %w", err)
		}
	}
	
	if backendURL == "" {
		return nil, fmt.Errorf("backend URL not configured")
	}

	apiKey, err := f.getSetting("ramble_ai_api_key")
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("Ramble AI API key not configured")
	}

	return NewRemoteAIService(f.client, f.ctx, backendURL, apiKey), nil
}

// createLocalService creates a local AI service with API keys pre-loaded
func (f *AIServiceFactory) createLocalService() (AIService, error) {
	// Get local API keys
	openaiKey, err := f.getSetting("openai_api_key")
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenAI API key: %w", err)
	}

	openrouterKey, err := f.getSetting("openrouter_api_key")
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenRouter API key: %w", err)
	}

	return NewLocalAIService(f.client, f.ctx, openaiKey, openrouterKey), nil
}

// Helper methods

func (f *AIServiceFactory) getSetting(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("setting key cannot be empty")
	}

	setting, err := f.client.Settings.
		Query().
		Where(settings.Key(key)).
		Only(f.ctx)

	if err != nil {
		// Return empty string if setting doesn't exist
		return "", nil
	}

	return setting.Value, nil
}

func (f *AIServiceFactory) getUseRemoteAIBackend() (bool, error) {
	// Check environment variable first
	if envValue := os.Getenv("USE_REMOTE_AI_BACKEND"); envValue != "" {
		return envValue == "true", nil
	}
	
	// Fallback to database setting
	value, err := f.getSetting("use_remote_ai_backend")
	if err != nil {
		return false, err
	}
	if value == "" {
		return false, nil // default to false
	}
	return value == "true", nil
}