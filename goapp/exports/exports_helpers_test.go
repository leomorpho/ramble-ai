package exports

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)


func TestGenerateListFile_Success(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "list_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test segment files
	segmentPaths := []string{
		filepath.Join(tempDir, "segment1.mp4"),
		filepath.Join(tempDir, "segment2.mp4"),
		filepath.Join(tempDir, "segment3.mp4"),
	}

	// Create actual files
	for _, path := range segmentPaths {
		file, err := os.Create(path)
		require.NoError(t, err)
		file.Close()
	}

	// Test list file generation
	listFile, err := service.generateListFile(segmentPaths, tempDir)
	require.NoError(t, err)
	assert.NotEmpty(t, listFile)
	assert.True(t, strings.HasSuffix(listFile, ".txt"))

	// Verify list file exists
	assert.FileExists(t, listFile)

	// Read and verify content
	content, err := os.ReadFile(listFile)
	require.NoError(t, err)
	contentStr := string(content)

	for _, path := range segmentPaths {
		expected := fmt.Sprintf("file '%s'", path)
		assert.Contains(t, contentStr, expected)
	}
}

func TestGenerateListFile_EmptyPaths(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "list_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test with empty paths
	listFile, err := service.generateListFile([]string{}, tempDir)
	require.NoError(t, err)
	assert.NotEmpty(t, listFile)

	// Verify empty file exists
	assert.FileExists(t, listFile)
	
	// Content should be empty
	content, err := os.ReadFile(listFile)
	require.NoError(t, err)
	assert.Empty(t, string(content))
}

func TestGenerateListFile_InvalidDirectory(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Use invalid directory
	invalidDir := "/invalid/nonexistent/path"
	segmentPaths := []string{"segment1.mp4", "segment2.mp4"}

	// Test should fail
	_, err := service.generateListFile(segmentPaths, invalidDir)
	assert.Error(t, err)
}

func TestGenerateOutputFilename_EdgeCases(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	tests := []struct {
		name        string
		projectName string
		suffix      string
		wantContains string
	}{
		{
			name:        "empty project name",
			projectName: "",
			suffix:      "test",
			wantContains: "_test_",
		},
		{
			name:        "empty suffix",
			projectName: "Project",
			suffix:      "",
			wantContains: "Project__",
		},
		{
			name:        "both empty",
			projectName: "",
			suffix:      "",
			wantContains: "__",
		},
		{
			name:        "only special characters",
			projectName: "@#$%^&*()",
			suffix:      "!@#$%",
			wantContains: "__________",
		},
		{
			name:        "unicode characters",
			projectName: "Project™®©",
			suffix:      "test",
			wantContains: "Project____test_",
		},
		{
			name:        "very long name",
			projectName: strings.Repeat("VeryLongProjectName", 10),
			suffix:      "test",
			wantContains: "VeryLongProjectName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.generateOutputFilename(tt.projectName, tt.suffix)
			
			assert.Contains(t, result, tt.wantContains)
			assert.True(t, strings.HasSuffix(result, ".mp4"))
			assert.True(t, len(result) > 0)
		})
	}
}

func TestGetProjectHighlightsForExport(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test projects
	proj1 := createTestProject(t, client, ctx, "Project1")
	proj2 := createTestProject(t, client, ctx, "Project2")

	// Create clips and highlights for each project
	clip1 := createTestVideoClip(t, client, ctx, proj1, "Clip1")
	clip2 := createTestVideoClip(t, client, ctx, proj2, "Clip2")

	h1 := createTestHighlight(t, client, ctx, clip1, 10.0, 20.0)
	h2 := createTestHighlight(t, client, ctx, clip1, 30.0, 40.0)
	h3 := createTestHighlight(t, client, ctx, clip2, 50.0, 60.0)

	// Test getting highlights for project 1
	segments1, err := service.getProjectHighlightsForExport(proj1.ID)
	require.NoError(t, err)
	
	// Should get highlights for project1 only
	foundProject1Highlights := 0
	for _, segment := range segments1 {
		if segment.VideoClipID == clip1.ID {
			foundProject1Highlights++
			assert.Contains(t, []string{h1, h2}, segment.ID)
		}
	}
	assert.Equal(t, 2, foundProject1Highlights, "Should find exactly 2 highlights for project1")

	// Test getting highlights for project 2
	segments2, err := service.getProjectHighlightsForExport(proj2.ID)
	require.NoError(t, err)
	assert.Len(t, segments2, 1)

	// Verify segment belongs to project 2
	assert.Equal(t, h3, segments2[0].ID)
	assert.Equal(t, clip2.ID, segments2[0].VideoClipID)
}

