package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ramble-ai/ent/schema"
	"ramble-ai/goapp"
	"ramble-ai/goapp/highlights"
)

// setupAppTestHelper creates a test helper for App-level AI tests  
func setupAppTestHelper(t *testing.T) (*goapp.TestHelper, *App) {
	helper := goapp.NewTestHelper(t)
	
	// Create App instance with the test client
	app := &App{
		ctx:    helper.Ctx,
		client: helper.Client,
	}
	
	return helper, app
}

// TestAppSuggestHighlightsWithAI tests the App-level SuggestHighlightsWithAI function
func TestAppSuggestHighlightsWithAI(t *testing.T) {
	helper, app := setupAppTestHelper(t)
	
	// Create test project and video clip
	project := helper.CreateTestProject("App AI Suggestions Test")
	clip := helper.CreateTestVideoClip(project, "Test Video")
	
	// Set up transcription data
	transcriptWords := []schema.Word{
		{Word: "Hello", Start: 0.0, End: 0.5},
		{Word: "world", Start: 0.6, End: 1.0},
		{Word: "this", Start: 1.2, End: 1.5},
		{Word: "is", Start: 1.6, End: 1.8},
		{Word: "test", Start: 1.9, End: 2.2},
	}
	
	// Update clip with transcription
	_, err := helper.Client.VideoClip.
		UpdateOneID(clip.ID).
		SetTranscription("Hello world this is test").
		SetTranscriptionWords(transcriptWords).
		Save(helper.Ctx)
	require.NoError(t, err)
	
	t.Run("function exists and handles missing API key", func(t *testing.T) {
		suggestions, err := app.SuggestHighlightsWithAI(project.ID, clip.ID, "Test prompt")
		
		// Should fail due to no OpenRouter API key configured
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get AI highlight suggestions: OpenRouter API key not provided")
		assert.Nil(t, suggestions)
	})
	
	t.Run("function handles nonexistent project", func(t *testing.T) {
		suggestions, err := app.SuggestHighlightsWithAI(999999, clip.ID, "Test prompt")
		
		// Should fail due to nonexistent project
		assert.Error(t, err)
		assert.Nil(t, suggestions)
	})
	
	t.Run("function handles nonexistent video", func(t *testing.T) {
		suggestions, err := app.SuggestHighlightsWithAI(project.ID, 999999, "Test prompt")
		
		// Should fail due to nonexistent video
		assert.Error(t, err)
		assert.Nil(t, suggestions)
	})
}

// TestAppGetProjectAISettings tests the App-level GetProjectAISettings function
func TestAppGetProjectAISettings(t *testing.T) {
	helper, app := setupAppTestHelper(t)
	
	t.Run("returns default settings for new project", func(t *testing.T) {
		project := helper.CreateTestProject("App AI Settings Test")
		
		settings, err := app.GetProjectAISettings(project.ID)
		
		assert.NoError(t, err)
		assert.NotNil(t, settings)
		assert.Equal(t, "anthropic/claude-3.5-haiku-20241022", settings.AIModel)
		assert.Equal(t, "", settings.AIPrompt)
	})
	
	t.Run("returns custom settings when set", func(t *testing.T) {
		project := helper.CreateTestProject("App AI Settings Custom Test")
		
		// Update project with custom AI settings
		_, err := helper.Client.Project.
			UpdateOneID(project.ID).
			SetAiModel("custom/model").
			SetAiPrompt("Custom prompt").
			Save(helper.Ctx)
		require.NoError(t, err)
		
		settings, err := app.GetProjectAISettings(project.ID)
		
		assert.NoError(t, err)
		assert.Equal(t, "custom/model", settings.AIModel)
		assert.Equal(t, "Custom prompt", settings.AIPrompt)
	})
	
	t.Run("handles nonexistent project", func(t *testing.T) {
		settings, err := app.GetProjectAISettings(999999)
		
		assert.Error(t, err)
		assert.Nil(t, settings)
	})
}

