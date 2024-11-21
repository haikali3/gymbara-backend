package models

type UserExerciseInput struct {
	ExerciseID int `json:"exercise_id"`
	Reps       int `json:"custom_reps"`
	Load       int `json:"custom_load"`
}

type UserExerciseRequest struct {
	SectionID int                 `json:"section_id"`
	Exercises []UserExerciseInput `json:"exercises"`
	UserEmail string              `json:"user_email"`
}