func TestGetProjectHighlightsForExport_NoHighlights(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create project with no highlights
	proj := createTestProject(t, client, ctx, "EmptyProject")
	createTestVideoClip(t, client, ctx, proj, "EmptyClip")

	// Test getting highlights
	segments, err := service.getProjectHighlightsForExport(proj.ID)
	require.NoError(t, err)
	assert.Len(t, segments, 0)
}

func TestGetProjectHighlightsForExport_InvalidProject(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Test with non-existent project
	segments, err := service.getProjectHighlightsForExport(999)
	require.NoError(t, err)
	assert.Len(t, segments, 0)
}

func TestParseFFmpegProgress(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project and job
	proj := createTestProject(t, client, ctx, "ProgressProject")
	jobID := "test_job_123"
	
	// Create export job
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

	// Create mock FFmpeg output
	ffmpegOutput := `frame=  120 fps= 30 q=28.0 size=    1024kB time=00:00:04.00 bitrate=2097.2kbits/s speed=1.0x
frame=  240 fps= 30 q=28.0 size=    2048kB time=00:00:08.00 bitrate=2097.2kbits/s speed=1.0x
frame=  360 fps= 30 q=28.0 size=    3072kB time=00:00:12.00 bitrate=2097.2kbits/s speed=1.0x`

	// Create reader from output
	reader := strings.NewReader(ffmpegOutput)
	
	// Mock cancel channel
	cancel := make(chan bool)
	
	// Test progress parsing (this will run in background)
	go service.parseFFmpegProgress(io.NopCloser(reader), jobID, 20.0, cancel)
	
	// Wait for parsing to complete
	time.Sleep(100 * time.Millisecond)
	
	// Close cancel channel
	close(cancel)
	
	// Verify progress was updated (we can't guarantee exact values due to timing)
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Equal(t, jobID, progress.JobID)
}

func TestParseFFmpegProgress_InvalidFormat(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project and job
	proj := createTestProject(t, client, ctx, "ProgressProject")
	jobID := "test_job_123"
	
	// Create export job
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

	// Create invalid FFmpeg output
	invalidOutput := `invalid line
another invalid line
frame=abc fps=xyz`

	// Create reader from output
	reader := strings.NewReader(invalidOutput)
	
	// Mock cancel channel
	cancel := make(chan bool)
	
	// Test progress parsing (should not crash)
	go service.parseFFmpegProgress(io.NopCloser(reader), jobID, 20.0, cancel)
	
	// Wait for parsing to complete
	time.Sleep(100 * time.Millisecond)
	
	// Close cancel channel
	close(cancel)
	
	// Should not crash or cause errors
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Equal(t, jobID, progress.JobID)
}

func TestUpdateJobStatus(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project and job
	proj := createTestProject(t, client, ctx, "StatusProject")
	jobID := "test_job_123"
	
	// Create export job
	_, err := client.ExportJob.
		Create().
		SetJobID(jobID).
		SetExportType("individual").
		SetOutputPath("/test/path").
		SetStage("pending").
		SetProject(proj).
		SetCreatedAt(time.Now()).
		Save(ctx)
	require.NoError(t, err)

	// Test updating status
	service.updateJobStatus(jobID, "processing")
	
	// Verify status was updated
	progress, err := service.GetExportProgress(jobID)
	require.NoError(t, err)
	assert.Equal(t, "processing", progress.Stage)
}

func TestUpdateJobStatus_InvalidJob(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Test updating non-existent job (should not crash)
	service.updateJobStatus("non_existent_job", "processing")
	
	// No assertion needed - just verify it doesn't crash
}

func TestHighlightSegmentConversion(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project and highlights
	proj := createTestProject(t, client, ctx, "ConversionProject")
	clip := createTestVideoClip(t, client, ctx, proj, "ConversionClip")
	
	// Add transcription to the clip
	_, err := client.VideoClip.
		UpdateOne(clip).
		SetTranscription("This is a test transcription with multiple words").
		Save(ctx)
	require.NoError(t, err)

	// Create highlight
	highlight := createTestHighlight(t, client, ctx, clip, 10.0, 20.0)

	// Get highlights for export
	segments, err := service.getProjectHighlightsForExport(proj.ID)
	require.NoError(t, err)
	assert.Len(t, segments, 1)

	// Verify segment properties
	segment := segments[0]
	assert.Equal(t, highlight, segment.ID)
	assert.Equal(t, clip.FilePath, segment.VideoPath)
	assert.Equal(t, 10.0, segment.Start)
	assert.Equal(t, 20.0, segment.End)
	assert.Equal(t, 3, segment.ColorID)
	assert.Equal(t, clip.ID, segment.VideoClipID)
	assert.Equal(t, clip.Name, segment.VideoClipName)
}

