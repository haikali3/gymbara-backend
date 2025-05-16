package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/internal/middleware"
	"github.com/haikali3/gymbara-backend/pkg/models"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"go.uber.org/zap"
)

const (
	defaultLimit = 10
	maxLimit     = 100
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

	// Get and validate limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := defaultLimit

	if limitStr != "" {
		// First check if the string can be converted to an integer
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			utils.HandleError(w, "Invalid limit parameter: must be a number", http.StatusBadRequest, err)
			return
		}

		// Then check if the value is within allowed range
		if parsedLimit <= 0 {
			utils.HandleError(w, "Invalid limit parameter: must be greater than 0", http.StatusBadRequest, nil)
			return
		}
		if parsedLimit > maxLimit {
			utils.HandleError(w, "Invalid limit parameter: exceeds maximum allowed value", http.StatusBadRequest, nil)
			return
		}

		limit = parsedLimit
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
	count := 0
	for rows.Next() {
		if count >= limit {
			break
		}

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
		count++
	}

	utils.Logger.Info("User progress retrieved successfully",
		zap.Int("user_id", userID),
		zap.Int("records", len(progressData)),
		zap.Int("limit", limit))
	utils.WriteStandardResponse(w, http.StatusOK, "User progress retrieved successfully", progressData)
}
