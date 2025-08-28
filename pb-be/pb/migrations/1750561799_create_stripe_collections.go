package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Create products collection
		products := core.NewBaseCollection("products")
		products.ListRule = types.Pointer("")
		products.ViewRule = types.Pointer("")
		
		// Add fields to products
		products.Fields.Add(
			&core.TextField{
				Name:     "product_id",
				Required: true,
			},
			&core.BoolField{
				Name: "active",
			},
			&core.TextField{
				Name: "name",
			},
			&core.TextField{
				Name: "description",
			},
			&core.TextField{
				Name: "image",
			},
			&core.JSONField{
				Name: "metadata",
			},
			&core.NumberField{
				Name: "product_order",
			},
		)

		// Add unique index for product_id
		products.AddIndex("idx_products_product_id", true, "product_id", "")

		if err := app.Save(products); err != nil {
			return err
		}

		// Create prices collection
		prices := core.NewBaseCollection("prices")
		prices.ListRule = types.Pointer("")
		prices.ViewRule = types.Pointer("")
		
		prices.Fields.Add(
			&core.TextField{
				Name:     "price_id",
				Required: true,
			},
			&core.TextField{
				Name:     "product_id",
				Required: true,
			},
			&core.BoolField{
				Name: "active",
			},
			&core.TextField{
				Name: "description",
			},
			&core.TextField{
				Name: "currency",
			},
			&core.NumberField{
				Name: "unit_amount",
			},
			&core.TextField{
				Name: "type",
			},
			&core.TextField{
				Name: "interval",
			},
			&core.NumberField{
				Name: "interval_count",
			},
			&core.NumberField{
				Name: "trial_period_days",
			},
			&core.JSONField{
				Name: "metadata",
			},
		)

		prices.AddIndex("idx_prices_price_id", true, "price_id", "")

		if err := app.Save(prices); err != nil {
			return err
		}

		// Create stripe_customers collection - directly linked to users
		customers := core.NewBaseCollection("stripe_customers")
		
		customers.Fields.Add(
			&core.TextField{
				Name:     "stripe_customer_id",
				Required: true,
			},
			&core.RelationField{
				Name:          "user_id",
				Required:      true,
				CollectionId:  "_pb_users_auth_",
				MaxSelect:     1,
				CascadeDelete: true,
			},
		)

		customers.AddIndex("idx_customers_stripe_customer_id", true, "stripe_customer_id", "")
		customers.AddIndex("idx_customers_user_id", true, "user_id", "")

		// Set access rules - users can only see their own customer record
		customers.ListRule = types.Pointer("@request.auth.id != '' && @request.auth.id = user_id")
		customers.ViewRule = types.Pointer("@request.auth.id != '' && @request.auth.id = user_id")

		if err := app.Save(customers); err != nil {
			return err
		}

		// Create subscriptions collection - directly linked to users
		subscriptions := core.NewBaseCollection("subscriptions")
		
		subscriptions.Fields.Add(
			&core.TextField{
				Name:     "subscription_id",
				Required: true,
			},
			&core.RelationField{
				Name:          "user_id",
				Required:      true,
				CollectionId:  "_pb_users_auth_",
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.TextField{
				Name: "status",
			},
			&core.TextField{
				Name: "price_id",
			},
			&core.NumberField{
				Name: "quantity",
			},
			&core.BoolField{
				Name: "cancel_at_period_end",
			},
			&core.NumberField{
				Name: "current_period_start",
			},
			&core.NumberField{
				Name: "current_period_end",
			},
			&core.NumberField{
				Name: "ended_at",
			},
			&core.NumberField{
				Name: "cancel_at",
			},
			&core.NumberField{
				Name: "canceled_at",
			},
			&core.NumberField{
				Name: "trial_start",
			},
			&core.NumberField{
				Name: "trial_end",
			},
			&core.JSONField{
				Name: "metadata",
			},
		)

		subscriptions.AddIndex("idx_subscriptions_subscription_id", true, "subscription_id", "")

		// Set access rules - users can only see their own subscriptions
		subscriptions.ListRule = types.Pointer("@request.auth.id != '' && @request.auth.id = user_id")
		subscriptions.ViewRule = types.Pointer("@request.auth.id != '' && @request.auth.id = user_id")

		if err := app.Save(subscriptions); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: Delete collections in reverse order
		collections := []string{"subscriptions", "stripe_customers", "prices", "products"}
		
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