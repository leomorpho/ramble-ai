package exports

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"ramble-ai/ent"
	"ramble-ai/ent/exportjob"
	"ramble-ai/ent/project"
	"ramble-ai/goapp"
	"ramble-ai/goapp/highlights"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ExportProgress represents the current state of an export job
type ExportProgress struct {
	JobID          string     `json:"jobId"`
	Stage          string     `json:"stage"`
	Progress       float64    `json:"progress"`
	CurrentFile    string     `json:"currentFile"`
	TotalFiles     int        `json:"totalFiles"`
	ProcessedFiles int        `json:"processedFiles"`
	IsComplete     bool       `json:"isComplete"`
	HasError       bool       `json:"hasError"`
	ErrorMessage   string     `json:"errorMessage"`
	IsCancelled    bool       `json:"isCancelled"`
	ExportType     string     `json:"exportType"`
	OutputPath     string     `json:"outputPath"`
	CompletedAt    *time.Time `json:"completedAt"`
}

// ActiveExportJob represents an active export job
type ActiveExportJob struct {
	JobID    string
	Cancel   chan bool
	IsActive bool
}

// FFmpegProgress represents FFmpeg progress information
type FFmpegProgress struct {
	Frame    int64
	FPS      float64
	Bitrate  string
	Time     float64
	Duration float64
	Progress float64
}

// HighlightSegment represents a single highlight segment for export
type HighlightSegment = highlights.HighlightSegment

// Global active job manager (for cancellation and in-memory tracking)
var (
	activeJobs      = make(map[string]*ActiveExportJob)
	activeJobsMutex = sync.RWMutex{}
)

// ExportService provides export functionality
type ExportService struct {
	client *ent.Client
	ctx    context.Context
}

// NewExportService creates a new export service
func NewExportService(client *ent.Client, ctx context.Context) *ExportService {
	return &ExportService{
		client: client,
		ctx:    ctx,
	}
}

// ExportStitchedHighlights exports all highlights from a project as a single stitched video
func (s *ExportService) ExportStitchedHighlights(projectID int, outputFolder string, paddingSeconds float64) (string, error) {
	// Generate unique job ID
	jobID := fmt.Sprintf("export_%d_%d", projectID, time.Now().UnixNano())

	// Create database record
	// Get the project to create relation
	proj, err := s.client.Project.Get(s.ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("failed to get project: %w", err)
	}

	// Retry job creation if database is locked
	var dbJob *ent.ExportJob
	for i := 0; i < 5; i++ {
		var createErr error
		dbJob, createErr = s.client.ExportJob.
			Create().
			SetJobID(jobID).
			SetExportType("stitched").
			SetOutputPath(outputFolder).
			SetStage("pending").
			SetCreatedAt(time.Now()).
			SetProject(proj).
			Save(s.ctx)

		if createErr == nil {
			break
		}
		
		// Check if it's a database lock error
		if strings.Contains(createErr.Error(), "database table is locked") {
			time.Sleep(time.Duration(i*10) * time.Millisecond)
			continue
		}
		
		// Other errors, don't retry
		return "", fmt.Errorf("failed to create export job record: %w", createErr)
	}

	if dbJob == nil {
		return "", fmt.Errorf("failed to create export job record after retries")
	}

	// Create active job
	activeJob := &ActiveExportJob{
		JobID:    jobID,
		Cancel:   make(chan bool),
		IsActive: true,
	}

	// Register active job
	activeJobsMutex.Lock()
	activeJobs[jobID] = activeJob
	activeJobsMutex.Unlock()

	// Run export in background
	go s.performStitchedExport(dbJob, activeJob, paddingSeconds)

	return jobID, nil
}

