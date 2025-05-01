// internal/payment/cancel-subscription.go
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
		utils.WriteStandardResponse(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}
	if req.SubscriptionID == "" {
		utils.WriteStandardResponse(w, http.StatusBadRequest, "Subscription ID is required", nil)
		return
	}

	// Initialize Stripe
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Verify subscription exists in Stripe
	stripeSub, err := subscription.Get(req.SubscriptionID, nil)
	if err != nil {
		utils.Logger.Error("Subscription not found in Stripe", zap.Error(err))
		utils.WriteStandardResponse(w, http.StatusNotFound, "Subscription ID not found in Stripe", nil)
		return
	}
	utils.Logger.Info("Subscription found in Stripe", zap.String("subscription_id", stripeSub.ID))

	// Cancel in Stripe
	canceledSub, err := subscription.Cancel(req.SubscriptionID, nil)
	if err != nil {
		utils.Logger.Error("Failed to cancel subscription", zap.Error(err))
		utils.WriteStandardResponse(w, http.StatusInternalServerError, "Failed to cancel subscription", nil)
		return
	}

	// Lookup the user in our DB
	var userID string
	err = database.DB.QueryRow(
		"SELECT user_id FROM Subscriptions WHERE stripe_subscription_id = $1",
		req.SubscriptionID,
	).Scan(&userID)
	if err != nil {
		utils.Logger.Error("Failed to get user ID from subscription", zap.Error(err))
		utils.WriteStandardResponse(w, http.StatusInternalServerError, "Could not find user for this subscription", nil)
		return
	}

	// Mark user as non-premium
	_, err = database.DB.Exec(
		"UPDATE Users SET is_premium = FALSE WHERE id = $1",
		userID,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			utils.Logger.Warn("Subscription ID not found in database", zap.String("subscription_id", req.SubscriptionID))
			utils.WriteStandardResponse(w, http.StatusNotFound, "Subscription not found in database", nil)
			return
		}
		utils.Logger.Error("Failed to update user status", zap.Error(err))
		utils.WriteStandardResponse(w, http.StatusInternalServerError, "Could not update user subscription status", nil)
		return
	}

	// Success response
	utils.Logger.Info("Subscription canceled successfully", zap.String("subscription_id", req.SubscriptionID))
	utils.WriteStandardResponse(w, http.StatusOK, "Subscription cancelled successfully", canceledSub)
}
