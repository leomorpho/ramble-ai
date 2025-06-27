package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"MYAPP/ent"
	_ "github.com/mattn/go-sqlite3"
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

	// Create Ent client
	client := ent.NewClient(ent.Driver(ent.Dialect.SQLite, db))

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

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// CreateProject creates a new project
func (a *App) CreateProject(name, description, path string) (*ent.Project, error) {
	return a.client.Project.
		Create().
		SetName(name).
		SetDescription(description).
		SetPath(path).
		Save(a.ctx)
}

// GetProjects returns all projects
func (a *App) GetProjects() ([]*ent.Project, error) {
	return a.client.Project.
		Query().
		WithVideoClips().
		All(a.ctx)
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
		Where(ent.VideoClip.HasProjectWith(ent.Project.ID(projectID))).
		All(a.ctx)
}

// Close closes the database connection
func (a *App) Close() error {
	return a.client.Close()
}
