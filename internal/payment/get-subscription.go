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
	err := database.DB.QueryRow(
		"SELECT stripe_subscription_id FROM Subscriptions WHERE user_id = (SELECT id FROM Users WHERE email = $1) LIMIT 1",
		email,
	).Scan(&subscriptionID)

	if err != nil {
		utils.Logger.Error("Failed to fetch subscription", zap.Error(err))
		http.Error(w, "No active subscription found", http.StatusNotFound)
		return
	}

	utils.Logger.Info("Subscription found", zap.String("subscription_id", subscriptionID))

	// ✅ Return the subscription ID
	response := SubscriptionResponse{
		SubscriptionID: subscriptionID,
		IsActive:       true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
