package goapp

import (
	"fmt"
	"log"
	"os/exec"

	"ramble-ai/binaries"
)

// GetFFmpegCommand creates an exec.Cmd for FFmpeg using the embedded binary only
func GetFFmpegCommand(args ...string) (*exec.Cmd, error) {
	// Always use the embedded FFmpeg binary - no fallback to system ffmpeg
	ffmpegPath, err := binaries.GetFFmpegPath()
	if err != nil {
		// Include debug info in error to help diagnose embedding issues
		debugInfo := binaries.GetFFmpegDebugInfo()
		log.Printf("FFmpeg embedding failed: %v, debug info: %+v", err, debugInfo)
		return nil, fmt.Errorf("embedded FFmpeg binary not available: %w", err)
	}

	log.Printf("Using embedded FFmpeg binary: %s", ffmpegPath)
	return exec.Command(ffmpegPath, args...), nil
}
