package goapp

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// Legacy GetFFmpegCommand for compatibility - now uses system FFmpeg
// This function is deprecated and will be removed once all callers are updated
func GetFFmpegCommand(args ...string) (*exec.Cmd, error) {
	log.Printf("[FFMPEG] Using system FFmpeg with args: %v", args)
	return exec.Command("ffmpeg", args...), nil
}

// getBundledFFmpegPath returns the path to the bundled FFmpeg binary in the app bundle
func getBundledFFmpegPath() string {
	// Get the path to the current executable
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("[FFMPEG] Failed to get executable path: %v", err)
		return ""
	}

	log.Printf("[FFMPEG] Executable path: %s", execPath)

	// Check if we're inside an app bundle (path contains .app/Contents/MacOS/)
	execDir := filepath.Dir(execPath)
	log.Printf("[FFMPEG] Executable directory: %s", execDir)
	log.Printf("[FFMPEG] Base of exec dir: %s", filepath.Base(execDir))
	log.Printf("[FFMPEG] Dir of exec dir: %s", filepath.Dir(execDir))
	log.Printf("[FFMPEG] Base of dir of exec dir: %s", filepath.Base(filepath.Dir(execDir)))
	log.Printf("[FFMPEG] Extension of dir of dir of exec dir: %s", filepath.Ext(filepath.Dir(filepath.Dir(execDir))))

	if filepath.Base(execDir) == "MacOS" && 
	   filepath.Base(filepath.Dir(execDir)) == "Contents" && 
	   filepath.Ext(filepath.Dir(filepath.Dir(execDir))) == ".app" {
		// We're in an app bundle - look for ffmpeg alongside the main executable
		bundledPath := filepath.Join(execDir, "ffmpeg")
		log.Printf("[FFMPEG] App bundle detected, bundled FFmpeg path: %s", bundledPath)
		return bundledPath
	}

	log.Printf("[FFMPEG] Not in an app bundle")
	return ""
}

// ExtractAudio extracts audio from video using ffmpeg-go library
func ExtractAudio(videoPath, outputPath string) error {
	log.Printf("[FFMPEG] Extracting audio from %s to %s", videoPath, outputPath)
	
	err := ffmpeg.Input(videoPath).
		Output(outputPath, ffmpeg.KwArgs{
			"vn":     "",       // No video
			"acodec": "mp3",    // MP3 codec (guaranteed Whisper support)
			"ar":     "16000",  // 16kHz sample rate (optimal for Whisper)
			"ac":     "1",      // Mono channel
			"b:a":    "24k",    // Low bitrate for space savings
			"af":     "highpass=f=80,lowpass=f=8000", // Filter frequencies outside speech range
		}).
		OverWriteOutput().
		Silent(true).
		Run()
	
	if err != nil {
		return fmt.Errorf("audio extraction failed: %w", err)
	}
	
	log.Printf("[FFMPEG] Audio extraction completed successfully")
	return nil
}

// ExtractAudioChunk extracts a specific time range from audio file
func ExtractAudioChunk(audioFile string, startTime, duration float64, outputPath string) error {
	log.Printf("[FFMPEG] Extracting audio chunk: %s [%.2fs - %.2fs] -> %s", audioFile, startTime, duration, outputPath)
	
	err := ffmpeg.Input(audioFile).
		Output(outputPath, ffmpeg.KwArgs{
			"ss":     fmt.Sprintf("%.2f", startTime),
			"t":      fmt.Sprintf("%.2f", duration),
			"acodec": "pcm_s16le",
			"ar":     "16000",
			"ac":     "1",
			"f":      "wav",
		}).
		OverWriteOutput().
		Silent(true).
		Run()
		
	if err != nil {
		return fmt.Errorf("audio chunk extraction failed: %w", err)
	}
	
	log.Printf("[FFMPEG] Audio chunk extraction completed successfully")
	return nil
}

// GenerateThumbnail generates a thumbnail image from video
func GenerateThumbnail(videoPath, outputPath string, timeOffset string) error {
	log.Printf("[FFMPEG] Generating thumbnail from %s at %s -> %s", videoPath, timeOffset, outputPath)
	
	err := ffmpeg.Input(videoPath).
		Output(outputPath, ffmpeg.KwArgs{
			"ss":      timeOffset,
			"vframes": "1",        // Extract one frame
			"vf":      "scale=320:240:force_original_aspect_ratio=decrease,pad=320:240:(ow-iw)/2:(oh-ih)/2", // Scale with padding
			"q:v":     "2",        // High quality
		}).
		OverWriteOutput().
		Silent(true).
		Run()
		
	if err != nil {
		return fmt.Errorf("thumbnail generation failed: %w", err)
	}
	
	log.Printf("[FFMPEG] Thumbnail generation completed successfully")
	return nil
}

