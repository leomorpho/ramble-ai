package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"MYAPP/ent"
	"MYAPP/ent/exportjob"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/mattn/go-sqlite3"
)

// Repository handles database operations
type Repository struct {
	client *ent.Client
	ctx    context.Context
}

// NewRepository creates a new database repository
func NewRepository() (*Repository, error) {
	// Initialize database
	db, err := sql.Open("sqlite3", "database.db?_fk=1")
	if err != nil {
		return nil, fmt.Errorf("failed opening connection to sqlite: %w", err)
	}

	// Create Ent client with proper dialect
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))

	return &Repository{
		client: client,
	}, nil
}

// Initialize sets up the repository with context and runs migrations
func (r *Repository) Initialize(ctx context.Context) error {
	r.ctx = ctx

	// Run database migrations
	if err := r.client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("failed creating schema resources: %w", err)
	}

	log.Println("Database initialized and migrations applied")
	
	// Recover any incomplete export jobs
	if err := r.recoverActiveExportJobs(); err != nil {
		log.Printf("Failed to recover active export jobs: %v", err)
	}

	return nil
}

// Close closes the database connection
func (r *Repository) Close() error {
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	return nil
}

// GetClient returns the ent client for direct access when needed
func (r *Repository) GetClient() *ent.Client {
	return r.client
}

// GetContext returns the repository context
func (r *Repository) GetContext() context.Context {
	return r.ctx
}

// recoverActiveExportJobs restores export jobs that were running when the app was closed
func (r *Repository) recoverActiveExportJobs() error {
	// Find jobs that are not complete and not cancelled
	activeJobs, err := r.client.ExportJob.
		Query().
		Where(
			exportjob.IsComplete(false),
			exportjob.IsCancelled(false),
		).
		All(r.ctx)

	if err != nil {
		return fmt.Errorf("failed to get active export jobs: %w", err)
	}

	// Mark incomplete jobs as cancelled since we can't resume them
	for _, job := range activeJobs {
		log.Printf("Marking incomplete export job %s as cancelled", job.JobID)
		_, err := r.client.ExportJob.
			UpdateOne(job).
			SetIsCancelled(true).
			SetStage("cancelled").
			SetErrorMessage("Application was restarted during export").
			SetIsComplete(true).
			SetCompletedAt(time.Now()).
			SetUpdatedAt(time.Now()).
			Save(r.ctx)
		if err != nil {
			log.Printf("Failed to cancel job %s: %v", job.JobID, err)
		}
	}

	return nil
}