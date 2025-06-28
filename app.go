package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"MYAPP/ent"
	"MYAPP/ent/project"
	"MYAPP/ent/schema"
	"MYAPP/ent/settings"
	"MYAPP/ent/videoclip"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ProjectResponse represents a project response for the frontend
type ProjectResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// VideoClipResponse represents a video clip response for the frontend
type VideoClipResponse struct {
	ID                     int     `json:"id"`
	Name                   string  `json:"name"`
	Description            string  `json:"description"`
	FilePath               string  `json:"filePath"`
	FileName               string  `json:"fileName"`
	FileSize               int64   `json:"fileSize"`
	Duration               float64 `json:"duration"`
	Format                 string  `json:"format"`
	Width                  int     `json:"width"`
	Height                 int     `json:"height"`
	ProjectID              int     `json:"projectId"`
	CreatedAt              string  `json:"createdAt"`
	UpdatedAt              string  `json:"updatedAt"`
	Exists                 bool    `json:"exists"`
	ThumbnailURL           string  `json:"thumbnailUrl"`
	Transcription          string  `json:"transcription"`
	TranscriptionWords     []Word      `json:"transcriptionWords"`
	TranscriptionLanguage  string      `json:"transcriptionLanguage"`
	TranscriptionDuration  float64     `json:"transcriptionDuration"`
	Highlights             []Highlight `json:"highlights"`
}

// LocalVideoFile represents a local video file for the frontend
type LocalVideoFile struct {
	Name     string `json:"name"`
	FilePath string `json:"filePath"`
	FileName string `json:"fileName"`
	FileSize int64  `json:"fileSize"`
	Format   string `json:"format"`
	Exists   bool   `json:"exists"`
}

// App struct
type App struct {
	ctx    context.Context
	client *ent.Client
}

// Helper function to convert schema.Word to Word
func schemaWordsToWords(schemaWords []schema.Word) []Word {
	words := make([]Word, len(schemaWords))
	for i, sw := range schemaWords {
		words[i] = Word{
			Word:  sw.Word,
			Start: sw.Start,
			End:   sw.End,
		}
	}
	return words
}

