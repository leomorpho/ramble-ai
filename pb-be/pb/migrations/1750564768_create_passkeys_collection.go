package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// First, delete existing passkeys collection if it exists
		if collection, err := app.FindCollectionByNameOrId("passkeys"); err == nil {
			if err := app.Delete(collection); err != nil {
				return err
			}
		}

		// Create passkeys collection
		passkeys := core.NewBaseCollection("passkeys")
		
		passkeys.Fields.Add(
			&core.RelationField{
				Name:          "user",
				Required:      true,
				CollectionId:  "_pb_users_auth_",
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.TextField{
				Name:     "credential_id",
				Required: true,
			},
			&core.JSONField{
				Name:     "credentials",
				Required: true,
			},
			&core.TextField{
				Name: "name",
			},
		)

		passkeys.AddIndex("idx_passkeys_credential_id", true, "credential_id", "")

		// Set access rules - only users can see their own passkeys
		passkeys.ListRule = types.Pointer("user = @request.auth.id")
		passkeys.ViewRule = types.Pointer("user = @request.auth.id")
		passkeys.DeleteRule = types.Pointer("user = @request.auth.id")

		if err := app.Save(passkeys); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: Delete passkeys collection
		collection, err := app.FindCollectionByNameOrId("passkeys")
		if err != nil {
			return nil // Collection doesn't exist, skip
		}
		return app.Delete(collection)
	})
}
