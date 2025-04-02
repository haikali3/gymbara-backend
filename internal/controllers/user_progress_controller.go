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
