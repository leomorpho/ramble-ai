package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		// Seed Products
		productsCollection, err := app.FindCollectionByNameOrId("products")
		if err != nil {
			return err
		}

		products := []map[string]any{
			{
				"product_id":    "prod_starter_plan",
				"active":        true,
				"name":          "Starter Plan",
				"description":   "Perfect for individuals getting started",
				"image":         "",
				"metadata":      `{"category": "subscription", "tier": "basic"}`,
				"product_order": 1,
			},
			{
				"product_id":    "prod_professional_plan",
				"active":        true,
				"name":          "Professional Plan",
				"description":   "Ideal for growing businesses and teams",
				"image":         "",
				"metadata":      `{"category": "subscription", "tier": "professional"}`,
				"product_order": 2,
			},
			{
				"product_id":    "prod_enterprise_plan",
				"active":        true,
				"name":          "Enterprise Plan",
				"description":   "Advanced features for large organizations",
				"image":         "",
				"metadata":      `{"category": "subscription", "tier": "enterprise"}`,
				"product_order": 3,
			},
			{
				"product_id":    "prod_one_time_credit",
				"active":        true,
				"name":          "Extra Credits",
				"description":   "One-time purchase for additional credits",
				"image":         "",
				"metadata":      `{"category": "one_time", "type": "credits"}`,
				"product_order": 4,
			},
		}

		for _, product := range products {
			record := core.NewRecord(productsCollection)
			record.Load(product)
			if err := app.Save(record); err != nil {
				return err
			}
		}

		// Seed Prices
		pricesCollection, err := app.FindCollectionByNameOrId("prices")
		if err != nil {
			return err
		}

		prices := []map[string]any{
			// Starter Plan prices
			{
				"price_id":          "price_starter_monthly",
				"product_id":        "prod_starter_plan",
				"active":            true,
				"description":       "Starter Plan - Monthly",
				"currency":          "usd",
				"unit_amount":       999, // $9.99
				"type":              "recurring",
				"interval":          "month",
				"interval_count":    1,
				"trial_period_days": 7,
				"metadata":          `{"features": "Basic features, 5 projects, 10GB storage"}`,
			},
			{
				"price_id":          "price_starter_yearly",
				"product_id":        "prod_starter_plan",
				"active":            true,
				"description":       "Starter Plan - Yearly",
				"currency":          "usd",
				"unit_amount":       9999, // $99.99 (2 months free)
				"type":              "recurring",
				"interval":          "year",
				"interval_count":    1,
				"trial_period_days": 14,
				"metadata":          `{"features": "Basic features, 5 projects, 10GB storage", "discount": "17% off"}`,
			},
			// Professional Plan prices
			{
				"price_id":          "price_professional_monthly",
				"product_id":        "prod_professional_plan",
				"active":            true,
				"description":       "Professional Plan - Monthly",
				"currency":          "usd",
				"unit_amount":       2999, // $29.99
				"type":              "recurring",
				"interval":          "month",
				"interval_count":    1,
				"trial_period_days": 14,
				"metadata":          `{"features": "Advanced features, 50 projects, 100GB storage, Priority support"}`,
			},
			{
				"price_id":          "price_professional_yearly",
				"product_id":        "prod_professional_plan",
				"active":            true,
				"description":       "Professional Plan - Yearly",
				"currency":          "usd",
				"unit_amount":       29999, // $299.99 (2 months free)
				"type":              "recurring",
				"interval":          "year",
				"interval_count":    1,
				"trial_period_days": 30,
				"metadata":          `{"features": "Advanced features, 50 projects, 100GB storage, Priority support", "discount": "17% off"}`,
			},
			// Enterprise Plan prices
			{
				"price_id":          "price_enterprise_monthly",
				"product_id":        "prod_enterprise_plan",
				"active":            true,
				"description":       "Enterprise Plan - Monthly",
				"currency":          "usd",
				"unit_amount":       9999, // $99.99
				"type":              "recurring",
				"interval":          "month",
				"interval_count":    1,
				"trial_period_days": 30,
				"metadata":          `{"features": "Enterprise features, Unlimited projects, 1TB storage, Dedicated support, SSO"}`,
			},
			{
				"price_id":          "price_enterprise_yearly",
				"product_id":        "prod_enterprise_plan",
				"active":            true,
				"description":       "Enterprise Plan - Yearly",
				"currency":          "usd",
				"unit_amount":       99999, // $999.99 (2 months free)
				"type":              "recurring",
				"interval":          "year",
				"interval_count":    1,
				"trial_period_days": 30,
				"metadata":          `{"features": "Enterprise features, Unlimited projects, 1TB storage, Dedicated support, SSO", "discount": "17% off"}`,
			},
			// One-time credit packages
			{
				"price_id":          "price_credits_small",
				"product_id":        "prod_one_time_credit",
				"active":            true,
				"description":       "100 Extra Credits",
				"currency":          "usd",
				"unit_amount":       499, // $4.99
				"type":             "one_time",
				"interval":          "",
				"interval_count":    0,
				"trial_period_days": types.Pointer(0),
				"metadata":          `{"credits": 100, "bonus": "5% bonus credits"}`,
			},
			{
				"price_id":          "price_credits_medium",
				"product_id":        "prod_one_time_credit",
				"active":            true,
				"description":       "500 Extra Credits",
				"currency":          "usd",
				"unit_amount":       1999, // $19.99
				"type":             "one_time",
				"interval":          "",
				"interval_count":    0,
				"trial_period_days": types.Pointer(0),
				"metadata":          `{"credits": 500, "bonus": "10% bonus credits"}`,
			},
			{
				"price_id":          "price_credits_large",
				"product_id":        "prod_one_time_credit",
				"active":            true,
				"description":       "1000 Extra Credits",
				"currency":          "usd",
				"unit_amount":       3499, // $34.99
				"type":             "one_time",
				"interval":          "",
				"interval_count":    0,
				"trial_period_days": types.Pointer(0),
				"metadata":          `{"credits": 1000, "bonus": "15% bonus credits"}`,
			},
		}

		for _, price := range prices {
			record := core.NewRecord(pricesCollection)
			record.Load(price)
			if err := app.Save(record); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		// Clean up seed data on migration down
		collections := []string{"prices", "products"}
		
		for _, collectionName := range collections {
			collection, err := app.FindCollectionByNameOrId(collectionName)
			if err != nil {
				continue // Collection might not exist
			}

			records, err := app.FindAllRecords(collection.Name)
			if err != nil {
				continue
			}

			for _, record := range records {
				if err := app.Delete(record); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
