package middleware

import (
	"net/http"
	"time"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/pkg/utils"
)

func RequirePremium(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Corrected type assertion
		userIDValue := r.Context().Value(UserIDKey)
		userID, ok := userIDValue.(int)
		if !ok {
			utils.HandleError(w, "Invalid user ID in context", http.StatusUnauthorized, nil)
			return
		}

		//query subs
		var expiration time.Time
		err := database.DB.QueryRow("SELECT expiration_date FROM subscriptions WHERE user_id = $1", userID).Scan(&expiration)
		if err != nil || time.Now().After(expiration) {
			utils.HandleError(w, "Access denied: Subscription required", http.StatusPaymentRequired, nil)
			return
		}

		next(w, r)
	}
}
