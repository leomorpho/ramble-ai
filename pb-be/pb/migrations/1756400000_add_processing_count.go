package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("processed_files")
		if err != nil {
			return err
		}

		// Add processing_count field to track how many times a file has been processed
		minVal := 1.0
		maxVal := 2.0
		collection.Fields.Add(&core.NumberField{
			Name:     "processing_count",
			Required: false,
			Min:      &minVal, // Minimum count is 1 (when first processed)
			Max:      &maxVal, // Maximum allowed reprocessing is 2 times
		})

		// Add index for efficient querying by filename and user
		collection.AddIndex("idx_processed_files_user_filename", false, "user_id", "filename")

		return app.Save(collection)
	}, func(app core.App) error {
		// Rollback: Remove processing_count field
		collection, err := app.FindCollectionByNameOrId("processed_files")
		if err != nil {
			return nil // Collection doesn't exist, nothing to rollback
		}

		// Remove the processing_count field
		for i, field := range collection.Fields {
			if field.GetName() == "processing_count" {
				collection.Fields = append(collection.Fields[:i], collection.Fields[i+1:]...)
				break
			}
		}

		// Remove the index
		for i, indexName := range collection.Indexes {
			if indexName == "idx_processed_files_user_filename" {
				collection.Indexes = append(collection.Indexes[:i], collection.Indexes[i+1:]...)
				break
			}
		}

		return app.Save(collection)
	})
}