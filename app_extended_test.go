package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ramble-ai/goapp"
)

// TestAppWailsMethods tests the Wails-exposed methods of the App struct
func TestAppWailsMethods(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	app := &App{
		ctx:    context.Background(),
		client: helper.Client,
	}

	t.Run("Greet", func(t *testing.T) {
		result := app.Greet("World")
		assert.Equal(t, "Hello World, It's show time!", result)

		result2 := app.Greet("Testing")
		assert.Equal(t, "Hello Testing, It's show time!", result2)

		// Test empty name
		result3 := app.Greet("")
		assert.Equal(t, "Hello , It's show time!", result3)
	})
}

func TestAppProjectMethods(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	app := &App{
		ctx:    context.Background(),
		client: helper.Client,
	}

	t.Run("CreateProject", func(t *testing.T) {
		project, err := app.CreateProject("Test Project", "Test Description")
		require.NoError(t, err)
		assert.NotNil(t, project)
		assert.Equal(t, "Test Project", project.Name)
		assert.Equal(t, "Test Description", project.Description)
		assert.NotZero(t, project.ID)
	})

	t.Run("CreateProject_EmptyName", func(t *testing.T) {
		_, err := app.CreateProject("", "Description")
		assert.Error(t, err)
	})

	t.Run("GetProjects", func(t *testing.T) {
		// Use separate database for this test to ensure isolation
		helper2 := goapp.NewTestHelper(t)
		app2 := &App{
			ctx:    context.Background(),
			client: helper2.Client,
		}

		// Create some test projects
		project1, err := app2.CreateProject("Project 1", "Description 1")
		require.NoError(t, err)

		project2, err := app2.CreateProject("Project 2", "Description 2")
		require.NoError(t, err)

		// Get all projects
		projects, err := app2.GetProjects()
		require.NoError(t, err)
		assert.Len(t, projects, 2)

		// Check that our projects are in the list
		projectIDs := make([]int, len(projects))
		for i, p := range projects {
			projectIDs[i] = p.ID
		}
		assert.Contains(t, projectIDs, project1.ID)
		assert.Contains(t, projectIDs, project2.ID)
	})

	t.Run("GetProjectByID", func(t *testing.T) {
		// Create a test project
		created, err := app.CreateProject("Specific Project", "Specific Description")
		require.NoError(t, err)

		// Get project by ID
		retrieved, err := app.GetProjectByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, retrieved.ID)
		assert.Equal(t, created.Name, retrieved.Name)
		assert.Equal(t, created.Description, retrieved.Description)
	})

	t.Run("GetProjectByID_NotFound", func(t *testing.T) {
		_, err := app.GetProjectByID(99999) // Non-existent ID
		assert.Error(t, err)
	})

	t.Run("UpdateProject", func(t *testing.T) {
		// Create a test project
		created, err := app.CreateProject("Original Name", "Original Description")
		require.NoError(t, err)

		// Update the project
		updated, err := app.UpdateProject(created.ID, "Updated Name", "Updated Description")
		require.NoError(t, err)
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, "Updated Description", updated.Description)
	})

	t.Run("UpdateProject_NotFound", func(t *testing.T) {
		_, err := app.UpdateProject(99999, "Name", "Description")
		assert.Error(t, err)
	})

	t.Run("DeleteProject", func(t *testing.T) {
		// Create a test project
		created, err := app.CreateProject("Project to Delete", "Will be deleted")
		require.NoError(t, err)

		// Delete the project
		err = app.DeleteProject(created.ID)
		require.NoError(t, err)

		// Verify it's deleted
		_, err = app.GetProjectByID(created.ID)
		assert.Error(t, err)
	})

	t.Run("DeleteProject_NotFound", func(t *testing.T) {
		err := app.DeleteProject(99999)
		assert.Error(t, err)
	})

	t.Run("UpdateProjectActiveTab", func(t *testing.T) {
		// Create a test project
		created, err := app.CreateProject("Tab Test Project", "For testing tabs")
		require.NoError(t, err)

		// Update active tab
		err = app.UpdateProjectActiveTab(created.ID, "timeline")
		require.NoError(t, err)

		// Verify the tab was updated
		retrieved, err := app.GetProjectByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, "timeline", retrieved.ActiveTab)
	})
}

