package goapp

import (
	"os"
	"os/exec"
)

// GetFFmpegCommand creates an exec.Cmd for FFmpeg using the extracted binary path
func GetFFmpegCommand(args ...string) *exec.Cmd {
	// First try to use the extracted FFmpeg binary from environment
	if ffmpegPath := os.Getenv("FFMPEG_PATH"); ffmpegPath != "" {
		return exec.Command(ffmpegPath, args...)
	}

	// Fallback to system FFmpeg if extraction failed
	return exec.Command("ffmpeg", args...)
}
