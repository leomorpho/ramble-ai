package projects

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"MYAPP/ent"
	"MYAPP/ent/project"
	"MYAPP/ent/schema"
	"MYAPP/ent/videoclip"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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

// ProjectResponse represents a project response for the frontend
type ProjectResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	ActiveTab   string `json:"activeTab"`
}

// VideoClipResponse represents a video clip response for the frontend
type VideoClipResponse struct {
	ID                    int         `json:"id"`
	Name                  string      `json:"name"`
	Description           string      `json:"description"`
	FilePath              string      `json:"filePath"`
	FileName              string      `json:"fileName"`
	FileSize              int64       `json:"fileSize"`
	Duration              float64     `json:"duration"`
	Format                string      `json:"format"`
	Width                 int         `json:"width"`
	Height                int         `json:"height"`
	ProjectID             int         `json:"projectId"`
	CreatedAt             string      `json:"createdAt"`
	UpdatedAt             string      `json:"updatedAt"`
	Exists                bool        `json:"exists"`
	ThumbnailURL          string      `json:"thumbnailUrl"`
	Transcription         string      `json:"transcription"`
	TranscriptionWords    []Word      `json:"transcriptionWords"`
	TranscriptionLanguage string      `json:"transcriptionLanguage"`
	TranscriptionDuration float64     `json:"transcriptionDuration"`
	Highlights            []Highlight `json:"highlights"`
}

// LocalVideoFile represents a local video file
type LocalVideoFile struct {
	Name     string `json:"name"`
	FilePath string `json:"filePath"`
	FileName string `json:"fileName"`
	FileSize int64  `json:"fileSize"`
	Format   string `json:"format"`
	Exists   bool   `json:"exists"`
}

// ProjectService provides project and video clip management functionality
type ProjectService struct {
	client *ent.Client
	ctx    context.Context
}

// NewProjectService creates a new project service
func NewProjectService(client *ent.Client, ctx context.Context) *ProjectService {
	return &ProjectService{
		client: client,
		ctx:    ctx,
	}
}

// CreateProject creates a new project
func (s *ProjectService) CreateProject(name, description string) (*ProjectResponse, error) {
	// Create project directory path
	projectPath := filepath.Join("projects", name)
	
	// Create the project in the database
	proj, err := s.client.Project.
		Create().
		SetName(name).
		SetDescription(description).
		SetPath(projectPath).
		Save(s.ctx)
		
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	
	return &ProjectResponse{
		ID:          proj.ID,
		Name:        proj.Name,
		Description: proj.Description,
		Path:        proj.Path,
		CreatedAt:   proj.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   proj.UpdatedAt.Format("2006-01-02 15:04:05"),
		ActiveTab:   proj.ActiveTab,
	}, nil
}

// GetProjects returns all projects
func (s *ProjectService) GetProjects() ([]*ProjectResponse, error) {
	projects, err := s.client.Project.Query().All(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	
	var responses []*ProjectResponse
	for _, proj := range projects {
		responses = append(responses, &ProjectResponse{
			ID:          proj.ID,
			Name:        proj.Name,
			Description: proj.Description,
			Path:        proj.Path,
			CreatedAt:   proj.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   proj.UpdatedAt.Format("2006-01-02 15:04:05"),
			ActiveTab:   proj.ActiveTab,
		})
	}
	
	return responses, nil
}

// GetProjectByID returns a project by ID
func (s *ProjectService) GetProjectByID(id int) (*ProjectResponse, error) {
	proj, err := s.client.Project.
		Query().
		Where(project.ID(id)).
		Only(s.ctx)
		
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	return &ProjectResponse{
		ID:          proj.ID,
		Name:        proj.Name,
		Description: proj.Description,
		Path:        proj.Path,
		CreatedAt:   proj.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   proj.UpdatedAt.Format("2006-01-02 15:04:05"),
		ActiveTab:   proj.ActiveTab,
	}, nil
}

// UpdateProject updates a project
func (s *ProjectService) UpdateProject(id int, name, description string) (*ProjectResponse, error) {
	// Update project path based on new name
	projectPath := filepath.Join("projects", name)
	
	proj, err := s.client.Project.
		UpdateOneID(id).
		SetName(name).
		SetDescription(description).
		SetPath(projectPath).
		Save(s.ctx)
		
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}
	
	return &ProjectResponse{
		ID:          proj.ID,
		Name:        proj.Name,
		Description: proj.Description,
		Path:        proj.Path,
		CreatedAt:   proj.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   proj.UpdatedAt.Format("2006-01-02 15:04:05"),
		ActiveTab:   proj.ActiveTab,
	}, nil
}

