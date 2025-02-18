package payment

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/billingportal/session"
	"go.uber.org/zap"
)

// 2️⃣ Stripe Customer Portal
// 💡 Purpose: This is where users manage their existing subscriptions (change plans, cancel, update payment method).

// 🔹 How It Works
// A user already subscribed wants to change or cancel their plan.
// They click a "Manage Subscription" button.
// Backend creates a Customer Portal session.
// The user is redirected to Stripe Billing where they can:
// Upgrade/Downgrade plans
// Cancel subscriptions
// Update payment methods

type CustomerPortalRequest struct {
	Email string `json:"email"`
}

// Customer Portal:
// This is where users manage their existing subscriptions (change plans, cancel, update payment method).

// HandleCustomerPortal redirects users to Stripe's customer portal
func HandleCustomerPortal(w http.ResponseWriter, r *http.Request) {
	var req CustomerPortalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Retrieve user’s Stripe customer ID from database
	var customerID string
	// ! where to get stripe customer id, from stripe? why need it? save payment method?
	err := database.DB.QueryRow("SELECT stripe_customer_id FROM Users WHERE email = $1", req.Email).Scan(&customerID)
	if err != nil {
		http.Error(w, "User not found or missing Stripe ID", http.StatusNotFound)
		return
	}

	// Create Customer Portal session
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(os.Getenv("FRONTEND_URL") + "/payment/dashboard"),
	}
	portalSession, err := session.New(params)
	if err != nil {
		utils.Logger.Error("Failed to create Stripe Customer Portal session", zap.Error(err))
		http.Error(w, "Could not create customer portal session", http.StatusInternalServerError)
		return
	}

	// Send URL to frontend
	resp := map[string]string{"url": portalSession.URL}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		utils.Logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
