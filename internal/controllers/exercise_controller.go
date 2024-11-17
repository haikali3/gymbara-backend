package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/internal/models"
)

// Helper to write JSON responses
func writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Unable to encode response", http.StatusInternalServerError)
		log.Println("JSON encoding error:", err)
	}
}

// Helper for handling errors
func handleError(w http.ResponseWriter, msg string, status int, err error) {
	http.Error(w, msg, status)
	if err != nil {
		log.Println(msg, err)
	}
}

// Get workout sections
func GetWorkoutSections(w http.ResponseWriter, r *http.Request) {
	query := "SELECT id, name, route FROM WorkoutSections"

	rows, err := database.DB.Query(query)
	if err != nil {
		handleError(w, "Unable to query workout sections", http.StatusInternalServerError, err)
		return
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	var workoutSections []models.WorkoutSection
	for rows.Next() {
		var workoutSection models.WorkoutSection
		if err := rows.Scan(&workoutSection.ID, &workoutSection.Name, &workoutSection.Route); err != nil {
			handleError(w, "Unable to scan workout sections", http.StatusInternalServerError, err)
			return
		}
		workoutSections = append(workoutSections, workoutSection)
	}
	writeJSONResponse(w, http.StatusOK, workoutSections)
}

// Get exercises for initial load
func GetExercisesList(w http.ResponseWriter, r *http.Request) {
	workoutSectionID := r.URL.Query().Get("workout_section_id")

	if workoutSectionID == "" {
		handleError(w, "Missing workout_section_id parameter", http.StatusBadRequest, nil)
		return
	}

	query := `
        SELECT e.name, ed.reps, ed.working_sets, ed.load
        FROM Exercises e
        JOIN ExerciseDetails ed ON e.id = ed.exercise_id
        WHERE e.workout_section_id = $1
    `
	rows, err := database.DB.Query(query, workoutSectionID)
	if err != nil {
		handleError(w, "Unable to query exercise basic details", http.StatusInternalServerError, err)
		return
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	var exerciseList []models.ExerciseDetails
	for rows.Next() {
		var detail models.ExerciseDetails
		if err := rows.Scan(&detail.Name, &detail.Reps, &detail.WorkSets, &detail.Load); err != nil {
			handleError(w, "Unable to scan exercise details", http.StatusInternalServerError, err)
			return
		}
		exerciseList = append(exerciseList, detail)
	}
	writeJSONResponse(w, http.StatusOK, exerciseList)
}

// Get detailed exercise information
func GetExerciseDetails(w http.ResponseWriter, r *http.Request) {
	workoutSectionID := r.URL.Query().Get("workout_section_id")

	if workoutSectionID == "" {
		handleError(w, "Missing workout_section_id parameter", http.StatusBadRequest, nil)
		return
	}

	query := `
        SELECT e.name, ed.warmup_sets, ed.working_sets, ed.reps, ed.load, ed.rpe, ed.rest_time
		FROM Exercises e
		JOIN ExerciseDetails ed ON e.id = ed.exercise_id
		WHERE e.workout_section_id = $1
    `
	rows, err := database.DB.Query(query, workoutSectionID)
	if err != nil {
		handleError(w, "Unable to query exercise details", http.StatusInternalServerError, err)
		return
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	var exerciseDetails []models.ExerciseDetails
	for rows.Next() {
		var detail models.ExerciseDetails
		if err := rows.Scan(&detail.Name, &detail.WarmupSets, &detail.WorkSets, &detail.Reps, &detail.Load, &detail.RPE, &detail.RestTime); err != nil {
			handleError(w, "Unable to scan exercise details", http.StatusInternalServerError, err)
			return
		}
		exerciseDetails = append(exerciseDetails, detail)
	}
	writeJSONResponse(w, http.StatusOK, exerciseDetails)
}
