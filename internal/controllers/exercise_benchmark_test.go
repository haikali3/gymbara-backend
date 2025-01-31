package controllers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/internal/middleware"
	"github.com/haikali3/gymbara-backend/pkg/models"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func parseDate(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}

var sampleRequest = models.UserExerciseRequest{
	SectionID: 1,
	Exercises: []models.UserExerciseInput{
		{ExerciseID: 1, Reps: 10, Load: 100, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 2, Reps: 12, Load: 80, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 3, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 4, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 5, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 6, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 7, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 8, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 9, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 10, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 11, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 12, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 13, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 14, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 15, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 16, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 17, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 18, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 19, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
		{ExerciseID: 20, Reps: 15, Load: 70, SubmittedAt: parseDate("2025-02-01")},
	},
}

func setupBenchmark() {
	// ✅ Check for `.env` in multiple locations
	envPaths := []string{"../../.env.development", "../.env.development", ".env.development"}

	var err error
	for _, path := range envPaths {
		err = godotenv.Load(path)
		if err == nil {
			log.Printf("✅ Loaded environment variables from %s\n", path)
			break
		}
	}

	// ⚠️ If no `.env` is found, warn but don't exit
	if err != nil {
		log.Println("⚠️ Warning: No .env file found. Using default environment variables.")
	}

	utils.Logger, _ = zap.NewDevelopment()

	// ✅ Read database credentials
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// ✅ Ensure values are not empty (prevent nil database connection)
	if dbUser == "" || dbPassword == "" || dbHost == "" || dbPort == "" || dbName == "" {
		log.Fatal("❌ Missing database credentials in environment variables.")
	}

	// ✅ Initialize database connection
	database.DB, err = sql.Open("postgres",
		"postgres://"+dbUser+":"+dbPassword+"@"+dbHost+":"+dbPort+"/"+dbName+"?sslmode=disable")

	if err != nil {
		log.Fatal("❌ Failed to connect to database in benchmark test:", err)
	}
}

func generateRequestBody(reqData models.UserExerciseRequest) *bytes.Buffer {
	body, _ := json.Marshal(reqData)
	return bytes.NewBuffer(body)
}

func cleanupTestData() {
	_, err := database.DB.Exec(`
		DELETE FROM UserExercisesDetails WHERE submitted_at >= '2025-02-01';
	`)
	if err != nil {
		log.Printf("❌ Failed to clean up test data: %v", err)
	} else {
		log.Println("✅ Test data cleaned up successfully")
	}
}

func BenchmarkSubmitUserExerciseDetails(b *testing.B) {
	setupBenchmark()

	reqBody := generateRequestBody(sampleRequest)
	req, err := http.NewRequest("POST", "/api/submit-exercise", reqBody)
	if err != nil {
		b.Fatalf("❌ Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1) // Mock User ID = 1
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// ✅ Clean test data before running benchmark
	cleanupTestData()

	// Reset timer before benchmarking
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req.Body = io.NopCloser(generateRequestBody(sampleRequest))
		SubmitUserExerciseDetails(w, req)
		w = httptest.NewRecorder()
	}

	// ✅ Clean test data after running benchmark
	cleanupTestData()
}
