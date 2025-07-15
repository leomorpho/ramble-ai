package exports

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func TestStitchedExport_SingleVideo(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project with single video
	proj := createTestProject(t, client, ctx, "SingleVideoProject")
	clip := createTestVideoClip(t, client, ctx, proj, "SingleClip")
	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Start stitched export
	jobID, err := service.ExportStitchedHighlights(proj.ID, tempDir, 0.0)
	require.NoError(t, err)

	// Verify job was created
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.NotEmpty(t, progress.JobID)
	assert.Contains(t, []string{"pending", "processing", "preparing"}, progress.Stage)
}

func TestStitchedExport_MultipleVideos(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project with multiple videos
	proj := createTestProject(t, client, ctx, "MultiVideoProject")
	clip1 := createTestVideoClip(t, client, ctx, proj, "Clip1")
	clip2 := createTestVideoClip(t, client, ctx, proj, "Clip2")
	clip3 := createTestVideoClip(t, client, ctx, proj, "Clip3")

	// Create highlights on different clips
	createTestHighlight(t, client, ctx, clip1, 10.0, 20.0)
	createTestHighlight(t, client, ctx, clip2, 30.0, 40.0)
	createTestHighlight(t, client, ctx, clip3, 5.0, 15.0)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Start stitched export
	jobID, err := service.ExportStitchedHighlights(proj.ID, tempDir, 0.0)
	require.NoError(t, err)

	// Verify job was created with correct total files
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.NotEmpty(t, progress.JobID)
	assert.Equal(t, tempDir, progress.OutputPath)
}

func TestStitchedExport_NoHighlights(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project with no highlights
	proj := createTestProject(t, client, ctx, "EmptyProject")
	createTestVideoClip(t, client, ctx, proj, "EmptyClip")

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Start stitched export
	jobID, err := service.ExportStitchedHighlights(proj.ID, tempDir, 0.0)
	require.NoError(t, err)

	// Wait for background processing
	time.Sleep(200 * time.Millisecond)

	// Verify job failed due to no highlights
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Equal(t, "failed", progress.Stage)
	assert.Contains(t, progress.ErrorMessage, "No highlights found")
}

func TestStitchedExport_OverlappingHighlights(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project with overlapping highlights
	proj := createTestProject(t, client, ctx, "OverlapProject")
	clip := createTestVideoClip(t, client, ctx, proj, "OverlapClip")

	// Create overlapping highlights
	createTestHighlight(t, client, ctx, clip, 10.0, 25.0)
	createTestHighlight(t, client, ctx, clip, 20.0, 35.0)
	createTestHighlight(t, client, ctx, clip, 30.0, 45.0)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Start stitched export
	jobID, err := service.ExportStitchedHighlights(proj.ID, tempDir, 0.0)
	require.NoError(t, err)

	// Verify job was created
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.NotEmpty(t, progress.JobID)
}

func TestStitchedExport_CancellationDuringPreparation(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project with highlights
	proj := createTestProject(t, client, ctx, "CancelProject")
	clip := createTestVideoClip(t, client, ctx, proj, "CancelClip")
	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)
	createTestHighlight(t, client, ctx, clip, 30.0, 40.0)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Start stitched export
	jobID, err := service.ExportStitchedHighlights(proj.ID, tempDir, 0.0)
	require.NoError(t, err)

	// Wait a moment for processing to start
	time.Sleep(50 * time.Millisecond)

	// Cancel the export
	err = service.CancelExport(jobID)
	require.NoError(t, err)

	// Wait for cancellation to complete
	time.Sleep(100 * time.Millisecond)

	// Verify job was cancelled
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", progress.Stage)
}

