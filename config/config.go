package config

import (
	"os"
	"strconv"
	"time"
)

// Config structure to hold configuration values
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string

	// Database connection pool settings
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration
}

// LoadConfig loads environment variables and returns a Config struct
func LoadConfig() *Config {

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "youruser"),
		DBPassword: getEnv("DB_PASSWORD", "yourpassword"),
		DBName:     getEnv("DB_NAME", "yourdbname"),
		ServerPort: getEnv("SERVER_PORT", "8080"),

		// Connection pool configuration
		DBMaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		DBConnMaxIdleTime: getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", 1*time.Minute),
	}
}

// Helper function to get environment variables or return a default value
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
