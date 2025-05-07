// internal/payment/renew_subscription.go
package payment

import (
	"net/http"
	"os"
	"time"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/internal/middleware"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/subscription"
	"go.uber.org/zap"
)

// RenewSubscriptionRequest represents the JSON payload for renewing.
type RenewSubscriptionRequest struct {
	// No fields needed as we get email from auth context
}

// RenewSubscriptionResponse is what we return on success.
type RenewSubscriptionResponse struct {
	Message     string `json:"message"`
	NextRenewal string `json:"next_renewal"`
	URL         string `json:"url,omitempty"`
}

// RenewSubscription either resumes a pending cancel or starts a fresh sub.
func RenewSubscription(w http.ResponseWriter, r *http.Request) {
	// POST request only
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// ✅ Extract email from AuthMiddleware context
	email, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok || email == "" {
		utils.Logger.Error("Failed to retrieve email from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	utils.Logger.Info("Email retrieved in RenewSubscription", zap.String("email", email))

	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	frontendURL := os.Getenv("FRONTEND_URL")
	priceID := os.Getenv("STRIPE_PRICE_ID")

	if stripeKey == "" || frontendURL == "" || priceID == "" {
		utils.Logger.Error("Missing required environment variables")
		http.Error(w, "Missing Stripe config", http.StatusInternalServerError)
		return
	}

	// 1) Look up current subscription details
	var subscriptionID string
	var expiry time.Time
	var customerID string
	err := database.DB.QueryRow(
		`SELECT s.stripe_subscription_id, s.expiration_date, u.stripe_customer_id 
			FROM Subscriptions s
			JOIN Users u ON s.user_id = u.id
			WHERE u.email = $1
			ORDER BY s.expiration_date DESC
			LIMIT 1`,
		email,
	).Scan(&subscriptionID, &expiry, &customerID)
	if err != nil {
		utils.Logger.Error("Failed to fetch subscription details", zap.Error(err))
		utils.WriteStandardResponse(w, http.StatusInternalServerError, "Could not fetch subscription", nil)
		return
	}

	if customerID == "" {
		utils.Logger.Error("No customer ID found for subscription")
		utils.WriteStandardResponse(w, http.StatusInternalServerError, "No customer ID found", nil)
		return
	}

	stripe.Key = stripeKey
	now := time.Now()

	// 2) Decide resume vs. new Checkout
	var updatedSub *stripe.Subscription
	var sess *stripe.CheckoutSession

	if now.Before(expiry) {
		// a) Resume the pending cancellation
		updatedSub, err = subscription.Update(
			subscriptionID,
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
			Customer:           stripe.String(customerID),
			PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
			LineItems: []*stripe.CheckoutSessionLineItemParams{{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			}},
			Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
			SuccessURL: stripe.String(frontendURL + "/payment/success?session_id={CHECKOUT_SESSION_ID}"),
			CancelURL:  stripe.String(frontendURL + "/payment/cancel"),
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
		newExpiry, subscriptionID,
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
		zap.String("subscription_id", subscriptionID),
		zap.Time("next_renewal", nextRenew),
	)
	utils.WriteStandardResponse(w, http.StatusOK, "Subscription renewed", payload)
}
