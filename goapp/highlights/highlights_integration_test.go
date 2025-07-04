package highlights

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"MYAPP/ent"
	"MYAPP/ent/schema"
	"MYAPP/ent/enttest"
	_ "github.com/mattn/go-sqlite3"
)

// Integration tests that require a real database connection
// These tests use an in-memory SQLite database for testing

func setupTestClient(t *testing.T) *ent.Client {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	return client
}

func createTestProject(t *testing.T, client *ent.Client, ctx context.Context) *ent.Project {
	project, err := client.Project.
		Create().
		SetName("Test Project").
		SetDescription("Test project for highlight testing").
		SetPath("/test/project/path").
		Save(ctx)
	require.NoError(t, err)
	return project
}

func createTestVideoClipWithHighlights(t *testing.T, client *ent.Client, ctx context.Context, project *ent.Project) *ent.VideoClip {
	words := []schema.Word{
		{Word: "Hello", Start: 0.0, End: 0.5},
		{Word: "world", Start: 0.5, End: 1.0},
		{Word: "this", Start: 1.0, End: 1.5},
		{Word: "is", Start: 1.5, End: 2.0},
		{Word: "a", Start: 2.0, End: 2.5},
		{Word: "test", Start: 2.5, End: 3.0},
		{Word: "video", Start: 3.0, End: 3.5},
	}

	highlights := []schema.Highlight{
		{
			ID:    "h1",
			Start: 0.0,
			End:   1.0,
			Color: "red",
		},
		{
			ID:    "h2",
			Start: 2.0,
			End:   3.0,
			Color: "blue",
		},
	}

	suggestedHighlights := []schema.Highlight{
		{
			ID:    "s1",
			Start: 1.0,
			End:   2.0,
			Color: "yellow",
		},
	}

	clip, err := client.VideoClip.
		Create().
		SetName("Test Video").
		SetFilePath("/test/video.mp4").
		SetDuration(10.0).
		SetTranscription("Hello world this is a test video").
		SetTranscriptionWords(words).
		SetHighlights(highlights).
		SetSuggestedHighlights(suggestedHighlights).
		SetProject(project).
		Save(ctx)
	require.NoError(t, err)
	return clip
}

func TestHighlightService_GetSuggestedHighlights_Integration(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()
	service := NewHighlightService(client, ctx)

	// Create test data
	project := createTestProject(t, client, ctx)
	clip := createTestVideoClipWithHighlights(t, client, ctx, project)

	// Test GetSuggestedHighlights
	suggestions, err := service.GetSuggestedHighlights(clip.ID)
	require.NoError(t, err)
	assert.Len(t, suggestions, 1)
	
	suggestion := suggestions[0]
	assert.Equal(t, "s1", suggestion.ID)
	assert.Equal(t, "yellow", suggestion.Color)
	assert.Equal(t, "this is a", suggestion.Text)
	assert.Equal(t, 2, suggestion.Start) // Word index
	assert.Equal(t, 4, suggestion.End)   // Word index
}

func TestHighlightService_ClearSuggestedHighlights_Integration(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()
	service := NewHighlightService(client, ctx)

	// Create test data
	project := createTestProject(t, client, ctx)
	clip := createTestVideoClipWithHighlights(t, client, ctx, project)

	// Verify suggested highlights exist
	suggestions, err := service.GetSuggestedHighlights(clip.ID)
	require.NoError(t, err)
	assert.Len(t, suggestions, 1)

	// Clear suggested highlights
	err = service.ClearSuggestedHighlights(clip.ID)
	require.NoError(t, err)

	// Verify suggested highlights are cleared
	suggestions, err = service.GetSuggestedHighlights(clip.ID)
	require.NoError(t, err)
	assert.Len(t, suggestions, 0)
}

