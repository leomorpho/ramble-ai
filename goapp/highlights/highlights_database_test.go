package highlights

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ramble-ai/ent/schema"
	"ramble-ai/goapp"
)

func TestDeleteSuggestedHighlight(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewHighlightService(helper.Client, helper.Ctx)

	// Create a test project and video clip
	project := helper.CreateTestProject("Test Project")
	videoClip := helper.CreateTestVideoClip(project, "test_video.mp4")

	// Create some suggested highlights
	suggestedHighlights := []schema.Highlight{
		{ID: "suggestion_1", Start: 0.0, End: 2.0, ColorID: 1},
		{ID: "suggestion_2", Start: 3.0, End: 5.0, ColorID: 2},
		{ID: "suggestion_3", Start: 6.0, End: 8.0, ColorID: 3},
	}

	// Update video clip with suggested highlights
	_, err := helper.Client.VideoClip.
		UpdateOneID(videoClip.ID).
		SetSuggestedHighlights(suggestedHighlights).
		Save(helper.Ctx)
	require.NoError(t, err)

	t.Run("delete existing suggested highlight", func(t *testing.T) {
		err := service.DeleteSuggestedHighlight(videoClip.ID, "suggestion_2")
		require.NoError(t, err)

		// Verify the highlight was deleted
		updatedClip, err := helper.Client.VideoClip.Get(helper.Ctx, videoClip.ID)
		require.NoError(t, err)
		
		// Should have 2 highlights left
		assert.Len(t, updatedClip.SuggestedHighlights, 2)
		
		// Verify the right highlight was deleted
		highlightIDs := make([]string, len(updatedClip.SuggestedHighlights))
		for i, h := range updatedClip.SuggestedHighlights {
			highlightIDs[i] = h.ID
		}
		assert.Contains(t, highlightIDs, "suggestion_1")
		assert.Contains(t, highlightIDs, "suggestion_3")
		assert.NotContains(t, highlightIDs, "suggestion_2")
	})

	t.Run("delete non-existent suggested highlight", func(t *testing.T) {
		err := service.DeleteSuggestedHighlight(videoClip.ID, "non_existent")
		require.NoError(t, err) // Should not error, just no-op

		// Verify no highlights were removed
		updatedClip, err := helper.Client.VideoClip.Get(helper.Ctx, videoClip.ID)
		require.NoError(t, err)
		assert.Len(t, updatedClip.SuggestedHighlights, 2)
	})

	t.Run("delete from non-existent video clip", func(t *testing.T) {
		err := service.DeleteSuggestedHighlight(99999, "suggestion_1")
		assert.Error(t, err)
	})
}

