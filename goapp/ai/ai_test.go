package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ramble-ai/goapp"
)

func TestNewApiKeyService(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewApiKeyService(helper.Client, helper.Ctx)
	
	assert.NotNil(t, service)
	assert.Equal(t, helper.Client, service.client)
	assert.Equal(t, helper.Ctx, service.ctx)
}

func TestGetSetting(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewApiKeyService(helper.Client, helper.Ctx)

	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{"normal setting", "test_key", "test_value", "test_value"},
		{"empty value", "empty_key", "", ""},
		{"special chars", "special_key", "value with !@# symbols", "value with !@# symbols"},
		{"unicode value", "unicode_key", "测试值 🔧", "测试值 🔧"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First create the setting
			helper.CreateTestSetting(tt.key, tt.value)

			// Test getSetting
			result, err := service.getSetting(tt.key)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSetting_EmptyKey(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewApiKeyService(helper.Client, helper.Ctx)

	result, err := service.getSetting("")
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "setting key cannot be empty")
}

func TestGetSetting_NotFound(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewApiKeyService(helper.Client, helper.Ctx)

	result, err := service.getSetting("nonexistent_key")
	assert.NoError(t, err) // Current implementation returns no error
	assert.Empty(t, result)
}

func TestGetOpenAIApiKey(t *testing.T) {
	t.Run("valid key", func(t *testing.T) {
		helper := goapp.NewTestHelper(t)
		service := NewApiKeyService(helper.Client, helper.Ctx)
		
		key := helper.MockOpenAIKey()
		helper.CreateTestSetting("openai_api_key", key)

		result, err := service.getOpenAIApiKey()
		assert.NoError(t, err)
		assert.Equal(t, key, result)
	})

	t.Run("test key", func(t *testing.T) {
		helper := goapp.NewTestHelper(t)
		service := NewApiKeyService(helper.Client, helper.Ctx)
		
		key := "sk-test123"
		helper.CreateTestSetting("openai_api_key", key)

		result, err := service.getOpenAIApiKey()
		assert.NoError(t, err)
		assert.Equal(t, key, result)
	})
}

func TestGetOpenAIApiKey_NotSet(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewApiKeyService(helper.Client, helper.Ctx)

	// Don't set any API key
	result, err := service.getOpenAIApiKey()
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetOpenRouterApiKey(t *testing.T) {
	t.Run("valid key", func(t *testing.T) {
		helper := goapp.NewTestHelper(t)
		service := NewApiKeyService(helper.Client, helper.Ctx)
		
		key := helper.MockOpenRouterKey()
		helper.CreateTestSetting("openrouter_api_key", key)

		result, err := service.getOpenRouterApiKey()
		assert.NoError(t, err)
		assert.Equal(t, key, result)
	})

	t.Run("test key", func(t *testing.T) {
		helper := goapp.NewTestHelper(t)
		service := NewApiKeyService(helper.Client, helper.Ctx)
		
		key := "sk-or-test123"
		helper.CreateTestSetting("openrouter_api_key", key)

		result, err := service.getOpenRouterApiKey()
		assert.NoError(t, err)
		assert.Equal(t, key, result)
	})
}

func TestGetOpenRouterApiKey_NotSet(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewApiKeyService(helper.Client, helper.Ctx)

	// Don't set any API key
	result, err := service.getOpenRouterApiKey()
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestApiKeyServiceIntegration(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewApiKeyService(helper.Client, helper.Ctx)

	// Set both API keys
	openaiKey := helper.MockOpenAIKey()
	openrouterKey := helper.MockOpenRouterKey()
	
	helper.CreateTestSetting("openai_api_key", openaiKey)
	helper.CreateTestSetting("openrouter_api_key", openrouterKey)

	// Test retrieving both keys
	retrievedOpenAI, err := service.getOpenAIApiKey()
	require.NoError(t, err)
	assert.Equal(t, openaiKey, retrievedOpenAI)

	retrievedOpenRouter, err := service.getOpenRouterApiKey()
	require.NoError(t, err)
	assert.Equal(t, openrouterKey, retrievedOpenRouter)

	// Test other settings don't interfere
	helper.CreateTestSetting("other_setting", "other_value")
	
	// Keys should still be retrievable
	retrievedOpenAI2, err := service.getOpenAIApiKey()
	require.NoError(t, err)
	assert.Equal(t, openaiKey, retrievedOpenAI2)

	retrievedOpenRouter2, err := service.getOpenRouterApiKey()
	require.NoError(t, err)
	assert.Equal(t, openrouterKey, retrievedOpenRouter2)
}

func TestApiKeyService_EdgeCases(t *testing.T) {
	t.Run("long API key", func(t *testing.T) {
		helper := goapp.NewTestHelper(t)
		service := NewApiKeyService(helper.Client, helper.Ctx)

		// Test very long API key
		longKey := "sk-" + string(make([]byte, 1000))
		for i := range longKey[3:] {
			longKey = longKey[:3+i] + "a" + longKey[3+i+1:]
		}
		
		helper.CreateTestSetting("openai_api_key", longKey)
		
		result, err := service.getOpenAIApiKey()
		require.NoError(t, err)
		assert.Equal(t, longKey, result)
	})

	t.Run("special characters", func(t *testing.T) {
		helper := goapp.NewTestHelper(t)
		service := NewApiKeyService(helper.Client, helper.Ctx)

		// Test API key with special characters
		specialKey := "sk-test!@#$%^&*()_+{}[]|\\:;\"'<>?,./"
		helper.CreateTestSetting("openrouter_api_key", specialKey)
		
		result, err := service.getOpenRouterApiKey()
		require.NoError(t, err)
		assert.Equal(t, specialKey, result)
	})

	t.Run("whitespace preservation", func(t *testing.T) {
		helper := goapp.NewTestHelper(t)
		service := NewApiKeyService(helper.Client, helper.Ctx)

		// Test key with whitespace
		whitespaceKey := "  sk-test-with-spaces  "
		helper.CreateTestSetting("openai_api_key", whitespaceKey)
		
		result, err := service.getOpenAIApiKey()
		require.NoError(t, err)
		assert.Equal(t, whitespaceKey, result) // Should preserve whitespace
	})
}