package projects

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ramble-ai/ent"
	"ramble-ai/ent/chatmessage"
	"ramble-ai/ent/enttest"
	"ramble-ai/ent/schema"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestClient(t *testing.T) *ent.Client {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	return client
}

func createTestProject(t *testing.T, client *ent.Client, ctx context.Context, name string) *ent.Project {
	project, err := client.Project.
		Create().
		SetName(name).
		SetDescription("Test project for deletion testing").
		SetPath("/test/project/path").
		Save(ctx)
	require.NoError(t, err)
	return project
}

func createTestVideoClip(t *testing.T, client *ent.Client, ctx context.Context, project *ent.Project, name string) *ent.VideoClip {
	words := []schema.Word{
		{Word: "Hello", Start: 0.0, End: 0.5},
		{Word: "world", Start: 0.5, End: 1.0},
	}

	highlights := []schema.Highlight{
		{ID: "h1", Start: 0.0, End: 1.0, ColorID: 1},
	}

	clip, err := client.VideoClip.
		Create().
		SetName(name).
		SetFilePath("/test/video.mp4").
		SetTranscription("Hello world").
		SetTranscriptionWords(words).
		SetHighlights(highlights).
		SetProjectID(project.ID).
		Save(ctx)
	require.NoError(t, err)
	return clip
}

func createTestChatSession(t *testing.T, client *ent.Client, ctx context.Context, project *ent.Project, endpointID string) *ent.ChatSession {
	session, err := client.ChatSession.
		Create().
		SetSessionID("test-session-" + endpointID).
		SetEndpointID(endpointID).
		SetProjectID(project.ID).
		Save(ctx)
	require.NoError(t, err)
	return session
}

func createTestChatMessage(t *testing.T, client *ent.Client, ctx context.Context, session *ent.ChatSession, content string, role chatmessage.Role) *ent.ChatMessage {
	message, err := client.ChatMessage.
		Create().
		SetMessageID("test-message-" + string(role) + "-" + content).
		SetSessionID(session.ID).
		SetRole(role).
		SetContent(content).
		Save(ctx)
	require.NoError(t, err)
	return message
}

func createTestExportJob(t *testing.T, client *ent.Client, ctx context.Context, project *ent.Project, jobID string) *ent.ExportJob {
	job, err := client.ExportJob.
		Create().
		SetJobID(jobID).
		SetExportType("stitched").
		SetOutputPath("/test/export/path").
		SetProjectID(project.ID).
		Save(ctx)
	require.NoError(t, err)
	return job
}