// ExtractVideoSegment extracts a video segment with optional padding
func ExtractVideoSegment(inputPath, outputPath string, startTime, duration float64) error {
	log.Printf("[FFMPEG] Extracting video segment: %s [%.3fs + %.3fs] -> %s", inputPath, startTime, duration, outputPath)
	
	err := ffmpeg.Input(inputPath).
		Output(outputPath, ffmpeg.KwArgs{
			"ss": fmt.Sprintf("%.3f", startTime),
			"t":  fmt.Sprintf("%.3f", duration),
			"c":  "copy", // Use stream copy for speed (no re-encoding)
		}).
		OverWriteOutput().
		Silent(true).
		Run()
		
	if err != nil {
		return fmt.Errorf("video segment extraction failed: %w", err)
	}
	
	log.Printf("[FFMPEG] Video segment extraction completed successfully")
	return nil
}

// ExportVideoSegmentHighQuality extracts a video segment with high quality encoding
func ExportVideoSegmentHighQuality(inputPath, outputPath string, startTime, duration float64) error {
	log.Printf("[FFMPEG] Exporting high-quality video segment: %s [%.3fs + %.3fs] -> %s", inputPath, startTime, duration, outputPath)
	
	err := ffmpeg.Input(inputPath, ffmpeg.KwArgs{"ss": fmt.Sprintf("%.3f", startTime)}).
		Output(outputPath, ffmpeg.KwArgs{
			"t":         fmt.Sprintf("%.3f", duration),
			"c:v":       "libx264",
			"preset":    "ultrafast",
			"crf":       "18",
			"c:a":       "copy",
			"movflags":  "+faststart",
		}).
		OverWriteOutput().
		Silent(true).
		Run()
		
	if err != nil {
		return fmt.Errorf("high-quality video segment export failed: %w", err)
	}
	
	log.Printf("[FFMPEG] High-quality video segment export completed successfully")
	return nil
}

// CheckFFmpegAvailability checks if ffmpeg is available in system PATH
func CheckFFmpegAvailability() error {
	log.Printf("[FFMPEG] Checking system FFmpeg availability")
	
	// Try to run ffmpeg version command
	_, err := exec.Command("ffmpeg", "-version").CombinedOutput()
	if err != nil {
		return fmt.Errorf("FFmpeg not available in system PATH: %w", err)
	}
	
	log.Printf("[FFMPEG] System FFmpeg is available")
	return nil
}

