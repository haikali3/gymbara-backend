package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/haikali3/gymbara-backend/config"
	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/internal/routes"

	"github.com/joho/godotenv"
)

func loadEnv() {
	// default "development"
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
		os.Setenv("APP_ENV", env)
	}

	wd, _ := os.Getwd()
	fmt.Println("Current Working Directory:", wd)

	envFile := fmt.Sprintf("../.env.%s", env) //beware of the path for .env
	log.Printf("Loading environment: %s from file: %s\n", env, envFile)

	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("Error loading %s file", envFile)
	}

	// Print a specific message if the environment is set to "production"
	if env == "production" {
		log.Println("\033[33;1m Running in production mode. Ensure sensitive data is secured \033[0m")
	}

	fmt.Println("Environment:", os.Getenv("APP_ENV"))
	fmt.Println("BACKEND_BASE_URL:", os.Getenv("BACKEND_BASE_URL"))
	fmt.Println("FRONTEND_URL:", os.Getenv("FRONTEND_URL"))
	fmt.Println("GOOGLE_CLIENT_ID:", os.Getenv("GOOGLE_CLIENT_ID"))
	fmt.Println("GOOGLE_CLIENT_SECRET:", os.Getenv("GOOGLE_CLIENT_SECRET"))
}

func selectEnvironment() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Select Environment:")
	fmt.Println("1. Development")
	fmt.Println("2. Production")

	for {
		fmt.Print("Enter number (1 or 2): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			return "development"
		case "2":
			return "production"
		default:
			fmt.Println("Invalid input. Please enter 1 or 2.")
		}
	}
}

func main() {
	env := selectEnvironment()
	os.Setenv("APP_ENV", env)

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
