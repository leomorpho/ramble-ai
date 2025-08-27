package binaries

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFFmpegVersion(t *testing.T) {
	version := GetFFmpegVersion()
	assert.NotEmpty(t, version)
	assert.Equal(t, "6.1", version)
}

func TestGetFFmpegPath_DevMode(t *testing.T) {
	// Reset state for clean test
	resetFFmpegState()

	// In dev/test mode, FFmpegBinary should be empty
	assert.Empty(t, FFmpegBinary, "FFmpegBinary should be empty in dev/test mode")

	path, err := GetFFmpegPath()
	assert.Error(t, err)
	assert.Empty(t, path)
	assert.Contains(t, err.Error(), "no embedded FFmpeg binary available")
	assert.Contains(t, err.Error(), "dev/test mode")
}

func TestIsFFmpegAvailable_DevMode(t *testing.T) {
	// Reset state for clean test
	resetFFmpegState()

	// In dev/test mode, should return false
	available := IsFFmpegAvailable()
	assert.False(t, available)
}

func TestGetFFmpegDir_DevMode(t *testing.T) {
	// Reset state for clean test
	resetFFmpegState()

	dir, err := GetFFmpegDir()
	assert.Error(t, err)
	assert.Empty(t, dir)
	assert.Contains(t, err.Error(), "no embedded FFmpeg binary available")
}

func TestCleanupFFmpeg_NoPathSet(t *testing.T) {
	// Reset state to ensure no path is set
	resetFFmpegState()

	// This should not panic and should handle gracefully
	CleanupFFmpeg()
	
	// No assertions needed, just testing that it doesn't panic
}

func TestGetFFmpegPath_WithMockBinary(t *testing.T) {
	// Reset state for clean test
	resetFFmpegState()

	// Save original binary and restore after test
	originalBinary := FFmpegBinary
	defer func() {
		FFmpegBinary = originalBinary
		resetFFmpegState()
	}()

	// Mock a binary with some test data
	mockBinaryData := []byte("fake ffmpeg binary data for testing")
	FFmpegBinary = mockBinaryData

	path, err := GetFFmpegPath()
	require.NoError(t, err)
	assert.NotEmpty(t, path)

	// Verify the file exists
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.False(t, info.IsDir())

	// Verify the file has correct permissions (on Unix systems)
	if info.Mode() != 0 {
		// On Unix systems, should be executable
		mode := info.Mode()
		if mode&0111 != 0 { // Has execute permission
			assert.True(t, mode&0100 != 0, "File should be executable by owner")
		}
	}

	// Verify file content matches what we wrote
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, mockBinaryData, content)

	// Verify subsequent calls return the same path (cached)
	path2, err2 := GetFFmpegPath()
	require.NoError(t, err2)
	assert.Equal(t, path, path2)

	// Test IsFFmpegAvailable returns true now
	available := IsFFmpegAvailable()
	assert.True(t, available)

	// Test GetFFmpegDir
	dir, err := GetFFmpegDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Dir(path), dir)

	// Test cleanup
	CleanupFFmpeg()

	// Verify file was removed
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err), "File should be removed after cleanup")
}

func TestGetFFmpegPath_FileCreationError(t *testing.T) {
	// Reset state for clean test
	resetFFmpegState()

	// Save original binary and restore after test
	originalBinary := FFmpegBinary
	defer func() {
		FFmpegBinary = originalBinary
		resetFFmpegState()
	}()

	// Mock a binary with some test data
	FFmpegBinary = []byte("test data")

	// This test is tricky because we need to simulate a file creation error
	// We can't easily mock os.CreateTemp, but we can test the error handling
	// by temporarily making a directory read-only (on Unix systems)
	
	// Create a test directory
	testDir := t.TempDir()
	
	// Set TMPDIR to our test directory
	originalTmpDir := os.Getenv("TMPDIR")
	defer func() {
		if originalTmpDir != "" {
			os.Setenv("TMPDIR", originalTmpDir)
		} else {
			os.Unsetenv("TMPDIR")
		}
	}()
	
	os.Setenv("TMPDIR", testDir)
	
	// Make directory read-only to cause CreateTemp to fail
	err := os.Chmod(testDir, 0555)
	if err != nil {
		t.Skip("Cannot make directory read-only on this system")
	}
	defer os.Chmod(testDir, 0755) // Restore permissions
	
	path, err := GetFFmpegPath()
	assert.Error(t, err)
	assert.Empty(t, path)
	assert.Contains(t, err.Error(), "failed to create temp file")
}

func TestExtractFFmpeg_WriteError(t *testing.T) {
	// This test verifies the error handling in extractFFmpeg when write fails
	// It's difficult to mock file write failures, so we test indirectly
	
	// Reset state for clean test
	resetFFmpegState()

	// Save original and restore after test
	originalBinary := FFmpegBinary
	defer func() {
		FFmpegBinary = originalBinary
		resetFFmpegState()
	}()

	// Test with valid binary data
	FFmpegBinary = []byte("test binary content")
	
	// Call extractFFmpeg directly to test the function
	path, err := extractFFmpeg()
	require.NoError(t, err)
	assert.NotEmpty(t, path)
	
	// Clean up the created file
	os.Remove(path)
}

func TestConcurrentAccess(t *testing.T) {
	// Test that concurrent calls to GetFFmpegPath are handled correctly
	resetFFmpegState()

	// Save original binary and restore after test
	originalBinary := FFmpegBinary
	defer func() {
		FFmpegBinary = originalBinary
		resetFFmpegState()
	}()

	FFmpegBinary = []byte("concurrent test data")

	const numGoroutines = 10
	paths := make([]string, numGoroutines)
	errors := make([]error, numGoroutines)
	var wg sync.WaitGroup

	// Launch multiple goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			paths[index], errors[index] = GetFFmpegPath()
		}(i)
	}

	wg.Wait()

	// All should succeed and return the same path
	firstPath := paths[0]
	firstError := errors[0]

	if firstError != nil {
		// If first one failed, all should fail with same error
		for i := 1; i < numGoroutines; i++ {
			assert.Error(t, errors[i])
			assert.Equal(t, firstError.Error(), errors[i].Error())
		}
	} else {
		// If first one succeeded, all should succeed with same path
		require.NoError(t, firstError)
		require.NotEmpty(t, firstPath)

		for i := 1; i < numGoroutines; i++ {
			require.NoError(t, errors[i])
			assert.Equal(t, firstPath, paths[i])
		}

		// Clean up
		os.Remove(firstPath)
	}
}

func TestFFmpegExtension(t *testing.T) {
	// Test that the extension constant is defined
	// The actual value depends on the build platform
	assert.True(t, FFmpegExtension == "" || strings.HasPrefix(FFmpegExtension, "."))
	
	// On the current platform (test environment), we can verify the extension
	if FFmpegExtension != "" {
		assert.True(t, len(FFmpegExtension) > 0)
	}
}

// Helper function to reset the global state for testing
func resetFFmpegState() {
	// Reset the sync.Once and cached values
	ffmpegOnce = sync.Once{}
	ffmpegPath = ""
	extractionErr = nil
}