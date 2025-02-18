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

// 1️⃣ Monthly Subscription (Stripe Checkout)
// 💡 Purpose: This is how users subscribe to a plan.

// 🔹 How It Works
// The user clicks a button to subscribe.
// The backend creates a Stripe Checkout Session.
// The user is redirected to Stripe Checkout.
// Stripe processes the payment and starts the subscription.
// 🔹 Backend Flow
// 1️⃣ User sends a request to create a subscription.
// 2️⃣ Backend creates a Checkout Session.
// 3️⃣ Backend returns the Stripe Checkout URL to the frontend.
// 4️⃣ Frontend redirects the user to complete payment.

// ✅ Example API for Monthly Subscription

type SubscriptionRequest struct {
	Email string `json:"email"`
}

func CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Stripe price ID (set this in your Stripe dashboard)
	priceID := os.Getenv("STRIPE_PRICE_ID")

	// Set your secret key. Remember to switch to your live secret key in production.
	// See your keys here: https://dashboard.stripe.com/apikeys
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
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
		SuccessURL: stripe.String(os.Getenv("FRONTEND_URL") + "/payment/success"),
		CancelURL:  stripe.String(os.Getenv("FRONTEND_URL") + "/payment/cancel"),
	}

	session, err := session.New(params)
	if err != nil {
		utils.Logger.Error("Failed to create Stripe checkout session", zap.Error(err))
		http.Error(w, "Could not create checkout session", http.StatusInternalServerError)
		return
	}

	// Send checkout session URL to frontend
	resp := map[string]string{"url": session.URL}
	json.NewEncoder(w).Encode(resp)

}