// DeleteProject deletes a project
func (s *ProjectService) DeleteProject(id int) error {
	err := s.client.Project.DeleteOneID(id).Exec(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

// CreateVideoClip creates a new video clip
func (s *ProjectService) CreateVideoClip(projectID int, filePath string) (*VideoClipResponse, error) {
	// Validate that it's a video file
	if !s.isVideoFile(filePath) {
		return nil, fmt.Errorf("file is not a supported video format")
	}
	
	// Get file information
	fileSize, format, exists := s.getFileInfo(filePath)
	if !exists {
		return nil, fmt.Errorf("file does not exist")
	}
	
	// Extract filename without extension for default name
	fileName := filepath.Base(filePath)
	name := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	
	// Create the video clip in the database
	clip, err := s.client.VideoClip.
		Create().
		SetName(name).
		SetFilePath(filePath).
		SetFileSize(fileSize).
		SetFormat(format).
		SetProjectID(projectID).
		Save(s.ctx)
		
	if err != nil {
		return nil, fmt.Errorf("failed to create video clip: %w", err)
	}
	
	return &VideoClipResponse{
		ID:                    clip.ID,
		Name:                  clip.Name,
		Description:           clip.Description,
		FilePath:              clip.FilePath,
		FileName:              fileName,
		FileSize:              fileSize,
		Duration:              clip.Duration,
		Format:                format,
		Width:                 clip.Width,
		Height:                clip.Height,
		ProjectID:             0, // Will need to be loaded separately
		CreatedAt:             clip.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:             clip.UpdatedAt.Format("2006-01-02 15:04:05"),
		Exists:                exists,
		ThumbnailURL:          s.getThumbnailURL(filePath),
		Transcription:         clip.Transcription,
		TranscriptionWords:    s.schemaWordsToWords(clip.TranscriptionWords),
		TranscriptionLanguage: clip.TranscriptionLanguage,
		TranscriptionDuration: clip.TranscriptionDuration,
		Highlights:            s.schemaHighlightsToHighlights(clip.Highlights),
	}, nil
}

// GetVideoClipsByProject returns all video clips for a project
func (s *ProjectService) GetVideoClipsByProject(projectID int) ([]*VideoClipResponse, error) {
	clips, err := s.client.VideoClip.
		Query().
		Where(videoclip.HasProjectWith(project.ID(projectID))).
		All(s.ctx)
		
	if err != nil {
		return nil, fmt.Errorf("failed to get video clips: %w", err)
	}
	
	var responses []*VideoClipResponse
	for _, clip := range clips {
		fileSize, format, exists := s.getFileInfo(clip.FilePath)
		fileName := filepath.Base(clip.FilePath)
		
		responses = append(responses, &VideoClipResponse{
			ID:                    clip.ID,
			Name:                  clip.Name,
			Description:           clip.Description,
			FilePath:              clip.FilePath,
			FileName:              fileName,
			FileSize:              fileSize,
			Duration:              clip.Duration,
			Format:                format,
			Width:                 clip.Width,
			Height:                clip.Height,
			ProjectID:             projectID,
			CreatedAt:             clip.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:             clip.UpdatedAt.Format("2006-01-02 15:04:05"),
			Exists:                exists,
			ThumbnailURL:          s.getThumbnailURL(clip.FilePath),
			Transcription:         clip.Transcription,
			TranscriptionWords:    s.schemaWordsToWords(clip.TranscriptionWords),
			TranscriptionLanguage: clip.TranscriptionLanguage,
			TranscriptionDuration: clip.TranscriptionDuration,
			Highlights:            s.schemaHighlightsToHighlights(clip.Highlights),
		})
	}
	
	return responses, nil
}

// UpdateVideoClip updates a video clip
func (s *ProjectService) UpdateVideoClip(id int, name, description string) (*VideoClipResponse, error) {
	clip, err := s.client.VideoClip.
		UpdateOneID(id).
		SetName(name).
		SetDescription(description).
		Save(s.ctx)
		
	if err != nil {
		return nil, fmt.Errorf("failed to update video clip: %w", err)
	}
	
	fileSize, format, exists := s.getFileInfo(clip.FilePath)
	fileName := filepath.Base(clip.FilePath)
	
	return &VideoClipResponse{
		ID:                    clip.ID,
		Name:                  clip.Name,
		Description:           clip.Description,
		FilePath:              clip.FilePath,
		FileName:              fileName,
		FileSize:              fileSize,
		Duration:              clip.Duration,
		Format:                format,
		Width:                 clip.Width,
		Height:                clip.Height,
		ProjectID:             0, // Will need to be loaded separately
		CreatedAt:             clip.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:             clip.UpdatedAt.Format("2006-01-02 15:04:05"),
		Exists:                exists,
		ThumbnailURL:          s.getThumbnailURL(clip.FilePath),
		Transcription:         clip.Transcription,
		TranscriptionWords:    s.schemaWordsToWords(clip.TranscriptionWords),
		TranscriptionLanguage: clip.TranscriptionLanguage,
		TranscriptionDuration: clip.TranscriptionDuration,
		Highlights:            s.schemaHighlightsToHighlights(clip.Highlights),
	}, nil
}

// DeleteVideoClip deletes a video clip
func (s *ProjectService) DeleteVideoClip(id int) error {
	err := s.client.VideoClip.DeleteOneID(id).Exec(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to delete video clip: %w", err)
	}
	return nil
}

// SelectVideoFiles opens a file dialog to select video files
func (s *ProjectService) SelectVideoFiles(ctx context.Context) ([]*LocalVideoFile, error) {
	files, err := runtime.OpenMultipleFilesDialog(ctx, runtime.OpenDialogOptions{
		Title: "Select Video Files",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Video Files",
				Pattern:     "*.mp4;*.avi;*.mov;*.mkv;*.wmv;*.flv;*.webm;*.m4v;*.3gp;*.ogv",
			},
		},
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to open file dialog: %w", err)
	}
	
	var videoFiles []*LocalVideoFile
	for _, filePath := range files {
		if s.isVideoFile(filePath) {
			fileSize, format, exists := s.getFileInfo(filePath)
			fileName := filepath.Base(filePath)
			name := strings.TrimSuffix(fileName, filepath.Ext(fileName))
			
			videoFiles = append(videoFiles, &LocalVideoFile{
				Name:     name,
				FilePath: filePath,
				FileName: fileName,
				FileSize: fileSize,
				Format:   format,
				Exists:   exists,
			})
		}
	}
	
	return videoFiles, nil
}

// GetVideoFileInfo returns information about a video file
func (s *ProjectService) GetVideoFileInfo(filePath string) (*LocalVideoFile, error) {
	if !s.isVideoFile(filePath) {
		return nil, fmt.Errorf("file is not a supported video format")
	}
	
	fileSize, format, exists := s.getFileInfo(filePath)
	fileName := filepath.Base(filePath)
	name := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	
	return &LocalVideoFile{
		Name:     name,
		FilePath: filePath,
		FileName: fileName,
		FileSize: fileSize,
		Format:   format,
		Exists:   exists,
	}, nil
}

// GetVideoURL returns a URL for accessing a video file
func (s *ProjectService) GetVideoURL(filePath string) (string, error) {
	if !s.isVideoFile(filePath) {
		return "", fmt.Errorf("file is not a supported video format")
	}
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("video file does not exist")
	}
	
	// Create a URL-safe path for the video to work with the asset middleware
	encodedPath := url.QueryEscape(filePath)
	videoURL := fmt.Sprintf("/api/video/%s", encodedPath)
	
	return videoURL, nil
}

