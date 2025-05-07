// internal/payment/cancel-subscription.go
package payment

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

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

// CancelSubscription cancels a user's Stripe subscription at period end.
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

	// 1) Schedule cancellation at period end
	updatedSub, err := subscription.Update(
		req.SubscriptionID,
		&stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		},
	)
	if err != nil {
		utils.Logger.Error("Failed to schedule cancellation", zap.Error(err))
		utils.WriteStandardResponse(w, http.StatusInternalServerError, "Failed to schedule cancellation", nil)
		return
	}

	// 2) Persist the expiration_date in our DB
	expiry := time.Unix(updatedSub.CurrentPeriodEnd, 0)
	_, err = database.DB.Exec(
		`UPDATE Subscriptions
				SET expiration_date = $1
			WHERE stripe_subscription_id = $2`,
		expiry, req.SubscriptionID,
	)
	if err != nil {
		utils.Logger.Error("Failed to update subscription expiry in DB", zap.Error(err))
		utils.WriteStandardResponse(w, http.StatusInternalServerError, "Could not update subscription expiry", nil)
		return
	}

	// // 3) (Optional) mark user non-premium immediately in UI
	// _, _ = database.DB.Exec(
	// 	`UPDATE Users
	// 			SET is_premium = FALSE
	// 		WHERE id = (
	// 			SELECT user_id FROM Subscriptions WHERE stripe_subscription_id = $1
	// 		)`,
	// 	req.SubscriptionID,
	// )

	// 4) Build response payload
	nextRenew := expiry.AddDate(0, 1, 0)
	payload := map[string]string{
		"expiration_date": expiry.Format(time.RFC3339),
		"next_renewal":    nextRenew.Format(time.RFC3339),
		"message":         "Your subscription has been cancelled. You remain active until " + expiry.Format("Jan 2, 2006") + ". If you don’t renew, it will auto-renew on " + nextRenew.Format("Jan 2, 2006") + ".",
	}

	utils.Logger.Info("Cancellation scheduled at period end",
		zap.String("subscription_id", req.SubscriptionID),
		zap.Time("expires_on", expiry),
	)
	utils.WriteStandardResponse(w, http.StatusOK, "Subscription cancellation scheduled", payload)
}
