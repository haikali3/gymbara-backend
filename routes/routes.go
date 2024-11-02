package routes

import (
	"net/http"

	"github.com/haikali3/gymbara-backend/controllers"
)

func RegisterRoutes() {
	http.HandleFunc("/exercises", controllers.GetExercises) // Get all exercises

	http.HandleFunc("/workout-sections", controllers.GetWorkoutSections) // Get all workout sections (e.g., Upper, Lower, Full)

	http.HandleFunc("/workout-sections/list", controllers.GetExercisesList) // Get basic list of exercises for a workout section

	http.HandleFunc("/workout-sections/details", controllers.GetExerciseDetails) // Get detailed exercise info for a workout section
}