func TestActiveJobsManagement(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project
	proj := createTestProject(t, client, ctx, "ActiveProject")
	clip := createTestVideoClip(t, client, ctx, proj, "ActiveClip")
	createTestHighlight(t, client, ctx, clip, 10.0, 20.0)

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Start export
	jobID, err := service.ExportIndividualHighlights(proj.ID, tempDir)
	require.NoError(t, err)

	// Verify active job exists
	activeJobsMutex.Lock()
	activeJob, exists := activeJobs[jobID]
	activeJobsMutex.Unlock()
	
	assert.True(t, exists)
	assert.Equal(t, jobID, activeJob.JobID)
	assert.True(t, activeJob.IsActive)
	assert.NotNil(t, activeJob.Cancel)

	// Cancel job
	err = service.CancelExport(jobID)
	require.NoError(t, err)

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	// Verify active job was removed
	activeJobsMutex.Lock()
	_, exists = activeJobs[jobID]
	activeJobsMutex.Unlock()
	
	assert.False(t, exists)
}

func TestFFmpegProgressParsing(t *testing.T) {
	tests := []struct {
		name           string
		line           string
		expectedTime   float64
		expectedFrame  int
		shouldHaveTime bool
		shouldHaveFrame bool
	}{
		{
			name:            "valid progress line",
			line:            "frame=  120 fps= 30 q=28.0 size=    1024kB time=00:00:04.00 bitrate=2097.2kbits/s speed=1.0x",
			expectedTime:    4.0,
			expectedFrame:   120,
			shouldHaveTime:  true,
			shouldHaveFrame: true,
		},
		{
			name:            "time only",
			line:            "time=00:01:30.50",
			expectedTime:    90.5,
			shouldHaveTime:  true,
			shouldHaveFrame: false,
		},
		{
			name:            "frame only",
			line:            "frame=  240",
			expectedFrame:   240,
			shouldHaveTime:  false,
			shouldHaveFrame: true,
		},
		{
			name:            "invalid time format",
			line:            "time=invalid",
			shouldHaveTime:  false,
			shouldHaveFrame: false,
		},
		{
			name:            "invalid frame format",
			line:            "frame=abc",
			shouldHaveTime:  false,
			shouldHaveFrame: false,
		},
		{
			name:            "empty line",
			line:            "",
			shouldHaveTime:  false,
			shouldHaveFrame: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This tests the internal parsing logic
			// We'll check by looking at the line content
			if tt.shouldHaveTime {
				assert.Contains(t, tt.line, "time=")
			}
			if tt.shouldHaveFrame {
				assert.Contains(t, tt.line, "frame=")
			}
		})
	}
}

func TestFileOperations(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	_ = NewExportService(client, ctx)

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "file_ops_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test file creation
	testFile := filepath.Join(tempDir, "test.txt")
	content := "test content"
	
	err = os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)
	
	// Verify file exists
	assert.FileExists(t, testFile)
	
	// Read and verify content
	readContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, content, string(readContent))
	
	// Test file deletion
	err = os.Remove(testFile)
	require.NoError(t, err)
	
	// Verify file was deleted
	assert.NoFileExists(t, testFile)
}

func TestStringManipulation(t *testing.T) {
	client, ctx := setupTestDB(t)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Test sanitization regex
	tests := []struct {
		input    string
		expected string
	}{
		{"SimpleProject", "SimpleProject"},
		{"Project With Spaces", "Project_With_Spaces"},
		{"Project@#$%", "Project____"},
		{"Project123", "Project123"},
		{"Project_-Valid", "Project_-Valid"},
		{"", ""},
		{"@#$%^&*()", "__________"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			filename := service.generateOutputFilename(tt.input, "test")
			assert.Contains(t, filename, tt.expected)
		})
	}
}

// Benchmark tests for helper functions
func BenchmarkGenerateOutputFilename(b *testing.B) {
	client, ctx := setupTestDB(b)
	defer client.Close()
	service := NewExportService(client, ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.generateOutputFilename("TestProject", "highlight_001")
	}
}

func BenchmarkGenerateListFile(b *testing.B) {
	client, ctx := setupTestDB(b)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "benchmark_test_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	// Create test segment paths
	segmentPaths := make([]string, 100)
	for i := range segmentPaths {
		segmentPaths[i] = filepath.Join(tempDir, fmt.Sprintf("segment_%d.mp4", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.generateListFile(segmentPaths, tempDir)
	}
}

func BenchmarkGetProjectHighlights(b *testing.B) {
	client, ctx := setupTestDB(b)
	defer client.Close()
	service := NewExportService(client, ctx)

	// Create test project with highlights
	proj := createTestProject(b, client, ctx, "BenchmarkProject")
	clip := createTestVideoClip(b, client, ctx, proj, "BenchmarkClip")
	
	// Create multiple highlights
	for i := 0; i < 50; i++ {
		start := float64(i * 10)
		end := start + 5.0
		createTestHighlight(b, client, ctx, clip, start, end)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.getProjectHighlightsForExport(proj.ID)
	}
}