// EnsureFFmpeg ensures FFmpeg is available by always using our app-specific version
func EnsureFFmpeg(ctx context.Context, settingsService interface{ GetFFmpegReady() (bool, error); SaveFFmpegReady(bool) error }, emitEvent func(string, ...interface{})) error {
	// Detect environment
	isCI := os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true"
	
	log.Printf("[FFMPEG] === FFmpeg Initialization Started ===")
	log.Printf("[FFMPEG] Runtime Platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	
	// Check if this is a universal binary by checking the actual binary file
	execPath, err := os.Executable()
	isUniversal := false
	if err == nil {
		// Check if the binary is universal using the 'file' command
		cmd := exec.Command("file", execPath)
		if output, err := cmd.CombinedOutput(); err == nil {
			outputStr := string(output)
			if strings.Contains(outputStr, "universal binary") {
				isUniversal = true
			}
		}
	}
	
	if isUniversal {
		log.Printf("[FFMPEG] Universal Binary: Supporting both Intel and ARM, currently running on %s", runtime.GOARCH)
	} else {
		log.Printf("[FFMPEG] Native Binary: Built and running on %s", runtime.GOARCH)
	}
	
	log.Printf("[FFMPEG] CI Environment: %v", isCI)
	log.Printf("[FFMPEG] Working Directory: %s", func() string { wd, _ := os.Getwd(); return wd }())
	
	// Check database first
	if ready, err := settingsService.GetFFmpegReady(); err == nil && ready {
		log.Printf("[FFMPEG] Database indicates FFmpeg is ready, verifying...")
		ffmpegPath, err := GetDownloadedFFmpegPath()
		if err == nil && TestFFmpegBinary(ffmpegPath) {
			log.Printf("[FFMPEG] ✅ Using existing app-specific FFmpeg at: %s", ffmpegPath)
			os.Setenv("FFMPEG_BINARY", ffmpegPath)
			return nil
		}
		// Binary missing despite DB flag - reset flag and re-download
		log.Printf("[FFMPEG] ⚠️ App-specific FFmpeg not working despite DB flag, re-downloading")
		log.Printf("[FFMPEG] Path error: %v", err)
		settingsService.SaveFFmpegReady(false)
	} else if err != nil {
		log.Printf("[FFMPEG] Database check failed: %v", err)
	}
	
	// Emit download start event
	emitEvent("ffmpeg_downloading")
	log.Printf("[FFMPEG] 📥 Starting FFmpeg download process...")
	
	if err := downloadFFmpeg(); err != nil {
		// Provide user-friendly error message with diagnostics
		var errorMsg string
		if strings.Contains(err.Error(), "unsupported platform") {
			errorMsg = fmt.Sprintf("Platform not supported: %s/%s", runtime.GOOS, runtime.GOARCH)
		} else if strings.Contains(err.Error(), "HTTP") || strings.Contains(err.Error(), "timeout") {
			errorMsg = fmt.Sprintf("Download failed (network issue): %v", err)
		} else if strings.Contains(err.Error(), "permission") || strings.Contains(err.Error(), "denied") {
			errorMsg = fmt.Sprintf("Permission error during download: %v", err)
		} else if strings.Contains(err.Error(), "disk") || strings.Contains(err.Error(), "space") {
			errorMsg = fmt.Sprintf("Insufficient disk space: %v", err)
		} else {
			errorMsg = fmt.Sprintf("Download failed: %v", err)
		}
		
		log.Printf("[FFMPEG] ❌ Download failed: %s", errorMsg)
		emitEvent("ffmpeg_error", errorMsg)
		return fmt.Errorf(errorMsg)
	}
	
	log.Printf("[FFMPEG] Download completed, verifying binary...")
	
	// Test the newly downloaded FFmpeg
	ffmpegPath, err := GetDownloadedFFmpegPath()
	if err != nil {
		errorMsg := fmt.Sprintf("failed to get downloaded FFmpeg path: %v", err)
		log.Printf("[FFMPEG] ❌ Path lookup failed: %s", errorMsg)
		emitEvent("ffmpeg_error", errorMsg)
		return fmt.Errorf(errorMsg)
	}
	
	log.Printf("[FFMPEG] Testing downloaded FFmpeg binary...")
	if !TestFFmpegBinary(ffmpegPath) {
		// Provide more detailed error information for troubleshooting
		fileInfo, statErr := os.Stat(ffmpegPath)
		var errorDetails []string
		
		if statErr != nil {
			errorDetails = append(errorDetails, fmt.Sprintf("File check failed: %v", statErr))
		} else {
			errorDetails = append(errorDetails, fmt.Sprintf("File size: %d MB", fileInfo.Size()/(1024*1024)))
			errorDetails = append(errorDetails, fmt.Sprintf("Permissions: %s", fileInfo.Mode()))
			
			// Check if file is executable
			if fileInfo.Mode().Perm()&0111 == 0 {
				errorDetails = append(errorDetails, "Binary is not executable")
			}
		}
		
		// Test execution and capture output
		cmd := exec.Command(ffmpegPath, "-version")
		output, execErr := cmd.CombinedOutput()
		if execErr != nil {
			errorDetails = append(errorDetails, fmt.Sprintf("Execution error: %v", execErr))
			
			// Check for specific sandbox/entitlement issues
			if strings.Contains(execErr.Error(), "operation not permitted") {
				errorDetails = append(errorDetails, "Sandbox restriction: App lacks permission to execute downloaded binaries")
				errorDetails = append(errorDetails, "This may require code signing with proper entitlements")
				if runtime.GOOS == "darwin" {
					errorDetails = append(errorDetails, "Required entitlement: com.apple.security.cs.disable-executable-page-protection")
				}
			} else if len(output) > 0 {
				outputStr := string(output)
				if strings.Contains(outputStr, "killed") {
					errorDetails = append(errorDetails, "Binary was killed (likely security/quarantine issue)")
				} else if strings.Contains(outputStr, "bad CPU type") {
					errorDetails = append(errorDetails, "Architecture mismatch (Intel vs ARM)")
				} else if len(outputStr) > 100 {
					errorDetails = append(errorDetails, fmt.Sprintf("Output: %s...", outputStr[:100]))
				} else {
					errorDetails = append(errorDetails, fmt.Sprintf("Output: %s", outputStr))
				}
			}
		}
		
		// Add platform and binary type information
		execPath, err := os.Executable()
		isUniversal := false
		if err == nil {
			cmd := exec.Command("file", execPath)
			if output, err := cmd.CombinedOutput(); err == nil {
				if strings.Contains(string(output), "universal binary") {
					isUniversal = true
				}
			}
		}
		
		if isUniversal {
			errorDetails = append(errorDetails, fmt.Sprintf("Universal binary running on %s/%s", runtime.GOOS, runtime.GOARCH))
		} else {
			errorDetails = append(errorDetails, fmt.Sprintf("Platform: %s/%s", runtime.GOOS, runtime.GOARCH))
		}
		
		if isCI {
			errorDetails = append(errorDetails, "Running in CI environment")
		}
		
		// Create detailed error message for toast
		errorMsg := fmt.Sprintf("FFmpeg verification failed: %s", strings.Join(errorDetails, "; "))
		
		log.Printf("[FFMPEG] ❌ Binary test failed with details: %v", errorDetails)
		emitEvent("ffmpeg_error", errorMsg)
		return fmt.Errorf(errorMsg)
	}
	
	// Mark as ready in database
	if err := settingsService.SaveFFmpegReady(true); err != nil {
		log.Printf("[FFMPEG] ⚠️ Warning: failed to save ready state to database: %v", err)
	}
	
	// Set environment variable for ffmpeg-go to use our binary
	os.Setenv("FFMPEG_BINARY", ffmpegPath)
	log.Printf("[FFMPEG] Set FFMPEG_BINARY environment variable to: %s", ffmpegPath)
	
	// Emit completion event
	emitEvent("ffmpeg_ready")
	log.Printf("[FFMPEG] ✅ Successfully configured app-specific FFmpeg")
	log.Printf("[FFMPEG] === FFmpeg Initialization Complete ===")
	return nil
}

// GetDownloadedFFmpegPath returns the path where we store our app-specific FFmpeg
func GetDownloadedFFmpegPath() (string, error) {
	// Get user data directory (same logic as in app.go)
	var userDataDir string
	
	// Check if we're in development mode
	if _, err := os.Stat("go.mod"); err == nil {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}
		userDataDir = cwd
	} else {
		// Production mode
		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user config directory: %w", err)
		}
		userDataDir = filepath.Join(userConfigDir, "RambleAI")
	}
	
	// Use dedicated binaries subdirectory to avoid conflicts
	binariesDir := filepath.Join(userDataDir, "binaries")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(binariesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create binaries directory: %w", err)
	}
	
	// FFmpeg binary name depends on OS
	binaryName := "ffmpeg"
	if runtime.GOOS == "windows" {
		binaryName = "ffmpeg.exe"
	}
	
	ffmpegPath := filepath.Join(binariesDir, binaryName)
	
	// Check if file exists
	if _, err := os.Stat(ffmpegPath); os.IsNotExist(err) {
		return "", fmt.Errorf("app-specific FFmpeg not found at %s", ffmpegPath)
	}
	
	return ffmpegPath, nil
}

