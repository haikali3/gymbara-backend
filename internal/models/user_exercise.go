package models

import "time"

type UserExerciseInput struct {
	ExerciseID  int       `json:"exercise_id"`
	Reps        int       `json:"custom_reps"`
	Load        int       `json:"custom_load"`
	SubmittedAt time.Time `json:"submitted_at" time_format:"2006-01-02"`
}

type UserExerciseRequest struct {
	SectionID int                 `json:"section_id"`
	Exercises []UserExerciseInput `json:"exercises"`
	UserEmail string              `json:"user_email"`
}
