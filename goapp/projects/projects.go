package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ramble-ai/goapp"
	"ramble-ai/goapp/ai"

	"ramble-ai/ent"
	"ramble-ai/ent/chatsession"
	"ramble-ai/ent/chatmessage"
	"ramble-ai/ent/exportjob"
	"ramble-ai/ent/project"
	"ramble-ai/ent/schema"
	"ramble-ai/ent/settings"
	"ramble-ai/ent/videoclip"
	highlightsservice "ramble-ai/goapp/highlights"
	"ramble-ai/goapp/realtime"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Transcription state constants
const (
	TranscriptionStateIdle         = "idle"
	TranscriptionStateChecking     = "checking"
	TranscriptionStateTranscribing = "transcribing"
	TranscriptionStateCompleted    = "completed"
	TranscriptionStateError        = "error"
)

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

// NewlineSection represents a newline section with an optional title
type NewlineSection struct {
	Type  string `json:"type"`
	Title string `json:"title"`
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

// Segment represents a segment of transcribed audio
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

// TranscriptionResponse represents the response returned to the frontend
type TranscriptionResponse struct {
	Success       bool    `json:"success"`
	Message       string  `json:"message"`
	Transcription string  `json:"transcription,omitempty"`
	Words         []Word  `json:"words,omitempty"`
	Language      string  `json:"language,omitempty"`
	Duration      float64 `json:"duration,omitempty"`
}

// VideoClipResponse represents a video clip response for the frontend
type VideoClipResponse struct {
	ID                       int         `json:"id"`
	Name                     string      `json:"name"`
	Description              string      `json:"description"`
	FilePath                 string      `json:"filePath"`
	FileName                 string      `json:"fileName"`
	FileSize                 int64       `json:"fileSize"`
	Duration                 float64     `json:"duration"`
	Format                   string      `json:"format"`
	Width                    int         `json:"width"`
	Height                   int         `json:"height"`
	ProjectID                int         `json:"projectId"`
	CreatedAt                string      `json:"createdAt"`
	UpdatedAt                string      `json:"updatedAt"`
	Exists                   bool        `json:"exists"`
	ThumbnailURL             string      `json:"thumbnailUrl"`
	Transcription            string      `json:"transcription"`
	TranscriptionWords       []Word      `json:"transcriptionWords"`
	TranscriptionLanguage    string      `json:"transcriptionLanguage"`
	TranscriptionDuration    float64     `json:"transcriptionDuration"`
	TranscriptionState       string      `json:"transcriptionState"`
	TranscriptionError       string      `json:"transcriptionError"`
	TranscriptionStartedAt   string      `json:"transcriptionStartedAt"`
	TranscriptionCompletedAt string      `json:"transcriptionCompletedAt"`
	Highlights               []Highlight `json:"highlights"`
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

// DeleteProject deletes a project and all its related entities
func (s *ProjectService) DeleteProject(id int) error {
	// Start a transaction to ensure all deletions succeed or fail together
	tx, err := s.client.Tx(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Delete in reverse dependency order to avoid foreign key constraints

	// 1. Delete all chat messages for chat sessions belonging to this project
	chatSessions, err := tx.ChatSession.Query().
		Where(chatsession.HasProjectWith(project.ID(id))).
		All(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to query chat sessions: %w", err)
	}

	for _, session := range chatSessions {
		_, err = tx.ChatMessage.Delete().
			Where(chatmessage.SessionID(session.ID)).
			Exec(s.ctx)
		if err != nil {
			return fmt.Errorf("failed to delete chat messages for session %d: %w", session.ID, err)
		}
	}

	// 2. Delete all chat sessions belonging to this project
	_, err = tx.ChatSession.Delete().
		Where(chatsession.HasProjectWith(project.ID(id))).
		Exec(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to delete chat sessions: %w", err)
	}

	// 3. Delete all export jobs belonging to this project
	_, err = tx.ExportJob.Delete().
		Where(exportjob.HasProjectWith(project.ID(id))).
		Exec(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to delete export jobs: %w", err)
	}

	// 4. Delete all video clips belonging to this project
	_, err = tx.VideoClip.Delete().
		Where(videoclip.HasProjectWith(project.ID(id))).
		Exec(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to delete video clips: %w", err)
	}

	// 5. Finally, delete the project itself
	err = tx.Project.DeleteOneID(id).Exec(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
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
		ID:                       clip.ID,
		Name:                     clip.Name,
		Description:              clip.Description,
		FilePath:                 clip.FilePath,
		FileName:                 fileName,
		FileSize:                 fileSize,
		Duration:                 clip.Duration,
		Format:                   format,
		Width:                    clip.Width,
		Height:                   clip.Height,
		ProjectID:                0, // Will need to be loaded separately
		CreatedAt:                clip.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:                clip.UpdatedAt.Format("2006-01-02 15:04:05"),
		Exists:                   exists,
		ThumbnailURL:             s.getThumbnailURL(filePath),
		Transcription:            clip.Transcription,
		TranscriptionWords:       s.schemaWordsToWords(clip.TranscriptionWords),
		TranscriptionLanguage:    clip.TranscriptionLanguage,
		TranscriptionDuration:    clip.TranscriptionDuration,
		TranscriptionState:       clip.TranscriptionState,
		TranscriptionError:       clip.TranscriptionError,
		TranscriptionStartedAt:   s.formatTime(clip.TranscriptionStartedAt),
		TranscriptionCompletedAt: s.formatTime(clip.TranscriptionCompletedAt),
		Highlights:               s.schemaHighlightsToHighlights(clip.Highlights),
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
			ID:                       clip.ID,
			Name:                     clip.Name,
			Description:              clip.Description,
			FilePath:                 clip.FilePath,
			FileName:                 fileName,
			FileSize:                 fileSize,
			Duration:                 clip.Duration,
			Format:                   format,
			Width:                    clip.Width,
			Height:                   clip.Height,
			ProjectID:                projectID,
			CreatedAt:                clip.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:                clip.UpdatedAt.Format("2006-01-02 15:04:05"),
			Exists:                   exists,
			ThumbnailURL:             s.getThumbnailURL(clip.FilePath),
			Transcription:            clip.Transcription,
			TranscriptionWords:       s.schemaWordsToWords(clip.TranscriptionWords),
			TranscriptionLanguage:    clip.TranscriptionLanguage,
			TranscriptionDuration:    clip.TranscriptionDuration,
			TranscriptionState:       clip.TranscriptionState,
			TranscriptionError:       clip.TranscriptionError,
			TranscriptionStartedAt:   s.formatTime(clip.TranscriptionStartedAt),
			TranscriptionCompletedAt: s.formatTime(clip.TranscriptionCompletedAt),
			Highlights:               s.schemaHighlightsToHighlights(clip.Highlights),
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
		ID:                       clip.ID,
		Name:                     clip.Name,
		Description:              clip.Description,
		FilePath:                 clip.FilePath,
		FileName:                 fileName,
		FileSize:                 fileSize,
		Duration:                 clip.Duration,
		Format:                   format,
		Width:                    clip.Width,
		Height:                   clip.Height,
		ProjectID:                0, // Will need to be loaded separately
		CreatedAt:                clip.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:                clip.UpdatedAt.Format("2006-01-02 15:04:05"),
		Exists:                   exists,
		ThumbnailURL:             s.getThumbnailURL(clip.FilePath),
		Transcription:            clip.Transcription,
		TranscriptionWords:       s.schemaWordsToWords(clip.TranscriptionWords),
		TranscriptionLanguage:    clip.TranscriptionLanguage,
		TranscriptionDuration:    clip.TranscriptionDuration,
		TranscriptionState:       clip.TranscriptionState,
		TranscriptionError:       clip.TranscriptionError,
		TranscriptionStartedAt:   s.formatTime(clip.TranscriptionStartedAt),
		TranscriptionCompletedAt: s.formatTime(clip.TranscriptionCompletedAt),
		Highlights:               s.schemaHighlightsToHighlights(clip.Highlights),
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
	colorCounter := 1
	for _, h := range highlights {
		colorID := h.ColorID

		// Validate ColorID - if invalid (0, negative, or out of range), assign a valid one
		if colorID < 1 || colorID > 20 {
			fmt.Printf("Warning: Invalid ColorID %d for highlight %s, assigning ColorID %d\n", h.ColorID, h.ID, colorCounter)
			colorID = colorCounter
			colorCounter++
			if colorCounter > 20 {
				colorCounter = 1 // Wrap around to color 1
			}
		}

		schemaHighlights = append(schemaHighlights, schema.Highlight{
			ID:      h.ID,
			Start:   h.Start,
			End:     h.End,
			ColorID: colorID,
		})
	}

	// Get the video clip to find its project ID
	clip, err := s.client.VideoClip.
		Query().
		Where(videoclip.ID(clipID)).
		WithProject().
		Only(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to get video clip for real-time update: %w", err)
	}

	_, err = s.client.VideoClip.
		UpdateOneID(clipID).
		SetHighlights(schemaHighlights).
		Save(s.ctx)
	if err != nil {
		return err
	}

	// Broadcast real-time update
	if clip.Edges.Project != nil {
		projectID := strconv.Itoa(clip.Edges.Project.ID)
		manager := realtime.GetManager()

		// Get the full project highlights structure for broadcasting
		highlightService := highlightsservice.NewHighlightService(s.client, s.ctx)
		projectHighlights, err := highlightService.GetProjectHighlights(clip.Edges.Project.ID)
		if err == nil {
			manager.BroadcastHighlightsUpdate(projectID, projectHighlights)
		}
	}

	return nil
}

// UpdateVideoClipSuggestedHighlights updates the suggested highlights for a video clip
func (s *ProjectService) UpdateVideoClipSuggestedHighlights(clipID int, suggestedHighlights []Highlight) error {
	// Convert Highlights to schema.Highlights for database storage
	var schemaHighlights []schema.Highlight
	colorCounter := 1
	for _, h := range suggestedHighlights {
		colorID := h.ColorID

		// Validate ColorID - if invalid (0, negative, or out of range), assign a valid one
		if colorID < 1 || colorID > 20 {
			fmt.Printf("Warning: Invalid ColorID %d for suggested highlight %s, assigning ColorID %d\n", h.ColorID, h.ID, colorCounter)
			colorID = colorCounter
			colorCounter++
			if colorCounter > 20 {
				colorCounter = 1 // Wrap around to color 1
			}
		}

		schemaHighlights = append(schemaHighlights, schema.Highlight{
			ID:      h.ID,
			Start:   h.Start,
			End:     h.End,
			ColorID: colorID,
		})
	}

	// Get clip with project information for broadcasting
	clip, err := s.client.VideoClip.
		Query().
		Where(videoclip.ID(clipID)).
		WithProject().
		Only(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to get video clip: %w", err)
	}

	_, err = s.client.VideoClip.
		UpdateOneID(clipID).
		SetSuggestedHighlights(schemaHighlights).
		Save(s.ctx)

	if err != nil {
		return err
	}

	// Broadcast real-time update for suggested highlights
	if clip.Edges.Project != nil {
		projectIDStr := strconv.Itoa(clip.Edges.Project.ID)
		manager := realtime.GetManager()

		// Get the full project highlights structure for broadcasting
		highlightService := highlightsservice.NewHighlightService(s.client, s.ctx)
		projectHighlights, err := highlightService.GetProjectHighlights(clip.Edges.Project.ID)
		if err == nil {
			manager.BroadcastHighlightsUpdate(projectIDStr, projectHighlights)
		}
	}

	return nil
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

	// Convert []string to []interface{} for database storage
	interfaceOrder := make([]interface{}, len(highlightOrder))
	for i, v := range highlightOrder {
		interfaceOrder[i] = v
	}

	// Update the highlight order in the project schema
	_, err = s.client.Project.
		UpdateOneID(projectID).
		SetHighlightOrder(interfaceOrder).
		Save(s.ctx)

	if err != nil {
		return fmt.Errorf("failed to update project highlight order: %w", err)
	}

	// Broadcast real-time update
	projectIDStr := strconv.Itoa(projectID)
	manager := realtime.GetManager()
	manager.BroadcastHighlightsReorder(projectIDStr, interfaceOrder)

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
			ID:      sh.ID,
			Start:   sh.Start,
			End:     sh.End,
			ColorID: sh.ColorID,
		})
	}
	return highlights
}

// formatTime formats a time value to a string, handling both time.Time and *time.Time
func (s *ProjectService) formatTime(t interface{}) string {
	switch v := t.(type) {
	case *time.Time:
		if v == nil {
			return ""
		}
		return v.Format("2006-01-02 15:04:05")
	case time.Time:
		if v.IsZero() {
			return ""
		}
		return v.Format("2006-01-02 15:04:05")
	default:
		return ""
	}
}

// equalStringSlices checks if two string slices are equal
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// equalInterfaceSlices checks if two interface{} slices are equal
func equalInterfaceSlices(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if !equalInterfaces(v, b[i]) {
			return false
		}
	}
	return true
}