// TestFFmpegBinary tests if an FFmpeg binary is working with detailed error reporting
func TestFFmpegBinary(path string) bool {
	log.Printf("[FFMPEG] Testing binary at: %s", path)
	
	// Check if file exists and get info
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Printf("[FFMPEG] Binary file check failed: %v", err)
		return false
	}
	
	log.Printf("[FFMPEG] Binary file info - Size: %d bytes, Mode: %s", fileInfo.Size(), fileInfo.Mode())
	
	// Check if file is executable
	if fileInfo.Mode().Perm()&0111 == 0 {
		log.Printf("[FFMPEG] Binary is not executable, attempting to fix permissions")
		if err := os.Chmod(path, 0755); err != nil {
			log.Printf("[FFMPEG] Failed to set executable permissions: %v", err)
			return false
		}
		log.Printf("[FFMPEG] Set executable permissions on binary")
	}
	
	// Test FFmpeg execution
	cmd := exec.Command(path, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[FFMPEG] Binary execution failed: %v", err)
		log.Printf("[FFMPEG] Command output: %s", string(output))
		
		// Check for specific macOS security and sandbox issues
		if runtime.GOOS == "darwin" {
			if strings.Contains(err.Error(), "operation not permitted") {
				log.Printf("[FFMPEG] 🚫 Sandbox restriction detected: App lacks permission to execute downloaded binaries")
				log.Printf("[FFMPEG] This requires proper entitlements and code signing")
			} else if strings.Contains(string(output), "killed") {
				log.Printf("[FFMPEG] 🔒 Possible macOS security/quarantine issue detected")
			}
		}
		
		return false
	}
	
	// Log successful execution info
	outputStr := strings.TrimSpace(string(output))
	firstLine := strings.Split(outputStr, "\n")[0]
	log.Printf("[FFMPEG] Binary test successful: %s", firstLine)
	return true
}

