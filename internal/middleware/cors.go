package middleware

import (
	"net/http"
	"os"

	"github.com/haikali3/gymbara-backend/pkg/utils"
	"go.uber.org/zap"
)

// Middleware to handle CORS
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//get frontend url from env variable
		frontendURL := os.Getenv("FRONTEND_URL")

		utils.Logger.Info("Frontend URL initialized",
			zap.String("frontend_url", frontendURL),
		)

		if frontendURL == "" {
			frontendURL = "http://localhost:3000" // Fallback to localhost if not set
		}

		w.Header().Set("Access-Control-Allow-Origin", frontendURL)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		// If it's a preflight request, return without processing further
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass to the next handler
		next.ServeHTTP(w, r)
	})
}
