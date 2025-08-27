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