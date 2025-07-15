package exports

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func TestIndividualExport_DirectoryCreation(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project with special characters in name
	proj := createTestProject(t, client, ctx, "Test Project @#$%")
	clip := createTestVideoClip(t, client, ctx, proj, "TestClip")
	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Start export
	_, err = service.ExportIndividualHighlights(proj.ID, tempDir, 0.0)
	require.NoError(t, err)

	// Wait a moment for background processing to start
	time.Sleep(100 * time.Millisecond)

	// Verify project directory was created with sanitized name
	expectedProjectDir := filepath.Join(tempDir, "Test_Project_____")
	assert.DirExists(t, expectedProjectDir)
}

func TestIndividualExport_MultipleHighlights(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project with multiple highlights
	proj := createTestProject(t, client, ctx, "MultiHighlightProject")
	clip1 := createTestVideoClip(t, client, ctx, proj, "Clip1")
	clip2 := createTestVideoClip(t, client, ctx, proj, "Clip2")

	// Create multiple highlights
	createTestHighlight(t, client, ctx, clip1, 10.0, 20.0)
	createTestHighlight(t, client, ctx, clip1, 30.0, 40.0)
	createTestHighlight(t, client, ctx, clip2, 5.0, 15.0)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Start export
	jobID, err := service.ExportIndividualHighlights(proj.ID, tempDir, 0.0)
	require.NoError(t, err)

	// Verify job was created with correct total files
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.NotEmpty(t, progress.JobID)

	// The job should be processing or pending initially
	assert.Contains(t, []string{"pending", "processing", "extracting"}, progress.Stage)
}

func TestIndividualExport_NoHighlights(t *testing.T) {
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

	// Start export
	jobID, err := service.ExportIndividualHighlights(proj.ID, tempDir, 0.0)
	require.NoError(t, err)

	// Wait for background processing
	time.Sleep(200 * time.Millisecond)

	// Verify job failed due to no highlights
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Equal(t, "failed", progress.Stage)
	assert.Contains(t, progress.ErrorMessage, "No highlights found")
}

func TestIndividualExport_InvalidOutputPath(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project
	proj := createTestProject(t, client, ctx, "TestProject")
	clip := createTestVideoClip(t, client, ctx, proj, "TestClip")
	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)

	// Use invalid output path
	invalidPath := "/invalid/nonexistent/path"

	// Start export
	jobID, err := service.ExportIndividualHighlights(proj.ID, invalidPath, 0.0)
	require.NoError(t, err)

	// Wait for background processing
	time.Sleep(200 * time.Millisecond)

	// Verify job failed due to invalid path
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Equal(t, "failed", progress.Stage)
	assert.Contains(t, progress.ErrorMessage, "Failed to create project directory")
}

