package stripe

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/subscription"
	"github.com/stripe/stripe-go/v79/webhook"
)

// HandleWebhook processes Stripe webhook events
func HandleWebhook(e *core.RequestEvent, app *pocketbase.PocketBase) error {
	const MaxBodyBytes = int64(65536)
	e.Request.Body = http.MaxBytesReader(e.Response, e.Request.Body, MaxBodyBytes)

	payload, err := io.ReadAll(e.Request.Body)
	if err != nil {
		return e.JSON(http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
	}

	// Verify webhook signature
	endpointSecret := os.Getenv("STRIPE_SECRET_WHSEC")
	event, err := webhook.ConstructEvent(payload, e.Request.Header.Get("Stripe-Signature"), endpointSecret)
	if err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Handle the event
	switch event.Type {
	case "product.created", "product.updated":
		var product stripe.Product
		if err := json.Unmarshal(event.Data.Raw, &product); err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			return e.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		if err := upsertProduct(app, &product); err != nil {
			log.Printf("Error upserting product: %v", err)
		}

	case "price.created", "price.updated":
		var price stripe.Price
		if err := json.Unmarshal(event.Data.Raw, &price); err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			return e.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		if err := upsertPrice(app, &price); err != nil {
			log.Printf("Error upserting price: %v", err)
		}

	case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			return e.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		if err := upsertSubscription(app, &sub); err != nil {
			log.Printf("Error upserting subscription: %v", err)
		}

	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			return e.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		if session.Mode == "subscription" && session.Subscription != nil {
			// Retrieve the subscription to get full details
			sub, err := subscription.Get(session.Subscription.ID, nil)
			if err != nil {
				log.Printf("Error retrieving subscription: %v", err)
			} else {
				if err := upsertSubscription(app, sub); err != nil {
					log.Printf("Error upserting subscription from checkout: %v", err)
				}
			}
		}

	default:
		log.Printf("Unhandled event type: %s", event.Type)
	}

	return e.JSON(http.StatusOK, map[string]string{"status": "success"})
}

// upsertProduct creates or updates a product record in PocketBase
func upsertProduct(app *pocketbase.PocketBase, stripeProduct *stripe.Product) error {
	collection, err := app.FindCollectionByNameOrId("products")
	if err != nil {
		return err
	}

	// Try to find existing record
	record, err := app.FindFirstRecordByFilter("products", "product_id = {:product_id}", map[string]any{
		"product_id": stripeProduct.ID,
	})

	if err != nil {
		// Create new record
		record = core.NewRecord(collection)
	}

	// Update record fields
	record.Set("product_id", stripeProduct.ID)
	record.Set("active", stripeProduct.Active)
	record.Set("name", stripeProduct.Name)
	record.Set("description", stripeProduct.Description)

	if len(stripeProduct.Images) > 0 {
		record.Set("image", stripeProduct.Images[0])
	}

	// Convert metadata to JSON
	if metadata, err := json.Marshal(stripeProduct.Metadata); err == nil {
		record.Set("metadata", string(metadata))
	}

	return app.Save(record)
}

// upsertPrice creates or updates a price record in PocketBase
func upsertPrice(app *pocketbase.PocketBase, stripePrice *stripe.Price) error {
	collection, err := app.FindCollectionByNameOrId("prices")
	if err != nil {
		return err
	}

	// Try to find existing record
	record, err := app.FindFirstRecordByFilter("prices", "price_id = {:price_id}", map[string]any{
		"price_id": stripePrice.ID,
	})

	if err != nil {
		// Create new record
		record = core.NewRecord(collection)
	}

	// Update record fields
	record.Set("price_id", stripePrice.ID)
	record.Set("product_id", stripePrice.Product.ID)
	record.Set("active", stripePrice.Active)
	record.Set("currency", stripePrice.Currency)
	record.Set("unit_amount", stripePrice.UnitAmount)
	record.Set("type", stripePrice.Type)

	if stripePrice.Recurring != nil {
		record.Set("interval", stripePrice.Recurring.Interval)
		record.Set("interval_count", stripePrice.Recurring.IntervalCount)
	}

	// Convert metadata to JSON
	if metadata, err := json.Marshal(stripePrice.Metadata); err == nil {
		record.Set("metadata", string(metadata))
	}

	return app.Save(record)
}

// upsertSubscription creates or updates a subscription record in PocketBase
func upsertSubscription(app *pocketbase.PocketBase, stripeSub *stripe.Subscription) error {
	collection, err := app.FindCollectionByNameOrId("subscriptions")
	if err != nil {
		return err
	}

	// Get user ID from customer
	userID, err := getUserIDFromCustomer(app, stripeSub.Customer.ID)
	if err != nil {
		return err
	}

	// Try to find existing record
	record, err := app.FindFirstRecordByFilter("subscriptions", "subscription_id = {:subscription_id}", map[string]any{
		"subscription_id": stripeSub.ID,
	})

	if err != nil {
		// Create new record
		record = core.NewRecord(collection)
	}

	// Update record fields
	record.Set("subscription_id", stripeSub.ID)
	record.Set("user_id", userID)
	record.Set("status", stripeSub.Status)
	record.Set("quantity", stripeSub.Items.Data[0].Quantity)
	record.Set("cancel_at_period_end", stripeSub.CancelAtPeriodEnd)
	record.Set("current_period_start", stripeSub.CurrentPeriodStart)
	record.Set("current_period_end", stripeSub.CurrentPeriodEnd)

	if stripeSub.Items != nil && len(stripeSub.Items.Data) > 0 {
		record.Set("price_id", stripeSub.Items.Data[0].Price.ID)
	}

	if stripeSub.EndedAt > 0 {
		record.Set("ended_at", stripeSub.EndedAt)
	}
	if stripeSub.CancelAt > 0 {
		record.Set("cancel_at", stripeSub.CancelAt)
	}
	if stripeSub.CanceledAt > 0 {
		record.Set("canceled_at", stripeSub.CanceledAt)
	}
	if stripeSub.TrialStart > 0 {
		record.Set("trial_start", stripeSub.TrialStart)
	}
	if stripeSub.TrialEnd > 0 {
		record.Set("trial_end", stripeSub.TrialEnd)
	}

	// Convert metadata to JSON
	if metadata, err := json.Marshal(stripeSub.Metadata); err == nil {
		record.Set("metadata", string(metadata))
	}

	return app.Save(record)
}