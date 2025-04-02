package controllers

import (
	"net/http"
	"time"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/internal/middleware"
	"github.com/haikali3/gymbara-backend/pkg/models"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"go.uber.org/zap"
)

// Get exercises for initial load
func GetExercisesList(w http.ResponseWriter, r *http.Request) {
	workoutSectionID := r.URL.Query().Get("workout_section_id")
	if workoutSectionID == "" {
		utils.HandleError(w, "Missing workout_section_id parameter", http.StatusBadRequest, nil)
		return
	}

	cacheKey := "exercise_list_" + workoutSectionID

	// check if response is in the cache, if no, query db
	if cachedData, found := workoutCache.Get(cacheKey); found {
		utils.Logger.Info("Returning cached exercise list", zap.String("sectionID", workoutSectionID))
		utils.WriteStandardResponse(w, http.StatusOK, "Exercise list retrieved successfully (from cache)", cachedData)
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

	// store cache for 3 hours
	workoutCache.Set(cacheKey, exerciseList, 3*time.Hour)

	utils.WriteStandardResponse(w, http.StatusOK, "Exercise list retrieved successfully", exerciseList)
}

// Get detailed exercise information
func GetExerciseDetails(w http.ResponseWriter, r *http.Request) {
	//query param
	workoutSectionID := r.URL.Query().Get("workout_section_id")
	if workoutSectionID == "" {
		utils.HandleError(w, "Missing workout_section_id parameter", http.StatusBadRequest, nil)
		return
	}

	cacheKey := "exercise_details_" + workoutSectionID

	// Check cache first
	if cachedData, found := workoutCache.Get(cacheKey); found {
		utils.Logger.Info("Returning cached exercise details", zap.String("workout_section_id", workoutSectionID))
		utils.WriteStandardResponse(w, http.StatusOK, "Exercise details retrieved successfully (from cache)", cachedData)
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
		if err := rows.Scan(
			&detail.ID,
			&detail.Name,
			&detail.WarmupSets,
			&detail.WorkSets,
			&detail.Reps,
			&detail.Load,
			&detail.RPE,
			&detail.RestTime,
		); err != nil {
			utils.HandleError(w, "Unable to scan exercise details", http.StatusInternalServerError, err)
			return
		}
		exerciseDetails = append(exerciseDetails, detail)
	}
	utils.Logger.Info("Retrieved exercise details",
		zap.Int("count", len(exerciseDetails)),
		zap.String("workout_section_id", workoutSectionID),
	)

	workoutCache.Set(cacheKey, exerciseDetails, 3*time.Hour)

	utils.WriteStandardResponse(w, http.StatusOK, "Exercise details retrieved successfully", exerciseDetails)
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
		var exerciseID, customReps int
		var exerciseName string
		var customLoad float64
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
	utils.WriteStandardResponse(w, http.StatusOK, "User progress retrieved successfully", progressData)
}