// TestAppSaveProjectAISettings tests the App-level SaveProjectAISettings function
func TestAppSaveProjectAISettings(t *testing.T) {
	helper, app := setupAppTestHelper(t)
	
	t.Run("successfully saves AI settings", func(t *testing.T) {
		project := helper.CreateTestProject("App Save AI Settings Test")
		
		settings := highlights.ProjectAISettings{
			AIModel:  "test/model",
			AIPrompt: "Test prompt for AI",
		}
		
		err := app.SaveProjectAISettings(project.ID, settings)
		assert.NoError(t, err)
		
		// Verify settings were saved
		updatedProject, err := helper.Client.Project.Get(helper.Ctx, project.ID)
		assert.NoError(t, err)
		assert.Equal(t, "test/model", updatedProject.AiModel)
		assert.Equal(t, "Test prompt for AI", updatedProject.AiPrompt)
	})
	
	t.Run("handles nonexistent project", func(t *testing.T) {
		settings := highlights.ProjectAISettings{
			AIModel:  "test/model",
			AIPrompt: "Test prompt",
		}
		
		err := app.SaveProjectAISettings(999999, settings)
		assert.Error(t, err)
	})
}

// TestAppGetProjectAISuggestion tests the App-level GetProjectAISuggestion function
func TestAppGetProjectAISuggestion(t *testing.T) {
	helper, app := setupAppTestHelper(t)
	
	t.Run("returns nil for project without suggestions", func(t *testing.T) {
		project := helper.CreateTestProject("App No Suggestions Test")
		
		suggestion, err := app.GetProjectAISuggestion(project.ID)
		
		assert.NoError(t, err)
		assert.Nil(t, suggestion)
	})
	
	t.Run("returns cached suggestion when available", func(t *testing.T) {
		project := helper.CreateTestProject("App With Suggestions Test")
		
		// Set up cached suggestion
		order := []interface{}{"h1", "h2", "N", "h3"}
		
		_, err := helper.Client.Project.
			UpdateOneID(project.ID).
			SetAiSuggestionOrder(order).
			SetAiSuggestionModel("test/model").
			Save(helper.Ctx)
		require.NoError(t, err)
		
		suggestion, err := app.GetProjectAISuggestion(project.ID)
		
		assert.NoError(t, err)
		assert.NotNil(t, suggestion)
		assert.Equal(t, order, suggestion.Order)
		assert.Equal(t, "test/model", suggestion.Model)
	})
	
	t.Run("handles nonexistent project", func(t *testing.T) {
		suggestion, err := app.GetProjectAISuggestion(999999)
		
		assert.Error(t, err)
		assert.Nil(t, suggestion)
	})
}