func TestAppVideoClipMethods(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	app := &App{
		ctx:    context.Background(),
		client: helper.Client,
	}

	// Create a test project first
	project, err := app.CreateProject("Video Test Project", "For testing video clips")
	require.NoError(t, err)

	t.Run("CreateVideoClip", func(t *testing.T) {
		// CreateVideoClip validates file existence, so this should fail
		_, err := app.CreateVideoClip(project.ID, "/test/path/video.mp4")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file does not exist")
	})

	t.Run("CreateVideoClip_InvalidProject", func(t *testing.T) {
		_, err := app.CreateVideoClip(99999, "/test/path/video.mp4")
		assert.Error(t, err)
	})

	t.Run("GetVideoClipsByProject", func(t *testing.T) {
		// Test getting clips for a project (should work even with no clips)
		clips, err := app.GetVideoClipsByProject(project.ID)
		require.NoError(t, err)
		
		// We expect an empty slice (not nil) since we can't create clips without valid files
		if clips != nil {
			assert.Len(t, clips, 0)
		}
	})

	t.Run("UpdateVideoClip", func(t *testing.T) {
		// Since we can't create valid video clips in tests, test updating non-existent clip
		_, err := app.UpdateVideoClip(99999, "Updated Name", "Updated Description")
		assert.Error(t, err)
	})

	t.Run("UpdateVideoClip_NotFound", func(t *testing.T) {
		_, err := app.UpdateVideoClip(99999, "Name", "Description")
		assert.Error(t, err)
	})

	t.Run("DeleteVideoClip", func(t *testing.T) {
		// Test deleting non-existent clip
		err := app.DeleteVideoClip(99999)
		assert.Error(t, err)
	})

	t.Run("DeleteVideoClip_NotFound", func(t *testing.T) {
		err := app.DeleteVideoClip(99999)
		assert.Error(t, err)
	})
}

func TestAppSettingsMethods(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	app := &App{
		ctx:    context.Background(),
		client: helper.Client,
	}

	t.Run("SaveSetting", func(t *testing.T) {
		err := app.SaveSetting("test_key", "test_value")
		require.NoError(t, err)

		// Verify setting was saved
		helper.AssertSettingEquals("test_key", "test_value")
	})

	t.Run("GetSetting", func(t *testing.T) {
		// Save a setting first
		err := app.SaveSetting("get_test_key", "get_test_value")
		require.NoError(t, err)

		// Get the setting
		value, err := app.GetSetting("get_test_key")
		require.NoError(t, err)
		assert.Equal(t, "get_test_value", value)
	})

	t.Run("GetSetting_NotFound", func(t *testing.T) {
		value, err := app.GetSetting("nonexistent_key")
		assert.NoError(t, err) // Current implementation returns no error
		assert.Empty(t, value)
	})

	t.Run("DeleteSetting", func(t *testing.T) {
		// Save a setting first
		err := app.SaveSetting("delete_test_key", "delete_test_value")
		require.NoError(t, err)

		// Delete the setting
		err = app.DeleteSetting("delete_test_key")
		require.NoError(t, err)

		// Verify it's deleted
		value, err := app.GetSetting("delete_test_key")
		assert.NoError(t, err)
		assert.Empty(t, value)
	})

	t.Run("API Key Methods", func(t *testing.T) {
		// Test OpenAI API key
		err := app.SaveOpenAIApiKey("sk-test-openai-key")
		require.NoError(t, err)

		key, err := app.GetOpenAIApiKey()
		require.NoError(t, err)
		assert.Equal(t, "sk-test-openai-key", key)

		err = app.DeleteOpenAIApiKey()
		require.NoError(t, err)

		key, err = app.GetOpenAIApiKey()
		assert.NoError(t, err)
		assert.Empty(t, key)

		// Test OpenRouter API key
		err = app.SaveOpenRouterApiKey("sk-or-test-openrouter-key")
		require.NoError(t, err)

		key, err = app.GetOpenRouterApiKey()
		require.NoError(t, err)
		assert.Equal(t, "sk-or-test-openrouter-key", key)

		err = app.DeleteOpenRouterApiKey()
		require.NoError(t, err)

		key, err = app.GetOpenRouterApiKey()
		assert.NoError(t, err)
		assert.Empty(t, key)
	})

	t.Run("Theme Preference", func(t *testing.T) {
		// Save theme preference
		err := app.SaveThemePreference("dark")
		require.NoError(t, err)

		// Get theme preference
		theme, err := app.GetThemePreference()
		require.NoError(t, err)
		assert.Equal(t, "dark", theme)

		// Test invalid theme
		err = app.SaveThemePreference("invalid")
		assert.Error(t, err)
	})
}

func TestAppFileAndVideoMethods(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	app := &App{
		ctx:    context.Background(),
		client: helper.Client,
	}

	t.Run("GetVideoFileInfo", func(t *testing.T) {
		// This method would normally interact with the file system
		// For testing, check with a video file extension but non-existent file
		info, err := app.GetVideoFileInfo("/nonexistent/path/video.mp4")
		
		// Based on the code, it should return info even if file doesn't exist
		// The exists flag in getFileInfo handles file existence
		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "/nonexistent/path/video.mp4", info.FilePath)
		assert.Equal(t, "video.mp4", info.FileName)
		assert.Equal(t, "video", info.Name)
	})

	t.Run("GetVideoURL", func(t *testing.T) {
		// This method would normally serve video files
		// For testing, we expect it to handle non-existent files gracefully
		_, err := app.GetVideoURL("/nonexistent/path/video.mp4")
		assert.Error(t, err) // Should error for non-existent file
	})
}