// UpdateVideoClipHighlights updates the highlights for a video clip
func (s *ProjectService) UpdateVideoClipHighlights(clipID int, highlights []Highlight) error {
	// Save current state to history before making changes
	err := s.saveHighlightsState(clipID)
	if err != nil {
		// Log error but don't fail the update
		fmt.Printf("Warning: failed to save highlights state to history: %v\n", err)
	}

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
	
	_, err = s.client.VideoClip.
		UpdateOneID(clipID).
		SetHighlights(schemaHighlights).
		Save(s.ctx)
		
	return err
}

// UpdateVideoClipSuggestedHighlights updates the suggested highlights for a video clip
func (s *ProjectService) UpdateVideoClipSuggestedHighlights(clipID int, suggestedHighlights []Highlight) error {
	// Convert Highlights to schema.Highlights for database storage
	var schemaHighlights []schema.Highlight
	for _, h := range suggestedHighlights {
		schemaHighlights = append(schemaHighlights, schema.Highlight{
			ID:    h.ID,
			Start: h.Start,
			End:   h.End,
			Color: h.Color,
		})
	}
	
	_, err := s.client.VideoClip.
		UpdateOneID(clipID).
		SetSuggestedHighlights(schemaHighlights).
		Save(s.ctx)
		
	return err
}