func TestHighlightService_DeleteHighlight_Integration(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()
	service := NewHighlightService(client, ctx)

	// Create test data
	project := createTestProject(t, client, ctx)
	clip := createTestVideoClipWithHighlights(t, client, ctx, project)

	// Verify initial highlights
	updatedClip, err := client.VideoClip.Get(ctx, clip.ID)
	require.NoError(t, err)
	assert.Len(t, updatedClip.Highlights, 2)

	// Delete one highlight
	err = service.DeleteHighlight(clip.ID, "h1")
	require.NoError(t, err)

	// Verify highlight was deleted
	updatedClip, err = client.VideoClip.Get(ctx, clip.ID)
	require.NoError(t, err)
	assert.Len(t, updatedClip.Highlights, 1)
	assert.Equal(t, "h2", updatedClip.Highlights[0].ID)
}

func TestHighlightService_GetProjectHighlights_Integration(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()
	service := NewHighlightService(client, ctx)

	// Create test data
	project := createTestProject(t, client, ctx)
	clip1 := createTestVideoClipWithHighlights(t, client, ctx, project)
	
	// Create second clip with different highlights
	clip2, err := client.VideoClip.
		Create().
		SetName("Test Video 2").
		SetFilePath("/test/video2.mp4").
		SetDuration(5.0).
		SetTranscription("Another test video").
		SetTranscriptionWords([]schema.Word{
			{Word: "Another", Start: 0.0, End: 0.5},
			{Word: "test", Start: 0.5, End: 1.0},
			{Word: "video", Start: 1.0, End: 1.5},
		}).
		SetHighlights([]schema.Highlight{
			{
				ID:    "h3",
				Start: 0.0,
				End:   1.0,
				Color: "green",
			},
		}).
		SetProject(project).
		Save(ctx)
	require.NoError(t, err)

	// Get project highlights
	projectHighlights, err := service.GetProjectHighlights(project.ID)
	require.NoError(t, err)
	assert.Len(t, projectHighlights, 2)

	// Verify first clip highlights
	var clip1Highlights *ProjectHighlight
	var clip2Highlights *ProjectHighlight
	
	for _, ph := range projectHighlights {
		if ph.VideoClipID == clip1.ID {
			clip1Highlights = &ph
		} else if ph.VideoClipID == clip2.ID {
			clip2Highlights = &ph
		}
	}

	require.NotNil(t, clip1Highlights)
	require.NotNil(t, clip2Highlights)

	// Verify clip1 highlights
	assert.Equal(t, "Test Video", clip1Highlights.VideoClipName)
	assert.Equal(t, "/test/video.mp4", clip1Highlights.FilePath)
	assert.Equal(t, 10.0, clip1Highlights.Duration)
	assert.Len(t, clip1Highlights.Highlights, 2)
	assert.Equal(t, "Hello world", clip1Highlights.Highlights[0].Text)
	assert.Equal(t, "a test", clip1Highlights.Highlights[1].Text)

	// Verify clip2 highlights
	assert.Equal(t, "Test Video 2", clip2Highlights.VideoClipName)
	assert.Equal(t, "/test/video2.mp4", clip2Highlights.FilePath)
	assert.Equal(t, 5.0, clip2Highlights.Duration)
	assert.Len(t, clip2Highlights.Highlights, 1)
	assert.Equal(t, "Another test", clip2Highlights.Highlights[0].Text)
}

func TestHighlightService_GetProjectHighlightsForExport_Integration(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()
	service := NewHighlightService(client, ctx)

	// Create test data
	project := createTestProject(t, client, ctx)
	clip := createTestVideoClipWithHighlights(t, client, ctx, project)

	// Get highlights for export
	segments, err := service.GetProjectHighlightsForExport(project.ID)
	require.NoError(t, err)
	assert.Len(t, segments, 2)

	// Verify first segment
	segment1 := segments[0]
	assert.Equal(t, "h1", segment1.ID)
	assert.Equal(t, "/test/video.mp4", segment1.VideoPath)
	assert.Equal(t, 0.0, segment1.Start)
	assert.Equal(t, 1.0, segment1.End)
	assert.Equal(t, "red", segment1.Color)
	assert.Equal(t, "Hello world", segment1.Text)
	assert.Equal(t, clip.ID, segment1.VideoClipID)
	assert.Equal(t, "Test Video", segment1.VideoClipName)

	// Verify second segment
	segment2 := segments[1]
	assert.Equal(t, "h2", segment2.ID)
	assert.Equal(t, "/test/video.mp4", segment2.VideoPath)
	assert.Equal(t, 2.0, segment2.Start)
	assert.Equal(t, 3.0, segment2.End)
	assert.Equal(t, "blue", segment2.Color)
	assert.Equal(t, "a test", segment2.Text)
	assert.Equal(t, clip.ID, segment2.VideoClipID)
	assert.Equal(t, "Test Video", segment2.VideoClipName)
}

