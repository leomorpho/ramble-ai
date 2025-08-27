package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Create user_otps collection for storing OTP codes
		userOtps := core.NewBaseCollection("user_otps")
		
		userOtps.Fields.Add(
			&core.RelationField{
				Name:          "user_id",
				Required:      true,
				CollectionId:  "_pb_users_auth_",
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.TextField{
				Name:     "otp_code",
				Required: true,
				Max:      6, // 6-digit OTP
				Min:      6,
			},
			&core.SelectField{
				Name:     "purpose",
				Required: true,
				Values:   []string{"signup_verification", "email_change", "password_reset"},
			},
			&core.DateField{
				Name:     "expires_at",
				Required: true,
			},
			&core.BoolField{
				Name: "used",
			},
			&core.TextField{
				Name: "email", // Store the email this OTP was sent to
			},
		)

		// Add indexes for efficient querying
		userOtps.AddIndex("idx_user_otps_user_id", false, "user_id", "")
		userOtps.AddIndex("idx_user_otps_code", false, "otp_code", "")
		userOtps.AddIndex("idx_user_otps_expires", false, "expires_at", "")

		// Security rules - users can only access their own OTPs
		userOtps.ListRule = types.Pointer("@request.auth.id != '' && user_id = @request.auth.id")
		userOtps.ViewRule = types.Pointer("@request.auth.id != '' && user_id = @request.auth.id")
		userOtps.CreateRule = types.Pointer("@request.auth.id != '' && user_id = @request.auth.id")
		userOtps.UpdateRule = types.Pointer("@request.auth.id != '' && user_id = @request.auth.id")
		userOtps.DeleteRule = types.Pointer("@request.auth.id != '' && user_id = @request.auth.id")

		if err := app.Save(userOtps); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: Delete the user_otps collection
		collection, err := app.FindCollectionByNameOrId("user_otps")
		if err != nil {
			return nil // Collection doesn't exist, nothing to delete
		}
		
		return app.Delete(collection)
	})
}