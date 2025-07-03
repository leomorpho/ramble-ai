package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"MYAPP/ent"
	"MYAPP/ent/exportjob"
	"MYAPP/ent/project"
	"MYAPP/ent/schema"
	"MYAPP/ent/settings"
	"MYAPP/goapp/ai"
	"MYAPP/goapp/highlights"
	"MYAPP/goapp/projects"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)


// ProjectAISettings represents AI settings for a project
type ProjectAISettings struct {
	AIModel  string `json:"aiModel"`
	AIPrompt string `json:"aiPrompt"`
}

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
	if key == "" {
		return fmt.Errorf("setting key cannot be empty")
	}

	// Check if setting already exists
	existingSetting, err := a.client.Settings.
		Query().
		Where(settings.Key(key)).
		Only(a.ctx)

	if err != nil {
		// Setting doesn't exist, create new one
		_, err = a.client.Settings.
			Create().
			SetKey(key).
			SetValue(value).
			Save(a.ctx)

		if err != nil {
			return fmt.Errorf("failed to create setting: %w", err)
		}
	} else {
		// Setting exists, update it
		_, err = a.client.Settings.
			UpdateOne(existingSetting).
			SetValue(value).
			Save(a.ctx)

		if err != nil {
			return fmt.Errorf("failed to update setting: %w", err)
		}
	}

	return nil
}

// GetSetting retrieves a setting value by key from the database
func (a *App) GetSetting(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("setting key cannot be empty")
	}

	setting, err := a.client.Settings.
		Query().
		Where(settings.Key(key)).
		Only(a.ctx)

	if err != nil {
		// Return empty string if setting doesn't exist
		return "", nil
	}

	return setting.Value, nil
}

// DeleteSetting removes a setting from the database
func (a *App) DeleteSetting(key string) error {
	if key == "" {
		return fmt.Errorf("setting key cannot be empty")
	}

	_, err := a.client.Settings.
		Delete().
		Where(settings.Key(key)).
		Exec(a.ctx)

	if err != nil {
		return fmt.Errorf("failed to delete setting: %w", err)
	}

	return nil
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
	if theme != "light" && theme != "dark" {
		return fmt.Errorf("theme must be either 'light' or 'dark'")
	}
	return a.SaveSetting("theme_preference", theme)
}

// GetThemePreference retrieves the user's preferred theme, defaults to "light"
func (a *App) GetThemePreference() (string, error) {
	theme, err := a.GetSetting("theme_preference")
	if err != nil {
		return "light", err
	}
	if theme == "" {
		return "light", nil // Default to light theme
	}
	return theme, nil
}

// TestOpenAIApiKeyResponse represents the response from testing the API key
type TestOpenAIApiKeyResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
	Model   string `json:"model,omitempty"`
}

// TestOpenRouterApiKeyResponse represents the response from testing the OpenRouter API key
type TestOpenRouterApiKeyResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
	Model   string `json:"model,omitempty"`
}

// TestOpenAIApiKey tests if the stored OpenAI API key is valid
func (a *App) TestOpenAIApiKey() (*TestOpenAIApiKeyResponse, error) {
	// Get the stored API key
	apiKey, err := a.GetOpenAIApiKey()
	if err != nil {
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: "Failed to retrieve API key from database",
		}, nil
	}

	if apiKey == "" {
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: "No API key found. Please set your OpenAI API key first.",
		}, nil
	}

	// Test the API key with a simple request to the models endpoint
	return a.testOpenAIConnection(apiKey)
}

// testOpenAIConnection makes a test request to OpenAI API
func (a *App) testOpenAIConnection(apiKey string) (*TestOpenAIApiKeyResponse, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request to list models (lightweight endpoint)
	req, err := http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
	if err != nil {
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: "Failed to create test request",
		}, nil
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: "Failed to connect to OpenAI API. Please check your internet connection.",
		}, nil
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: "Failed to read API response",
		}, nil
	}

	// Check response status
	switch resp.StatusCode {
	case http.StatusOK:
		// Parse response to get a model name
		var modelsResp struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}

		if err := json.Unmarshal(body, &modelsResp); err == nil && len(modelsResp.Data) > 0 {
			// Find Whisper model or use first available
			modelName := modelsResp.Data[0].ID
			for _, model := range modelsResp.Data {
				if strings.Contains(model.ID, "whisper") {
					modelName = model.ID
					break
				}
			}

			return &TestOpenAIApiKeyResponse{
				Valid:   true,
				Message: "API key is valid and working!",
				Model:   modelName,
			}, nil
		}

		return &TestOpenAIApiKeyResponse{
			Valid:   true,
			Message: "API key is valid and working!",
		}, nil

	case http.StatusUnauthorized:
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: "Invalid API key. Please check your OpenAI API key.",
		}, nil

	case http.StatusTooManyRequests:
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: "Rate limit exceeded. Please try again later.",
		}, nil

	case http.StatusForbidden:
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: "API key doesn't have sufficient permissions.",
		}, nil

	default:
		return &TestOpenAIApiKeyResponse{
			Valid:   false,
			Message: fmt.Sprintf("API test failed with status %d: %s", resp.StatusCode, string(body)),
		}, nil
	}
}

