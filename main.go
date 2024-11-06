package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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

	fmt.Println("GOOGLE_CLIENT_ID:", os.Getenv("GOOGLE_CLIENT_ID"))
	fmt.Println("GOOGLE_CLIENT_SECRET:", os.Getenv("GOOGLE_CLIENT_SECRET"))

	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
