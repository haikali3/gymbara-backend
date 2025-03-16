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

	stripeSubscriptionID := session.Subscription.ID
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
	err = database.DB.QueryRow("SELECT id, stripe_customer_id FROM Users WHERE email = $1", email).Scan(&userID, &stripeCustomerID)

	if err != nil {
		utils.Logger.Warn("User not found, creating new user", zap.String("email", email))

		// Create new user with stripe_customer_id
		err = database.DB.QueryRow(
			"INSERT INTO Users (email, stripe_customer_id, is_premium) VALUES ($1, $2, TRUE) RETURNING id",
			email, stripeCustomerID,
		).Scan(&userID)

		if err != nil {
			utils.Logger.Error("Failed to create new user", zap.Error(err))
			http.Error(w, "User creation failed", http.StatusInternalServerError)
			return
		}
	} else {
		// Update existing user with stripe_customer_id if missing
		if stripeCustomerID == nil {
			_, err = database.DB.Exec("UPDATE Users SET stripe_customer_id = $1, is_premium = TRUE WHERE id = $2", stripeCustomerID, userID)
			if err != nil {
				utils.Logger.Error("Failed to update user with stripe customer ID", zap.Error(err))
				http.Error(w, "Failed to update user", http.StatusInternalServerError)
				return
			}
		}
	}

	// Ensure subscription exists
	var existingSubscriptionID string
	err = database.DB.QueryRow("SELECT stripe_subscription_id FROM Subscriptions WHERE user_id = $1", userID).Scan(&existingSubscriptionID)

	paidDate := time.Now()
	expirationDate := paidDate.AddDate(0, 1, 0)

	// If subscription exists, update it
	if err == nil {
		_, err = database.DB.Exec(
			"UPDATE Subscriptions SET paid_date = $1, expiration_date = $2 WHERE user_id = $3",
			paidDate, expirationDate, userID,
		)
		if err != nil {
			utils.Logger.Error("Failed to update existing subscription", zap.Error(err))
			http.Error(w, "Failed to update subscription", http.StatusInternalServerError)
			return
		}
	} else {
		// If no subscription exists, create a new one
		_, err = database.DB.Exec(
			"INSERT INTO Subscriptions (user_id, paid_date, expiration_date, stripe_subscription_id) VALUES ($1, $2, $3, $4)",
			userID, paidDate, expirationDate, stripeSubscriptionID,
		)
		if err != nil {
			utils.Logger.Error("Failed to insert subscription record", zap.Error(err))
			http.Error(w, "Failed to record subscription", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
