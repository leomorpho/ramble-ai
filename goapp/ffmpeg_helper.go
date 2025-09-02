package goapp

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

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

// EnsureFFmpeg ensures FFmpeg is available by downloading it if necessary
func EnsureFFmpeg() error {
	log.Printf("[FFMPEG] Ensuring FFmpeg availability")
	
	// Check if we already have a downloaded FFmpeg
	ffmpegPath, err := getDownloadedFFmpegPath()
	if err == nil {
		// Test if the downloaded FFmpeg works
		if testFFmpegBinary(ffmpegPath) {
			log.Printf("[FFMPEG] Using downloaded FFmpeg at: %s", ffmpegPath)
			return nil
		}
		log.Printf("[FFMPEG] Downloaded FFmpeg not working, removing and re-downloading")
		os.Remove(ffmpegPath)
	}
	
	// Try system FFmpeg first
	if CheckFFmpegAvailability() == nil {
		log.Printf("[FFMPEG] Using system FFmpeg")
		return nil
	}
	
	// Download FFmpeg
	log.Printf("[FFMPEG] Neither downloaded nor system FFmpeg available, downloading...")
	if err := downloadFFmpeg(); err != nil {
		return fmt.Errorf("failed to download FFmpeg: %w", err)
	}
	
	// Test the newly downloaded FFmpeg
	ffmpegPath, err = getDownloadedFFmpegPath()
	if err != nil {
		return fmt.Errorf("failed to get downloaded FFmpeg path: %w", err)
	}
	
	if !testFFmpegBinary(ffmpegPath) {
		return fmt.Errorf("downloaded FFmpeg is not working")
	}
	
	log.Printf("[FFMPEG] Successfully downloaded and verified FFmpeg at: %s", ffmpegPath)
	return nil
}

// getDownloadedFFmpegPath returns the path where we store downloaded FFmpeg
func getDownloadedFFmpegPath() (string, error) {
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
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(userDataDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create app data directory: %w", err)
	}
	
	// FFmpeg binary name depends on OS
	binaryName := "ffmpeg"
	if runtime.GOOS == "windows" {
		binaryName = "ffmpeg.exe"
	}
	
	ffmpegPath := filepath.Join(userDataDir, binaryName)
	
	// Check if file exists
	if _, err := os.Stat(ffmpegPath); os.IsNotExist(err) {
		return "", fmt.Errorf("downloaded FFmpeg not found at %s", ffmpegPath)
	}
	
	return ffmpegPath, nil
}

// testFFmpegBinary tests if an FFmpeg binary is working
func testFFmpegBinary(path string) bool {
	_, err := exec.Command(path, "-version").CombinedOutput()
	return err == nil
}

// downloadFFmpeg downloads FFmpeg for the current platform
func downloadFFmpeg() error {
	// Map Go runtime to ffbinaries platform
	var platform string
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "darwin/amd64", "darwin/arm64":
		platform = "macos-64"
	case "linux/amd64":
		platform = "linux-64"
	case "linux/386":
		platform = "linux-32"
	case "linux/arm64":
		platform = "linux-arm64"
	case "windows/amd64":
		platform = "windows-64"
	default:
		return fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	
	// Get download URL from ffbinaries API
	downloadURL := fmt.Sprintf("https://github.com/ffbinaries/ffbinaries-prebuilt/releases/download/v6.1/ffmpeg-6.1-%s.zip", platform)
	log.Printf("[FFMPEG] Downloading from: %s", downloadURL)
	
	// Download the zip file
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download FFmpeg: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download FFmpeg: HTTP %d", resp.StatusCode)
	}
	
	// Create temporary file
	tempFile, err := os.CreateTemp("", "ffmpeg-*.zip")
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
	
	// Get destination path
	userDataDir := filepath.Dir(zipPath) // Use temp dir first
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
		if err := os.MkdirAll(userDataDir, 0755); err != nil {
			return fmt.Errorf("failed to create app data directory: %w", err)
		}
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
			
			// Create destination file
			destPath := filepath.Join(userDataDir, binaryName)
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
			
			// Remove quarantine on macOS
			if runtime.GOOS == "darwin" {
				exec.Command("xattr", "-d", "com.apple.quarantine", destPath).Run()
			}
			
			return nil
		}
	}
	
	return fmt.Errorf("FFmpeg binary not found in zip file")
}
