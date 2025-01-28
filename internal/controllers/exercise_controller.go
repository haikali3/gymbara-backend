package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"strings"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/internal/middleware"
	"github.com/haikali3/gymbara-backend/pkg/models"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"go.uber.org/zap"
)

// Get workout sections
func GetWorkoutSections(w http.ResponseWriter, r *http.Request) {
	query := "SELECT id, name, route FROM WorkoutSections"

	stmt, err := database.DB.Prepare(query)
	if err != nil {
		utils.HandleError(w, "Unable to prepare statement", http.StatusInternalServerError, err)
		return
	}
	defer stmt.Close()

	// use prepared statement
	rows, err := stmt.Query()
	if err != nil {
		utils.HandleError(w, "Unable to query workout sections", http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	var workoutSections []models.WorkoutSection
	for rows.Next() {
		var workoutSection models.WorkoutSection
		if err := rows.Scan(&workoutSection.ID, &workoutSection.Name, &workoutSection.Route); err != nil {
			utils.HandleError(w, "Unable to scan workout sections", http.StatusInternalServerError, err)
			return
		}
		workoutSections = append(workoutSections, workoutSection)
	}
	if len(workoutSections) == 0 {
		utils.HandleError(w, "No workout sections found", http.StatusNotFound, nil)
		return
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

	//prepare statement to avoid sql injection
	stmt, err := database.DB.Prepare(query)
	if err != nil {
		utils.HandleError(w, "Unable to prepare statement for querying exercises", http.StatusInternalServerError, err)
		return
	}
	defer stmt.Close()

	// use prepared statement
	rows, err := stmt.Query(workoutSectionID)
	if err != nil {
		utils.HandleError(w, "Unable to query exercise basic details", http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	var exerciseList []models.ExerciseDetails
	for rows.Next() {
		var detail models.ExerciseDetails
		if err := rows.Scan(&detail.Name, &detail.Reps, &detail.WorkSets, &detail.Load); err != nil {
			utils.HandleError(w, "Unable to scan exercise details", http.StatusInternalServerError, err)
			return
		}
		exerciseList = append(exerciseList, detail)
	}
	utils.Logger.Info("Retrieved exercises",
		zap.Int("count", len(exerciseList)),
		zap.String("workout_section_id", workoutSectionID),
	)
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
		utils.Logger.Error("Failed to query exercise details",
			zap.String("workout_section_id", workoutSectionID),
			zap.Error(err),
		)
		return
	}
	defer rows.Close()

	var exerciseDetails []models.ExerciseDetails
	for rows.Next() {
		var detail models.ExerciseDetails
		if err := rows.Scan(&detail.Name, &detail.WarmupSets, &detail.WorkSets, &detail.Reps, &detail.Load, &detail.RPE, &detail.RestTime); err != nil {
			utils.HandleError(w, "Unable to scan exercise details", http.StatusInternalServerError, err)
			return
		}
		exerciseDetails = append(exerciseDetails, detail)
	}
	utils.Logger.Info("Retrieved exercise details",
		zap.Int("count", len(exerciseDetails)),
		zap.String("workout_section_id", workoutSectionID),
	)
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

	stmt, err := database.DB.Prepare(query)
	if err != nil {
		utils.HandleError(w, "Unable to prepare statement for querying workout sections and exercises", http.StatusInternalServerError, err)
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		utils.HandleError(w, fmt.Sprintf("Unable to query workout sections and exercises for workout_section_ids: %v. Query: %s", workoutSectionIDs, query), http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	//map is unordered lol
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
		if exercise.ID != 0 {
			sectionsMap[sectionID].Exercises = append(sectionsMap[sectionID].Exercises, exercise)
		}
	}

	sections := make([]models.WorkoutSectionWithExercises, 0, len(sectionsMap))
	for _, section := range sectionsMap {
		sections = append(sections, *section)
	}

	sort.Slice(sections, func(i, j int) bool {
		return sections[i].ID < sections[j].ID
	})

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

	//TODO: Validate all inputs (e.g., section_id, exercise_id, reps, load) upfront, before starting the transaction.
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
			utils.Logger.Error("Transaction rolled back due to panic", zap.Any("panic", p))
			panic(p)
		} else if err != nil {
			tx.Rollback()
			utils.Logger.Error("Transaction rolled back due to error", zap.Error(err))
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				utils.Logger.Error("Transaction commit failed", zap.Error(commitErr))
			} else {
				utils.Logger.Info("Transaction committed successfully")
			}
		}
	}()

	//validate user existence using OAuth email or ID
	userIDValue := r.Context().Value(middleware.UserIDKey)
	if userIDValue == nil {
		utils.HandleError(w, "User ID missing or invalid in request context", http.StatusUnauthorized, nil)
		return
	}

	userID, ok := userIDValue.(int)
	if !ok {
		utils.HandleError(w, "Invalid user ID type in request context", http.StatusUnauthorized, nil)
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
	var invalidExercises []string
	for _, exercise := range request.Exercises {
		if exercise.Reps <= 0 || exercise.Load <= 0 {
			invalidExercises = append(invalidExercises,
				fmt.Sprintf("Exercise ID %d: Reps=%d, Load=%d",
					exercise.ExerciseID, exercise.Reps, exercise.Load))
		}
	}

	if len(invalidExercises) > 0 {
		utils.HandleError(w,
			fmt.Sprintf("Invalid reps or load for exercises: %v",
				strings.Join(invalidExercises, "; ")),
			http.StatusBadRequest,
			nil)
		return
	}

	for _, exercise := range request.Exercises {
		//check exercise id exist
		var exerciseExists bool
		err = tx.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM Exercises WHERE id = $1)
		`, exercise.ExerciseID).Scan(&exerciseExists)
		if err != nil || !exerciseExists {
			utils.HandleError(w, fmt.Sprintf("Invalid exercise_id: %d doesn't exist", exercise.ExerciseID), http.StatusBadRequest, nil)
			return
		}

		//TODO: use batch SQL Query
		//! how to handle if user add new exercise? it should INSERT only for Date(to track progress)
		//execute upsert for each exercise
		_, err = tx.Exec(`
		INSERT INTO UserExercisesDetails (user_workout_id, exercise_id, custom_reps, custom_load, submitted_at)
		VALUES ($1, $2, $3, $4, CURRENT_DATE)
		ON CONFLICT (user_workout_id, exercise_id) DO UPDATE
		SET custom_reps = $3, custom_load = $4, submitted_at = CURRENT_DATE
	`, userWorkoutID, exercise.ExerciseID, exercise.Reps, exercise.Load)
		if err != nil {
			utils.HandleError(w, "failed to insert or update user exercise details", http.StatusInternalServerError, err)
			return
		}
	}

	var updatedExercises []models.UserExerciseInput
	for _, exercise := range request.Exercises {

		updatedExercises = append(updatedExercises, models.UserExerciseInput{
			ExerciseID:  exercise.ExerciseID,
			Reps:        exercise.Reps,
			Load:        exercise.Load,
			SubmittedAt: time.Now(),
		})
	}
	// return success response
	utils.WriteJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"message":           "user exercise details submitted successfully",
		"user_workout_id":   userWorkoutID,
		"updated_exercises": updatedExercises,
	})
}
