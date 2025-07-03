package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"MYAPP/db"
	"MYAPP/ent/schema"
	"MYAPP/services"
	"MYAPP/utils"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// TranscriptionResponse represents the response from a transcription operation
type TranscriptionResponse struct {
	Success       bool    `json:"success"`
	Message       string  `json:"message"`
	Transcription string  `json:"transcription,omitempty"`
	Words         []Word  `json:"words,omitempty"`
	Language      string  `json:"language,omitempty"`
	Duration      float64 `json:"duration,omitempty"`
}

// Word represents a single word in a transcription with timing
type Word struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// Segment represents a segment in a WhisperResponse
type Segment struct {
	ID               int     `json:"id"`
	Seek             int     `json:"seek"`
	Start            float64 `json:"start"`
	End              float64 `json:"end"`
	Text             string  `json:"text"`
	Tokens           []int   `json:"tokens"`
	Temperature      float64 `json:"temperature"`
	AvgLogprob       float64 `json:"avg_logprob"`
	CompressionRatio float64 `json:"compression_ratio"`
	NoSpeechProb     float64 `json:"no_speech_prob"`
	Words            []Word  `json:"words"`
}

// WhisperResponse represents the response from OpenAI Whisper API
type WhisperResponse struct {
	Task     string    `json:"task"`
	Language string    `json:"language"`
	Duration float64   `json:"duration"`
	Text     string    `json:"text"`
	Segments []Segment `json:"segments"`
	Words    []Word    `json:"words"`
}

// App struct contains all services and dependencies
type App struct {
	ctx               context.Context
	repository        *db.Repository
	projectService    *services.ProjectService
	videoClipService  *services.VideoClipService
	settingsService   *services.SettingsService
}

// NewApp creates a new App application struct with all services
func NewApp() *App {
	// Initialize database repository
	repository, err := db.NewRepository()
	if err != nil {
		log.Fatalf("failed to initialize database repository: %v", err)
	}

	return &App{
		repository: repository,
	}
}

// Startup is called when the app starts. The context is saved so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize repository with context
	if err := a.repository.Initialize(ctx); err != nil {
		log.Printf("failed to initialize repository: %v", err)
		return
	}

	log.Println("Application started and database initialized")

	// Initialize services with the database client and context
	client := a.repository.GetClient()
	a.projectService = services.NewProjectService(client, ctx)
	a.videoClipService = services.NewVideoClipService(client, ctx)
	a.settingsService = services.NewSettingsService(client, ctx)
}

// Shutdown is called when the app shuts down
func (a *App) Shutdown(ctx context.Context) {
	// Close the database connection
	if err := a.repository.Close(); err != nil {
		log.Printf("failed to close database connection: %v", err)
	}
	log.Println("Application shutdown completed")
}

// GetContext returns the application context
func (a *App) GetContext() context.Context {
	return a.ctx
}

// GetProjectService returns the project service
func (a *App) GetProjectService() *services.ProjectService {
	return a.projectService
}

// GetVideoClipService returns the video clip service  
func (a *App) GetVideoClipService() *services.VideoClipService {
	return a.videoClipService
}

// GetSettingsService returns the settings service
func (a *App) GetSettingsService() *services.SettingsService {
	return a.settingsService
}

// GetRepository returns the database repository
func (a *App) GetRepository() *db.Repository {
	return a.repository
}

