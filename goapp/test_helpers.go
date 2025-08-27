package goapp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"ramble-ai/ent"
	"ramble-ai/ent/enttest"
	"ramble-ai/ent/migrate"
	"ramble-ai/ent/schema"
	_ "github.com/mattn/go-sqlite3"
)

// TestHelper provides shared utilities for testing goapp components
type TestHelper struct {
	Client *ent.Client
	Ctx    context.Context
	T      testing.TB
}

// NewTestHelper creates a new test helper with an in-memory database
func NewTestHelper(t testing.TB) *TestHelper {
	// Use unique database name per test to avoid sharing issues
	dbName := fmt.Sprintf("file:test_%d?mode=memory&cache=shared&_fk=1&_journal_mode=WAL&_busy_timeout=5000", time.Now().UnixNano())
	client := enttest.Open(t, "sqlite3", dbName)
	ctx := context.Background()

	// Ensure all migrations are run
	err := client.Schema.Create(ctx, migrate.WithGlobalUniqueID(true))
	require.NoError(t, err)

	// Set up cleanup for when the test completes
	t.Cleanup(func() {
		client.Close()
	})

	return &TestHelper{
		Client: client,
		Ctx:    ctx,
		T:      t,
	}
}

// CreateTestProject creates a test project with given name
func (h *TestHelper) CreateTestProject(name string) *ent.Project {
	proj, err := h.Client.Project.
		Create().
		SetName(name).
		SetDescription("Test project: " + name).
		SetPath("/test/path/" + name).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(h.Ctx)
	require.NoError(h.T, err)
	return proj
}

// CreateTestVideoClip creates a test video clip for the given project
func (h *TestHelper) CreateTestVideoClip(project *ent.Project, name string) *ent.VideoClip {
	clip, err := h.Client.VideoClip.
		Create().
		SetName(name).
		SetDescription("Test clip: " + name).
		SetFilePath("/test/video/" + name + ".mp4").
		SetFileSize(1000000).
		SetDuration(60.0).
		SetFormat("mp4").
		SetWidth(1920).
		SetHeight(1080).
		SetProject(project).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(h.Ctx)
	require.NoError(h.T, err)
	return clip
}

// CreateTestHighlight creates a test highlight on the given video clip
func (h *TestHelper) CreateTestHighlight(clip *ent.VideoClip, start, end float64) string {
	highlightID := fmt.Sprintf("h_%d", time.Now().UnixNano())

	// Get fresh clip data to ensure we have latest highlights
	freshClip, err := h.Client.VideoClip.Get(h.Ctx, clip.ID)
	require.NoError(h.T, err)

	// Get existing highlights
	existingHighlights := freshClip.Highlights

	// Add new highlight
	newHighlight := schema.Highlight{
		ID:      highlightID,
		Start:   start,
		End:     end,
		ColorID: 3, // Red
	}

	updatedHighlights := append(existingHighlights, newHighlight)

	// Update video clip with new highlights
	_, err = h.Client.VideoClip.
		UpdateOne(freshClip).
		SetHighlights(updatedHighlights).
		Save(h.Ctx)
	require.NoError(h.T, err)

	return highlightID
}

// CreateTestSetting creates a test setting
func (h *TestHelper) CreateTestSetting(key, value string) *ent.Settings {
	setting, err := h.Client.Settings.
		Create().
		SetKey(key).
		SetValue(value).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(h.Ctx)
	require.NoError(h.T, err)
	return setting
}

// MockOpenAIKey creates a mock OpenAI API key for testing
func (h *TestHelper) MockOpenAIKey() string {
	return "sk-test_mock_openai_key_for_testing_123456789abcdef"
}

// MockOpenRouterKey creates a mock OpenRouter API key for testing
func (h *TestHelper) MockOpenRouterKey() string {
	return "sk-or-test_mock_openrouter_key_for_testing_123456789abcdef"
}

// AssertProjectExists asserts that a project with the given ID exists
func (h *TestHelper) AssertProjectExists(projectID int) *ent.Project {
	project, err := h.Client.Project.Get(h.Ctx, projectID)
	require.NoError(h.T, err)
	require.NotNil(h.T, project)
	return project
}

// AssertProjectNotExists asserts that a project with the given ID does not exist
func (h *TestHelper) AssertProjectNotExists(projectID int) {
	_, err := h.Client.Project.Get(h.Ctx, projectID)
	require.Error(h.T, err)
}

// AssertSettingEquals asserts that a setting has the expected value
func (h *TestHelper) AssertSettingEquals(key, expectedValue string) {
	// Get all settings and find the one with matching key
	settings, err := h.Client.Settings.Query().All(h.Ctx)
	require.NoError(h.T, err)
	
	for _, s := range settings {
		if s.Key == key {
			require.Equal(h.T, expectedValue, s.Value)
			return
		}
	}
	
	h.T.Errorf("Setting with key '%s' not found", key)
}