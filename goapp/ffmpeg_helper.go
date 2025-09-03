package goapp

import (
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

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// Legacy GetFFmpegCommand for compatibility - now uses system FFmpeg
// This function is deprecated and will be removed once all callers are updated
func GetFFmpegCommand(args ...string) (*exec.Cmd, error) {
	log.Printf("[FFMPEG] Using system FFmpeg with args: %v", args)
	return exec.Command("ffmpeg", args...), nil
}

// Legacy functions for backward compatibility - these now redirect to system FFmpeg
func GetBundledFFmpegPath() string {
	// Redirect to system FFmpeg detection
	systemPath, err := FindSystemFFmpeg()
	if err != nil {
		return ""
	}
	return systemPath
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

// EnsureFFmpeg ensures FFmpeg is available by checking system installation
func EnsureFFmpeg(ctx context.Context, settingsService interface{}, emitEvent func(string, ...interface{})) error {
	log.Printf("[FFMPEG] === FFmpeg Initialization Started ===")
	log.Printf("[FFMPEG] Runtime Platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	
	// Check for system FFmpeg installation
	ffmpegPath, err := FindSystemFFmpeg()
	if err != nil {
		log.Printf("[FFMPEG] ❌ System FFmpeg not found: %v", err)
		emitEvent("ffmpeg_not_found", err.Error())
		return err
	}
	
	log.Printf("[FFMPEG] Testing system FFmpeg binary at: %s", ffmpegPath)
	testResult, testDetails := TestFFmpegBinaryWithDetails(ffmpegPath)
	if !testResult {
		errorMsg := fmt.Sprintf("System FFmpeg failed verification. Test details: %s", testDetails)
		log.Printf("[FFMPEG] ❌ %s", errorMsg)
		emitEvent("ffmpeg_error", errorMsg)
		return fmt.Errorf(errorMsg)
	}
	
	// Set environment variable for ffmpeg-go to use the system binary
	os.Setenv("FFMPEG_BINARY", ffmpegPath)
	log.Printf("[FFMPEG] ✅ Using system FFmpeg binary at: %s", ffmpegPath)
	
	// Emit ready event
	emitEvent("ffmpeg_ready")
	log.Printf("[FFMPEG] === FFmpeg Initialization Complete ===")
	return nil
}

// getAppDataDir returns the application data directory (sandbox-safe)
func getAppDataDir() (string, error) {
	// Check if we're in development mode by looking for go.mod file
	if _, err := os.Stat("go.mod"); err == nil {
		// In development mode, use current directory
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}
		return cwd, nil
	}

	// In production mode, use Application Support directory (sandbox-safe)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Use ~/Library/Application Support/RambleAI for production
	appDataDir := filepath.Join(homeDir, "Library", "Application Support", "RambleAI")
	if err := os.MkdirAll(appDataDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create app data directory: %w", err)
	}

	return appDataDir, nil
}

// FindSystemFFmpeg looks for FFmpeg in common system locations
func FindSystemFFmpeg() (string, error) {
	log.Printf("[FFMPEG] 🔍 Searching for system FFmpeg installation...")
	
	// First check app's own installation (sandbox-safe)
	if appDataDir, err := getAppDataDir(); err == nil {
		appFFmpeg := filepath.Join(appDataDir, "bin", "ffmpeg")
		if _, err := os.Stat(appFFmpeg); err == nil {
			log.Printf("[FFMPEG] ✅ Found FFmpeg in app directory: %s", appFFmpeg)
			return appFFmpeg, nil
		}
	}
	
	// Common FFmpeg installation locations on macOS
	homeDir, _ := os.UserHomeDir()
	commonLocations := []string{
		"/opt/homebrew/bin/ffmpeg",   // ARM64 Homebrew
		"/usr/local/bin/ffmpeg",      // Intel Homebrew / manual install
		"/Applications/FFmpeg.app/Contents/MacOS/ffmpeg", // App bundle install
	}
	
	// Add user-specific location if home directory is available
	if homeDir != "" {
		userFFmpeg := filepath.Join(homeDir, ".local", "bin", "ffmpeg")
		commonLocations = append(commonLocations, userFFmpeg)
	}
	
	// First check PATH
	if pathFFmpeg, err := exec.LookPath("ffmpeg"); err == nil {
		log.Printf("[FFMPEG] ✅ Found FFmpeg in PATH: %s", pathFFmpeg)
		return pathFFmpeg, nil
	}
	
	// Then check common locations
	for _, location := range commonLocations {
		if _, err := os.Stat(location); err == nil {
			log.Printf("[FFMPEG] ✅ Found FFmpeg at: %s", location)
			return location, nil
		}
	}
	
	log.Printf("[FFMPEG] ❌ FFmpeg not found in any system location")
	return "", fmt.Errorf("FFmpeg not found. Please install FFmpeg to continue using video processing features")
}

// InstallFFmpeg downloads and installs FFmpeg to the system
func InstallFFmpeg(ctx context.Context, emitEvent func(string, ...interface{})) error {
	log.Printf("[FFMPEG] 🔽 Starting FFmpeg installation...")
	
	// Determine architecture and download URL
	var downloadURL string
	switch runtime.GOARCH {
	case "arm64":
		downloadURL = "https://ffmpeg.martin-riedl.de/redirect/latest/macos/arm64/release/ffmpeg.zip"
		log.Printf("[FFMPEG] Installing ARM64 FFmpeg for Apple Silicon")
	case "amd64":
		downloadURL = "https://ffmpeg.martin-riedl.de/redirect/latest/macos/amd64/release/ffmpeg.zip"
		log.Printf("[FFMPEG] Installing Intel FFmpeg for x86_64")
	default:
		return fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}
	
	// Get the app's data directory (works in sandbox)
	appDataDir, err := getAppDataDir()
	if err != nil {
		return fmt.Errorf("failed to get app data directory: %w", err)
	}
	
	// Create temporary directory in app's data directory (sandbox-safe)
	tempDir := filepath.Join(appDataDir, "ffmpeg-install-temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Download FFmpeg
	emitEvent("ffmpeg_install_progress", "Downloading FFmpeg...")
	zipPath := filepath.Join(tempDir, "ffmpeg.zip")
	
	log.Printf("[FFMPEG] Downloading from: %s", downloadURL)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download FFmpeg: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download FFmpeg: HTTP %d", resp.StatusCode)
	}
	
	// Save to file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()
	
	_, err = io.Copy(zipFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save FFmpeg download: %w", err)
	}
	
	// Extract FFmpeg
	emitEvent("ffmpeg_install_progress", "Extracting FFmpeg...")
	cmd := exec.Command("unzip", "-o", zipPath, "-d", tempDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract FFmpeg: %w", err)
	}
	
	// Find extracted binary
	ffmpegBinary := filepath.Join(tempDir, "ffmpeg")
	if _, err := os.Stat(ffmpegBinary); err != nil {
		return fmt.Errorf("FFmpeg binary not found after extraction: %w", err)
	}
	
	// Install FFmpeg to app's data directory (sandbox-safe)
	emitEvent("ffmpeg_install_progress", "Installing FFmpeg...")
	
	// App data directory was already retrieved earlier for temp directory
	
	// Install to app's bin directory
	appBinDir := filepath.Join(appDataDir, "bin")
	if err := os.MkdirAll(appBinDir, 0755); err != nil {
		return fmt.Errorf("failed to create app bin directory: %w", err)
	}
	
	installPath := filepath.Join(appBinDir, "ffmpeg")
	
	// Copy FFmpeg binary
	if err := copyFile(ffmpegBinary, installPath); err != nil {
		return fmt.Errorf("failed to install FFmpeg: %w", err)
	}
	
	// Make executable
	if err := os.Chmod(installPath, 0755); err != nil {
		return fmt.Errorf("failed to make FFmpeg executable: %w", err)
	}
	
	log.Printf("[FFMPEG] ✅ FFmpeg installed successfully: %s", installPath)
	
	// Test installation
	emitEvent("ffmpeg_install_progress", "Verifying installation...")
	if !TestFFmpegBinary(installPath) {
		return fmt.Errorf("FFmpeg installation failed verification")
	}
	
	log.Printf("[FFMPEG] ✅ FFmpeg successfully installed to: %s", installPath)
	emitEvent("ffmpeg_install_complete", installPath)
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	return err
}

// TestFFmpegBinary tests if an FFmpeg binary is working with detailed error reporting (legacy)
func TestFFmpegBinary(path string) bool {
	result, _ := TestFFmpegBinaryWithDetails(path)
	return result
}

// TestFFmpegBinaryWithDetails tests if an FFmpeg binary is working and returns user-visible details
func TestFFmpegBinaryWithDetails(path string) (bool, string) {
	log.Printf("[FFMPEG] 🔍 Testing binary functionality at: %s", path)
	
	var detailsBuilder []string
	
	// Check if file exists and get info
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Printf("[FFMPEG] Binary file check failed: %v", err)
		return false, fmt.Sprintf("File not accessible: %v", err)
	}
	
	log.Printf("[FFMPEG] Binary file info - Size: %d bytes, Mode: %s", fileInfo.Size(), fileInfo.Mode())
	detailsBuilder = append(detailsBuilder, fmt.Sprintf("File found (size: %d bytes)", fileInfo.Size()))
	
	// Check if file is executable
	if fileInfo.Mode().Perm()&0111 == 0 {
		log.Printf("[FFMPEG] Binary is not executable, attempting to fix permissions")
		if err := os.Chmod(path, 0755); err != nil {
			log.Printf("[FFMPEG] Failed to set executable permissions: %v", err)
			return false, fmt.Sprintf("%s; Failed to set executable permissions: %v", strings.Join(detailsBuilder, "; "), err)
		}
		log.Printf("[FFMPEG] Set executable permissions on binary")
		detailsBuilder = append(detailsBuilder, "Fixed permissions")
	}
	
	// Test FFmpeg execution
	cmd := exec.Command(path, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[FFMPEG] Binary execution failed: %v", err)
		log.Printf("[FFMPEG] Command output: %s", string(output))
		
		errorDetail := fmt.Sprintf("Execution failed: %v", err)
		
		// Check for specific macOS security and sandbox issues
		if runtime.GOOS == "darwin" {
			if strings.Contains(err.Error(), "operation not permitted") {
				log.Printf("[FFMPEG] 🚫 Sandbox restriction detected: App lacks permission to execute downloaded binaries")
				errorDetail += "; macOS sandbox restriction (needs proper entitlements and code signing)"
			} else if strings.Contains(string(output), "killed") {
				log.Printf("[FFMPEG] 🔒 Possible macOS security/quarantine issue detected")
				errorDetail += "; possible macOS security/quarantine issue"
			}
		}
		
		return false, fmt.Sprintf("%s; %s", strings.Join(detailsBuilder, "; "), errorDetail)
	}
	
	// Log successful execution info
	outputStr := strings.TrimSpace(string(output))
	firstLine := strings.Split(outputStr, "\n")[0]
	log.Printf("[FFMPEG] Binary test successful: %s", firstLine)
	detailsBuilder = append(detailsBuilder, fmt.Sprintf("Test successful: %s", firstLine))
	
	return true, strings.Join(detailsBuilder, "; ")
}

