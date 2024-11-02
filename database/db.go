package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/haikali3/gymbara-backend/config"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect(cfg *config.Config) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Set database connection pool limits
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	if err = DB.Ping(); err != nil {
		log.Fatal("Database connection is not alive:", err)
	}

	log.Println("Database connected successfully.")
}
