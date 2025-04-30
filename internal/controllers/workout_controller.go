// internal/controllers/workout_controller.go

package controllers

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/pkg/cache"
	"github.com/haikali3/gymbara-backend/pkg/models"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"go.uber.org/zap"
)

// init cache instance
var workoutCache = cache.WorkoutCache

func GetWorkoutSections(w http.ResponseWriter, r *http.Request) {
	cacheKey := "workout_sections"

	// check if response is in the cache
	if cachedData, found := workoutCache.Get(cacheKey); found {
		utils.Logger.Info("Returning cached workout sections")
		utils.WriteStandardResponse(w, http.StatusOK, "Workout sections retrieved successfully (from cache)", cachedData)
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

	utils.WriteStandardResponse(w, http.StatusOK, "Workout sections retrieved successfully", workoutSections)
}

func GetWorkoutSectionsWithExercises(w http.ResponseWriter, r *http.Request) {
	workoutSectionIDs := r.URL.Query()["workout_section_ids"]
	utils.Logger.Debug("Workout section IDs received", zap.Strings("workout_section_ids", workoutSectionIDs))
	if len(workoutSectionIDs) == 0 {
		utils.HandleError(w, "Missing workout_section_ids parameter", http.StatusBadRequest, nil)
		return
	}

	cacheKey := "workout_sections_with_exercises_" + strings.Join(workoutSectionIDs, "_")

	if cachedData, found := workoutCache.Get(cacheKey); found {
		utils.Logger.Info("Returning cached workout sections with exercises", zap.String("cacheKey", cacheKey))
		utils.WriteStandardResponse(w, http.StatusOK, "Workout sections with exercises retrieved successfully (from cache)", cachedData)
		return
	}

	placeholders, args := utils.GeneratePlaceholders(len(workoutSectionIDs))
	for i, id := range workoutSectionIDs {
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT
			ws.id              AS section_id,
			ws.name            AS section_name,
			ws.route           AS section_route,
			COALESCE(e.id, 0)    AS exercise_id,
			COALESCE(e.name, '') AS exercise_name
		FROM WorkoutSections ws
		LEFT JOIN Exercises e
			ON ws.id = e.workout_section_id
		WHERE ws.id IN (%s)
		ORDER BY ws.id, e.id;
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

	// ✅ Store in cache for 24 hours
	utils.Logger.Info("Storing workout sections with exercises in cache", zap.String("cacheKey", cacheKey))
	workoutCache.Set(cacheKey, sections, 3*time.Hour)

	utils.WriteStandardResponse(w, http.StatusOK, "Workout sections with exercises retrieved successfully", sections)
}
