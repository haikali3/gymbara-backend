package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/haikali3/gymbara-backend/config"
	"github.com/haikali3/gymbara-backend/pkg/models"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

var DB *sql.DB

func Connect(cfg *config.Config) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		utils.Logger.Fatal("Failed to connect to database:", zap.String("error", err.Error()))
	}

	// Configure connection pool with values from config
	DB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	DB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	DB.SetConnMaxLifetime(cfg.DBConnMaxLifetime)
	DB.SetConnMaxIdleTime(cfg.DBConnMaxIdleTime)

	utils.Logger.Info("Database connection pool configured",
		zap.Int("maxOpenConns", cfg.DBMaxOpenConns),
		zap.Int("maxIdleConns", cfg.DBMaxIdleConns),
		zap.Duration("connMaxLifetime", cfg.DBConnMaxLifetime),
		zap.Duration("connMaxIdleTime", cfg.DBConnMaxIdleTime),
	)

	//initialize prepared statement
	PrepareStatements()

	if err = DB.Ping(); err != nil {
		utils.Logger.Fatal("Database connection is not alive:", zap.String("error", err.Error()))
	}
	utils.Logger.Info("Database connected successfully.")
}

// inserts or updates a user in the database after OAuth2 login
func StoreUserInDB(user models.GoogleUser, provider string) (int, error) {
	var userID int
	query := `
			INSERT INTO Users (username, email, oauth_provider, oauth_id)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (email) DO UPDATE
			SET username = EXCLUDED.username
			RETURNING id
	`
	err := DB.QueryRow(query, user.Name, user.Email, provider, user.ID).Scan(&userID)
	if err != nil {
		log.Println("Error storing user in DB:", err)
		return 0, err
	}
	log.Println("User stored in DB with ID:", userID)
	return userID, nil
}

func StoreUserWithToken(user models.GoogleUser, accessToken string, refreshToken string) error {
	//TODO: is it normal for this access token will update current row and also other row for column access token?
	utils.Logger.Debug("Updating user with access and refresh tokens",
		zap.String("email", user.Email),
		zap.String("accessToken", accessToken),
		zap.String("refreshToken", refreshToken),
		// zap.String("accessToken", "***"), // Mask sensitive data in logs
		// zap.String("refreshToken", "***"),
	)

	_, err := DB.Exec(`
			INSERT INTO Users (username, email, oauth_provider, oauth_id, access_token, refresh_token)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (email) DO UPDATE
			SET username = EXCLUDED.username, 
					access_token = EXCLUDED.access_token,
					refresh_token = COALESCE(NULLIF(EXCLUDED.refresh_token, ''), Users.refresh_token)
	`, user.Name, user.Email, "google", user.ID, accessToken, refreshToken)

	if err != nil {
		utils.Logger.Error("Failed to store user with token", zap.Error(err))
		return fmt.Errorf("failed to store user with token: %w", err)
	}

	return nil
}

func Close() {
	CloseStatement()

	if DB != nil {
		if err := DB.Close(); err != nil {
			utils.Logger.Error("Failed to close database connection", zap.Error(err))
		} else {
			utils.Logger.Info("Database connection closed successfully.")
		}
	}
}
