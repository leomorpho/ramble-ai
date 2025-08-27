package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Get companies collection reference
		companies, err := app.FindCollectionByNameOrId("companies")
		if err != nil {
			return err
		}

		// Create customer_contacts collection
		customerContacts := core.NewBaseCollection("customer_contacts")
		
		customerContacts.Fields.Add(
			&core.RelationField{
				Name:          "company_id",
				Required:      true,
				CollectionId:  companies.Id,
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.TextField{
				Name:     "phone_number",
				Required: false,
			},
			&core.TextField{
				Name:     "email",
				Required: false,
			},
			&core.TextField{
				Name: "name",
			},
		)

		// Add unique indexes for phone and email per company
		customerContacts.AddIndex("idx_customer_contacts_company_phone", true, "company_id", "phone_number")
		customerContacts.AddIndex("idx_customer_contacts_company_email", true, "company_id", "email")

		// Security rules - only company employees can access
		customerContacts.ListRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = company_id")
		customerContacts.ViewRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = company_id")
		customerContacts.CreateRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = company_id")
		customerContacts.UpdateRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = company_id")
		customerContacts.DeleteRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = company_id")

		if err := app.Save(customerContacts); err != nil {
			return err
		}

		// Create conversations collection
		conversations := core.NewBaseCollection("conversations")
		
		conversations.Fields.Add(
			&core.RelationField{
				Name:          "company_id",
				Required:      true,
				CollectionId:  companies.Id,
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.RelationField{
				Name:          "customer_contact_id",
				Required:      true,
				CollectionId:  customerContacts.Id,
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.SelectField{
				Name:     "channel",
				Required: true,
				Values:   []string{"email", "sms", "call"},
			},
			&core.DateField{
				Name: "last_message_at",
			},
		)

		// Index for efficient querying
		conversations.AddIndex("idx_conversations_company", false, "company_id", "customer_contact_id")
		conversations.AddIndex("idx_conversations_recent", false, "company_id", "last_message_at")

		// Security rules - only company employees can access
		conversations.ListRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = company_id")
		conversations.ViewRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = company_id")
		conversations.CreateRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = company_id")
		conversations.UpdateRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = company_id")
		conversations.DeleteRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = company_id")

		if err := app.Save(conversations); err != nil {
			return err
		}

		// Create messages collection
		messages := core.NewBaseCollection("messages")
		
		messages.Fields.Add(
			&core.RelationField{
				Name:          "conversation_id",
				Required:      true,
				CollectionId:  conversations.Id,
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.SelectField{
				Name:     "author",
				Required: true,
				Values:   []string{"customer", "company"},
			},
			&core.RelationField{
				Name:          "author_user_id",
				Required:      false,
				CollectionId:  "_pb_users_auth_",
				MaxSelect:     1,
				CascadeDelete: false,
			},
			&core.TextField{
				Name:     "content",
				Required: false,
			},
			&core.FileField{
				Name:      "audio_file",
				MaxSelect: 1,
				MaxSize:   104857600, // 100MB for audio files
				MimeTypes: []string{
					"audio/mpeg",
					"audio/mp3",
					"audio/wav",
					"audio/ogg",
					"audio/webm",
				},
			},
		)

		// Index for efficient message ordering
		messages.AddIndex("idx_messages_conversation", false, "conversation_id", "")

		// Security rules - company employees can access messages through conversation
		messages.ListRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = @collection.conversations.company_id")
		messages.ViewRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = @collection.conversations.company_id")
		messages.CreateRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = @collection.conversations.company_id")
		messages.UpdateRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = @collection.conversations.company_id")
		messages.DeleteRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = @collection.conversations.company_id")

		if err := app.Save(messages); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: Delete collections in reverse order
		collections := []string{"messages", "conversations", "customer_contacts"}
		
		for _, name := range collections {
			collection, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				continue // Collection doesn't exist, skip
			}
			if err := app.Delete(collection); err != nil {
				return err
			}
		}
		
		return nil
	})
}