// TestOpenRouterApiKey tests if the stored OpenRouter API key is valid
func (a *App) TestOpenRouterApiKey() (*TestOpenRouterApiKeyResponse, error) {
	// Get the stored API key
	apiKey, err := a.GetOpenRouterApiKey()
	if err != nil {
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: "Failed to retrieve API key from database",
		}, nil
	}

	if apiKey == "" {
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: "No API key found. Please set your OpenRouter API key first.",
		}, nil
	}

	// Test the API key with a simple request to the models endpoint
	return a.testOpenRouterConnection(apiKey)
}

// testOpenRouterConnection makes a test request to OpenRouter API
func (a *App) testOpenRouterConnection(apiKey string) (*TestOpenRouterApiKeyResponse, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request to list models (lightweight endpoint)
	req, err := http.NewRequest("GET", "https://openrouter.ai/api/v1/models", nil)
	if err != nil {
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: "Failed to create test request",
		}, nil
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: "Failed to connect to OpenRouter API. Please check your internet connection.",
		}, nil
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: "Failed to read API response",
		}, nil
	}

	// Check response status
	switch resp.StatusCode {
	case http.StatusOK:
		// Parse response to get a model name
		var modelsResp struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}

		if err := json.Unmarshal(body, &modelsResp); err == nil && len(modelsResp.Data) > 0 {
			// Find a suitable model or use first available
			modelName := modelsResp.Data[0].ID
			for _, model := range modelsResp.Data {
				if strings.Contains(strings.ToLower(model.ID), "gpt") || strings.Contains(strings.ToLower(model.ID), "claude") {
					modelName = model.ID
					break
				}
			}

			return &TestOpenRouterApiKeyResponse{
				Valid:   true,
				Message: "API key is valid and working!",
				Model:   modelName,
			}, nil
		}

		return &TestOpenRouterApiKeyResponse{
			Valid:   true,
			Message: "API key is valid and working!",
		}, nil

	case http.StatusUnauthorized:
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: "Invalid API key. Please check your OpenRouter API key.",
		}, nil

	case http.StatusTooManyRequests:
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: "Rate limit exceeded. Please try again later.",
		}, nil

	case http.StatusForbidden:
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: "API key doesn't have sufficient permissions.",
		}, nil

	default:
		return &TestOpenRouterApiKeyResponse{
			Valid:   false,
			Message: fmt.Sprintf("API test failed with status %d: %s", resp.StatusCode, string(body)),
		}, nil
	}
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
	// Get OpenRouter API key
	apiKey, err := a.GetOpenRouterApiKey()
	if err != nil || apiKey == "" {
		return nil, fmt.Errorf("OpenRouter API key not configured")
	}

	// Get project AI settings
	aiSettings, err := a.GetProjectAISettings(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project AI settings: %w", err)
	}

	// Use custom prompt if provided, otherwise use project's saved prompt
	prompt := customPrompt
	if prompt == "" {
		prompt = aiSettings.AIPrompt
	}

	// Get all project highlights
	projectHighlights, err := a.GetProjectHighlights(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project highlights: %w", err)
	}

	if len(projectHighlights) == 0 {
		return []string{}, nil
	}

	// Create a minimal map of ID to highlight text for AI processing
	highlightMap := make(map[string]string)
	var highlightIDs []string

	for _, ph := range projectHighlights {
		for _, highlight := range ph.Highlights {
			highlightMap[highlight.ID] = highlight.Text
			highlightIDs = append(highlightIDs, highlight.ID)
		}
	}

	if len(highlightMap) == 0 {
		return []string{}, nil
	}

	// Call OpenRouter API to get AI reordering
	reorderedIDs, err := a.callOpenRouterForReordering(apiKey, aiSettings.AIModel, highlightMap, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI reordering: %w", err)
	}

	// Validate that all IDs are present in the reordered list
	if len(reorderedIDs) != len(highlightIDs) {
		log.Printf("AI reordering returned %d IDs but expected %d", len(reorderedIDs), len(highlightIDs))
		// Fallback to original order if counts don't match
		return highlightIDs, nil
	}

	// Validate that all original IDs are present
	originalIDSet := make(map[string]bool)
	for _, id := range highlightIDs {
		originalIDSet[id] = true
	}

	for _, id := range reorderedIDs {
		if !originalIDSet[id] {
			log.Printf("AI reordering returned unknown ID: %s", id)
			// Fallback to original order if unknown IDs are present
			return highlightIDs, nil
		}
	}

	// Save AI suggestion to database
	err = a.saveAISuggestion(projectID, reorderedIDs, aiSettings.AIModel)
	if err != nil {
		log.Printf("Failed to save AI suggestion to database: %v", err)
		// Don't fail the request if saving fails, just log the error
	}

	return reorderedIDs, nil
}


