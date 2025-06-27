package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"MYAPP/ent"
	"MYAPP/ent/project"
	"MYAPP/ent/videoclip"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/mattn/go-sqlite3"
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
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	FilePath    string  `json:"filePath"`
	FileName    string  `json:"fileName"`
	FileSize    int64   `json:"fileSize"`
	Duration    float64 `json:"duration"`
	Format      string  `json:"format"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	ProjectID   int     `json:"projectId"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
	Exists      bool    `json:"exists"`
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

	return &App{
		client: client,
	}
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
		ID:          videoClip.ID,
		Name:        videoClip.Name,
		Description: videoClip.Description,
		FilePath:    videoClip.FilePath,
		FileName:    fileName,
		FileSize:    videoClip.FileSize,
		Duration:    videoClip.Duration,
		Format:      videoClip.Format,
		Width:       videoClip.Width,
		Height:      videoClip.Height,
		ProjectID:   projectID,
		CreatedAt:   videoClip.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   videoClip.UpdatedAt.Format("2006-01-02 15:04:05"),
		Exists:      true,
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
			ID:          clip.ID,
			Name:        clip.Name,
			Description: clip.Description,
			FilePath:    clip.FilePath,
			FileName:    fileName,
			FileSize:    clip.FileSize,
			Duration:    clip.Duration,
			Format:      clip.Format,
			Width:       clip.Width,
			Height:      clip.Height,
			ProjectID:   projectID,
			CreatedAt:   clip.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   clip.UpdatedAt.Format("2006-01-02 15:04:05"),
			Exists:      exists,
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
		ID:          updatedClip.ID,
		Name:        updatedClip.Name,
		Description: updatedClip.Description,
		FilePath:    updatedClip.FilePath,
		FileName:    fileName,
		FileSize:    updatedClip.FileSize,
		Duration:    updatedClip.Duration,
		Format:      updatedClip.Format,
		Width:       updatedClip.Width,
		Height:      updatedClip.Height,
		ProjectID:   updatedClip.Edges.Project.ID,
		CreatedAt:   updatedClip.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   updatedClip.UpdatedAt.Format("2006-01-02 15:04:05"),
		Exists:      exists,
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

// Close closes the database connection
func (a *App) Close() error {
	return a.client.Close()
}