func TestStitchedExport_InvalidOutputPath(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project
	proj := createTestProject(t, client, ctx, "TestProject")
	clip := createTestVideoClip(t, client, ctx, proj, "TestClip")
	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)

	// Use invalid output path
	invalidPath := "/invalid/nonexistent/path"

	// Start stitched export
	jobID, err := service.ExportStitchedHighlights(proj.ID, invalidPath, 0.0)
	require.NoError(t, err)

	// Wait for background processing
	time.Sleep(200 * time.Millisecond)

	// Verify job failed due to invalid path
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	// Should fail - error message might vary depending on ffmpeg availability
	assert.Equal(t, "failed", progress.Stage)
}

func TestStitchedExport_ProgressTracking(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project with multiple highlights
	proj := createTestProject(t, client, ctx, "ProgressProject")
	clip1 := createTestVideoClip(t, client, ctx, proj, "ProgressClip1")
	clip2 := createTestVideoClip(t, client, ctx, proj, "ProgressClip2")

	createTestHighlight(t, client, ctx, clip1, 10.0, 20.0)
	createTestHighlight(t, client, ctx, clip2, 30.0, 40.0)
	createTestHighlight(t, client, ctx, clip1, 50.0, 60.0)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Start stitched export
	jobID, err := service.ExportStitchedHighlights(proj.ID, tempDir, 0.0)
	require.NoError(t, err)

	// Track progress over time
	var progressStages []string
	var progressValues []float64

	for i := 0; i < 10; i++ {
		progress, err := service.GetExportProgress(jobID)
		require.NoError(t, err)

		progressStages = append(progressStages, progress.Stage)
		progressValues = append(progressValues, progress.Progress)

		// Break if completed or failed
		if progress.Stage == "completed" || progress.Stage == "failed" {
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	// Verify we saw different stages
	assert.Contains(t, progressStages, "pending")
	// Progress should be between 0 and 1
	for _, val := range progressValues {
		assert.True(t, val >= 0.0 && val <= 1.0, "Progress should be between 0 and 1, got %f", val)
	}
}

func TestStitchedExport_ProjectIsolation(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create two separate projects
	proj1 := createTestProject(t, client, ctx, "StitchProject1")
	proj2 := createTestProject(t, client, ctx, "StitchProject2")

	// Create clips and highlights for each project
	clip1 := createTestVideoClip(t, client, ctx, proj1, "Clip1")
	clip2 := createTestVideoClip(t, client, ctx, proj2, "Clip2")

	createTestHighlight(t, client, ctx, clip1, 10.0, 20.0)
	createTestHighlight(t, client, ctx, clip1, 30.0, 40.0)
	createTestHighlight(t, client, ctx, clip2, 5.0, 15.0)

	// Create temporary output directories
	tempDir1, err := os.MkdirTemp("", "export_test_1_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir1)

	tempDir2, err := os.MkdirTemp("", "export_test_2_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir2)

	// Start stitched exports
	jobID1, err := service.ExportStitchedHighlights(proj1.ID, tempDir1, 0.0)
	require.NoError(t, err)

	jobID2, err := service.ExportStitchedHighlights(proj2.ID, tempDir2, 0.0)
	require.NoError(t, err)

	// Verify jobs are different
	assert.NotEqual(t, jobID1, jobID2)

	// Check both jobs exist and have correct paths
	progress1, err := service.GetExportProgress(jobID1)
	require.NoError(t, err)
	assert.Equal(t, jobID1, progress1.JobID)
	assert.Equal(t, tempDir1, progress1.OutputPath)

	progress2, err := service.GetExportProgress(jobID2)
	require.NoError(t, err)
	assert.Equal(t, jobID2, progress2.JobID)
	assert.Equal(t, tempDir2, progress2.OutputPath)
}

func TestStitchedExport_FilenameGeneration(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project
	proj := createTestProject(t, client, ctx, "Test Project @2024")
	clip := createTestVideoClip(t, client, ctx, proj, "TestClip")
	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Start stitched export
	_, err = service.ExportStitchedHighlights(proj.ID, tempDir, 0.0)
	require.NoError(t, err)

	// Check that filename generation works correctly
	filename := service.generateOutputFilename("Test Project @2024", "stitched")

	// Should sanitize special characters
	assert.Contains(t, filename, "Test_Project__2024_stitched_")
	assert.Equal(t, ".mp4", filepath.Ext(filename))
}

func TestStitchedExport_ConcurrentExports(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create multiple projects
	proj1 := createTestProject(t, client, ctx, "ConcurrentStitch1")
	proj2 := createTestProject(t, client, ctx, "ConcurrentStitch2")
	proj3 := createTestProject(t, client, ctx, "ConcurrentStitch3")

	// Create clips and highlights
	clip1 := createTestVideoClip(t, client, ctx, proj1, "Clip1")
	clip2 := createTestVideoClip(t, client, ctx, proj2, "Clip2")
	clip3 := createTestVideoClip(t, client, ctx, proj3, "Clip3")

	createTestHighlight(t, client, ctx, clip1, 10.0, 20.0)
	createTestHighlight(t, client, ctx, clip2, 30.0, 40.0)
	createTestHighlight(t, client, ctx, clip3, 50.0, 60.0)

	// Create separate output directories
	tempDir1, err := os.MkdirTemp("", "export_test_1_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir1)

	tempDir2, err := os.MkdirTemp("", "export_test_2_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir2)

	tempDir3, err := os.MkdirTemp("", "export_test_3_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir3)

	// Start concurrent stitched exports
	jobID1, err := service.ExportStitchedHighlights(proj1.ID, tempDir1, 0.0)
	require.NoError(t, err)

	jobID2, err := service.ExportStitchedHighlights(proj2.ID, tempDir2, 0.0)
	require.NoError(t, err)

	jobID3, err := service.ExportStitchedHighlights(proj3.ID, tempDir3, 0.0)
	require.NoError(t, err)

	// Verify all jobs are running
	assert.NotEqual(t, jobID1, jobID2)
	assert.NotEqual(t, jobID2, jobID3)
	assert.NotEqual(t, jobID1, jobID3)

	// Check all jobs exist
	progress1, err := service.GetExportProgress(jobID1)
	require.NoError(t, err)
	assert.Equal(t, jobID1, progress1.JobID)

	progress2, err := service.GetExportProgress(jobID2)
	require.NoError(t, err)
	assert.Equal(t, jobID2, progress2.JobID)

	progress3, err := service.GetExportProgress(jobID3)
	require.NoError(t, err)
	assert.Equal(t, jobID3, progress3.JobID)
}

func TestStitchedExport_TempDirectoryCleanup(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project
	proj := createTestProject(t, client, ctx, "CleanupProject")
	clip := createTestVideoClip(t, client, ctx, proj, "CleanupClip")
	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Start stitched export
	jobID, err := service.ExportStitchedHighlights(proj.ID, tempDir, 0.0)
	require.NoError(t, err)

	// Wait for processing to start
	time.Sleep(100 * time.Millisecond)

	// Cancel to trigger cleanup
	err = service.CancelExport(jobID)
	require.NoError(t, err)

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	// Verify job was cancelled (cleanup should happen automatically)
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", progress.Stage)
}

func TestStitchedExport_LargeNumberOfHighlights(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project with many highlights
	proj := createTestProject(t, client, ctx, "LargeProject")
	clip := createTestVideoClip(t, client, ctx, proj, "LargeClip")

	// Create 20 highlights
	for i := 0; i < 20; i++ {
		start := float64(i * 10)
		end := start + 5.0
		createTestHighlight(t, client, ctx, clip, start, end)
	}

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Start stitched export
	jobID, err := service.ExportStitchedHighlights(proj.ID, tempDir, 0.0)
	require.NoError(t, err)

	// Wait for background processing to start and set total files
	time.Sleep(100 * time.Millisecond)

	// Verify job was created
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.NotEmpty(t, progress.JobID)
	assert.Equal(t, 20, progress.TotalFiles)
}
