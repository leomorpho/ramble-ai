package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Create api_keys collection for Ramble AI API key management
		apiKeys := core.NewBaseCollection("api_keys")
		
		// Add fields to api_keys
		apiKeys.Fields.Add(
			&core.TextField{
				Name:     "key_hash",
				Required: true,
			},
			&core.RelationField{
				Name:          "user_id",
				Required:      true,
				CollectionId:  "_pb_users_auth_",
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.BoolField{
				Name: "active",
			},
			&core.TextField{
				Name: "name",
			},
		)

		// Add unique index for key_hash
		apiKeys.AddIndex("idx_api_keys_key_hash", true, "key_hash", "")
		
		// Add index for user_id for efficient lookups
		apiKeys.AddIndex("idx_api_keys_user_id", false, "user_id", "")

		// Security rules - only allow users to manage their own API keys
		apiKeys.ListRule = types.Pointer("@request.auth.id != '' && user_id = @request.auth.id")
		apiKeys.ViewRule = types.Pointer("@request.auth.id != '' && user_id = @request.auth.id")
		apiKeys.CreateRule = types.Pointer("@request.auth.id != '' && user_id = @request.auth.id")
		apiKeys.UpdateRule = types.Pointer("@request.auth.id != '' && user_id = @request.auth.id")
		apiKeys.DeleteRule = types.Pointer("@request.auth.id != '' && user_id = @request.auth.id")

		if err := app.Save(apiKeys); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: Delete api_keys collection
		collection, err := app.FindCollectionByNameOrId("api_keys")
		if err != nil {
			return nil // Collection doesn't exist, nothing to rollback
		}
		
		return app.Delete(collection)
	})
}