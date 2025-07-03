package main

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"MYAPP/ent"
	"MYAPP/goapp/ai"
	"MYAPP/goapp/exports"
	"MYAPP/goapp/highlights"
	"MYAPP/goapp/projects"
	"MYAPP/goapp/settings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// App struct
type App struct {
	ctx    context.Context
	client *ent.Client
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Initialize database
	db, err := sql.Open("sqlite3", "database.db?_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}

	// Create Ent client with proper dialect
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))

	app := &App{
		client: client,
	}

	return app
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Run database migrations
	if err := a.client.Schema.Create(ctx); err != nil {
		log.Printf("failed creating schema resources: %v", err)
	}

	log.Println("Database initialized and migrations applied")

	// Recover any incomplete export jobs
	if err := a.RecoverActiveExportJobs(); err != nil {
		log.Printf("Failed to recover active export jobs: %v", err)
	}
}

// shutdown is called when the app shuts down
func (a *App) shutdown(ctx context.Context) {
	// Close the database connection
	if err := a.client.Close(); err != nil {
		log.Printf("failed to close database connection: %v", err)
	}
}

// createAssetMiddleware creates middleware for serving video files via AssetServer
func (a *App) createAssetMiddleware() assetserver.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if this is a video request
			if strings.HasPrefix(r.URL.Path, "/api/video/") {
				a.handleVideoRequest(w, r)
				return
			}
			// Check if this is a thumbnail request
			if strings.HasPrefix(r.URL.Path, "/api/thumbnail/") {
				a.handleThumbnailRequest(w, r)
				return
			}
			// Pass to next handler for non-video requests
			next.ServeHTTP(w, r)
		})
	}
}

// handleVideoRequest handles video file requests with HTTP range support
func (a *App) handleVideoRequest(w http.ResponseWriter, r *http.Request) {
	// Extract file path from URL
	filePath := r.URL.Path[11:] // Remove "/api/video/"
	log.Printf("[VIDEO] Raw path: %s", r.URL.Path)
	log.Printf("[VIDEO] Extracted path: %s", filePath)

	decodedPath, err := url.QueryUnescape(filePath)
	if err != nil {
		log.Printf("[VIDEO] URL decode error: %v", err)
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	log.Printf("[VIDEO] Decoded path: %s", decodedPath)

	// Security check - ensure file exists and is a video
	if !a.isVideoFile(decodedPath) {
		http.Error(w, "Not a video file", http.StatusBadRequest)
		return
	}

	file, err := os.Open(decodedPath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "File info error", http.StatusInternalServerError)
		return
	}

	// Set content type based on file extension
	contentType := a.getContentType(decodedPath)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")

	// Handle range requests for video seeking
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		a.handleRangeRequest(w, r, file, fileInfo.Size(), rangeHeader)
		return
	}

	// Serve the entire file
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	io.Copy(w, file)
}

