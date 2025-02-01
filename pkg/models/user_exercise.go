package models

import (
	"encoding/json"
	"time"
)

type UserExerciseInput struct {
	ExerciseID  int       `json:"exercise_id"`
	Reps        int       `json:"custom_reps"`
	Load        int       `json:"custom_load"`
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

// Custom JSON unmarshaller to parse `submitted_at` as `YYYY-MM-DD`
func (u *UserExerciseInput) UnmarshalJSON(data []byte) error {
	type Alias UserExerciseInput
	aux := &struct {
		SubmittedAt string `json:"submitted_at"`
		*Alias
	}{
		Alias: (*Alias)(u),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// // handle empty `submitted_at` field
	// if aux.SubmittedAt == "" {
	// 	u.SubmittedAt = time.Time{} // set to zero value of time.Time if empty
	// 	return nil
	// }

	// Parse non-empty `submitted_at`
	parsedTime, err := time.Parse("2006-01-02", aux.SubmittedAt)
	if err != nil {
		return err
	}
	u.SubmittedAt = parsedTime

	return nil
}
