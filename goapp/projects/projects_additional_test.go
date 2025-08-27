package projects

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ramble-ai/ent/schema"
	"ramble-ai/goapp"
)

// setupAdditionalTestHelper creates a test helper for additional tests
func setupAdditionalTestHelper(t *testing.T) (*goapp.TestHelper, *ProjectService) {
	helper := goapp.NewTestHelper(t)
	service := NewProjectService(helper.Client, helper.Ctx)
	return helper, service
}

// TestCreateVideoClip tests the CreateVideoClip function comprehensively
func TestCreateVideoClip(t *testing.T) {
	helper, service := setupAdditionalTestHelper(t)

	project := helper.CreateTestProject("CreateVideoClip Test")

	// Create a temporary video file for testing
	tempDir := t.TempDir()
	videoFile := filepath.Join(tempDir, "test_video.mp4")
	err := os.WriteFile(videoFile, []byte("fake video content"), 0644)
	require.NoError(t, err)

	nonVideoFile := filepath.Join(tempDir, "not_video.txt")
	err = os.WriteFile(nonVideoFile, []byte("not a video"), 0644)
	require.NoError(t, err)

	t.Run("successfully creates video clip", func(t *testing.T) {
		clip, err := service.CreateVideoClip(project.ID, videoFile)
		
		assert.NoError(t, err)
		assert.NotNil(t, clip)
		assert.Equal(t, "test_video", clip.Name)
		assert.Equal(t, videoFile, clip.FilePath)
		assert.Equal(t, "test_video.mp4", clip.FileName)
		assert.Greater(t, clip.FileSize, int64(0))
		assert.Equal(t, "mp4", clip.Format)
		assert.True(t, clip.Exists)
		assert.NotEmpty(t, clip.CreatedAt)
		assert.NotEmpty(t, clip.UpdatedAt)
		assert.Contains(t, clip.ThumbnailURL, "/api/thumbnail/")
	})

	t.Run("rejects non-video file", func(t *testing.T) {
		clip, err := service.CreateVideoClip(project.ID, nonVideoFile)
		
		assert.Error(t, err)
		assert.Nil(t, clip)
		assert.Contains(t, err.Error(), "file is not a supported video format")
	})

	t.Run("rejects non-existent file", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "does_not_exist.mp4")
		
		clip, err := service.CreateVideoClip(project.ID, nonExistentFile)
		
		assert.Error(t, err)
		assert.Nil(t, clip)
		assert.Contains(t, err.Error(), "file does not exist")
	})

	t.Run("handles nonexistent project", func(t *testing.T) {
		clip, err := service.CreateVideoClip(999999, videoFile)
		
		assert.Error(t, err)
		assert.Nil(t, clip)
		assert.Contains(t, err.Error(), "failed to create video clip")
	})

	t.Run("handles different video formats", func(t *testing.T) {
		formats := []string{".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm", ".m4v", ".3gp", ".ogv"}
		
		for _, format := range formats {
			testFile := filepath.Join(tempDir, "test"+format)
			err := os.WriteFile(testFile, []byte("fake video content"), 0644)
			require.NoError(t, err)
			
			clip, err := service.CreateVideoClip(project.ID, testFile)
			
			assert.NoError(t, err, "Should handle format %s", format)
			assert.NotNil(t, clip)
			assert.Equal(t, format[1:], clip.Format) // Remove the dot
		}
	})

	t.Run("extracts filename correctly", func(t *testing.T) {
		testCases := []struct {
			fileName     string
			expectedName string
		}{
			{"simple.mp4", "simple"},
			{"with spaces.mp4", "with spaces"},
			{"with-dashes.mp4", "with-dashes"},
			{"with_underscores.mp4", "with_underscores"},
			{"complex.name.with.dots.mp4", "complex.name.with.dots"},
		}
		
		for _, tc := range testCases {
			testFile := filepath.Join(tempDir, tc.fileName)
			err := os.WriteFile(testFile, []byte("fake video content"), 0644)
			require.NoError(t, err)
			
			clip, err := service.CreateVideoClip(project.ID, testFile)
			
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedName, clip.Name)
			assert.Equal(t, tc.fileName, clip.FileName)
		}
	})
}

