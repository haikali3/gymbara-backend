package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/haikali3/gymbara-backend/database"
	"github.com/haikali3/gymbara-backend/models"
)

func GetExercises(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(`
        SELECT ed.id, e.name AS ExerciseName, ed.warmup_sets, ed.working_sets, ed.reps, ed.load, ed.rpe, ed.rest_time
        FROM ExerciseDetails ed
        JOIN Exercises e ON ed.exercise_id = e.id
    `)
	if err != nil {
		http.Error(w, "Unable to query database", http.StatusInternalServerError)
		log.Println("Database query error:", err)
		return
	}
	defer rows.Close()

	var exercises []models.Exercise
	for rows.Next() {
		var ex models.Exercise
		if err := rows.Scan(&ex.ID, &ex.ExerciseName, &ex.WarmupSets, &ex.WorkSets, &ex.Reps, &ex.Load, &ex.RPE, &ex.RestTime); err != nil {
			http.Error(w, "Unable to scan database row", http.StatusInternalServerError)
			log.Println("Row scan error:", err)
			return
		}
		exercises = append(exercises, ex)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(exercises); err != nil {
		http.Error(w, "Unable to encode response", http.StatusInternalServerError)
		log.Println("JSON encoding error:", err)
	}
}
