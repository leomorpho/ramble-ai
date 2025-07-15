package settings

import (
	"context"
	"fmt"

	"MYAPP/ent"
	"MYAPP/ent/settings"
)

// SettingsService provides settings management functionality
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
func (s *SettingsService) SaveSetting(key, value string) error {
	if key == "" {
		return fmt.Errorf("setting key cannot be empty")
	}

	// Check if setting already exists
	existingSetting, err := s.client.Settings.
		Query().
		Where(settings.Key(key)).
		Only(s.ctx)

	if err != nil {
		// Setting doesn't exist, create new one
		_, err = s.client.Settings.
			Create().
			SetKey(key).
			SetValue(value).
			Save(s.ctx)

		if err != nil {
			return fmt.Errorf("failed to create setting: %w", err)
		}
	} else {
		// Setting exists, update it
		_, err = s.client.Settings.
			UpdateOne(existingSetting).
			SetValue(value).
			Save(s.ctx)

		if err != nil {
			return fmt.Errorf("failed to update setting: %w", err)
		}
	}

	return nil
}

// GetSetting retrieves a setting value by key from the database
func (s *SettingsService) GetSetting(key string) (string, error) {
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

// DeleteSetting removes a setting from the database
func (s *SettingsService) DeleteSetting(key string) error {
	if key == "" {
		return fmt.Errorf("setting key cannot be empty")
	}

	_, err := s.client.Settings.
		Delete().
		Where(settings.Key(key)).
		Exec(s.ctx)

	if err != nil {
		return fmt.Errorf("failed to delete setting: %w", err)
	}

	return nil
}

// SaveThemePreference saves the user's preferred theme (light or dark)
func (s *SettingsService) SaveThemePreference(theme string) error {
	if theme != "light" && theme != "dark" {
		return fmt.Errorf("theme must be either 'light' or 'dark'")
	}
	return s.SaveSetting("theme_preference", theme)
}

// GetThemePreference retrieves the user's preferred theme, defaults to "light"
func (s *SettingsService) GetThemePreference() (string, error) {
	theme, err := s.GetSetting("theme_preference")
	if err != nil {
		return "light", err
	}
	if theme == "" {
		return "light", nil // Default to light theme
	}
	return theme, nil
}
