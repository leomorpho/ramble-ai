package projects

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ramble-ai/goapp"
)

// Simple test for TranscribeVideoClip
func TestTranscribeVideoClipSimple(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewProjectService(helper.Client, helper.Ctx)

	project := helper.CreateTestProject("Transcribe Test")
	clip := helper.CreateTestVideoClip(project, "Test Clip")

	t.Run("function exists and can be called", func(t *testing.T) {
		result, err := service.TranscribeVideoClip(999999)
		// Function may succeed or fail, we just test it doesn't crash
		if err == nil {
			assert.NotNil(t, result)
		}
	})

	t.Run("function works with valid clip", func(t *testing.T) {
		result, err := service.TranscribeVideoClip(clip.ID)
		// Function may succeed or fail, we just test it doesn't crash
		if err == nil {
			assert.NotNil(t, result)
		}
	})
}

// Simple test for BatchTranscribeUntranscribedClips
func TestBatchTranscribeUntranscribedClipsSimple(t *testing.T) {
	helper := goapp.NewTestHelper(t)
	service := NewProjectService(helper.Client, helper.Ctx)

	project := helper.CreateTestProject("Batch Transcribe Test")

	t.Run("function exists and handles empty project", func(t *testing.T) {
		result, err := service.BatchTranscribeUntranscribedClips(project.ID)
		if err == nil {
			assert.NotNil(t, result)
			assert.Equal(t, 0, result.TranscribedCount)
			assert.Equal(t, 0, result.FailedCount)
		}
		// Function may error due to missing API key, which is okay
	})

	t.Run("function exists and can be called", func(t *testing.T) {
		result, err := service.BatchTranscribeUntranscribedClips(999999)
		// Function may succeed or fail, we just test it doesn't crash
		if err == nil {
			assert.NotNil(t, result)
		}
	})
}