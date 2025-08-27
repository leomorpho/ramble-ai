package exports

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"ramble-ai/ent"
	"ramble-ai/ent/enttest"
	"ramble-ai/ent/migrate"
	"ramble-ai/ent/schema"
	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates a test database for testing
func setupTestDB(t testing.TB) (*ent.Client, context.Context) {
	// Clean up any active jobs from previous tests
	cleanupActiveJobs()

	// Use unique database name per test to avoid sharing issues
	dbName := fmt.Sprintf("file:ent_%d?mode=memory&cache=shared&_fk=1&_journal_mode=WAL&_busy_timeout=5000", time.Now().UnixNano())
	client := enttest.Open(t, "sqlite3", dbName)
	ctx := context.Background()

	// Ensure all migrations are run - use forceful recreation
	err := client.Schema.Create(ctx, migrate.WithGlobalUniqueID(true))
	require.NoError(t, err)

	// Set up cleanup for when the test completes
	t.Cleanup(func() {
		// Wait for active jobs to complete before cleanup
		waitForActiveJobsToComplete(t, 2*time.Second)
		cleanupActiveJobs()
		// Give more time for background goroutines to finish
		time.Sleep(100 * time.Millisecond)
	})

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

	// Get fresh clip data to ensure we have latest highlights
	freshClip, err := client.VideoClip.Get(ctx, clip.ID)
	require.NoError(t, err)

	// Get existing highlights
	existingHighlights := freshClip.Highlights

	// Add new highlight
	newHighlight := schema.Highlight{
		ID:      highlightID,
		Start:   start,
		End:     end,
		ColorID: 3, // Red
	}

	updatedHighlights := append(existingHighlights, newHighlight)

	// Update video clip with new highlights
	_, err = client.VideoClip.
		UpdateOne(freshClip).
		SetHighlights(updatedHighlights).
		Save(ctx)
	require.NoError(t, err)

	return highlightID
}

// cleanupActiveJobs clears the global activeJobs map to prevent test interference
func cleanupActiveJobs() {
	activeJobsMutex.Lock()
	defer activeJobsMutex.Unlock()

	// Cancel any active jobs gracefully
	for jobID, activeJob := range activeJobs {
		if activeJob.IsActive && activeJob.Cancel != nil {
			// Send cancel signal without closing channel to avoid panics
			select {
			case activeJob.Cancel <- true:
				// Signal sent successfully
			default:
				// Channel was full or closed, skip
			}
		}
		delete(activeJobs, jobID)
	}

	// Clear the map completely
	activeJobs = make(map[string]*ActiveExportJob)
}

// waitForActiveJobsToComplete waits for all active jobs to complete or timeout
func waitForActiveJobsToComplete(t testing.TB, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		activeJobsMutex.RLock()
		activeCount := len(activeJobs)
		activeJobsMutex.RUnlock()
		
		if activeCount == 0 {
			return
		}
		
		// Cancel all active jobs to speed up cleanup
		activeJobsMutex.RLock()
		for _, job := range activeJobs {
			if job.Cancel != nil {
				select {
				case job.Cancel <- true:
				default:
					// Channel might be full or closed, continue
				}
			}
		}
		activeJobsMutex.RUnlock()
		
		time.Sleep(100 * time.Millisecond)
	}
	
	// Force cleanup regardless of timeout
	cleanupActiveJobs()
}
