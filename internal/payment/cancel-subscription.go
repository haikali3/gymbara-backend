package payment

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/haikali3/gymbara-backend/pkg/utils"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/subscription"
	"go.uber.org/zap"
)

// CancelSubscriptionRequest represents the JSON payload for cancellation.
type CancelSubscriptionRequest struct {
	SubscriptionID string `json:"subscription_id"`
}

// CancelSubscription cancels a user's Stripe subscription.
func CancelSubscription(w http.ResponseWriter, r *http.Request) {
	var req CancelSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.SubscriptionID == "" {
		http.Error(w, "Subscription ID is required", http.StatusBadRequest)
		return
	}

	// Set your Stripe secret key
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Cancel the subscription at Stripe
	canceledSub, err := subscription.Cancel(req.SubscriptionID, nil)
	if err != nil {
		utils.Logger.Error("Failed to cancel subscription", zap.Error(err))
		http.Error(w, "Failed to cancel subscription", http.StatusInternalServerError)
		return
	}

	// Optionally, update your local database to mark the subscription as cancelled.
	// err = database.MarkSubscriptionCancelled(req.SubscriptionID)
	// if err != nil {
	//     utils.Logger.Error("Failed to update subscription in database", zap.Error(err))
	// }

	// Return a success response with the canceled subscription details.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(canceledSub); err != nil {
		utils.Logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
