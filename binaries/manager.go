package binaries

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	ffmpegPath    string
	ffmpegOnce    sync.Once
	extractionErr error
)

// GetFFmpegPath returns the path to the extracted FFmpeg binary.
// It extracts the binary on first call and caches the path for subsequent calls.
func GetFFmpegPath() (string, error) {
	ffmpegOnce.Do(func() {
		ffmpegPath, extractionErr = extractFFmpeg()
	})
	return ffmpegPath, extractionErr
}

// extractFFmpeg extracts the embedded FFmpeg binary to a temporary file
func extractFFmpeg() (string, error) {
	// Create a temporary file with the appropriate extension
	tmpFile, err := os.CreateTemp("", "ffmpeg"+FFmpegExtension)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	// Write the embedded binary data to the temp file
	if _, err := tmpFile.Write(FFmpegBinary); err != nil {
		os.Remove(tmpFile.Name()) // Clean up on error
		return "", fmt.Errorf("failed to write FFmpeg binary: %w", err)
	}

	// Make the file executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
			os.Remove(tmpFile.Name()) // Clean up on error
			return "", fmt.Errorf("failed to make FFmpeg executable: %w", err)
		}
	}

	path := tmpFile.Name()
	log.Printf("FFmpeg extracted to: %s", path)
	return path, nil
}

// CleanupFFmpeg removes the extracted FFmpeg binary (call on app shutdown)
func CleanupFFmpeg() {
	if ffmpegPath != "" {
		if err := os.Remove(ffmpegPath); err != nil {
			log.Printf("Failed to cleanup FFmpeg binary: %v", err)
		} else {
			log.Printf("FFmpeg binary cleaned up: %s", ffmpegPath)
		}
	}
}

// GetFFmpegVersion returns the version of the embedded FFmpeg binary
func GetFFmpegVersion() string {
	return "6.1" // Version of the embedded FFmpeg binaries
}

// IsFFmpegAvailable checks if FFmpeg extraction was successful
func IsFFmpegAvailable() bool {
	_, err := GetFFmpegPath()
	return err == nil
}

// GetFFmpegDir returns the directory containing the FFmpeg binary
func GetFFmpegDir() (string, error) {
	path, err := GetFFmpegPath()
	if err != nil {
		return "", err
	}
	return filepath.Dir(path), nil
}
