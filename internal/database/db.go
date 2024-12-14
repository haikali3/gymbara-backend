package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/haikali3/gymbara-backend/config"
	"github.com/haikali3/gymbara-backend/internal/models"
	"github.com/haikali3/gymbara-backend/internal/utils"
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

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

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

func StoreUserWithToken(user models.GoogleUser, accessToken string) error {
	//TODO: is it normal for this access token will update current row and also other row for column access token?
	// im not sure...
	utils.Logger.Info("Updating user with access token", zap.String("email", user.Email), zap.String("accessToken", accessToken))

	_, err := DB.Exec(`
			INSERT INTO Users (username, email, oauth_provider, oauth_id, access_token)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (email) DO UPDATE
			SET username = EXCLUDED.username, access_token = EXCLUDED.access_token
	`, user.Name, user.Email, "google", user.ID, accessToken)
	return err
}
