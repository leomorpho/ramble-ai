package assetshandler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAssetHandler(t *testing.T) {
	handler := NewAssetHandler()
	assert.NotNil(t, handler)
}

func TestIsVideoFile(t *testing.T) {
	handler := NewAssetHandler()

	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{"mp4 file", "/path/to/video.mp4", true},
		{"avi file", "/path/to/video.avi", true},
		{"mov file", "/path/to/video.mov", true},
		{"mkv file", "/path/to/video.mkv", true},
		{"wmv file", "/path/to/video.wmv", true},
		{"MP4 uppercase", "/path/to/VIDEO.MP4", true},
		{"webm file", "/path/to/video.webm", true},
		{"txt file", "/path/to/document.txt", false},
		{"jpg file", "/path/to/image.jpg", false},
		{"mp3 file", "/path/to/audio.mp3", false},
		{"no extension", "/path/to/file", false},
		{"empty path", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.IsVideoFile(tt.filePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetContentType(t *testing.T) {
	handler := NewAssetHandler()

	tests := []struct {
		name        string
		filePath    string
		expectedCT  string
	}{
		{"mp4 file", "/path/to/video.mp4", "video/mp4"},
		{"avi file", "/path/to/video.avi", "video/x-msvideo"},
		{"mov file", "/path/to/video.mov", "video/quicktime"},
		{"mkv file", "/path/to/video.mkv", "video/x-matroska"},
		{"wmv file", "/path/to/video.wmv", "video/x-ms-wmv"},
		{"webm file", "/path/to/video.webm", "video/webm"},
		{"MP4 uppercase", "/path/to/VIDEO.MP4", "video/mp4"},
		{"unknown extension", "/path/to/file.xyz", "application/octet-stream"},
		{"no extension", "/path/to/file", "application/octet-stream"},
		{"empty path", "", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.GetContentType(tt.filePath)
			assert.Equal(t, tt.expectedCT, result)
		})
	}
}

func TestGetThumbnailURL(t *testing.T) {
	handler := NewAssetHandler()

	tests := []struct {
		name        string
		filePath    string
		expectedURL string
	}{
		{
			"basic video file", 
			"/path/to/video.mp4",
			"/api/thumbnail/%2Fpath%2Fto%2Fvideo.mp4", // URL encoded path
		},
		{
			"video with spaces",
			"/path/with spaces/video.mp4",
			"/api/thumbnail/%2Fpath%2Fwith+spaces%2Fvideo.mp4",
		},
		{
			"empty path (not video)",
			"",
			"", // Returns empty for non-video files
		},
		{
			"non-video file",
			"/path/to/document.txt",
			"", // Returns empty for non-video files
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.GetThumbnailURL(tt.filePath)
			assert.Equal(t, tt.expectedURL, result)
		})
	}
}

func TestIsDirectFileRequest(t *testing.T) {
	handler := NewAssetHandler()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"direct file request", "/api/files/direct/test.mp4", false}, // Excluded by /api/ prefix
		{"another direct file", "/api/files/direct/video/clip.avi", false}, // Excluded by /api/ prefix
		{"absolute video file", "/home/user/video.mp4", true},
		{"windows video file", "C:/videos/test.avi", true},
		{"video API request", "/api/video/123", false},
		{"thumbnail request", "/api/thumbnail/abc123", false},
		{"root path", "/", false},
		{"empty path", "", false},
		{"other API", "/api/projects", false},
		{"non-video absolute path", "/home/user/document.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.isDirectFileRequest(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFileInfo(t *testing.T) {
	handler := NewAssetHandler()

	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.mp4")
	testContent := []byte("fake video content for testing")
	err := os.WriteFile(testFile, testContent, 0644)
	require.NoError(t, err)

	t.Run("existing file", func(t *testing.T) {
		size, format, exists := handler.GetFileInfo(testFile)
		assert.True(t, exists)
		assert.Equal(t, int64(len(testContent)), size)
		assert.Equal(t, "mp4", format) // Returns format without "video/" prefix
	})

	t.Run("non-existent file", func(t *testing.T) {
		nonExistentFile := filepath.Join(tmpDir, "nonexistent.mp4")
		size, format, exists := handler.GetFileInfo(nonExistentFile)
		assert.False(t, exists)
		assert.Equal(t, int64(0), size)
		assert.Equal(t, "", format)
	})

	t.Run("empty path", func(t *testing.T) {
		size, format, exists := handler.GetFileInfo("")
		assert.False(t, exists)
		assert.Equal(t, int64(0), size)
		assert.Equal(t, "", format)
	})
}

func TestCreateAssetMiddleware(t *testing.T) {
	handler := NewAssetHandler()
	middleware := handler.CreateAssetMiddleware()

	// Create a mock next handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("next handler called"))
	})

	wrappedHandler := middleware(nextHandler)

	t.Run("sets CORS headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		// Check CORS headers are set
		assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", rr.Header().Get("Access-Control-Allow-Headers"))
	})

	t.Run("handles OPTIONS preflight", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		// Should not call next handler for OPTIONS
		assert.Empty(t, rr.Body.String())
	})

	t.Run("calls next handler for non-OPTIONS", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		// Should call next handler
		assert.Equal(t, "next handler called", rr.Body.String())
	})

	t.Run("handles video requests", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/video/test.mp4", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		// Should handle video request (will fail because file doesn't exist)
		// but we can check that it doesn't call the next handler
		assert.NotEqual(t, "next handler called", rr.Body.String())
	})

	t.Run("handles thumbnail requests", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/thumbnail/abc123", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		// Should handle thumbnail request
		assert.NotEqual(t, "next handler called", rr.Body.String())
	})

	t.Run("handles direct file requests", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/files/direct/test.mp4", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		// Should handle direct file request
		assert.NotEqual(t, "next handler called", rr.Body.String())
	})
}

