package payment

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/haikali3/gymbara-backend/internal/database"
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

	// ✅ Fetch subscription from Stripe to verify existence
	stripeSub, err := subscription.Get(req.SubscriptionID, nil)
	if err != nil {
		utils.Logger.Error("Subscription not found in Stripe", zap.Error(err))
		http.Error(w, "Subscription ID not found in Stripe", http.StatusNotFound)
		return
	}

	utils.Logger.Info("Subscription found in Stripe", zap.String("subscription_id", stripeSub.ID))

	// Cancel the subscription at Stripe
	canceledSub, err := subscription.Cancel(req.SubscriptionID, nil)
	if err != nil {
		utils.Logger.Error("Failed to cancel subscription", zap.Error(err))
		http.Error(w, "Failed to cancel subscription", http.StatusInternalServerError)
		return
	}

	// ✅ Update database to reflect cancellation
	_, err = database.DB.Exec(
		"UPDATE Subscriptions SET is_active = FALSE WHERE stripe_subscription_id = $1",
		req.SubscriptionID,
	)
	if err != nil {
		// TODO: DEBUG HERE, can cancel subscription but is_premium is on Users table and not on subscriptions table
		utils.Logger.Error("Failed to update subscription status in database", zap.Error(err))
		http.Error(w, "Failed to update subscription in database", http.StatusInternalServerError)
		return
	}

	// Return a success response with the canceled subscription details.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(canceledSub); err != nil {
		utils.Logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	utils.Logger.Info("Subscription canceled successfully", zap.String("subscription_id", req.SubscriptionID))
}
