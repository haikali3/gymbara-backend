package routes

import (
	"net/http"
	"time"

	oauth "github.com/haikali3/gymbara-backend/internal/auth"
	"github.com/haikali3/gymbara-backend/internal/controllers"
	"github.com/haikali3/gymbara-backend/internal/middleware"
)

func RegisterRoutes() {
	const maxRequests = 10
	const duration = time.Second

	secureHandler := func(handler http.HandlerFunc) http.Handler {
		rateLimitedHandler := middleware.RateLimit(maxRequests, duration)
		corsHandler := middleware.CORS
		authHandler := middleware.AuthMiddleware

		return rateLimitedHandler(corsHandler(authHandler(handler)))
	}

	// Workout routes
	http.Handle("/workout-sections", secureHandler(controllers.GetWorkoutSections))
	http.Handle("/workout-sections/list", secureHandler(controllers.GetExercisesList))
	http.Handle("/workout-sections/details", secureHandler(controllers.GetExerciseDetails))
	http.Handle("/workout-sections/exercises", secureHandler(controllers.GetWorkoutSectionsWithExercises))

	// User submit exercise details
	http.Handle("/workout-sections/user-exercise-details", secureHandler(controllers.SubmitUserExerciseDetails))
	// Fetch user submitted exercise detail
	http.Handle("/user/progress", secureHandler(controllers.GetUserProgress))

	//fetch user details
	http.Handle("/api/user-info", secureHandler(controllers.GetUserInfoHandler))

	// OAuth routes
	http.Handle("/oauth/login", middleware.RateLimit(maxRequests, duration)(http.HandlerFunc(oauth.GoogleLoginHandler)))
	http.Handle("/oauth/callback", middleware.RateLimit(maxRequests, duration)(http.HandlerFunc(oauth.GoogleCallbackHandler)))

	// stripe
	// http.Handle("/create-checkout-session", middleware.RateLimit(maxRequests, duration)(http.HandlerFunc(controllers.CreateCheckoutSession)))
}
