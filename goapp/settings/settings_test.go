package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ramble-ai/goapp"
)

func TestNewSettingsService(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewSettingsService(helper.Client, helper.Ctx)
	
	assert.NotNil(t, service)
	assert.Equal(t, helper.Client, service.client)
	assert.Equal(t, helper.Ctx, service.ctx)
}

func TestSaveSetting(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewSettingsService(helper.Client, helper.Ctx)

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"simple setting", "test_key", "test_value"},
		{"empty value", "empty_key", ""},
		{"special chars", "special_key", "value with spaces and symbols !@#"},
		{"long value", "long_key", "this is a very long value that might be stored as text in the database"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SaveSetting(tt.key, tt.value)
			require.NoError(t, err)
			
			// Verify setting was saved
			helper.AssertSettingEquals(tt.key, tt.value)
		})
	}
}

func TestSaveSetting_Update(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewSettingsService(helper.Client, helper.Ctx)

	key := "update_test_key"
	originalValue := "original_value"
	newValue := "updated_value"

	// Save original setting
	err := service.SaveSetting(key, originalValue)
	require.NoError(t, err)
	helper.AssertSettingEquals(key, originalValue)

	// Update setting
	err = service.SaveSetting(key, newValue)
	require.NoError(t, err)
	helper.AssertSettingEquals(key, newValue)
}

func TestGetSetting(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewSettingsService(helper.Client, helper.Ctx)

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"normal setting", "get_test_key", "get_test_value"},
		{"empty value", "get_empty_key", ""},
		{"unicode value", "get_unicode_key", "测试值 with émojis 🔧"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First save the setting
			err := service.SaveSetting(tt.key, tt.value)
			require.NoError(t, err)

			// Then get it
			value, err := service.GetSetting(tt.key)
			require.NoError(t, err)
			assert.Equal(t, tt.value, value)
		})
	}
}

func TestGetSetting_NotFound(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewSettingsService(helper.Client, helper.Ctx)

	value, err := service.GetSetting("nonexistent_key")
	// Based on the current implementation, GetSetting returns empty string and nil error for missing keys
	assert.NoError(t, err)
	assert.Empty(t, value)
}

func TestDeleteSetting(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewSettingsService(helper.Client, helper.Ctx)

	key := "delete_test_key"
	value := "delete_test_value"

	// Save setting
	err := service.SaveSetting(key, value)
	require.NoError(t, err)
	helper.AssertSettingEquals(key, value)

	// Delete setting
	err = service.DeleteSetting(key)
	require.NoError(t, err)

	// Verify setting is deleted - GetSetting returns empty string for missing keys
	deletedValue, err := service.GetSetting(key)
	assert.NoError(t, err)
	assert.Empty(t, deletedValue)
}

func TestDeleteSetting_NotFound(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewSettingsService(helper.Client, helper.Ctx)

	// DeleteSetting doesn't explicitly check if key exists, so it should succeed
	err := service.DeleteSetting("nonexistent_key")
	assert.NoError(t, err)
}

func TestSaveThemePreference(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewSettingsService(helper.Client, helper.Ctx)

	tests := []struct {
		name      string
		theme     string
		expectErr bool
	}{
		{"light theme", "light", false},
		{"dark theme", "dark", false},
		{"invalid theme", "system", true}, // Only light/dark allowed per current implementation
		{"invalid custom", "custom-blue", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SaveThemePreference(tt.theme)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				// Verify theme was saved with correct key
				helper.AssertSettingEquals("theme_preference", tt.theme)
			}
		})
	}
}

func TestGetThemePreference(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewSettingsService(helper.Client, helper.Ctx)

	tests := []struct {
		name          string
		theme         string
		expectedTheme string
	}{
		{"light theme", "light", "light"},
		{"dark theme", "dark", "dark"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save theme preference
			err := service.SaveThemePreference(tt.theme)
			require.NoError(t, err)

			// Get theme preference
			theme, err := service.GetThemePreference()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedTheme, theme)
		})
	}
}

func TestGetThemePreference_NotSet(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewSettingsService(helper.Client, helper.Ctx)

	// Get theme preference when not set - should return default
	theme, err := service.GetThemePreference()
	require.NoError(t, err)
	assert.Equal(t, "light", theme) // Default theme per implementation
}

func TestSettingsIntegration(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewSettingsService(helper.Client, helper.Ctx)

	// Test saving multiple settings
	settings := map[string]string{
		"api_key":      "sk-test123",
		"max_retries":  "5",
		"timeout":      "30",
		"debug_mode":   "true",
		"environment":  "test",
	}

	// Save all settings
	for key, value := range settings {
		err := service.SaveSetting(key, value)
		require.NoError(t, err)
	}

	// Verify all settings
	for key, expectedValue := range settings {
		value, err := service.GetSetting(key)
		require.NoError(t, err)
		assert.Equal(t, expectedValue, value)
	}

	// Update one setting
	err := service.SaveSetting("max_retries", "10")
	require.NoError(t, err)

	value, err := service.GetSetting("max_retries")
	require.NoError(t, err)
	assert.Equal(t, "10", value)

	// Delete one setting
	err = service.DeleteSetting("debug_mode")
	require.NoError(t, err)

	deletedValue, err := service.GetSetting("debug_mode")
	assert.NoError(t, err)
	assert.Empty(t, deletedValue)

	// Verify other settings still exist
	for key, expectedValue := range settings {
		if key == "debug_mode" {
			continue
		}
		if key == "max_retries" {
			expectedValue = "10" // Updated value
		}
		
		value, err := service.GetSetting(key)
		require.NoError(t, err)
		assert.Equal(t, expectedValue, value)
	}
}