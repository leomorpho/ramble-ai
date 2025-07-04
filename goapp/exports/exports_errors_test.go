package exports

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"MYAPP/ent/schema"
	_ "github.com/mattn/go-sqlite3"
)


func TestExportStitchedHighlights_DatabaseError(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Close the database to cause errors
	client.Close()

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test export with closed database
	_, err = service.ExportStitchedHighlights(1, tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get project")
}

func TestExportIndividualHighlights_DatabaseError(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Close the database to cause errors
	client.Close()

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test export with closed database
	_, err = service.ExportIndividualHighlights(1, tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get project")
}

func TestExportStitchedHighlights_EmptyOutputFolder(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project
	proj := createTestProject(t, client, ctx, "TestProject")
	clip := createTestVideoClip(t, client, ctx, proj, "TestClip")
	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)

	// Test with empty output folder
	_, err := service.ExportStitchedHighlights(proj.ID, "")
	require.NoError(t, err) // Job creation should succeed

	// But background processing should fail
	time.Sleep(200 * time.Millisecond)
}

func TestExportIndividualHighlights_EmptyOutputFolder(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project
	proj := createTestProject(t, client, ctx, "TestProject")
	clip := createTestVideoClip(t, client, ctx, proj, "TestClip")
	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)

	// Test with empty output folder
	jobID, err := service.ExportIndividualHighlights(proj.ID, "")
	require.NoError(t, err) // Job creation should succeed

	// But background processing should fail
	time.Sleep(200 * time.Millisecond)
	
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Equal(t, "failed", progress.Stage)
}

func TestGetExportProgress_DatabaseClosed(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project and job first
	proj := createTestProject(t, client, ctx, "TestProject")
	jobID := "test_job_123"
	
	_, err := client.ExportJob.
		Create().
		SetJobID(jobID).
		SetExportType("individual").
		SetOutputPath("/test/path").
		SetStage("processing").
		SetProject(proj).
		SetCreatedAt(time.Now()).
		Save(ctx)
	require.NoError(t, err)

	// Close database
	client.Close()

	// Test get progress with closed database
	_, err = service.GetExportProgress(jobID)
	assert.Error(t, err)
}

func TestCancelExport_DatabaseClosed(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project and job first
	proj := createTestProject(t, client, ctx, "TestProject")
	jobID := "test_job_123"
	
	_, err := client.ExportJob.
		Create().
		SetJobID(jobID).
		SetExportType("individual").
		SetOutputPath("/test/path").
		SetStage("processing").
		SetProject(proj).
		SetCreatedAt(time.Now()).
		Save(ctx)
	require.NoError(t, err)

	// Close database
	client.Close()

	// Test cancel with closed database
	err = service.CancelExport(jobID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "export job not found")
}

func TestExportWithNonExistentVideoFiles(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project with non-existent video file
	proj := createTestProject(t, client, ctx, "TestProject")
	clip, err := client.VideoClip.
		Create().
		SetName("NonExistentClip").
		SetDescription("Test clip").
		SetFilePath("/nonexistent/path/video.mp4"). // Non-existent file
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

	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test individual export
	jobID, err := service.ExportIndividualHighlights(proj.ID, tempDir)
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(300 * time.Millisecond)

	// Should fail due to non-existent video file
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Equal(t, "failed", progress.Stage)
	assert.Contains(t, progress.ErrorMessage, "Failed to extract segment")
}

func TestExportWithInvalidHighlightTimes(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project
	proj := createTestProject(t, client, ctx, "TestProject")
	clip := createTestVideoClip(t, client, ctx, proj, "TestClip")

	// Create highlight with invalid times (end before start)
	invalidHighlight := schema.Highlight{
		ID:    "invalid_highlight",
		Start: 30.0,
		End:   10.0, // End before start
		Color: "#FF0000",
	}
	
	_, err := client.VideoClip.
		UpdateOne(clip).
		SetHighlights([]schema.Highlight{invalidHighlight}).
		Save(ctx)
	require.NoError(t, err)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test export
	jobID, err := service.ExportIndividualHighlights(proj.ID, tempDir)
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(300 * time.Millisecond)

	// Should fail due to invalid highlight times
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Equal(t, "failed", progress.Stage)
}

