package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Create processed_files collection for video/audio transcription usage tracking
		processedFiles := core.NewBaseCollection("processed_files")
		
		// Add fields to processed_files
		processedFiles.Fields.Add(
			&core.RelationField{
				Name:          "user_id",
				Required:      true,
				CollectionId:  "_pb_users_auth_",
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.TextField{
				Name: "filename",
			},
			&core.NumberField{
				Name: "file_size_bytes",
			},
			&core.NumberField{
				Name: "duration_seconds",
			},
			&core.NumberField{
				Name: "processing_time_ms",
			},
			&core.SelectField{
				Name: "status",
				Values: []string{"processing", "completed", "failed"},
			},
			&core.NumberField{
				Name: "transcript_length",
			},
			&core.NumberField{
				Name: "words_count",
			},
			&core.TextField{
				Name: "model_used",
			},
			&core.TextField{
				Name: "client_ip",
			},
			&core.DateField{
				Name: "created",
			},
			&core.DateField{
				Name: "updated",
			},
		)

		// Add indexes for efficient querying
		processedFiles.AddIndex("idx_processed_files_user_id", false, "user_id", "")
		processedFiles.AddIndex("idx_processed_files_status", false, "status", "")

		// Security rules - users can only access their own processed files
		processedFiles.ListRule = types.Pointer("@request.auth.id != '' && user_id = @request.auth.id")
		processedFiles.ViewRule = types.Pointer("@request.auth.id != '' && user_id = @request.auth.id")
		processedFiles.CreateRule = types.Pointer("@request.auth.id != '' && user_id = @request.auth.id")
		processedFiles.UpdateRule = types.Pointer("@request.auth.id != '' && user_id = @request.auth.id")
		processedFiles.DeleteRule = types.Pointer("@request.auth.id != '' && user_id = @request.auth.id")

		if err := app.Save(processedFiles); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: Delete processed_files collection
		collection, err := app.FindCollectionByNameOrId("processed_files")
		if err != nil {
			return nil // Collection doesn't exist, nothing to rollback
		}
		
		return app.Delete(collection)
	})
}