func TestHighlightService_AI_Settings_Integration(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()
	service := NewHighlightService(client, ctx)

	// Create test data
	project := createTestProject(t, client, ctx)

	// Test getting default AI settings (returns empty strings when no settings exist)
	settings, err := service.GetProjectHighlightAISettings(project.ID)
	require.NoError(t, err)
	assert.Equal(t, "", settings.AIModel)
	assert.Equal(t, "", settings.AIPrompt)

	// Test saving AI settings
	newSettings := ProjectHighlightAISettings{
		AIModel:  "anthropic/claude-3-haiku",
		AIPrompt: "Custom prompt for highlight suggestions",
	}
	
	err = service.SaveProjectHighlightAISettings(project.ID, newSettings)
	require.NoError(t, err)

	// Test getting saved AI settings
	savedSettings, err := service.GetProjectHighlightAISettings(project.ID)
	require.NoError(t, err)
	assert.Equal(t, "anthropic/claude-3-haiku", savedSettings.AIModel)
	assert.Equal(t, "Custom prompt for highlight suggestions", savedSettings.AIPrompt)
}

func TestHighlightService_GetProjectHighlightOrder_Integration(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()
	service := NewHighlightService(client, ctx)

	// Create test data
	project := createTestProject(t, client, ctx)

	// Test getting highlight order when none exists
	order, err := service.GetProjectHighlightOrder(project.ID)
	require.NoError(t, err)
	assert.Empty(t, order)

	// Manually create a highlight order setting
	settingKey := "project_1_highlight_order"
	_, err = client.Settings.
		Create().
		SetKey(settingKey).
		SetValue(`["h3", "h1", "h2"]`).
		Save(ctx)
	require.NoError(t, err)

	// Test getting saved highlight order
	order, err = service.GetProjectHighlightOrder(project.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"h3", "h1", "h2"}, order)
}

