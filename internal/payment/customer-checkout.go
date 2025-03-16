package payment

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/haikali3/gymbara-backend/pkg/utils"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"go.uber.org/zap"
)

type SubscriptionRequest struct {
	Email string `json:"email"`
}

func CreateSubscription(w http.ResponseWriter, r *http.Request) {
	// Set the Stripe API Key before calling Stripe
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	// Stripe price ID (set this in your Stripe dashboard) > Product Catalog(Sidebar) > Triple Dot > PriceID
	priceID := os.Getenv("STRIPE_PRICE_ID")
	frontendURL := os.Getenv("FRONTEND_URL")

	// Validate required environment variables
	if stripeKey == "" || priceID == "" || frontendURL == "" {
		utils.Logger.Error("Missing required environment variables", zap.String("StripeKey", stripeKey), zap.String("PriceID", priceID))
		http.Error(w, "Stripe API Key, Price ID, or Frontend URL is missing", http.StatusInternalServerError)
		return
	}

	// Set Stripe API key
	stripe.Key = stripeKey

	utils.Logger.Debug("Using Stripe API Key", zap.String("StripeKey", stripeKey))

	var req SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// 1. Create Stripe Customer
	// var CustomerID string
	// err := database.DB.QueryRow("SELECT stripe_customer_id FROM Users")

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		CustomerEmail:      stripe.String(req.Email),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(frontendURL + "/payment/success"),
		CancelURL:  stripe.String(frontendURL + "/payment/cancel"),
	}

	session, err := session.New(params)
	if err != nil {
		utils.Logger.Error("Failed to create Stripe checkout session", zap.Error(err))
		http.Error(w, "Could not create checkout session", http.StatusInternalServerError)
		return
	}

	// Send checkout session URL to frontend
	resp := map[string]string{"url": session.URL}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		utils.Logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
