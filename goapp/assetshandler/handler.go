package assetshandler

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// AssetHandler provides asset serving functionality for video files and thumbnails
type AssetHandler struct {
	// Can add configuration options here if needed
}

// NewAssetHandler creates a new asset handler service
func NewAssetHandler() *AssetHandler {
	return &AssetHandler{}
}

// CreateAssetMiddleware creates middleware for serving video files via AssetServer
// Uses gahara's approach for direct file serving with CORS support
func (h *AssetHandler) CreateAssetMiddleware() assetserver.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers for all requests (gahara approach)
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Check if this is a video request
			if strings.HasPrefix(r.URL.Path, "/api/video/") {
				h.HandleVideoRequest(w, r)
				return
			}
			// Check if this is a thumbnail request
			if strings.HasPrefix(r.URL.Path, "/api/thumbnail/") {
				h.HandleThumbnailRequest(w, r)
				return
			}
			
			// For all other requests (including direct file paths), use gahara's approach
			// Serve files directly from the filesystem with security checks
			if h.isDirectFileRequest(r.URL.Path) {
				h.HandleDirectFileRequest(w, r)
				return
			}
			
			// Pass to next handler for frontend assets
			next.ServeHTTP(w, r)
		})
	}
}

// HandleVideoRequest handles video file requests with HTTP range support
func (h *AssetHandler) HandleVideoRequest(w http.ResponseWriter, r *http.Request) {
	// Extract file path from URL
	filePath := r.URL.Path[11:] // Remove "/api/video/"
	log.Printf("[VIDEO] Raw path: %s", r.URL.Path)
	log.Printf("[VIDEO] Extracted path: %s", filePath)

	decodedPath, err := url.QueryUnescape(filePath)
	if err != nil {
		log.Printf("[VIDEO] URL decode error: %v", err)
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	log.Printf("[VIDEO] Decoded path: %s", decodedPath)

	// Security check - ensure file exists and is a video
	if !h.IsVideoFile(decodedPath) {
		http.Error(w, "Not a video file", http.StatusBadRequest)
		return
	}

	file, err := os.Open(decodedPath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "File info error", http.StatusInternalServerError)
		return
	}

	// Set content type based on file extension
	contentType := h.GetContentType(decodedPath)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")

	// Handle range requests for video seeking
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		h.HandleRangeRequest(w, r, file, fileInfo.Size(), rangeHeader)
		return
	}

	// Serve the entire file
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	io.Copy(w, file)
}