// Helper function to convert schema.Highlight to Highlight
func schemaHighlightsToHighlights(schemaHighlights []schema.Highlight) []Highlight {
	highlights := make([]Highlight, len(schemaHighlights))
	for i, sh := range schemaHighlights {
		highlights[i] = Highlight{
			ID:    sh.ID,
			Start: sh.Start,
			End:   sh.End,
			Color: sh.Color,
		}
	}
	return highlights
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
		"-vframes", "1",   // Extract 1 frame
		"-vf", "scale=320:240:force_original_aspect_ratio=decrease,pad=320:240:(ow-iw)/2:(oh-ih)/2", // Scale to 320x240 with padding
		"-q:v", "2",       // High quality
		"-y",              // Overwrite output file
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
func (a *App) CreateProject(name, description string) (*ProjectResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}

	// Create a default project path
	projectPath := filepath.Join("projects", name)

	project, err := a.client.Project.
		Create().
		SetName(name).
		SetDescription(description).
		SetPath(projectPath).
		Save(a.ctx)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return &ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		Path:        project.Path,
		CreatedAt:   project.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   project.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// GetProjects returns all projects
func (a *App) GetProjects() ([]*ProjectResponse, error) {
	projects, err := a.client.Project.
		Query().
		WithVideoClips().
		All(a.ctx)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	var responses []*ProjectResponse
	for _, project := range projects {
		responses = append(responses, &ProjectResponse{
			ID:          project.ID,
			Name:        project.Name,
			Description: project.Description,
			Path:        project.Path,
			CreatedAt:   project.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   project.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return responses, nil
}

// GetProjectByID returns a project by its ID
func (a *App) GetProjectByID(id int) (*ProjectResponse, error) {
	project, err := a.client.Project.
		Query().
		Where(project.ID(id)).
		WithVideoClips().
		Only(a.ctx)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get project with ID %d: %w", id, err)
	}

	return &ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		Path:        project.Path,
		CreatedAt:   project.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   project.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateProject updates an existing project
func (a *App) UpdateProject(id int, name, description string) (*ProjectResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}

	// Update the project path if name changed
	projectPath := filepath.Join("projects", name)

	updatedProject, err := a.client.Project.
		UpdateOneID(id).
		SetName(name).
		SetDescription(description).
		SetPath(projectPath).
		Save(a.ctx)
	
	if err != nil {
		return nil, fmt.Errorf("failed to update project with ID %d: %w", id, err)
	}

	return &ProjectResponse{
		ID:          updatedProject.ID,
		Name:        updatedProject.Name,
		Description: updatedProject.Description,
		Path:        updatedProject.Path,
		CreatedAt:   updatedProject.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   updatedProject.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// DeleteProject deletes a project by its ID
func (a *App) DeleteProject(id int) error {
	err := a.client.Project.
		DeleteOneID(id).
		Exec(a.ctx)
	
	if err != nil {
		return fmt.Errorf("failed to delete project with ID %d: %w", id, err)
	}

	return nil
}

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

// CreateVideoClip creates a new video clip with file validation
func (a *App) CreateVideoClip(projectID int, filePath string) (*VideoClipResponse, error) {
	// Validate file exists and is a video
	if !a.isVideoFile(filePath) {
		return nil, fmt.Errorf("file is not a supported video format")
	}
	
	fileSize, format, exists := a.getFileInfo(filePath)
	if !exists {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}
	
	// Check if this file path already exists for this project
	existingClip, err := a.client.VideoClip.
		Query().
		Where(
			videoclip.HasProjectWith(project.ID(projectID)),
			videoclip.FilePath(filePath),
		).
		Only(a.ctx)
	
	if err == nil {
		// File already exists for this project, return the existing clip
		fileName := filepath.Base(existingClip.FilePath)
		_, _, fileExists := a.getFileInfo(existingClip.FilePath)
		
		return &VideoClipResponse{
			ID:                    existingClip.ID,
			Name:                  existingClip.Name,
			Description:           existingClip.Description,
			FilePath:              existingClip.FilePath,
			FileName:              fileName,
			FileSize:              existingClip.FileSize,
			Duration:              existingClip.Duration,
			Format:                existingClip.Format,
			Width:                 existingClip.Width,
			Height:                existingClip.Height,
			ProjectID:             projectID,
			CreatedAt:             existingClip.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:             existingClip.UpdatedAt.Format("2006-01-02 15:04:05"),
			Exists:                fileExists,
			ThumbnailURL:          a.getThumbnailURL(existingClip.FilePath),
			Transcription:         existingClip.Transcription,
			TranscriptionWords:    schemaWordsToWords(existingClip.TranscriptionWords),
			TranscriptionLanguage: existingClip.TranscriptionLanguage,
			TranscriptionDuration: existingClip.TranscriptionDuration,
			Highlights:            schemaHighlightsToHighlights(existingClip.Highlights),
		}, fmt.Errorf("video file already added to this project")
	}
	
	fileName := filepath.Base(filePath)
	name := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	
	// Create video clip in database
	videoClip, err := a.client.VideoClip.
		Create().
		SetName(name).
		SetDescription("").
		SetFilePath(filePath).
		SetFormat(format).
		SetFileSize(fileSize).
		SetProjectID(projectID).
		Save(a.ctx)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create video clip: %w", err)
	}
	
	return &VideoClipResponse{
		ID:           videoClip.ID,
		Name:         videoClip.Name,
		Description:  videoClip.Description,
		FilePath:     videoClip.FilePath,
		FileName:              fileName,
		FileSize:              videoClip.FileSize,
		Duration:              videoClip.Duration,
		Format:                videoClip.Format,
		Width:                 videoClip.Width,
		Height:                videoClip.Height,
		ProjectID:             projectID,
		CreatedAt:             videoClip.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:             videoClip.UpdatedAt.Format("2006-01-02 15:04:05"),
		Exists:                true,
		ThumbnailURL:          a.getThumbnailURL(videoClip.FilePath),
		Transcription:         videoClip.Transcription,
		TranscriptionWords:    schemaWordsToWords(videoClip.TranscriptionWords),
		TranscriptionLanguage: videoClip.TranscriptionLanguage,
		TranscriptionDuration: videoClip.TranscriptionDuration,
		Highlights:            schemaHighlightsToHighlights(videoClip.Highlights),
	}, nil
}

// GetVideoClipsByProject returns all video clips for a project
func (a *App) GetVideoClipsByProject(projectID int) ([]*VideoClipResponse, error) {
	clips, err := a.client.VideoClip.
		Query().
		Where(videoclip.HasProjectWith(project.ID(projectID))).
		All(a.ctx)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get video clips: %w", err)
	}
	
	var responses []*VideoClipResponse
	for _, clip := range clips {
		fileName := filepath.Base(clip.FilePath)
		_, _, exists := a.getFileInfo(clip.FilePath)
		
		responses = append(responses, &VideoClipResponse{
			ID:                    clip.ID,
			Name:                  clip.Name,
			Description:           clip.Description,
			FilePath:              clip.FilePath,
			FileName:              fileName,
			FileSize:              clip.FileSize,
			Duration:              clip.Duration,
			Format:                clip.Format,
			Width:                 clip.Width,
			Height:                clip.Height,
			ProjectID:             projectID,
			CreatedAt:             clip.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:             clip.UpdatedAt.Format("2006-01-02 15:04:05"),
			Exists:                exists,
			ThumbnailURL:          a.getThumbnailURL(clip.FilePath),
			Transcription:         clip.Transcription,
			TranscriptionWords:    schemaWordsToWords(clip.TranscriptionWords),
			TranscriptionLanguage: clip.TranscriptionLanguage,
			TranscriptionDuration: clip.TranscriptionDuration,
			Highlights:            schemaHighlightsToHighlights(clip.Highlights),
		})
	}
	
	return responses, nil
}

// UpdateVideoClip updates a video clip's metadata
func (a *App) UpdateVideoClip(id int, name, description string) (*VideoClipResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("video clip name cannot be empty")
	}
	
	updatedClip, err := a.client.VideoClip.
		UpdateOneID(id).
		SetName(name).
		SetDescription(description).
		Save(a.ctx)
	
	if err != nil {
		return nil, fmt.Errorf("failed to update video clip: %w", err)
	}
	
	fileName := filepath.Base(updatedClip.FilePath)
	_, _, exists := a.getFileInfo(updatedClip.FilePath)
	
	return &VideoClipResponse{
		ID:                    updatedClip.ID,
		Name:                  updatedClip.Name,
		Description:           updatedClip.Description,
		FilePath:              updatedClip.FilePath,
		FileName:              fileName,
		FileSize:              updatedClip.FileSize,
		Duration:              updatedClip.Duration,
		Format:                updatedClip.Format,
		Width:                 updatedClip.Width,
		Height:                updatedClip.Height,
		ProjectID:             updatedClip.Edges.Project.ID,
		CreatedAt:             updatedClip.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:             updatedClip.UpdatedAt.Format("2006-01-02 15:04:05"),
		Exists:                exists,
		ThumbnailURL:          a.getThumbnailURL(updatedClip.FilePath),
		Transcription:         updatedClip.Transcription,
		TranscriptionWords:    schemaWordsToWords(updatedClip.TranscriptionWords),
		TranscriptionLanguage: updatedClip.TranscriptionLanguage,
		TranscriptionDuration: updatedClip.TranscriptionDuration,
		Highlights:            schemaHighlightsToHighlights(updatedClip.Highlights),
	}, nil
}

// DeleteVideoClip deletes a video clip
func (a *App) DeleteVideoClip(id int) error {
	err := a.client.VideoClip.
		DeleteOneID(id).
		Exec(a.ctx)
	
	if err != nil {
		return fmt.Errorf("failed to delete video clip: %w", err)
	}
	
	return nil
}

// SelectVideoFiles opens a file dialog to select video files
func (a *App) SelectVideoFiles() ([]*LocalVideoFile, error) {
	// Open file dialog for multiple video files
	filePaths, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Video Files",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Video Files",
				Pattern:     "*.mp4;*.mov;*.avi;*.mkv;*.wmv;*.flv;*.webm;*.m4v;*.mpg;*.mpeg",
			},
		},
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to open file dialog: %w", err)
	}

	var videoFiles []*LocalVideoFile
	for _, filePath := range filePaths {
		if !a.isVideoFile(filePath) {
			continue // Skip non-video files
		}

		fileSize, format, exists := a.getFileInfo(filePath)
		if !exists {
			continue // Skip files that don't exist
		}

		fileName := filepath.Base(filePath)
		name := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		videoFiles = append(videoFiles, &LocalVideoFile{
			Name:     name,
			FilePath: filePath,
			FileName: fileName,
			FileSize: fileSize,
			Format:   format,
			Exists:   true,
		})
	}

	return videoFiles, nil
}

