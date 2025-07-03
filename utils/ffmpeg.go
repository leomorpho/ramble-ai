package utils

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FFmpegProgress represents progress information from FFmpeg
type FFmpegProgress struct {
	Frame    int64
	FPS      float64
	Bitrate  string
	Time     float64
	Duration float64
	Progress float64
}

// ProgressCallback is the type for progress callback functions
type ProgressCallback func(progress float64, status string)

// SegmentInfo represents a video segment for extraction
type SegmentInfo struct {
	FilePath  string
	StartTime float64
	EndTime   float64
}

// IsVideoFile checks if the given file path is a video file
func IsVideoFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	videoExtensions := []string{".mp4", ".mov", ".avi", ".mkv", ".wmv", ".flv", ".webm", ".m4v", ".mpg", ".mpeg"}
	
	for _, validExt := range videoExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

// GenerateThumbnail generates a thumbnail for a video file using FFmpeg
func GenerateThumbnail(videoPath string) (string, error) {
	// Create thumbnails directory if it doesn't exist
	thumbnailsDir := "thumbnails"
	if err := os.MkdirAll(thumbnailsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create thumbnails directory: %w", err)
	}

	// Generate unique filename based on video path hash
	hash := md5.Sum([]byte(videoPath))
	thumbnailFilename := hex.EncodeToString(hash[:]) + ".jpg"
	thumbnailPath := filepath.Join(thumbnailsDir, thumbnailFilename)

	// Check if thumbnail already exists
	if _, err := os.Stat(thumbnailPath); err == nil {
		log.Printf("[THUMBNAIL] Using existing thumbnail: %s", thumbnailPath)
		return thumbnailPath, nil
	}

	log.Printf("[THUMBNAIL] Generating new thumbnail for: %s", videoPath)

	// Use ffmpeg to generate thumbnail at 3 seconds
	cmd := exec.Command("ffmpeg", 
		"-i", videoPath,
		"-ss", "00:00:03", // Seek to 3 seconds
		"-vframes", "1",   // Extract 1 frame
		"-vf", "scale=320:240:force_original_aspect_ratio=decrease,pad=320:240:(ow-iw)/2:(oh-ih)/2", // Scale to 320x240 with padding
		"-q:v", "2",       // High quality
		"-y",              // Overwrite output file
		thumbnailPath,
	)

	// Run ffmpeg command
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[THUMBNAIL] ffmpeg error: %v, output: %s", err, string(output))
		return "", fmt.Errorf("ffmpeg failed: %w", err)
	}

	log.Printf("[THUMBNAIL] Successfully generated: %s", thumbnailPath)
	return thumbnailPath, nil
}

// ExtractAudio extracts audio from a video file using ffmpeg
func ExtractAudio(videoPath string) (string, error) {
	// Create temp directory for audio files
	tempDir := "temp_audio"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate unique audio filename
	hash := md5.Sum([]byte(videoPath + fmt.Sprintf("%d", time.Now().UnixNano())))
	audioFilename := hex.EncodeToString(hash[:]) + ".mp3"
	audioPath := filepath.Join(tempDir, audioFilename)

	log.Printf("[TRANSCRIPTION] Extracting audio from: %s to: %s", videoPath, audioPath)

	// Use ffmpeg to extract audio
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vn",                    // No video
		"-acodec", "mp3",         // Audio codec
		"-ar", "16000",           // Sample rate (16kHz for Whisper)
		"-ac", "1",               // Mono channel
		"-b:a", "64k",            // Bitrate
		"-y",                     // Overwrite output file
		audioPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[TRANSCRIPTION] ffmpeg error: %v, output: %s", err, string(output))
		return "", fmt.Errorf("ffmpeg failed: %w", err)
	}

	log.Printf("[TRANSCRIPTION] Audio extracted successfully: %s", audioPath)
	return audioPath, nil
}

// GetVideoDuration gets the duration of a video file using ffprobe
func GetVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		videoPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get video duration: %w", err)
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

