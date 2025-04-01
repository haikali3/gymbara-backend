package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type APIResponse struct {
	Status     string      `json:"status"`     // "success" | "fail" | "error"
	StatusCode int         `json:"statusCode"` // HTTP status code
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
}

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

func WriteStandardResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	status := "success"
	if statusCode >= 400 && statusCode < 500 {
		status = "fail"
	} else if statusCode >= 500 {
		status = "error"
	}

	response := APIResponse{
		Status:     status,
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