// downloadFFmpeg downloads FFmpeg for the current platform
func downloadFFmpeg() error {
	// Detect CI environment
	isCI := os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true"
	
	// For universal binaries, runtime.GOARCH shows the actual running architecture
	log.Printf("[FFMPEG] Starting download for runtime platform: %s/%s (CI: %v)", runtime.GOOS, runtime.GOARCH, isCI)
	
	// Check if this is a universal binary
	execPath, err := os.Executable()
	isUniversal := false
	if err == nil {
		cmd := exec.Command("file", execPath)
		if output, err := cmd.CombinedOutput(); err == nil {
			if strings.Contains(string(output), "universal binary") {
				isUniversal = true
			}
		}
	}
	
	if isUniversal {
		log.Printf("[FFMPEG] Universal binary detected, running on: %s", runtime.GOARCH)
	} else {
		log.Printf("[FFMPEG] Native binary running on: %s", runtime.GOARCH)
	}
	
	// Map Go runtime to ffbinaries platform
	var platform string
	currentPlatform := runtime.GOOS + "/" + runtime.GOARCH
	switch currentPlatform {
	case "darwin/amd64":
		platform = "macos-64"
		log.Printf("[FFMPEG] Using Intel macOS FFmpeg binary for Intel runtime")
	case "darwin/arm64":
		platform = "macos-64"  // ffbinaries doesn't have separate ARM build yet
		if isUniversal {
			log.Printf("[FFMPEG] Universal binary - using Intel FFmpeg via Rosetta for ARM runtime")
		} else {
			log.Printf("[FFMPEG] Using Intel FFmpeg binary for ARM Mac (Rosetta 2)")
		}
	case "linux/amd64":
		platform = "linux-64"
	case "linux/386":
		platform = "linux-32"
	case "linux/arm64":
		platform = "linux-arm64"
	case "windows/amd64":
		platform = "windows-64"
	default:
		return fmt.Errorf("unsupported platform: %s", currentPlatform)
	}
	
	// Get download URL from ffbinaries API
	downloadURL := fmt.Sprintf("https://github.com/ffbinaries/ffbinaries-prebuilt/releases/download/v6.1/ffmpeg-6.1-%s.zip", platform)
	log.Printf("[FFMPEG] Downloading from: %s", downloadURL)
	
	// Download the zip file with retry logic
	var resp *http.Response
	var downloadErr error
	maxRetries := 3
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("[FFMPEG] Download attempt %d/%d", attempt, maxRetries)
		
		client := &http.Client{
			Timeout: 60 * time.Second, // 60-second timeout
		}
		
		resp, downloadErr = client.Get(downloadURL)
		if downloadErr == nil && resp.StatusCode == http.StatusOK {
			log.Printf("[FFMPEG] Download successful on attempt %d", attempt)
			break
		}
		
		if resp != nil {
			resp.Body.Close()
			log.Printf("[FFMPEG] Download attempt %d failed: HTTP %d", attempt, resp.StatusCode)
		} else {
			log.Printf("[FFMPEG] Download attempt %d failed: %v", attempt, downloadErr)
		}
		
		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * 2 * time.Second
			log.Printf("[FFMPEG] Waiting %v before retry", waitTime)
			time.Sleep(waitTime)
		}
	}
	
	if downloadErr != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			defer resp.Body.Close()
			return fmt.Errorf("failed to download FFmpeg after %d attempts: HTTP %d", maxRetries, resp.StatusCode)
		}
		return fmt.Errorf("failed to download FFmpeg after %d attempts: %w", maxRetries, downloadErr)
	}
	defer resp.Body.Close()
	
	log.Printf("[FFMPEG] Download completed successfully, Content-Length: %s", resp.Header.Get("Content-Length"))
	
	// Get app data directory for temp file - same logic as getDownloadedFFmpegPath
	var userDataDir string
	if _, err := os.Stat("go.mod"); err == nil {
		// Development mode
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		userDataDir = cwd
	} else {
		// Production mode
		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get user config directory: %w", err)
		}
		userDataDir = filepath.Join(userConfigDir, "RambleAI")
	}
	
	// Create temp file in app data directory instead of system temp
	tempDir := filepath.Join(userDataDir, "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	tempFile, err := os.CreateTemp(tempDir, "ffmpeg-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	
	// Copy response to temp file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save download: %w", err)
	}
	
	// Extract FFmpeg binary
	return extractFFmpeg(tempFile.Name())
}

