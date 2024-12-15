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
	"github.com/haikali3/gymbara-backend/internal/utils"
	"go.uber.org/zap"

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
	utils.Logger.Info("Current Working Directory", zap.String("path", wd))

	envFile := fmt.Sprintf("../.env.%s", env) //beware of the path for .env
	utils.Logger.Info("Loading environment file", zap.String("environment", env), zap.String("file", envFile))

	err := godotenv.Load(envFile)
	if err != nil {
		utils.Logger.Fatal("Error loading environment file", zap.String("file", envFile), zap.Error(err))
	}

	// Print a specific message if the environment is set to "production"
	if env == "production" {
		log.Println("\033[33;1m Running in production mode. Ensure sensitive data is secured \033[0m")
	}

	utils.Logger.Info("Environment variables",
		zap.String("environment", os.Getenv("APP_ENV")),
		zap.String("backend_base_url", os.Getenv("BACKEND_BASE_URL")),
		zap.String("frontend_url", os.Getenv("FRONTEND_URL")),
		zap.String("google_client_id", os.Getenv("GOOGLE_CLIENT_ID")),
		zap.String("google_client_secret", os.Getenv("GOOGLE_CLIENT_SECRET")),
	)
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
	// initialize logger
	utils.InitializeLogger()
	defer utils.SyncLogger() // ensure the logger flush on app exit

	env := selectEnvironment()
	os.Setenv("APP_ENV", env)

	loadEnv()

	cfg := config.LoadConfig()

	utils.Logger.Info("Starting Gymbara backend",
		zap.String("environment", os.Getenv("APP_ENV")),
		zap.String("backend_base_url", os.Getenv("BACKEND_BASE_URL")),
		zap.String("frontend_url", os.Getenv("FRONTEND_URL")),
	)

	// Print the loaded configuration for verification
	utils.Logger.Info("Loaded configuration",
		zap.String("environment", os.Getenv("APP_ENV")),
		zap.String("db_host", cfg.DBHost),
		zap.String("server_port", cfg.ServerPort),
	)

	database.Connect(cfg) // Pass config to database connection function
	defer database.Close()

	routes.RegisterRoutes()

	utils.Logger.Info("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		utils.Logger.Fatal("Failed to start server", zap.Error(err))
	}

}
