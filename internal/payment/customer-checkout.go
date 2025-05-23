// internal/payment/customer-checkout.go
package payment

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
	"go.uber.org/zap"
)

type SubscriptionRequest struct {
	Email string `json:"email"`
}

func CreateSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	priceID := os.Getenv("STRIPE_PRICE_ID")
	frontendURL := os.Getenv("FRONTEND_URL")

	if stripeKey == "" || priceID == "" || frontendURL == "" {
		utils.Logger.Error("Missing required environment variables")
		utils.WriteStandardResponse(w, http.StatusInternalServerError, "Missing Stripe config", nil)
		return
	}

	stripe.Key = stripeKey

	var req SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteStandardResponse(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}
	req.Email = strings.ToLower(req.Email)
	utils.Logger.Info("Email received for subscription", zap.String("email", req.Email))

	// Step 1: Look up user in DB
	var userID int
	var stripeCustomerID *string
	err := database.DB.QueryRow("SELECT id, stripe_customer_id FROM Users WHERE email = $1", req.Email).Scan(&userID, &stripeCustomerID)
	if err != nil && err.Error() != "sql: no rows in result set" {
		utils.Logger.Error("DB error fetching user", zap.Error(err))
		utils.WriteStandardResponse(w, http.StatusInternalServerError, "Internal error", nil)
		return
	}

	// Step 2: Use valid Stripe customer ID or create new one
	var customerID string
	validCustomer := false

	if stripeCustomerID != nil && *stripeCustomerID != "" {
		// Try to verify the Stripe customer exists
		_, err := customer.Get(*stripeCustomerID, nil)
		if err == nil {
			customerID = *stripeCustomerID
			validCustomer = true
			utils.Logger.Info("Reusing existing Stripe customer ID", zap.String("customer_id", customerID))
		} else {
			utils.Logger.Warn("Invalid Stripe customer ID, creating new one", zap.String("customer_id", *stripeCustomerID))
		}
	}

	if !validCustomer {
		custParams := &stripe.CustomerParams{Email: stripe.String(req.Email)}
		newCust, err := customer.New(custParams)
		if err != nil {
			utils.Logger.Error("Failed to create Stripe customer", zap.Error(err))
			utils.WriteStandardResponse(w, http.StatusInternalServerError, "Failed to create Stripe customer", nil)
			return
		}
		customerID = newCust.ID

		// Save new customer ID to DB if user exists
		if userID != 0 {
			_, err = database.DB.Exec("UPDATE Users SET stripe_customer_id = $1 WHERE id = $2", customerID, userID)
			if err != nil {
				utils.Logger.Warn("Failed to update stripe_customer_id", zap.Error(err))
			}
		}
	}

	// Step 3: Prevent duplicate subscriptions
	if userID != 0 {
		var existingSub string
		err = database.DB.QueryRow(`
			SELECT stripe_subscription_id
			FROM Subscriptions
			WHERE user_id = $1 AND expiration_date > $2
		`, userID, time.Now()).Scan(&existingSub)

		if err == nil && existingSub != "" {
			utils.Logger.Warn("User already has an active subscription", zap.Int("user_id", userID))
			utils.WriteStandardResponse(w, http.StatusBadRequest, "You already have an active subscription", nil)
			return
		}
	}

	// Step 4: Create checkout session
	params := &stripe.CheckoutSessionParams{
		Customer:           stripe.String(customerID),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(frontendURL + "/payment/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(frontendURL + "/payment/cancel"),
	}

	s, err := session.New(params)
	if err != nil {
		utils.Logger.Error("Failed to create Stripe checkout session", zap.Error(err))
		utils.WriteStandardResponse(w, http.StatusInternalServerError, "Could not create checkout session", nil)
		return
	}

	utils.WriteStandardResponse(w, http.StatusOK, "Checkout session created", map[string]string{"url": s.URL})
}