// ExportIndividualHighlights exports each highlight as a separate video file
func (s *ExportService) ExportIndividualHighlights(projectID int, outputFolder string, paddingSeconds float64) (string, error) {
	// Generate unique job ID
	jobID := fmt.Sprintf("export_%d_%d", projectID, time.Now().UnixNano())

	// Create database record
	// Get the project to create relation
	proj, err := s.client.Project.Get(s.ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("failed to get project: %w", err)
	}

	// Retry job creation if database is locked
	var dbJob *ent.ExportJob
	for i := 0; i < 5; i++ {
		var createErr error
		dbJob, createErr = s.client.ExportJob.
			Create().
			SetJobID(jobID).
			SetExportType("individual").
			SetOutputPath(outputFolder).
			SetStage("pending").
			SetCreatedAt(time.Now()).
			SetProject(proj).
			Save(s.ctx)

		if createErr == nil {
			break
		}
		
		// Check if it's a database lock error
		if strings.Contains(createErr.Error(), "database table is locked") {
			time.Sleep(time.Duration(i*10) * time.Millisecond)
			continue
		}
		
		// Other errors, don't retry
		return "", fmt.Errorf("failed to create export job record: %w", createErr)
	}

	if dbJob == nil {
		return "", fmt.Errorf("failed to create export job record after retries")
	}

	// Create active job
	activeJob := &ActiveExportJob{
		JobID:    jobID,
		Cancel:   make(chan bool),
		IsActive: true,
	}

	// Register active job
	activeJobsMutex.Lock()
	activeJobs[jobID] = activeJob
	activeJobsMutex.Unlock()

	// Run export in background
	go s.performIndividualExport(dbJob, activeJob, paddingSeconds)

	return jobID, nil
}

// GetExportProgress returns the current progress of an export job
func (s *ExportService) GetExportProgress(jobID string) (*ExportProgress, error) {
	// Get job from database
	job, err := s.client.ExportJob.
		Query().
		Where(exportjob.JobID(jobID)).
		Only(s.ctx)

	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}

	// Check if job is cancelled
	isCancelled := job.IsCancelled

	return &ExportProgress{
		JobID:          job.JobID,
		Stage:          job.Stage,
		Progress:       job.Progress,
		CurrentFile:    job.CurrentFile,
		TotalFiles:     job.TotalFiles,
		ProcessedFiles: job.ProcessedFiles,
		IsComplete:     job.IsComplete,
		HasError:       job.HasError,
		ErrorMessage:   job.ErrorMessage,
		IsCancelled:    isCancelled,
		ExportType:     job.ExportType,
		OutputPath:     job.OutputPath,
		CompletedAt:    &job.CompletedAt,
	}, nil
}

// CancelExport cancels an active export job
func (s *ExportService) CancelExport(jobID string) error {
	// Send cancellation signal to active job
	activeJobsMutex.RLock()
	if activeJob, exists := activeJobs[jobID]; exists && activeJob.IsActive {
		activeJobsMutex.RUnlock()
		select {
		case activeJob.Cancel <- true:
			log.Printf("Sent cancellation signal to job %s", jobID)
		default:
			log.Printf("Cancel channel full for job %s", jobID)
		}
	} else {
		activeJobsMutex.RUnlock()
	}

	// Update database status - first check if job exists
	_, err := s.client.ExportJob.
		Query().
		Where(exportjob.JobID(jobID)).
		Only(s.ctx)

	if err != nil {
		return fmt.Errorf("job not found: %w", err)
	}

	// Retry job update if database is locked
	var updateErr error
	for i := 0; i < 5; i++ {
		_, updateErr = s.client.ExportJob.
			Update().
			Where(exportjob.JobID(jobID)).
			SetIsCancelled(true).
			SetStage("cancelled").
			SetIsComplete(true).
			Save(s.ctx)

		if updateErr == nil {
			break
		}
		
		// Check if it's a database lock error
		if strings.Contains(updateErr.Error(), "database table is locked") {
			time.Sleep(time.Duration(i*10) * time.Millisecond)
			continue
		}
		
		// Other errors, don't retry
		return fmt.Errorf("failed to update job status: %w", updateErr)
	}

	if updateErr != nil {
		return fmt.Errorf("failed to update job status after retries: %w", updateErr)
	}

	return nil
}

// GetProjectExportJobs retrieves all export jobs for a project
func (s *ExportService) GetProjectExportJobs(projectID int) ([]*ExportProgress, error) {
	// Get all export jobs for the project
	jobs, err := s.client.ExportJob.
		Query().
		Where(exportjob.HasProjectWith(project.ID(projectID))).
		Order(ent.Desc(exportjob.FieldCreatedAt)).
		All(s.ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get export jobs: %w", err)
	}

	var progress []*ExportProgress
	for _, job := range jobs {
		progress = append(progress, &ExportProgress{
			JobID:          job.JobID,
			Stage:          job.Stage,
			Progress:       job.Progress,
			CurrentFile:    job.CurrentFile,
			TotalFiles:     job.TotalFiles,
			ProcessedFiles: job.ProcessedFiles,
			IsComplete:     job.IsComplete,
			HasError:       job.HasError,
			ErrorMessage:   job.ErrorMessage,
			IsCancelled:    job.IsCancelled,
		})
	}

	return progress, nil
}