func TestGetProjectHighlightOrderWithTitles(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewHighlightService(helper.Client, helper.Ctx)

	// Create a test project
	project := helper.CreateTestProject("Test Project")

	t.Run("project with no highlight order", func(t *testing.T) {
		result, err := service.GetProjectHighlightOrderWithTitles(project.ID)
		require.NoError(t, err)
		
		// Should return empty slice when no order is set
		if result != nil {
			assert.Len(t, result, 0)
		}
	})

	t.Run("project with mixed highlight order and titles", func(t *testing.T) {
		// First, set a highlight order with mixed content
		highlightOrder := []interface{}{
			"highlight_1",
			map[string]interface{}{
				"title": "Section 1",
				"type":  "N",
			},
			"highlight_2",
			"highlight_3",
		}

		// Save the order using the project service (we need to create this first)
		_, err := helper.Client.Project.
			UpdateOneID(project.ID).
			SetHighlightOrder(highlightOrder).
			Save(helper.Ctx)
		require.NoError(t, err)

		// Now test getting the order
		result, err := service.GetProjectHighlightOrderWithTitles(project.ID)
		require.NoError(t, err)
		assert.Len(t, result, 4)

		// Verify the structure
		assert.Equal(t, "highlight_1", result[0])
		
		// Second item should be the section object
		sectionObj, ok := result[1].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Section 1", sectionObj["title"])
		assert.Equal(t, "N", sectionObj["type"])
		
		assert.Equal(t, "highlight_2", result[2])
		assert.Equal(t, "highlight_3", result[3])
	})

	t.Run("non-existent project", func(t *testing.T) {
		result, err := service.GetProjectHighlightOrderWithTitles(99999)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGetSuggestedHighlights(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewHighlightService(helper.Client, helper.Ctx)

	// Create a test project and video clip
	project := helper.CreateTestProject("Test Project")
	videoClip := helper.CreateTestVideoClip(project, "test_video.mp4")

	t.Run("video clip with no suggested highlights", func(t *testing.T) {
		result, err := service.GetSuggestedHighlights(videoClip.ID)
		require.NoError(t, err)
		
		// Should return empty slice (or nil is also acceptable)
		if result != nil {
			assert.Len(t, result, 0)
		}
	})

	t.Run("video clip with suggested highlights", func(t *testing.T) {
		// Create suggested highlights with transcription words
		words := []schema.Word{
			{Word: "Hello", Start: 0.0, End: 0.5},
			{Word: "world", Start: 0.5, End: 1.0},
			{Word: "this", Start: 1.0, End: 1.5},
			{Word: "is", Start: 1.5, End: 2.0},
			{Word: "a", Start: 2.0, End: 2.5},
			{Word: "test", Start: 2.5, End: 3.0},
		}

		suggestedHighlights := []schema.Highlight{
			{ID: "suggestion_1", Start: 0.0, End: 1.0, ColorID: 1},
			{ID: "suggestion_2", Start: 1.5, End: 3.0, ColorID: 2},
		}

		// Update video clip with suggested highlights and words
		_, err := helper.Client.VideoClip.
			UpdateOneID(videoClip.ID).
			SetSuggestedHighlights(suggestedHighlights).
			SetTranscriptionWords(words).
			SetTranscription("Hello world this is a test").
			Save(helper.Ctx)
		require.NoError(t, err)

		// Get suggested highlights
		result, err := service.GetSuggestedHighlights(videoClip.ID)
		require.NoError(t, err)
		assert.Len(t, result, 2)

		// Verify the structure
		assert.Equal(t, "suggestion_1", result[0].ID)
		assert.Equal(t, 1, result[0].ColorID)
		assert.Equal(t, "Hello world", result[0].Text) // Should extract text from words

		assert.Equal(t, "suggestion_2", result[1].ID)
		assert.Equal(t, 2, result[1].ColorID)
		assert.Equal(t, "is a test", result[1].Text)
	})

	t.Run("non-existent video clip", func(t *testing.T) {
		result, err := service.GetSuggestedHighlights(99999)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGetProjectHighlights(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewHighlightService(helper.Client, helper.Ctx)

	// Create a test project
	project := helper.CreateTestProject("Test Project")

	t.Run("project with no video clips", func(t *testing.T) {
		result, err := service.GetProjectHighlights(project.ID)
		require.NoError(t, err)
		
		// Should return empty slice (or nil is also acceptable)
		if result != nil {
			assert.Len(t, result, 0)
		}
	})

	t.Run("project with video clips and highlights", func(t *testing.T) {
		// Create video clips with highlights
		videoClip1 := helper.CreateTestVideoClip(project, "video1.mp4")
		videoClip2 := helper.CreateTestVideoClip(project, "video2.mp4")

		// Add transcription words and highlights
		words1 := []schema.Word{
			{Word: "First", Start: 0.0, End: 0.5},
			{Word: "video", Start: 0.5, End: 1.0},
			{Word: "content", Start: 1.0, End: 1.5},
		}
		highlights1 := []schema.Highlight{
			{ID: "h1", Start: 0.0, End: 1.0, ColorID: 1},
			{ID: "h2", Start: 1.0, End: 1.5, ColorID: 2},
		}

		words2 := []schema.Word{
			{Word: "Second", Start: 0.0, End: 0.5},
			{Word: "video", Start: 0.5, End: 1.0},
		}
		highlights2 := []schema.Highlight{
			{ID: "h3", Start: 0.0, End: 1.0, ColorID: 3},
		}

		// Update video clips
		_, err := helper.Client.VideoClip.
			UpdateOneID(videoClip1.ID).
			SetName("Video 1").
			SetTranscriptionWords(words1).
			SetTranscription("First video content").
			SetHighlights(highlights1).
			SetDuration(5.0).
			Save(helper.Ctx)
		require.NoError(t, err)

		_, err = helper.Client.VideoClip.
			UpdateOneID(videoClip2.ID).
			SetName("Video 2").
			SetTranscriptionWords(words2).
			SetTranscription("Second video").
			SetHighlights(highlights2).
			SetDuration(3.0).
			Save(helper.Ctx)
		require.NoError(t, err)

		// Get project highlights
		result, err := service.GetProjectHighlights(project.ID)
		require.NoError(t, err)
		assert.Len(t, result, 2) // Should have 2 video clips

		// Verify video clips are present
		clipNames := make([]string, len(result))
		for i, clip := range result {
			clipNames[i] = clip.VideoClipName
		}
		assert.Contains(t, clipNames, "Video 1")
		assert.Contains(t, clipNames, "Video 2")

		// Find Video 1 and verify its highlights
		var video1 *ProjectHighlight
		for i := range result {
			if result[i].VideoClipName == "Video 1" {
				video1 = &result[i]
				break
			}
		}
		require.NotNil(t, video1)
		assert.Len(t, video1.Highlights, 2)
		assert.Equal(t, "First video", video1.Highlights[0].Text)
		assert.Equal(t, "content", video1.Highlights[1].Text)
	})

	t.Run("non-existent project", func(t *testing.T) {
		result, err := service.GetProjectHighlights(99999)
		require.NoError(t, err)
		
		// Should return empty slice for non-existent project (or nil is also acceptable)
		if result != nil {
			assert.Len(t, result, 0)
		}
	})
}

func TestGetProjectHighlightsForExport(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewHighlightService(helper.Client, helper.Ctx)

	// Create a test project
	project := helper.CreateTestProject("Test Project")

	t.Run("project with no highlights", func(t *testing.T) {
		result, err := service.GetProjectHighlightsForExport(project.ID)
		require.NoError(t, err)
		
		// Should return empty slice (or nil is also acceptable)
		if result != nil {
			assert.Len(t, result, 0)
		}
	})

	t.Run("project with highlights for export", func(t *testing.T) {
		// Create video clip with highlights
		videoClip := helper.CreateTestVideoClip(project, "export_video.mp4")

		words := []schema.Word{
			{Word: "Export", Start: 0.0, End: 0.5},
			{Word: "this", Start: 0.5, End: 1.0},
			{Word: "highlight", Start: 1.0, End: 1.5},
			{Word: "content", Start: 1.5, End: 2.0},
		}
		highlights := []schema.Highlight{
			{ID: "export_h1", Start: 0.0, End: 1.5, ColorID: 1},
			{ID: "export_h2", Start: 1.5, End: 2.0, ColorID: 2},
		}

		// Update video clip
		_, err := helper.Client.VideoClip.
			UpdateOneID(videoClip.ID).
			SetName("Export Video").
			SetFilePath("/test/export_video.mp4").
			SetTranscriptionWords(words).
			SetTranscription("Export this highlight content").
			SetHighlights(highlights).
			Save(helper.Ctx)
		require.NoError(t, err)

		// Get highlights for export
		result, err := service.GetProjectHighlightsForExport(project.ID)
		require.NoError(t, err)
		assert.Len(t, result, 2) // Should have 2 highlight segments

		// Verify segments
		assert.Equal(t, "export_h1", result[0].ID)
		assert.Equal(t, "/test/export_video.mp4", result[0].VideoPath)
		assert.Equal(t, "Export Video", result[0].VideoClipName)
		assert.Equal(t, videoClip.ID, result[0].VideoClipID)
		assert.Equal(t, "Export this highlight", result[0].Text)

		assert.Equal(t, "export_h2", result[1].ID)
		assert.Equal(t, "content", result[1].Text)
	})

	t.Run("non-existent project", func(t *testing.T) {
		result, err := service.GetProjectHighlightsForExport(99999)
		require.NoError(t, err)
		
		// Should return empty slice for non-existent project (or nil is also acceptable)
		if result != nil {
			assert.Len(t, result, 0)
		}
	})
}

func TestGetProjectHighlightAISettings(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewHighlightService(helper.Client, helper.Ctx)

	// Create a test project
	project := helper.CreateTestProject("Test Project")

	t.Run("project with default AI settings", func(t *testing.T) {
		result, err := service.GetProjectHighlightAISettings(project.ID)
		require.NoError(t, err)
		
		// Should return empty strings when no settings exist (based on current implementation)
		assert.NotNil(t, result)
		assert.Equal(t, "", result.AIModel)  // Current behavior returns empty string
		assert.Equal(t, "", result.AIPrompt) // Current behavior returns empty string
	})

	t.Run("project with custom AI settings", func(t *testing.T) {
		// Create custom AI settings
		customSettings := ProjectHighlightAISettings{
			AIModel:  "openai/gpt-4",
			AIPrompt: "Find the most engaging moments in this video transcript.",
		}

		// Save the settings
		err := service.SaveProjectHighlightAISettings(project.ID, customSettings)
		require.NoError(t, err)

		// Retrieve the settings
		result, err := service.GetProjectHighlightAISettings(project.ID)
		require.NoError(t, err)
		
		assert.Equal(t, "openai/gpt-4", result.AIModel)
		assert.Equal(t, "Find the most engaging moments in this video transcript.", result.AIPrompt)
	})

	t.Run("non-existent project", func(t *testing.T) {
		result, err := service.GetProjectHighlightAISettings(99999)
		require.NoError(t, err)
		
		// Should return empty strings even for non-existent project (based on current implementation)
		assert.NotNil(t, result)
		assert.Equal(t, "", result.AIModel)
	})
}

func TestSaveProjectHighlightAISettings(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewHighlightService(helper.Client, helper.Ctx)

	// Create a test project
	project := helper.CreateTestProject("Test Project")

	t.Run("save valid AI settings", func(t *testing.T) {
		settings := ProjectHighlightAISettings{
			AIModel:  "openai/gpt-4",
			AIPrompt: "Custom AI prompt for highlights",
		}

		err := service.SaveProjectHighlightAISettings(project.ID, settings)
		require.NoError(t, err)

		// Verify settings were saved by retrieving them
		result, err := service.GetProjectHighlightAISettings(project.ID)
		require.NoError(t, err)
		assert.Equal(t, settings.AIModel, result.AIModel)
		assert.Equal(t, settings.AIPrompt, result.AIPrompt)
	})

	t.Run("update existing AI settings", func(t *testing.T) {
		updatedSettings := ProjectHighlightAISettings{
			AIModel:  "openai/gpt-3.5-turbo",
			AIPrompt: "Updated prompt for better results",
		}

		err := service.SaveProjectHighlightAISettings(project.ID, updatedSettings)
		require.NoError(t, err)

		// Verify settings were updated
		result, err := service.GetProjectHighlightAISettings(project.ID)
		require.NoError(t, err)
		assert.Equal(t, updatedSettings.AIModel, result.AIModel)
		assert.Equal(t, updatedSettings.AIPrompt, result.AIPrompt)
	})

	t.Run("non-existent project", func(t *testing.T) {
		settings := ProjectHighlightAISettings{
			AIModel:  "openai/gpt-4",
			AIPrompt: "Test prompt",
		}

		// Should not error for non-existent project (settings stored by key)
		err := service.SaveProjectHighlightAISettings(99999, settings)
		require.NoError(t, err)
	})
}

func TestGetProjectHighlightOrder(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewHighlightService(helper.Client, helper.Ctx)

	// Create a test project
	project := helper.CreateTestProject("Test Project")

	t.Run("project with no highlight order", func(t *testing.T) {
		result, err := service.GetProjectHighlightOrder(project.ID)
		require.NoError(t, err)
		
		// Should return empty slice when no order is set
		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("project with highlight order", func(t *testing.T) {
		// Set a highlight order (mix of strings and objects)
		highlightOrder := []interface{}{
			"highlight_1",
			"highlight_2", 
			map[string]interface{}{"title": "Section", "type": "N"},
			"highlight_3",
		}

		_, err := helper.Client.Project.
			UpdateOneID(project.ID).
			SetHighlightOrder(highlightOrder).
			Save(helper.Ctx)
		require.NoError(t, err)

		// Get the highlight order
		result, err := service.GetProjectHighlightOrder(project.ID)
		require.NoError(t, err)
		
		// Should return all string items including "N" markers (based on current implementation)
		expectedOrder := []string{"highlight_1", "highlight_2", "N", "highlight_3"}
		assert.Equal(t, expectedOrder, result)
	})

	t.Run("non-existent project", func(t *testing.T) {
		result, err := service.GetProjectHighlightOrder(99999)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}