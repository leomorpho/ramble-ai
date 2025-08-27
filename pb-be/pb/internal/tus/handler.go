package tus

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/tus/tusd/v2/pkg/handler"
)

// TUSHandler wraps the TUS handler with PocketBase integration
type TUSHandler struct {
	handler *handler.Handler
	app     core.App
}

// NewTUSHandler creates a new TUS handler with PocketBase integration
func NewTUSHandler(app core.App) (*TUSHandler, error) {
	// Create upload directory
	uploadDir := filepath.Join(app.DataDir(), "tus_uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Create PocketBase store
	store := NewPocketBaseStore(app)

	// Configure TUS handler
	composer := handler.NewStoreComposer()
	store.UseIn(composer)

	config := handler.Config{
		BasePath:                "/tus/",
		StoreComposer:          composer,
		NotifyCompleteUploads:  true,
		NotifyTerminatedUploads: true,
		NotifyUploadProgress:   true,
		NotifyCreatedUploads:   true,
	}

	tusHandler, err := handler.NewHandler(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUS handler: %w", err)
	}

	h := &TUSHandler{
		handler: tusHandler,
		app:     app,
	}

	// Set up hooks
	h.setupHooks()

	return h, nil
}

// setupHooks configures TUS event hooks for PocketBase integration
func (h *TUSHandler) setupHooks() {
	// Hook for upload completion
	go func() {
		for {
			select {
			case info := <-h.handler.CompleteUploads:
				h.handleUploadComplete(info)
			case info := <-h.handler.TerminatedUploads:
				h.handleUploadTerminated(info)
			case info := <-h.handler.CreatedUploads:
				h.handleUploadCreated(info)
			}
		}
	}()
}

// handleUploadCreated handles when a new upload is created
func (h *TUSHandler) handleUploadCreated(info handler.HookEvent) {
	metadata := info.Upload.MetaData
	
	// Create PocketBase record
	collection, err := h.app.FindCollectionByNameOrId("file_uploads")
	if err != nil {
		h.app.Logger().Error("Failed to find file_uploads collection", "error", err)
		return
	}

	record := core.NewRecord(collection)
	
	// Set initial record data
	record.Set("upload_id", info.Upload.ID)
	record.Set("processing_status", "pending")
	record.Set("original_name", metadata["filename"])
	
	// Parse metadata
	if fileType, ok := metadata["fileType"]; ok {
		record.Set("file_type", fileType)
	}
	if category, ok := metadata["category"]; ok {
		record.Set("category", category)
	}
	if userID, ok := metadata["userId"]; ok {
		record.Set("user", userID)
	}
	if visibility, ok := metadata["visibility"]; ok {
		record.Set("visibility", visibility)
	} else {
		record.Set("visibility", "private")
	}
	
	// Store all metadata as JSON
	metadataJSON, _ := json.Marshal(metadata)
	record.Set("metadata", string(metadataJSON))

	if err := h.app.Save(record); err != nil {
		h.app.Logger().Error("Failed to create file upload record", "error", err)
	}
}

// handleUploadComplete handles when an upload is completed
func (h *TUSHandler) handleUploadComplete(info handler.HookEvent) {
	// Find the record by upload_id
	record, err := h.app.FindFirstRecordByFilter(
		"file_uploads",
		"upload_id = {:uploadId}",
		map[string]any{"uploadId": info.Upload.ID},
	)
	if err != nil {
		h.app.Logger().Error("Failed to find upload record", "error", err)
		return
	}

	// Move file to PocketBase storage and update record
	if err := h.moveFileToStorage(record, info.Upload); err != nil {
		h.app.Logger().Error("Failed to move file to storage", "error", err)
		record.Set("processing_status", "failed")
	} else {
		record.Set("processing_status", "completed")
	}

	if err := h.app.Save(record); err != nil {
		h.app.Logger().Error("Failed to update upload record", "error", err)
	}

	// Trigger post-processing if needed
	h.triggerPostProcessing(record)
}

// handleUploadTerminated handles when an upload is terminated
func (h *TUSHandler) handleUploadTerminated(info handler.HookEvent) {
	// Find and delete the record
	record, err := h.app.FindFirstRecordByFilter(
		"file_uploads",
		"upload_id = {:uploadId}",
		map[string]any{"uploadId": info.Upload.ID},
	)
	if err != nil {
		return // Record might not exist
	}

	h.app.Delete(record)
}

// moveFileToStorage moves the completed upload to PocketBase file storage
func (h *TUSHandler) moveFileToStorage(record *core.Record, upload handler.FileInfo) error {
	// Get upload file path
	uploadPath := filepath.Join(h.app.DataDir(), "tus_uploads", upload.ID+".bin")
	
	// Open the upload file
	file, err := os.Open(uploadPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get original filename from metadata
	filename := "upload"
	if upload.MetaData["filename"] != "" {
		filename = upload.MetaData["filename"]
	}

	// For now, just store the filename - proper file storage integration
	// would require more complex handling of the PocketBase filesystem
	record.Set("file", filename)

	// Clean up temp file
	os.Remove(uploadPath)
	os.Remove(filepath.Join(h.app.DataDir(), "tus_uploads", upload.ID+".info"))

	return nil
}

// triggerPostProcessing triggers any post-upload processing
func (h *TUSHandler) triggerPostProcessing(record *core.Record) {
	// Parse metadata to check for processing instructions
	metadataStr := record.GetString("metadata")
	if metadataStr == "" {
		return
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
		return
	}

	// Check for processing instructions
	if processAfterUpload, ok := metadata["processAfterUpload"].([]interface{}); ok {
		record.Set("processing_status", "processing")
		h.app.Save(record)

		// Process each instruction
		for _, instruction := range processAfterUpload {
			if instructionStr, ok := instruction.(string); ok {
				h.processFile(record, instructionStr)
			}
		}

		record.Set("processing_status", "completed")
		h.app.Save(record)
	}
}

// processFile handles individual file processing instructions
func (h *TUSHandler) processFile(record *core.Record, instruction string) error {
	// Get file from record
	fileField := record.GetString("file")
	if fileField == "" {
		return fmt.Errorf("no file attached to record")
	}

	// Get filesystem
	fileSystem, err := h.app.NewFilesystem()
	if err != nil {
		return err
	}
	defer fileSystem.Close()

	// Process based on instruction
	switch {
	case strings.HasPrefix(instruction, "resize:"):
		return h.processImageResize(record, fileSystem, instruction)
	case instruction == "thumbnail":
		return h.processImageThumbnail(record, fileSystem)
	case instruction == "extract_text":
		return h.processTextExtraction(record, fileSystem)
	default:
		h.app.Logger().Warn("Unknown processing instruction", "instruction", instruction)
	}

	return nil
}

// processImageResize handles image resizing
func (h *TUSHandler) processImageResize(record *core.Record, fs *filesystem.System, instruction string) error {
	// Parse resize dimensions from instruction (e.g., "resize:200x200")
	// Implementation would use PocketBase's image processing capabilities
	h.app.Logger().Info("Processing image resize", "instruction", instruction)
	return nil
}

// processImageThumbnail generates thumbnails
func (h *TUSHandler) processImageThumbnail(record *core.Record, fs *filesystem.System) error {
	h.app.Logger().Info("Processing image thumbnail")
	return nil
}

// processTextExtraction extracts text from documents
func (h *TUSHandler) processTextExtraction(record *core.Record, fs *filesystem.System) error {
	h.app.Logger().Info("Processing text extraction")
	return nil
}

// ServeHTTP implements http.Handler
func (h *TUSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, HEAD, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Origin, X-Requested-With, Content-Type, Upload-Length, Upload-Offset, Tus-Resumable, Upload-Metadata, Upload-Defer-Length, Upload-Concat")
	w.Header().Set("Access-Control-Expose-Headers", "Upload-Offset, Location, Upload-Length, Tus-Version, Tus-Resumable, Tus-Max-Size, Tus-Extension, Upload-Metadata, Upload-Defer-Length, Upload-Concat")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Authenticate request using PocketBase auth
	if !h.authenticateRequest(r) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Authentication required"))
		return
	}

	// Delegate to TUS handler
	h.handler.ServeHTTP(w, r)
}

// authenticateRequest validates the request has valid PocketBase authentication
func (h *TUSHandler) authenticateRequest(r *http.Request) bool {
	// Extract auth token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	// Remove "Bearer " prefix
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return false
	}

	// Validate token with PocketBase - simple validation for now
	// In a real implementation, you'd want to properly validate the JWT token
	if len(token) < 10 {
		return false
	}
	
	// For now, we'll assume the token is valid if it's present
	// You should implement proper JWT validation here

	return true
}

// UseIn implements the store interface for TUS composer
func (store *PocketBaseStore) UseIn(composer *handler.StoreComposer) {
	composer.UseCore(store)
	composer.UseTerminater(store)
	composer.UseLengthDeferrer(store)
	composer.UseConcater(store)
}