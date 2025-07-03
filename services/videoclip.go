package services

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
	"MYAPP/utils"
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

// VideoClipResponse represents a video clip response for the frontend
type VideoClipResponse struct {
	ID                     int         `json:"id"`
	Name                   string      `json:"name"`
	Description            string      `json:"description"`
	FilePath               string      `json:"filePath"`
	FileName               string      `json:"fileName"`
	FileSize               int64       `json:"fileSize"`
	Duration               float64     `json:"duration"`
	Format                 string      `json:"format"`
	Width                  int         `json:"width"`
	Height                 int         `json:"height"`
	ProjectID              int         `json:"projectId"`
	CreatedAt              string      `json:"createdAt"`
	UpdatedAt              string      `json:"updatedAt"`
	Exists                 bool        `json:"exists"`
	ThumbnailURL           string      `json:"thumbnailUrl"`
	Transcription          string      `json:"transcription"`
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

// VideoClipService handles video clip-related operations
type VideoClipService struct {
	client *ent.Client
	ctx    context.Context
}

// NewVideoClipService creates a new video clip service
func NewVideoClipService(client *ent.Client, ctx context.Context) *VideoClipService {
	return &VideoClipService{
		client: client,
		ctx:    ctx,
	}
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

// CreateVideoClip creates a new video clip with file validation
func (vcs *VideoClipService) CreateVideoClip(projectID int, filePath string) (*VideoClipResponse, error) {
	// Validate file exists and is a video
	if !utils.IsVideoFile(filePath) {
		return nil, fmt.Errorf("file is not a supported video format")
	}
	
	fileSize, format, exists := utils.GetFileInfo(filePath)
	if !exists {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}
	
	// Check if this file path already exists for this project
	existingClip, err := vcs.client.VideoClip.
		Query().
		Where(
			videoclip.HasProjectWith(project.ID(projectID)),
			videoclip.FilePath(filePath),
		).
		Only(vcs.ctx)
	
	if err == nil {
		// File already exists for this project, return the existing clip
		fileName := filepath.Base(existingClip.FilePath)
		_, _, fileExists := utils.GetFileInfo(existingClip.FilePath)
		
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
			ThumbnailURL:          utils.GetThumbnailURL(existingClip.FilePath),
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
	videoClip, err := vcs.client.VideoClip.
		Create().
		SetName(name).
		SetDescription("").
		SetFilePath(filePath).
		SetFormat(format).
		SetFileSize(fileSize).
		SetProjectID(projectID).
		Save(vcs.ctx)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create video clip: %w", err)
	}
	
	return &VideoClipResponse{
		ID:                    videoClip.ID,
		Name:                  videoClip.Name,
		Description:           videoClip.Description,
		FilePath:              videoClip.FilePath,
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
		ThumbnailURL:          utils.GetThumbnailURL(videoClip.FilePath),
		Transcription:         videoClip.Transcription,
		TranscriptionWords:    schemaWordsToWords(videoClip.TranscriptionWords),
		TranscriptionLanguage: videoClip.TranscriptionLanguage,
		TranscriptionDuration: videoClip.TranscriptionDuration,
		Highlights:            schemaHighlightsToHighlights(videoClip.Highlights),
	}, nil
}

// GetVideoClipsByProject returns all video clips for a project
func (vcs *VideoClipService) GetVideoClipsByProject(projectID int) ([]*VideoClipResponse, error) {
	clips, err := vcs.client.VideoClip.
		Query().
		Where(videoclip.HasProjectWith(project.ID(projectID))).
		All(vcs.ctx)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get video clips: %w", err)
	}
	
	var responses []*VideoClipResponse
	for _, clip := range clips {
		fileName := filepath.Base(clip.FilePath)
		_, _, exists := utils.GetFileInfo(clip.FilePath)
		
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
			ThumbnailURL:          utils.GetThumbnailURL(clip.FilePath),
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
func (vcs *VideoClipService) UpdateVideoClip(id int, name, description string) (*VideoClipResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("video clip name cannot be empty")
	}
	
	updatedClip, err := vcs.client.VideoClip.
		UpdateOneID(id).
		SetName(name).
		SetDescription(description).
		Save(vcs.ctx)
	
	if err != nil {
		return nil, fmt.Errorf("failed to update video clip: %w", err)
	}
	
	fileName := filepath.Base(updatedClip.FilePath)
	_, _, exists := utils.GetFileInfo(updatedClip.FilePath)
	
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
		ThumbnailURL:          utils.GetThumbnailURL(updatedClip.FilePath),
		Transcription:         updatedClip.Transcription,
		TranscriptionWords:    schemaWordsToWords(updatedClip.TranscriptionWords),
		TranscriptionLanguage: updatedClip.TranscriptionLanguage,
		TranscriptionDuration: updatedClip.TranscriptionDuration,
		Highlights:            schemaHighlightsToHighlights(updatedClip.Highlights),
	}, nil
}

// DeleteVideoClip deletes a video clip
func (vcs *VideoClipService) DeleteVideoClip(id int) error {
	err := vcs.client.VideoClip.
		DeleteOneID(id).
		Exec(vcs.ctx)
	
	if err != nil {
		return fmt.Errorf("failed to delete video clip: %w", err)
	}
	
	return nil
}

// SelectVideoFiles opens a file dialog to select video files
func (vcs *VideoClipService) SelectVideoFiles() ([]*LocalVideoFile, error) {
	// Open file dialog for multiple video files
	filePaths, err := runtime.OpenMultipleFilesDialog(vcs.ctx, runtime.OpenDialogOptions{
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
		if !utils.IsVideoFile(filePath) {
			continue // Skip non-video files
		}

		fileSize, format, exists := utils.GetFileInfo(filePath)
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
func (vcs *VideoClipService) GetVideoFileInfo(filePath string) (*LocalVideoFile, error) {
	if !utils.IsVideoFile(filePath) {
		return nil, fmt.Errorf("file is not a supported video format")
	}
	
	fileSize, format, exists := utils.GetFileInfo(filePath)
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
func (vcs *VideoClipService) GetVideoURL(filePath string) (string, error) {
	if !utils.IsVideoFile(filePath) {
		return "", fmt.Errorf("file is not a supported video format")
	}
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", filePath)
	}
	
	// Encode file path for URL safety
	encodedPath := url.QueryEscape(filePath)
	videoURL := fmt.Sprintf("/api/video/%s", encodedPath)
	
	// Return AssetServer URL that will work in the webview
	return videoURL, nil
}