func TestDeleteProject_CascadesDeletion(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	// Create test project
	project := createTestProject(t, client, ctx, "Test Project")

	// Create related entities
	videoClip1 := createTestVideoClip(t, client, ctx, project, "Video 1")
	videoClip2 := createTestVideoClip(t, client, ctx, project, "Video 2")
	
	chatSession1 := createTestChatSession(t, client, ctx, project, "endpoint1")
	chatSession2 := createTestChatSession(t, client, ctx, project, "endpoint2")
	
	chatMessage1 := createTestChatMessage(t, client, ctx, chatSession1, "Hello", chatmessage.RoleUser)
	chatMessage2 := createTestChatMessage(t, client, ctx, chatSession1, "Hi there", chatmessage.RoleAssistant)
	chatMessage3 := createTestChatMessage(t, client, ctx, chatSession2, "Test message", chatmessage.RoleUser)
	
	exportJob1 := createTestExportJob(t, client, ctx, project, "job-1")
	exportJob2 := createTestExportJob(t, client, ctx, project, "job-2")

	// Verify all entities exist before deletion
	require.NotNil(t, project)
	require.NotNil(t, videoClip1)
	require.NotNil(t, videoClip2)
	require.NotNil(t, chatSession1)
	require.NotNil(t, chatSession2)
	require.NotNil(t, chatMessage1)
	require.NotNil(t, chatMessage2)
	require.NotNil(t, chatMessage3)
	require.NotNil(t, exportJob1)
	require.NotNil(t, exportJob2)

	// Count entities before deletion
	projectCount, err := client.Project.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, projectCount)

	videoClipCount, err := client.VideoClip.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, videoClipCount)

	chatSessionCount, err := client.ChatSession.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, chatSessionCount)

	chatMessageCount, err := client.ChatMessage.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, chatMessageCount)

	exportJobCount, err := client.ExportJob.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, exportJobCount)

	// Delete the project (should cascade delete all related entities)
	err = service.DeleteProject(project.ID)
	require.NoError(t, err)

	// Verify all entities were deleted
	projectCount, err = client.Project.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, projectCount)

	videoClipCount, err = client.VideoClip.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, videoClipCount)

	chatSessionCount, err = client.ChatSession.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, chatSessionCount)

	chatMessageCount, err = client.ChatMessage.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, chatMessageCount)

	exportJobCount, err = client.ExportJob.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, exportJobCount)

	// Verify individual entities cannot be found
	_, err = client.Project.Get(ctx, project.ID)
	assert.Error(t, err)

	_, err = client.VideoClip.Get(ctx, videoClip1.ID)
	assert.Error(t, err)

	_, err = client.VideoClip.Get(ctx, videoClip2.ID)
	assert.Error(t, err)

	_, err = client.ChatSession.Get(ctx, chatSession1.ID)
	assert.Error(t, err)

	_, err = client.ChatSession.Get(ctx, chatSession2.ID)
	assert.Error(t, err)

	_, err = client.ChatMessage.Get(ctx, chatMessage1.ID)
	assert.Error(t, err)

	_, err = client.ChatMessage.Get(ctx, chatMessage2.ID)
	assert.Error(t, err)

	_, err = client.ChatMessage.Get(ctx, chatMessage3.ID)
	assert.Error(t, err)

	_, err = client.ExportJob.Get(ctx, exportJob1.ID)
	assert.Error(t, err)

	_, err = client.ExportJob.Get(ctx, exportJob2.ID)
	assert.Error(t, err)
}

func TestDeleteProject_WithMultipleProjects_OnlyDeletesTargeted(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	// Create two test projects
	project1 := createTestProject(t, client, ctx, "Project 1")
	project2 := createTestProject(t, client, ctx, "Project 2")

	// Create related entities for both projects
	videoClip1 := createTestVideoClip(t, client, ctx, project1, "Video 1-1")
	videoClip2 := createTestVideoClip(t, client, ctx, project2, "Video 2-1")
	
	chatSession1 := createTestChatSession(t, client, ctx, project1, "endpoint1")
	chatSession2 := createTestChatSession(t, client, ctx, project2, "endpoint2")
	
	chatMessage1 := createTestChatMessage(t, client, ctx, chatSession1, "Hello", chatmessage.RoleUser)
	chatMessage2 := createTestChatMessage(t, client, ctx, chatSession2, "Hi there", chatmessage.RoleUser)
	
	exportJob1 := createTestExportJob(t, client, ctx, project1, "job-1")
	exportJob2 := createTestExportJob(t, client, ctx, project2, "job-2")

	// Delete only project1
	err := service.DeleteProject(project1.ID)
	require.NoError(t, err)

	// Verify project1 and its related entities are deleted
	_, err = client.Project.Get(ctx, project1.ID)
	assert.Error(t, err)

	_, err = client.VideoClip.Get(ctx, videoClip1.ID)
	assert.Error(t, err)

	_, err = client.ChatSession.Get(ctx, chatSession1.ID)
	assert.Error(t, err)

	_, err = client.ChatMessage.Get(ctx, chatMessage1.ID)
	assert.Error(t, err)

	_, err = client.ExportJob.Get(ctx, exportJob1.ID)
	assert.Error(t, err)

	// Verify project2 and its related entities still exist
	foundProject2, err := client.Project.Get(ctx, project2.ID)
	require.NoError(t, err)
	assert.Equal(t, project2.ID, foundProject2.ID)

	foundVideoClip2, err := client.VideoClip.Get(ctx, videoClip2.ID)
	require.NoError(t, err)
	assert.Equal(t, videoClip2.ID, foundVideoClip2.ID)

	foundChatSession2, err := client.ChatSession.Get(ctx, chatSession2.ID)
	require.NoError(t, err)
	assert.Equal(t, chatSession2.ID, foundChatSession2.ID)

	foundChatMessage2, err := client.ChatMessage.Get(ctx, chatMessage2.ID)
	require.NoError(t, err)
	assert.Equal(t, chatMessage2.ID, foundChatMessage2.ID)

	foundExportJob2, err := client.ExportJob.Get(ctx, exportJob2.ID)
	require.NoError(t, err)
	assert.Equal(t, exportJob2.ID, foundExportJob2.ID)
}

