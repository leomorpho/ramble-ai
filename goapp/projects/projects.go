package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"MYAPP/ent"
	"MYAPP/ent/project"
	"MYAPP/ent/schema"
	"MYAPP/ent/settings"
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
	
	// Create a URL-safe path for the video
	encodedPath := url.QueryEscape(filePath)
	videoURL := fmt.Sprintf("wails://getVideoFile?path=%s", encodedPath)
	
	return videoURL, nil
}

// UpdateVideoClipHighlights updates the highlights for a video clip
func (s *ProjectService) UpdateVideoClipHighlights(clipID int, highlights []Highlight) error {
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
	
	_, err := s.client.VideoClip.
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

// UpdateProjectHighlightOrder updates the highlight order for a project
func (s *ProjectService) UpdateProjectHighlightOrder(projectID int, highlightOrder []string) error {
	// Convert highlight order to JSON for storage
	highlightOrderJSON, err := json.Marshal(highlightOrder)
	if err != nil {
		return fmt.Errorf("failed to marshal highlight order: %w", err)
	}
	
	// Store in settings table with project-specific key
	settingKey := fmt.Sprintf("project_%d_highlight_order", projectID)
	
	// Check if setting exists
	existing, err := s.client.Settings.
		Query().
		Where(settings.Key(settingKey)).
		Only(s.ctx)
		
	if err != nil {
		// Setting doesn't exist, create it
		_, err = s.client.Settings.
			Create().
			SetKey(settingKey).
			SetValue(string(highlightOrderJSON)).
			Save(s.ctx)
	} else {
		// Setting exists, update it
		_, err = s.client.Settings.
			UpdateOne(existing).
			SetValue(string(highlightOrderJSON)).
			Save(s.ctx)
	}
	
	return err
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
	return fmt.Sprintf("wails://getThumbnail?path=%s", encodedPath)
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