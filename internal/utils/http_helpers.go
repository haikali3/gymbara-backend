// internal/utils/http_helpers.go
package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
