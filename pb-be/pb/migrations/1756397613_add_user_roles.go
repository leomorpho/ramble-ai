package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Add role field to users collection
		collection, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		// Add role field if it doesn't exist
		if collection.Fields.GetByName("role") == nil {
			collection.Fields.Add(
				&core.SelectField{
					Name:     "role",
					Required: false,
					Values:   []string{"user", "admin"},
				},
			)
		}

		if err := app.Save(collection); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: Remove role field from users collection
		collection, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return nil // Collection doesn't exist, nothing to rollback
		}
		
		// Remove role field if it exists
		if roleField := collection.Fields.GetByName("role"); roleField != nil {
			collection.Fields.RemoveById(roleField.GetId())
			return app.Save(collection)
		}
		
		return nil
	})
}