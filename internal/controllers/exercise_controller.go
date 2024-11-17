package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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

// helper for palceholders
func generatePlaceholders(count int) (string, []interface{}) {
	placeholders := make([]string, count)
	args := make([]interface{}, count)
	for i := 0; i < count; i++ {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = i + 1
	}
	return strings.Join(placeholders, ","), args
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

func GetWorkoutSectionsWithExercises(w http.ResponseWriter, r *http.Request) {
	workoutSectionIDs := r.URL.Query()["workout_section_ids"]
	// ids := strings.Split(workoutSectionIDs, ",")
	if len(workoutSectionIDs) == 0 {
		handleError(w, "Missing workout_section_ids parameter", http.StatusBadRequest, nil)
		return
	}

	placeholders, args := generatePlaceholders(len(workoutSectionIDs))
	for i, id := range workoutSectionIDs {
		args[i] = id
	}

	query := fmt.Sprintf(`
        SELECT
            ws.id AS section_id,
            ws.name AS section_name,
            ws.route AS section_route,
            e.id AS exercise_id,
            e.name AS exercise_name
        FROM
            WorkoutSections ws
        LEFT JOIN
            Exercises e
        ON
            ws.id = e.workout_section_id
        WHERE
            ws.id IN (%s)
        ORDER BY
            ws.id, e.id;
    `, placeholders)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		handleError(w, "Unable to query workout sections and exercises", http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	// Map the results to the desired JSON structure
	sectionsMap := make(map[int]*models.WorkoutSectionWithExercises)
	for rows.Next() {
		var sectionID int
		var sectionName, sectionRoute string
		// var exerciseID *int
		var exercise models.Exercise

		err := rows.Scan(
			&sectionID,
			&sectionName,
			&sectionRoute,
			&exercise.ID,
			&exercise.ExerciseName,
		)
		if err != nil {
			handleError(w, "Unable to scan workout sections and exercises", http.StatusInternalServerError, err)
			return
		}

		if _, exists := sectionsMap[sectionID]; !exists {
			sectionsMap[sectionID] = &models.WorkoutSectionWithExercises{
				ID:        sectionID,
				Name:      sectionName,
				Route:     sectionRoute,
				Exercises: []models.Exercise{},
			}
		}

		// Add the exercise to the section
		if exercise.ID != 0 { // Only add exercises if the ID is valid
			sectionsMap[sectionID].Exercises = append(sectionsMap[sectionID].Exercises, exercise)
		}
	}

	sections := make([]models.WorkoutSectionWithExercises, 0, len(sectionsMap))
	for _, section := range sectionsMap {
		sections = append(sections, *section)
	}

	writeJSONResponse(w, http.StatusOK, sections)
}
