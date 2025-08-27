package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// 1. Update users collection to add avatar thumbnails
		usersCollection, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		// Find the avatar field and update it with thumbnails
		for _, field := range usersCollection.Fields {
			if fileField, ok := field.(*core.FileField); ok && fileField.Name == "avatar" {
				// Add comprehensive thumbnail sizes for avatars
				fileField.Thumbs = []string{
					"32x32",   // Small avatar (nav, lists)
					"64x64",   // Medium avatar (cards, profile previews)
					"128x128", // Large avatar (profile pages)
					"200x200", // Extra large avatar (detailed views)
					"32x0",    // Small width-constrained
					"64x0",    // Medium width-constrained
					"128x0",   // Large width-constrained
					"0x32",    // Small height-constrained
					"0x64",    // Medium height-constrained
					"0x128",   // Large height-constrained
				}
			}
		}

		if err := app.Save(usersCollection); err != nil {
			return err
		}

		// 2. Create file_uploads collection with comprehensive thumbnail support
		// First, delete existing file_uploads collection if it exists
		if collection, err := app.FindCollectionByNameOrId("file_uploads"); err == nil {
			if err := app.Delete(collection); err != nil {
				return err
			}
		}

		// Create file_uploads collection
		fileUploads := core.NewBaseCollection("file_uploads")
		
		fileUploads.Fields.Add(
			&core.FileField{
				Name:      "file",
				Required:  false,
				MaxSelect: 1,
				MaxSize:   104857600, // 100MB
				MimeTypes: []string{}, // Allow all file types
				Thumbs: []string{
					"32x32", "64x64", "100x100", "128x128", "200x200", "300x300", "400x400",
					"600x400", "400x600", "32x0", "64x0", "128x0", "200x0",
					"0x32", "0x64", "0x128", "0x200",
					"800x600f", "400x300f", "200x150f",
					"100x100t", "200x200b", "300x200t",
				},
				Protected: false,
			},
			&core.TextField{
				Name:     "upload_id",
				Required: true,
				Pattern:  "^[a-zA-Z0-9+/=_-]+$",
			},
			&core.JSONField{
				Name:     "metadata",
				Required: false,
				MaxSize:  2000000,
			},
			&core.SelectField{
				Name:     "processing_status",
				Required: true,
				MaxSelect: 1,
				Values: []string{
					"pending",
					"processing", 
					"completed",
					"failed",
				},
			},
			&core.TextField{
				Name:     "file_type",
				Required: true,
				Pattern:  "^(avatar|document|media|temp)$",
			},
			&core.TextField{
				Name:     "category",
				Required: false,
				Max:      100,
			},
			&core.RelationField{
				Name:          "user",
				Required:      true,
				CollectionId:  "_pb_users_auth_",
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.SelectField{
				Name:      "visibility",
				Required:  true,
				MaxSelect: 1,
				Values: []string{
					"public",
					"private",
					"shared",
				},
			},
			&core.JSONField{
				Name:     "processed_variants",
				Required: false,
				MaxSize:  2000000,
			},
			&core.TextField{
				Name:     "original_name",
				Required: false,
				Max:      255,
			},
		)

		// Add indexes
		fileUploads.AddIndex("idx_file_uploads_upload_id", true, "upload_id", "")
		fileUploads.AddIndex("idx_file_uploads_user", false, "user", "")
		fileUploads.AddIndex("idx_file_uploads_file_type", false, "file_type", "")

		// Set access rules
		fileUploads.ListRule = types.Pointer("@request.auth.id = user.id")
		fileUploads.ViewRule = types.Pointer("@request.auth.id = user.id || visibility = \"public\"")
		fileUploads.CreateRule = types.Pointer("@request.auth.id != \"\"")
		fileUploads.UpdateRule = types.Pointer("@request.auth.id = user.id")
		fileUploads.DeleteRule = types.Pointer("@request.auth.id = user.id")

		if err := app.Save(fileUploads); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: Remove thumbnails from users avatar field
		usersCollection, err := app.FindCollectionByNameOrId("users")
		if err == nil {
			// Find the avatar field and remove thumbnails
			for _, field := range usersCollection.Fields {
				if fileField, ok := field.(*core.FileField); ok && fileField.Name == "avatar" {
					fileField.Thumbs = nil
				}
			}
			app.Save(usersCollection)
		}

		// Rollback: Delete file_uploads collection
		collection, err := app.FindCollectionByNameOrId("file_uploads")
		if err != nil {
			return nil // Collection doesn't exist, skip
		}
		return app.Delete(collection)
	})
}