package middleware

import (
	"context"
	"net/http"

	"github.com/haikali3/gymbara-backend/internal/auth"
	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"go.uber.org/zap"
)

// custom context key type to avoid collision
type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
)

// AuthMiddleware extracts the user ID and email from the JWT token
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract and validate access token from cookie
		accessToken, err := r.Cookie("access_token")
		if err != nil {
			utils.Logger.Error("Access token cookie missing", zap.Error(err))
			http.Error(w, "Access token not found", http.StatusUnauthorized)
			return
		}

		utils.Logger.Info("Received access token", zap.String("token", accessToken.Value))

		// Validate access token and get user ID
		userID, err := auth.ValidateToken(accessToken.Value)
		if err != nil {
			utils.Logger.Error("Invalid access token", zap.Error(err))
			http.Error(w, "Invalid access token", http.StatusUnauthorized)
			return
		}

		utils.Logger.Info("Token validated successfully", zap.Int("user_id", userID))

		// Fetch user email from database
		var email string
		err = database.DB.QueryRow("SELECT email FROM Users WHERE id = $1", userID).Scan(&email)
		if err != nil {
			utils.Logger.Error("Failed to fetch user email from database", zap.Error(err))
			http.Error(w, "User email not found", http.StatusUnauthorized)
			return
		}

		utils.Logger.Info("Attaching user email to context", zap.String("email", email))

		// Attach userID and email to request context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, UserEmailKey, email)

		// Pass modified request to next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