// RecoverActiveExportJobs restores export jobs that were running when the app was closed
func (s *ExportService) RecoverActiveExportJobs() error {
	// Find all export jobs that were processing but not completed
	jobs, err := s.client.ExportJob.
		Query().
		Where(
			exportjob.IsComplete(false),
			exportjob.IsCancelled(false),
		).
		All(s.ctx)

	if err != nil {
		return fmt.Errorf("failed to query active export jobs: %w", err)
	}

	// Mark all previously active jobs as failed with a recovery message
	for _, job := range jobs {
		_, err := s.client.ExportJob.
			UpdateOne(job).
			SetHasError(true).
			SetStage("recovery").
			SetErrorMessage("Export interrupted - application was closed during processing").
			SetIsComplete(true).
			Save(s.ctx)

		if err != nil {
			log.Printf("Failed to update job %s status during recovery: %v", job.JobID, err)
		} else {
			log.Printf("Marked interrupted job %s as failed", job.JobID)
		}
	}

	if len(jobs) > 0 {
		log.Printf("Recovered %d interrupted export jobs", len(jobs))
	}

	return nil
}

// performStitchedExport performs the actual stitched export in the background
func (s *ExportService) performStitchedExport(dbJob *ent.ExportJob, activeJob *ActiveExportJob, paddingSeconds float64) {
	defer func() {
		// Cleanup active job
		activeJobsMutex.Lock()
		delete(activeJobs, dbJob.JobID)
		activeJobsMutex.Unlock()
	}()

	// Update status to processing
	s.updateJobStatus(dbJob.JobID, "processing")

	// Get all highlights for the project
	// Get project from job relation
	proj, err := dbJob.QueryProject().Only(s.ctx)
	if err != nil {
		s.updateJobFailed(dbJob.JobID, fmt.Sprintf("Failed to get project: %v", err))
		return
	}

	segments, err := s.getProjectHighlightsForExport(proj.ID)
	if err != nil {
		s.updateJobFailed(dbJob.JobID, fmt.Sprintf("Failed to get highlights: %v", err))
		return
	}

	s.updateJobProgress(dbJob.JobID, "preparing", 0.0, "", 0, 0)

	if len(segments) == 0 {
		s.updateJobFailed(dbJob.JobID, "No highlights found to export")
		return
	}

	// Project already retrieved above, no need to get it again

	// Create temp directory for segments
	tempDir, err := os.MkdirTemp("", "export_*")
	if err != nil {
		s.updateJobFailed(dbJob.JobID, fmt.Sprintf("Failed to create temp directory: %v", err))
		return
	}
	defer os.RemoveAll(tempDir)

	s.updateJobProgress(dbJob.JobID, "preparing", 0.0, "", len(segments), 0)

	// Extract all segments
	var segmentPaths []string
	for i, segment := range segments {
		// Check for cancellation
		select {
		case <-activeJob.Cancel:
			s.updateJobCancelled(dbJob.JobID)
			return
		default:
		}

		// Calculate progress
		progress := float64(i) / float64(len(segments)) * 0.7 // 70% for extraction

		// Get file name without extension
		fileName := filepath.Base(segment.VideoPath)
		if lastDot := strings.LastIndex(fileName, "."); lastDot != -1 {
			fileName = fileName[:lastDot]
		}

		s.updateJobProgress(dbJob.JobID, "extracting", progress, fileName, len(segments), i)

		segmentPath, err := s.extractHighlightSegmentWithProgress(segment, tempDir, i+1, dbJob.JobID, activeJob.Cancel, paddingSeconds)
		if err != nil {
			s.updateJobFailed(dbJob.JobID, fmt.Sprintf("Failed to extract segment %d: %v", i+1, err))
			return
		}
		segmentPaths = append(segmentPaths, segmentPath)
	}

	// Create list file for concatenation
	s.updateJobProgress(dbJob.JobID, "stitching", 0.8, "Combining highlight segments", len(segments), len(segments))

	outputFile := filepath.Join(dbJob.OutputPath, s.generateOutputFilename(proj.Name, "stitched"))

	if err := s.stitchSegments(segmentPaths, outputFile, dbJob.JobID, activeJob.Cancel); err != nil {
		s.updateJobFailed(dbJob.JobID, fmt.Sprintf("Failed to stitch segments: %v", err))
		return
	}

	// Update job as completed
	s.updateJobCompleted(dbJob.JobID, outputFile)
	completionMessage := fmt.Sprintf("Successfully exported %d highlights to %s", len(segments), filepath.Base(outputFile))
	s.updateJobProgress(dbJob.JobID, "complete", 1.0, completionMessage, len(segments), len(segments))
}

