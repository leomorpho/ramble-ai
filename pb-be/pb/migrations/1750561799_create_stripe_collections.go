package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Create companies collection first (needed for relations)
		companies := core.NewBaseCollection("companies")
		companies.Fields.Add(
			&core.TextField{
				Name:     "name",
				Required: true,
			},
			&core.RelationField{
				Name:          "owner_id",
				Required:      true,
				CollectionId:  "_pb_users_auth_",
				MaxSelect:     1,
				CascadeDelete: false,
			},
			&core.TextField{
				Name:     "stripe_customer_id",
				Required: false,
			},
			&core.TextField{
				Name:     "domain",
				Required: false,
			},
			&core.FileField{
				Name:      "logo",
				MaxSelect: 1,
				MaxSize:   5242880, // 5MB
				MimeTypes: []string{
					"image/jpeg",
					"image/png",
					"image/svg+xml",
					"image/webp",
				},
			},
			&core.JSONField{
				Name: "settings",
			},
		)

		// Add unique index for stripe_customer_id
		companies.AddIndex("idx_companies_stripe_customer_id", true, "stripe_customer_id", "")

		// Temporarily set simple rules for companies (will update after employees is created)
		companies.ListRule = types.Pointer("@request.auth.id != '' && @request.auth.id = owner_id")
		companies.ViewRule = types.Pointer("@request.auth.id != '' && @request.auth.id = owner_id")
		companies.CreateRule = types.Pointer("@request.auth.id != '' && @request.auth.id = owner_id")
		companies.UpdateRule = types.Pointer("@request.auth.id = owner_id")
		companies.DeleteRule = types.Pointer("@request.auth.id = owner_id")

		if err := app.Save(companies); err != nil {
			return err
		}

		// Create employees collection
		employees := core.NewBaseCollection("employees")
		employees.Fields.Add(
			&core.RelationField{
				Name:          "user_id",
				Required:      true,
				CollectionId:  "_pb_users_auth_",
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.RelationField{
				Name:          "company_id",
				Required:      true,
				CollectionId:  companies.Id,
				MaxSelect:     1,
				CascadeDelete: true,
			},
			&core.SelectField{
				Name:     "role",
				Required: true,
				Values:   []string{"owner", "admin", "member"},
			},
			&core.DateField{
				Name: "joined_at",
			},
		)

		// Add indexes for employees
		employees.AddIndex("idx_employees_user_id", true, "user_id", "")
		employees.AddIndex("idx_employees_company_id", false, "company_id", "")

		// Security rules for employees
		employees.ListRule = types.Pointer("@request.auth.id = user_id || (@request.auth.id = @collection.companies.owner_id && company_id = @collection.companies.id)")
		employees.ViewRule = types.Pointer("@request.auth.id = user_id || (@request.auth.id = @collection.companies.owner_id && company_id = @collection.companies.id)")
		// For now, keep create as backend-only. We'll handle company setup through a custom endpoint or allow any authenticated user to create
		employees.CreateRule = types.Pointer("@request.auth.id != ''")
		employees.UpdateRule = types.Pointer("@request.auth.id = @collection.companies.owner_id && company_id = @collection.companies.id")
		employees.DeleteRule = types.Pointer("@request.auth.id = @collection.companies.owner_id && company_id = @collection.companies.id")

		if err := app.Save(employees); err != nil {
			return err
		}

		// Now update companies collection with full security rules
		companiesUpdate, err := app.FindCollectionByNameOrId("companies")
		if err != nil {
			return err
		}
		companiesUpdate.ListRule = types.Pointer("@request.auth.id != '' && (@request.auth.id = owner_id || @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = id)")
		companiesUpdate.ViewRule = types.Pointer("@request.auth.id != '' && (@request.auth.id = owner_id || @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = id)")
		
		if err := app.Save(companiesUpdate); err != nil {
			return err
		}

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

		// Create stripe_customers collection
		customers := core.NewBaseCollection("stripe_customers")
		
		customers.Fields.Add(
			&core.TextField{
				Name:     "stripe_customer_id",
				Required: true,
			},
			&core.RelationField{
				Name:          "company_id",
				Required:      true,
				CollectionId:  companies.Id,
				MaxSelect:     1,
				CascadeDelete: true,
			},
		)

		customers.AddIndex("idx_customers_stripe_customer_id", true, "stripe_customer_id", "")
		customers.AddIndex("idx_customers_company_id", true, "company_id", "")

		// Set access rules - only company members can see their company's customer record
		customers.ListRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = company_id")
		customers.ViewRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = company_id")

		if err := app.Save(customers); err != nil {
			return err
		}

		// Create subscriptions collection
		subscriptions := core.NewBaseCollection("subscriptions")
		
		subscriptions.Fields.Add(
			&core.TextField{
				Name:     "subscription_id",
				Required: true,
			},
			&core.RelationField{
				Name:          "company_id",
				Required:      true,
				CollectionId:  companies.Id,
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

		// Set access rules - only company members can see their company's subscriptions
		subscriptions.ListRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = company_id")
		subscriptions.ViewRule = types.Pointer("@request.auth.id != '' && @collection.employees.user_id = @request.auth.id && @collection.employees.company_id = company_id")

		if err := app.Save(subscriptions); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: Delete collections in reverse order
		collections := []string{"subscriptions", "stripe_customers", "prices", "products", "employees", "companies"}
		
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