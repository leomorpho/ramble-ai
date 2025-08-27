package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ramble-ai/ent/schema"
	"ramble-ai/goapp"
)

// setupAppAdditionalTestHelper creates a test helper for App-level additional tests  
func setupAppAdditionalTestHelper(t *testing.T) (*goapp.TestHelper, *App) {
	helper := goapp.NewTestHelper(t)
	
	// Create App instance with the test client
	app := &App{
		ctx:    helper.Ctx,
		client: helper.Client,
	}
	
	return helper, app
}

// TestAppCreateVideoClip tests the App-level CreateVideoClip function
func TestAppCreateVideoClip(t *testing.T) {
	helper, app := setupAppAdditionalTestHelper(t)
	
	project := helper.CreateTestProject("App CreateVideoClip Test")
	
	// Create a temporary video file for testing
	tempDir := t.TempDir()
	videoFile := filepath.Join(tempDir, "app_test_video.mp4")
	err := os.WriteFile(videoFile, []byte("fake video content for app test"), 0644)
	require.NoError(t, err)
	
	nonVideoFile := filepath.Join(tempDir, "not_video.txt")
	err = os.WriteFile(nonVideoFile, []byte("not a video"), 0644)
	require.NoError(t, err)
	
	t.Run("successfully creates video clip through app", func(t *testing.T) {
		clip, err := app.CreateVideoClip(project.ID, videoFile)
		
		assert.NoError(t, err)
		assert.NotNil(t, clip)
		assert.Equal(t, "app_test_video", clip.Name)
		assert.Equal(t, videoFile, clip.FilePath)
		assert.Equal(t, "app_test_video.mp4", clip.FileName)
		assert.Greater(t, clip.FileSize, int64(0))
		assert.Equal(t, "mp4", clip.Format)
		assert.True(t, clip.Exists)
		assert.Contains(t, clip.ThumbnailURL, "/api/thumbnail/")
	})
	
	t.Run("rejects non-video file through app", func(t *testing.T) {
		clip, err := app.CreateVideoClip(project.ID, nonVideoFile)
		
		assert.Error(t, err)
		assert.Nil(t, clip)
		assert.Contains(t, err.Error(), "file is not a supported video format")
	})
	
	t.Run("handles nonexistent project through app", func(t *testing.T) {
		clip, err := app.CreateVideoClip(999999, videoFile)
		
		assert.Error(t, err)
		assert.Nil(t, clip)
		assert.Contains(t, err.Error(), "failed to create video clip")
	})
	
	t.Run("handles nonexistent file through app", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "does_not_exist.mp4")
		
		clip, err := app.CreateVideoClip(project.ID, nonExistentFile)
		
		assert.Error(t, err)
		assert.Nil(t, clip)
		assert.Contains(t, err.Error(), "file does not exist")
	})
}

// TestAppSelectVideoFiles tests the App-level SelectVideoFiles function
func TestAppSelectVideoFiles(t *testing.T) {
	t.Run("function exists and handles dialog error gracefully", func(t *testing.T) {
		t.Skip("Skipping SelectVideoFiles test as it requires valid Wails runtime context")
		// This test verifies the function exists and can be called
		// In a real application context with proper Wails initialization, it would work correctly
		// In test environment, Wails runtime functions fail due to missing GUI context
	})
}

// TestAppHideHighlight tests the App-level HideHighlight function
func TestAppHideHighlight(t *testing.T) {
	helper, app := setupAppAdditionalTestHelper(t)
	
	project := helper.CreateTestProject("App Hide Highlight Test")
	
	t.Run("successfully hides highlight through app", func(t *testing.T) {
		err := app.HideHighlight(project.ID, "app_highlight_1")
		assert.NoError(t, err)
		
		// Verify it was hidden
		hidden, err := app.GetHiddenHighlights(project.ID)
		require.NoError(t, err)
		assert.Contains(t, hidden, "app_highlight_1")
	})
	
	t.Run("handles nonexistent project through app", func(t *testing.T) {
		err := app.HideHighlight(999999, "some_highlight")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get project")
	})
	
	t.Run("handles duplicate hide requests through app", func(t *testing.T) {
		err := app.HideHighlight(project.ID, "app_highlight_duplicate")
		assert.NoError(t, err)
		
		// Hide the same highlight again
		err = app.HideHighlight(project.ID, "app_highlight_duplicate")
		assert.NoError(t, err) // Should not error
		
		// Verify it's only listed once
		hidden, err := app.GetHiddenHighlights(project.ID)
		require.NoError(t, err)
		
		count := 0
		for _, h := range hidden {
			if h == "app_highlight_duplicate" {
				count++
			}
		}
		assert.Equal(t, 1, count, "Should not duplicate hidden highlights")
	})
}

// TestAppRedoHighlightsChange tests the App-level RedoHighlightsChange function
func TestAppRedoHighlightsChange(t *testing.T) {
	helper, app := setupAppAdditionalTestHelper(t)
	
	project := helper.CreateTestProject("App Redo Highlights Test")
	clip := helper.CreateTestVideoClip(project, "App Test Clip")
	
	t.Run("handles no history through app", func(t *testing.T) {
		_, err := app.RedoHighlightsChange(clip.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no history available")
	})
	
	t.Run("handles nonexistent clip through app", func(t *testing.T) {
		_, err := app.RedoHighlightsChange(999999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get video clip")
	})
	
	t.Run("complete undo/redo workflow through app", func(t *testing.T) {
		// Create highlights with proper history setup
		highlights1 := []schema.Highlight{
			{ID: "app_h1", Start: 0, End: 5, ColorID: 1},
		}
		highlights2 := []schema.Highlight{
			{ID: "app_h1", Start: 0, End: 5, ColorID: 1},
			{ID: "app_h2", Start: 10, End: 15, ColorID: 2},
		}
		
		// Set up proper history state with index at second entry (1) so we can undo then redo
		_, err := helper.Client.VideoClip.
			UpdateOneID(clip.ID).
			SetHighlights(highlights2).
			SetHighlightsHistory([][]schema.Highlight{highlights1, highlights2}).
			SetHighlightsHistoryIndex(1). // Set to second entry
			Save(helper.Ctx)
		require.NoError(t, err)
		
		// Undo through app (should go to index 0)
		result, err := app.UndoHighlightsChange(clip.ID)
		require.NoError(t, err)
		assert.Len(t, result, 1) // Should have 1 highlight
		assert.Equal(t, "app_h1", result[0].ID)
		
		// Redo through app (should go back to index 1)
		result, err = app.RedoHighlightsChange(clip.ID)
		require.NoError(t, err)
		assert.Len(t, result, 2) // Should have 2 highlights
		assert.Equal(t, "app_h1", result[0].ID)
		assert.Equal(t, "app_h2", result[1].ID)
		
		// Try to redo beyond available history
		_, err = app.RedoHighlightsChange(clip.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot redo further")
	})
}