// performIndividualExport performs the actual individual export in the background
func (s *ExportService) performIndividualExport(dbJob *ent.ExportJob, activeJob *ActiveExportJob, paddingSeconds float64) {
	defer func() {
		// Cleanup active job
		activeJobsMutex.Lock()
		delete(activeJobs, dbJob.JobID)
		activeJobsMutex.Unlock()
	}()

	// Update status to processing
	s.updateJobStatus(dbJob.JobID, "processing")

	// Get all highlights for the project
	// Get project from job relation
	proj, err := dbJob.QueryProject().Only(s.ctx)
	if err != nil {
		s.updateJobFailed(dbJob.JobID, fmt.Sprintf("Failed to get project: %v", err))
		return
	}

	segments, err := s.getProjectHighlightsForExport(proj.ID)
	if err != nil {
		s.updateJobFailed(dbJob.JobID, fmt.Sprintf("Failed to get highlights: %v", err))
		return
	}

	if len(segments) == 0 {
		s.updateJobFailed(dbJob.JobID, "No highlights found to export")
		return
	}

	// Create subdirectory with project name
	sanitizedProjectName := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(proj.Name, "_")
	projectDir := filepath.Join(dbJob.OutputPath, sanitizedProjectName)

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		s.updateJobFailed(dbJob.JobID, fmt.Sprintf("Failed to create project directory: %v", err))
		return
	}

	s.updateJobProgress(dbJob.JobID, "extracting", 0.0, "", len(segments), 0)

	// Export each segment
	for i, segment := range segments {
		// Check for cancellation
		select {
		case <-activeJob.Cancel:
			s.updateJobCancelled(dbJob.JobID)
			return
		default:
		}

		// Calculate progress
		progress := float64(i) / float64(len(segments))

		// Get file name without extension
		fileName := filepath.Base(segment.VideoPath)
		if lastDot := strings.LastIndex(fileName, "."); lastDot != -1 {
			fileName = fileName[:lastDot]
		}

		s.updateJobProgress(dbJob.JobID, "extracting", progress, fileName, len(segments), i)

		// Generate simple numbered filename for this highlight in the project directory
		outputFile := filepath.Join(projectDir, fmt.Sprintf("%d.mp4", i+1))

		err := s.extractHighlightSegmentDirectWithProgress(segment, outputFile, dbJob.JobID, activeJob.Cancel, paddingSeconds)
		if err != nil {
			s.updateJobFailed(dbJob.JobID, fmt.Sprintf("Failed to extract segment %d: %v", i+1, err))
			return
		}
	}

	// Update job as completed
	s.updateJobCompleted(dbJob.JobID, projectDir)
	completionMessage := fmt.Sprintf("Successfully exported %d individual highlights", len(segments))
	s.updateJobProgress(dbJob.JobID, "complete", 1.0, completionMessage, len(segments), len(segments))
}

// getProjectHighlightsForExport retrieves all highlights for a project in the correct order
func (s *ExportService) getProjectHighlightsForExport(projectID int) ([]HighlightSegment, error) {
	service := highlights.NewHighlightService(s.client, s.ctx)
	return service.GetProjectHighlightsForExport(projectID)
}

