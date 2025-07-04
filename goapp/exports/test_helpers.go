package exports

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"MYAPP/ent"
	"MYAPP/ent/enttest"
	"MYAPP/ent/schema"
	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates a test database for testing
func setupTestDB(t testing.TB) (*ent.Client, context.Context) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	ctx := context.Background()
	return client, ctx
}

// createTestProject creates a test project
func createTestProject(t testing.TB, client *ent.Client, ctx context.Context, name string) *ent.Project {
	proj, err := client.Project.
		Create().
		SetName(name).
		SetDescription("Test project").
		SetPath("/test/path").
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	require.NoError(t, err)
	return proj
}

// createTestVideoClip creates a test video clip
func createTestVideoClip(t testing.TB, client *ent.Client, ctx context.Context, proj *ent.Project, name string) *ent.VideoClip {
	clip, err := client.VideoClip.
		Create().
		SetName(name).
		SetDescription("Test clip").
		SetFilePath("/test/video.mp4").
		SetFileSize(1000000).
		SetDuration(60.0).
		SetFormat("mp4").
		SetWidth(1920).
		SetHeight(1080).
		SetProject(proj).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	require.NoError(t, err)
	return clip
}

// createTestHighlight creates a test highlight on a video clip
func createTestHighlight(t testing.TB, client *ent.Client, ctx context.Context, clip *ent.VideoClip, start, end float64) string {
	highlightID := fmt.Sprintf("h_%d", time.Now().UnixNano())
	
	// Get existing highlights
	existingHighlights := clip.Highlights
	
	// Add new highlight
	newHighlight := schema.Highlight{
		ID:    highlightID,
		Start: start,
		End:   end,
		Color: "#FF0000",
	}
	
	updatedHighlights := append(existingHighlights, newHighlight)
	
	// Update video clip with new highlights
	_, err := client.VideoClip.
		UpdateOne(clip).
		SetHighlights(updatedHighlights).
		Save(ctx)
	require.NoError(t, err)
	
	return highlightID
}