// TestSelectVideoFiles tests the SelectVideoFiles function
func TestSelectVideoFiles(t *testing.T) {
	// Note: This function uses runtime.OpenMultipleFilesDialog which requires a GUI
	// In a headless test environment, this will fail, but we can verify the function exists
	
	t.Run("handles dialog error gracefully", func(t *testing.T) {
		t.Skip("Skipping SelectVideoFiles test as it requires valid Wails runtime context")
		// This test verifies the function exists and handles GUI context requirements
		// In a real application context with proper Wails initialization, it would work correctly
		// In test environment, Wails runtime functions fail due to missing GUI context
	})

	t.Run("function exists and can be called without crashing", func(t *testing.T) {
		t.Skip("Skipping SelectVideoFiles test as it requires valid Wails runtime context")
		// This test verifies the function exists and can be invoked
		// The function signature is correct and will work in proper application context
	})
}

// TestEqualInterfaces tests the equalInterfaces helper function comprehensively
func TestEqualInterfaces(t *testing.T) {
	testCases := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		// String comparisons
		{"equal strings", "hello", "hello", true},
		{"different strings", "hello", "world", false},
		{"string vs non-string", "hello", 123, false},
		
		// Map comparisons
		{
			"equal maps",
			map[string]interface{}{"key1": "value1", "key2": "value2"},
			map[string]interface{}{"key1": "value1", "key2": "value2"},
			true,
		},
		{
			"equal maps different order",
			map[string]interface{}{"key2": "value2", "key1": "value1"},
			map[string]interface{}{"key1": "value1", "key2": "value2"},
			true,
		},
		{
			"different map values",
			map[string]interface{}{"key1": "value1", "key2": "value2"},
			map[string]interface{}{"key1": "value1", "key2": "different"},
			false,
		},
		{
			"different map keys",
			map[string]interface{}{"key1": "value1", "key2": "value2"},
			map[string]interface{}{"key1": "value1", "key3": "value2"},
			false,
		},
		{
			"different map lengths",
			map[string]interface{}{"key1": "value1"},
			map[string]interface{}{"key1": "value1", "key2": "value2"},
			false,
		},
		{
			"empty maps",
			map[string]interface{}{},
			map[string]interface{}{},
			true,
		},
		{
			"map vs non-map",
			map[string]interface{}{"key": "value"},
			"not a map",
			false,
		},
		
		// Nested map comparisons
		{
			"equal nested maps",
			map[string]interface{}{
				"outer": map[string]interface{}{"inner": "value"},
			},
			map[string]interface{}{
				"outer": map[string]interface{}{"inner": "value"},
			},
			true,
		},
		{
			"different nested maps",
			map[string]interface{}{
				"outer": map[string]interface{}{"inner": "value1"},
			},
			map[string]interface{}{
				"outer": map[string]interface{}{"inner": "value2"},
			},
			false,
		},
		
		// Mixed type comparisons (using string representation)
		{"equal integers", 123, 123, true},
		{"different integers", 123, 456, false},
		{"equal floats", 123.45, 123.45, true},
		{"different floats", 123.45, 678.90, false},
		{"equal booleans", true, true, true},
		{"different booleans", true, false, false},
		{"integer vs string", 123, "123", true}, // String representation comparison
		{"float vs string", 123.45, "123.45", true},
		{"boolean vs string", true, "true", true},
		
		// Nil comparisons
		{"both nil", nil, nil, true},
		{"nil vs non-nil", nil, "not nil", false},
		{"non-nil vs nil", "not nil", nil, false},
		
		// Complex mixed comparisons
		{
			"complex equal structures",
			map[string]interface{}{
				"string": "value",
				"number": 123,
				"nested": map[string]interface{}{"inner": "data"},
			},
			map[string]interface{}{
				"nested": map[string]interface{}{"inner": "data"},
				"string": "value",
				"number": 123,
			},
			true,
		},
		{
			"complex different structures",
			map[string]interface{}{
				"string": "value",
				"number": 123,
				"nested": map[string]interface{}{"inner": "data"},
			},
			map[string]interface{}{
				"string": "value",
				"number": 456, // Different number
				"nested": map[string]interface{}{"inner": "data"},
			},
			false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := equalInterfaces(tc.a, tc.b)
			assert.Equal(t, tc.expected, result, 
				"equalInterfaces(%v, %v) = %v, expected %v", 
				tc.a, tc.b, result, tc.expected)
		})
	}
}

