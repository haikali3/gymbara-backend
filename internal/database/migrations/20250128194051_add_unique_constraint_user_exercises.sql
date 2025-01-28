-- +goose Up
-- +goose StatementBegin
-- Add a unique constraint to prevent duplicate entries for the same workout, exercise, and date.
ALTER TABLE UserExercisesDetails
ADD CONSTRAINT unique_user_exercise_submission UNIQUE (user_workout_id, exercise_id, submitted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove the unique constraint on user_workout_id, exercise_id, and submitted_at.
ALTER TABLE UserExercisesDetails
DROP CONSTRAINT IF EXISTS unique_user_exercise_submission;
-- +goose StatementEnd