// callOpenRouterForReordering calls the OpenRouter API to get intelligent highlight reordering
func (a *App) callOpenRouterForReordering(apiKey string, model string, highlightMap map[string]string, customPrompt string) ([]string, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 60 * time.Second, // AI requests can take longer
	}

	// Build the prompt for AI reordering
	prompt := a.buildReorderingPrompt(highlightMap, customPrompt)

	// Create request payload
	requestData := highlights.OpenRouterRequest{
		Model: model, // Use the project-specific model
		Messages: []highlights.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/yourusername/video-app") // Required by OpenRouter
	req.Header.Set("X-Title", "Video Highlight Reordering")                     // Optional but recommended

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenRouter API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var openRouterResp highlights.OpenRouterResponse
	err = json.Unmarshal(body, &openRouterResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if openRouterResp.Error != nil {
		return nil, fmt.Errorf("OpenRouter API error: %s", openRouterResp.Error.Message)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices received from AI")
	}

	// Extract the reordered IDs from the AI response
	aiResponse := openRouterResp.Choices[0].Message.Content
	reorderedIDs, err := a.parseAIReorderingResponse(aiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return reorderedIDs, nil
}

// buildReorderingPrompt creates a prompt for the AI to reorder highlights intelligently
func (a *App) buildReorderingPrompt(highlightMap map[string]string, customPrompt string) string {
	// Use default YouTube expert prompt if no custom prompt provided
	var basePrompt string
	if customPrompt != "" {
		basePrompt = customPrompt
	} else {
		basePrompt = `You are an expert YouTuber and content creator with millions of subscribers, known for creating highly engaging videos that maximize viewer retention and satisfaction. Your task is to reorder these video highlight segments to create the highest quality video possible.

Reorder these segments using your expertise in:
- Hook creation and audience retention
- Storytelling and narrative structure
- Pacing and rhythm for maximum engagement
- Building emotional connections with viewers
- Creating viral-worthy content flow
- Strategic placement of key moments

Feel free to completely restructure the order - move any segment to any position if it will improve video quality and viewer experience.`
	}

	prompt := basePrompt + `

Here are the video highlight segments:

`

	// Convert map to sorted slice for consistent ordering in prompt
	type highlightEntry struct {
		id   string
		text string
	}
	var entries []highlightEntry
	for id, text := range highlightMap {
		entries = append(entries, highlightEntry{id: id, text: text})
	}

	// Sort entries by ID for consistent ordering
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].id < entries[j].id
	})

	for i, entry := range entries {
		prompt += fmt.Sprintf("%d. ID: %s\n", i+1, entry.id)
		prompt += fmt.Sprintf("   Content: %s\n\n", entry.text)
	}

	prompt += `

Analyze these segments and reorder them to create the highest quality video possible for maximum viewer engagement and retention.

IMPORTANT: Respond with ONLY a JSON array containing the highlight IDs in the new order. Do not include any explanation, reasoning, or additional text.

Example format: ["id1", "id2", "id3", ...]`

	return prompt
}

// parseAIReorderingResponse extracts the reordered highlight IDs from the AI response
func (a *App) parseAIReorderingResponse(response string) ([]string, error) {
	// Clean the response - remove any markdown formatting
	cleanResponse := strings.TrimSpace(response)
	cleanResponse = strings.Trim(cleanResponse, "`")
	if strings.HasPrefix(cleanResponse, "json") {
		cleanResponse = strings.TrimPrefix(cleanResponse, "json")
		cleanResponse = strings.TrimSpace(cleanResponse)
	}

	// Try to parse as JSON array
	var reorderedIDs []string
	err := json.Unmarshal([]byte(cleanResponse), &reorderedIDs)
	if err != nil {
		// If direct parsing fails, try to extract JSON from the response
		// Look for JSON array pattern
		jsonStart := strings.Index(cleanResponse, "[")
		jsonEnd := strings.LastIndex(cleanResponse, "]")

		if jsonStart >= 0 && jsonEnd > jsonStart {
			jsonPart := cleanResponse[jsonStart : jsonEnd+1]
			err = json.Unmarshal([]byte(jsonPart), &reorderedIDs)
			if err != nil {
				return nil, fmt.Errorf("failed to parse JSON array from AI response: %w", err)
			}
		} else {
			return nil, fmt.Errorf("no valid JSON array found in AI response")
		}
	}

	return reorderedIDs, nil
}

// Export-related types and structs
type ExportProgress struct {
	JobID          string  `json:"jobId"`
	Stage          string  `json:"stage"`
	Progress       float64 `json:"progress"`
	CurrentFile    string  `json:"currentFile"`
	TotalFiles     int     `json:"totalFiles"`
	ProcessedFiles int     `json:"processedFiles"`
	IsComplete     bool    `json:"isComplete"`
	HasError       bool    `json:"hasError"`
	ErrorMessage   string  `json:"errorMessage"`
	IsCancelled    bool    `json:"isCancelled"`
}

type ActiveExportJob struct {
	JobID    string
	Cancel   chan bool
	IsActive bool
}

// Global active job manager (for cancellation and in-memory tracking)
var (
	activeJobs      = make(map[string]*ActiveExportJob)
	activeJobsMutex = sync.RWMutex{}
)

// FFmpeg progress tracking
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

// SelectExportFolder opens a dialog for the user to select an export folder
func (a *App) SelectExportFolder() (string, error) {
	options := runtime.OpenDialogOptions{
		Title:   "Select Export Folder",
		Filters: []runtime.FileFilter{},
	}

	folder, err := runtime.OpenDirectoryDialog(a.ctx, options)
	if err != nil {
		return "", fmt.Errorf("failed to open directory dialog: %w", err)
	}

	return folder, nil
}