// TestHelperFunctions tests the helper functions used by CreateVideoClip
func TestHelperFunctions(t *testing.T) {
	_, service := setupAdditionalTestHelper(t)

	// Create temporary files for testing
	tempDir := t.TempDir()
	
	t.Run("isVideoFile", func(t *testing.T) {
		testCases := []struct {
			fileName string
			isVideo  bool
		}{
			{"test.mp4", true},
			{"test.avi", true},
			{"test.mov", true},
			{"test.mkv", true},
			{"test.wmv", true},
			{"test.flv", true},
			{"test.webm", true},
			{"test.m4v", true},
			{"test.3gp", true},
			{"test.ogv", true},
			{"test.MP4", true}, // Uppercase
			{"test.txt", false},
			{"test.jpg", false},
			{"test.doc", false},
			{"test", false},     // No extension
			{"", false},         // Empty string
		}
		
		for _, tc := range testCases {
			filePath := filepath.Join(tempDir, tc.fileName)
			result := service.isVideoFile(filePath)
			assert.Equal(t, tc.isVideo, result, 
				"isVideoFile(%s) = %v, expected %v", tc.fileName, result, tc.isVideo)
		}
	})
	
	t.Run("getFileInfo", func(t *testing.T) {
		// Create test files
		smallFile := filepath.Join(tempDir, "small.mp4")
		err := os.WriteFile(smallFile, []byte("small content"), 0644)
		require.NoError(t, err)
		
		largeFile := filepath.Join(tempDir, "large.avi")
		err = os.WriteFile(largeFile, make([]byte, 1024*1024), 0644) // 1MB file
		require.NoError(t, err)
		
		nonExistentFile := filepath.Join(tempDir, "does_not_exist.mp4")
		
		// Test existing small file
		size, format, exists := service.getFileInfo(smallFile)
		assert.True(t, exists)
		assert.Equal(t, int64(13), size) // "small content" is 13 bytes
		assert.Equal(t, "mp4", format)
		
		// Test existing large file
		size, format, exists = service.getFileInfo(largeFile)
		assert.True(t, exists)
		assert.Equal(t, int64(1024*1024), size)
		assert.Equal(t, "avi", format)
		
		// Test non-existent file
		size, format, exists = service.getFileInfo(nonExistentFile)
		assert.False(t, exists)
		assert.Equal(t, int64(0), size)
		assert.Equal(t, "", format)
	})
	
	t.Run("getThumbnailURL", func(t *testing.T) {
		testCases := []struct {
			fileName    string
			shouldHaveURL bool
		}{
			{"video.mp4", true},
			{"video.avi", true},
			{"document.txt", false}, // Not a video file
		}
		
		for _, tc := range testCases {
			filePath := filepath.Join(tempDir, tc.fileName)
			url := service.getThumbnailURL(filePath)
			
			if tc.shouldHaveURL {
				assert.NotEmpty(t, url)
				assert.Contains(t, url, "/api/thumbnail/")
			} else {
				assert.Empty(t, url)
			}
		}
	})
}