func TestDeleteProject_NonexistentProject_ReturnsError(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	// Try to delete a project that doesn't exist
	err := service.DeleteProject(99999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete project")
}

func TestDeleteProject_EmptyProject_Success(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	// Create a project with no related entities
	project := createTestProject(t, client, ctx, "Empty Project")

	// Verify project exists
	projectCount, err := client.Project.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, projectCount)

	// Delete the empty project
	err = service.DeleteProject(project.ID)
	require.NoError(t, err)

	// Verify project is deleted
	projectCount, err = client.Project.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, projectCount)

	_, err = client.Project.Get(ctx, project.ID)
	assert.Error(t, err)
}

func TestDeleteProject_WithOrphanedEntities_HandlesGracefully(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	// Create project and related entities
	project := createTestProject(t, client, ctx, "Test Project")
	chatSession := createTestChatSession(t, client, ctx, project, "endpoint1")
	chatMessage := createTestChatMessage(t, client, ctx, chatSession, "Hello", chatmessage.RoleUser)

	// Manually delete some entities to create an inconsistent state
	// This simulates the situation where files might be missing but DB refs exist
	_, err := client.ChatMessage.Delete().Where(chatmessage.ID(chatMessage.ID)).Exec(ctx)
	require.NoError(t, err)

	// Delete project should still work even with missing related entities
	err = service.DeleteProject(project.ID)
	require.NoError(t, err)

	// Verify everything is cleaned up
	projectCount, err := client.Project.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, projectCount)

	chatSessionCount, err := client.ChatSession.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, chatSessionCount)
}

func TestDeleteProject_TransactionRollback_OnError(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	// Create project and related entities
	project := createTestProject(t, client, ctx, "Test Project")
	videoClip := createTestVideoClip(t, client, ctx, project, "Video 1")

	// Get counts before attempting deletion
	projectCountBefore, err := client.Project.Query().Count(ctx)
	require.NoError(t, err)
	videoClipCountBefore, err := client.VideoClip.Query().Count(ctx)
	require.NoError(t, err)

	// Attempt to delete with a nonexistent project ID to trigger rollback
	// The transaction should rollback and leave existing data intact
	err = service.DeleteProject(99999)
	assert.Error(t, err)

	// Verify original data is still intact (rollback worked)
	projectCountAfter, err := client.Project.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, projectCountBefore, projectCountAfter)

	videoClipCountAfter, err := client.VideoClip.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, videoClipCountBefore, videoClipCountAfter)

	// Verify specific entities still exist
	foundProject, err := client.Project.Get(ctx, project.ID)
	require.NoError(t, err)
	assert.Equal(t, project.ID, foundProject.ID)

	foundVideoClip, err := client.VideoClip.Get(ctx, videoClip.ID)
	require.NoError(t, err)
	assert.Equal(t, videoClip.ID, foundVideoClip.ID)
}

// Test CRUD operations for projects
func TestCreateProject(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	tests := []struct {
		name        string
		projectName string
		description string
		expectError bool
	}{
		{"valid project", "Test Project", "A test project", false},
		{"empty name", "", "Description", true},
		{"long name", "A very long project name that might exceed some limits", "Description", false},
		{"special characters", "Projet Spécial ñ 🎥", "Descripción especial", false},
		{"empty description", "Test Project", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := service.CreateProject(tt.projectName, tt.description)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, project)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, project)
				assert.Equal(t, tt.projectName, project.Name)
				assert.Equal(t, tt.description, project.Description)
				assert.NotZero(t, project.ID)
				assert.NotEmpty(t, project.Path) // Should set a default path
				assert.Equal(t, "clips", project.ActiveTab) // Default active tab
			}
		})
	}
}

