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
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	priceID := os.Getenv("STRIPE_PRICE_ID")
	frontendURL := os.Getenv("FRONTEND_URL")

	if stripeKey == "" || priceID == "" || frontendURL == "" {
		utils.Logger.Error("Missing required environment variables", zap.String("StripeKey", stripeKey), zap.String("PriceID", priceID))
		http.Error(w, "Stripe API Key, Price ID, or Frontend URL is missing", http.StatusInternalServerError)
		return
	}

	stripe.Key = stripeKey

	var req SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(req.Email)
	utils.Logger.Info("Email received for subscription", zap.String("email", req.Email))

	// Step 1: Check user in DB
	var userID int
	var stripeCustomerID *string
	err := database.DB.QueryRow("SELECT id, stripe_customer_id FROM Users WHERE email = $1", req.Email).Scan(&userID, &stripeCustomerID)
	if err != nil && err.Error() != "sql: no rows in result set" {
		utils.Logger.Error("Failed to fetch user from DB", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Step 2: Use existing Stripe customer ID or create new one
	var customerID string
	if stripeCustomerID != nil && *stripeCustomerID != "" {
		customerID = *stripeCustomerID
		utils.Logger.Info("Reusing existing Stripe customer ID", zap.String("customer_id", customerID))
	} else {
		// Create new Stripe customer
		customerParams := &stripe.CustomerParams{
			Email: stripe.String(req.Email),
		}
		stripeCustomer, err := customer.New(customerParams)
		if err != nil {
			utils.Logger.Error("Failed to create Stripe customer", zap.Error(err))
			http.Error(w, "Failed to create Stripe customer", http.StatusInternalServerError)
			return
		}
		customerID = stripeCustomer.ID

		// Save new Stripe customer ID if user exists
		if err == nil && userID != 0 {
			_, err = database.DB.Exec("UPDATE Users SET stripe_customer_id = $1 WHERE id = $2", customerID, userID)
			if err != nil {
				utils.Logger.Warn("Failed to update stripe_customer_id in DB", zap.Error(err))
			}
		}
	}

	// Step 3: Check if user already has an active subscription
	if userID != 0 {
		var subID string
		err = database.DB.QueryRow(
			"SELECT stripe_subscription_id FROM Subscriptions WHERE user_id = $1 AND expiration_date > $2",
			userID, time.Now(),
		).Scan(&subID)
		if err == nil && subID != "" {
			utils.Logger.Warn("User already has an active subscription", zap.Int("user_id", userID))
			http.Error(w, "You already have an active subscription", http.StatusBadRequest)
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
		http.Error(w, "Could not create checkout session", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"url": s.URL}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		utils.Logger.Error("Failed to encode checkout session URL", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
