package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/internal/models"
	"github.com/haikali3/gymbara-backend/internal/utils"
)

// Get workout sections
func GetWorkoutSections(w http.ResponseWriter, r *http.Request) {
	query := "SELECT id, name, route FROM WorkoutSections"

	rows, err := database.DB.Query(query)
	if err != nil {
		utils.HandleError(w, "Unable to query workout sections", http.StatusInternalServerError, err)
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
			utils.HandleError(w, "Unable to scan workout sections", http.StatusInternalServerError, err)
			return
		}
		workoutSections = append(workoutSections, workoutSection)
	}
	utils.WriteJSONResponse(w, http.StatusOK, workoutSections)
}

// Get exercises for initial load
func GetExercisesList(w http.ResponseWriter, r *http.Request) {
	workoutSectionID := r.URL.Query().Get("workout_section_id")

	if workoutSectionID == "" {
		utils.HandleError(w, "Missing workout_section_id parameter", http.StatusBadRequest, nil)
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
		utils.HandleError(w, "Unable to query exercise basic details", http.StatusInternalServerError, err)
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
			utils.HandleError(w, "Unable to scan exercise details", http.StatusInternalServerError, err)
			return
		}
		exerciseList = append(exerciseList, detail)
	}
	utils.WriteJSONResponse(w, http.StatusOK, exerciseList)
}

// Get detailed exercise information
func GetExerciseDetails(w http.ResponseWriter, r *http.Request) {
	workoutSectionID := r.URL.Query().Get("workout_section_id")

	if workoutSectionID == "" {
		utils.HandleError(w, "Missing workout_section_id parameter", http.StatusBadRequest, nil)
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
		utils.HandleError(w, "Unable to query exercise details", http.StatusInternalServerError, err)
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
			utils.HandleError(w, "Unable to scan exercise details", http.StatusInternalServerError, err)
			return
		}
		exerciseDetails = append(exerciseDetails, detail)
	}
	utils.WriteJSONResponse(w, http.StatusOK, exerciseDetails)
}

func GetWorkoutSectionsWithExercises(w http.ResponseWriter, r *http.Request) {
	workoutSectionIDs := r.URL.Query()["workout_section_ids"]
	if len(workoutSectionIDs) == 0 {
		utils.HandleError(w, "Missing workout_section_ids parameter", http.StatusBadRequest, nil)
		return
	}

	placeholders, args := utils.GeneratePlaceholders(len(workoutSectionIDs))
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
		utils.HandleError(w, fmt.Sprintf("Unable to query workout sections and exercises for workout_section_ids: %v. Query: %s", workoutSectionIDs, query), http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	// Map the results to the desired JSON structure
	sectionsMap := make(map[int]*models.WorkoutSectionWithExercises)
	for rows.Next() {
		var sectionID int
		var sectionName, sectionRoute string
		var exercise models.ExerciseMinimal

		err := rows.Scan(
			&sectionID,
			&sectionName,
			&sectionRoute,
			&exercise.ID,
			&exercise.ExerciseName,
		)
		if err != nil {
			utils.HandleError(w, fmt.Sprintf("Error scanning row for section_id: %d. Partial data: SectionName=%s, Route=%s", sectionID, sectionName, sectionRoute), http.StatusInternalServerError, err)
			return
		}

		if _, exists := sectionsMap[sectionID]; !exists {
			sectionsMap[sectionID] = &models.WorkoutSectionWithExercises{
				ID:        sectionID,
				Name:      sectionName,
				Route:     sectionRoute,
				Exercises: []models.ExerciseMinimal{},
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

	utils.WriteJSONResponse(w, http.StatusOK, sections)
}

func SubmitUserExerciseDetails(w http.ResponseWriter, r *http.Request) {
	var request models.UserExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.HandleError(w, "Invalid request body", http.StatusBadRequest, err)
		return
	}

	// validate the request
	if request.SectionID == 0 || len(request.Exercises) == 0 {
		utils.HandleError(w, "Missing required fields: section_id or exercises", http.StatusBadRequest, nil)
		return
	}

	// Check for duplicate exercise IDs
	if duplicateExerciseID, found := utils.HasDuplicateExerciseIDs(request.Exercises); found {
		utils.HandleError(w, fmt.Sprintf("Duplicate(exercise_id: %d) found in request", duplicateExerciseID), http.StatusBadRequest, nil)
		return
	}

	// begin db transaction
	tx, err := database.DB.Begin()
	if err != nil {
		utils.HandleError(w, "Failed to start database transaction", http.StatusInternalServerError, err)
		return
	}

	// rollback or commit transaction appropriately
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	//validate user existence using OAuth email or ID
	var userID int
	err = tx.QueryRow(`
		SELECT id FROM Users WHERE email = $1
	`, request.UserEmail).Scan(&userID)
	if err != nil {
		utils.HandleError(w, "User not found. Please log in.", http.StatusUnauthorized, err)
		return
	}

	//check UserWorkout exist, create if not
	var userWorkoutID int
	err = tx.QueryRow(`
		INSERT INTO UserWorkouts (user_id, section_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, section_id) DO UPDATE
		SET user_id = EXCLUDED.user_id
		RETURNING id
	`, userID, request.SectionID).Scan(&userWorkoutID)
	if err != nil {
		utils.HandleError(w, fmt.Sprintf("Failed to insert or update user workout for user_id: %d, section_id: %d", userID, request.SectionID), http.StatusInternalServerError, err)
		return
	}

	// validate each exercise ID and  insert or update custom details
	for _, exercise := range request.Exercises {
		//validate input
		if exercise.Reps <= 0 || exercise.Load <= 0 {
			utils.HandleError(w, fmt.Sprintf("Invalid reps or load for exercise_id: %d", exercise.ExerciseID), http.StatusBadRequest, nil)
			return
		}

		//check exercise id exist
		var exerciseExists bool
		err = tx.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM Exercises WHERE id = $1)
		`, exercise.ExerciseID).Scan(&exerciseExists)
		if err != nil || !exerciseExists {
			utils.HandleError(w, fmt.Sprintf("Invalid exercise_id: %d doesn't exist", exercise.ExerciseID), http.StatusBadRequest, nil)
			return
		}

		//execute upsert query
		_, err = tx.Exec(`
		INSERT INTO UserExercisesDetails (user_workout_id, exercise_id, custom_reps, custom_load)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_workout_id, exercise_id) DO UPDATE
		SET custom_reps = $3, custom_load = $4
	`, userWorkoutID, exercise.ExerciseID, exercise.Reps, exercise.Load)
		if err != nil {
			utils.HandleError(w, "failed to insert or update user exercise details", http.StatusInternalServerError, err)
			return
		}
	}

	var updatedExercises []models.UserExerciseInput
	for _, exercise := range request.Exercises {
		updatedExercises = append(updatedExercises, models.UserExerciseInput{
			ExerciseID: exercise.ExerciseID,
			Reps:       exercise.Reps,
			Load:       exercise.Load,
		})
	}
	// return success response
	utils.WriteJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"message":           "user exercise details submitted successfully",
		"user_workout_id":   userWorkoutID,
		"updated_exercises": updatedExercises,
	})
}
