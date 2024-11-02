package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/haikali3/gymbara-backend/database"
	"github.com/haikali3/gymbara-backend/models"
)

/*
	{
		"id": 24,
		"name": "Incline Machine Press",
		"warmup_sets": 1,
		"work_sets": 2,
		"reps": "8-10",
		"load": 0,
		"rpe": "9-10",
		"rest_time": "~2 MINS"
	},
*/
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

func GetWorkoutSections(w http.ResponseWriter, r *http.Request) { //upper lower full
	rows, err := database.DB.Query("SELECT id, name FROM WorkoutSections")
	if err != nil {
		http.Error(w, "Unable to query workout sections", http.StatusInternalServerError)
		log.Println("Database query error:", err)
		return
	}
	defer rows.Close()

	var workoutSections []models.WorkoutSection
	for rows.Next() {
		var workoutSection models.WorkoutSection
		if err := rows.Scan(&workoutSection.ID, &workoutSection.Name); err != nil {
			http.Error(w, "Unable to scan workout sections", http.StatusInternalServerError)
			log.Println("Row scan error:", err)
			return
		}
		workoutSections = append(workoutSections, workoutSection)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workoutSections)
}

// this func is only for initial load of exercises.
func GetExercisesList(w http.ResponseWriter, r *http.Request) {
	workoutSectionID := r.URL.Query().Get("workout_section_id")

	if workoutSectionID == "" {
		http.Error(w, "Missing workout_section_id parameter", http.StatusBadRequest)
		log.Println("Missing workout_section_id parameter")
		return
	}

	rows, err := database.DB.Query(` 
        SELECT e.name, ed.reps, ed.working_sets, ed.load
        FROM Exercises e
        JOIN ExerciseDetails ed ON e.id = ed.exercise_id
        WHERE e.workout_section_id = $1
    `, workoutSectionID)
	if err != nil {
		http.Error(w, "Unable to query exercise basic details", http.StatusInternalServerError)
		log.Println("Database query error:", err)
	}
	defer rows.Close()

	var exerciseDetails []models.ExerciseDetails
	for rows.Next() {
		var detail models.ExerciseDetails
		if err := rows.Scan(&detail.Name, &detail.Reps, &detail.WorkSets, &detail.Load); err != nil {
			http.Error(w, "Unable to scan exercise details", http.StatusInternalServerError)
			log.Println("Row scan error:", err)
			return
		}
		exerciseDetails = append(exerciseDetails, detail)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exerciseDetails)
}

// detail exercise
func GetExerciseDetails(w http.ResponseWriter, r *http.Request) {
	workoutSectionID := r.URL.Query().Get("workout_section_id")

	if workoutSectionID == "" {
		http.Error(w, "Missing workout_section_id parameter", http.StatusBadRequest)
		log.Println("Missing workout_section_id parameter")
		return
	}

	rows, err := database.DB.Query(`
        SELECT e.name, ed.warmup_sets, ed.working_sets, ed.reps, ed.load, ed.rpe, ed.rest_time
		FROM Exercises e
		JOIN ExerciseDetails ed ON e.id = ed.exercise_id
		WHERE e.workout_section_id = $1
    `, workoutSectionID)
	if err != nil {
		http.Error(w, "Unable to query exercise basic details", http.StatusInternalServerError)
		log.Println("Database query error:", err)
	}
	defer rows.Close()

	var exerciseDetails []models.ExerciseDetails
	for rows.Next() {
		var detail models.ExerciseDetails
		if err := rows.Scan(&detail.Name, &detail.WarmupSets, &detail.WorkSets, &detail.Reps, &detail.Load, &detail.RPE, &detail.RestTime); err != nil {
			http.Error(w, "Unable to scan exercise details", http.StatusInternalServerError)
			log.Println("Row scan error:", err)
			return
		}
		exerciseDetails = append(exerciseDetails, detail)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exerciseDetails)
}