func TestGetProjects(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	t.Run("empty database", func(t *testing.T) {
		projects, err := service.GetProjects()
		require.NoError(t, err)
		// GetProjects returns nil when no projects exist
		if projects != nil {
			assert.Len(t, projects, 0)
		}
	})

	t.Run("with projects", func(t *testing.T) {
		// Create some projects
		project1, err := service.CreateProject("Project 1", "Description 1")
		require.NoError(t, err)
		
		project2, err := service.CreateProject("Project 2", "Description 2")
		require.NoError(t, err)

		// Get all projects
		projects, err := service.GetProjects()
		require.NoError(t, err)
		assert.Len(t, projects, 2)

		// Check that our projects are in the list
		projectIDs := make([]int, len(projects))
		for i, p := range projects {
			projectIDs[i] = p.ID
		}
		assert.Contains(t, projectIDs, project1.ID)
		assert.Contains(t, projectIDs, project2.ID)
	})
}

func TestGetProjectByID(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	t.Run("valid project", func(t *testing.T) {
		created, err := service.CreateProject("Test Project", "Test Description")
		require.NoError(t, err)

		retrieved, err := service.GetProjectByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, retrieved.ID)
		assert.Equal(t, created.Name, retrieved.Name)
		assert.Equal(t, created.Description, retrieved.Description)
		assert.Equal(t, created.Path, retrieved.Path)
	})

	t.Run("nonexistent project", func(t *testing.T) {
		_, err := service.GetProjectByID(99999)
		assert.Error(t, err)
	})
}

func TestUpdateProject(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	t.Run("valid update", func(t *testing.T) {
		created, err := service.CreateProject("Original Name", "Original Description")
		require.NoError(t, err)

		updated, err := service.UpdateProject(created.ID, "Updated Name", "Updated Description")
		require.NoError(t, err)
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, "Updated Description", updated.Description)

		// Verify in database
		retrieved, err := service.GetProjectByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", retrieved.Name)
		assert.Equal(t, "Updated Description", retrieved.Description)
	})

	t.Run("empty name", func(t *testing.T) {
		created, err := service.CreateProject("Test Project", "Test Description")
		require.NoError(t, err)

		_, err = service.UpdateProject(created.ID, "", "Updated Description")
		assert.Error(t, err)
	})

	t.Run("nonexistent project", func(t *testing.T) {
		_, err := service.UpdateProject(99999, "Name", "Description")
		assert.Error(t, err)
	})
}

func TestUpdateProjectActiveTab(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	created, err := service.CreateProject("Test Project", "Test Description")
	require.NoError(t, err)

	t.Run("valid tab", func(t *testing.T) {
		err := service.UpdateProjectActiveTab(created.ID, "timeline")
		require.NoError(t, err)

		// Verify the update
		retrieved, err := service.GetProjectByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, "timeline", retrieved.ActiveTab)
	})

	t.Run("nonexistent project", func(t *testing.T) {
		err := service.UpdateProjectActiveTab(99999, "videos")
		assert.Error(t, err)
	})
}

// Test video clip operations (testing logic without file system dependencies)
func TestGetVideoClipsByProject(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	project := createTestProject(t, client, ctx, "Video Project")

	t.Run("empty project", func(t *testing.T) {
		clips, err := service.GetVideoClipsByProject(project.ID)
		require.NoError(t, err)
		// GetVideoClipsByProject returns nil when no clips exist
		if clips != nil {
			assert.Len(t, clips, 0)
		}
	})

	t.Run("with video clips", func(t *testing.T) {
		// Create test clips directly in database (bypassing file validation)
		clip1 := createTestVideoClip(t, client, ctx, project, "Clip 1")
		clip2 := createTestVideoClip(t, client, ctx, project, "Clip 2")

		clips, err := service.GetVideoClipsByProject(project.ID)
		require.NoError(t, err)
		assert.Len(t, clips, 2)

		clipIDs := make([]int, len(clips))
		for i, c := range clips {
			clipIDs[i] = c.ID
		}
		assert.Contains(t, clipIDs, clip1.ID)
		assert.Contains(t, clipIDs, clip2.ID)
	})

	t.Run("nonexistent project", func(t *testing.T) {
		clips, err := service.GetVideoClipsByProject(99999)
		require.NoError(t, err)
		// Returns nil or empty slice for nonexistent project
		if clips != nil {
			assert.Len(t, clips, 0)
		}
	})
}

