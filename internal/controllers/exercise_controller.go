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
	"github.com/haikali3/gymbara-backend/pkg/cache"
	"github.com/haikali3/gymbara-backend/pkg/models"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"go.uber.org/zap"
)

// init cache instance
var workoutCache = cache.NewCache()

// Get workout sections
func GetWorkoutSections(w http.ResponseWriter, r *http.Request) {
	cacheKey := "workout_sections"

	// check if response is in the cache
	if cachedData, found := workoutCache.Get(cacheKey); found {
		utils.Logger.Info("Returning cached workout sections")
		utils.WriteJSONResponse(w, http.StatusOK, cachedData)
		return
	}

	rows, err := database.StmtGetWorkoutSections.Query()
	if err != nil {
		utils.HandleError(w, "Unable to query workout sections", http.StatusInternalServerError, err)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			utils.Logger.Error("Failed to close rows", zap.Error(err))
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
	if len(workoutSections) == 0 {
		utils.HandleError(w, "No workout sections found", http.StatusNotFound, nil)
		return
	}

	// store cache for 3 hours
	workoutCache.Set(cacheKey, workoutSections, 3*time.Hour)

	utils.WriteJSONResponse(w, http.StatusOK, workoutSections)
}

// Get exercises for initial load
func GetExercisesList(w http.ResponseWriter, r *http.Request) {
	workoutSectionID := r.URL.Query().Get("workout_section_id")
	if workoutSectionID == "" {
		utils.HandleError(w, "Missing workout_section_id parameter", http.StatusBadRequest, nil)
		return
	}

	// use the pre-prepared statement directly
	rows, err := database.StmtGetExercisesBySectionID.Query(workoutSectionID)
	if err != nil {
		utils.HandleError(w, "Unable to query exercises for workout_section_id: "+workoutSectionID, http.StatusInternalServerError, err)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			utils.Logger.Error("Failed to close rows", zap.Error(err))
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
	utils.Logger.Info("Retrieved exercises",
		zap.Int("count", len(exerciseList)),
		zap.String("workout_section_id", workoutSectionID),
	)
	utils.WriteJSONResponse(w, http.StatusOK, exerciseList)
}

// Get detailed exercise information
func GetExerciseDetails(w http.ResponseWriter, r *http.Request) {
	//query param
	workoutSectionID := r.URL.Query().Get("workout_section_id")
	if workoutSectionID == "" {
		utils.HandleError(w, "Missing workout_section_id parameter", http.StatusBadRequest, nil)
		return
	}

	rows, err := database.StmtGetExerciseDetails.Query(workoutSectionID)
	if err != nil {
		utils.HandleError(w, "Unable to query exercise details", http.StatusInternalServerError, err)
		utils.Logger.Error("Failed to query exercise details",
			zap.String("workout_section_id", workoutSectionID),
			zap.Error(err),
		)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			utils.Logger.Error("Failed to close rows", zap.Error(err))
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
	defer func() {
		if err := stmt.Close(); err != nil {
			utils.Logger.Error("Failed to close statement", zap.Error(err))
		}
	}()

	rows, err := stmt.Query(args...)
	if err != nil {
		utils.HandleError(w, fmt.Sprintf("Unable to query workout sections and exercises for workout_section_ids: %v. Query: %s", workoutSectionIDs, query), http.StatusInternalServerError, err)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			utils.Logger.Error("Failed to close rows", zap.Error(err))
		}
	}()

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
				utils.Logger.Info("Transaction committed successfully")
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
				fmt.Sprintf("Exercise ID %d: Reps=%d, Load=%d",
					exercise.ExerciseID, exercise.Reps, exercise.Load))
			continue
		}

		utils.Logger.Info("Adding exercise to batch",
			zap.Int("user_workout_id", userWorkoutID),
			zap.Int("exercise_id", exercise.ExerciseID),
			zap.Int("reps", exercise.Reps),
			zap.Int("load", exercise.Load),
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
			utils.HandleError(w, fmt.Sprintf("Failed to insert user exercise details for exercise_id: %d", request.Exercises[len(insertedExercises)].ExerciseID), http.StatusInternalServerError, txErr)
			return
		}

		utils.Logger.Debug("Batch insert query generated", zap.String("query", query))
	}

	// return success response
	utils.WriteJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"message":            "user exercises details submitted successfully",
		"user_workout_id":    userWorkoutID,
		"inserted_exercises": insertedExercises,
	})
	utils.Logger.Info("User exercise details submitted successfully", zap.Int("user_workout_id", userWorkoutID))
}

func GetUserProgress(w http.ResponseWriter, r *http.Request) {
	userIDValue := r.Context().Value(middleware.UserIDKey)
	if userIDValue == nil {
		utils.HandleError(w, "Unauthorized: User ID missing", http.StatusUnauthorized, nil)
		return
	}

	userID, ok := userIDValue.(int)
	if !ok {
		utils.HandleError(w, "Invalid User ID", http.StatusUnauthorized, nil)
		return
	}

	rows, err := database.StmtGetUserProgress.Query(userID)
	if err != nil {
		utils.HandleError(w, "Unable to retrieve user progress", http.StatusInternalServerError, err)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			utils.Logger.Error("Failed to close rows", zap.Error(err))
		}
	}()

	var progressData []models.UserProgressResponse
	for rows.Next() {
		var exerciseID int
		var exerciseName string
		var customLoad, customReps int
		var submittedAt time.Time

		if err := rows.Scan(&exerciseID, &exerciseName, &customLoad, &customReps, &submittedAt); err != nil {
			utils.HandleError(w, "Error scanning user progress data", http.StatusInternalServerError, err)
			return
		}

		// later find better way just for format time
		progressData = append(progressData, models.UserProgressResponse{
			ExerciseID:   exerciseID,
			ExerciseName: exerciseName,
			CustomLoad:   customLoad,
			CustomReps:   customReps,
			SubmittedAt:  submittedAt.Format("2006-01-02"),
		})
	}

	utils.Logger.Info("User progress retrieved successfully", zap.Int("user_id", userID), zap.Int("records", len(progressData)))
	utils.WriteJSONResponse(w, http.StatusOK, progressData)
}
