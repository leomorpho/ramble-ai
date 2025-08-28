package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Create banners collection for system announcements and updates
		banners := core.NewBaseCollection("banners")
		
		// Add fields to banners
		banners.Fields.Add(
			&core.TextField{
				Name:     "title",
				Required: true,
			},
			&core.TextField{
				Name:     "message",
				Required: true,
			},
			&core.SelectField{
				Name:     "type",
				Required: true,
				Values:   []string{"info", "warning", "success", "error"},
			},
			&core.BoolField{
				Name: "active",
			},
			&core.BoolField{
				Name: "requires_auth",
			},
			&core.TextField{
				Name: "action_url",
			},
			&core.TextField{
				Name: "action_text",
			},
			&core.DateField{
				Name: "expires_at",
			},
		)

		// Add index for active banners for efficient queries
		banners.AddIndex("idx_banners_active", false, "active", "")
		
		// Add index for expiration for cleanup queries
		banners.AddIndex("idx_banners_expires_at", false, "expires_at", "")

		// Security rules
		// Public banners - anyone can read active banners that don't require auth
		banners.ListRule = types.Pointer("active = true && (requires_auth = false || @request.auth.id != '')")
		banners.ViewRule = types.Pointer("active = true && (requires_auth = false || @request.auth.id != '')")
		
		// Only admins can create/update/delete banners
		banners.CreateRule = types.Pointer("@request.auth.id != '' && @request.auth.role = 'admin'")
		banners.UpdateRule = types.Pointer("@request.auth.id != '' && @request.auth.role = 'admin'")
		banners.DeleteRule = types.Pointer("@request.auth.id != '' && @request.auth.role = 'admin'")

		if err := app.Save(banners); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: Delete banners collection
		collection, err := app.FindCollectionByNameOrId("banners")
		if err != nil {
			return nil // Collection doesn't exist, nothing to rollback
		}
		
		return app.Delete(collection)
	})
}