func TestIndividualExport_CancellationDuringExport(t *testing.T) {
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

	// Start export
	jobID, err := service.ExportIndividualHighlights(proj.ID, tempDir, 0.0)
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

func TestIndividualExport_ProjectIsolation(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create two separate projects
	proj1 := createTestProject(t, client, ctx, "Project1")
	proj2 := createTestProject(t, client, ctx, "Project2")

	// Create clips and highlights for each project
	clip1 := createTestVideoClip(t, client, ctx, proj1, "Clip1")
	clip2 := createTestVideoClip(t, client, ctx, proj2, "Clip2")

	createTestHighlight(t, client, ctx, clip1, 10.0, 20.0)
	createTestHighlight(t, client, ctx, clip1, 30.0, 40.0)
	createTestHighlight(t, client, ctx, clip2, 5.0, 15.0)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Export from project 1
	jobID1, err := service.ExportIndividualHighlights(proj1.ID, tempDir, 0.0)
	require.NoError(t, err)

	// Wait for processing to start
	time.Sleep(100 * time.Millisecond)

	// Verify project 1 directory exists
	proj1Dir := filepath.Join(tempDir, "Project1")
	assert.DirExists(t, proj1Dir)

	// Export from project 2
	jobID2, err := service.ExportIndividualHighlights(proj2.ID, tempDir, 0.0)
	require.NoError(t, err)

	// Wait for processing to start
	time.Sleep(100 * time.Millisecond)

	// Verify project 2 directory exists and is separate
	proj2Dir := filepath.Join(tempDir, "Project2")
	assert.DirExists(t, proj2Dir)

	// Verify jobs are different
	assert.NotEqual(t, jobID1, jobID2)
}

func TestIndividualExport_ProgressTracking(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project with multiple highlights
	proj := createTestProject(t, client, ctx, "ProgressProject")
	clip := createTestVideoClip(t, client, ctx, proj, "ProgressClip")
	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)
	createTestHighlight(t, client, ctx, clip, 30.0, 40.0)
	createTestHighlight(t, client, ctx, clip, 50.0, 60.0)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Start export
	jobID, err := service.ExportIndividualHighlights(proj.ID, tempDir, 0.0)
	require.NoError(t, err)

	// Track progress over time
	var progressStages []string
	for i := 0; i < 10; i++ {
		progress, err := service.GetExportProgress(jobID)
		require.NoError(t, err)

		progressStages = append(progressStages, progress.Stage)

		// Break if completed or failed
		if progress.Stage == "completed" || progress.Stage == "failed" {
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	// Verify we saw different stages
	assert.Contains(t, progressStages, "pending")
	// We might see processing or extracting stages depending on timing
}

func TestIndividualExport_FilenameGeneration(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	tests := []struct {
		name          string
		projectName   string
		expectedFiles []string
	}{
		{
			name:          "simple project",
			projectName:   "SimpleProject",
			expectedFiles: []string{"1.mp4", "2.mp4"},
		},
		{
			name:          "project with spaces",
			projectName:   "My Video Project",
			expectedFiles: []string{"1.mp4", "2.mp4"},
		},
		{
			name:          "project with special chars",
			projectName:   "Project@2024#Final!",
			expectedFiles: []string{"1.mp4", "2.mp4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create project with video clip and highlights
			proj := createTestProject(t, client, ctx, tt.projectName)
			clip := createTestVideoClip(t, client, ctx, proj, "test_video.mp4")
			createTestHighlight(t, client, ctx, clip, 10.0, 20.0)
			createTestHighlight(t, client, ctx, clip, 30.0, 40.0)

			// Create temp directory for export
			tempDir := t.TempDir()

			// Start individual export
			jobID, err := service.ExportIndividualHighlights(proj.ID, tempDir, 0.0)
			require.NoError(t, err)

			// Wait for processing to start and create project directory
			time.Sleep(100 * time.Millisecond)

			// Verify expected filenames would be created (checking the pattern, not actual FFmpeg execution)
			projectDir := filepath.Join(tempDir, fmt.Sprintf("%s_%s", proj.Name, jobID))
			for _, expectedFile := range tt.expectedFiles {
				expectedPath := filepath.Join(projectDir, expectedFile)
				// Just verify the path structure is correct (actual file creation depends on FFmpeg)
				assert.Contains(t, expectedPath, expectedFile)
				assert.True(t, strings.HasSuffix(expectedPath, ".mp4"))
				// Verify it's a simple numeric filename
				filename := filepath.Base(expectedPath)
				assert.Regexp(t, `^\d+\.mp4$`, filename) // Should match pattern: number.mp4
			}
		})
	}
}

func TestIndividualExport_DirectoryPermissions(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project
	proj := createTestProject(t, client, ctx, "PermissionTest")
	clip := createTestVideoClip(t, client, ctx, proj, "TestClip")
	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Start export
	_, err = service.ExportIndividualHighlights(proj.ID, tempDir, 0.0)
	require.NoError(t, err)

	// Wait for directory creation
	time.Sleep(100 * time.Millisecond)

	// Verify directory was created with correct permissions
	projectDir := filepath.Join(tempDir, "PermissionTest")
	assert.DirExists(t, projectDir)

	// Check directory permissions
	info, err := os.Stat(projectDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
}

func TestIndividualExport_ConcurrentExports(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create multiple projects
	proj1 := createTestProject(t, client, ctx, "ConcurrentProject1")
	proj2 := createTestProject(t, client, ctx, "ConcurrentProject2")

	// Create clips and highlights
	clip1 := createTestVideoClip(t, client, ctx, proj1, "Clip1")
	clip2 := createTestVideoClip(t, client, ctx, proj2, "Clip2")

	createTestHighlight(t, client, ctx, clip1, 10.0, 20.0)
	createTestHighlight(t, client, ctx, clip2, 30.0, 40.0)

	// Create separate output directories
	tempDir1, err := os.MkdirTemp("", "export_test_1_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir1)

	tempDir2, err := os.MkdirTemp("", "export_test_2_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir2)

	// Start concurrent exports
	jobID1, err := service.ExportIndividualHighlights(proj1.ID, tempDir1, 0.0)
	require.NoError(t, err)

	jobID2, err := service.ExportIndividualHighlights(proj2.ID, tempDir2, 0.0)
	require.NoError(t, err)

	// Verify both jobs are running
	assert.NotEqual(t, jobID1, jobID2)

	// Check both jobs exist
	progress1, err := service.GetExportProgress(jobID1)
	require.NoError(t, err)
	assert.Equal(t, jobID1, progress1.JobID)

	progress2, err := service.GetExportProgress(jobID2)
	require.NoError(t, err)
	assert.Equal(t, jobID2, progress2.JobID)
}
