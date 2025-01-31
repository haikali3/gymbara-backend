package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/haikali3/gymbara-backend/pkg/models"
)

// Sample test request body
var sampleRequest = models.UserExerciseRequest{
	SectionID: 1,
	Exercises: []models.UserExerciseInput{
		{ExerciseID: 101, Reps: 10, Load: 100},
		{ExerciseID: 102, Reps: 12, Load: 80},
	},
}

// Helper function to convert request to JSON
func generateRequestBody(reqData models.UserExerciseRequest) *bytes.Buffer {
	body, _ := json.Marshal(reqData)
	return bytes.NewBuffer(body)
}

// Benchmark for SubmitUserExerciseDetails
func BenchmarkSubmitUserExerciseDetails(b *testing.B) {
	reqBody := generateRequestBody(sampleRequest)

	// Create a test HTTP request
	req, err := http.NewRequest("POST", "/api/submit-exercise", reqBody)
	if err != nil {
		b.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to capture response
	w := httptest.NewRecorder()

	// Reset the timer before benchmarking
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		SubmitUserExerciseDetails(w, req)
	}
}
