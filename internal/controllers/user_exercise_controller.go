package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/internal/middleware"
	"github.com/haikali3/gymbara-backend/pkg/models"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"go.uber.org/zap"
)

// Case 1: First-time submission (New row inserted)
// Case 2: Duplicate submission on the same day (Row updated)
func SubmitUserExerciseDetails(w http.ResponseWriter, r *http.Request) {
	var request models.UserExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.HandleError(w, "Invalid request body", http.StatusBadRequest, err)
		return
	}

	// validate the request
	if request.SectionID == 0 || len(request.Exercises) == 0 {
		utils.Logger.Warn("Missing required fields in user exercise submission",
			zap.Int("section_id", request.SectionID),
			zap.Int("exercise_count", len(request.Exercises)),
		)
		utils.HandleError(w, "Missing required fields: section_id or exercises", http.StatusBadRequest, nil)
		return
	}

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

	// begin db transaction
	tx, err := database.DB.Begin()
	if err != nil {
		utils.HandleError(w, "Failed to start database transaction", http.StatusInternalServerError, err)
		return
	}

	// rollback or commit transaction appropriately
	var txErr error
	defer func() {
		if p := recover(); p != nil {
			if err := tx.Rollback(); err != nil {
				utils.Logger.Error("Transaction rollback failed", zap.Error(err))
			}
			panic(p)
		} else if txErr != nil {
			if err := tx.Rollback(); err != nil {
				utils.Logger.Error("Transaction rollback failed", zap.Error(err))
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				utils.Logger.Error("Transaction commit failed", zap.Error(commitErr))
				utils.HandleError(w, "Transaction commit failed", http.StatusInternalServerError, commitErr)
				return
			} else {
				utils.Logger.Info("Transaction committed successfully",
					zap.Int("user_id", userID),
					zap.Int("section_id", request.SectionID),
				)
			}
		}
	}()

	// insert or update UserWorkouts
	var userWorkoutID int
	txErr = tx.QueryRow(`
		INSERT INTO UserWorkouts (user_id, section_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, section_id) DO UPDATE
		SET user_id = EXCLUDED.user_id
		RETURNING id
	`, userID, request.SectionID).Scan(&userWorkoutID)
	if txErr != nil {
		utils.HandleError(
			w,
			fmt.Sprintf("Failed to insert or update user workout for user_id: %d, section_id: %d. Error: %v",
				userID,
				request.SectionID,
				txErr),
			http.StatusInternalServerError, txErr)
		return
	}

	// batch(collect exercise id) and check if exercise exist
	exerciseIDs := make([]int, len(request.Exercises))
	for i, exercise := range request.Exercises {
		exerciseIDs[i] = exercise.ExerciseID
	}

	queryPlaceholders := make([]string, len(exerciseIDs))
	queryValues := make([]interface{}, len(exerciseIDs))
	for i, id := range exerciseIDs {
		queryPlaceholders[i] = fmt.Sprintf("$%d", i+1)
		queryValues[i] = id
	}

	query := fmt.Sprintf(`SELECT id FROM Exercises WHERE id IN (%s)`, strings.Join(queryPlaceholders, ","))
	rows, txErr := tx.Query(query, queryValues...)
	if txErr != nil {
		utils.HandleError(w, "Failed to validate exercise IDs", http.StatusInternalServerError, txErr)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			utils.Logger.Error("Failed to close rows", zap.Error(err))
		}
	}()

	validExerciseIDs := make(map[int]bool)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err == nil {
			validExerciseIDs[id] = true
		}
	}

	// prepare for batch insert
	var invalidExercises []string
	var placeholders []string
	var values []interface{}
	var insertedExercises []models.UserExerciseInput

	// use the same time for batch
	currentTime := time.Now()

	for i, exercise := range request.Exercises {
		//check if exercise exist in db
		if !validExerciseIDs[exercise.ExerciseID] {
			utils.Logger.Warn("Attempt to insert invalid exercise ID", zap.Int("exercise_id", exercise.ExerciseID))
			utils.HandleError(w, fmt.Sprintf("Invalid exercise_id: %d doesn't exist", exercise.ExerciseID), http.StatusBadRequest, nil)
			return
		}

		if exercise.Reps <= 0 || exercise.Load <= 0 {
			invalidExercises = append(invalidExercises,
				fmt.Sprintf("Exercise ID %d: Reps=%d, Load=%.2f",
					exercise.ExerciseID, exercise.Reps, exercise.Load))
			continue
		}

		utils.Logger.Info("Adding exercise to batch",
			zap.Int("user_workout_id", userWorkoutID),
			zap.Int("exercise_id", exercise.ExerciseID),
			zap.Int("reps", exercise.Reps),
			zap.Float64("load", exercise.Load),
		)
		// use placeholders batch insert with timestamp
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))
		values = append(values, userWorkoutID, exercise.ExerciseID, exercise.Reps, exercise.Load, currentTime)

		//return value for response
		insertedExercises = append(insertedExercises, models.UserExerciseInput{
			ExerciseID:  exercise.ExerciseID,
			Reps:        exercise.Reps,
			Load:        exercise.Load,
			SubmittedAt: currentTime,
		})
	}

	if len(invalidExercises) > 0 {
		utils.Logger.Warn("Invalid exercises detected", zap.Strings("invalid_exercises", invalidExercises))
		utils.HandleError(w,
			fmt.Sprintf("Invalid reps or load for exercises: %v",
				strings.Join(invalidExercises, "; ")),
			http.StatusBadRequest,
			nil)
		return
	}

	utils.Logger.Info("Executing batch insert for user exercises", zap.Int("exercise_count", len(request.Exercises)))

	// batch insert
	if len(placeholders) > 0 {
		query := fmt.Sprintf(`
		INSERT INTO UserExercisesDetails (user_workout_id, exercise_id, custom_reps, custom_load, submitted_at)
		VALUES %s
		ON CONFLICT ON CONSTRAINT unique_user_exercise_submission
		DO UPDATE
		SET custom_reps = EXCLUDED.custom_reps, 
				custom_load = EXCLUDED.custom_load,
				submitted_at = EXCLUDED.submitted_at
		`, strings.Join(placeholders, ", "))

		_, txErr = tx.Exec(query, values...)
		if txErr != nil {
			exerciseID := 0
			if len(insertedExercises) > 0 {
				exerciseID = insertedExercises[len(insertedExercises)-1].ExerciseID
			}
			utils.HandleError(
				w,
				fmt.Sprintf("Failed to insert user exercise details for exercise_id: %d", exerciseID),
				http.StatusInternalServerError,
				txErr,
			)
			return
		}

		utils.Logger.Debug("Batch insert query generated", zap.String("query", query))
	}

	// ✅ Invalidate cache when exercises are updated
	sectionIDStr := strconv.Itoa(request.SectionID)
	if request.SectionID > 0 {
		workoutCache.Delete("exercise_list_" + sectionIDStr)
		workoutCache.Delete("exercise_details_" + sectionIDStr)
	}

	utils.Logger.Info("Cache invalidated for updated workout sections and exercises")
	utils.Logger.Info("Cache invalidated", zap.String("cacheKey", "workout_sections"))
	utils.Logger.Info("Cache invalidated", zap.String("cacheKey", "exercise_list_"+sectionIDStr))
	utils.Logger.Info("Cache invalidated", zap.String("cacheKey", "exercise_details_"+sectionIDStr))

	// return success response
	utils.WriteStandardResponse(w, http.StatusCreated, "User exercise details submitted successfully", map[string]interface{}{
		"user_workout_id":    userWorkoutID,
		"inserted_exercises": insertedExercises,
	})

	utils.Logger.Info("User exercise details submitted successfully", zap.Int("user_workout_id", userWorkoutID))
}
