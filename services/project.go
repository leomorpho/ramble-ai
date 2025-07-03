package services

import (
	"context"
	"fmt"
	"path/filepath"

	"MYAPP/ent"
	"MYAPP/ent/project"
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

// ProjectService handles project-related operations
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

// CreateProject creates a new project with a default path
func (ps *ProjectService) CreateProject(name, description string) (*ProjectResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}

	// Create a default project path
	projectPath := filepath.Join("projects", name)

	project, err := ps.client.Project.
		Create().
		SetName(name).
		SetDescription(description).
		SetPath(projectPath).
		Save(ps.ctx)
	
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
func (ps *ProjectService) GetProjects() ([]*ProjectResponse, error) {
	projects, err := ps.client.Project.
		Query().
		WithVideoClips().
		All(ps.ctx)
	
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
func (ps *ProjectService) GetProjectByID(id int) (*ProjectResponse, error) {
	project, err := ps.client.Project.
		Query().
		Where(project.ID(id)).
		WithVideoClips().
		Only(ps.ctx)
	
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
func (ps *ProjectService) UpdateProject(id int, name, description string) (*ProjectResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}

	// Update the project path if name changed
	projectPath := filepath.Join("projects", name)

	updatedProject, err := ps.client.Project.
		UpdateOneID(id).
		SetName(name).
		SetDescription(description).
		SetPath(projectPath).
		Save(ps.ctx)
	
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
func (ps *ProjectService) DeleteProject(id int) error {
	err := ps.client.Project.
		DeleteOneID(id).
		Exec(ps.ctx)
	
	if err != nil {
		return fmt.Errorf("failed to delete project with ID %d: %w", id, err)
	}

	return nil
}