func TestHandleVideoRequest(t *testing.T) {
	handler := NewAssetHandler()

	// Create a temporary test video file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.mp4")
	testContent := []byte("fake video content for testing")
	err := os.WriteFile(testFile, testContent, 0644)
	require.NoError(t, err)

	t.Run("existing file without range", func(t *testing.T) {
		// Encode the file path
		encodedPath := strings.Replace(testFile, "/", "%2F", -1)
		req := httptest.NewRequest(http.MethodGet, "/api/video/"+encodedPath, nil)
		rr := httptest.NewRecorder()

		handler.HandleVideoRequest(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "video/mp4", rr.Header().Get("Content-Type"))
		assert.Equal(t, testContent, rr.Body.Bytes())
	})

	t.Run("non-existent file", func(t *testing.T) {
		nonExistentPath := "/nonexistent/file.mp4"
		encodedPath := strings.Replace(nonExistentPath, "/", "%2F", -1)
		req := httptest.NewRequest(http.MethodGet, "/api/video/"+encodedPath, nil)
		rr := httptest.NewRecorder()

		handler.HandleVideoRequest(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("invalid URL encoding", func(t *testing.T) {
		// Create a request with properly formed URL but invalid encoded content
		req := httptest.NewRequest(http.MethodGet, "/api/video/", nil)
		// Manually set the URL path to something that will fail URL decoding
		req.URL.Path = "/api/video/invalid%"
		rr := httptest.NewRecorder()

		handler.HandleVideoRequest(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestHandleThumbnailRequest(t *testing.T) {
	handler := NewAssetHandler()

	t.Run("invalid URL encoding", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/thumbnail/", nil)
		// Manually set invalid URL encoding that will fail QueryUnescape
		req.URL.Path = "/api/thumbnail/invalid%"
		rr := httptest.NewRecorder()

		handler.HandleThumbnailRequest(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("URL encoded non-existent file", func(t *testing.T) {
		// URL encode "/nonexistent/file.mp4"
		nonExistentPath := "/nonexistent/file.mp4"
		encodedPath := url.QueryEscape(nonExistentPath)
		req := httptest.NewRequest(http.MethodGet, "/api/thumbnail/"+encodedPath, nil)
		rr := httptest.NewRecorder()

		handler.HandleThumbnailRequest(rr, req)

		// Should return 404 since file doesn't exist
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestHandleDirectFileRequest(t *testing.T) {
	handler := NewAssetHandler()

	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "direct_test.mp4")
	testContent := []byte("direct file content")
	err := os.WriteFile(testFile, testContent, 0644)
	require.NoError(t, err)

	t.Run("existing file", func(t *testing.T) {
		// URL encode the full file path after /api/files/direct/
		encodedPath := strings.Replace(testFile, "/", "%2F", -1)
		req := httptest.NewRequest(http.MethodGet, "/api/files/direct/"+encodedPath, nil)
		rr := httptest.NewRecorder()

		handler.HandleDirectFileRequest(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "video/mp4", rr.Header().Get("Content-Type"))
		assert.Equal(t, testContent, rr.Body.Bytes())
	})

	t.Run("non-existent file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/files/direct/nonexistent.mp4", nil)
		rr := httptest.NewRecorder()

		handler.HandleDirectFileRequest(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// Benchmark tests for performance-critical functions
func TestHandleRangeRequest(t *testing.T) {
	handler := NewAssetHandler()

	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "range_test.mp4")
	testContent := []byte("0123456789abcdefghijklmnopqrstuvwxyz") // 36 bytes
	err := os.WriteFile(testFile, testContent, 0644)
	require.NoError(t, err)

	file, err := os.Open(testFile)
	require.NoError(t, err)
	defer file.Close()

	fileInfo, err := file.Stat()
	require.NoError(t, err)

	t.Run("valid range request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		handler.HandleRangeRequest(rr, req, file, fileInfo.Size(), "bytes=0-9")

		assert.Equal(t, http.StatusPartialContent, rr.Code)
		assert.Equal(t, "bytes 0-9/36", rr.Header().Get("Content-Range"))
		assert.Equal(t, "10", rr.Header().Get("Content-Length"))
		assert.Equal(t, testContent[0:10], rr.Body.Bytes())
	})

	t.Run("invalid range format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		handler.HandleRangeRequest(rr, req, file, fileInfo.Size(), "notbytes=0-9")

		assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, rr.Code)
	})

	t.Run("invalid range parts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		handler.HandleRangeRequest(rr, req, file, fileInfo.Size(), "bytes=0-9-15")

		assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, rr.Code)
	})

	t.Run("range with empty end", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		handler.HandleRangeRequest(rr, req, file, fileInfo.Size(), "bytes=10-")

		assert.Equal(t, http.StatusPartialContent, rr.Code)
		assert.Equal(t, "bytes 10-35/36", rr.Header().Get("Content-Range"))
		assert.Equal(t, testContent[10:], rr.Body.Bytes())
	})

	t.Run("invalid start range", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		handler.HandleRangeRequest(rr, req, file, fileInfo.Size(), "bytes=abc-10")

		assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, rr.Code)
	})

	t.Run("range beyond file size", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		handler.HandleRangeRequest(rr, req, file, fileInfo.Size(), "bytes=50-60")

		assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, rr.Code)
	})
}

func TestVideoRequestWithRange(t *testing.T) {
	handler := NewAssetHandler()

	// Create a temporary test video file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "range_video.mp4")
	testContent := []byte("video content for range testing")
	err := os.WriteFile(testFile, testContent, 0644)
	require.NoError(t, err)

	t.Run("video request with range header", func(t *testing.T) {
		// Encode the file path
		encodedPath := strings.Replace(testFile, "/", "%2F", -1)
		req := httptest.NewRequest(http.MethodGet, "/api/video/"+encodedPath, nil)
		req.Header.Set("Range", "bytes=0-5")
		rr := httptest.NewRecorder()

		handler.HandleVideoRequest(rr, req)

		assert.Equal(t, http.StatusPartialContent, rr.Code)
		assert.Equal(t, "video/mp4", rr.Header().Get("Content-Type"))
		assert.Equal(t, "bytes", rr.Header().Get("Accept-Ranges"))
		assert.Equal(t, testContent[0:6], rr.Body.Bytes())
	})
}

func TestGenerateThumbnail(t *testing.T) {
	handler := NewAssetHandler()

	t.Run("thumbnail directory creation", func(t *testing.T) {
		// Use a non-existent video path - this will fail at ffmpeg stage but test directory creation
		nonExistentVideo := "/nonexistent/video.mp4"
		
		// This will create the thumbnails directory but fail at ffmpeg
		_, err := handler.GenerateThumbnail(nonExistentVideo)
		
		// Should get an error from FFmpeg creation (in dev/test mode no embedded binary)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "FFmpeg not available")
		
		// But thumbnails directory should exist
		_, err = os.Stat("thumbnails")
		assert.NoError(t, err)
		
		// Clean up
		os.RemoveAll("thumbnails")
	})
}

// Benchmark tests for performance-critical functions
func BenchmarkIsVideoFile(b *testing.B) {
	handler := NewAssetHandler()
	testPaths := []string{
		"/path/to/video.mp4",
		"/path/to/video.avi",
		"/path/to/document.txt",
		"/path/to/image.jpg",
		"",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			handler.IsVideoFile(path)
		}
	}
}

func BenchmarkGetContentType(b *testing.B) {
	handler := NewAssetHandler()
	testPaths := []string{
		"/path/to/video.mp4",
		"/path/to/video.avi",
		"/path/to/video.mov",
		"/path/to/file.unknown",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			handler.GetContentType(path)
		}
	}
}

func BenchmarkGetThumbnailURL(b *testing.B) {
	handler := NewAssetHandler()
	testPaths := []string{
		"/path/to/video.mp4",
		"/very/long/path/to/video/file.avi",
		"/short.mov",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			handler.GetThumbnailURL(path)
		}
	}
}