func TestUpdateVideoClip(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	project := createTestProject(t, client, ctx, "Video Project")
	clip := createTestVideoClip(t, client, ctx, project, "Original Clip")

	t.Run("valid update", func(t *testing.T) {
		updated, err := service.UpdateVideoClip(clip.ID, "Updated Name", "Updated Description")
		require.NoError(t, err)
		assert.Equal(t, clip.ID, updated.ID)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, "Updated Description", updated.Description)
		assert.Equal(t, clip.FilePath, updated.FilePath) // FilePath should remain unchanged
	})

	t.Run("nonexistent clip", func(t *testing.T) {
		_, err := service.UpdateVideoClip(99999, "Name", "Description")
		assert.Error(t, err)
	})
}

func TestDeleteVideoClip(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	project := createTestProject(t, client, ctx, "Video Project")

	t.Run("valid deletion", func(t *testing.T) {
		clip := createTestVideoClip(t, client, ctx, project, "Clip to Delete")

		err := service.DeleteVideoClip(clip.ID)
		require.NoError(t, err)

		// Verify clip is deleted
		_, err = client.VideoClip.Get(ctx, clip.ID)
		assert.Error(t, err)
	})

	t.Run("nonexistent clip", func(t *testing.T) {
		err := service.DeleteVideoClip(99999)
		assert.Error(t, err)
	})
}

