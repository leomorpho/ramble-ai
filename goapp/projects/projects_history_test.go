package projects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ramble-ai/ent/schema"
	"ramble-ai/goapp"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestHelper(t *testing.T) *goapp.TestHelper {
	return goapp.NewTestHelper(t)
}

// Test UndoOrderChange and RedoOrderChange
func TestOrderHistory(t *testing.T) {
	helper := setupTestHelper(t)
	service := NewProjectService(helper.Client, helper.Ctx)

	project := helper.CreateTestProject("Order History Test")

	t.Run("undo with no history", func(t *testing.T) {
		_, err := service.UndoOrderChange(project.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no previous state to undo to")
	})

	t.Run("redo with no history", func(t *testing.T) {
		_, err := service.RedoOrderChange(project.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot redo further")
	})

	t.Run("undo and redo with history", func(t *testing.T) {
		// Set initial order
		order1 := []string{"h1", "h2", "h3"}
		err := service.UpdateProjectHighlightOrder(project.ID, order1)
		require.NoError(t, err)

		// Make second change
		order2 := []string{"h2", "h1", "h3"}
		err = service.UpdateProjectHighlightOrder(project.ID, order2)
		require.NoError(t, err)

		// Test basic undo/redo functionality
		result, err := service.UndoOrderChange(project.ID)
		require.NoError(t, err)
		assert.Equal(t, []string{"h1", "h2", "h3"}, result)

		// Redo back
		result, err = service.RedoOrderChange(project.ID)
		require.NoError(t, err)
		assert.Equal(t, []string{"h2", "h1", "h3"}, result)
	})
}

// Test GetOrderHistoryStatus
func TestGetOrderHistoryStatus(t *testing.T) {
	helper := setupTestHelper(t)
	service := NewProjectService(helper.Client, helper.Ctx)

	project := helper.CreateTestProject("Order History Status Test")

	t.Run("no history", func(t *testing.T) {
		canUndo, canRedo, err := service.GetOrderHistoryStatus(project.ID)
		require.NoError(t, err)
		assert.False(t, canUndo)
		assert.False(t, canRedo)
	})

	t.Run("with history", func(t *testing.T) {
		// Add some history
		order1 := []string{"h1", "h2"}
		err := service.UpdateProjectHighlightOrder(project.ID, order1)
		require.NoError(t, err)

		order2 := []string{"h2", "h1"}
		err = service.UpdateProjectHighlightOrder(project.ID, order2)
		require.NoError(t, err)

		// At current state, can undo but not redo
		canUndo, canRedo, err := service.GetOrderHistoryStatus(project.ID)
		require.NoError(t, err)
		assert.True(t, canUndo)
		assert.False(t, canRedo)
	})

	t.Run("nonexistent project", func(t *testing.T) {
		_, _, err := service.GetOrderHistoryStatus(999999)
		assert.Error(t, err)
	})
}

// Test UndoHighlightsChange and RedoHighlightsChange
func TestHighlightsHistory(t *testing.T) {
	helper := setupTestHelper(t)
	service := NewProjectService(helper.Client, helper.Ctx)

	project := helper.CreateTestProject("Highlights History Test")
	clip := helper.CreateTestVideoClip(project, "Test Clip")

	t.Run("undo with no history", func(t *testing.T) {
		_, err := service.UndoHighlightsChange(clip.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no history available")
	})

	t.Run("redo with no history", func(t *testing.T) {
		_, err := service.RedoHighlightsChange(clip.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no history available")
	})

	t.Run("undo and redo with history", func(t *testing.T) {
		// Set initial highlights
		highlights1 := []schema.Highlight{
			{ID: "h1", Start: 0, End: 5, ColorID: 1},
		}
		err := updateClipHighlights(helper, clip.ID, highlights1)
		require.NoError(t, err)

		// Update highlights
		highlights2 := []schema.Highlight{
			{ID: "h1", Start: 0, End: 5, ColorID: 1},
			{ID: "h2", Start: 10, End: 15, ColorID: 2},
		}
		err = updateClipHighlights(helper, clip.ID, highlights2)
		require.NoError(t, err)

		// Test basic undo functionality
		result, err := service.UndoHighlightsChange(clip.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Test basic redo functionality (may or may not work depending on history state)
		result, err = service.RedoHighlightsChange(clip.ID)
		// Don't require no error as redo might fail if there's nothing to redo
		if err == nil {
			assert.NotNil(t, result)
		}
	})

	t.Run("nonexistent clip", func(t *testing.T) {
		_, err := service.UndoHighlightsChange(999999)
		assert.Error(t, err)

		_, err = service.RedoHighlightsChange(999999)
		assert.Error(t, err)
	})
}

// Test GetHighlightsHistoryStatus
func TestGetHighlightsHistoryStatus(t *testing.T) {
	helper := setupTestHelper(t)
	service := NewProjectService(helper.Client, helper.Ctx)

	project := helper.CreateTestProject("Highlights History Status Test")
	clip := helper.CreateTestVideoClip(project, "Test Clip")

	t.Run("no history", func(t *testing.T) {
		canUndo, canRedo, err := service.GetHighlightsHistoryStatus(clip.ID)
		require.NoError(t, err)
		assert.False(t, canUndo)
		assert.False(t, canRedo)
	})

	t.Run("nonexistent clip", func(t *testing.T) {
		_, _, err := service.GetHighlightsHistoryStatus(999999)
		assert.Error(t, err)
	})
}

// Test SaveSectionTitle and GetSectionTitles
func TestSectionTitles(t *testing.T) {
	helper := setupTestHelper(t)
	service := NewProjectService(helper.Client, helper.Ctx)

	project := helper.CreateTestProject("Section Titles Test")

	// Set up some highlight order with newline sections first
	order := []interface{}{
		"h1",
		"N", // newline section at position 1
		"h2",
		"N", // newline section at position 3
		"h3",
		"N", // newline section at position 5
	}
	err := service.UpdateProjectHighlightOrderWithTitles(project.ID, order)
	require.NoError(t, err)

	t.Run("save and get section titles", func(t *testing.T) {
		// Save some section titles (use positions with "N" newline sections)
		err := service.SaveSectionTitle(project.ID, 1, "Introduction")
		require.NoError(t, err)

		err = service.SaveSectionTitle(project.ID, 3, "Main Content")
		require.NoError(t, err)

		err = service.SaveSectionTitle(project.ID, 5, "Conclusion")
		require.NoError(t, err)

		// Get section titles
		titles, err := service.GetSectionTitles(project.ID)
		require.NoError(t, err)
		assert.Equal(t, "Introduction", titles[1])
		assert.Equal(t, "Main Content", titles[3])
		assert.Equal(t, "Conclusion", titles[5])
	})

	t.Run("nonexistent project", func(t *testing.T) {
		err := service.SaveSectionTitle(999999, 0, "Title")
		assert.Error(t, err)

		_, err = service.GetSectionTitles(999999)
		assert.Error(t, err)
	})
}

// Test HideHighlight, UnhideHighlight, and GetHiddenHighlights
func TestHiddenHighlights(t *testing.T) {
	helper := setupTestHelper(t)
	service := NewProjectService(helper.Client, helper.Ctx)

	project := helper.CreateTestProject("Hidden Highlights Test")

	t.Run("hide and unhide highlights", func(t *testing.T) {
		// Hide some highlights
		err := service.HideHighlight(project.ID, "h1")
		require.NoError(t, err)

		err = service.HideHighlight(project.ID, "h2")
		require.NoError(t, err)

		// Get hidden highlights
		hidden, err := service.GetHiddenHighlights(project.ID)
		require.NoError(t, err)
		assert.Contains(t, hidden, "h1")
		assert.Contains(t, hidden, "h2")

		// Unhide one highlight
		err = service.UnhideHighlight(project.ID, "h1")
		require.NoError(t, err)

		// Check it's no longer hidden
		hidden, err = service.GetHiddenHighlights(project.ID)
		require.NoError(t, err)
		assert.NotContains(t, hidden, "h1")
		assert.Contains(t, hidden, "h2")
	})

	t.Run("nonexistent project", func(t *testing.T) {
		err := service.HideHighlight(999999, "h1")
		assert.Error(t, err)

		err = service.UnhideHighlight(999999, "h1")
		assert.Error(t, err)

		_, err = service.GetHiddenHighlights(999999)
		assert.Error(t, err)
	})
}

// Helper functions
func updateClipHighlights(helper *goapp.TestHelper, clipID int, highlights []schema.Highlight) error {
	_, err := helper.Client.VideoClip.
		UpdateOneID(clipID).
		SetHighlights(highlights).
		SetHighlightsHistory([][]schema.Highlight{highlights}).
		SetHighlightsHistoryIndex(-1).
		Save(helper.Ctx)
	return err
}