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
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// FFmpeg version management
const MinimumFFmpegVersion = "4.4.0"

// parseFFmpegVersion extracts version from FFmpeg -version output
func parseFFmpegVersion(output string) (string, error) {
	// FFmpeg version output typically looks like: "ffmpeg version 4.4.2 Copyright ..."
	// We want to extract "4.4.2" from this
	versionRegex := regexp.MustCompile(`ffmpeg version (\d+\.\d+(?:\.\d+)?)`)
	matches := versionRegex.FindStringSubmatch(output)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse FFmpeg version from output")
	}
	return matches[1], nil
}

// compareVersions compares two semantic versions (e.g., "4.4.2" vs "4.4.0")
// Returns true if version >= minimum
func compareVersions(version, minimum string) (bool, error) {
	versionParts := strings.Split(version, ".")
	minimumParts := strings.Split(minimum, ".")
	
	// Ensure both have at least major.minor format
	for len(versionParts) < 2 {
		versionParts = append(versionParts, "0")
	}
	for len(minimumParts) < 2 {
		minimumParts = append(minimumParts, "0")
	}
	
	// Compare major, minor, patch
	for i := 0; i < 3; i++ {
		var versionNum, minimumNum int
		var err error
		
		if i < len(versionParts) {
			versionNum, err = strconv.Atoi(versionParts[i])
			if err != nil {
				return false, fmt.Errorf("invalid version number: %s", versionParts[i])
			}
		}
		
		if i < len(minimumParts) {
			minimumNum, err = strconv.Atoi(minimumParts[i])
			if err != nil {
				return false, fmt.Errorf("invalid minimum version number: %s", minimumParts[i])
			}
		}
		
		if versionNum > minimumNum {
			return true, nil
		} else if versionNum < minimumNum {
			return false, nil
		}
		// If equal, continue to next part
	}
	
	// All parts are equal, so version meets minimum
	return true, nil
}

// getFFmpegVersion gets the version string for a specific FFmpeg binary
func getFFmpegVersion(binaryPath string) (string, error) {
	cmd := exec.Command(binaryPath, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get FFmpeg version: %w", err)
	}
	
	return parseFFmpegVersion(string(output))
}

// isFFmpegVersionCompatible checks if the given binary meets minimum version requirements
func isFFmpegVersionCompatible(binaryPath string) (bool, string, error) {
	version, err := getFFmpegVersion(binaryPath)
	if err != nil {
		return false, "", err
	}
	
	compatible, err := compareVersions(version, MinimumFFmpegVersion)
	if err != nil {
		return false, version, err
	}
	
	return compatible, version, nil
}

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

