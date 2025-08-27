package stripe

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/customer"
)

// getOrCreateStripeCustomer retrieves an existing Stripe customer ID or creates a new one for a company
func getOrCreateStripeCustomer(app *pocketbase.PocketBase, companyID string) (string, error) {
	// First try to get existing customer
	customerID, err := getStripeCustomerID(app, companyID)
	if err == nil {
		return customerID, nil
	}

	// Get company info
	company, err := app.FindRecordById("companies", companyID)
	if err != nil {
		return "", err
	}

	// Get owner info for email
	ownerID := company.GetString("owner_id")
	owner, err := app.FindRecordById("users", ownerID)
	if err != nil {
		return "", err
	}

	// Create new Stripe customer
	params := &stripe.CustomerParams{
		Email: stripe.String(owner.GetString("email")),
		Name:  stripe.String(company.GetString("name")),
		Metadata: map[string]string{
			"company_id": companyID,
			"owner_id":   ownerID,
		},
	}

	stripeCustomer, err := customer.New(params)
	if err != nil {
		return "", err
	}

	// Save customer record in PocketBase
	collection, err := app.FindCollectionByNameOrId("stripe_customers")
	if err != nil {
		return "", err
	}

	record := core.NewRecord(collection)
	record.Set("stripe_customer_id", stripeCustomer.ID)
	record.Set("company_id", companyID)

	if err := app.Save(record); err != nil {
		return "", err
	}

	return stripeCustomer.ID, nil
}

// getStripeCustomerID retrieves the Stripe customer ID for a given company ID
func getStripeCustomerID(app *pocketbase.PocketBase, companyID string) (string, error) {
	record, err := app.FindFirstRecordByFilter("stripe_customers", "company_id = {:company_id}", map[string]any{
		"company_id": companyID,
	})
	if err != nil {
		return "", err
	}

	return record.GetString("stripe_customer_id"), nil
}

// getCompanyIDFromCustomer retrieves the company ID associated with a Stripe customer ID
func getCompanyIDFromCustomer(app *pocketbase.PocketBase, customerID string) (string, error) {
	record, err := app.FindFirstRecordByFilter("stripe_customers", "stripe_customer_id = {:customer_id}", map[string]any{
		"customer_id": customerID,
	})
	if err != nil {
		return "", err
	}

	return record.GetString("company_id"), nil
}

// getCompanyIDFromUserID retrieves the company ID for a given user ID
func getCompanyIDFromUserID(app *pocketbase.PocketBase, userID string) (string, error) {
	record, err := app.FindFirstRecordByFilter("employees", "user_id = {:user_id}", map[string]any{
		"user_id": userID,
	})
	if err != nil {
		return "", err
	}

	return record.GetString("company_id"), nil
}