// ExtractSegment extracts a single video segment to a temp file
func ExtractSegment(segment SegmentInfo, tempDir string, index int) (string, error) {
	outputPath := filepath.Join(tempDir, fmt.Sprintf("segment_%03d.mp4", index))
	
	// Use ffmpeg to extract the segment
	cmd := exec.Command("ffmpeg",
		"-i", segment.FilePath,
		"-ss", fmt.Sprintf("%.3f", segment.StartTime),
		"-to", fmt.Sprintf("%.3f", segment.EndTime),
		"-c:v", "libx264",
		"-c:a", "aac",
		"-y",
		outputPath,
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w, output: %s", err, string(output))
	}
	
	return outputPath, nil
}

// ExtractSegmentDirect extracts a video segment directly to the output file
func ExtractSegmentDirect(segment SegmentInfo, outputPath string) error {
	// Use ffmpeg to extract the segment
	cmd := exec.Command("ffmpeg",
		"-i", segment.FilePath,
		"-ss", fmt.Sprintf("%.3f", segment.StartTime),
		"-to", fmt.Sprintf("%.3f", segment.EndTime),
		"-c:v", "libx264",
		"-c:a", "aac",
		"-y",
		outputPath,
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w, output: %s", err, string(output))
	}
	
	return nil
}

// StitchVideoClips combines multiple video clips into one
func StitchVideoClips(clipPaths []string, outputPath string) error {
	if len(clipPaths) == 0 {
		return fmt.Errorf("no clips to stitch")
	}
	
	// Create concat file for ffmpeg
	concatFile := filepath.Join(filepath.Dir(outputPath), "concat_list.txt")
	defer os.Remove(concatFile)
	
	file, err := os.Create(concatFile)
	if err != nil {
		return fmt.Errorf("failed to create concat file: %w", err)
	}
	defer file.Close()
	
	for _, clipPath := range clipPaths {
		_, err := file.WriteString(fmt.Sprintf("file '%s'\n", clipPath))
		if err != nil {
			return fmt.Errorf("failed to write to concat file: %w", err)
		}
	}
	
	// Use ffmpeg to concatenate clips
	cmd := exec.Command("ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", concatFile,
		"-c", "copy",
		"-y",
		outputPath,
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg concat failed: %w, output: %s", err, string(output))
	}
	
	return nil
}

// ParseFFmpegProgress parses FFmpeg progress output
func ParseFFmpegProgress(line string) *FFmpegProgress {
	// FFmpeg progress line format: frame=   123 fps= 12 q=28.0 size=    1234kB time=00:01:23.45 bitrate= 567.8kbits/s speed=1.23x
	frameRegex := regexp.MustCompile(`frame=\s*(\d+)`)
	fpsRegex := regexp.MustCompile(`fps=\s*([\d.]+)`)
	timeRegex := regexp.MustCompile(`time=(\d{2}):(\d{2}):([\d.]+)`)
	bitrateRegex := regexp.MustCompile(`bitrate=\s*([\d.]+)kbits/s`)

	progress := &FFmpegProgress{}

	if match := frameRegex.FindStringSubmatch(line); len(match) > 1 {
		if frame, err := strconv.ParseInt(match[1], 10, 64); err == nil {
			progress.Frame = frame
		}
	}

	if match := fpsRegex.FindStringSubmatch(line); len(match) > 1 {
		if fps, err := strconv.ParseFloat(match[1], 64); err == nil {
			progress.FPS = fps
		}
	}

	if match := timeRegex.FindStringSubmatch(line); len(match) > 3 {
		hours, _ := strconv.ParseFloat(match[1], 64)
		minutes, _ := strconv.ParseFloat(match[2], 64)
		seconds, _ := strconv.ParseFloat(match[3], 64)
		progress.Time = hours*3600 + minutes*60 + seconds
	}

	if match := bitrateRegex.FindStringSubmatch(line); len(match) > 1 {
		progress.Bitrate = match[1] + "kbits/s"
	}

	return progress
}