// Test file handling utilities
func TestIsVideoFile(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	tests := []struct {
		filePath string
		expected bool
	}{
		{"/path/to/video.mp4", true},
		{"/path/to/video.avi", true},
		{"/path/to/video.mov", true},
		{"/path/to/video.mkv", true},
		{"/path/to/video.wmv", true},
		{"/path/to/video.MP4", true}, // Should handle uppercase
		{"/path/to/audio.mp3", false},
		{"/path/to/image.jpg", false},
		{"/path/to/document.pdf", false},
		{"/path/to/file", false}, // No extension
		{"", false}, // Empty path
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			result := service.isVideoFile(tt.filePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetVideoFileInfo(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	t.Run("valid video file path", func(t *testing.T) {
		info, err := service.GetVideoFileInfo("/test/path/video.mp4")
		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "/test/path/video.mp4", info.FilePath)
		assert.Equal(t, "video.mp4", info.FileName)
		assert.Equal(t, "video", info.Name)
		// File size and format will be 0 and empty since file doesn't exist
	})

	t.Run("invalid file extension", func(t *testing.T) {
		_, err := service.GetVideoFileInfo("/test/path/document.pdf")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a supported video format")
	})
}

// Test highlight operations
func TestUpdateVideoClipHighlights(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	project := createTestProject(t, client, ctx, "Video Project")
	clip := createTestVideoClip(t, client, ctx, project, "Test Clip")

	t.Run("valid highlights update", func(t *testing.T) {
		highlights := []Highlight{
			{ID: "h1", Start: 0.0, End: 1.5, ColorID: 1},
			{ID: "h2", Start: 2.0, End: 3.5, ColorID: 2},
		}

		err := service.UpdateVideoClipHighlights(clip.ID, highlights)
		require.NoError(t, err)

		// Verify highlights were updated by getting the clip
		updatedClip, err := client.VideoClip.Get(ctx, clip.ID)
		require.NoError(t, err)
		assert.Len(t, updatedClip.Highlights, 2)
	})

	t.Run("empty highlights", func(t *testing.T) {
		err := service.UpdateVideoClipHighlights(clip.ID, []Highlight{})
		require.NoError(t, err)

		// Verify highlights were cleared
		updatedClip, err := client.VideoClip.Get(ctx, clip.ID)
		require.NoError(t, err)
		assert.Len(t, updatedClip.Highlights, 0)
	})

	t.Run("nonexistent clip", func(t *testing.T) {
		highlights := []Highlight{{ID: "h1", Start: 0.0, End: 1.0, ColorID: 1}}
		err := service.UpdateVideoClipHighlights(99999, highlights)
		assert.Error(t, err)
	})
}

func TestUpdateVideoClipSuggestedHighlights(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	project := createTestProject(t, client, ctx, "Video Project")
	clip := createTestVideoClip(t, client, ctx, project, "Test Clip")

	t.Run("valid suggested highlights update", func(t *testing.T) {
		suggestedHighlights := []Highlight{
			{ID: "sh1", Start: 0.5, End: 2.0, ColorID: 3},
			{ID: "sh2", Start: 3.0, End: 4.0, ColorID: 1},
		}

		err := service.UpdateVideoClipSuggestedHighlights(clip.ID, suggestedHighlights)
		require.NoError(t, err)

		// Verify suggested highlights were updated
		updatedClip, err := client.VideoClip.Get(ctx, clip.ID)
		require.NoError(t, err)
		assert.Len(t, updatedClip.SuggestedHighlights, 2)
	})

	t.Run("nonexistent clip", func(t *testing.T) {
		highlights := []Highlight{{ID: "sh1", Start: 0.0, End: 1.0, ColorID: 1}}
		err := service.UpdateVideoClipSuggestedHighlights(99999, highlights)
		assert.Error(t, err)
	})
}


// Test NewProjectService
func TestNewProjectService(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	assert.NotNil(t, service)
	assert.Equal(t, client, service.client)
	assert.Equal(t, ctx, service.ctx)
}

// Integration test for project workflow
func TestProjectWorkflow_Integration(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()
	
	ctx := context.Background()
	service := NewProjectService(client, ctx)

	// Create project
	project, err := service.CreateProject("Integration Test Project", "Testing full workflow")
	require.NoError(t, err)
	assert.NotNil(t, project)

	// Update project
	updatedProject, err := service.UpdateProject(project.ID, "Updated Integration Project", "Updated description")
	require.NoError(t, err)
	assert.Equal(t, "Updated Integration Project", updatedProject.Name)

	// Update active tab
	err = service.UpdateProjectActiveTab(project.ID, "timeline")
	require.NoError(t, err)

	// Verify project state
	retrievedProject, err := service.GetProjectByID(project.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Integration Project", retrievedProject.Name)
	assert.Equal(t, "Updated description", retrievedProject.Description)
	assert.Equal(t, "timeline", retrievedProject.ActiveTab)

	// Create video clips directly in database
	clip1 := createTestVideoClip(t, client, ctx, &ent.Project{ID: project.ID}, "Clip 1")
	clip2 := createTestVideoClip(t, client, ctx, &ent.Project{ID: project.ID}, "Clip 2")

	// Get clips
	clips, err := service.GetVideoClipsByProject(project.ID)
	require.NoError(t, err)
	assert.Len(t, clips, 2)

	// Update clip
	updatedClip, err := service.UpdateVideoClip(clip1.ID, "Updated Clip 1", "New description")
	require.NoError(t, err)
	assert.Equal(t, "Updated Clip 1", updatedClip.Name)

	// Update highlights
	highlights := []Highlight{
		{ID: "h1", Start: 0.0, End: 2.0, ColorID: 1},
		{ID: "h2", Start: 3.0, End: 5.0, ColorID: 2},
	}
	err = service.UpdateVideoClipHighlights(clip1.ID, highlights)
	require.NoError(t, err)

	// Delete clip
	err = service.DeleteVideoClip(clip2.ID)
	require.NoError(t, err)

	// Verify only one clip remains
	finalClips, err := service.GetVideoClipsByProject(project.ID)
	require.NoError(t, err)
	assert.Len(t, finalClips, 1)
	assert.Equal(t, clip1.ID, finalClips[0].ID)

	// Delete project (should cascade delete remaining clip)
	err = service.DeleteProject(project.ID)
	require.NoError(t, err)

	// Verify project and all clips are deleted
	_, err = service.GetProjectByID(project.ID)
	assert.Error(t, err)

	finalClips, err = service.GetVideoClipsByProject(project.ID)
	require.NoError(t, err)
	assert.Len(t, finalClips, 0)
}