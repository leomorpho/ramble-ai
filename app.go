package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ramble-ai/ent"
	"ramble-ai/goapp"
	"ramble-ai/goapp/assetshandler"
	"ramble-ai/goapp/ai"
	"ramble-ai/goapp/chatbot"
	"ramble-ai/goapp/config"
	"ramble-ai/goapp/exports"
	"ramble-ai/goapp/highlights"
	"ramble-ai/goapp/projects"
	"ramble-ai/goapp/realtime"
	"ramble-ai/goapp/settings"
	"ramble-ai/goapp/version"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	client *ent.Client
}

// getUserDataDir returns the user data directory for the application
func getUserDataDir() (string, error) {
	// Check if we're in development mode by looking for go.mod file
	if _, err := os.Stat("go.mod"); err == nil {
		// In development mode, use current directory
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}
		return cwd, nil
	}

	// In production mode, use user config directory
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}

	appDataDir := filepath.Join(userConfigDir, "RambleAI")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(appDataDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create app data directory: %w", err)
	}

	return appDataDir, nil
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Get user data directory
	userDataDir, err := getUserDataDir()
	if err != nil {
		log.Fatalf("failed to get user data directory: %v", err)
	}

	// Initialize database in user data directory
	dbPath := filepath.Join(userDataDir, "database.db")
	db, err := sql.Open("sqlite3", dbPath+"?_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}

	log.Printf("Database initialized at: %s", dbPath)

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

	// Set context for real-time manager
	manager := realtime.GetManager()
	manager.SetContext(ctx)

	// Run database migrations
	if err := a.client.Schema.Create(ctx); err != nil {
		log.Printf("failed creating schema resources: %v", err)
	}

	// Seed development API key if in development mode (unconditionally)
	if err := a.SeedDevAPIKeyOnStartup(); err != nil {
		log.Printf("Warning: Failed to seed development API key: %v", err)
	}

	log.Println("Database initialized and migrations applied")

	// Create event emitter function
	emitEvent := func(eventName string, data ...interface{}) {
		// For single string data, emit it directly instead of as an array
		if len(data) == 1 {
			runtime.EventsEmit(ctx, eventName, data[0])
		} else {
			runtime.EventsEmit(ctx, eventName, data...)
		}
	}
	
	// Initialize bundled FFmpeg immediately - no delay needed since it's synchronous
	if err := goapp.EnsureFFmpeg(ctx, nil, emitEvent); err != nil {
		log.Printf("Failed to ensure FFmpeg availability: %v", err)
	} else {
		log.Printf("FFmpeg initialized successfully")
	}

	// Recover any incomplete export jobs
	if err := a.RecoverActiveExportJobs(); err != nil {
		log.Printf("Failed to recover active export jobs: %v", err)
	}
}

// shutdown is called when the app shuts down
func (a *App) shutdown(ctx context.Context) {
	// Shutdown real-time manager
	manager := realtime.GetManager()
	manager.Shutdown()

	// Close the database connection
	if err := a.client.Close(); err != nil {
		log.Printf("failed to close database connection: %v", err)
	}
}

// createAssetMiddleware creates middleware for serving video files via AssetServer
func (a *App) createAssetMiddleware() assetserver.Middleware {
	assetHandler := assetshandler.NewAssetHandler()
	originalMiddleware := assetHandler.CreateAssetMiddleware()

	// Wrap the original middleware to handle SSE endpoints
	return func(next http.Handler) http.Handler {
		// Apply the original middleware first
		wrappedHandler := originalMiddleware(next)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if this is an SSE request
			if strings.HasPrefix(r.URL.Path, "/api/sse/highlights") {
				manager := realtime.GetManager()
				manager.HandleSSEConnection(w, r)
				return
			}

			// For all other requests, use the original wrapped handler
			wrappedHandler.ServeHTTP(w, r)
		})
	}
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