// GetVideoFileInfo returns information about a local video file
func (a *App) GetVideoFileInfo(filePath string) (*LocalVideoFile, error) {
	if !a.isVideoFile(filePath) {
		return nil, fmt.Errorf("file is not a supported video format")
	}
	
	fileSize, format, exists := a.getFileInfo(filePath)
	if !exists {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}
	
	fileName := filepath.Base(filePath)
	name := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	
	return &LocalVideoFile{
		Name:     name,
		FilePath: filePath,
		FileName: fileName,
		FileSize: fileSize,
		Format:   format,
		Exists:   true,
	}, nil
}

// GetVideoURL returns a URL that can be used to access the video file via AssetServer
func (a *App) GetVideoURL(filePath string) (string, error) {
	if !a.isVideoFile(filePath) {
		return "", fmt.Errorf("file is not a supported video format")
	}
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", filePath)
	}
	
	// Encode file path for URL safety
	encodedPath := url.QueryEscape(filePath)
	videoURL := fmt.Sprintf("/api/video/%s", encodedPath)
	
	log.Printf("[VIDEO] Original path: %s", filePath)
	log.Printf("[VIDEO] Encoded path: %s", encodedPath)
	log.Printf("[VIDEO] Final URL: %s", videoURL)
	
	// Return AssetServer URL that will work in the webview
	return videoURL, nil
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