// ExtractSegmentWithProgress extracts a video segment with progress tracking
func ExtractSegmentWithProgress(segment SegmentInfo, tempDir string, index int, progressCallback ProgressCallback, cancel chan bool) (string, error) {
	outputPath := filepath.Join(tempDir, fmt.Sprintf("segment_%03d.mp4", index))

	// Get video duration for the highlight segment
	duration := segment.EndTime - segment.StartTime

	cmd := exec.Command("ffmpeg",
		"-i", segment.FilePath,
		"-ss", fmt.Sprintf("%.3f", segment.StartTime),
		"-to", fmt.Sprintf("%.3f", segment.EndTime),
		"-c:v", "libx264",
		"-c:a", "aac",
		"-progress", "pipe:1",
		"-y",
		outputPath,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Monitor progress
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "time=") {
				if progress := ParseFFmpegProgress(line); progress.Time > 0 && duration > 0 {
					clipProgress := progress.Time / duration
					if clipProgress > 1.0 {
						clipProgress = 1.0
					}
					if progressCallback != nil {
						progressCallback(clipProgress, "extracting")
					}
				}
			}
		}
	}()

	// Wait for completion or cancellation
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-cancel:
		cmd.Process.Kill()
		return "", fmt.Errorf("extraction cancelled")
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("ffmpeg failed: %w", err)
		}
	}

	return outputPath, nil
}

// ExtractSegmentDirectWithProgress extracts a video segment directly with progress tracking
func ExtractSegmentDirectWithProgress(segment SegmentInfo, outputPath string, progressCallback ProgressCallback, cancel chan bool) error {
	duration := segment.EndTime - segment.StartTime

	cmd := exec.Command("ffmpeg",
		"-i", segment.FilePath,
		"-ss", fmt.Sprintf("%.3f", segment.StartTime),
		"-to", fmt.Sprintf("%.3f", segment.EndTime),
		"-c:v", "libx264",
		"-c:a", "aac",
		"-progress", "pipe:1",
		"-y",
		outputPath,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Monitor progress
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "time=") {
				if progress := ParseFFmpegProgress(line); progress.Time > 0 && duration > 0 {
					clipProgress := progress.Time / duration
					if clipProgress > 1.0 {
						clipProgress = 1.0
					}
					if progressCallback != nil {
						progressCallback(clipProgress, "extracting")
					}
				}
			}
		}
	}()

	// Wait for completion or cancellation
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-cancel:
		cmd.Process.Kill()
		return fmt.Errorf("extraction cancelled")
	case err := <-done:
		if err != nil {
			return fmt.Errorf("ffmpeg failed: %w", err)
		}
	}

	return nil
}

// StitchVideoClipsWithProgress combines multiple video clips with progress tracking
func StitchVideoClipsWithProgress(clipPaths []string, outputPath string, progressCallback ProgressCallback, cancel chan bool) error {
	if len(clipPaths) == 0 {
		return fmt.Errorf("no clips to stitch")
	}

	// Calculate total duration for progress tracking
	var totalDuration float64
	for _, clipPath := range clipPaths {
		if duration, err := GetVideoDuration(clipPath); err == nil {
			totalDuration += duration
		}
	}

	// Create concat file for ffmpeg
	concatFile := filepath.Join(filepath.Dir(outputPath), "concat_list.txt")
	defer os.Remove(concatFile)

	file, err := os.Create(concatFile)
	if err != nil {
		return fmt.Errorf("failed to create concat file: %w", err)
	}
	defer file.Close()

	for _, clipPath := range clipPaths {
		_, err := file.WriteString(fmt.Sprintf("file '%s'\n", clipPath))
		if err != nil {
			return fmt.Errorf("failed to write to concat file: %w", err)
		}
	}

	cmd := exec.Command("ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", concatFile,
		"-c", "copy",
		"-progress", "pipe:1",
		"-y",
		outputPath,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Monitor progress
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "time=") {
				if progress := ParseFFmpegProgress(line); progress.Time > 0 && totalDuration > 0 {
					stitchProgress := progress.Time / totalDuration
					if stitchProgress > 1.0 {
						stitchProgress = 1.0
					}
					if progressCallback != nil {
						progressCallback(stitchProgress, "stitching")
					}
				}
			}
		}
	}()

	// Wait for completion or cancellation
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-cancel:
		cmd.Process.Kill()
		return fmt.Errorf("stitching cancelled")
	case err := <-done:
		if err != nil {
			return fmt.Errorf("ffmpeg concat failed: %w", err)
		}
	}

	return nil
}