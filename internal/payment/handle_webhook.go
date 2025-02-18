package payment

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
	"go.uber.org/zap"
)

func HandleWebhook(w http.ResponseWriter, req *http.Request) {
	const MaxBodyBytes = int64(65536)
	bodyReader := http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusServiceUnavailable)
		return
	}

	// Replace this endpoint secret with your endpoint's unique secret
	// If you are testing with the CLI, find the secret by running 'stripe listen'
	// If you are using an endpoint defined with the API or dashboard, look in your webhook settings
	// at https://dashboard.stripe.com/webhooks

	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	signatureHeader := req.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signatureHeader, endpointSecret)
	if err != nil {
		utils.Logger.Error("Webhook signature verification failed.", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}

	switch event.Type {
	case "customer.subscription.created", "customer.subscription.updated":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			http.Error(w, "Error parsing webhook JSON", http.StatusBadRequest)
			return
		}

		// Calculate expiration date (1 month later)
		expirationDate := time.Now().AddDate(0, 1, 0).Format("2006-01-02")

		// Update database with new subscription details
		_, err = database.DB.Exec(`
			UPDATE Subscriptions 
			SET stripe_subscription_id = $1, paid_date = NOW(), expiration_date = $2 
			WHERE user_id = (SELECT id FROM Users WHERE email = $3)
		`, subscription.ID, expirationDate, subscription.Customer.Email)

		if err != nil {
			log.Printf("DB error: %v", err)
		}

		// Also set is_premium = TRUE in Users table
		_, err = database.DB.Exec(`
			UPDATE Users SET is_premium = TRUE WHERE email = $1
		`, subscription.Customer.Email)

		if err != nil {
			utils.Logger.Error("DB error.", zap.Error(err))
		}
		utils.Logger.Info("Subscription created.", zap.String("subscription_id", subscription.ID))
		// Then define and call a func to handle the successful attachment of a PaymentMethod.
		// handleSubscriptionCreated(subscription)

	// Unmarshal the event data into an appropriate struct depending on its Type

	case "customer.subscription.trial_will_end":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			utils.Logger.Error("Error parsing webhook JSON.", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		utils.Logger.Info("Subscription trial will end.", zap.String("subscription_id", subscription.ID))
		// Then define and call a func to handle the successful attachment of a PaymentMethod.
		// handleSubscriptionTrialWillEnd(subscription)

	case "entitlements.active_entitlement_summary.updated":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			utils.Logger.Error("Error parsing webhook JSON.", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		utils.Logger.Info("Active entitlement summary updated.", zap.String("subscription_id", subscription.ID))
		// Then define and call a func to handle active entitlement summary updated.
		// handleEntitlementUpdated(subscription)

	case "customer.subscription.deleted":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			http.Error(w, "Error parsing webhook JSON", http.StatusBadRequest)
			return
		}

		// Cancel subscription and set user as non-premium
		_, err = database.DB.Exec(`
			UPDATE Users SET is_premium = FALSE WHERE id = (SELECT user_id FROM Subscriptions WHERE stripe_subscription_id = $1)
		`, subscription.ID)
		if err != nil {
			log.Printf("DB error: %v", err)
		}

		_, err = database.DB.Exec(`
			UPDATE Subscriptions SET expiration_date = NOW() WHERE stripe_subscription_id = $1
		`, subscription.ID)
		if err != nil {
			log.Printf("DB error: %v", err)
		}

		if err != nil {
			log.Printf("DB error: %v", err)
		}
		utils.Logger.Info("Subscription deleted.", zap.String("subscription_id", subscription.ID))
		// Then define and call a func to handle the deleted subscription.
		// handleSubscriptionCanceled(subscription)

	default:
		utils.Logger.Warn("Unhandled event type.", zap.String("event_type", string(event.Type)))
	}
	w.WriteHeader(http.StatusOK)
}