// calculatePaddedTimes calculates start and end times with padding, respecting video boundaries
func (s *ExportService) calculatePaddedTimes(segment HighlightSegment, paddingSeconds float64) (float64, float64, error) {
	// Calculate padded start time (never go below 0)
	paddedStart := math.Max(0, segment.Start-paddingSeconds)

	// For end time, we'll use the original end + padding
	// Note: We don't validate against video duration here since FFmpeg will handle it gracefully
	// by using the maximum available duration if we exceed the video length
	paddedEnd := segment.End + paddingSeconds

	return paddedStart, paddedEnd, nil
}

// extractHighlightSegment extracts a single highlight segment to a temp file
func (s *ExportService) extractHighlightSegment(segment HighlightSegment, tempDir string, index int) (string, error) {
	// Generate output filename
	outputPath := filepath.Join(tempDir, fmt.Sprintf("segment_%03d.mp4", index))

	// Build FFmpeg command - using input seeking + minimal re-encoding for precision
	duration := segment.End - segment.Start
	cmd := goapp.GetFFmpegCommand(
		"-ss", fmt.Sprintf("%.3f", segment.Start),
		"-i", segment.VideoPath,
		"-t", fmt.Sprintf("%.3f", duration),
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-crf", "18",
		"-c:a", "copy",
		"-movflags", "+faststart",
		"-y",
		outputPath,
	)

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg error: %v\nOutput: %s", err, string(output))
	}

	return outputPath, nil
}

// extractHighlightSegmentDirect extracts a highlight segment directly to the output file
func (s *ExportService) extractHighlightSegmentDirect(segment HighlightSegment, outputPath string) error {
	// Build FFmpeg command - using input seeking + minimal re-encoding for precision
	duration := segment.End - segment.Start
	cmd := goapp.GetFFmpegCommand(
		"-ss", fmt.Sprintf("%.3f", segment.Start),
		"-i", segment.VideoPath,
		"-t", fmt.Sprintf("%.3f", duration),
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-crf", "18",
		"-c:a", "copy",
		"-movflags", "+faststart",
		"-y",
		outputPath,
	)

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %v\nOutput: %s", err, string(output))
	}

	return nil
}

