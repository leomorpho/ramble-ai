package goapp

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"ramble-ai/binaries"
)

// GetFFmpegCommand creates an exec.Cmd for FFmpeg using bundled binary first, then embedded fallback
func GetFFmpegCommand(args ...string) (*exec.Cmd, error) {
	// First try to use bundled FFmpeg binary in app bundle (no permissions needed)
	if bundledPath := getBundledFFmpegPath(); bundledPath != "" {
		log.Printf("[FFMPEG] Found bundled path candidate: %s", bundledPath)
		if _, err := os.Stat(bundledPath); err == nil {
			log.Printf("[FFMPEG] ✅ Using bundled FFmpeg binary: %s", bundledPath)
			return exec.Command(bundledPath, args...), nil
		} else {
			log.Printf("[FFMPEG] ❌ Bundled FFmpeg not accessible: %v", err)
		}
	} else {
		log.Printf("[FFMPEG] No bundled path found")
	}

	// Fallback to embedded binary extraction (requires temp file permissions)
	log.Printf("[FFMPEG] Falling back to embedded binary extraction")
	ffmpegPath, err := binaries.GetFFmpegPath()
	if err != nil {
		// Include debug info in error to help diagnose embedding issues
		debugInfo := binaries.GetFFmpegDebugInfo()
		log.Printf("FFmpeg embedding failed: %v, debug info: %+v", err, debugInfo)
		return nil, fmt.Errorf("FFmpeg not available - bundled binary not found and embedded extraction failed: %w", err)
	}

	log.Printf("[FFMPEG] ✅ Using embedded FFmpeg binary: %s", ffmpegPath)
	return exec.Command(ffmpegPath, args...), nil
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