// ExportStitchedHighlights exports all highlights as a single stitched video
func (a *App) ExportStitchedHighlights(projectID int, outputFolder string) (string, error) {
	jobID := fmt.Sprintf("stitched_%d_%d", projectID, time.Now().UnixNano())

	// Get project info for directory naming
	project, err := a.client.Project.Get(a.ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("failed to get project: %w", err)
	}

	// Create timestamped directory
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	projectDirName := fmt.Sprintf("%s_%s", sanitizeFilename(project.Name), timestamp)
	projectDir := filepath.Join(outputFolder, projectDirName)

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create database record with project directory path
	dbJob, err := a.client.ExportJob.Create().
		SetJobID(jobID).
		SetExportType("stitched").
		SetOutputPath(projectDir).
		SetStage("starting").
		SetProgress(0.0).
		SetProjectID(projectID).
		Save(a.ctx)

	if err != nil {
		return "", fmt.Errorf("failed to create export job: %w", err)
	}

	// Create active job for cancellation tracking
	activeJob := &ActiveExportJob{
		JobID:    jobID,
		Cancel:   make(chan bool, 1),
		IsActive: true,
	}

	activeJobsMutex.Lock()
	activeJobs[jobID] = activeJob
	activeJobsMutex.Unlock()

	// Start export job in goroutine
	go a.performStitchedExport(dbJob, activeJob)

	return jobID, nil
}

// ExportIndividualHighlights exports each highlight as a separate file
func (a *App) ExportIndividualHighlights(projectID int, outputFolder string) (string, error) {
	jobID := fmt.Sprintf("individual_%d_%d", projectID, time.Now().UnixNano())

	// Get project info for directory naming
	project, err := a.client.Project.Get(a.ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("failed to get project: %w", err)
	}

	// Create timestamped directory
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	projectDirName := fmt.Sprintf("%s_%s", sanitizeFilename(project.Name), timestamp)
	projectDir := filepath.Join(outputFolder, projectDirName)

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create database record with project directory path
	dbJob, err := a.client.ExportJob.Create().
		SetJobID(jobID).
		SetExportType("individual").
		SetOutputPath(projectDir).
		SetStage("starting").
		SetProgress(0.0).
		SetProjectID(projectID).
		Save(a.ctx)

	if err != nil {
		return "", fmt.Errorf("failed to create export job: %w", err)
	}

	// Create active job for cancellation tracking
	activeJob := &ActiveExportJob{
		JobID:    jobID,
		Cancel:   make(chan bool, 1),
		IsActive: true,
	}

	activeJobsMutex.Lock()
	activeJobs[jobID] = activeJob
	activeJobsMutex.Unlock()

	// Start export job in goroutine
	go a.performIndividualExport(dbJob, activeJob)

	return jobID, nil
}

// GetExportProgress returns the current progress of an export job
func (a *App) GetExportProgress(jobID string) (*ExportProgress, error) {
	// Get job from database
	dbJob, err := a.client.ExportJob.
		Query().
		Where(exportjob.JobID(jobID)).
		First(a.ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("export job not found: %s", jobID)
		}
		return nil, fmt.Errorf("failed to get export job: %w", err)
	}

	return &ExportProgress{
		JobID:          dbJob.JobID,
		Stage:          dbJob.Stage,
		Progress:       dbJob.Progress,
		CurrentFile:    dbJob.CurrentFile,
		TotalFiles:     dbJob.TotalFiles,
		ProcessedFiles: dbJob.ProcessedFiles,
		IsComplete:     dbJob.IsComplete,
		HasError:       dbJob.HasError,
		ErrorMessage:   dbJob.ErrorMessage,
		IsCancelled:    dbJob.IsCancelled,
	}, nil
}

// CancelExport cancels an ongoing export job
func (a *App) CancelExport(jobID string) error {
	// Update database to mark as cancelled
	_, err := a.client.ExportJob.
		Update().
		Where(exportjob.JobID(jobID)).
		SetIsCancelled(true).
		SetStage("cancelled").
		SetUpdatedAt(time.Now()).
		Save(a.ctx)

	if err != nil {
		return fmt.Errorf("failed to cancel export job in database: %w", err)
	}

	// Signal active job to cancel
	activeJobsMutex.RLock()
	activeJob, exists := activeJobs[jobID]
	activeJobsMutex.RUnlock()

	if exists && activeJob.IsActive {
		select {
		case activeJob.Cancel <- true:
		default:
		}
	}

	return nil
}