// generateListFile creates a concat list file for FFmpeg
func (s *ExportService) generateListFile(segmentPaths []string, tempDir string) (string, error) {
	listPath := filepath.Join(tempDir, "concat_list.txt")
	file, err := os.Create(listPath)
	if err != nil {
		return "", fmt.Errorf("failed to create list file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, path := range segmentPaths {
		// FFmpeg concat requires specific format
		fmt.Fprintf(writer, "file '%s'\n", path)
	}
	writer.Flush()

	return listPath, nil
}

// generateOutputFilename creates a unique filename for the export
func (s *ExportService) generateOutputFilename(projectName, suffix string) string {
	// Sanitize project name
	sanitized := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(projectName, "_")

	// Create timestamp
	timestamp := time.Now().Format("20060102_150405")

	// Generate filename
	return fmt.Sprintf("%s_%s_%s.mp4", sanitized, suffix, timestamp)
}

// Database update helper functions
func (s *ExportService) updateJobStatus(jobID, stage string) {
	// Retry job update if database is locked
	for i := 0; i < 5; i++ {
		_, err := s.client.ExportJob.
			Update().
			Where(exportjob.JobID(jobID)).
			SetStage(stage).
			Save(s.ctx)

		if err == nil {
			return
		}
		
		// Check if it's a database lock error
		if strings.Contains(err.Error(), "database table is locked") {
			time.Sleep(time.Duration(i*10) * time.Millisecond)
			continue
		}
		
		// Other errors, log and return
		log.Printf("Failed to update job status: %v", err)
		return
	}
	
	log.Printf("Failed to update job status after retries")
}

func (s *ExportService) updateJobProgress(jobID, stage string, progress float64, currentFile string, totalFiles, processedFiles int) {
	// Retry job update if database is locked
	for i := 0; i < 5; i++ {
		_, err := s.client.ExportJob.
			Update().
			Where(exportjob.JobID(jobID)).
			SetStage(stage).
			SetProgress(progress).
			SetCurrentFile(currentFile).
			SetTotalFiles(totalFiles).
			SetProcessedFiles(processedFiles).
			Save(s.ctx)

		if err == nil {
			return
		}
		
		// Check if it's a database lock error
		if strings.Contains(err.Error(), "database table is locked") {
			time.Sleep(time.Duration(i*10) * time.Millisecond)
			continue
		}
		
		// Other errors, log and return
		log.Printf("Failed to update job progress: %v", err)
		return
	}
	
	log.Printf("Failed to update job progress after retries")
}

func (s *ExportService) updateJobCompleted(jobID, outputPath string) {
	// Retry job update if database is locked
	for i := 0; i < 5; i++ {
		_, err := s.client.ExportJob.
			Update().
			Where(exportjob.JobID(jobID)).
			SetIsComplete(true).
			SetProgress(1.0).
			SetCompletedAt(time.Now()).
			SetOutputPath(outputPath).
			Save(s.ctx)

		if err == nil {
			return
		}
		
		// Check if it's a database lock error
		if strings.Contains(err.Error(), "database table is locked") {
			time.Sleep(time.Duration(i*10) * time.Millisecond)
			continue
		}
		
		// Other errors, log and return
		log.Printf("Failed to update job as completed: %v", err)
		return
	}
	
	log.Printf("Failed to update job as completed after retries")
}

func (s *ExportService) updateJobFailed(jobID, errorMessage string) {
	// Retry job update if database is locked
	for i := 0; i < 5; i++ {
		_, err := s.client.ExportJob.
			Update().
			Where(exportjob.JobID(jobID)).
			SetHasError(true).
			SetStage("failed").
			SetErrorMessage(errorMessage).
			SetIsComplete(true).
			Save(s.ctx)

		if err == nil {
			return
		}
		
		// Check if it's a database lock error
		if strings.Contains(err.Error(), "database table is locked") {
			time.Sleep(time.Duration(i*10) * time.Millisecond)
			continue
		}
		
		// Other errors, log and return
		log.Printf("Failed to update job as failed: %v", err)
		return
	}
	
	log.Printf("Failed to update job as failed after retries")
}

func (s *ExportService) updateJobCancelled(jobID string) {
	// Retry job update if database is locked
	for i := 0; i < 5; i++ {
		_, err := s.client.ExportJob.
			Update().
			Where(exportjob.JobID(jobID)).
			SetIsCancelled(true).
			SetStage("cancelled").
			SetIsComplete(true).
			Save(s.ctx)

		if err == nil {
			return
		}
		
		// Check if it's a database lock error
		if strings.Contains(err.Error(), "database table is locked") {
			time.Sleep(time.Duration(i*10) * time.Millisecond)
			continue
		}
		
		// Other errors, log and return
		log.Printf("Failed to update job as cancelled: %v", err)
		return
	}
	
	log.Printf("Failed to update job as cancelled after retries")
}

// extractHighlightSegmentWithProgress extracts a highlight with progress tracking
func (s *ExportService) extractHighlightSegmentWithProgress(segment HighlightSegment, tempDir string, index int, jobID string, cancel chan bool, paddingSeconds float64) (string, error) {
	// Generate output filename
	outputPath := filepath.Join(tempDir, fmt.Sprintf("segment_%03d.mp4", index))

	// Calculate padded times
	paddedStart, paddedEnd, err := s.calculatePaddedTimes(segment, paddingSeconds)
	if err != nil {
		return "", fmt.Errorf("failed to calculate padded times: %w", err)
	}

	// Build FFmpeg command with progress tracking - using input seeking + minimal re-encoding for precision
	duration := paddedEnd - paddedStart
	cmd := goapp.GetFFmpegCommand(
		"-progress", "pipe:1",
		"-ss", fmt.Sprintf("%.3f", paddedStart),
		"-i", segment.VideoPath,
		"-t", fmt.Sprintf("%.3f", duration),
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-crf", "18",
		"-c:a", "copy",
		"-movflags", "+faststart",
		"-y",
		outputPath,
	)

	// Create pipes for stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Monitor for cancellation
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	// Process output
	go s.parseFFmpegProgress(stdout, jobID, segment.End-segment.Start, cancel)

	// Wait for completion or cancellation
	select {
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("ffmpeg error: %v", err)
		}
		return outputPath, nil
	case <-cancel:
		cmd.Process.Kill()
		return "", fmt.Errorf("export cancelled")
	}
}