// UpdateProjectActiveTab updates the active tab for a project
func (s *ProjectService) UpdateProjectActiveTab(projectID int, activeTab string) error {
	_, err := s.client.Project.
		UpdateOneID(projectID).
		SetActiveTab(activeTab).
		Save(s.ctx)
		
	if err != nil {
		return fmt.Errorf("failed to update project active tab: %w", err)
	}
	
	return nil
}

// UpdateProjectHighlightOrder updates the highlight order for a project
func (s *ProjectService) UpdateProjectHighlightOrder(projectID int, highlightOrder []string) error {
	// Save current state to history before making changes
	err := s.saveOrderState(projectID)
	if err != nil {
		// Log error but don't fail the update
		fmt.Printf("Warning: failed to save order state to history: %v\n", err)
	}

	// Update the highlight order in the project schema
	_, err = s.client.Project.
		UpdateOneID(projectID).
		SetHighlightOrder(highlightOrder).
		Save(s.ctx)
	
	if err != nil {
		return fmt.Errorf("failed to update project highlight order: %w", err)
	}
	
	return nil
}

// Helper functions

// isVideoFile checks if a file is a supported video format
func (s *ProjectService) isVideoFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	videoExts := []string{".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm", ".m4v", ".3gp", ".ogv"}
	
	for _, videoExt := range videoExts {
		if ext == videoExt {
			return true
		}
	}
	return false
}

// getFileInfo returns file size, format, and existence status
func (s *ProjectService) getFileInfo(filePath string) (int64, string, bool) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, "", false
	}
	
	ext := strings.ToLower(filepath.Ext(filePath))
	format := strings.TrimPrefix(ext, ".")
	
	return info.Size(), format, true
}

// getThumbnailURL returns a thumbnail URL for a video file
func (s *ProjectService) getThumbnailURL(filePath string) string {
	if !s.isVideoFile(filePath) {
		return ""
	}
	
	encodedPath := url.QueryEscape(filePath)
	return fmt.Sprintf("/api/thumbnail/%s", encodedPath)
}

// schemaWordsToWords converts schema.Word slice to Word slice
func (s *ProjectService) schemaWordsToWords(schemaWords []schema.Word) []Word {
	var words []Word
	for _, sw := range schemaWords {
		words = append(words, Word{
			Word:  sw.Word,
			Start: sw.Start,
			End:   sw.End,
		})
	}
	return words
}

// schemaHighlightsToHighlights converts schema.Highlight slice to Highlight slice
func (s *ProjectService) schemaHighlightsToHighlights(schemaHighlights []schema.Highlight) []Highlight {
	var highlights []Highlight
	for _, sh := range schemaHighlights {
		highlights = append(highlights, Highlight{
			ID:    sh.ID,
			Start: sh.Start,
			End:   sh.End,
			Color: sh.Color,
		})
	}
	return highlights
}

// History Management Functions

