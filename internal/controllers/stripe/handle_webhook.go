package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/haikali3/gymbara-backend/pkg/utils"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
	"go.uber.org/zap"
)

func handleWebhook(w http.ResponseWriter, req *http.Request) {
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

	endpointSecret := os.Getenv("STRIPE_SECRET_KEY")
	signatureHeader := req.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signatureHeader, endpointSecret)
	if err != nil {
		utils.Logger.Error("Webhook signature verification failed.", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}

	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "customer.subscription.deleted":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		utils.Logger.Info("Subscription deleted.", zap.String("subscription_id", subscription.ID))
		// Then define and call a func to handle the deleted subscription.
		// handleSubscriptionCanceled(subscription)

	case "customer.subscription.updated":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		utils.Logger.Info("Subscription updated.", zap.String("subscription_id", subscription.ID))
		// Then define and call a func to handle the successful attachment of a PaymentMethod.
		// handleSubscriptionUpdated(subscription)

	case "customer.subscription.created":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		utils.Logger.Info("Subscription created.", zap.String("subscription_id", subscription.ID))
		// Then define and call a func to handle the successful attachment of a PaymentMethod.
		// handleSubscriptionCreated(subscription)

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

	default:
		utils.Logger.Warn("Unhandled event type.", zap.String("event_type", string(event.Type)))
	}
	w.WriteHeader(http.StatusOK)
}