// UpdateProjectActiveTab updates the active tab for a project
func (a *App) UpdateProjectActiveTab(projectID int, activeTab string) error {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.UpdateProjectActiveTab(projectID, activeTab)
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

// SaveUseRemoteAIBackend saves the remote AI backend toggle setting
func (a *App) SaveUseRemoteAIBackend(useRemote bool) error {
	value := "false"
	if useRemote {
		value = "true"
	}
	return a.SaveSetting("use_remote_ai_backend", value)
}

// GetUseRemoteAIBackend always returns true (remote backend only)
func (a *App) GetUseRemoteAIBackend() (bool, error) {
	// Always use remote backend - keeping function for compatibility
	return true, nil
}

// SaveRemoteAIBackendURL saves the remote AI backend URL setting
func (a *App) SaveRemoteAIBackendURL(url string) error {
	return a.SaveSetting("remote_ai_backend_url", url)
}

// GetRemoteAIBackendURL retrieves the remote AI backend URL setting
func (a *App) GetRemoteAIBackendURL() (string, error) {
	return a.GetSetting("remote_ai_backend_url")
}

// IsDevMode checks if the application is running in development mode
func (a *App) IsDevMode() bool {
	// Check if we're in development mode by looking for go.mod file
	_, err := os.Stat("go.mod")
	return err == nil
}

// IsRemoteBackendOverriddenByEnv checks if the remote backend setting is overridden by environment variables
func (a *App) IsRemoteBackendOverriddenByEnv() bool {
	// In production, always return true since it's compiled in
	if config.IsProduction() {
		return true
	}
	// In development, check for env override
	return os.Getenv("USE_REMOTE_AI_BACKEND") != ""
}

// GetRambleFrontendURL returns the Ramble AI frontend URL for API key acquisition
func (a *App) GetRambleFrontendURL() string {
	// Use build-time configuration with env override in development
	return config.GetFrontendURL()
}

// GetDevAPIKey returns the development API key if in development mode
func (a *App) GetDevAPIKey() string {
	// Only return the dev key if we're in development mode
	if !a.IsDevMode() {
		return ""
	}
	// This matches the key defined in PocketBase seed.go
	return "ra-dev-12345678901234567890123456789012"
}


// SeedDevAPIKey seeds the development API key in the settings if in dev mode and remote backend is enabled
func (a *App) SeedDevAPIKey() error {
	// Only seed if in development mode
	if !a.IsDevMode() {
		return nil
	}

	// Check if remote backend is enabled
	useRemote, err := a.GetUseRemoteAIBackend()
	if err != nil {
		return err
	}

	if !useRemote {
		return nil // No need to seed if not using remote backend
	}

	// Get existing API key
	existingKey, err := a.GetRambleAIApiKey()
	if err != nil {
		return err
	}

	// Only seed if no key is set or if the current key is empty
	if existingKey == "" {
		devKey := a.GetDevAPIKey()
		if devKey != "" {
			return a.SaveRambleAIApiKey(devKey)
		}
	}

	return nil
}

// SeedDevAPIKeyOnStartup seeds the development API key unconditionally if in dev mode
func (a *App) SeedDevAPIKeyOnStartup() error {
	// Only seed if in development mode
	if !a.IsDevMode() {
		return nil
	}

	// Get existing API key
	existingKey, err := a.GetRambleAIApiKey()
	if err != nil {
		return err
	}

	// Always seed the dev key if we're in development and no key exists
	if existingKey == "" {
		devKey := a.GetDevAPIKey()
		if devKey != "" {
			log.Printf("🌱 Seeding development API key: %s", devKey)
			return a.SaveRambleAIApiKey(devKey)
		}
	} else {
		log.Printf("✅ Development API key already exists: %s", existingKey[:16]+"...")
	}

	return nil
}

// SaveRambleAIApiKey saves the Ramble AI API key securely
func (a *App) SaveRambleAIApiKey(apiKey string) error {
	return a.SaveSetting("ramble_ai_api_key", apiKey)
}

// GetRambleAIApiKey retrieves the Ramble AI API key
func (a *App) GetRambleAIApiKey() (string, error) {
	return a.GetSetting("ramble_ai_api_key")
}

// GetBackendURL returns the backend URL configured for the current environment
func (a *App) GetBackendURL() string {
	return config.GetRemoteBackendURL()
}

// DeleteRambleAIApiKey removes the Ramble AI API key
func (a *App) DeleteRambleAIApiKey() error {
	return a.DeleteSetting("ramble_ai_api_key")
}

// Word represents a single word with timing information
type Word struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// Highlight represents a highlighted text region with timestamps
type Highlight struct {
	ID      string  `json:"id"`
	Start   float64 `json:"start"`
	End     float64 `json:"end"`
	ColorID int     `json:"colorId"`
}

// TranscribeVideoClip transcribes audio from a video clip using the AI factory
func (a *App) TranscribeVideoClip(clipID int) (*projects.TranscriptionResponse, error) {
	// First, get the video clip to check if it exists and get the file path
	clip, err := a.client.VideoClip.Get(a.ctx, clipID)
	if err != nil {
		return &projects.TranscriptionResponse{
			Success: false,
			Message: "Video clip not found",
		}, nil
	}

	// Check if file exists
	if _, err := os.Stat(clip.FilePath); os.IsNotExist(err) {
		return &projects.TranscriptionResponse{
			Success: false,
			Message: "Video file not found",
		}, nil
	}

	// Extract audio from video first to reduce file size
	audioPath, err := a.extractAudioFromVideo(clip.FilePath)
	if err != nil {
		return &projects.TranscriptionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to extract audio: %v", err),
		}, nil
	}
	defer os.Remove(audioPath) // Clean up temporary audio file

	// Create AI service using factory (which will choose local or remote based on settings)
	factory := ai.NewAIServiceFactory(a.client, a.ctx)
	aiService, err := factory.CreateService()
	if err != nil {
		return &projects.TranscriptionResponse{
			Success: false,
			Message: fmt.Sprintf("AI service configuration error: %v", err),
		}, nil
	}

	// Process the extracted audio file instead of the full video
	result, err := aiService.ProcessAudio(audioPath)
	if err != nil {
		return &projects.TranscriptionResponse{
			Success: false,
			Message: fmt.Sprintf("Transcription failed: %v", err),
		}, nil
	}

	// Convert AI result to projects format and save to database
	projectService := projects.NewProjectService(a.client, a.ctx)
	return projectService.SaveTranscriptionResult(clipID, result)
}

