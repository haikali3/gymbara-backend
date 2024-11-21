// internal/utils/http_helpers.go
package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/haikali3/gymbara-backend/internal/models"
)

func WriteJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Unable to encode response", http.StatusInternalServerError)
		log.Println("JSON encoding error:", err)
	}
}

func HandleError(w http.ResponseWriter, msg string, status int, err error) {
	http.Error(w, msg, status)
	if err != nil {
		log.Println(msg, err)
	}
}

func GeneratePlaceholders(count int) (string, []interface{}) {
	placeholders := make([]string, count)
	args := make([]interface{}, count)
	for i := 0; i < count; i++ {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = i + 1
	}
	return strings.Join(placeholders, ","), args
}

// HasDuplicateExerciseIDs checks for duplicate exercise IDs in the input.
func HasDuplicateExerciseIDs(exercises []models.UserExerciseInput) (int, bool) {
	seen := make(map[int]bool)
	for _, exercise := range exercises {
		if seen[exercise.ExerciseID] {
			return exercise.ExerciseID, true // Duplicate found
		}
		seen[exercise.ExerciseID] = true
	}
	return 0, false // No duplicates
}