// EnsureFFmpeg ensures FFmpeg is available with version-based management and auto-download
func EnsureFFmpeg(ctx context.Context, settingsService interface{}, emitEvent func(string, ...interface{})) error {
	log.Printf("[FFMPEG] === FFmpeg Initialization Started ===")
	log.Printf("[FFMPEG] Runtime Platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	log.Printf("[FFMPEG] Minimum required version: %s", MinimumFFmpegVersion)
	
	// Step 1: Try to find compatible FFmpeg installation
	ffmpegPath, err := FindSystemFFmpeg()
	if err != nil {
		log.Printf("[FFMPEG] ❌ No compatible FFmpeg found: %v", err)
		
		// Step 2: Check if we should auto-download
		downloaded, downloadedVersion, stateErr := isFFmpegDownloaded()
		if stateErr != nil {
			log.Printf("[FFMPEG] ⚠️  Failed to check download state: %v", stateErr)
		}
		
		if downloaded {
			log.Printf("[FFMPEG] ℹ️  FFmpeg was previously downloaded (v%s) but not found", downloadedVersion)
			// The downloaded version exists in state but wasn't found by FindSystemFFmpeg
			// This could mean the file was deleted or corrupted
			log.Printf("[FFMPEG] 🔄 Downloaded FFmpeg missing, will re-download")
		}
		
		// Auto-download FFmpeg
		log.Printf("[FFMPEG] 🔄 Attempting automatic FFmpeg download...")
		emitEvent("ffmpeg_auto_download_started", "Downloading FFmpeg automatically...")
		
		if downloadErr := InstallFFmpeg(ctx, emitEvent); downloadErr != nil {
			errorMsg := fmt.Sprintf("Auto-download failed: %v", downloadErr)
			log.Printf("[FFMPEG] ❌ %s", errorMsg)
			emitEvent("ffmpeg_not_found", errorMsg)
			return fmt.Errorf("FFmpeg auto-download failed: %w", downloadErr)
		}
		
		// Try to find FFmpeg again after download
		ffmpegPath, err = FindSystemFFmpeg()
		if err != nil {
			errorMsg := fmt.Sprintf("FFmpeg still not found after download: %v", err)
			log.Printf("[FFMPEG] ❌ %s", errorMsg)
			emitEvent("ffmpeg_error", errorMsg)
			return fmt.Errorf(errorMsg)
		}
		
		log.Printf("[FFMPEG] ✅ FFmpeg auto-download completed successfully")
	}
	
	// Step 3: Verify the found/downloaded FFmpeg works
	log.Printf("[FFMPEG] Testing FFmpeg binary at: %s", ffmpegPath)
	testResult, testDetails := TestFFmpegBinaryWithDetails(ffmpegPath)
	if !testResult {
		errorMsg := fmt.Sprintf("FFmpeg failed verification. Test details: %s", testDetails)
		log.Printf("[FFMPEG] ❌ %s", errorMsg)
		emitEvent("ffmpeg_error", errorMsg)
		return fmt.Errorf(errorMsg)
	}
	
	// Step 4: Set environment and emit ready event
	os.Setenv("FFMPEG_BINARY", ffmpegPath)
	
	// Log final version info
	if version, vErr := getFFmpegVersion(ffmpegPath); vErr == nil {
		log.Printf("[FFMPEG] ✅ Using FFmpeg v%s at: %s", version, ffmpegPath)
	} else {
		log.Printf("[FFMPEG] ✅ Using FFmpeg binary at: %s", ffmpegPath)
	}
	
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

// FindSystemFFmpeg looks for FFmpeg with version-based priority
func FindSystemFFmpeg() (string, error) {
	log.Printf("[FFMPEG] 🔍 Searching for FFmpeg installation (minimum version: %s)...", MinimumFFmpegVersion)
	
	// Priority 1: Check app's own installation first (downloaded version)
	// This ensures we always use the downloaded version if it exists
	if appDataDir, err := getAppDataDir(); err == nil {
		appFFmpeg := filepath.Join(appDataDir, "bin", "ffmpeg")
		if _, err := os.Stat(appFFmpeg); err == nil {
			// Test the downloaded version
			if compatible, version, err := isFFmpegVersionCompatible(appFFmpeg); err == nil {
				if compatible {
					log.Printf("[FFMPEG] ✅ Found compatible downloaded FFmpeg v%s: %s", version, appFFmpeg)
					return appFFmpeg, nil
				} else {
					log.Printf("[FFMPEG] ⚠️  Downloaded FFmpeg v%s is below minimum version %s: %s", version, MinimumFFmpegVersion, appFFmpeg)
				}
			} else {
				log.Printf("[FFMPEG] ⚠️  Downloaded FFmpeg failed version check: %v", err)
			}
		}
	}
	
	// Priority 2: Check system locations for compatible versions
	systemLocations := []string{}
	
	// Add PATH location if available
	if pathFFmpeg, err := exec.LookPath("ffmpeg"); err == nil {
		systemLocations = append(systemLocations, pathFFmpeg)
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
	
	systemLocations = append(systemLocations, commonLocations...)
	
	// Check each system location for version compatibility
	for _, location := range systemLocations {
		if _, err := os.Stat(location); err == nil {
			if compatible, version, err := isFFmpegVersionCompatible(location); err == nil {
				if compatible {
					log.Printf("[FFMPEG] ✅ Found compatible system FFmpeg v%s: %s", version, location)
					return location, nil
				} else {
					log.Printf("[FFMPEG] ⚠️  System FFmpeg v%s is below minimum version %s: %s", version, MinimumFFmpegVersion, location)
				}
			} else {
				log.Printf("[FFMPEG] ⚠️  System FFmpeg failed version check at %s: %v", location, err)
			}
		}
	}
	
	log.Printf("[FFMPEG] ❌ No compatible FFmpeg installation found (minimum version: %s)", MinimumFFmpegVersion)
	return "", fmt.Errorf("no compatible FFmpeg found. Minimum version required: %s", MinimumFFmpegVersion)
}

// State tracking for download management
func getFFmpegStateFile() (string, error) {
	appDataDir, err := getAppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appDataDir, ".ffmpeg_state"), nil
}

// markFFmpegDownloaded creates a state file indicating FFmpeg has been downloaded
func markFFmpegDownloaded(version string) error {
	stateFile, err := getFFmpegStateFile()
	if err != nil {
		return err
	}
	
	stateContent := fmt.Sprintf("downloaded_version=%s\ndownloaded_at=%d\n", version, time.Now().Unix())
	return os.WriteFile(stateFile, []byte(stateContent), 0644)
}

// isFFmpegDownloaded checks if FFmpeg has been previously downloaded
func isFFmpegDownloaded() (bool, string, error) {
	stateFile, err := getFFmpegStateFile()
	if err != nil {
		return false, "", err
	}
	
	content, err := os.ReadFile(stateFile)
	if os.IsNotExist(err) {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	
	// Parse version from state file
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "downloaded_version=") {
			version := strings.TrimPrefix(line, "downloaded_version=")
			return true, version, nil
		}
	}
	
	return true, "", nil
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
	
	// Get version of installed FFmpeg and mark as downloaded
	if version, err := getFFmpegVersion(installPath); err == nil {
		if err := markFFmpegDownloaded(version); err != nil {
			log.Printf("[FFMPEG] ⚠️  Failed to mark FFmpeg as downloaded: %v", err)
		} else {
			log.Printf("[FFMPEG] ✅ Marked FFmpeg v%s as downloaded", version)
		}
	}
	
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

