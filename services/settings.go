package services

import (
	"context"
	"fmt"

	"MYAPP/ent"
	"MYAPP/ent/settings"
)

// SettingsService handles settings-related operations
type SettingsService struct {
	client *ent.Client
	ctx    context.Context
}

// NewSettingsService creates a new settings service
func NewSettingsService(client *ent.Client, ctx context.Context) *SettingsService {
	return &SettingsService{
		client: client,
		ctx:    ctx,
	}
}

// SaveSetting saves a setting key-value pair to the database
func (ss *SettingsService) SaveSetting(key, value string) error {
	if key == "" {
		return fmt.Errorf("setting key cannot be empty")
	}

	// Check if setting already exists
	existingSetting, err := ss.client.Settings.
		Query().
		Where(settings.Key(key)).
		Only(ss.ctx)

	if err != nil {
		// Setting doesn't exist, create new one
		_, err = ss.client.Settings.
			Create().
			SetKey(key).
			SetValue(value).
			Save(ss.ctx)
		
		if err != nil {
			return fmt.Errorf("failed to create setting: %w", err)
		}
	} else {
		// Setting exists, update it
		_, err = ss.client.Settings.
			UpdateOne(existingSetting).
			SetValue(value).
			Save(ss.ctx)
		
		if err != nil {
			return fmt.Errorf("failed to update setting: %w", err)
		}
	}

	return nil
}

// GetSetting retrieves a setting value by key from the database
func (ss *SettingsService) GetSetting(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("setting key cannot be empty")
	}

	setting, err := ss.client.Settings.
		Query().
		Where(settings.Key(key)).
		Only(ss.ctx)

	if err != nil {
		// Return empty string if setting doesn't exist
		return "", nil
	}

	return setting.Value, nil
}

// DeleteSetting removes a setting from the database
func (ss *SettingsService) DeleteSetting(key string) error {
	if key == "" {
		return fmt.Errorf("setting key cannot be empty")
	}

	_, err := ss.client.Settings.
		Delete().
		Where(settings.Key(key)).
		Exec(ss.ctx)

	if err != nil {
		return fmt.Errorf("failed to delete setting: %w", err)
	}

	return nil
}

// SaveOpenAIApiKey saves the OpenAI API key securely
func (ss *SettingsService) SaveOpenAIApiKey(apiKey string) error {
	return ss.SaveSetting("openai_api_key", apiKey)
}

// GetOpenAIApiKey retrieves the OpenAI API key
func (ss *SettingsService) GetOpenAIApiKey() (string, error) {
	return ss.GetSetting("openai_api_key")
}

// DeleteOpenAIApiKey removes the OpenAI API key
func (ss *SettingsService) DeleteOpenAIApiKey() error {
	return ss.DeleteSetting("openai_api_key")
}

// SaveOpenRouterApiKey saves the OpenRouter API key securely
func (ss *SettingsService) SaveOpenRouterApiKey(apiKey string) error {
	return ss.SaveSetting("openrouter_api_key", apiKey)
}

// GetOpenRouterApiKey retrieves the OpenRouter API key
func (ss *SettingsService) GetOpenRouterApiKey() (string, error) {
	return ss.GetSetting("openrouter_api_key")
}

// DeleteOpenRouterApiKey removes the OpenRouter API key
func (ss *SettingsService) DeleteOpenRouterApiKey() error {
	return ss.DeleteSetting("openrouter_api_key")
}