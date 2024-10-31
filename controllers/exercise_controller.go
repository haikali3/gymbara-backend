package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/haikali3/gymbara-backend/database"
	"github.com/haikali3/gymbara-backend/models"
)

func GetExercises(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, name, warmup_sets, work_sets, reps, load, rpe, rest_time FROM ExerciseDetails")
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
			return
		}
		exercises = append(exercises, ex)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exercises)
}