// OnFileDrop handles file drops from the OS using Wails v2 drag and drop API
func (a *App) OnFileDrop(ctx context.Context, x, y int, paths []string) {
	log.Printf("Files dropped at (%d, %d): %v", x, y, paths)
	
	// Filter for video files only
	videoFiles := []string{}
	for _, path := range paths {
		if utils.IsVideoFile(path) {
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

// Project Service Methods - Delegate to the project service

// CreateProject creates a new project
func (a *App) CreateProject(name, description string) (*services.ProjectResponse, error) {
	return a.projectService.CreateProject(name, description)
}

// GetProjects returns all projects
func (a *App) GetProjects() ([]*services.ProjectResponse, error) {
	return a.projectService.GetProjects()
}

// GetProjectByID returns a project by its ID
func (a *App) GetProjectByID(id int) (*services.ProjectResponse, error) {
	return a.projectService.GetProjectByID(id)
}

// UpdateProject updates an existing project
func (a *App) UpdateProject(id int, name, description string) (*services.ProjectResponse, error) {
	return a.projectService.UpdateProject(id, name, description)
}

// DeleteProject deletes a project by its ID
func (a *App) DeleteProject(id int) error {
	return a.projectService.DeleteProject(id)
}

// Video Clip Service Methods - Delegate to the video clip service

// CreateVideoClip creates a new video clip
func (a *App) CreateVideoClip(projectID int, filePath string) (*services.VideoClipResponse, error) {
	return a.videoClipService.CreateVideoClip(projectID, filePath)
}

// GetVideoClipsByProject returns all video clips for a project
func (a *App) GetVideoClipsByProject(projectID int) ([]*services.VideoClipResponse, error) {
	return a.videoClipService.GetVideoClipsByProject(projectID)
}

// UpdateVideoClip updates a video clip's metadata
func (a *App) UpdateVideoClip(id int, name, description string) (*services.VideoClipResponse, error) {
	return a.videoClipService.UpdateVideoClip(id, name, description)
}

// DeleteVideoClip deletes a video clip
func (a *App) DeleteVideoClip(id int) error {
	return a.videoClipService.DeleteVideoClip(id)
}

// SelectVideoFiles opens a file dialog to select video files
func (a *App) SelectVideoFiles() ([]*services.LocalVideoFile, error) {
	return a.videoClipService.SelectVideoFiles()
}

// GetVideoFileInfo returns information about a local video file
func (a *App) GetVideoFileInfo(filePath string) (*services.LocalVideoFile, error) {
	return a.videoClipService.GetVideoFileInfo(filePath)
}

// GetVideoURL returns a URL for accessing the video file
func (a *App) GetVideoURL(filePath string) (string, error) {
	return a.videoClipService.GetVideoURL(filePath)
}

// Settings Service Methods - Delegate to the settings service

// SaveSetting saves a setting key-value pair
func (a *App) SaveSetting(key, value string) error {
	return a.settingsService.SaveSetting(key, value)
}

// GetSetting retrieves a setting value by key
func (a *App) GetSetting(key string) (string, error) {
	return a.settingsService.GetSetting(key)
}

// DeleteSetting removes a setting
func (a *App) DeleteSetting(key string) error {
	return a.settingsService.DeleteSetting(key)
}

// SaveOpenAIApiKey saves the OpenAI API key
func (a *App) SaveOpenAIApiKey(apiKey string) error {
	return a.settingsService.SaveOpenAIApiKey(apiKey)
}

// GetOpenAIApiKey retrieves the OpenAI API key
func (a *App) GetOpenAIApiKey() (string, error) {
	return a.settingsService.GetOpenAIApiKey()
}

// DeleteOpenAIApiKey removes the OpenAI API key
func (a *App) DeleteOpenAIApiKey() error {
	return a.settingsService.DeleteOpenAIApiKey()
}

// SaveOpenRouterApiKey saves the OpenRouter API key
func (a *App) SaveOpenRouterApiKey(apiKey string) error {
	return a.settingsService.SaveOpenRouterApiKey(apiKey)
}

// GetOpenRouterApiKey retrieves the OpenRouter API key
func (a *App) GetOpenRouterApiKey() (string, error) {
	return a.settingsService.GetOpenRouterApiKey()
}

// DeleteOpenRouterApiKey removes the OpenRouter API key
func (a *App) DeleteOpenRouterApiKey() error {
	return a.settingsService.DeleteOpenRouterApiKey()
}

// GetThemePreference retrieves the user's theme preference
func (a *App) GetThemePreference() (string, error) {
	theme, err := a.settingsService.GetSetting("theme")
	if err != nil {
		return "light", err
	}
	if theme == "" {
		return "light", nil
	}
	return theme, nil
}

// SaveThemePreference saves the user's theme preference
func (a *App) SaveThemePreference(theme string) error {
	return a.settingsService.SaveSetting("theme", theme)
}

// TODO: Implement these functions properly when services are ready

// GetProjectAISettings gets AI settings for a project
func (a *App) GetProjectAISettings(projectID int) (map[string]string, error) {
	return map[string]string{"aiModel": "", "aiPrompt": ""}, nil
}

// SaveProjectAISettings saves AI settings for a project
func (a *App) SaveProjectAISettings(projectID int, settings map[string]string) error {
	return nil
}

// GetProjectAISuggestion gets AI suggestion for a project
func (a *App) GetProjectAISuggestion(projectID int) (map[string]interface{}, error) {
	return map[string]interface{}{"order": []string{}, "model": "", "createdAt": ""}, nil
}

// ReorderHighlightsWithAI reorders highlights using AI
func (a *App) ReorderHighlightsWithAI(projectID int, model string, prompt string) ([]string, error) {
	return []string{}, nil
}

// GetProjectHighlights gets highlights for a project
func (a *App) GetProjectHighlights(projectID int) ([]interface{}, error) {
	return []interface{}{}, nil
}

// GetProjectHighlightOrder gets highlight order for a project
func (a *App) GetProjectHighlightOrder(projectID int) ([]string, error) {
	return []string{}, nil
}

// UpdateProjectHighlightOrder updates highlight order for a project
func (a *App) UpdateProjectHighlightOrder(projectID int, order []string) error {
	return nil
}

// DeleteHighlight deletes a highlight
func (a *App) DeleteHighlight(videoClipID int, highlightID string) error {
	return nil
}

// UpdateVideoClipHighlights updates highlights for a video clip
func (a *App) UpdateVideoClipHighlights(videoClipID int, highlights []interface{}) error {
	return nil
}

// GetProjectHighlightAISettings gets highlight AI settings for a project
func (a *App) GetProjectHighlightAISettings(projectID int) (map[string]string, error) {
	return map[string]string{"aiModel": "", "aiPrompt": ""}, nil
}

// SaveProjectHighlightAISettings saves highlight AI settings for a project
func (a *App) SaveProjectHighlightAISettings(projectID int, settings map[string]string) error {
	return nil
}

// SuggestHighlightsWithAI suggests highlights using AI
func (a *App) SuggestHighlightsWithAI(videoClipID int, model string, prompt string) ([]interface{}, error) {
	return []interface{}{}, nil
}

// GetSuggestedHighlights gets suggested highlights for a video
func (a *App) GetSuggestedHighlights(videoClipID int) ([]interface{}, error) {
	return []interface{}{}, nil
}

// ClearSuggestedHighlights clears suggested highlights for a video
func (a *App) ClearSuggestedHighlights(videoClipID int) error {
	return nil
}

// UpdateVideoClipSuggestedHighlights updates suggested highlights for a video clip
func (a *App) UpdateVideoClipSuggestedHighlights(videoClipID int, highlights []interface{}) error {
	return nil
}

// TranscribeVideoClip transcribes a video clip using OpenAI Whisper
func (a *App) TranscribeVideoClip(videoClipID int) (*TranscriptionResponse, error) {
	// Get the video clip from database
	client := a.repository.GetClient()
	videoClip, err := client.VideoClip.Get(a.ctx, videoClipID)
	if err != nil {
		return &TranscriptionResponse{
			Success: false,
			Message: fmt.Sprintf("Video clip not found: %v", err),
		}, nil
	}

	// Check if file exists
	if _, err := os.Stat(videoClip.FilePath); os.IsNotExist(err) {
		return &TranscriptionResponse{
			Success: false,
			Message: "Video file not found on disk",
		}, nil
	}

	// Get OpenAI API key
	apiKey, err := a.settingsService.GetOpenAIApiKey()
	if err != nil || apiKey == "" {
		return &TranscriptionResponse{
			Success: false,
			Message: "OpenAI API key not configured",
		}, nil
	}

	// Extract audio from video
	log.Printf("Extracting audio from video: %s", videoClip.FilePath)
	audioFilePath, err := a.extractAudio(videoClip.FilePath)
	if err != nil {
		return &TranscriptionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to extract audio: %v", err),
		}, nil
	}
	defer os.Remove(audioFilePath) // Clean up temp file

	// Transcribe audio using OpenAI Whisper
	log.Printf("Transcribing audio file: %s", audioFilePath)
	whisperResponse, err := a.transcribeAudio(audioFilePath, apiKey)
	if err != nil {
		return &TranscriptionResponse{
			Success: false,
			Message: fmt.Sprintf("Transcription failed: %v", err),
		}, nil
	}

	// Store transcription in database
	// Convert Word types to schema.Word types
	schemaWords := make([]schema.Word, len(whisperResponse.Words))
	for i, word := range whisperResponse.Words {
		schemaWords[i] = schema.Word{
			Word:  word.Word,
			Start: word.Start,
			End:   word.End,
		}
	}
	
	_, err = client.VideoClip.UpdateOneID(videoClipID).
		SetTranscription(whisperResponse.Text).
		SetTranscriptionWords(schemaWords).
		SetTranscriptionLanguage(whisperResponse.Language).
		SetTranscriptionDuration(whisperResponse.Duration).
		Save(a.ctx)
	
	if err != nil {
		log.Printf("Failed to save transcription to database: %v", err)
		// Continue anyway, return the transcription even if save failed
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

// Export-related functions
func (a *App) SelectExportFolder() (string, error) {
	// TODO: Implement folder selection dialog
	return "/tmp", nil // Return a valid path instead of empty string
}

func (a *App) ExportStitchedHighlights(projectID int, exportPath string) (string, error) {
	// TODO: Implement stitched highlights export
	return "job-" + fmt.Sprintf("%d", projectID), nil // Return a job ID
}

func (a *App) ExportIndividualHighlights(projectID int, exportPath string) (string, error) {
	// TODO: Implement individual highlights export
	return "job-" + fmt.Sprintf("%d", projectID), nil // Return a job ID
}

func (a *App) GetExportProgress(jobID string) (map[string]interface{}, error) {
	return map[string]interface{}{"progress": 0, "status": "idle"}, nil
}

func (a *App) CancelExport(jobID string) error {
	return nil
}

func (a *App) GetProjectExportJobs(projectID int) ([]interface{}, error) {
	return []interface{}{}, nil
}

// Test functions for API keys
func (a *App) TestOpenAIApiKey() (map[string]interface{}, error) {
	// TODO: Implement actual API key testing
	return map[string]interface{}{
		"success": false, 
		"message": "API key testing not yet implemented",
		"valid":   false,
	}, nil
}

func (a *App) TestOpenRouterApiKey() (map[string]interface{}, error) {
	// TODO: Implement actual API key testing
	return map[string]interface{}{
		"success": false, 
		"message": "API key testing not yet implemented",
		"valid":   false,
	}, nil
}

// extractAudio extracts audio from a video file using ffmpeg
func (a *App) extractAudio(videoPath string) (string, error) {
	// Create temporary audio file
	tempDir := os.TempDir()
	audioFileName := fmt.Sprintf("temp_audio_%d.mp3", time.Now().UnixNano())
	audioPath := filepath.Join(tempDir, audioFileName)

	// Use ffmpeg to extract audio
	cmd := exec.Command("ffmpeg", 
		"-i", videoPath,           // Input video file
		"-vn",                     // Disable video recording
		"-acodec", "mp3",          // Audio codec
		"-ar", "16000",            // Sample rate (16kHz, optimal for Whisper)
		"-ac", "1",                // Mono channel
		"-y",                      // Overwrite output file
		audioPath,                 // Output audio file
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Printf("Running ffmpeg command: %s", strings.Join(cmd.Args, " "))
	
	if err := cmd.Run(); err != nil {
		log.Printf("FFmpeg error: %s", stderr.String())
		return "", fmt.Errorf("ffmpeg failed: %v - %s", err, stderr.String())
	}

	// Verify the audio file was created
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return "", fmt.Errorf("audio file was not created")
	}

	log.Printf("Audio extracted successfully: %s", audioPath)
	return audioPath, nil
}

// transcribeAudio transcribes an audio file using OpenAI Whisper API
func (a *App) transcribeAudio(audioPath, apiKey string) (*WhisperResponse, error) {
	// Open the audio file
	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %v", err)
	}
	defer file.Close()

	// Create multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file field
	fileWriter, err := writer.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %v", err)
	}
	
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %v", err)
	}

	// Add model field
	writer.WriteField("model", "whisper-1")
	
	// Add response format for word-level timestamps
	writer.WriteField("response_format", "verbose_json")
	writer.WriteField("timestamp_granularities[]", "word")

	writer.Close()

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Make request with timeout
	client := &http.Client{
		Timeout: 2 * time.Minute, // 2 minute timeout for transcription
	}

	log.Printf("Making OpenAI Whisper API request...")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var whisperResponse WhisperResponse
	if err := json.Unmarshal(body, &whisperResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	log.Printf("Transcription completed successfully, duration: %.2fs, language: %s", 
		whisperResponse.Duration, whisperResponse.Language)
	
	return &whisperResponse, nil
}