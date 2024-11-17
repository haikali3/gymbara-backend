package routes

import (
	"net/http"

	oauth "github.com/haikali3/gymbara-backend/internal/auth"
	"github.com/haikali3/gymbara-backend/internal/controllers"
	"github.com/haikali3/gymbara-backend/internal/middleware"
)

func RegisterRoutes() {
	// Workout routes
	http.Handle("/workout-sections", middleware.CORS(http.HandlerFunc(controllers.GetWorkoutSections)))                             // Get all workout sections
	http.Handle("/workout-sections/list", middleware.CORS(http.HandlerFunc(controllers.GetExercisesList)))                          // Get basic list of exercises
	http.Handle("/workout-sections/details", middleware.CORS(http.HandlerFunc(controllers.GetExerciseDetails)))                     // Get detailed exercise info
	http.Handle("/workout-sections/with-exercises", middleware.CORS(http.HandlerFunc(controllers.GetWorkoutSectionsWithExercises))) // Get workout sections with exercises

	//user submit exercise details
	http.Handle("/workout-sections/user-exercise-details", middleware.CORS(http.HandlerFunc(controllers.SubmitUserExerciseDetails))) // Get workout sections with exercises

	// OAuth routes
	http.HandleFunc("/oauth/login", oauth.GoogleLoginHandler)
	http.HandleFunc("/oauth/callback", oauth.GoogleCallbackHandler)

	//fetch user details
	http.Handle("/api/user-info", middleware.CORS(http.HandlerFunc(controllers.GetUserInfoHandler)))
}