// extractHighlightSegmentDirectWithProgress extracts directly with progress tracking
func (s *ExportService) extractHighlightSegmentDirectWithProgress(segment HighlightSegment, outputPath, jobID string, cancel chan bool, paddingSeconds float64) error {
	// Calculate padded times
	paddedStart, paddedEnd, err := s.calculatePaddedTimes(segment, paddingSeconds)
	if err != nil {
		return fmt.Errorf("failed to calculate padded times: %w", err)
	}

	// Build FFmpeg command with progress tracking - using input seeking + minimal re-encoding for precision
	duration := paddedEnd - paddedStart
	cmd := goapp.GetFFmpegCommand(
		"-progress", "pipe:1",
		"-ss", fmt.Sprintf("%.3f", paddedStart),
		"-i", segment.VideoPath,
		"-t", fmt.Sprintf("%.3f", duration),
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-crf", "18",
		"-c:a", "copy",
		"-movflags", "+faststart",
		"-y",
		outputPath,
	)

	// Create pipes for stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Monitor for cancellation
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	// Process output
	go s.parseFFmpegProgress(stdout, jobID, segment.End-segment.Start, cancel)

	// Wait for completion or cancellation
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("ffmpeg error: %v", err)
		}
		return nil
	case <-cancel:
		cmd.Process.Kill()
		return fmt.Errorf("export cancelled")
	}
}

// parseFFmpegProgress parses FFmpeg progress output
func (s *ExportService) parseFFmpegProgress(stdout io.ReadCloser, jobID string, duration float64, cancel chan bool) {
	scanner := bufio.NewScanner(stdout)
	progress := &FFmpegProgress{Duration: duration}

	for scanner.Scan() {
		line := scanner.Text()

		// Parse progress lines
		if strings.HasPrefix(line, "out_time_ms=") {
			if timeMsStr := strings.TrimPrefix(line, "out_time_ms="); timeMsStr != "" {
				if timeMs, err := strconv.ParseInt(timeMsStr, 10, 64); err == nil {
					progress.Time = float64(timeMs) / 1000000.0 // Convert microseconds to seconds
					if progress.Duration > 0 {
						progress.Progress = progress.Time / progress.Duration
						if progress.Progress > 1.0 {
							progress.Progress = 1.0
						}
					}
				}
			}
		}

		// Check for cancellation
		select {
		case <-cancel:
			return
		default:
		}
	}
}

// stitchSegments combines multiple video segments into one
func (s *ExportService) stitchSegments(segmentPaths []string, outputPath string, jobID string, cancel chan bool) error {
	// Create concat list file
	tempDir := filepath.Dir(segmentPaths[0])
	listFile, err := s.generateListFile(segmentPaths, tempDir)
	if err != nil {
		return err
	}

	// Build FFmpeg concat command with progress
	cmd := goapp.GetFFmpegCommand(
		"-progress", "pipe:1",
		"-f", "concat",
		"-safe", "0",
		"-i", listFile,
		"-c", "copy",
		"-movflags", "+faststart",
		"-y",
		outputPath,
	)

	// Create pipes for stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg concat: %w", err)
	}

	// Monitor for cancellation
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	// Process output for progress
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			// Parse progress and update accordingly
			if strings.HasPrefix(line, "out_time_ms=") {
				// Calculate overall progress for stitching
				// This is simplified - in practice you'd calculate based on total duration
				overallProgress := 0.8 + (0.2 * 0.5) // Base 80% + some progress
				s.updateJobProgress(jobID, "stitching", overallProgress, "Combining clips", 0, 0)
			}
		}
	}()

	// Wait for completion or cancellation
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("ffmpeg concat error: %v", err)
		}
		return nil
	case <-cancel:
		cmd.Process.Kill()
		return fmt.Errorf("export cancelled")
	}
}

// SelectExportFolder opens a dialog for the user to select an export folder
func (s *ExportService) SelectExportFolder(ctx context.Context) (string, error) {
	options := runtime.OpenDialogOptions{
		Title:   "Select Export Folder",
		Filters: []runtime.FileFilter{},
	}

	folder, err := runtime.OpenDirectoryDialog(ctx, options)
	if err != nil {
		return "", fmt.Errorf("failed to open directory dialog: %w", err)
	}

	return folder, nil
}
