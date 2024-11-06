package routes

import (
	"net/http"

	oauth "github.com/haikali3/gymbara-backend/auth"
	"github.com/haikali3/gymbara-backend/controllers"
)

// Middleware to handle CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow specific origin
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		// Allow specific HTTP methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		// Allow specific headers
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