// HandleRangeRequest handles HTTP range requests for efficient video seeking
func (h *AssetHandler) HandleRangeRequest(w http.ResponseWriter, r *http.Request, file *os.File, fileSize int64, rangeHeader string) {
	// Parse range header (e.g., "bytes=0-1023")
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	rangeSpec := rangeHeader[6:] // Remove "bytes="
	parts := strings.Split(rangeSpec, "-")
	if len(parts) != 2 {
		http.Error(w, "Invalid range format", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	var start, end int64
	var err error

	// Parse start
	if parts[0] != "" {
		start, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil || start < 0 {
			http.Error(w, "Invalid start range", http.StatusRequestedRangeNotSatisfiable)
			return
		}
	}

	// Parse end
	if parts[1] != "" {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil || end >= fileSize {
			end = fileSize - 1
		}
	} else {
		end = fileSize - 1
	}

	// Validate range
	if start > end || start >= fileSize {
		http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	contentLength := end - start + 1

	// Set response headers for partial content
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	w.WriteHeader(http.StatusPartialContent)

	// Seek to start position and copy the requested range
	file.Seek(start, 0)
	io.CopyN(w, file, contentLength)
}

// GetContentType returns the appropriate MIME type for video files
func (h *AssetHandler) GetContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	case ".mkv":
		return "video/x-matroska"
	case ".webm":
		return "video/webm"
	case ".flv":
		return "video/x-flv"
	case ".wmv":
		return "video/x-ms-wmv"
	case ".m4v":
		return "video/x-m4v"
	case ".mpg", ".mpeg":
		return "video/mpeg"
	default:
		return "application/octet-stream"
	}
}

// HandleThumbnailRequest handles video thumbnail requests
func (h *AssetHandler) HandleThumbnailRequest(w http.ResponseWriter, r *http.Request) {
	// Extract file path from URL
	filePath := r.URL.Path[15:] // Remove "/api/thumbnail/"
	log.Printf("[THUMBNAIL] Raw path: %s", r.URL.Path)
	log.Printf("[THUMBNAIL] Extracted path: %s", filePath)

	decodedPath, err := url.QueryUnescape(filePath)
	if err != nil {
		log.Printf("[THUMBNAIL] URL decode error: %v", err)
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	log.Printf("[THUMBNAIL] Decoded path: %s", decodedPath)

	// Security check - ensure file exists and is a video
	if !h.IsVideoFile(decodedPath) {
		http.Error(w, "Not a video file", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(decodedPath); os.IsNotExist(err) {
		http.Error(w, "Video file not found", http.StatusNotFound)
		return
	}

	// Generate or get existing thumbnail
	thumbnailPath, err := h.GenerateThumbnail(decodedPath)
	if err != nil {
		log.Printf("[THUMBNAIL] Generation error: %v", err)
		http.Error(w, "Failed to generate thumbnail", http.StatusInternalServerError)
		return
	}

	// Serve the thumbnail file
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours
	http.ServeFile(w, r, thumbnailPath)
}

// GenerateThumbnail generates a thumbnail for the video file
func (h *AssetHandler) GenerateThumbnail(videoPath string) (string, error) {
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

	// Use ffmpeg to generate thumbnail at 10% of video duration
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-ss", "00:00:03", // Seek to 3 seconds
		"-vframes", "1", // Extract 1 frame
		"-vf", "scale=320:240:force_original_aspect_ratio=decrease,pad=320:240:(ow-iw)/2:(oh-ih)/2", // Scale to 320x240 with padding
		"-q:v", "2", // High quality
		"-y", // Overwrite output file
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

// GetThumbnailURL returns a URL for the video thumbnail
func (h *AssetHandler) GetThumbnailURL(filePath string) string {
	if !h.IsVideoFile(filePath) {
		return ""
	}

	// Encode file path for URL safety
	encodedPath := url.QueryEscape(filePath)
	return fmt.Sprintf("/api/thumbnail/%s", encodedPath)
}

// IsVideoFile checks if a file is a supported video format
func (h *AssetHandler) IsVideoFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	videoExtensions := []string{".mp4", ".mov", ".avi", ".mkv", ".wmv", ".flv", ".webm", ".m4v", ".mpg", ".mpeg"}

	for _, validExt := range videoExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

// GetFileInfo extracts file information from the filesystem
func (h *AssetHandler) GetFileInfo(filePath string) (int64, string, bool) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, "", false
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	format := strings.TrimPrefix(ext, ".")

	return fileInfo.Size(), format, true
}

// isDirectFileRequest checks if the request is for a direct file path (gahara approach)
func (h *AssetHandler) isDirectFileRequest(path string) bool {
	// Exclude frontend assets and development server paths
	if strings.HasPrefix(path, "/_app/") || 
	   strings.HasPrefix(path, "/api/") ||
	   strings.HasPrefix(path, "/@") ||
	   strings.HasPrefix(path, "/node_modules/") ||
	   strings.Contains(path, ".svelte-kit") ||
	   strings.Contains(path, "/src/") {
		log.Printf("[DIRECT] Excluding frontend asset path: %s", path)
		return false
	}
	
	// Check if this looks like an absolute file path with a video extension
	// Support ANY absolute path (Unix: starts with /, Windows: drive letter format)
	isAbsolutePath := strings.HasPrefix(path, "/") || // Unix-style absolute path
		(len(path) >= 3 && path[1] == ':' && (path[2] == '\\' || path[2] == '/')) // Windows drive letter
	
	if isAbsolutePath && len(path) > 3 {
		// Check if it has a video file extension
		ext := strings.ToLower(filepath.Ext(path))
		videoExtensions := []string{".mp4", ".mov", ".avi", ".mkv", ".wmv", ".flv", ".webm", ".m4v", ".mpg", ".mpeg"}
		for _, validExt := range videoExtensions {
			if ext == validExt {
				log.Printf("[DIRECT] Accepting direct file request: %s (extension: %s)", path, ext)
				return true
			}
		}
		log.Printf("[DIRECT] Path is absolute but not a video file: %s (extension: %s)", path, ext)
	} else {
		log.Printf("[DIRECT] Path is not absolute: %s", path)
	}
	
	log.Printf("[DIRECT] Rejecting file request: %s", path)
	return false
}

// HandleDirectFileRequest handles direct file serving (gahara approach)
func (h *AssetHandler) HandleDirectFileRequest(w http.ResponseWriter, r *http.Request) {
	// Clean the file path from URL
	filePath, err := url.QueryUnescape(r.URL.Path)
	if err != nil {
		log.Printf("[DIRECT] URL decode error: %v", err)
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}
	
	// Clean the path for security
	filePath = filepath.Clean(filePath)
	
	log.Printf("[DIRECT] Serving file: %s", filePath)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("[DIRECT] File not found: %s", err)
		http.NotFound(w, r)
		return
	}
	
	// Serve the file directly (gahara approach)
	http.ServeFile(w, r, filePath)
}