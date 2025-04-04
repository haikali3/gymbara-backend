package middleware

import (
	"net/http"
	"time"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/pkg/utils"
)

func RequirePremium(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDValue := r.Context().Value(UserIDKey)
		userID, ok := userIDValue.(int)
		if !ok {
			utils.WriteStandardResponse(w, http.StatusUnauthorized, "Invalid user ID in context", nil)
			return
		}

		var expiration time.Time
		err := database.DB.QueryRow(`
			SELECT expiration_date 
			FROM subscriptions 
			WHERE user_id = $1 
			ORDER BY expiration_date DESC 
			LIMIT 1
		`, userID).Scan(&expiration)

		if err != nil {
			utils.WriteStandardResponse(w, http.StatusPaymentRequired, "Access denied: No active subscription", nil)
			return
		}

		if time.Now().After(expiration) {
			utils.WriteStandardResponse(w, http.StatusPaymentRequired, "Access denied: Subscription expired", nil)
			return
		}

		next(w, r)
	}
}