// handleRangeRequest handles HTTP range requests for efficient video seeking
func (a *App) handleRangeRequest(w http.ResponseWriter, r *http.Request, file *os.File, fileSize int64, rangeHeader string) {
	// Parse range header (e.g., "bytes=0-1023")
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	rangeSpec := rangeHeader[6:] // Remove "bytes="
	parts := strings.Split(rangeSpec, "-")
	if len(parts) != 2 {
		http.Error(w, "Invalid range format", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	var start, end int64
	var err error

	// Parse start
	if parts[0] != "" {
		start, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil || start < 0 {
			http.Error(w, "Invalid start range", http.StatusRequestedRangeNotSatisfiable)
			return
		}
	}

	// Parse end
	if parts[1] != "" {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil || end >= fileSize {
			end = fileSize - 1
		}
	} else {
		end = fileSize - 1
	}

	// Validate range
	if start > end || start >= fileSize {
		http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	contentLength := end - start + 1

	// Set response headers for partial content
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	w.WriteHeader(http.StatusPartialContent)

	// Seek to start position and copy the requested range
	file.Seek(start, 0)
	io.CopyN(w, file, contentLength)
}

// getContentType returns the appropriate MIME type for video files
func (a *App) getContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	case ".mkv":
		return "video/x-matroska"
	case ".webm":
		return "video/webm"
	case ".flv":
		return "video/x-flv"
	case ".wmv":
		return "video/x-ms-wmv"
	case ".m4v":
		return "video/x-m4v"
	case ".mpg", ".mpeg":
		return "video/mpeg"
	default:
		return "application/octet-stream"
	}
}

// handleThumbnailRequest handles video thumbnail requests
func (a *App) handleThumbnailRequest(w http.ResponseWriter, r *http.Request) {
	// Extract file path from URL
	filePath := r.URL.Path[15:] // Remove "/api/thumbnail/"
	log.Printf("[THUMBNAIL] Raw path: %s", r.URL.Path)
	log.Printf("[THUMBNAIL] Extracted path: %s", filePath)

	decodedPath, err := url.QueryUnescape(filePath)
	if err != nil {
		log.Printf("[THUMBNAIL] URL decode error: %v", err)
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	log.Printf("[THUMBNAIL] Decoded path: %s", decodedPath)

	// Security check - ensure file exists and is a video
	if !a.isVideoFile(decodedPath) {
		http.Error(w, "Not a video file", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(decodedPath); os.IsNotExist(err) {
		http.Error(w, "Video file not found", http.StatusNotFound)
		return
	}

	// Generate or get existing thumbnail
	thumbnailPath, err := a.generateThumbnail(decodedPath)
	if err != nil {
		log.Printf("[THUMBNAIL] Generation error: %v", err)
		http.Error(w, "Failed to generate thumbnail", http.StatusInternalServerError)
		return
	}

	// Serve the thumbnail file
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours
	http.ServeFile(w, r, thumbnailPath)
}

// generateThumbnail generates a thumbnail for the video file
func (a *App) generateThumbnail(videoPath string) (string, error) {
	// Create thumbnails directory if it doesn't exist
	thumbnailsDir := "thumbnails"
	if err := os.MkdirAll(thumbnailsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create thumbnails directory: %w", err)
	}

	// Generate unique filename based on video path hash
	hash := md5.Sum([]byte(videoPath))
	thumbnailFilename := hex.EncodeToString(hash[:]) + ".jpg"
	thumbnailPath := filepath.Join(thumbnailsDir, thumbnailFilename)

	// Check if thumbnail already exists
	if _, err := os.Stat(thumbnailPath); err == nil {
		log.Printf("[THUMBNAIL] Using existing thumbnail: %s", thumbnailPath)
		return thumbnailPath, nil
	}

	log.Printf("[THUMBNAIL] Generating new thumbnail for: %s", videoPath)

	// Use ffmpeg to generate thumbnail at 10% of video duration
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-ss", "00:00:03", // Seek to 3 seconds
		"-vframes", "1", // Extract 1 frame
		"-vf", "scale=320:240:force_original_aspect_ratio=decrease,pad=320:240:(ow-iw)/2:(oh-ih)/2", // Scale to 320x240 with padding
		"-q:v", "2", // High quality
		"-y", // Overwrite output file
		thumbnailPath,
	)

	// Run ffmpeg command
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[THUMBNAIL] ffmpeg error: %v, output: %s", err, string(output))
		return "", fmt.Errorf("ffmpeg failed: %w", err)
	}

	log.Printf("[THUMBNAIL] Successfully generated: %s", thumbnailPath)
	return thumbnailPath, nil
}

// getThumbnailURL returns a URL for the video thumbnail
func (a *App) getThumbnailURL(filePath string) string {
	if !a.isVideoFile(filePath) {
		return ""
	}

	// Encode file path for URL safety
	encodedPath := url.QueryEscape(filePath)
	return fmt.Sprintf("/api/thumbnail/%s", encodedPath)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// CreateProject creates a new project with a default path
func (a *App) CreateProject(name, description string) (*projects.ProjectResponse, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.CreateProject(name, description)
}

// GetProjects returns all projects
func (a *App) GetProjects() ([]*projects.ProjectResponse, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.GetProjects()
}

// GetProjectByID returns a project by its ID
func (a *App) GetProjectByID(id int) (*projects.ProjectResponse, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.GetProjectByID(id)
}

// UpdateProject updates an existing project
func (a *App) UpdateProject(id int, name, description string) (*projects.ProjectResponse, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.UpdateProject(id, name, description)
}

// DeleteProject deletes a project by its ID
func (a *App) DeleteProject(id int) error {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.DeleteProject(id)
}

// CreateVideoClip creates a new video clip with file validation
func (a *App) CreateVideoClip(projectID int, filePath string) (*projects.VideoClipResponse, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.CreateVideoClip(projectID, filePath)
}

// GetVideoClipsByProject returns all video clips for a project
func (a *App) GetVideoClipsByProject(projectID int) ([]*projects.VideoClipResponse, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.GetVideoClipsByProject(projectID)
}

// UpdateVideoClip updates a video clip's metadata
func (a *App) UpdateVideoClip(id int, name, description string) (*projects.VideoClipResponse, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.UpdateVideoClip(id, name, description)
}

// DeleteVideoClip deletes a video clip
func (a *App) DeleteVideoClip(id int) error {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.DeleteVideoClip(id)
}

// SelectVideoFiles opens a file dialog to select video files
func (a *App) SelectVideoFiles() ([]*projects.LocalVideoFile, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.SelectVideoFiles(a.ctx)
}

// GetVideoFileInfo returns information about a local video file
func (a *App) GetVideoFileInfo(filePath string) (*projects.LocalVideoFile, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.GetVideoFileInfo(filePath)
}

// GetVideoURL returns a URL that can be used to access the video file via AssetServer
func (a *App) GetVideoURL(filePath string) (string, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.GetVideoURL(filePath)
}

// Helper functions needed by HTTP handlers and other parts of the app

// isVideoFile checks if a file is a supported video format
func (a *App) isVideoFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	videoExtensions := []string{".mp4", ".mov", ".avi", ".mkv", ".wmv", ".flv", ".webm", ".m4v", ".mpg", ".mpeg"}

	for _, validExt := range videoExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

// getFileInfo extracts file information from the filesystem
func (a *App) getFileInfo(filePath string) (int64, string, bool) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, "", false
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	format := strings.TrimPrefix(ext, ".")

	return fileInfo.Size(), format, true
}

// Close closes the database connection
func (a *App) Close() error {
	return a.client.Close()
}

// SaveSetting saves a setting key-value pair to the database
func (a *App) SaveSetting(key, value string) error {
	service := settings.NewSettingsService(a.client, a.ctx)
	return service.SaveSetting(key, value)
}

// GetSetting retrieves a setting value by key from the database
func (a *App) GetSetting(key string) (string, error) {
	service := settings.NewSettingsService(a.client, a.ctx)
	return service.GetSetting(key)
}

// DeleteSetting removes a setting from the database
func (a *App) DeleteSetting(key string) error {
	service := settings.NewSettingsService(a.client, a.ctx)
	return service.DeleteSetting(key)
}

// SaveOpenAIApiKey saves the OpenAI API key securely
func (a *App) SaveOpenAIApiKey(apiKey string) error {
	return a.SaveSetting("openai_api_key", apiKey)
}

// GetOpenAIApiKey retrieves the OpenAI API key
func (a *App) GetOpenAIApiKey() (string, error) {
	return a.GetSetting("openai_api_key")
}

// DeleteOpenAIApiKey removes the OpenAI API key
func (a *App) DeleteOpenAIApiKey() error {
	return a.DeleteSetting("openai_api_key")
}

// SaveOpenRouterApiKey saves the OpenRouter API key securely
func (a *App) SaveOpenRouterApiKey(apiKey string) error {
	return a.SaveSetting("openrouter_api_key", apiKey)
}

// GetOpenRouterApiKey retrieves the OpenRouter API key
func (a *App) GetOpenRouterApiKey() (string, error) {
	return a.GetSetting("openrouter_api_key")
}

// DeleteOpenRouterApiKey removes the OpenRouter API key
func (a *App) DeleteOpenRouterApiKey() error {
	return a.DeleteSetting("openrouter_api_key")
}

// SaveThemePreference saves the user's preferred theme (light or dark)
func (a *App) SaveThemePreference(theme string) error {
	service := settings.NewSettingsService(a.client, a.ctx)
	return service.SaveThemePreference(theme)
}

// GetThemePreference retrieves the user's preferred theme, defaults to "light"
func (a *App) GetThemePreference() (string, error) {
	service := settings.NewSettingsService(a.client, a.ctx)
	return service.GetThemePreference()
}

// Type aliases for AI service responses
type TestOpenAIApiKeyResponse = ai.TestOpenAIApiKeyResponse
type TestOpenRouterApiKeyResponse = ai.TestOpenRouterApiKeyResponse

// TestOpenAIApiKey tests if the stored OpenAI API key is valid
func (a *App) TestOpenAIApiKey() (*TestOpenAIApiKeyResponse, error) {
	service := ai.NewTranscriptionService(a.client, a.ctx)
	return service.TestOpenAIApiKey()
}

// TestOpenRouterApiKey tests if the stored OpenRouter API key is valid
func (a *App) TestOpenRouterApiKey() (*TestOpenRouterApiKeyResponse, error) {
	service := ai.NewTranscriptionService(a.client, a.ctx)
	return service.TestOpenRouterApiKey()
}

// Word represents a single word with timing information
type Word struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// Highlight represents a highlighted text region with timestamps
type Highlight struct {
	ID    string  `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Color string  `json:"color"`
}

// TranscribeVideoClip transcribes audio from a video clip using the AI service
func (a *App) TranscribeVideoClip(clipID int) (*ai.TranscriptionResponse, error) {
	transcriptionService := ai.NewTranscriptionService(a.client, a.ctx)
	return transcriptionService.TranscribeVideoClip(clipID)
}

// UpdateVideoClipHighlights updates the highlights for a video clip
func (a *App) UpdateVideoClipHighlights(clipID int, highlights []projects.Highlight) error {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.UpdateVideoClipHighlights(clipID, highlights)
}

// UpdateVideoClipSuggestedHighlights updates the suggested highlights for a video clip
func (a *App) UpdateVideoClipSuggestedHighlights(clipID int, suggestedHighlights []projects.Highlight) error {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.UpdateVideoClipSuggestedHighlights(clipID, suggestedHighlights)
}

// DeleteHighlight removes a specific highlight from a video clip by highlight ID
func (a *App) DeleteHighlight(clipID int, highlightID string) error {
	service := highlights.NewHighlightService(a.client, a.ctx)
	return service.DeleteHighlight(clipID, highlightID)
}

// Type aliases for backwards compatibility
type HighlightWithText = highlights.HighlightWithText
type ProjectHighlight = highlights.ProjectHighlight
type ProjectHighlightAISettings = highlights.ProjectHighlightAISettings
type HighlightSuggestion = highlights.HighlightSuggestion

// GetProjectHighlights returns all highlights from all video clips in a project
func (a *App) GetProjectHighlights(projectID int) ([]ProjectHighlight, error) {
	service := highlights.NewHighlightService(a.client, a.ctx)
	return service.GetProjectHighlights(projectID)
}

// UpdateProjectHighlightOrder updates the custom order of highlights for a project
func (a *App) UpdateProjectHighlightOrder(projectID int, highlightOrder []string) error {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.UpdateProjectHighlightOrder(projectID, highlightOrder)
}

// GetProjectHighlightOrder retrieves the custom highlight order for a project
func (a *App) GetProjectHighlightOrder(projectID int) ([]string, error) {
	service := highlights.NewHighlightService(a.client, a.ctx)
	return service.GetProjectHighlightOrder(projectID)
}

// ReorderHighlightsWithAI uses OpenRouter API to intelligently reorder highlights
func (a *App) ReorderHighlightsWithAI(projectID int, customPrompt string) ([]string, error) {
	service := highlights.NewAIService(a.client, a.ctx)
	return service.ReorderHighlightsWithAI(projectID, customPrompt, a.GetOpenRouterApiKey, a.GetProjectHighlights)
}

// Export-related type aliases
type ExportProgress = exports.ExportProgress
type HighlightSegment = highlights.HighlightSegment

// SelectExportFolder opens a dialog for the user to select an export folder
func (a *App) SelectExportFolder() (string, error) {
	service := exports.NewExportService(a.client, a.ctx)
	return service.SelectExportFolder(a.ctx)
}

// ExportStitchedHighlights exports all highlights as a single stitched video
func (a *App) ExportStitchedHighlights(projectID int, outputFolder string) (string, error) {
	service := exports.NewExportService(a.client, a.ctx)
	return service.ExportStitchedHighlights(projectID, outputFolder)
}

// ExportIndividualHighlights exports each highlight as a separate file
func (a *App) ExportIndividualHighlights(projectID int, outputFolder string) (string, error) {
	service := exports.NewExportService(a.client, a.ctx)
	return service.ExportIndividualHighlights(projectID, outputFolder)
}

// GetExportProgress returns the current progress of an export job
func (a *App) GetExportProgress(jobID string) (*ExportProgress, error) {
	service := exports.NewExportService(a.client, a.ctx)
	return service.GetExportProgress(jobID)
}

// CancelExport cancels an ongoing export job
func (a *App) CancelExport(jobID string) error {
	service := exports.NewExportService(a.client, a.ctx)
	return service.CancelExport(jobID)
}

// GetProjectExportJobs returns all export jobs for a project
func (a *App) GetProjectExportJobs(projectID int) ([]*ExportProgress, error) {
	service := exports.NewExportService(a.client, a.ctx)
	return service.GetProjectExportJobs(projectID)
}

// RecoverActiveExportJobs restores export jobs that were running when the app was closed
func (a *App) RecoverActiveExportJobs() error {
	service := exports.NewExportService(a.client, a.ctx)
	return service.RecoverActiveExportJobs()
}

// GetProjectAISettings gets the AI settings for a specific project
func (a *App) GetProjectAISettings(projectID int) (*highlights.ProjectAISettings, error) {
	service := highlights.NewAIService(a.client, a.ctx)
	return service.GetProjectAISettings(projectID)
}

// SaveProjectAISettings saves the AI settings for a specific project
func (a *App) SaveProjectAISettings(projectID int, settings highlights.ProjectAISettings) error {
	service := highlights.NewAIService(a.client, a.ctx)
	return service.SaveProjectAISettings(projectID, settings)
}

// GetProjectAISuggestion retrieves cached AI suggestion for a project
func (a *App) GetProjectAISuggestion(projectID int) (*highlights.ProjectAISuggestion, error) {
	service := highlights.NewAIService(a.client, a.ctx)
	return service.GetProjectAISuggestion(projectID)
}

// GetProjectHighlightAISettings retrieves AI settings for highlight suggestions
func (a *App) GetProjectHighlightAISettings(projectID int) (*ProjectHighlightAISettings, error) {
	service := highlights.NewHighlightService(a.client, a.ctx)
	return service.GetProjectHighlightAISettings(projectID)
}

// SaveProjectHighlightAISettings saves AI settings for highlight suggestions
func (a *App) SaveProjectHighlightAISettings(projectID int, settings ProjectHighlightAISettings) error {
	service := highlights.NewHighlightService(a.client, a.ctx)
	return service.SaveProjectHighlightAISettings(projectID, settings)
}

// SuggestHighlightsWithAI generates AI-powered highlight suggestions for a video
func (a *App) SuggestHighlightsWithAI(projectID int, videoID int, customPrompt string) ([]HighlightSuggestion, error) {
	service := highlights.NewAIService(a.client, a.ctx)
	return service.SuggestHighlightsWithAI(projectID, videoID, customPrompt, a.GetOpenRouterApiKey)
}

// GetSuggestedHighlights retrieves saved suggested highlights for a video
func (a *App) GetSuggestedHighlights(videoID int) ([]HighlightSuggestion, error) {
	service := highlights.NewHighlightService(a.client, a.ctx)
	return service.GetSuggestedHighlights(videoID)
}

// ClearSuggestedHighlights removes all suggested highlights for a video
func (a *App) ClearSuggestedHighlights(videoID int) error {
	service := highlights.NewHighlightService(a.client, a.ctx)
	return service.ClearSuggestedHighlights(videoID)
}