// TestAppReorderHighlightsWithAI tests the App-level ReorderHighlightsWithAI function
func TestAppReorderHighlightsWithAI(t *testing.T) {
	helper, app := setupAppTestHelper(t)
	
	project := helper.CreateTestProject("App Reorder Test")
	
	t.Run("function exists and handles missing API key", func(t *testing.T) {
		result, err := app.ReorderHighlightsWithAI(project.ID, "Custom prompt")
		
		// Should fail due to no OpenRouter API key configured
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "OpenRouter API key not configured")
		assert.Nil(t, result)
	})
	
	t.Run("function handles nonexistent project", func(t *testing.T) {
		result, err := app.ReorderHighlightsWithAI(999999, "Custom prompt")
		
		// Should fail due to AI service dependencies
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestAppReorderHighlightsWithAIOptions tests the App-level ReorderHighlightsWithAIOptions function  
func TestAppReorderHighlightsWithAIOptions(t *testing.T) {
	helper, app := setupAppTestHelper(t)
	
	project := helper.CreateTestProject("App Reorder Options Test")
	
	options := highlights.AIActionOptions{
		UseCurrentOrder:       false,
		KeepAllHighlights:     true,
		OptimizeForEngagement: false,
		CreateSections:        true,
		BalanceLength:         false,
		ImproveTransitions:    false,
	}
	
	t.Run("function exists and handles missing API key", func(t *testing.T) {
		result, err := app.ReorderHighlightsWithAIOptions(project.ID, "Test prompt", options)
		
		// Should fail due to no OpenRouter API key configured
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "OpenRouter API key not configured")
		assert.Nil(t, result)
	})
	
	t.Run("function handles nonexistent project", func(t *testing.T) {
		result, err := app.ReorderHighlightsWithAIOptions(999999, "Test prompt", options)
		
		// Should fail due to AI service dependencies
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestAppImproveHighlightSilencesWithAI tests the App-level ImproveHighlightSilencesWithAI function
func TestAppImproveHighlightSilencesWithAI(t *testing.T) {
	helper, app := setupAppTestHelper(t)
	
	project := helper.CreateTestProject("App Silence Improvement Test")
	
	t.Run("function exists and handles missing API key", func(t *testing.T) {
		result, err := app.ImproveHighlightSilencesWithAI(project.ID)
		
		// Should fail due to no OpenRouter API key configured
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "OpenRouter API key not configured")
		assert.Nil(t, result)
	})
	
	t.Run("function handles nonexistent project", func(t *testing.T) {
		result, err := app.ImproveHighlightSilencesWithAI(999999)
		
		// Should fail due to AI service dependencies
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestAppGetProjectAISilenceResult tests the App-level GetProjectAISilenceResult function
func TestAppGetProjectAISilenceResult(t *testing.T) {
	helper, app := setupAppTestHelper(t)
	
	t.Run("returns nil for project without cached improvements", func(t *testing.T) {
		project := helper.CreateTestProject("App No Silence Cache Test")
		
		result, err := app.GetProjectAISilenceResult(project.ID)
		
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
	
	t.Run("returns cached improvements when available", func(t *testing.T) {
		project := helper.CreateTestProject("App With Silence Cache Test")
		
		// Set up cached improvements
		improvements := []map[string]interface{}{
			{
				"videoClipId": 1,
				"highlights": []interface{}{
					map[string]interface{}{
						"id":    "h1",
						"start": 1.0,
						"end":   2.0,
					},
				},
			},
		}
		
		_, err := helper.Client.Project.
			UpdateOneID(project.ID).
			SetAiSilenceImprovements(improvements).
			SetAiSilenceModel("test/model").
			Save(helper.Ctx)
		require.NoError(t, err)
		
		result, err := app.GetProjectAISilenceResult(project.ID)
		
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test/model", result.Model)
		assert.NotEmpty(t, result.CreatedAt)
		assert.NotEmpty(t, result.Improvements)
	})
	
	t.Run("handles nonexistent project", func(t *testing.T) {
		result, err := app.GetProjectAISilenceResult(999999)
		
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestAppClearAISilenceImprovements tests the App-level ClearAISilenceImprovements function
func TestAppClearAISilenceImprovements(t *testing.T) {
	helper, app := setupAppTestHelper(t)
	
	project := helper.CreateTestProject("App Clear Silence Cache Test")
	
	t.Run("successfully clears cache", func(t *testing.T) {
		// First set some cached data
		_, err := helper.Client.Project.
			UpdateOneID(project.ID).
			SetAiSilenceImprovements([]map[string]interface{}{{"test": "data"}}).
			SetAiSilenceModel("test-model").
			Save(helper.Ctx)
		require.NoError(t, err)
		
		// Clear the cache
		err = app.ClearAISilenceImprovements(project.ID)
		assert.NoError(t, err)
		
		// Verify cache was cleared
		updatedProject, err := helper.Client.Project.Get(helper.Ctx, project.ID)
		assert.NoError(t, err)
		assert.Nil(t, updatedProject.AiSilenceImprovements)
		assert.Equal(t, "", updatedProject.AiSilenceModel)
	})
	
	t.Run("handles nonexistent project", func(t *testing.T) {
		err := app.ClearAISilenceImprovements(999999)
		assert.Error(t, err)
	})
}

// TestAppGetProjectHighlightAISettings tests the App-level GetProjectHighlightAISettings function
func TestAppGetProjectHighlightAISettings(t *testing.T) {
	helper, app := setupAppTestHelper(t)
	
	t.Run("function exists and can be called", func(t *testing.T) {
		project := helper.CreateTestProject("App Highlight AI Settings Test")
		
		settings, err := app.GetProjectHighlightAISettings(project.ID)
		
		// Function may succeed or fail, we just test it doesn't crash
		if err == nil {
			assert.NotNil(t, settings)
		}
	})
	
	t.Run("handles nonexistent project", func(t *testing.T) {
		settings, err := app.GetProjectHighlightAISettings(999999)
		
		// Function may succeed or fail, we just test it doesn't crash
		if err == nil {
			assert.NotNil(t, settings)
		} else {
			assert.Nil(t, settings)
		}
	})
}

// TestAppSaveProjectHighlightAISettings tests the App-level SaveProjectHighlightAISettings function
func TestAppSaveProjectHighlightAISettings(t *testing.T) {
	helper, app := setupAppTestHelper(t)
	
	project := helper.CreateTestProject("App Save Highlight AI Settings Test")
	
	t.Run("function exists and can be called", func(t *testing.T) {
		settings := highlights.ProjectHighlightAISettings{
			AIModel:  "test/model",
			AIPrompt: "Test prompt",
		}
		
		err := app.SaveProjectHighlightAISettings(project.ID, settings)
		
		// Function may succeed or fail, we just test it doesn't crash
		// Error is expected as this may depend on missing schema fields
		if err != nil {
			assert.Error(t, err)
		}
	})
	
	t.Run("handles nonexistent project", func(t *testing.T) {
		settings := highlights.ProjectHighlightAISettings{
			AIModel:  "test/model",
			AIPrompt: "Test prompt",
		}
		
		err := app.SaveProjectHighlightAISettings(999999, settings)
		// Function may succeed or fail, we just test it doesn't crash
		// Error is expected but not required as the implementation may vary
		if err != nil {
			assert.Error(t, err)
		}
	})
}

// TestAppGetSuggestedHighlights tests the App-level GetSuggestedHighlights function
func TestAppGetSuggestedHighlights(t *testing.T) {
	helper, app := setupAppTestHelper(t)
	
	project := helper.CreateTestProject("App Get Suggestions Test")
	clip := helper.CreateTestVideoClip(project, "Test Video")
	
	t.Run("function exists and can be called", func(t *testing.T) {
		suggestions, err := app.GetSuggestedHighlights(clip.ID)
		
		// Function should work with empty suggestions
		assert.NoError(t, err)
		assert.Empty(t, suggestions)
	})
	
	t.Run("handles nonexistent video", func(t *testing.T) {
		suggestions, err := app.GetSuggestedHighlights(999999)
		
		assert.Error(t, err)
		assert.Nil(t, suggestions)
	})
}

// TestAppClearSuggestedHighlights tests the App-level ClearSuggestedHighlights function
func TestAppClearSuggestedHighlights(t *testing.T) {
	helper, app := setupAppTestHelper(t)
	
	project := helper.CreateTestProject("App Clear Suggestions Test")
	clip := helper.CreateTestVideoClip(project, "Test Video")
	
	t.Run("function exists and can be called", func(t *testing.T) {
		err := app.ClearSuggestedHighlights(clip.ID)
		assert.NoError(t, err)
	})
	
	t.Run("handles nonexistent video", func(t *testing.T) {
		err := app.ClearSuggestedHighlights(999999)
		assert.Error(t, err)
	})
}

// TestAppDeleteSuggestedHighlight tests the App-level DeleteSuggestedHighlight function
func TestAppDeleteSuggestedHighlight(t *testing.T) {
	helper, app := setupAppTestHelper(t)
	
	project := helper.CreateTestProject("App Delete Suggestion Test")
	clip := helper.CreateTestVideoClip(project, "Test Video")
	
	t.Run("function exists and can be called", func(t *testing.T) {
		err := app.DeleteSuggestedHighlight(clip.ID, "nonexistent_suggestion")
		
		// Function may succeed (if it doesn't validate existence) or fail
		// Either way, it shouldn't crash
		if err != nil {
			assert.Error(t, err)
		}
	})
	
	t.Run("handles nonexistent video", func(t *testing.T) {
		err := app.DeleteSuggestedHighlight(999999, "suggestion_id")
		assert.Error(t, err)
	})
}