// Segment represents a segment of transcription with timing
type Segment struct {
	ID     int    `json:"id"`
	Seek   int    `json:"seek"`
	Start  float64 `json:"start"`
	End    float64 `json:"end"`
	Text   string  `json:"text"`
	Tokens []int   `json:"tokens"`
	Temperature float64 `json:"temperature"`
	AvgLogprob  float64 `json:"avg_logprob"`
	CompressionRatio float64 `json:"compression_ratio"`
	NoSpeechProb     float64 `json:"no_speech_prob"`
	Words   []Word  `json:"words"`
}

// WhisperResponse represents the detailed response from OpenAI Whisper API
type WhisperResponse struct {
	Task     string    `json:"task"`
	Language string    `json:"language"`
	Duration float64   `json:"duration"`
	Text     string    `json:"text"`
	Segments []Segment `json:"segments"`
	Words    []Word    `json:"words"`
}

// TranscriptionResponse represents the response from the transcription process
type TranscriptionResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Transcription string `json:"transcription,omitempty"`
	Words     []Word `json:"words,omitempty"`
	Language  string `json:"language,omitempty"`
	Duration  float64 `json:"duration,omitempty"`
}

// TranscribeVideoClip extracts audio from a video and transcribes it using OpenAI Whisper
func (a *App) TranscribeVideoClip(clipID int) (*TranscriptionResponse, error) {
	// Get the video clip
	clip, err := a.client.VideoClip.Get(a.ctx, clipID)
	if err != nil {
		return &TranscriptionResponse{
			Success: false,
			Message: "Video clip not found",
		}, nil
	}

	// Check if file exists
	if _, err := os.Stat(clip.FilePath); os.IsNotExist(err) {
		return &TranscriptionResponse{
			Success: false,
			Message: "Video file not found",
		}, nil
	}

	// Get OpenAI API key
	apiKey, err := a.GetOpenAIApiKey()
	if err != nil || apiKey == "" {
		return &TranscriptionResponse{
			Success: false,
			Message: "OpenAI API key not configured",
		}, nil
	}

	// Extract audio from video
	audioPath, err := a.extractAudio(clip.FilePath)
	if err != nil {
		return &TranscriptionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to extract audio: %v", err),
		}, nil
	}
	defer os.Remove(audioPath) // Clean up temporary audio file

	// Transcribe audio using OpenAI Whisper
	whisperResponse, err := a.transcribeAudio(audioPath, apiKey)
	if err != nil {
		return &TranscriptionResponse{
			Success: false,
			Message: fmt.Sprintf("Transcription failed: %v", err),
		}, nil
	}

	// Convert Word structs for storage
	var wordsForStorage []schema.Word
	for _, w := range whisperResponse.Words {
		wordsForStorage = append(wordsForStorage, schema.Word{
			Word:  w.Word,
			Start: w.Start,
			End:   w.End,
		})
	}

	// Save transcription to database
	_, err = a.client.VideoClip.
		UpdateOneID(clipID).
		SetTranscription(whisperResponse.Text).
		SetTranscriptionWords(wordsForStorage).
		SetTranscriptionLanguage(whisperResponse.Language).
		SetTranscriptionDuration(whisperResponse.Duration).
		Save(a.ctx)
	
	if err != nil {
		return &TranscriptionResponse{
			Success: false,
			Message: "Failed to save transcription",
		}, nil
	}

	return &TranscriptionResponse{
		Success:       true,
		Message:       "Transcription completed successfully",
		Transcription: whisperResponse.Text,
		Words:         whisperResponse.Words,
		Language:      whisperResponse.Language,
		Duration:      whisperResponse.Duration,
	}, nil
}