// performStitchedExport performs the actual stitched video export
func (a *App) performStitchedExport(dbJob *ent.ExportJob, activeJob *ActiveExportJob) {
	defer func() {
		// Mark job as complete and clean up
		activeJobsMutex.Lock()
		delete(activeJobs, dbJob.JobID)
		activeJobsMutex.Unlock()

		// Update database with completion status
		a.client.ExportJob.
			UpdateOne(dbJob).
			SetIsComplete(true).
			SetCompletedAt(time.Now()).
			SetUpdatedAt(time.Now()).
			Save(a.ctx)
	}()

	// Update stage to preparing
	a.updateJobProgress(dbJob.JobID, "preparing", 0.0, "", 0, 0)

	// Get project ID from job
	project, err := dbJob.QueryProject().First(a.ctx)
	if err != nil {
		a.updateJobError(dbJob.JobID, fmt.Sprintf("Failed to get project: %v", err))
		return
	}

	// Get all highlight segments for this project (in proper order)
	segments, err := a.getProjectHighlightsForExport(project.ID)
	if err != nil {
		a.updateJobError(dbJob.JobID, fmt.Sprintf("Failed to get highlight segments: %v", err))
		return
	}

	if len(segments) == 0 {
		a.updateJobError(dbJob.JobID, "No highlight segments found")
		return
	}

	// Update total files count
	a.updateJobProgress(dbJob.JobID, "preparing", 0.0, "", len(segments), 0)

	// Create temp directory for clips
	tempDir := filepath.Join("temp_export", dbJob.JobID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		a.updateJobError(dbJob.JobID, fmt.Sprintf("Failed to create temp directory: %v", err))
		return
	}
	defer os.RemoveAll(tempDir)

	// Extract individual segments with progress tracking
	var segmentPaths []string
	for i, segment := range segments {
		// Check for cancellation
		select {
		case <-activeJob.Cancel:
			a.updateJobCancelled(dbJob.JobID)
			return
		default:
		}

		// Update progress
		progress := float64(i) / float64(len(segments)) * 0.8 // 80% for extraction
		fileName := fmt.Sprintf("%s (%.1fs-%.1fs)", segment.VideoClipName, segment.Start, segment.End)
		a.updateJobProgress(dbJob.JobID, "extracting", progress, fileName, len(segments), i)

		segmentPath, err := a.extractHighlightSegmentWithProgress(segment, tempDir, i+1, dbJob.JobID, activeJob.Cancel)
		if err != nil {
			a.updateJobError(dbJob.JobID, fmt.Sprintf("Failed to extract segment %s: %v", fileName, err))
			return
		}
		segmentPaths = append(segmentPaths, segmentPath)
	}

	// Stitch segments together with progress tracking
	a.updateJobProgress(dbJob.JobID, "stitching", 0.8, "Combining highlight segments", len(segments), len(segments))

	// Create output file in the project directory
	outputFileName := fmt.Sprintf("%s_highlights_stitched.mp4", sanitizeFilename(project.Name))
	outputFile := filepath.Join(dbJob.OutputPath, outputFileName)
	err = a.stitchVideoClipsWithProgress(segmentPaths, outputFile, dbJob.JobID, activeJob.Cancel)
	if err != nil {
		a.updateJobError(dbJob.JobID, fmt.Sprintf("Failed to stitch segments: %v", err))
		return
	}

	// Mark as complete with directory info
	completionMessage := fmt.Sprintf("Exported to: %s", filepath.Base(dbJob.OutputPath))
	a.updateJobProgress(dbJob.JobID, "complete", 1.0, completionMessage, len(segments), len(segments))
}

// performIndividualExport performs the individual clips export
func (a *App) performIndividualExport(dbJob *ent.ExportJob, activeJob *ActiveExportJob) {
	defer func() {
		// Mark job as complete and clean up
		activeJobsMutex.Lock()
		delete(activeJobs, dbJob.JobID)
		activeJobsMutex.Unlock()

		// Update database with completion status
		a.client.ExportJob.
			UpdateOne(dbJob).
			SetIsComplete(true).
			SetCompletedAt(time.Now()).
			SetUpdatedAt(time.Now()).
			Save(a.ctx)
	}()

	// Get project ID from job
	project, err := dbJob.QueryProject().First(a.ctx)
	if err != nil {
		a.updateJobError(dbJob.JobID, fmt.Sprintf("Failed to get project: %v", err))
		return
	}

	// Get all highlight segments for this project (in proper order)
	segments, err := a.getProjectHighlightsForExport(project.ID)
	if err != nil {
		a.updateJobError(dbJob.JobID, fmt.Sprintf("Failed to get highlight segments: %v", err))
		return
	}

	if len(segments) == 0 {
		a.updateJobError(dbJob.JobID, "No highlight segments found")
		return
	}

	// Update stage and total files
	a.updateJobProgress(dbJob.JobID, "extracting", 0.0, "", len(segments), 0)

	// Extract individual segments with progress tracking
	for i, segment := range segments {
		// Check for cancellation
		select {
		case <-activeJob.Cancel:
			a.updateJobCancelled(dbJob.JobID)
			return
		default:
		}

		// Update progress
		progress := float64(i) / float64(len(segments))
		fileName := fmt.Sprintf("%s (%.1fs-%.1fs)", segment.VideoClipName, segment.Start, segment.End)
		a.updateJobProgress(dbJob.JobID, "extracting", progress, fileName, len(segments), i)

		// Create descriptive filename with segment info
		segmentName := fmt.Sprintf("%s_%.1fs-%.1fs",
			sanitizeFilename(strings.TrimSuffix(segment.VideoClipName, filepath.Ext(segment.VideoClipName))),
			segment.Start, segment.End)
		outputFile := filepath.Join(dbJob.OutputPath, fmt.Sprintf("%03d_%s.mp4", i+1, segmentName))

		err := a.extractHighlightSegmentDirectWithProgress(segment, outputFile, dbJob.JobID, activeJob.Cancel)
		if err != nil {
			a.updateJobError(dbJob.JobID, fmt.Sprintf("Failed to extract segment %s: %v", fileName, err))
			return
		}
	}

	// Mark as complete with directory info
	completionMessage := fmt.Sprintf("Exported to: %s", filepath.Base(dbJob.OutputPath))
	a.updateJobProgress(dbJob.JobID, "complete", 1.0, completionMessage, len(segments), len(segments))
}

