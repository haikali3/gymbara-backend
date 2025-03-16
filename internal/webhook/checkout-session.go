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
		http.Error(w, "Webhook signature verification failed", http.StatusBadRequest)
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

	// ✅ Ensure we have a valid Stripe Customer ID
	stripeCustomerID := session.Customer.ID
	if stripeCustomerID == "" && session.Metadata != nil {
		stripeCustomerID = session.Metadata["customer_id"]
	}
	if stripeCustomerID == "" {
		utils.Logger.Error("Stripe customer ID is missing", zap.Any("session", session))
		http.Error(w, "Stripe customer ID is missing", http.StatusBadRequest)
		return
	}

	// ✅ Ensure we have a valid Subscription ID
	var stripeSubscriptionID string
	if session.Subscription != nil {
		stripeSubscriptionID = session.Subscription.ID
	} else {
		utils.Logger.Error("Subscription data is nil", zap.Any("session", session))
		http.Error(w, "Subscription ID is missing", http.StatusBadRequest)
		return
	}

	// ✅ Ensure we have a valid Email
	var email string
	if session.CustomerDetails != nil {
		email = session.CustomerDetails.Email
	}
	if email == "" && session.Metadata != nil {
		email = session.Metadata["email"]
	}
	if email == "" {
		utils.Logger.Error("Email not found in session", zap.Any("session", session))
		http.Error(w, "Email not provided", http.StatusBadRequest)
		return
	}

	// ✅ Ensure user exists and update stripe_customer_id if necessary
	var userID int
	var existingStripeCustomerID *string

	err = database.DB.QueryRow("SELECT id, stripe_customer_id FROM Users WHERE email = $1", email).Scan(&userID, &existingStripeCustomerID)

	if err == nil {
		// ✅ User exists, update stripe_customer_id if missing
		if existingStripeCustomerID == nil || *existingStripeCustomerID == "" {
			_, err = database.DB.Exec("UPDATE Users SET stripe_customer_id = $1 WHERE id = $2", stripeCustomerID, userID)
			if err != nil {
				utils.Logger.Error("Failed to update stripe_customer_id", zap.Error(err))
				http.Error(w, "Failed to update user", http.StatusInternalServerError)
				return
			}
			utils.Logger.Info("Updated stripe_customer_id for existing user", zap.Int("userID", userID))
		}
	} else if err.Error() == "sql: no rows in result set" {
		// ✅ User not found, create a new one
		err = database.DB.QueryRow(
			"INSERT INTO Users (email, stripe_customer_id, is_premium) VALUES ($1, $2, TRUE) RETURNING id",
			email, stripeCustomerID,
		).Scan(&userID)

		if err != nil {
			utils.Logger.Error("Failed to create new user", zap.Error(err))
			http.Error(w, "User creation failed", http.StatusInternalServerError)
			return
		}
		utils.Logger.Info("Created new user with stripe_customer_id", zap.String("email", email), zap.Int("userID", userID))
	} else {
		// 🚨 Handle unexpected DB errors
		utils.Logger.Error("Database query failed", zap.Error(err))
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// ✅ Ensure subscription exists
	var existingSubscriptionID string
	err = database.DB.QueryRow("SELECT stripe_subscription_id FROM Subscriptions WHERE user_id = $1", userID).Scan(&existingSubscriptionID)

	paidDate := time.Now()
	expirationDate := paidDate.AddDate(0, 1, 0)

	// ✅ If subscription exists, update it
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
		utils.Logger.Info("Updated existing subscription", zap.Int("userID", userID))
	} else {
		// ✅ If no subscription exists, create a new one
		_, err = database.DB.Exec(
			"INSERT INTO Subscriptions (user_id, paid_date, expiration_date, stripe_subscription_id) VALUES ($1, $2, $3, $4)",
			userID, paidDate, expirationDate, stripeSubscriptionID,
		)
		if err != nil {
			utils.Logger.Error("Failed to insert subscription record", zap.Error(err))
			http.Error(w, "Failed to record subscription", http.StatusInternalServerError)
			return
		}
		utils.Logger.Info("Inserted new subscription", zap.Int("userID", userID), zap.String("stripeSubscriptionID", stripeSubscriptionID))
	}

	// ✅ Debugging Log: Check if `stripe_customer_id` was stored properly
	var testStripeCustomerID string
	err = database.DB.QueryRow("SELECT stripe_customer_id FROM Users WHERE id = $1", userID).Scan(&testStripeCustomerID)
	utils.Logger.Info("Database check - stripe_customer_id after update", zap.String("stripe_customer_id", testStripeCustomerID))

	w.WriteHeader(http.StatusOK)
}
