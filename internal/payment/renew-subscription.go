// internal/payment/renew_subscription.go
package payment

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/subscription"
	"go.uber.org/zap"
)

// RenewSubscriptionRequest represents the JSON payload for renewing.
type RenewSubscriptionRequest struct {
	SubscriptionID string `json:"subscription_id"`
	CustomerID     string `json:"customer_id"`
	PriceID        string `json:"price_id"`
	FrontendURL    string `json:"frontend_url"`
}

// RenewSubscriptionResponse is what we return on success.
type RenewSubscriptionResponse struct {
	Message     string `json:"message"`
	NextRenewal string `json:"next_renewal"`
	URL         string `json:"url,omitempty"`
}

// RenewSubscription either resumes a pending cancel or starts a fresh sub.
func RenewSubscription(w http.ResponseWriter, r *http.Request) {
	var req RenewSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteStandardResponse(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}
	if req.SubscriptionID == "" {
		utils.WriteStandardResponse(w, http.StatusBadRequest, "Subscription ID is required", nil)
		return
	}

	// 1) Look up current expiration_date
	var expiry time.Time
	err := database.DB.QueryRow(
		`SELECT expiration_date FROM Subscriptions WHERE stripe_subscription_id = $1`,
		req.SubscriptionID,
	).Scan(&expiry)
	if err != nil {
		utils.Logger.Error("Failed to fetch subscription expiry", zap.Error(err))
		utils.WriteStandardResponse(w, http.StatusInternalServerError, "Could not fetch subscription", nil)
		return
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	now := time.Now()

	// 2) Decide resume vs. new Checkout
	var updatedSub *stripe.Subscription
	var sess *stripe.CheckoutSession

	if now.Before(expiry) {
		// a) Resume the pending cancellation
		updatedSub, err = subscription.Update(
			req.SubscriptionID,
			&stripe.SubscriptionParams{
				CancelAtPeriodEnd: stripe.Bool(false),
			},
		)
		if err != nil {
			utils.Logger.Error("Failed to resume subscription", zap.Error(err))
			utils.WriteStandardResponse(w, http.StatusInternalServerError, "Could not resume subscription", nil)
			return
		}
	} else {
		// b) Already expired → create a new Checkout Session
		params := &stripe.CheckoutSessionParams{
			Customer:           stripe.String(req.CustomerID),
			PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
			LineItems: []*stripe.CheckoutSessionLineItemParams{{
				Price:    stripe.String(req.PriceID),
				Quantity: stripe.Int64(1),
			}},
			Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
			SuccessURL: stripe.String(req.FrontendURL + "/payment/success?session_id={CHECKOUT_SESSION_ID}"),
			CancelURL:  stripe.String(req.FrontendURL + "/payment/cancel"),
		}
		sess, err = session.New(params)
		if err != nil {
			utils.Logger.Error("Failed to create checkout session", zap.Error(err))
			utils.WriteStandardResponse(w, http.StatusInternalServerError, "Could not create checkout session", nil)
			return
		}

		// Return the redirect URL and exit
		resp := RenewSubscriptionResponse{URL: sess.URL}
		utils.WriteStandardResponse(w, http.StatusOK, "Redirect to Stripe Checkout", resp)
		return
	}

	// 3) Update our DB with the new expiry
	newExpiry := time.Unix(updatedSub.CurrentPeriodEnd, 0)
	_, err = database.DB.Exec(
		`UPDATE Subscriptions SET expiration_date = $1 WHERE stripe_subscription_id = $2`,
		newExpiry, req.SubscriptionID,
	)
	if err != nil {
		utils.Logger.Error("Failed to update subscription expiry in DB", zap.Error(err))
		utils.WriteStandardResponse(w, http.StatusInternalServerError, "Could not update expiry", nil)
		return
	}

	// 4) Return success message
	nextRenew := newExpiry.AddDate(0, 1, 0)
	payload := RenewSubscriptionResponse{
		Message:     "Your subscription has been renewed. It will auto-renew on " + nextRenew.Format("Jan 2, 2006") + ".",
		NextRenewal: nextRenew.Format(time.RFC3339),
	}
	utils.Logger.Info("Subscription renewed",
		zap.String("subscription_id", req.SubscriptionID),
		zap.Time("next_renewal", nextRenew),
	)
	utils.WriteStandardResponse(w, http.StatusOK, "Subscription renewed", payload)
}