// getProjectHighlightsForExport gets all highlights across all clips in the proper order
func (a *App) getProjectHighlightsForExport(projectID int) ([]HighlightSegment, error) {
	service := highlights.NewHighlightService(a.client, a.ctx)
	segments, err := service.GetProjectHighlightsForExport(projectID)
	if err != nil {
		return nil, err
	}

	// Apply custom ordering if it exists
	order, err := service.GetProjectHighlightOrder(projectID)
	if err == nil && len(order) > 0 {
		segments = service.ApplyHighlightOrder(segments, order)
	}

	return segments, nil
}


// extractHighlightSegment extracts a single highlight segment to a temp file
func (a *App) extractHighlightSegment(segment HighlightSegment, tempDir string, index int) (string, error) {
	outputPath := filepath.Join(tempDir, fmt.Sprintf("segment_%03d.mp4", index))

	// Use ffmpeg to extract the segment
	cmd := exec.Command("ffmpeg",
		"-i", segment.VideoPath,
		"-ss", fmt.Sprintf("%.3f", segment.Start),
		"-to", fmt.Sprintf("%.3f", segment.End),
		"-c:v", "libx264",
		"-c:a", "aac",
		"-y",
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w, output: %s", err, string(output))
	}

	return outputPath, nil
}

// extractHighlightSegmentDirect extracts a highlight segment directly to the output file
func (a *App) extractHighlightSegmentDirect(segment HighlightSegment, outputPath string) error {
	// Use ffmpeg to extract the segment
	cmd := exec.Command("ffmpeg",
		"-i", segment.VideoPath,
		"-ss", fmt.Sprintf("%.3f", segment.Start),
		"-to", fmt.Sprintf("%.3f", segment.End),
		"-c:v", "libx264",
		"-c:a", "aac",
		"-y",
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w, output: %s", err, string(output))
	}

	return nil
}

// stitchVideoClips combines multiple video clips into one
func (a *App) stitchVideoClips(clipPaths []string, outputPath string) error {
	if len(clipPaths) == 0 {
		return fmt.Errorf("no clips to stitch")
	}

	// Create concat file for ffmpeg
	concatFile := filepath.Join(filepath.Dir(outputPath), "concat_list.txt")
	defer os.Remove(concatFile)

	file, err := os.Create(concatFile)
	if err != nil {
		return fmt.Errorf("failed to create concat file: %w", err)
	}
	defer file.Close()

	for _, clipPath := range clipPaths {
		_, err := file.WriteString(fmt.Sprintf("file '%s'\n", clipPath))
		if err != nil {
			return fmt.Errorf("failed to write to concat file: %w", err)
		}
	}

	// Use ffmpeg to concatenate clips
	cmd := exec.Command("ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", concatFile,
		"-c", "copy",
		"-y",
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg concat failed: %w, output: %s", err, string(output))
	}

	return nil
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(filename string) string {
	// Replace invalid characters with underscores
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := filename
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}

// Database update helper functions
func (a *App) updateJobProgress(jobID, stage string, progress float64, currentFile string, totalFiles, processedFiles int) {
	a.client.ExportJob.
		Update().
		Where(exportjob.JobID(jobID)).
		SetStage(stage).
		SetProgress(progress).
		SetCurrentFile(currentFile).
		SetTotalFiles(totalFiles).
		SetProcessedFiles(processedFiles).
		SetUpdatedAt(time.Now()).
		Save(a.ctx)
}

func (a *App) updateJobError(jobID, errorMessage string) {
	a.client.ExportJob.
		Update().
		Where(exportjob.JobID(jobID)).
		SetHasError(true).
		SetErrorMessage(errorMessage).
		SetIsComplete(true).
		SetCompletedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(a.ctx)
}

func (a *App) updateJobCancelled(jobID string) {
	a.client.ExportJob.
		Update().
		Where(exportjob.JobID(jobID)).
		SetIsCancelled(true).
		SetStage("cancelled").
		SetIsComplete(true).
		SetCompletedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(a.ctx)
}

// FFmpeg progress tracking functions
func (a *App) parseFFmpegProgress(line string) *FFmpegProgress {
	// FFmpeg progress line format: frame=   123 fps= 12 q=28.0 size=    1234kB time=00:01:23.45 bitrate= 567.8kbits/s speed=1.23x
	frameRegex := regexp.MustCompile(`frame=\s*(\d+)`)
	fpsRegex := regexp.MustCompile(`fps=\s*([\d.]+)`)
	timeRegex := regexp.MustCompile(`time=(\d{2}):(\d{2}):([\d.]+)`)
	bitrateRegex := regexp.MustCompile(`bitrate=\s*([\d.]+)kbits/s`)

	progress := &FFmpegProgress{}

	if match := frameRegex.FindStringSubmatch(line); len(match) > 1 {
		if frame, err := strconv.ParseInt(match[1], 10, 64); err == nil {
			progress.Frame = frame
		}
	}

	if match := fpsRegex.FindStringSubmatch(line); len(match) > 1 {
		if fps, err := strconv.ParseFloat(match[1], 64); err == nil {
			progress.FPS = fps
		}
	}

	if match := timeRegex.FindStringSubmatch(line); len(match) > 3 {
		hours, _ := strconv.ParseFloat(match[1], 64)
		minutes, _ := strconv.ParseFloat(match[2], 64)
		seconds, _ := strconv.ParseFloat(match[3], 64)
		progress.Time = hours*3600 + minutes*60 + seconds
	}

	if match := bitrateRegex.FindStringSubmatch(line); len(match) > 1 {
		progress.Bitrate = match[1] + "kbits/s"
	}

	return progress
}

func (a *App) getVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		videoPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get video duration: %w", err)
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

// Enhanced ffmpeg functions with progress tracking
func (a *App) extractHighlightSegmentWithProgress(segment HighlightSegment, tempDir string, index int, jobID string, cancel chan bool) (string, error) {
	outputPath := filepath.Join(tempDir, fmt.Sprintf("segment_%03d.mp4", index))

	// Get video duration for the highlight segment
	duration := segment.End - segment.Start

	cmd := exec.Command("ffmpeg",
		"-i", segment.VideoPath,
		"-ss", fmt.Sprintf("%.3f", segment.Start),
		"-to", fmt.Sprintf("%.3f", segment.End),
		"-c:v", "libx264",
		"-c:a", "aac",
		"-progress", "pipe:1",
		"-y",
		outputPath,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Monitor progress
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "time=") {
				if progress := a.parseFFmpegProgress(line); progress.Time > 0 && duration > 0 {
					clipProgress := progress.Time / duration
					if clipProgress > 1.0 {
						clipProgress = 1.0
					}
					// Update progress for this specific clip extraction
					// This is a sub-progress within the overall job
				}
			}
		}
	}()

	// Wait for completion or cancellation
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-cancel:
		cmd.Process.Kill()
		return "", fmt.Errorf("extraction cancelled")
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("ffmpeg failed: %w", err)
		}
	}

	return outputPath, nil
}

