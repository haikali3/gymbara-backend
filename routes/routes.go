package routes

import (
	"net/http"
	"os"

	oauth "github.com/haikali3/gymbara-backend/auth"
	"github.com/haikali3/gymbara-backend/controllers"
)

// Middleware to handle CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//get frontend url from env variable
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:3000" // Fallback to localhost if not set
		}
		w.Header().Set("Access-Control-Allow-Origin", frontendURL)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		// If it's a preflight request, return without processing further
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass to the next handler
		next.ServeHTTP(w, r)
	})
}

func RegisterRoutes() {
	// Workout routes
	http.Handle("/workout-sections", corsMiddleware(http.HandlerFunc(controllers.GetWorkoutSections)))         // Get all workout sections
	http.Handle("/workout-sections/list", corsMiddleware(http.HandlerFunc(controllers.GetExercisesList)))      // Get basic list of exercises
	http.Handle("/workout-sections/details", corsMiddleware(http.HandlerFunc(controllers.GetExerciseDetails))) // Get detailed exercise info

	// OAuth routes
	http.HandleFunc("/oauth/login", oauth.GoogleLoginHandler)
	http.HandleFunc("/oauth/callback", oauth.GoogleCallbackHandler)
	http.HandleFunc("/oauth/logout", oauth.GoogleLogoutHandler)
}