// extractAudio extracts audio from a video file using ffmpeg
func (a *App) extractAudio(videoPath string) (string, error) {
	// Create temp directory for audio files
	tempDir := "temp_audio"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate unique audio filename
	hash := md5.Sum([]byte(videoPath + fmt.Sprintf("%d", time.Now().UnixNano())))
	audioFilename := hex.EncodeToString(hash[:]) + ".mp3"
	audioPath := filepath.Join(tempDir, audioFilename)

	log.Printf("[TRANSCRIPTION] Extracting audio from: %s to: %s", videoPath, audioPath)

	// Use ffmpeg to extract audio
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vn",                    // No video
		"-acodec", "mp3",         // Audio codec
		"-ar", "16000",           // Sample rate (16kHz for Whisper)
		"-ac", "1",               // Mono channel
		"-b:a", "64k",            // Bitrate
		"-y",                     // Overwrite output file
		audioPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[TRANSCRIPTION] ffmpeg error: %v, output: %s", err, string(output))
		return "", fmt.Errorf("ffmpeg failed: %w", err)
	}

	log.Printf("[TRANSCRIPTION] Audio extracted successfully: %s", audioPath)
	return audioPath, nil
}

// transcribeAudio sends audio to OpenAI Whisper API for transcription
func (a *App) transcribeAudio(audioPath, apiKey string) (*WhisperResponse, error) {
	// Create HTTP client with longer timeout for transcription
	client := &http.Client{
		Timeout: 120 * time.Second, // 2 minutes for transcription
	}

	// Open audio file
	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file field
	fileWriter, err := writer.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}

	// Add model field
	err = writer.WriteField("model", "whisper-1")
	if err != nil {
		return nil, fmt.Errorf("failed to add model field: %w", err)
	}

	// Add response format field for verbose JSON with timestamps
	err = writer.WriteField("response_format", "verbose_json")
	if err != nil {
		return nil, fmt.Errorf("failed to add response format field: %w", err)
	}

	// Add timestamp granularities for word-level timestamps
	err = writer.WriteField("timestamp_granularities[]", "word")
	if err != nil {
		return nil, fmt.Errorf("failed to add timestamp granularities field: %w", err)
	}

	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	log.Printf("[TRANSCRIPTION] Sending audio to OpenAI Whisper API")

	// Make request
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
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var whisperResponse WhisperResponse
	err = json.Unmarshal(body, &whisperResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transcription response: %w", err)
	}

	log.Printf("[TRANSCRIPTION] Transcription completed, text length: %d characters, words: %d", 
		len(whisperResponse.Text), len(whisperResponse.Words))

	return &whisperResponse, nil
}

// UpdateVideoClipHighlights updates the highlights for a video clip
func (a *App) UpdateVideoClipHighlights(clipID int, highlights []Highlight) error {
	// Convert Highlights to schema.Highlights for database storage
	var schemaHighlights []schema.Highlight
	for _, h := range highlights {
		schemaHighlights = append(schemaHighlights, schema.Highlight{
			ID:    h.ID,
			Start: h.Start,
			End:   h.End,
			Color: h.Color,
		})
	}

	// Update the video clip with new highlights
	_, err := a.client.VideoClip.
		UpdateOneID(clipID).
		SetHighlights(schemaHighlights).
		Save(a.ctx)
	
	if err != nil {
		return fmt.Errorf("failed to update video clip highlights: %w", err)
	}

	return nil
}