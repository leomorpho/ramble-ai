package goapp

import (
	"context"
	"fmt"
	"log"
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

// getBundledFFmpegPath returns the path to the bundled FFmpeg binary in the app bundle
func GetBundledFFmpegPath() string {
	// Get the path to the current executable
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("[FFMPEG] Failed to get executable path: %v", err)
		return ""
	}
	
	log.Printf("[FFMPEG] Executable path: %s", execPath)

	// Detect platform and get platform-specific bundled paths
	switch runtime.GOOS {
	case "darwin":
		return getBundledFFmpegPathMacOS(execPath)
	case "windows":
		// TODO: Implement Windows bundled FFmpeg detection
		log.Printf("[FFMPEG] Windows bundled FFmpeg support not yet implemented")
		return ""
	case "linux":
		// TODO: Implement Linux bundled FFmpeg detection  
		log.Printf("[FFMPEG] Linux bundled FFmpeg support not yet implemented")
		return ""
	default:
		log.Printf("[FFMPEG] Unsupported platform for bundled FFmpeg: %s", runtime.GOOS)
		return ""
	}
}

func getBundledFFmpegPathMacOS(execPath string) string {
	// Check if we're inside an app bundle (path contains .app/Contents/MacOS/)
	if !strings.Contains(execPath, ".app/Contents/MacOS/") {
		// Not in app bundle - check for development build
		if wd, err := os.Getwd(); err == nil {
			devFFmpegPath := filepath.Join(wd, "build", "bin", "RambleAI.app", "Contents", "Resources", "binaries", "ffmpeg")
			log.Printf("[FFMPEG] Looking for dev FFmpeg at: %s", devFFmpegPath)
			if _, err := os.Stat(devFFmpegPath); err == nil {
				log.Printf("[FFMPEG] ✅ Found development FFmpeg binary")
				return devFFmpegPath
			}
		}
		log.Printf("[FFMPEG] Not in app bundle and no dev build found")
		return ""
	}

	// We're in an app bundle - try multiple possible locations
	var possiblePaths []string
	
	// Method 1: Direct path extraction (works for non-sandboxed apps)
	if parts := strings.Split(execPath, ".app/Contents/MacOS/"); len(parts) >= 2 {
		appContentsDir := parts[0] + ".app/Contents"
		possiblePaths = append(possiblePaths, filepath.Join(appContentsDir, "Resources", "binaries", "ffmpeg"))
	}
	
	// Method 2: Common installation locations (for sandboxed apps)
	possiblePaths = append(possiblePaths, 
		"/Applications/RambleAI.app/Contents/Resources/binaries/ffmpeg",
		filepath.Join(os.Getenv("HOME"), "Applications", "RambleAI.app", "Contents", "Resources", "binaries", "ffmpeg"),
	)
	
	// Method 3: Development/testing location
	possiblePaths = append(possiblePaths, 
		"/Users/leoaudibert/Workspace/ramble-ai/build/bin/RambleAI.app/Contents/Resources/binaries/ffmpeg",
	)
	if wd, err := os.Getwd(); err == nil {
		buildPath := filepath.Join(wd, "build", "bin", "RambleAI.app", "Contents", "Resources", "binaries", "ffmpeg")
		possiblePaths = append(possiblePaths, buildPath)
	}
	
	// Try each possible path
	for _, bundledFFmpegPath := range possiblePaths {
		log.Printf("[FFMPEG] Looking for bundled FFmpeg at: %s", bundledFFmpegPath)
		if _, err := os.Stat(bundledFFmpegPath); err == nil {
			log.Printf("[FFMPEG] ✅ Found bundled FFmpeg binary")
			return bundledFFmpegPath
		} else {
			log.Printf("[FFMPEG] Bundled FFmpeg not found at this path: %v", err)
		}
	}
	
	log.Printf("[FFMPEG] No bundled FFmpeg found in any macOS location")
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

// EnsureFFmpeg ensures FFmpeg is available by using the bundled version
func EnsureFFmpeg(ctx context.Context, settingsService interface{}, emitEvent func(string, ...interface{})) error {
	log.Printf("[FFMPEG] === FFmpeg Initialization Started ===")
	log.Printf("[FFMPEG] Runtime Platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	
	// Check for bundled FFmpeg binary in the app bundle
	bundledPath := GetBundledFFmpegPath()
	if bundledPath == "" {
		errorMsg := "FFmpeg not found in app bundle. Please reinstall the application."
		log.Printf("[FFMPEG] ❌ %s", errorMsg)
		emitEvent("ffmpeg_error", errorMsg)
		return fmt.Errorf(errorMsg)
	}
	
	log.Printf("[FFMPEG] Testing bundled FFmpeg binary at: %s", bundledPath)
	if !TestFFmpegBinary(bundledPath) {
		errorMsg := "Bundled FFmpeg binary failed verification. Please reinstall the application."
		log.Printf("[FFMPEG] ❌ %s", errorMsg)
		emitEvent("ffmpeg_error", errorMsg)
		return fmt.Errorf(errorMsg)
	}
	
	// Set environment variable for ffmpeg-go to use our bundled binary
	os.Setenv("FFMPEG_BINARY", bundledPath)
	log.Printf("[FFMPEG] ✅ Using bundled FFmpeg binary at: %s", bundledPath)
	
	// Emit ready event immediately - no async operations
	emitEvent("ffmpeg_ready")
	log.Printf("[FFMPEG] === FFmpeg Initialization Complete ===")
	return nil
}

// Download-related functions removed - FFmpeg is now bundled in the app

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

