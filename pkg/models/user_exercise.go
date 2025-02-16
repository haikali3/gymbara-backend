package models

import (
	"encoding/json"
	"time"
)

type UserExerciseInput struct {
	ExerciseID  int       `json:"exercise_id"`
	Reps        int       `json:"custom_reps"`
	Load        float64   `json:"custom_load"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type UserExerciseRequest struct {
	SectionID int                 `json:"section_id"`
	Exercises []UserExerciseInput `json:"exercises"`
	UserEmail string              `json:"user_email"`
}

// Custom JSON marshaler for UserExerciseInput
func (u UserExerciseInput) MarshalJSON() ([]byte, error) {
	type Alias UserExerciseInput
	return json.Marshal(&struct {
		SubmittedAt string `json:"submitted_at"`
		*Alias
	}{
		SubmittedAt: u.SubmittedAt.Format("2006-01-02"), // Format as yyyy-mm-dd
		Alias:       (*Alias)(&u),
	})
}

// Response model for progress data
type UserProgressResponse struct {
	ExerciseID   int     `json:"exercise_id"`
	ExerciseName string  `json:"exercise_name"` //refactor to use UserExerciseInput model
	CustomReps   int     `json:"custom_reps"`
	CustomLoad   float64 `json:"custom_load"`
	SubmittedAt  string  `json:"submitted_at"`
}
