package payment

import (
	"encoding/json"
	"net/http"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/internal/middleware"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"go.uber.org/zap"
)

// SubscriptionResponse represents the JSON response with the active subscription.
type SubscriptionResponse struct {
	SubscriptionID string `json:"subscription_id"`
	IsActive       bool   `json:"is_active"`
	ExpirationDate string `json:"expiration_date"`
}

// GetSubscription retrieves the active subscription for a user.
func GetSubscription(w http.ResponseWriter, r *http.Request) {
	// ✅ Extract email from AuthMiddleware context
	email, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok || email == "" {
		utils.Logger.Error("Failed to retrieve email from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	utils.Logger.Info("Email retrieved in GetSubscription", zap.String("email", email))

	// ✅ Fetch subscription from database
	var subscriptionID string
	var expirationDate string

	err := database.DB.QueryRow(
		`SELECT stripe_subscription_id, expiration_date 
			FROM Subscriptions 
			WHERE user_id = (SELECT id FROM Users WHERE email = $1) 
			AND expiration_date > NOW()
			ORDER BY expiration_date DESC 
			LIMIT 1`,
		email,
	).Scan(&subscriptionID, &expirationDate)

	if err != nil {
		utils.Logger.Error("Failed to fetch subscription", zap.Error(err))
		http.Error(w, "No active subscription found", http.StatusNotFound)
		return
	}

	utils.Logger.Info("Subscription found",
		zap.String("subscription_id", subscriptionID),
		zap.String("expiration_date", expirationDate))

	// ✅ Return the subscription ID
	response := SubscriptionResponse{
		SubscriptionID: subscriptionID,
		IsActive:       true,
		ExpirationDate: expirationDate,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.Logger.Error("Failed to encode JSON response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
