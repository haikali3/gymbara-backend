package webhook

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
	"go.uber.org/zap"
)

func CheckoutSessionCompleted(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		utils.Logger.Error("Error reading request body", zap.Error(err))
		http.Error(w, "Error reading request body", http.StatusServiceUnavailable)
		return
	}

	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	signHeader := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signHeader, endpointSecret)
	if err != nil {
		utils.Logger.Error("Webhook signature verification failed", zap.Error(err))
		http.Error(w, "Webhook signature verfication failed", http.StatusBadRequest)
		return
	}

	if event.Type != "checkout.session.completed" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		utils.Logger.Error("Error parsing webhook JSON", zap.Error(err))
		http.Error(w, "Error parsing webhook JSON", http.StatusBadRequest)
		return
	}

	stripeCustomerID := session.Customer
	if session.Customer == nil {
		utils.Logger.Error("Customer data is nil")
		http.Error(w, "Customer data is missing", http.StatusBadRequest)
		return
	}

	stripeSubscriptionID := session.Subscription
	if session.Subscription == nil {
		utils.Logger.Error("Subscription data is nil")
		http.Error(w, "Subscription ID is missing", http.StatusBadRequest)
		return
	}

	email := session.CustomerDetails.Email
	if email == "" {
		email = session.Metadata["email"]
	}
	if email == "" {
		utils.Logger.Error("Email not found in session")
		http.Error(w, "Email not provided", http.StatusBadRequest)
		return
	}

	var userID int
	err = database.DB.QueryRow("SELECT id from Users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		utils.Logger.Error("User not found", zap.String("email", email), zap.Error(err))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	paidDate := time.Now()
	expirationDate := paidDate.AddDate(0, 1, 0)

	query := `
		INSERT INTO Subscriptions (user_id, paid_date, expiration_date, stripe_subscription_id, stripe_customer_id)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = database.DB.Exec(query, userID, paidDate, expirationDate, stripeSubscriptionID, stripeCustomerID)
	if err != nil {
		zap.L().Error("Failed to insert subscription record", zap.Error(err))
		http.Error(w, "Failed to record subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