// equalInterfaces checks if two interface{} values are equal
func equalInterfaces(a, b interface{}) bool {
	switch va := a.(type) {
	case string:
		if vb, ok := b.(string); ok {
			return va == vb
		}
		return false
	case map[string]interface{}:
		if vb, ok := b.(map[string]interface{}); ok {
			if len(va) != len(vb) {
				return false
			}
			for k, v := range va {
				if !equalInterfaces(v, vb[k]) {
					return false
				}
			}
			return true
		}
		return false
	default:
		return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
	}
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

	// Get current order from project schema and preserve full objects for history
	var currentOrder []interface{}
	if project.HighlightOrder != nil {
		// Store full objects in history to preserve titles
		currentOrder = make([]interface{}, len(project.HighlightOrder))
		copy(currentOrder, project.HighlightOrder)
	} else {
		currentOrder = []interface{}{}
	}

	// Get current history
	history := project.OrderHistory
	if history == nil {
		history = [][]interface{}{}
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
	currentIndex := project.OrderHistoryIndex

	// Special handling when we're at the current state (index == -1)
	if currentIndex == -1 {
		// Before undoing from current state, we need to save the current state to history
		// so we can redo back to it later

		// Get current order and preserve full objects
		var currentOrder []interface{}
		if project.HighlightOrder != nil {
			currentOrder = make([]interface{}, len(project.HighlightOrder))
			copy(currentOrder, project.HighlightOrder)
		}

		// Add current state to history if it's different from the last history entry
		if len(history) == 0 || !equalInterfaceSlices(currentOrder, history[len(history)-1]) {
			history = append(history, currentOrder)
			if len(history) > 20 {
				history = history[1:] // Remove oldest entry
			}

			// Update history in database
			project, err = s.client.Project.
				UpdateOneID(projectID).
				SetOrderHistory(history).
				Save(s.ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to update history: %w", err)
			}
		}
	}

	// Refresh history after potential update
	history = project.OrderHistory
	if history == nil || len(history) == 0 {
		return nil, fmt.Errorf("no history available")
	}

	// Calculate new index (move backward)
	var newIndex int
	if currentIndex == -1 {
		// We're at current state, move to second-to-last history entry (since we just added current state)
		if len(history) < 2 {
			return nil, fmt.Errorf("no previous state to undo to")
		}
		newIndex = len(history) - 2
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
		SetHighlightOrder(orderFromHistory).
		Save(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update history index and apply order: %w", err)
	}

	// Broadcast real-time update
	projectIDStr := strconv.Itoa(projectID)
	manager := realtime.GetManager()
	manager.BroadcastHighlightsReorder(projectIDStr, orderFromHistory)

	// Convert to string array for return value (for backward compatibility)
	var orderStrings []string
	for _, item := range orderFromHistory {
		switch v := item.(type) {
		case string:
			orderStrings = append(orderStrings, v)
		case map[string]interface{}:
			if typeVal, ok := v["type"].(string); ok && typeVal == "N" {
				orderStrings = append(orderStrings, "N")
			}
		default:
			orderStrings = append(orderStrings, fmt.Sprintf("%v", v))
		}
	}

	return orderStrings, nil
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
	if currentIndex == -1 {
		// Already at current state
		return nil, fmt.Errorf("cannot redo further")
	}

	if currentIndex >= len(history)-1 {
		// At the last history entry, check if we can move to current state
		// We can only move to current state (-1) if the current project order
		// is different from the last history entry
		var currentOrder []interface{}
		if project.HighlightOrder != nil {
			currentOrder = make([]interface{}, len(project.HighlightOrder))
			copy(currentOrder, project.HighlightOrder)
		}

		// If last history entry matches current state, we can't redo
		if equalInterfaceSlices(history[len(history)-1], currentOrder) {
			return nil, fmt.Errorf("cannot redo further")
		}

		// Move to current state
		newIndex := -1

		// Update project index
		_, err = s.client.Project.
			UpdateOneID(projectID).
			SetOrderHistoryIndex(newIndex).
			Save(s.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to update history index: %w", err)
		}

		// Convert current order to string array for return value
		var currentOrderStrings []string
		for _, item := range currentOrder {
			switch v := item.(type) {
			case string:
				currentOrderStrings = append(currentOrderStrings, v)
			case map[string]interface{}:
				if typeVal, ok := v["type"].(string); ok && typeVal == "N" {
					currentOrderStrings = append(currentOrderStrings, "N")
				}
			default:
				currentOrderStrings = append(currentOrderStrings, fmt.Sprintf("%v", v))
			}
		}

		// Return the current order (which is already applied)
		return currentOrderStrings, nil
	}

	// Normal case: move forward in history
	newIndex := currentIndex + 1

	// Get order from history
	orderFromHistory := history[newIndex]

	// Update project index and apply the order
	_, err = s.client.Project.
		UpdateOneID(projectID).
		SetOrderHistoryIndex(newIndex).
		SetHighlightOrder(orderFromHistory).
		Save(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update history index and apply order: %w", err)
	}

	// Broadcast real-time update
	projectIDStr := strconv.Itoa(projectID)
	manager := realtime.GetManager()
	manager.BroadcastHighlightsReorder(projectIDStr, orderFromHistory)

	// Convert to string array for return value (for backward compatibility)
	var orderStrings []string
	for _, item := range orderFromHistory {
		switch v := item.(type) {
		case string:
			orderStrings = append(orderStrings, v)
		case map[string]interface{}:
			if typeVal, ok := v["type"].(string); ok && typeVal == "N" {
				orderStrings = append(orderStrings, "N")
			}
		default:
			orderStrings = append(orderStrings, fmt.Sprintf("%v", v))
		}
	}

	return orderStrings, nil
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

	// Can undo if:
	// 1. We're at current state (index == -1) and have history
	// 2. We're in history and not at the oldest entry (index > 0)
	canUndo := (currentIndex == -1 && len(history) > 0) || currentIndex > 0

	// Can redo if:
	// 1. We're in history and not at the last entry (index < len(history)-1)
	// 2. We're at the last history entry and current state differs from it
	canRedo := false
	if currentIndex != -1 {
		if currentIndex < len(history)-1 {
			// Not at the last history entry
			canRedo = true
		} else if currentIndex == len(history)-1 {
			// At last history entry, check if current state differs
			var currentOrder []interface{}
			if project.HighlightOrder != nil {
				currentOrder = make([]interface{}, len(project.HighlightOrder))
				copy(currentOrder, project.HighlightOrder)
			}

			// Can redo if current state differs from last history entry
			canRedo = !equalInterfaceSlices(history[len(history)-1], currentOrder)
		}
	}

	return canUndo, canRedo, nil
}

// UndoHighlightsChange moves backward in highlights history
func (s *ProjectService) UndoHighlightsChange(clipID int) ([]Highlight, error) {
	// Get current video clip with history and project
	clip, err := s.client.VideoClip.
		Query().
		Where(videoclip.ID(clipID)).
		WithProject().
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

	// Broadcast real-time update
	if clip.Edges.Project != nil {
		projectIDStr := strconv.Itoa(clip.Edges.Project.ID)
		manager := realtime.GetManager()

		// Get the full project highlights structure for broadcasting
		highlightService := highlightsservice.NewHighlightService(s.client, s.ctx)
		projectHighlights, err := highlightService.GetProjectHighlights(clip.Edges.Project.ID)
		if err == nil {
			manager.BroadcastHighlightsUpdate(projectIDStr, projectHighlights)
		}
	}

	// Convert to return format
	return s.schemaHighlightsToHighlights(highlightsFromHistory), nil
}

// RedoHighlightsChange moves forward in highlights history
func (s *ProjectService) RedoHighlightsChange(clipID int) ([]Highlight, error) {
	// Get current video clip with history and project
	clip, err := s.client.VideoClip.
		Query().
		Where(videoclip.ID(clipID)).
		WithProject().
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

	// Broadcast real-time update
	if clip.Edges.Project != nil {
		projectIDStr := strconv.Itoa(clip.Edges.Project.ID)
		manager := realtime.GetManager()

		// Get the full project highlights structure for broadcasting
		highlightService := highlightsservice.NewHighlightService(s.client, s.ctx)
		projectHighlights, err := highlightService.GetProjectHighlights(clip.Edges.Project.ID)
		if err == nil {
			manager.BroadcastHighlightsUpdate(projectIDStr, projectHighlights)
		}
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

// SaveSectionTitle saves or updates the title for a newline section at a specific position
func (s *ProjectService) SaveSectionTitle(projectID int, position int, title string) error {
	// Get current highlight order
	highlightOrder, err := s.getProjectHighlightOrderWithTitles(projectID)
	if err != nil {
		return fmt.Errorf("failed to get current highlight order: %w", err)
	}

	// Validate position
	if position < 0 || position >= len(highlightOrder) {
		return fmt.Errorf("invalid position %d for highlight order of length %d", position, len(highlightOrder))
	}

	// Check if the item at position is a newline
	item := highlightOrder[position]
	switch v := item.(type) {
	case string:
		if v == "N" {
			// Convert simple "N" to rich newline object with title
			highlightOrder[position] = NewlineSection{Type: "N", Title: title}
		} else {
			return fmt.Errorf("item at position %d is not a newline section", position)
		}
	case map[string]interface{}:
		if typeVal, ok := v["type"].(string); ok && typeVal == "N" {
			// Update existing newline object
			v["title"] = title
			highlightOrder[position] = v
		} else {
			return fmt.Errorf("item at position %d is not a newline section", position)
		}
	case NewlineSection:
		if v.Type == "N" {
			// Update existing NewlineSection
			v.Title = title
			highlightOrder[position] = v
		} else {
			return fmt.Errorf("item at position %d is not a newline section", position)
		}
	default:
		return fmt.Errorf("item at position %d is not a newline section", position)
	}

	// Save the updated order
	return s.UpdateProjectHighlightOrderWithTitles(projectID, highlightOrder)
}

// GetSectionTitles retrieves all section titles from the project highlight order
func (s *ProjectService) GetSectionTitles(projectID int) (map[int]string, error) {
	highlightOrder, err := s.getProjectHighlightOrderWithTitles(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get highlight order: %w", err)
	}

	titles := make(map[int]string)
	for i, item := range highlightOrder {
		switch v := item.(type) {
		case map[string]interface{}:
			if typeVal, ok := v["type"].(string); ok && typeVal == "N" {
				if titleVal, ok := v["title"].(string); ok && titleVal != "" {
					titles[i] = titleVal
				}
			}
		case NewlineSection:
			if v.Type == "N" && v.Title != "" {
				titles[i] = v.Title
			}
		}
	}

	return titles, nil
}

// UpdateProjectHighlightOrderWithTitles updates the highlight order with rich newline objects
func (s *ProjectService) UpdateProjectHighlightOrderWithTitles(projectID int, highlightOrder []interface{}) error {
	// Save current state to history before making changes
	err := s.saveOrderState(projectID)
	if err != nil {
		// Log error but don't fail the update
		fmt.Printf("Warning: failed to save order state to history: %v\n", err)
	}

	// Convert interface{} array to the format expected by the database
	// The database expects a JSON-serializable array
	_, err = s.client.Project.
		UpdateOneID(projectID).
		SetHighlightOrder(highlightOrder).
		Save(s.ctx)

	if err != nil {
		return fmt.Errorf("failed to update project highlight order with titles: %w", err)
	}

	// Broadcast real-time update
	projectIDStr := strconv.Itoa(projectID)
	manager := realtime.GetManager()
	manager.BroadcastHighlightsReorder(projectIDStr, highlightOrder)

	return nil
}

// GetProjectHighlightOrderWithTitles retrieves the highlight order with rich newline objects
func (s *ProjectService) GetProjectHighlightOrderWithTitles(projectID int) ([]interface{}, error) {
	return s.getProjectHighlightOrderWithTitles(projectID)
}
// HideHighlight adds a highlight to the hidden highlights list
func (s *ProjectService) HideHighlight(projectID int, highlightID string) error {
	proj, err := s.client.Project.
		Query().
		Where(project.ID(projectID)).
		Only(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}
	
	// Get current hidden highlights
	hiddenHighlights := proj.HiddenHighlights
	if hiddenHighlights == nil {
		hiddenHighlights = []string{}
	}
	
	// Check if already hidden
	for _, id := range hiddenHighlights {
		if id == highlightID {
			return nil // Already hidden
		}
	}
	
	// Add to hidden highlights
	hiddenHighlights = append(hiddenHighlights, highlightID)
	
	// Update database
	_, err = s.client.Project.
		UpdateOneID(projectID).
		SetHiddenHighlights(hiddenHighlights).
		Save(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to hide highlight: %w", err)
	}
	
	// Remove from highlight order if present
	highlightOrder := proj.HighlightOrder
	if highlightOrder != nil {
		updatedOrder := make([]interface{}, 0, len(highlightOrder))
		for _, item := range highlightOrder {
			if str, ok := item.(string); !ok || str != highlightID {
				updatedOrder = append(updatedOrder, item)
			}
		}
		if len(updatedOrder) != len(highlightOrder) {
			_, err = s.client.Project.
				UpdateOneID(projectID).
				SetHighlightOrder(updatedOrder).
				Save(s.ctx)
			if err != nil {
				return fmt.Errorf("failed to update highlight order: %w", err)
			}
		}
	}
	
	// Broadcast real-time update
	projectIDStr := strconv.Itoa(projectID)
	manager := realtime.GetManager()
	manager.BroadcastHighlightsReorder(projectIDStr, proj.HighlightOrder)
	
	return nil
}
// UnhideHighlight removes a highlight from the hidden highlights list
func (s *ProjectService) UnhideHighlight(projectID int, highlightID string) error {
	proj, err := s.client.Project.
		Query().
		Where(project.ID(projectID)).
		Only(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}
	
	// Get current hidden highlights
	hiddenHighlights := proj.HiddenHighlights
	if hiddenHighlights == nil {
		return nil // Nothing to unhide
	}
	
	// Remove from hidden highlights
	updatedHidden := make([]string, 0, len(hiddenHighlights))
	for _, id := range hiddenHighlights {
		if id != highlightID {
			updatedHidden = append(updatedHidden, id)
		}
	}
	
	// Update database
	_, err = s.client.Project.
		UpdateOneID(projectID).
		SetHiddenHighlights(updatedHidden).
		Save(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to unhide highlight: %w", err)
	}
	
	// Broadcast real-time update
	projectIDStr := strconv.Itoa(projectID)
	manager := realtime.GetManager()
	manager.BroadcastHighlightsReorder(projectIDStr, proj.HighlightOrder)
	
	return nil
}
// GetHiddenHighlights retrieves the list of hidden highlight IDs for a project
func (s *ProjectService) GetHiddenHighlights(projectID int) ([]string, error) {
	proj, err := s.client.Project.
		Query().
		Where(project.ID(projectID)).
		Only(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	if proj.HiddenHighlights == nil {
		return []string{}, nil
	}
	
	return proj.HiddenHighlights, nil
}

// getProjectHighlightOrderWithTitles is a helper that gets the raw highlight order
func (s *ProjectService) getProjectHighlightOrderWithTitles(projectID int) ([]interface{}, error) {
	proj, err := s.client.Project.
		Query().
		Where(project.ID(projectID)).
		Only(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Return the highlight order as-is, which can contain both strings and objects
	if proj.HighlightOrder == nil {
		return []interface{}{}, nil
	}

	// Since the database field is already []interface{}, just return it directly
	return proj.HighlightOrder, nil
}

// Transcription Methods

// TranscribeVideoClip transcribes audio from a video clip using OpenAI Whisper
func (s *ProjectService) TranscribeVideoClip(clipID int) (*TranscriptionResponse, error) {
	// Update transcription state to checking
	err := s.updateTranscriptionState(clipID, TranscriptionStateChecking, "")
	if err != nil {
		log.Printf("[TRANSCRIPTION] Warning: failed to update state to checking: %v", err)
	}

	// Get the video clip
	clip, err := s.client.VideoClip.Get(s.ctx, clipID)
	if err != nil {
		s.updateTranscriptionState(clipID, TranscriptionStateError, "Video clip not found")
		return &TranscriptionResponse{
			Success: false,
			Message: "Video clip not found",
		}, nil
	}

	// Check if file exists
	if _, err := os.Stat(clip.FilePath); os.IsNotExist(err) {
		s.updateTranscriptionState(clipID, TranscriptionStateError, "Video file not found")
		return &TranscriptionResponse{
			Success: false,
			Message: "Video file not found",
		}, nil
	}

	// Get OpenAI API key
	apiKey, err := s.getOpenAIApiKey()
	if err != nil || apiKey == "" {
		s.updateTranscriptionState(clipID, TranscriptionStateError, "OpenAI API key not configured")
		return &TranscriptionResponse{
			Success: false,
			Message: "OpenAI API key not configured",
		}, nil
	}

	// Update state to transcribing
	err = s.updateTranscriptionState(clipID, TranscriptionStateTranscribing, "")
	if err != nil {
		log.Printf("[TRANSCRIPTION] Warning: failed to update state to transcribing: %v", err)
	}

	// Extract audio from video
	audioPath, err := s.extractAudio(clip.FilePath)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to extract audio: %v", err)
		s.updateTranscriptionState(clipID, TranscriptionStateError, errMsg)
		return &TranscriptionResponse{
			Success: false,
			Message: errMsg,
		}, nil
	}
	defer os.Remove(audioPath) // Clean up temporary audio file

	// Transcribe audio using AI service factory
	factory := ai.NewAIServiceFactory(s.client, s.ctx)
	aiService, err := factory.CreateService()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create AI service: %v", err)
		s.updateTranscriptionState(clipID, TranscriptionStateError, errMsg)
		return &TranscriptionResponse{
			Success: false,
			Message: errMsg,
		}, nil
	}
	
	result, err := aiService.ProcessAudio(audioPath)
	if err != nil {
		errMsg := fmt.Sprintf("Transcription failed: %v", err)
		s.updateTranscriptionState(clipID, TranscriptionStateError, errMsg)
		return &TranscriptionResponse{
			Success: false,
			Message: errMsg,
		}, nil
	}

	// Convert result to WhisperResponse format for compatibility
	var convertedSegments []Segment
	for _, seg := range result.Segments {
		var convertedWords []Word
		for _, w := range seg.Words {
			convertedWords = append(convertedWords, Word{
				Word:  w.Word,
				Start: w.Start,
				End:   w.End,
			})
		}
		convertedSegments = append(convertedSegments, Segment{
			ID:               seg.ID,
			Seek:             seg.Seek,
			Start:            seg.Start,
			End:              seg.End,
			Text:             seg.Text,
			Tokens:           seg.Tokens,
			Temperature:      seg.Temperature,
			AvgLogprob:       seg.AvgLogprob,
			CompressionRatio: seg.CompressionRatio,
			NoSpeechProb:     seg.NoSpeechProb,
			Words:            convertedWords,
		})
	}

	var convertedWords []Word
	for _, w := range result.Words {
		convertedWords = append(convertedWords, Word{
			Word:  w.Word,
			Start: w.Start,
			End:   w.End,
		})
	}

	whisperResponse := &WhisperResponse{
		Task:     "transcribe",
		Language: result.Language,
		Duration: result.Duration,
		Text:     result.Transcript,
		Segments: convertedSegments,
		Words:    convertedWords,
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

	// Save transcription to database and update state to completed
	_, err = s.client.VideoClip.
		UpdateOneID(clipID).
		SetTranscription(whisperResponse.Text).
		SetTranscriptionWords(wordsForStorage).
		SetTranscriptionLanguage(whisperResponse.Language).
		SetTranscriptionDuration(whisperResponse.Duration).
		SetTranscriptionState(TranscriptionStateCompleted).
		SetTranscriptionError("").
		SetTranscriptionCompletedAt(time.Now()).
		Save(s.ctx)

	if err != nil {
		s.updateTranscriptionState(clipID, TranscriptionStateError, "Failed to save transcription")
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

// BatchTranscribeResponse represents the response from batch transcription
type BatchTranscribeResponse struct {
	Success           bool     `json:"success"`
	Message           string   `json:"message"`
	TranscribedCount  int      `json:"transcribedCount"`
	SkippedCount      int      `json:"skippedCount"`
	FailedCount       int      `json:"failedCount"`
	FailedClips       []string `json:"failedClips"`
}

// BatchTranscribeUntranscribedClips transcribes all video clips in a project that haven't been transcribed yet
func (s *ProjectService) BatchTranscribeUntranscribedClips(projectID int) (*BatchTranscribeResponse, error) {
	// Get all video clips for the project that haven't been transcribed
	clips, err := s.client.VideoClip.
		Query().
		Where(
			videoclip.HasProjectWith(project.ID(projectID)),
			videoclip.Or(
				videoclip.TranscriptionIsNil(),
				videoclip.TranscriptionEQ(""),
				videoclip.TranscriptionStateEQ(TranscriptionStateError),
			),
		).
		All(s.ctx)
	
	if err != nil {
		return &BatchTranscribeResponse{
			Success: false,
			Message: "Failed to get video clips",
		}, err
	}

	if len(clips) == 0 {
		return &BatchTranscribeResponse{
			Success: true,
			Message: "No untranscribed video clips found",
			TranscribedCount: 0,
			SkippedCount: 0,
			FailedCount: 0,
		}, nil
	}

	log.Printf("[BATCH_TRANSCRIPTION] Starting batch transcription for project %d with %d clips", projectID, len(clips))

	// Check if OpenAI API key is configured once
	apiKey, err := s.getOpenAIApiKey()
	if err != nil || apiKey == "" {
		return &BatchTranscribeResponse{
			Success: false,
			Message: "OpenAI API key not configured",
		}, nil
	}

	var transcribedCount, skippedCount, failedCount int
	var failedClips []string

	// Process each clip
	for _, clip := range clips {
		log.Printf("[BATCH_TRANSCRIPTION] Processing clip: %s (ID: %d)", clip.Name, clip.ID)

		// Check if file exists
		if _, err := os.Stat(clip.FilePath); os.IsNotExist(err) {
			log.Printf("[BATCH_TRANSCRIPTION] File not found for clip %s: %s", clip.Name, clip.FilePath)
			failedCount++
			failedClips = append(failedClips, fmt.Sprintf("%s (file not found)", clip.Name))
			continue
		}

		// Start transcription for this clip
		result, err := s.TranscribeVideoClip(clip.ID)
		if err != nil {
			log.Printf("[BATCH_TRANSCRIPTION] Error transcribing clip %s: %v", clip.Name, err)
			failedCount++
			failedClips = append(failedClips, fmt.Sprintf("%s (error: %v)", clip.Name, err))
			continue
		}

		if result.Success {
			log.Printf("[BATCH_TRANSCRIPTION] Successfully transcribed clip: %s", clip.Name)
			transcribedCount++
		} else {
			log.Printf("[BATCH_TRANSCRIPTION] Failed to transcribe clip %s: %s", clip.Name, result.Message)
			failedCount++
			failedClips = append(failedClips, fmt.Sprintf("%s (%s)", clip.Name, result.Message))
		}
	}

	message := fmt.Sprintf("Batch transcription completed: %d transcribed, %d failed", transcribedCount, failedCount)
	if len(failedClips) > 0 {
		message = fmt.Sprintf("%s. Failed clips: %s", message, strings.Join(failedClips, ", "))
	}

	log.Printf("[BATCH_TRANSCRIPTION] Completed batch transcription for project %d: %s", projectID, message)

	return &BatchTranscribeResponse{
		Success:          transcribedCount > 0 || failedCount == 0,
		Message:          message,
		TranscribedCount: transcribedCount,
		SkippedCount:     skippedCount,
		FailedCount:      failedCount,
		FailedClips:      failedClips,
	}, nil
}

// extractAudio extracts audio from a video file using ffmpeg with optimized settings
func (s *ProjectService) extractAudio(videoPath string) (string, error) {
	// Create temp directory for audio files using system temp dir
	tempDir := filepath.Join(os.TempDir(), "ramble_audio")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate unique audio filename
	audioFilename := fmt.Sprintf("audio_%d.mp3", time.Now().UnixNano())
	audioPath := filepath.Join(tempDir, audioFilename)

	log.Printf("[TRANSCRIPTION] Extracting audio from: %s to: %s", videoPath, audioPath)

	// Use ffmpeg to extract audio with optimized settings for Whisper
	cmd := goapp.GetFFmpegCommand(
		"-i", videoPath,
		"-vn",            // No video
		"-acodec", "mp3", // MP3 codec (guaranteed Whisper support)
		"-ar", "16000",   // Sample rate (16kHz for Whisper)
		"-ac", "1",       // Mono channel
		"-b:a", "24k",    // Low bitrate for significant space savings (reduced from 64k)
		"-af", "highpass=f=80,lowpass=f=8000", // Filter frequencies outside speech range
		"-y",             // Overwrite output file
		audioPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[TRANSCRIPTION] ffmpeg error: %v, output: %s", err, string(output))
		return "", fmt.Errorf("ffmpeg failed: %w", err)
	}

	// Get file size for logging
	if stat, err := os.Stat(audioPath); err == nil {
		sizeMB := float64(stat.Size()) / (1024 * 1024)
		log.Printf("[TRANSCRIPTION] Audio extracted successfully: %s (%.2f MB)", audioPath, sizeMB)
	} else {
		log.Printf("[TRANSCRIPTION] Audio extracted successfully: %s", audioPath)
	}

	return audioPath, nil
}

// transcribeAudio transcribes audio using OpenAI Whisper API
func (s *ProjectService) transcribeAudio(audioPath, apiKey string) (*WhisperResponse, error) {
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

// getOpenAIApiKey retrieves the OpenAI API key from settings
func (s *ProjectService) getOpenAIApiKey() (string, error) {
	return s.getSetting("openai_api_key")
}

// getSetting retrieves a setting value by key
func (s *ProjectService) getSetting(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("setting key cannot be empty")
	}

	setting, err := s.client.Settings.
		Query().
		Where(settings.Key(key)).
		Only(s.ctx)

	if err != nil {
		// Return empty string if setting doesn't exist
		return "", nil
	}

	return setting.Value, nil
}

// updateTranscriptionState updates the transcription state and error message for a video clip
func (s *ProjectService) updateTranscriptionState(clipID int, state string, errorMsg string) error {
	update := s.client.VideoClip.UpdateOneID(clipID).SetTranscriptionState(state)

	if state == TranscriptionStateTranscribing && errorMsg == "" {
		update = update.SetTranscriptionStartedAt(time.Now())
	}

	if errorMsg != "" {
		update = update.SetTranscriptionError(errorMsg)
	} else {
		update = update.ClearTranscriptionError()
	}

	_, err := update.Save(s.ctx)
	return err
}

// SaveTranscriptionResult saves AI transcription result to database
func (s *ProjectService) SaveTranscriptionResult(clipID int, result *ai.AudioProcessingResult) (*TranscriptionResponse, error) {
	if result == nil {
		return &TranscriptionResponse{
			Success: false,
			Message: "No transcription result provided",
		}, nil
	}

	// Update transcription state to processing
	err := s.updateTranscriptionState(clipID, TranscriptionStateTranscribing, "")
	if err != nil {
		log.Printf("[TRANSCRIPTION] Warning: failed to update state to transcribing: %v", err)
	}

	// Convert AI Words to schema Words for storage
	var wordsForStorage []schema.Word
	for _, w := range result.Words {
		wordsForStorage = append(wordsForStorage, schema.Word{
			Word:  w.Word,
			Start: w.Start,
			End:   w.End,
		})
	}

	// Save transcription to database and update state to completed
	_, err = s.client.VideoClip.
		UpdateOneID(clipID).
		SetTranscription(result.Transcript).
		SetTranscriptionWords(wordsForStorage).
		SetTranscriptionLanguage(result.Language).
		SetTranscriptionDuration(result.Duration).
		SetTranscriptionState(TranscriptionStateCompleted).
		SetTranscriptionError("").
		SetTranscriptionCompletedAt(time.Now()).
		Save(s.ctx)

	if err != nil {
		s.updateTranscriptionState(clipID, TranscriptionStateError, "Failed to save transcription")
		return &TranscriptionResponse{
			Success: false,
			Message: "Failed to save transcription",
		}, nil
	}

	log.Printf("[TRANSCRIPTION] Transcription saved successfully, text length: %d characters, words: %d",
		len(result.Transcript), len(result.Words))

	// Convert AI Words to projects Words for response
	var responseWords []Word
	for _, w := range result.Words {
		responseWords = append(responseWords, Word{
			Word:  w.Word,
			Start: w.Start,
			End:   w.End,
		})
	}

	return &TranscriptionResponse{
		Success:       true,
		Message:       "Transcription completed successfully",
		Transcription: result.Transcript,
		Words:         responseWords,
		Language:      result.Language,
		Duration:      result.Duration,
	}, nil
}

// BatchTranscribeWithAIService transcribes all untranscribed video clips using provided AI service
func (s *ProjectService) BatchTranscribeWithAIService(projectID int, aiService ai.AIService) (*BatchTranscribeResponse, error) {
	// Get untranscribed clips  
	clips, err := s.client.VideoClip.
		Query().
		Where(
			videoclip.HasProjectWith(project.ID(projectID)),
			videoclip.TranscriptionStateNEQ(TranscriptionStateCompleted),
		).
		All(s.ctx)
	if err != nil {
		return &BatchTranscribeResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get clips: %v", err),
		}, nil
	}

	if len(clips) == 0 {
		return &BatchTranscribeResponse{
			Success: true,
			Message: "No clips found that need transcription",
		}, nil
	}

	log.Printf("[BATCH_TRANSCRIPTION] Starting batch transcription for project %d with %d clips", projectID, len(clips))

	var transcribedCount, skippedCount, failedCount int
	var failedClips []string

	// Process each clip
	for _, clip := range clips {
		log.Printf("[BATCH_TRANSCRIPTION] Processing clip: %s (ID: %d)", clip.Name, clip.ID)

		// Check if file exists
		if _, err := os.Stat(clip.FilePath); os.IsNotExist(err) {
			log.Printf("[BATCH_TRANSCRIPTION] File not found for clip %s: %s", clip.Name, clip.FilePath)
			failedCount++
			failedClips = append(failedClips, fmt.Sprintf("%s (file not found)", clip.Name))
			continue
		}

		// Process audio using AI service
		result, err := aiService.ProcessAudio(clip.FilePath)
		if err != nil {
			log.Printf("[BATCH_TRANSCRIPTION] Error transcribing clip %s: %v", clip.Name, err)
			failedCount++
			failedClips = append(failedClips, fmt.Sprintf("%s (error: %v)", clip.Name, err))
			continue
		}

		// Save transcription result
		transcriptionResponse, err := s.SaveTranscriptionResult(clip.ID, result)
		if err != nil {
			log.Printf("[BATCH_TRANSCRIPTION] Error saving transcription for clip %s: %v", clip.Name, err)
			failedCount++
			failedClips = append(failedClips, fmt.Sprintf("%s (save error: %v)", clip.Name, err))
			continue
		}

		if transcriptionResponse.Success {
			log.Printf("[BATCH_TRANSCRIPTION] Successfully transcribed clip: %s", clip.Name)
			transcribedCount++
		} else {
			log.Printf("[BATCH_TRANSCRIPTION] Failed to transcribe clip %s: %s", clip.Name, transcriptionResponse.Message)
			failedCount++
			failedClips = append(failedClips, fmt.Sprintf("%s (%s)", clip.Name, transcriptionResponse.Message))
		}
	}

	// Build summary message
	var message string
	if transcribedCount > 0 {
		message = fmt.Sprintf("Successfully transcribed %d clips", transcribedCount)
		if failedCount > 0 {
			message += fmt.Sprintf(", %d failed", failedCount)
		}
		if skippedCount > 0 {
			message += fmt.Sprintf(", %d skipped", skippedCount)
		}
	} else if failedCount > 0 {
		message = fmt.Sprintf("All %d clips failed transcription", failedCount)
	} else {
		message = "No clips processed"
	}

	log.Printf("[BATCH_TRANSCRIPTION] Completed: %s", message)

	return &BatchTranscribeResponse{
		Success:          transcribedCount > 0 || (failedCount == 0 && skippedCount > 0),
		Message:          message,
		TranscribedCount: transcribedCount,
		SkippedCount:     skippedCount,
		FailedCount:      failedCount,
		FailedClips:      failedClips,
	}, nil
}