// extractAudioFromVideo extracts audio from a video file using FFmpeg
// Returns the path to the temporary audio file
func (a *App) extractAudioFromVideo(videoPath string) (string, error) {
	// Create temp directory for audio files
	tempDir := filepath.Join(os.TempDir(), "ramble_audio")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate unique audio filename
	audioFilename := fmt.Sprintf("audio_%d.mp3", time.Now().UnixNano())
	audioPath := filepath.Join(tempDir, audioFilename)

	log.Printf("[AUDIO EXTRACTION] Extracting audio from: %s to: %s", videoPath, audioPath)

	// Use ffmpeg-go library to extract audio with optimized settings for Whisper
	if err := goapp.ExtractAudio(videoPath, audioPath); err != nil {
		return "", fmt.Errorf("failed to extract audio: %w", err)
	}

	// Get file size for logging
	if stat, err := os.Stat(audioPath); err == nil {
		sizeMB := float64(stat.Size()) / (1024 * 1024)
		log.Printf("[AUDIO EXTRACTION] Audio extracted successfully: %s (%.2f MB)", audioPath, sizeMB)
	}

	return audioPath, nil
}

// BatchTranscribeUntranscribedClips transcribes all untranscribed video clips in a project
func (a *App) BatchTranscribeUntranscribedClips(projectID int) (*projects.BatchTranscribeResponse, error) {
	// Create AI service using factory (which will choose local or remote based on settings)
	factory := ai.NewAIServiceFactory(a.client, a.ctx)
	aiService, err := factory.CreateService()
	if err != nil {
		return &projects.BatchTranscribeResponse{
			Success: false,
			Message: fmt.Sprintf("AI service configuration error: %v", err),
		}, nil
	}

	// Use project service to handle the batch operation with AI service
	projectService := projects.NewProjectService(a.client, a.ctx)
	return projectService.BatchTranscribeWithAIService(projectID, aiService)
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
type AIActionOptions = highlights.AIActionOptions

// ProjectAISilenceResult represents AI silence improvement result for Wails compatibility
type ProjectAISilenceResult struct {
	Improvements []highlights.ProjectHighlight `json:"improvements"`
	CreatedAt    string                        `json:"createdAt"`
	Model        string                        `json:"model"`
}

// HistoryStatus represents the undo/redo status for Wails compatibility
type HistoryStatus struct {
	CanUndo bool `json:"canUndo"`
	CanRedo bool `json:"canRedo"`
}

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

// NewlineSection represents a newline section with an optional title
type NewlineSection struct {
	Type  string `json:"type"`
	Title string `json:"title"`
}

// SaveSectionTitle saves or updates the title for a newline section at a specific position
func (a *App) SaveSectionTitle(projectID int, position int, title string) error {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.SaveSectionTitle(projectID, position, title)
}

// GetSectionTitles retrieves all section titles from the project highlight order
func (a *App) GetSectionTitles(projectID int) (map[int]string, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.GetSectionTitles(projectID)
}

// UpdateProjectHighlightOrderWithTitles updates the highlight order with rich newline objects
func (a *App) UpdateProjectHighlightOrderWithTitles(projectID int, highlightOrder []interface{}) error {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.UpdateProjectHighlightOrderWithTitles(projectID, highlightOrder)
}

// GetProjectHighlightOrderWithTitles retrieves the highlight order with rich newline objects
func (a *App) GetProjectHighlightOrderWithTitles(projectID int) ([]interface{}, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.GetProjectHighlightOrderWithTitles(projectID)
}
// HideHighlight adds a highlight to the hidden highlights list
func (a *App) HideHighlight(projectID int, highlightID string) error {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.HideHighlight(projectID, highlightID)
}
// UnhideHighlight removes a highlight from the hidden highlights list
func (a *App) UnhideHighlight(projectID int, highlightID string) error {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.UnhideHighlight(projectID, highlightID)
}
// GetHiddenHighlights retrieves the list of hidden highlight IDs for a project
func (a *App) GetHiddenHighlights(projectID int) ([]string, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.GetHiddenHighlights(projectID)
}

// ReorderHighlightsWithAI uses OpenRouter API to intelligently reorder highlights
func (a *App) ReorderHighlightsWithAI(projectID int, customPrompt string) ([]interface{}, error) {
	// Use the specialized highlights AI service for reordering
	service := highlights.NewAIService(a.client, a.ctx)
	
	// Define callback functions required by the highlights AI service
	getAPIKey := func() (string, error) {
		return a.GetOpenRouterApiKey()
	}
	
	getProjectHighlights := func(id int) ([]highlights.ProjectHighlight, error) {
		return a.GetProjectHighlights(id)
	}
	
	// Delegate to the specialized reordering implementation
	return service.ReorderHighlightsWithAI(projectID, customPrompt, getAPIKey, getProjectHighlights)
}

// ReorderHighlightsWithAIOptions uses AI service (local or remote) to intelligently reorder highlights with specific options
func (a *App) ReorderHighlightsWithAIOptions(projectID int, customPrompt string, options highlights.AIActionOptions) ([]interface{}, error) {
	// Use the specialized highlights AI service for reordering with options
	service := highlights.NewAIService(a.client, a.ctx)
	
	// Define callback functions required by the highlights AI service
	getAPIKey := func() (string, error) {
		return a.GetOpenRouterApiKey()
	}
	
	getProjectHighlights := func(id int) ([]highlights.ProjectHighlight, error) {
		return a.GetProjectHighlights(id)
	}
	
	// Delegate to the specialized reordering implementation with options
	return service.ReorderHighlightsWithAIOptions(projectID, customPrompt, options, getAPIKey, getProjectHighlights)
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
func (a *App) ExportStitchedHighlights(projectID int, outputFolder string, paddingSeconds float64) (string, error) {
	service := exports.NewExportService(a.client, a.ctx)
	return service.ExportStitchedHighlights(projectID, outputFolder, paddingSeconds)
}

// ExportIndividualHighlights exports each highlight as a separate file
func (a *App) ExportIndividualHighlights(projectID int, outputFolder string, paddingSeconds float64) (string, error) {
	service := exports.NewExportService(a.client, a.ctx)
	return service.ExportIndividualHighlights(projectID, outputFolder, paddingSeconds)
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
	factory := ai.NewAIServiceFactory(a.client, a.ctx)
	aiService, err := factory.CreateService()
	if err != nil {
		return nil, fmt.Errorf("failed to create AI service: %w", err)
	}
	
	// For local service, delegate to highlights service
	if _, ok := aiService.(*ai.LocalAIService); ok {
		highlightService := highlights.NewAIService(a.client, a.ctx)
		return highlightService.SuggestHighlightsWithAI(projectID, videoID, customPrompt)
	}
	
	// For remote service, we need to build the request manually
	// This would need to be implemented in the remote service
	return nil, fmt.Errorf("highlight suggestions not yet supported with remote AI backend")
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

// DeleteSuggestedHighlight removes a specific suggested highlight from a video
func (a *App) DeleteSuggestedHighlight(videoID int, suggestionID string) error {
	service := highlights.NewHighlightService(a.client, a.ctx)
	return service.DeleteSuggestedHighlight(videoID, suggestionID)
}

// ImproveHighlightSilencesWithAI uses AI to suggest improved timings for highlights with natural silence buffers
func (a *App) ImproveHighlightSilencesWithAI(projectID int) ([]ProjectHighlight, error) {
	factory := ai.NewAIServiceFactory(a.client, a.ctx)
	aiService, err := factory.CreateService()
	if err != nil {
		return nil, fmt.Errorf("failed to create AI service: %w", err)
	}
	
	// For local service, delegate to highlights service
	if _, ok := aiService.(*ai.LocalAIService); ok {
		highlightService := highlights.NewAIService(a.client, a.ctx)
		return highlightService.ImproveHighlightSilencesWithAI(projectID, a.GetOpenRouterApiKey)
	}
	
	// For remote service, we need to build the request manually
	// This would need to be implemented in the remote service
	return nil, fmt.Errorf("silence improvements not yet supported with remote AI backend")
}

// GetProjectAISilenceResult retrieves cached AI silence improvements for a project
func (a *App) GetProjectAISilenceResult(projectID int) (*ProjectAISilenceResult, error) {
	service := highlights.NewAIService(a.client, a.ctx)
	improvements, createdAt, model, err := service.GetProjectAISilenceImprovements(projectID)
	if err != nil {
		return nil, err
	}

	// If no cached improvements, return nil
	if len(improvements) == 0 {
		return nil, nil
	}

	return &ProjectAISilenceResult{
		Improvements: improvements,
		CreatedAt:    createdAt.Format("2006-01-02T15:04:05Z07:00"),
		Model:        model,
	}, nil
}

// ClearAISilenceImprovements clears cached AI silence improvements for a project
func (a *App) ClearAISilenceImprovements(projectID int) error {
	return highlights.ClearAISilenceImprovementsCache(a.ctx, a.client, projectID)
}

// History Management - Project Order Undo/Redo

// UndoOrderChange reverts to previous state in project highlight order history
func (a *App) UndoOrderChange(projectID int) ([]string, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.UndoOrderChange(projectID)
}

// RedoOrderChange moves forward in project highlight order history
func (a *App) RedoOrderChange(projectID int) ([]string, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.RedoOrderChange(projectID)
}

// GetOrderHistoryStatus returns current undo/redo availability for project order
func (a *App) GetOrderHistoryStatus(projectID int) (*HistoryStatus, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	canUndo, canRedo, err := service.GetOrderHistoryStatus(projectID)
	if err != nil {
		return nil, err
	}
	return &HistoryStatus{
		CanUndo: canUndo,
		CanRedo: canRedo,
	}, nil
}

// History Management - Video Clip Highlights Undo/Redo

// UndoHighlightsChange reverts to previous state in video clip highlights history
func (a *App) UndoHighlightsChange(clipID int) ([]projects.Highlight, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.UndoHighlightsChange(clipID)
}

// RedoHighlightsChange moves forward in video clip highlights history
func (a *App) RedoHighlightsChange(clipID int) ([]projects.Highlight, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	return service.RedoHighlightsChange(clipID)
}

// GetHighlightsHistoryStatus returns current undo/redo availability for video clip highlights
func (a *App) GetHighlightsHistoryStatus(clipID int) (*HistoryStatus, error) {
	service := projects.NewProjectService(a.client, a.ctx)
	canUndo, canRedo, err := service.GetHighlightsHistoryStatus(clipID)
	if err != nil {
		return nil, err
	}
	return &HistoryStatus{
		CanUndo: canUndo,
		CanRedo: canRedo,
	}, nil
}

// Chatbot Methods

// SendChatMessage sends a message to the AI chatbot and returns the response
func (a *App) SendChatMessage(request chatbot.ChatRequest) (*chatbot.ChatResponse, error) {
	service := chatbot.NewChatbotService(a.client, a.ctx, a.UpdateProjectHighlightOrderWithTitles)
	return service.SendMessage(request, a.GetOpenRouterApiKey)
}

// GetChatHistory retrieves the chat history for a project and endpoint
func (a *App) GetChatHistory(projectID int, endpointID string) (*chatbot.ChatHistoryResponse, error) {
	service := chatbot.NewChatbotService(a.client, a.ctx, a.UpdateProjectHighlightOrderWithTitles)
	return service.GetChatHistory(projectID, endpointID)
}

// ClearChatHistory clears the chat history for a project and endpoint
func (a *App) ClearChatHistory(projectID int, endpointID string) error {
	service := chatbot.NewChatbotService(a.client, a.ctx, a.UpdateProjectHighlightOrderWithTitles)
	return service.ClearChatHistory(projectID, endpointID)
}

// SaveChatModelSelection saves the selected model for a chat session
func (a *App) SaveChatModelSelection(projectID int, endpointID string, model string) error {
	service := chatbot.NewChatbotService(a.client, a.ctx, a.UpdateProjectHighlightOrderWithTitles)
	return service.SaveModelSelection(projectID, endpointID, model)
}

// GetAppVersion returns the current application version information
func (a *App) GetAppVersion() version.Info {
	return version.Get()
}

// IsFFmpegReady checks if bundled FFmpeg is available
func (a *App) IsFFmpegReady() bool {
	// Check if bundled FFmpeg is available and working
	bundledPath := goapp.GetBundledFFmpegPath()
	return bundledPath != "" && goapp.TestFFmpegBinary(bundledPath)
}

