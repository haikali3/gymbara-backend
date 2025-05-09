// internal/controllers/exercise_controller.go
package controllers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/pkg/models"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"go.uber.org/zap"
)

// ExerciseGuideDTO is the shape we return to clients
type ExerciseGuideDTO struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	Notes         string   `json:"notes"`
	Substitutions []string `json:"substitutions"`
}

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

// GET /workout-sections/exercises/42/guide
func GetExerciseGuide(w http.ResponseWriter, r *http.Request) {
	// e.g. "/workout-sections/exercises/42/guide"
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	// Expect exactly 4 segments:
	// ["workout-sections", "exercises", "<id>", "guide"]
	if len(parts) != 4 ||
		parts[0] != "workout-sections" ||
		parts[1] != "exercises" ||
		parts[3] != "guide" {
		http.NotFound(w, r)
		return
	}
	exID := parts[2]

	// Example: fetch two subs as part of the guide
	row := database.DB.QueryRow(`
        SELECT
					id, 
					name, 
					substitution_1, 
					substitution_2, 
					notes
        FROM Exercises
        WHERE id = $1;
    `, exID)

	var (
		id    int
		name  string
		sub1  sql.NullString
		sub2  sql.NullString
		notes sql.NullString
	)
	if err := row.Scan(&id, &name, &sub1, &sub2, &notes); err != nil {
		if err == sql.ErrNoRows {
			utils.HandleError(w, "Exercise not found", http.StatusNotFound, err)
		} else {
			utils.HandleError(w, "Error querying exercise guide", http.StatusInternalServerError, err)
		}
		return
	}

	// Build whatever “guide” payload you need—here we just wrap the two substitutions
	subs := []string{}

	if sub1.Valid && sub1.String != "" {
		subs = append(subs, sub1.String)
	}
	if sub2.Valid && sub2.String != "" {
		subs = append(subs, sub2.String)
	}

	// Assemble the DTO
	dto := ExerciseGuideDTO{
		ID:            id,
		Name:          name,
		Notes:         notes.String, // empty string if null
		Substitutions: subs,
	}

	utils.Logger.Info("Fetched exercise guide",
		zap.Int("exercise_id", id),
		zap.Any("guide", dto),
	)
	utils.WriteStandardResponse(w, http.StatusOK, "Exercise guide retrieved", dto)

}