func TestHighlightService_Error_Handling_Integration(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()
	service := NewHighlightService(client, ctx)

	// Test getting suggested highlights for non-existent video
	_, err := service.GetSuggestedHighlights(99999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get video clip")

	// Test clearing suggested highlights for non-existent video
	err = service.ClearSuggestedHighlights(99999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to clear suggested highlights")

	// Test deleting highlight for non-existent video
	err = service.DeleteHighlight(99999, "h1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get video clip")

	// Test getting project highlights for non-existent project
	highlights, err := service.GetProjectHighlights(99999)
	require.NoError(t, err) // Should not error, just return empty
	assert.Empty(t, highlights)
}

func TestHighlightService_getSetting_Integration(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()
	service := NewHighlightService(client, ctx)

	// Test getting non-existent setting
	value, err := service.getSetting("non_existent_key")
	assert.NoError(t, err) // Should not error for non-existent setting
	assert.Equal(t, "", value)

	// Test getting setting with empty key
	value, err = service.getSetting("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "setting key cannot be empty")

	// Create a setting
	_, err = client.Settings.
		Create().
		SetKey("test_key").
		SetValue("test_value").
		Save(ctx)
	require.NoError(t, err)

	// Test getting existing setting
	value, err = service.getSetting("test_key")
	require.NoError(t, err)
	assert.Equal(t, "test_value", value)
}

func TestHighlightService_saveSetting_Integration(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()
	service := NewHighlightService(client, ctx)

	// Test saving new setting
	err := service.saveSetting("new_key", "new_value")
	require.NoError(t, err)

	// Verify setting was saved
	value, err := service.getSetting("new_key")
	require.NoError(t, err)
	assert.Equal(t, "new_value", value)

	// Test updating existing setting
	err = service.saveSetting("new_key", "updated_value")
	require.NoError(t, err)

	// Verify setting was updated
	value, err = service.getSetting("new_key")
	require.NoError(t, err)
	assert.Equal(t, "updated_value", value)
}

// Test complex scenarios
func TestHighlightService_ComplexScenario_Integration(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()
	service := NewHighlightService(client, ctx)

	// Create test data with multiple projects and clips
	project1 := createTestProject(t, client, ctx)
	project2, err := client.Project.
		Create().
		SetName("Test Project 2").
		SetDescription("Second test project").
		SetPath("/test/project2/path").
		Save(ctx)
	require.NoError(t, err)

	// Create clips for both projects
	clip1 := createTestVideoClipWithHighlights(t, client, ctx, project1)
	clip2, err := client.VideoClip.
		Create().
		SetName("Project 2 Video").
		SetFilePath("/test/project2_video.mp4").
		SetDuration(8.0).
		SetTranscription("Project two video content").
		SetTranscriptionWords([]schema.Word{
			{Word: "Project", Start: 0.0, End: 0.5},
			{Word: "two", Start: 0.5, End: 1.0},
			{Word: "video", Start: 1.0, End: 1.5},
			{Word: "content", Start: 1.5, End: 2.0},
		}).
		SetHighlights([]schema.Highlight{
			{
				ID:    "p2h1",
				Start: 0.0,
				End:   1.0,
				Color: "purple",
			},
		}).
		SetProject(project2).
		Save(ctx)
	require.NoError(t, err)

	// Test that each project only returns its own highlights
	project1Highlights, err := service.GetProjectHighlights(project1.ID)
	require.NoError(t, err)
	assert.Len(t, project1Highlights, 1)
	assert.Equal(t, clip1.ID, project1Highlights[0].VideoClipID)

	project2Highlights, err := service.GetProjectHighlights(project2.ID)
	require.NoError(t, err)
	assert.Len(t, project2Highlights, 1)
	assert.Equal(t, clip2.ID, project2Highlights[0].VideoClipID)

	// Test export functionality with custom order
	// GetProjectHighlightsForExport should only get highlights from the specified project
	segments, err := service.GetProjectHighlightsForExport(project1.ID)
	require.NoError(t, err)
	assert.Len(t, segments, 2) // Gets only 2 highlights from project1

	// Apply custom order
	customOrder := []string{"h2", "h1"}
	orderedSegments := service.ApplyHighlightOrder(segments, customOrder)
	assert.Len(t, orderedSegments, 3)
	assert.Equal(t, "h2", orderedSegments[0].ID)
	assert.Equal(t, "h1", orderedSegments[1].ID)

	// Test AI settings for different projects
	aiSettings1 := ProjectHighlightAISettings{
		AIModel:  "openai/gpt-4",
		AIPrompt: "Project 1 custom prompt",
	}
	err = service.SaveProjectHighlightAISettings(project1.ID, aiSettings1)
	require.NoError(t, err)

	aiSettings2 := ProjectHighlightAISettings{
		AIModel:  "anthropic/claude-3-sonnet",
		AIPrompt: "Project 2 custom prompt",
	}
	err = service.SaveProjectHighlightAISettings(project2.ID, aiSettings2)
	require.NoError(t, err)

	// Verify each project has its own AI settings
	savedSettings1, err := service.GetProjectHighlightAISettings(project1.ID)
	require.NoError(t, err)
	assert.Equal(t, "openai/gpt-4", savedSettings1.AIModel)
	assert.Equal(t, "Project 1 custom prompt", savedSettings1.AIPrompt)

	savedSettings2, err := service.GetProjectHighlightAISettings(project2.ID)
	require.NoError(t, err)
	assert.Equal(t, "anthropic/claude-3-sonnet", savedSettings2.AIModel)
	assert.Equal(t, "Project 2 custom prompt", savedSettings2.AIPrompt)
}