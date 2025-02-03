package database

import (
	"database/sql"

	"github.com/haikali3/gymbara-backend/pkg/utils"
	"go.uber.org/zap"
)

var (
	StmtGetWorkoutSections      *sql.Stmt
	StmtGetExercisesBySectionID *sql.Stmt
	StmtGetExerciseDetails      *sql.Stmt
	StmtGetUserProgress         *sql.Stmt
)

func PrepareStatements() {
	var err error
	StmtGetWorkoutSections, err = DB.Prepare(
		"SELECT id, name, route FROM WorkoutSections",
	)
	if err != nil {
		utils.Logger.Fatal("Failed to prepare StmtGetWorkoutSections", zap.Error(err))
	}

	StmtGetExercisesBySectionID, err = DB.Prepare(`
    SELECT e.name, ed.reps, ed.working_sets, ed.load
    FROM Exercises e
    JOIN ExerciseDetails ed ON e.id = ed.exercise_id
    WHERE e.workout_section_id = $1
  `)
	if err != nil {
		utils.Logger.Fatal("Failed to prepare StmtGetExercisesBySectionID", zap.Error(err))
	}

	StmtGetExerciseDetails, err = DB.Prepare(`
  SELECT e.name, ed.warmup_sets, ed.working_sets, ed.reps, ed.load, ed.rpe, ed.rest_time
  FROM Exercises e
  JOIN ExerciseDetails ed ON e.id = ed.exercise_id
  WHERE e.workout_section_id = $1
  `)
	if err != nil {
		utils.Logger.Fatal("Failed to prepare StmtGetExerciseDetails", zap.Error(err))
	}

	StmtGetUserProgress, err = DB.Prepare(`
	SELECT ued.exercise_id, e.name, ued.custom_load, ued.custom_reps, ued.submitted_at
	FROM UserExercisesDetails ued
	JOIN Exercises e ON ued.exercise_id = e.id
	JOIN UserWorkouts uw ON ued.user_workout_id = uw.id
	WHERE uw.user_id = $1
	ORDER BY ued.submitted_at ASC
`)
	if err != nil {
		utils.Logger.Fatal("Failed to prepare StmtGetUserProgress", zap.Error(err))
	}
}

func CloseStatement() {
	if err := StmtGetWorkoutSections.Close(); err != nil {
		utils.Logger.Error("Failed to close StmtGetWorkoutSections", zap.Error(err))
	}
	if err := StmtGetExercisesBySectionID.Close(); err != nil {
		utils.Logger.Error("Failed to close StmtGetExercisesBySectionID", zap.Error(err))
	}
	if err := StmtGetExerciseDetails.Close(); err != nil {
		utils.Logger.Error("Failed to close StmtGetExerciseDetails", zap.Error(err))
	}
	if err := StmtGetUserProgress.Close(); err != nil {
		utils.Logger.Error("Failed to close StmtGetUserProgress", zap.Error(err))
	}
}