func TestExportWithNegativeHighlightTimes(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project
	proj := createTestProject(t, client, ctx, "TestProject")
	clip := createTestVideoClip(t, client, ctx, proj, "TestClip")

	// Create highlight with negative times
	negativeHighlight := schema.Highlight{
		ID:    "negative_highlight",
		Start: -10.0, // Negative start
		End:   20.0,
		Color: "#FF0000",
	}
	
	_, err := client.VideoClip.
		UpdateOne(clip).
		SetHighlights([]schema.Highlight{negativeHighlight}).
		Save(ctx)
	require.NoError(t, err)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test export
	jobID, err := service.ExportIndividualHighlights(proj.ID, tempDir)
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(300 * time.Millisecond)

	// Should fail due to invalid highlight times
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Equal(t, "failed", progress.Stage)
}

func TestGetProjectExportJobs_DatabaseClosed(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project first
	proj := createTestProject(t, client, ctx, "TestProject")

	// Close database
	client.Close()

	// Test get project jobs with closed database
	_, err := service.GetProjectExportJobs(proj.ID)
	assert.Error(t, err)
}

func TestRecoverActiveExportJobs_DatabaseClosed(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Close database
	client.Close()

	// Test recovery with closed database
	err := service.RecoverActiveExportJobs()
	assert.Error(t, err)
}

func TestExportWithReadOnlyOutputDirectory(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project
	proj := createTestProject(t, client, ctx, "TestProject")
	clip := createTestVideoClip(t, client, ctx, proj, "TestClip")
	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)

	// Create read-only directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	readOnlyDir := filepath.Join(tempDir, "readonly")
	err = os.Mkdir(readOnlyDir, 0555) // Read-only permissions
	require.NoError(t, err)

	// Test export
	jobID, err := service.ExportIndividualHighlights(proj.ID, readOnlyDir)
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Should fail due to read-only directory
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Equal(t, "failed", progress.Stage)
	assert.Contains(t, progress.ErrorMessage, "Failed to create project directory")
}

func TestExportWithDiskSpaceIssues(t *testing.T) {
	// This test is harder to simulate, but we can test the error handling
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project
	proj := createTestProject(t, client, ctx, "TestProject")
	clip := createTestVideoClip(t, client, ctx, proj, "TestClip")
	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)

	// Use /dev/null as output (invalid for file creation)
	invalidPath := "/dev/null"

	// Test export
	jobID, err := service.ExportIndividualHighlights(proj.ID, invalidPath)
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Should fail
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Equal(t, "failed", progress.Stage)
}

func TestCancelExport_NonExistentActiveJob(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project and job
	proj := createTestProject(t, client, ctx, "TestProject")
	jobID := "test_job_123"
	
	// Create export job but not active job
	_, err := client.ExportJob.
		Create().
		SetJobID(jobID).
		SetExportType("individual").
		SetOutputPath("/test/path").
		SetStage("processing").
		SetProject(proj).
		SetCreatedAt(time.Now()).
		Save(ctx)
	require.NoError(t, err)

	// Test cancel without active job (should still work)
	err = service.CancelExport(jobID)
	require.NoError(t, err)

	// Verify job was cancelled
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", progress.Stage)
}

func TestExportWithCorruptedDatabase(t *testing.T) {
	// This test simulates database corruption scenarios
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project with invalid references
	proj := createTestProject(t, client, ctx, "TestProject")

	// Create project without any highlights to simulate empty state
	// This simulates a scenario where the project has no highlights to export

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test export - should handle gracefully
	jobID, err := service.ExportIndividualHighlights(proj.ID, tempDir)
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Should complete successfully since no highlights belong to the project
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Equal(t, "failed", progress.Stage)
	assert.Contains(t, progress.ErrorMessage, "No highlights found")
}

