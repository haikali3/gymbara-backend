package main

import (
	"log"
	"net/http"

	oauth "github.com/haikali3/gymbara-backend/auth"
	"github.com/haikali3/gymbara-backend/config"
	"github.com/haikali3/gymbara-backend/database"
	"github.com/haikali3/gymbara-backend/routes"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := config.LoadConfig() // Load configuration

	database.Connect(cfg) // Pass config to database connection function

	routes.RegisterRoutes()

	log.Println("Starting server on :8080...")
	http.HandleFunc("/oauth/login", oauth.GoogleLoginHandler)
	http.HandleFunc("/oauth/callback", oauth.GoogleCallbackHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
