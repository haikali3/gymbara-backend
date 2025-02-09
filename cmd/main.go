package main

import (
	"log"
	"net/http"
	"os"

	"github.com/haikali3/gymbara-backend/config"
	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/internal/routes"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"go.uber.org/zap"

	"github.com/joho/godotenv"
)

func loadEnv() {
	// default "development"
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
		if err := os.Setenv("APP_ENV", env); err != nil {
			utils.Logger.Fatal("Error setting environment variable", zap.Error(err))
		}
	}

	wd, _ := os.Getwd()
	utils.Logger.Info("Current Working Directory", zap.String("path", wd))

	envFile := ".env." + env
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

func main() {
	// initialize logger
	utils.InitializeLogger()
	defer func() {
		if err := utils.SyncLogger(); err != nil {
			utils.Logger.Error("Failed to sync logger", zap.Error(err))
		}
	}()
	utils.Logger.Info("Logger initialized successfully with colors!")

	// Get APP_ENV from .air.toml (or default to "development")
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
		if err := os.Setenv("APP_ENV", env); err != nil {
			utils.Logger.Fatal("Error setting environment variable", zap.Error(err))
		}
	}

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
