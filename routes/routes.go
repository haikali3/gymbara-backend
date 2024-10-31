package routes

import (
	"net/http"

	"github.com/haikali3/gymbara-backend/controllers"
)

func RegisterRoutes() {
	http.HandleFunc("/exercises", controllers.GetExercises)
}