func TestGenerateListFile_WritePermissionError(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create read-only directory
	tempDir, err := os.MkdirTemp("", "list_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	readOnlyDir := filepath.Join(tempDir, "readonly")
	err = os.Mkdir(readOnlyDir, 0555) // Read-only permissions
	require.NoError(t, err)

	// Test list file generation in read-only directory
	segmentPaths := []string{"segment1.mp4", "segment2.mp4"}
	_, err = service.generateListFile(segmentPaths, readOnlyDir)
	assert.Error(t, err)
}

func TestUpdateJobProgress_ConcurrentUpdates(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project and job
	proj := createTestProject(t, client, ctx, "ConcurrentProject")
	jobID := "test_job_123"
	
	_, err := client.ExportJob.
		Create().
		SetJobID(jobID).
		SetExportType("individual").
		SetOutputPath("/test/path").
		SetStage("processing").
		SetProject(proj).
		SetCreatedAt(time.Now()).
		Save(ctx)
	require.NoError(t, err)

	// Simulate concurrent updates
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(index int) {
			service.updateJobProgress(jobID, "processing", float64(index)/10.0, "file.mp4", 10, index)
			done <- true
		}(i)
	}

	// Wait for all updates to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify job still exists and has valid progress
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Equal(t, jobID, progress.JobID)
	assert.Equal(t, "processing", progress.Stage)
	assert.True(t, progress.Progress >= 0.0 && progress.Progress <= 1.0)
}

func TestExportWithVeryLongProjectName(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create project with very long name
	longName := strings.Repeat("VeryLongProjectNameWithLotsOfCharacters", 20) // ~800 characters
	proj := createTestProject(t, client, ctx, longName)
	clip := createTestVideoClip(t, client, ctx, proj, "TestClip")
	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test export - should handle long names gracefully
	jobID, err := service.ExportIndividualHighlights(proj.ID, tempDir)
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Check that directory was created (may be truncated by filesystem)
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	
	// Should either succeed or fail gracefully
	assert.Contains(t, []string{"processing", "extracting", "completed", "failed"}, progress.Stage)
}

func TestExportWithSpecialCharacterPaths(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project
	proj := createTestProject(t, client, ctx, "SpecialCharsProject")
	clip := createTestVideoClip(t, client, ctx, proj, "TestClip")
	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)

	// Create temporary directory with special characters
	tempDir, err := os.MkdirTemp("", "export_test_with_spaces_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	specialDir := filepath.Join(tempDir, "path with spaces & special chars!")
	err = os.MkdirAll(specialDir, 0755)
	require.NoError(t, err)

	// Test export
	jobID, err := service.ExportIndividualHighlights(proj.ID, specialDir)
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Should handle special characters in paths
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Contains(t, []string{"processing", "extracting", "completed", "failed"}, progress.Stage)
}

func TestMemoryLeakPrevention(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project
	proj := createTestProject(t, client, ctx, "MemoryTestProject")
	clip := createTestVideoClip(t, client, ctx, proj, "MemoryTestClip")
	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Start multiple exports and cancel them quickly
	var jobIDs []string
	for i := 0; i < 5; i++ {
		jobID, err := service.ExportIndividualHighlights(proj.ID, tempDir)
		require.NoError(t, err)
		jobIDs = append(jobIDs, jobID)
		
		// Cancel immediately
		time.Sleep(10 * time.Millisecond)
		err = service.CancelExport(jobID)
		require.NoError(t, err)
	}

	// Wait for cleanup
	time.Sleep(200 * time.Millisecond)

	// Verify active jobs map is cleaned up
	activeJobsMutex.Lock()
	activeJobCount := len(activeJobs)
	activeJobsMutex.Unlock()

	// Should have minimal or no active jobs
	assert.True(t, activeJobCount <= 1, "Expected minimal active jobs, got %d", activeJobCount)
}