// extractFFmpeg extracts the FFmpeg binary from downloaded zip
func extractFFmpeg(zipPath string) error {
	// Open zip file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()
	
	// Get destination path - use same logic as getDownloadedFFmpegPath
	var userDataDir string
	if _, err := os.Stat("go.mod"); err == nil {
		// Development mode
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		userDataDir = cwd
	} else {
		// Production mode
		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get user config directory: %w", err)
		}
		userDataDir = filepath.Join(userConfigDir, "RambleAI")
	}
	
	// Use dedicated binaries subdirectory
	binariesDir := filepath.Join(userDataDir, "binaries")
	if err := os.MkdirAll(binariesDir, 0755); err != nil {
		return fmt.Errorf("failed to create binaries directory: %w", err)
	}
	
	// Find and extract FFmpeg binary
	binaryName := "ffmpeg"
	if runtime.GOOS == "windows" {
		binaryName = "ffmpeg.exe"
	}
	
	for _, file := range reader.File {
		if file.Name == binaryName {
			// Extract this file
			rc, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open file in zip: %w", err)
			}
			defer rc.Close()
			
			// Create destination file in binaries directory
			destPath := filepath.Join(binariesDir, binaryName)
			destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return fmt.Errorf("failed to create destination file: %w", err)
			}
			defer destFile.Close()
			
			// Copy file contents
			_, err = io.Copy(destFile, rc)
			if err != nil {
				return fmt.Errorf("failed to extract file: %w", err)
			}
			
			log.Printf("[FFMPEG] Extracted FFmpeg to: %s", destPath)
			
			// Set executable permissions explicitly
			if err := os.Chmod(destPath, 0755); err != nil {
				log.Printf("[FFMPEG] Warning: failed to set executable permissions: %v", err)
			}
			
			// Remove quarantine and security attributes on macOS
			if runtime.GOOS == "darwin" {
				log.Printf("[FFMPEG] Removing macOS security attributes")
				
				// Remove quarantine attribute
				cmd := exec.Command("xattr", "-d", "com.apple.quarantine", destPath)
				if output, err := cmd.CombinedOutput(); err != nil {
					log.Printf("[FFMPEG] Quarantine removal failed: %v, output: %s", err, string(output))
				} else {
					log.Printf("[FFMPEG] Successfully removed quarantine attribute")
				}
				
				// Remove all extended attributes as a more aggressive approach
				cmd = exec.Command("xattr", "-c", destPath)
				if output, err := cmd.CombinedOutput(); err != nil {
					log.Printf("[FFMPEG] Extended attributes clearing failed: %v, output: %s", err, string(output))
				} else {
					log.Printf("[FFMPEG] Cleared all extended attributes")
				}
				
				// Set additional execute permissions if in CI
				if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
					log.Printf("[FFMPEG] CI environment detected, setting additional permissions")
					if err := os.Chmod(destPath, 0755); err != nil {
						log.Printf("[FFMPEG] Failed to set CI permissions: %v", err)
					}
				}
			}
			
			return nil
		}
	}
	
	return fmt.Errorf("FFmpeg binary not found in zip file")
}