// saveOrderState saves the current highlight order to history before making changes
func (s *ProjectService) saveOrderState(projectID int) error {
	// Get current project with history
	project, err := s.client.Project.
		Query().
		Where(project.ID(projectID)).
		Only(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Get current order from project schema
	currentOrder := project.HighlightOrder
	if currentOrder == nil {
		currentOrder = []string{}
	}

	// Get current history
	history := project.OrderHistory
	if history == nil {
		history = [][]string{}
	}

	// Add current order to history (FIFO, max 20)
	history = append(history, currentOrder)
	if len(history) > 20 {
		history = history[1:] // Remove oldest entry
	}

	// Update project with new history and reset index to -1 (new change, can't redo)
	_, err = s.client.Project.
		UpdateOneID(projectID).
		SetOrderHistory(history).
		SetOrderHistoryIndex(-1).
		Save(s.ctx)

	return err
}

// saveHighlightsState saves the current highlights to history before making changes
func (s *ProjectService) saveHighlightsState(clipID int) error {
	// Get current video clip with history
	clip, err := s.client.VideoClip.
		Query().
		Where(videoclip.ID(clipID)).
		Only(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to get video clip: %w", err)
	}

	// Get current highlights
	currentHighlights := clip.Highlights

	// Get current history
	history := clip.HighlightsHistory
	if history == nil {
		history = [][]schema.Highlight{}
	}

	// Add current highlights to history (FIFO, max 20)
	history = append(history, currentHighlights)
	if len(history) > 20 {
		history = history[1:] // Remove oldest entry
	}

	// Update clip with new history and reset index to -1 (new change, can't redo)
	_, err = s.client.VideoClip.
		UpdateOneID(clipID).
		SetHighlightsHistory(history).
		SetHighlightsHistoryIndex(-1).
		Save(s.ctx)

	return err
}

// UndoOrderChange moves backward in order history
func (s *ProjectService) UndoOrderChange(projectID int) ([]string, error) {
	// Get current project with history
	project, err := s.client.Project.
		Query().
		Where(project.ID(projectID)).
		Only(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	history := project.OrderHistory
	if history == nil || len(history) == 0 {
		return nil, fmt.Errorf("no history available")
	}

	currentIndex := project.OrderHistoryIndex
	
	// Calculate new index (move backward)
	var newIndex int
	if currentIndex == -1 {
		// We're at current state, move to last history entry
		newIndex = len(history) - 1
	} else if currentIndex > 0 {
		// Move backward in history
		newIndex = currentIndex - 1
	} else {
		// Already at oldest entry
		return nil, fmt.Errorf("cannot undo further")
	}

	// Get order from history
	orderFromHistory := history[newIndex]

	// Update project index and apply the order
	_, err = s.client.Project.
		UpdateOneID(projectID).
		SetOrderHistoryIndex(newIndex).
		Save(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update history index: %w", err)
	}

	// Apply the order to settings
	err = s.UpdateProjectHighlightOrderWithoutHistory(projectID, orderFromHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to apply historical order: %w", err)
	}

	return orderFromHistory, nil
}

// RedoOrderChange moves forward in order history
func (s *ProjectService) RedoOrderChange(projectID int) ([]string, error) {
	// Get current project with history
	project, err := s.client.Project.
		Query().
		Where(project.ID(projectID)).
		Only(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	history := project.OrderHistory
	if history == nil || len(history) == 0 {
		return nil, fmt.Errorf("no history available")
	}

	currentIndex := project.OrderHistoryIndex
	
	// Calculate new index (move forward)
	if currentIndex == -1 || currentIndex >= len(history)-1 {
		// Already at newest entry or current state
		return nil, fmt.Errorf("cannot redo further")
	}

	newIndex := currentIndex + 1

	// Get order from history
	orderFromHistory := history[newIndex]

	// Update project index and apply the order
	_, err = s.client.Project.
		UpdateOneID(projectID).
		SetOrderHistoryIndex(newIndex).
		Save(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update history index: %w", err)
	}

	// Apply the order to settings
	err = s.UpdateProjectHighlightOrderWithoutHistory(projectID, orderFromHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to apply historical order: %w", err)
	}

	return orderFromHistory, nil
}

// GetOrderHistoryStatus returns whether undo/redo is available
func (s *ProjectService) GetOrderHistoryStatus(projectID int) (bool, bool, error) {
	// Get current project with history
	project, err := s.client.Project.
		Query().
		Where(project.ID(projectID)).
		Only(s.ctx)
	if err != nil {
		return false, false, fmt.Errorf("failed to get project: %w", err)
	}

	history := project.OrderHistory
	if history == nil || len(history) == 0 {
		return false, false, nil
	}

	currentIndex := project.OrderHistoryIndex
	
	// Can undo if we have history and we're not at the oldest entry
	canUndo := len(history) > 0 && (currentIndex == -1 || currentIndex > 0)
	
	// Can redo if we have history and we're not at the newest entry
	canRedo := len(history) > 0 && currentIndex != -1 && currentIndex < len(history)-1

	return canUndo, canRedo, nil
}

// UndoHighlightsChange moves backward in highlights history
func (s *ProjectService) UndoHighlightsChange(clipID int) ([]Highlight, error) {
	// Get current video clip with history
	clip, err := s.client.VideoClip.
		Query().
		Where(videoclip.ID(clipID)).
		Only(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get video clip: %w", err)
	}

	history := clip.HighlightsHistory
	if history == nil || len(history) == 0 {
		return nil, fmt.Errorf("no history available")
	}

	currentIndex := clip.HighlightsHistoryIndex
	
	// Calculate new index (move backward)
	var newIndex int
	if currentIndex == -1 {
		// We're at current state, move to last history entry
		newIndex = len(history) - 1
	} else if currentIndex > 0 {
		// Move backward in history
		newIndex = currentIndex - 1
	} else {
		// Already at oldest entry
		return nil, fmt.Errorf("cannot undo further")
	}

	// Get highlights from history
	highlightsFromHistory := history[newIndex]

	// Update clip index and apply the highlights
	_, err = s.client.VideoClip.
		UpdateOneID(clipID).
		SetHighlightsHistoryIndex(newIndex).
		SetHighlights(highlightsFromHistory).
		Save(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to apply historical highlights: %w", err)
	}

	// Convert to return format
	return s.schemaHighlightsToHighlights(highlightsFromHistory), nil
}

// RedoHighlightsChange moves forward in highlights history
func (s *ProjectService) RedoHighlightsChange(clipID int) ([]Highlight, error) {
	// Get current video clip with history
	clip, err := s.client.VideoClip.
		Query().
		Where(videoclip.ID(clipID)).
		Only(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get video clip: %w", err)
	}

	history := clip.HighlightsHistory
	if history == nil || len(history) == 0 {
		return nil, fmt.Errorf("no history available")
	}

	currentIndex := clip.HighlightsHistoryIndex
	
	// Calculate new index (move forward)
	if currentIndex == -1 || currentIndex >= len(history)-1 {
		// Already at newest entry or current state
		return nil, fmt.Errorf("cannot redo further")
	}

	newIndex := currentIndex + 1

	// Get highlights from history
	highlightsFromHistory := history[newIndex]

	// Update clip index and apply the highlights
	_, err = s.client.VideoClip.
		UpdateOneID(clipID).
		SetHighlightsHistoryIndex(newIndex).
		SetHighlights(highlightsFromHistory).
		Save(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to apply historical highlights: %w", err)
	}

	// Convert to return format
	return s.schemaHighlightsToHighlights(highlightsFromHistory), nil
}

// GetHighlightsHistoryStatus returns whether undo/redo is available for highlights
func (s *ProjectService) GetHighlightsHistoryStatus(clipID int) (bool, bool, error) {
	// Get current video clip with history
	clip, err := s.client.VideoClip.
		Query().
		Where(videoclip.ID(clipID)).
		Only(s.ctx)
	if err != nil {
		return false, false, fmt.Errorf("failed to get video clip: %w", err)
	}

	history := clip.HighlightsHistory
	if history == nil || len(history) == 0 {
		return false, false, nil
	}

	currentIndex := clip.HighlightsHistoryIndex
	
	// Can undo if we have history and we're not at the oldest entry
	canUndo := len(history) > 0 && (currentIndex == -1 || currentIndex > 0)
	
	// Can redo if we have history and we're not at the newest entry
	canRedo := len(history) > 0 && currentIndex != -1 && currentIndex < len(history)-1

	return canUndo, canRedo, nil
}

// UpdateProjectHighlightOrderWithoutHistory updates order without saving to history (used for undo/redo)
func (s *ProjectService) UpdateProjectHighlightOrderWithoutHistory(projectID int, highlightOrder []string) error {
	// Update the highlight order in the project schema directly (no history save)
	_, err := s.client.Project.
		UpdateOneID(projectID).
		SetHighlightOrder(highlightOrder).
		Save(s.ctx)
	
	if err != nil {
		return fmt.Errorf("failed to update project highlight order: %w", err)
	}
	
	return nil
}