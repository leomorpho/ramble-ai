package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"MYAPP/ent"
	"MYAPP/ent/project"
	"MYAPP/ent/videoclip"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/mattn/go-sqlite3"
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

// CreateVideoClip creates a new video clip
func (a *App) CreateVideoClip(projectID int, name, description, filePath string) (*ent.VideoClip, error) {
	return a.client.VideoClip.
		Create().
		SetName(name).
		SetDescription(description).
		SetFilePath(filePath).
		SetProjectID(projectID).
		Save(a.ctx)
}

// GetVideoClipsByProject returns all video clips for a project
func (a *App) GetVideoClipsByProject(projectID int) ([]*ent.VideoClip, error) {
	return a.client.VideoClip.
		Query().
		Where(videoclip.HasProjectWith(project.ID(projectID))).
		All(a.ctx)
}

// Close closes the database connection
func (a *App) Close() error {
	return a.client.Close()
}
