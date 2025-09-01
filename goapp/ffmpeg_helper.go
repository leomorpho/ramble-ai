package goapp

import (
	"log"
	"os/exec"

	"ramble-ai/binaries"
)

// GetFFmpegCommand creates an exec.Cmd for FFmpeg using the extracted binary path
func GetFFmpegCommand(args ...string) *exec.Cmd {
	// First try to use the extracted FFmpeg binary directly
	if ffmpegPath, err := binaries.GetFFmpegPath(); err == nil {
		log.Printf("Using embedded FFmpeg binary: %s", ffmpegPath)
		return exec.Command(ffmpegPath, args...)
	}

	// Fallback to system FFmpeg if extraction failed
	log.Printf("FFmpeg binary not available, falling back to system ffmpeg")
	return exec.Command("ffmpeg", args...)
}
