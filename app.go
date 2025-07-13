package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"MYAPP/ent"
	"MYAPP/goapp/ai"
	"MYAPP/goapp/assetshandler"
	"MYAPP/goapp/chatbot"
	"MYAPP/goapp/exports"
	"MYAPP/goapp/highlights"
	"MYAPP/goapp/projects"
	"MYAPP/goapp/realtime"
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

	// Set context for real-time manager
	manager := realtime.GetManager()
	manager.SetContext(ctx)

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

// Type aliases for AI service responses
type TestOpenAIApiKeyResponse = ai.TestOpenAIApiKeyResponse
type TestOpenRouterApiKeyResponse = ai.TestOpenRouterApiKeyResponse

// TestOpenAIApiKey tests if the stored OpenAI API key is valid
func (a *App) TestOpenAIApiKey() (*TestOpenAIApiKeyResponse, error) {
	service := ai.NewApiKeyService(a.client, a.ctx)
	return service.TestOpenAIApiKey()
}

// TestOpenRouterApiKey tests if the stored OpenRouter API key is valid
func (a *App) TestOpenRouterApiKey() (*TestOpenRouterApiKeyResponse, error) {
	service := ai.NewApiKeyService(a.client, a.ctx)
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

// TranscribeVideoClip transcribes audio from a video clip using the Projects service
func (a *App) TranscribeVideoClip(clipID int) (*projects.TranscriptionResponse, error) {
	projectService := projects.NewProjectService(a.client, a.ctx)
	return projectService.TranscribeVideoClip(clipID)
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

// ReorderHighlightsWithAI uses OpenRouter API to intelligently reorder highlights
func (a *App) ReorderHighlightsWithAI(projectID int, customPrompt string) ([]interface{}, error) {
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

// DeleteSuggestedHighlight removes a specific suggested highlight from a video
func (a *App) DeleteSuggestedHighlight(videoID int, suggestionID string) error {
	service := highlights.NewHighlightService(a.client, a.ctx)
	return service.DeleteSuggestedHighlight(videoID, suggestionID)
}

// ImproveHighlightSilencesWithAI uses AI to suggest improved timings for highlights with natural silence buffers
func (a *App) ImproveHighlightSilencesWithAI(projectID int) ([]ProjectHighlight, error) {
	service := highlights.NewAIService(a.client, a.ctx)
	return service.ImproveHighlightSilencesWithAI(projectID, a.GetOpenRouterApiKey)
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
