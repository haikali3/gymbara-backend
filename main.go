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

func loadEnv() {
	// default "development"
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
		os.Setenv("APP_ENV", env)
	}

	var envFile string
	if env == "production" {
		envFile = ".env.production"
	} else {
		envFile = ".env.development"
	}

	log.Printf("Loading environment: %s from file: %s\n", env, envFile)

	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("Error loading %s file", envFile)
	}

	fmt.Println("Environment:", os.Getenv("APP_ENV"))
	fmt.Println("BACKEND_BASE_URL:", os.Getenv("BACKEND_BASE_URL"))
	fmt.Println("FRONTEND_URL:", os.Getenv("FRONTEND_URL"))
	fmt.Println("GOOGLE_CLIENT_ID:", os.Getenv("GOOGLE_CLIENT_ID"))
	fmt.Println("GOOGLE_CLIENT_SECRET:", os.Getenv("GOOGLE_CLIENT_SECRET"))

}

func main() {
	loadEnv() //check if dev or prod .env

	cfg := config.LoadConfig()

	// Print the loaded configuration for verification
	fmt.Printf("Environment: %s\n", os.Getenv("APP_ENV"))
	fmt.Printf("DBHost: %s\n", cfg.DBHost)
	fmt.Printf("ServerPort: %s\n", cfg.ServerPort)

	database.Connect(cfg) // Pass config to database connection function
	routes.RegisterRoutes()

	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