func (a *App) extractHighlightSegmentDirectWithProgress(segment HighlightSegment, outputPath, jobID string, cancel chan bool) error {
	duration := segment.End - segment.Start

	cmd := exec.Command("ffmpeg",
		"-i", segment.VideoPath,
		"-ss", fmt.Sprintf("%.3f", segment.Start),
		"-to", fmt.Sprintf("%.3f", segment.End),
		"-c:v", "libx264",
		"-c:a", "aac",
		"-progress", "pipe:1",
		"-y",
		outputPath,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Monitor progress
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "time=") {
				if progress := a.parseFFmpegProgress(line); progress.Time > 0 && duration > 0 {
					clipProgress := progress.Time / duration
					if clipProgress > 1.0 {
						clipProgress = 1.0
					}
					// Could update sub-progress here if needed
				}
			}
		}
	}()

	// Wait for completion or cancellation
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-cancel:
		cmd.Process.Kill()
		return fmt.Errorf("extraction cancelled")
	case err := <-done:
		if err != nil {
			return fmt.Errorf("ffmpeg failed: %w", err)
		}
	}

	return nil
}

func (a *App) stitchVideoClipsWithProgress(clipPaths []string, outputPath, jobID string, cancel chan bool) error {
	if len(clipPaths) == 0 {
		return fmt.Errorf("no clips to stitch")
	}

	// Calculate total duration for progress tracking
	var totalDuration float64
	for _, clipPath := range clipPaths {
		if duration, err := a.getVideoDuration(clipPath); err == nil {
			totalDuration += duration
		}
	}

	// Create concat file for ffmpeg
	concatFile := filepath.Join(filepath.Dir(outputPath), "concat_list.txt")
	defer os.Remove(concatFile)

	file, err := os.Create(concatFile)
	if err != nil {
		return fmt.Errorf("failed to create concat file: %w", err)
	}
	defer file.Close()

	for _, clipPath := range clipPaths {
		_, err := file.WriteString(fmt.Sprintf("file '%s'\n", clipPath))
		if err != nil {
			return fmt.Errorf("failed to write to concat file: %w", err)
		}
	}

	cmd := exec.Command("ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", concatFile,
		"-c", "copy",
		"-progress", "pipe:1",
		"-y",
		outputPath,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Monitor progress
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "time=") {
				if progress := a.parseFFmpegProgress(line); progress.Time > 0 && totalDuration > 0 {
					stitchProgress := progress.Time / totalDuration
					if stitchProgress > 1.0 {
						stitchProgress = 1.0
					}
					// Update stitching progress (80% + 20% of stitching progress)
					overallProgress := 0.8 + (stitchProgress * 0.2)
					a.updateJobProgress(jobID, "stitching", overallProgress, "Combining clips", 0, 0)
				}
			}
		}
	}()

	// Wait for completion or cancellation
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-cancel:
		cmd.Process.Kill()
		return fmt.Errorf("stitching cancelled")
	case err := <-done:
		if err != nil {
			return fmt.Errorf("ffmpeg concat failed: %w", err)
		}
	}

	return nil
}

