package goapp

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

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