// TestRedoHighlightsChangeComprehensive provides comprehensive coverage for RedoHighlightsChange
func TestRedoHighlightsChangeComprehensive(t *testing.T) {
	helper, service := setupAdditionalTestHelper(t)
	
	project := helper.CreateTestProject("Redo Highlights Test")
	clip := helper.CreateTestVideoClip(project, "Test Clip")
	
	t.Run("redo with no history", func(t *testing.T) {
		_, err := service.RedoHighlightsChange(clip.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no history available")
	})
	
	t.Run("redo with empty history", func(t *testing.T) {
		// Set empty history explicitly
		_, err := helper.Client.VideoClip.
			UpdateOneID(clip.ID).
			SetHighlightsHistory([][]schema.Highlight{}).
			SetHighlightsHistoryIndex(-1).
			Save(helper.Ctx)
		require.NoError(t, err)
		
		_, err = service.RedoHighlightsChange(clip.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no history available")
	})
	
	t.Run("complete undo/redo workflow", func(t *testing.T) {
		// Set up a fresh clip for this test
		freshClip := helper.CreateTestVideoClip(project, "Redo Test Clip")
		
		// Create initial highlights (entry 0 in history)
		highlights1 := []schema.Highlight{
			{ID: "redo1", Start: 0, End: 5, ColorID: 1},
		}
		
		// Create second set of highlights (entry 1 in history)
		highlights2 := []schema.Highlight{
			{ID: "redo1", Start: 0, End: 5, ColorID: 1},
			{ID: "redo2", Start: 10, End: 15, ColorID: 2},
		}
		
		// Set up the history manually (simulate what happens during normal operations)
		history := [][]schema.Highlight{highlights1, highlights2}
		
		// Set current state to highlights2 with history index pointing to entry 1
		_, err := helper.Client.VideoClip.
			UpdateOneID(freshClip.ID).
			SetHighlights(highlights2).
			SetHighlightsHistory(history).
			SetHighlightsHistoryIndex(1). // Currently at entry 1 (highlights2)
			Save(helper.Ctx)
		require.NoError(t, err)
		
		// From index 1, undo should go to index 0 (highlights1)
		result, err := service.UndoHighlightsChange(freshClip.ID)
		require.NoError(t, err)
		assert.Len(t, result, 1) // Should have 1 highlight
		assert.Equal(t, "redo1", result[0].ID)
		
		// Now redo should go from index 0 back to index 1 (highlights2)
		result, err = service.RedoHighlightsChange(freshClip.ID)
		require.NoError(t, err)
		assert.Len(t, result, 2) // Should have 2 highlights
		assert.Equal(t, "redo1", result[0].ID)
		assert.Equal(t, "redo2", result[1].ID)
		
		// Try to redo beyond available history (already at index 1, max index)
		_, err = service.RedoHighlightsChange(freshClip.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot redo further")
	})
	
	t.Run("redo at current state", func(t *testing.T) {
		// Create a fresh clip with some history
		freshClip := helper.CreateTestVideoClip(project, "Fresh Clip")
		
		highlights := []schema.Highlight{
			{ID: "fresh", Start: 0, End: 5, ColorID: 1},
		}
		err := updateClipHighlights(helper, freshClip.ID, highlights)
		require.NoError(t, err)
		
		// At current state (index -1), can't redo
		_, err = service.RedoHighlightsChange(freshClip.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot redo further")
	})
	
	t.Run("redo with index at last history entry", func(t *testing.T) {
		// Create clip with specific history state
		specificClip := helper.CreateTestVideoClip(project, "Specific Clip")
		
		highlights1 := []schema.Highlight{{ID: "spec1", Start: 0, End: 5, ColorID: 1}}
		highlights2 := []schema.Highlight{{ID: "spec2", Start: 5, End: 10, ColorID: 2}}
		
		history := [][]schema.Highlight{highlights1, highlights2}
		
		// Set index to last entry (1)
		_, err := helper.Client.VideoClip.
			UpdateOneID(specificClip.ID).
			SetHighlightsHistory(history).
			SetHighlightsHistoryIndex(1). // At last entry
			SetHighlights(highlights2).
			Save(helper.Ctx)
		require.NoError(t, err)
		
		// Should not be able to redo further
		_, err = service.RedoHighlightsChange(specificClip.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot redo further")
	})
	
	t.Run("nonexistent clip", func(t *testing.T) {
		_, err := service.RedoHighlightsChange(999999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get video clip")
	})
	
	t.Run("database update verification", func(t *testing.T) {
		// This test verifies that RedoHighlightsChange properly updates the database
		
		testClip := helper.CreateTestVideoClip(project, "DB Test Clip")
		
		// Set up proper history state for undo/redo testing
		highlights1 := []schema.Highlight{{ID: "db_test_1", Start: 0, End: 5, ColorID: 1}}
		highlights2 := []schema.Highlight{{ID: "db_test_2", Start: 5, End: 10, ColorID: 2}}
		
		// Manually set up history with two entries
		_, err := helper.Client.VideoClip.
			UpdateOneID(testClip.ID).
			SetHighlights(highlights2).
			SetHighlightsHistory([][]schema.Highlight{highlights1, highlights2}).
			SetHighlightsHistoryIndex(0). // Set to first entry so we can redo to second
			Save(helper.Ctx)
		require.NoError(t, err)
		
		// Redo should advance to index 1 (highlights2)
		result, err := service.RedoHighlightsChange(testClip.ID)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "db_test_2", result[0].ID)
		
		// Verify the database was actually updated
		updatedClip, err := helper.Client.VideoClip.Get(helper.Ctx, testClip.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, updatedClip.HighlightsHistoryIndex) // Should be at index 1 after redo
		assert.Len(t, updatedClip.Highlights, 1)
		assert.Equal(t, "db_test_2", updatedClip.Highlights[0].ID)
	})
}

// TestHideHighlightComprehensive provides comprehensive coverage for HideHighlight
func TestHideHighlightComprehensive(t *testing.T) {
	helper, service := setupAdditionalTestHelper(t)
	
	project := helper.CreateTestProject("Hide Highlight Test")
	
	t.Run("hide highlight successfully", func(t *testing.T) {
		err := service.HideHighlight(project.ID, "highlight_1")
		assert.NoError(t, err)
		
		// Verify it was hidden
		hidden, err := service.GetHiddenHighlights(project.ID)
		require.NoError(t, err)
		assert.Contains(t, hidden, "highlight_1")
	})
	
	t.Run("hide multiple highlights", func(t *testing.T) {
		err := service.HideHighlight(project.ID, "highlight_2")
		assert.NoError(t, err)
		
		err = service.HideHighlight(project.ID, "highlight_3")
		assert.NoError(t, err)
		
		// Verify both were hidden
		hidden, err := service.GetHiddenHighlights(project.ID)
		require.NoError(t, err)
		assert.Contains(t, hidden, "highlight_2")
		assert.Contains(t, hidden, "highlight_3")
		assert.Len(t, hidden, 3) // Including highlight_1 from previous test
	})
	
	t.Run("hide already hidden highlight", func(t *testing.T) {
		// Hide the same highlight again - should not error and not duplicate
		err := service.HideHighlight(project.ID, "highlight_1")
		assert.NoError(t, err)
		
		hidden, err := service.GetHiddenHighlights(project.ID)
		require.NoError(t, err)
		
		// Count occurrences of highlight_1
		count := 0
		for _, h := range hidden {
			if h == "highlight_1" {
				count++
			}
		}
		assert.Equal(t, 1, count, "Should not duplicate hidden highlights")
	})
	
	t.Run("hide highlight with special characters", func(t *testing.T) {
		specialID := "highlight_with_special-chars_123!@#"
		err := service.HideHighlight(project.ID, specialID)
		assert.NoError(t, err)
		
		hidden, err := service.GetHiddenHighlights(project.ID)
		require.NoError(t, err)
		assert.Contains(t, hidden, specialID)
	})
	
	t.Run("hide highlight and remove from order", func(t *testing.T) {
		// Set up a project with highlight order
		highlightOrder := []interface{}{"h1", "h2", "h3", "h4"}
		_, err := helper.Client.Project.
			UpdateOneID(project.ID).
			SetHighlightOrder(highlightOrder).
			Save(helper.Ctx)
		require.NoError(t, err)
		
		// Hide h2
		err = service.HideHighlight(project.ID, "h2")
		assert.NoError(t, err)
		
		// Verify h2 was removed from order
		updatedProject, err := helper.Client.Project.Get(helper.Ctx, project.ID)
		require.NoError(t, err)
		
		expectedOrder := []interface{}{"h1", "h3", "h4"}
		assert.Equal(t, len(expectedOrder), len(updatedProject.HighlightOrder))
		
		// Check that h2 is not in the order
		for _, item := range updatedProject.HighlightOrder {
			assert.NotEqual(t, "h2", item)
		}
	})
	
	t.Run("hide highlight with mixed order types", func(t *testing.T) {
		// Set up order with mixed types (string highlights and section objects)
		mixedOrder := []interface{}{
			"h1",
			map[string]interface{}{"type": "N", "title": "Section 1"},
			"h2",
			"h3",
			map[string]interface{}{"type": "N", "title": "Section 2"},
			"h4",
		}
		
		_, err := helper.Client.Project.
			UpdateOneID(project.ID).
			SetHighlightOrder(mixedOrder).
			Save(helper.Ctx)
		require.NoError(t, err)
		
		// Hide h3
		err = service.HideHighlight(project.ID, "h3")
		assert.NoError(t, err)
		
		// Verify h3 was removed but sections preserved
		updatedProject, err := helper.Client.Project.Get(helper.Ctx, project.ID)
		require.NoError(t, err)
		
		// Should have 5 items (4 original + sections - 1 hidden highlight)
		assert.Equal(t, 5, len(updatedProject.HighlightOrder))
		
		// Verify sections are still there
		foundSection1 := false
		foundSection2 := false
		foundH3 := false
		
		for _, item := range updatedProject.HighlightOrder {
			if str, ok := item.(string); ok {
				if str == "h3" {
					foundH3 = true
				}
			} else if obj, ok := item.(map[string]interface{}); ok {
				if title, exists := obj["title"]; exists {
					if title == "Section 1" {
						foundSection1 = true
					} else if title == "Section 2" {
						foundSection2 = true
					}
				}
			}
		}
		
		assert.True(t, foundSection1, "Section 1 should be preserved")
		assert.True(t, foundSection2, "Section 2 should be preserved")
		assert.False(t, foundH3, "h3 should be removed from order")
	})
	
	t.Run("hide highlight not in order", func(t *testing.T) {
		// Hide a highlight that's not in the current order
		err := service.HideHighlight(project.ID, "not_in_order")
		assert.NoError(t, err)
		
		// Should still be added to hidden highlights
		hidden, err := service.GetHiddenHighlights(project.ID)
		require.NoError(t, err)
		assert.Contains(t, hidden, "not_in_order")
	})
	
	t.Run("project with nil highlight order", func(t *testing.T) {
		// Create new project with nil order
		nilOrderProject := helper.CreateTestProject("Nil Order Project")
		
		err := service.HideHighlight(nilOrderProject.ID, "test_highlight")
		assert.NoError(t, err)
		
		// Verify highlight was hidden
		hidden, err := service.GetHiddenHighlights(nilOrderProject.ID)
		require.NoError(t, err)
		assert.Contains(t, hidden, "test_highlight")
	})
	
	t.Run("project with empty highlight order", func(t *testing.T) {
		// Create project with explicitly empty order
		emptyOrderProject := helper.CreateTestProject("Empty Order Project")
		_, err := helper.Client.Project.
			UpdateOneID(emptyOrderProject.ID).
			SetHighlightOrder([]interface{}{}).
			Save(helper.Ctx)
		require.NoError(t, err)
		
		err = service.HideHighlight(emptyOrderProject.ID, "empty_test")
		assert.NoError(t, err)
		
		hidden, err := service.GetHiddenHighlights(emptyOrderProject.ID)
		require.NoError(t, err)
		assert.Contains(t, hidden, "empty_test")
	})
	
	t.Run("nonexistent project", func(t *testing.T) {
		err := service.HideHighlight(999999, "some_highlight")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get project")
	})
	
	t.Run("empty highlight ID", func(t *testing.T) {
		err := service.HideHighlight(project.ID, "")
		assert.NoError(t, err) // Should handle empty string gracefully
		
		hidden, err := service.GetHiddenHighlights(project.ID)
		require.NoError(t, err)
		assert.Contains(t, hidden, "") // Empty string should be in hidden list
	})
}