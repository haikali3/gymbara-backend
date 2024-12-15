// Desc: This file contains the middleware for authenticating the user
package middleware

import (
	"context"
	"net/http"

	"github.com/haikali3/gymbara-backend/internal/auth"
	"github.com/haikali3/gymbara-backend/internal/utils"
	"go.uber.org/zap"
)

// custom context key type to avoid collision
type contextKey string

const UserIDKey contextKey = "user_id"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract and validate access token from cookie
		accessToken, err := r.Cookie("access_token")
		if err != nil {
			utils.Logger.Error("Access token cookie missing", zap.Error(err))
			utils.HandleError(w, "Access token not found", http.StatusUnauthorized, err)
			return
		}

		utils.Logger.Info("Access token found", zap.String("cookie_name", accessToken.Value))

		// validate access token
		userID, err := auth.ValidateToken(accessToken.Value)
		if err != nil {
			utils.Logger.Error("Invalid access token", zap.Error(err))
			utils.HandleError(w, "Invalid access token", http.StatusUnauthorized, nil)
			return
		}

		// attach userID to request ctx
		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		//pass modified request to next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
