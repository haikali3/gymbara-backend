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
}