// GetProjectExportJobs returns all export jobs for a project
func (a *App) GetProjectExportJobs(projectID int) ([]*ExportProgress, error) {
	jobs, err := a.client.ExportJob.
		Query().
		Where(exportjob.HasProjectWith(project.ID(projectID))).
		Order(ent.Desc(exportjob.FieldCreatedAt)).
		All(a.ctx)

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

// GetProjectAISettings gets the AI settings for a specific project
func (a *App) GetProjectAISettings(projectID int) (*ProjectAISettings, error) {
	project, err := a.client.Project.
		Query().
		Where(project.ID(projectID)).
		Only(a.ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	aiModel := project.AiModel
	if aiModel == "" {
		aiModel = "anthropic/claude-3-haiku-20240307"
	}

	aiPrompt := project.AiPrompt

	return &ProjectAISettings{
		AIModel:  aiModel,
		AIPrompt: aiPrompt,
	}, nil
}

// SaveProjectAISettings saves the AI settings for a specific project
func (a *App) SaveProjectAISettings(projectID int, settings ProjectAISettings) error {
	_, err := a.client.Project.
		UpdateOneID(projectID).
		SetAiModel(settings.AIModel).
		SetAiPrompt(settings.AIPrompt).
		Save(a.ctx)

	if err != nil {
		return fmt.Errorf("failed to save project AI settings: %w", err)
	}

	return nil
}

// ProjectAISuggestion represents an AI suggestion for a project
type ProjectAISuggestion struct {
	Order     []string  `json:"order"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"createdAt"`
}


// Helper functions for word index and time conversion
func (a *App) timeToWordIndex(timeSeconds float64, transcriptWords []schema.Word) int {
	service := highlights.NewHighlightService(a.client, a.ctx)
	return service.TimeToWordIndex(timeSeconds, transcriptWords)
}

func (a *App) wordIndexToTime(wordIndex int, transcriptWords []schema.Word) float64 {
	service := highlights.NewHighlightService(a.client, a.ctx)
	return service.WordIndexToTime(wordIndex, transcriptWords)
}

// saveAISuggestion saves the AI suggestion to the database (internal helper)
func (a *App) saveAISuggestion(projectID int, reorderedIDs []string, model string) error {
	_, err := a.client.Project.
		UpdateOneID(projectID).
		SetAiSuggestionOrder(reorderedIDs).
		SetAiSuggestionModel(model).
		SetAiSuggestionCreatedAt(time.Now()).
		Save(a.ctx)

	if err != nil {
		return fmt.Errorf("failed to save AI suggestion: %w", err)
	}

	return nil
}

// GetProjectAISuggestion retrieves cached AI suggestion for a project
func (a *App) GetProjectAISuggestion(projectID int) (*ProjectAISuggestion, error) {
	project, err := a.client.Project.
		Query().
		Where(project.ID(projectID)).
		Only(a.ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Check if there's a cached AI suggestion
	if project.AiSuggestionOrder == nil {
		return nil, nil // No cached suggestion
	}

	return &ProjectAISuggestion{
		Order:     project.AiSuggestionOrder,
		Model:     project.AiSuggestionModel,
		CreatedAt: project.AiSuggestionCreatedAt,
	}, nil
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

// RecoverActiveExportJobs restores export jobs that were running when the app was closed
func (a *App) RecoverActiveExportJobs() error {
	// Find jobs that are not complete and not cancelled
	activeJobs, err := a.client.ExportJob.
		Query().
		Where(
			exportjob.IsComplete(false),
			exportjob.IsCancelled(false),
		).
		All(a.ctx)

	if err != nil {
		return fmt.Errorf("failed to get active export jobs: %w", err)
	}

	// Mark incomplete jobs as cancelled since we can't resume them
	for _, job := range activeJobs {
		log.Printf("Marking incomplete export job %s as cancelled", job.JobID)
		a.client.ExportJob.
			UpdateOne(job).
			SetIsCancelled(true).
			SetStage("cancelled").
			SetErrorMessage("Application was restarted during export").
			SetIsComplete(true).
			SetCompletedAt(time.Now()).
			SetUpdatedAt(time.Now()).
			Save(a.ctx)
	}

	return nil
}

// onFileDrop handles file drops from the OS using Wails v2 drag and drop API
func (a *App) onFileDrop(ctx context.Context, x, y int, paths []string) {
	log.Printf("Files dropped at (%d, %d): %v", x, y, paths)

	// Filter for video files only
	videoFiles := []string{}
	for _, path := range paths {
		if a.isVideoFile(path) {
			videoFiles = append(videoFiles, path)
		}
	}

	// Emit event to frontend with dropped video files
	runtime.EventsEmit(ctx, "files-dropped", map[string]interface{}{
		"x":     x,
		"y":     y,
		"paths": videoFiles,
	})
}
