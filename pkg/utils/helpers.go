// internal/utils/http_helpers.go
package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/haikali3/gymbara-backend/pkg/models"
	"go.uber.org/zap"
)

func WriteJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		Logger.Error("Failed to encode JSON response",
			zap.Error(err),
			zap.Int("status", http.StatusInternalServerError),
		)
		HandleError(w, "Unable to encode response", http.StatusInternalServerError, err)
	}
}

func HandleError(w http.ResponseWriter, msg string, status int, err error) {
	response := map[string]interface{}{
		"error":  msg,
		"status": status,
	}
	if err != nil {
		response["details"] = err.Error()
		Logger.Error(msg,
			zap.Int("status", status),
			zap.Error(err),
		)
	} else {
		Logger.Error(msg,
			zap.Int("status", status))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		Logger.Error("Failed to encode JSON response",
			zap.Error(err),
			zap.Int("status", http.StatusInternalServerError),
		)
	}
}

func GeneratePlaceholders(count int) (string, []interface{}) {
	placeholders := make([]string, count)
	args := make([]interface{}, count)
	for i := 0; i < count; i++ {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = i + 1
	}
	Logger.Debug("Generated placeholders",
		zap.Int("count", count),
		zap.Strings("placeholders", placeholders),
	)
	return strings.Join(placeholders, ","), args
}

func HasDuplicateExerciseIDs(exercises []models.UserExerciseInput) (int, bool) {
	seen := make(map[int]bool)
	for _, exercise := range exercises {
		if seen[exercise.ExerciseID] {
			Logger.Warn("Duplicate exercise ID found",
				zap.Int("exercise_id", exercise.ExerciseID))
			return exercise.ExerciseID, true
		}
		seen[exercise.ExerciseID] = true
	}
	Logger.Debug("No duplicate exercise IDs found",
		zap.Int("total_exercises", len(exercises)),
	)
	return 0, false // No duplicates
}
