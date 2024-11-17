package models

type UserExercise struct {
	ExerciseID int `json:"exercise_id"`
	Reps       int `json:"custom_reps"`
	Load       int `json:"custom_load"`
}

type UserExerciseRequest struct {
	UserWorkoutID int            `json:"user_workout_id"`
	Exercises     []UserExercise `json:"exercises"`
}
