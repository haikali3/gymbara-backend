package main

import (
	"github.com/haikali3/gymbara-backend/database"
	"github.com/haikali3/gymbara-backend/routes"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.Connect()
	routes.RegisterRoutes()

	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
