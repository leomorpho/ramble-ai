package binaries

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var (
	ffmpegPath       string
	ffmpegOnce       sync.Once
	extractionSuccess bool
	mu               sync.Mutex
)

// GetFFmpegPath returns the path to the extracted FFmpeg binary.
// It extracts the binary on first call and caches successful extractions.
// If extraction fails, it will retry on subsequent calls instead of caching the error.
func GetFFmpegPath() (string, error) {
	mu.Lock()
	defer mu.Unlock()
	
	// If we already successfully extracted, return the cached path
	if extractionSuccess && ffmpegPath != "" {
		return ffmpegPath, nil
	}
	
	// If no binary is embedded (dev/test mode), return permanent error
	if len(FFmpegBinary) == 0 {
		return "", fmt.Errorf("no embedded FFmpeg binary available (dev/test mode)")
	}
	
	// Attempt extraction (this will retry on failures)
	path, err := extractFFmpeg()
	if err != nil {
		log.Printf("FFmpeg extraction failed, will retry on next call: %v", err)
		return "", err
	}
	
	// Cache successful extraction
	ffmpegPath = path
	extractionSuccess = true
	log.Printf("FFmpeg extraction successful, cached path: %s", path)
	return path, nil
}

// extractFFmpeg extracts the embedded FFmpeg binary to a temporary file with retry logic
func extractFFmpeg() (string, error) {
	// If no binary is embedded (dev/test mode), return an error
	if len(FFmpegBinary) == 0 {
		return "", fmt.Errorf("no embedded FFmpeg binary available (dev/test mode)")
	}

	// Retry extraction with exponential backoff for temporary failures
	maxRetries := 3
	baseDelay := 100 * time.Millisecond
	
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1)) // exponential backoff: 100ms, 200ms, 400ms
			log.Printf("FFmpeg extraction attempt %d/%d failed, retrying in %v: %v", attempt, maxRetries, delay, lastErr)
			time.Sleep(delay)
		}

		// Create a temporary file with the appropriate extension
		tmpFile, err := os.CreateTemp("", "ffmpeg"+FFmpegExtension)
		if err != nil {
			lastErr = fmt.Errorf("failed to create temp file: %w", err)
			continue
		}

		// Write the embedded binary data to the temp file
		if _, err := tmpFile.Write(FFmpegBinary); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name()) // Clean up on error
			lastErr = fmt.Errorf("failed to write FFmpeg binary: %w", err)
			continue
		}

		tmpFile.Close() // Close before chmod

		// Make the file executable on Unix systems
		if runtime.GOOS != "windows" {
			if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
				os.Remove(tmpFile.Name()) // Clean up on error
				lastErr = fmt.Errorf("failed to make FFmpeg executable: %w", err)
				continue
			}
		}

		path := tmpFile.Name()
		log.Printf("FFmpeg extracted successfully to: %s (attempt %d/%d)", path, attempt+1, maxRetries)
		return path, nil
	}

	return "", fmt.Errorf("ffmpeg extraction failed after %d attempts, last error: %w", maxRetries, lastErr)
}

// CleanupFFmpeg removes the extracted FFmpeg binary (call on app shutdown)
func CleanupFFmpeg() {
	mu.Lock()
	defer mu.Unlock()
	
	if extractionSuccess && ffmpegPath != "" {
		if err := os.Remove(ffmpegPath); err != nil {
			log.Printf("Failed to cleanup FFmpeg binary: %v", err)
		} else {
			log.Printf("FFmpeg binary cleaned up: %s", ffmpegPath)
		}
		// Reset state after cleanup
		ffmpegPath = ""
		extractionSuccess = false
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

// GetEmbeddedBinarySize returns the size of the embedded FFmpeg binary
func GetEmbeddedBinarySize() int {
	return len(FFmpegBinary)
}

// GetFFmpegDebugInfo returns debug information about FFmpeg embedding
func GetFFmpegDebugInfo() map[string]interface{} {
	info := map[string]interface{}{
		"embedded_size":  GetEmbeddedBinarySize(),
		"extension":      FFmpegExtension,
		"available":      IsFFmpegAvailable(),
		"version":        GetFFmpegVersion(),
		"runtime_goos":   runtime.GOOS,
		"runtime_goarch": runtime.GOARCH,
	}
	
	if path, err := GetFFmpegPath(); err == nil {
		info["extracted_path"] = path
		if stat, err := os.Stat(path); err == nil {
			info["extracted_size"] = stat.Size()
		}
	} else {
		info["error"] = err.Error()
	}
	
	return info
}
