-- +goose Up
-- +goose StatementBegin
ALTER TABLE UserExercisesDetails
ADD CONSTRAINT unique_user_workout_exercise UNIQUE (user_workout_id, exercise_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE UserExercisesDetails
DROP CONSTRAINT unique_user_workout_exercise;
-- +goose StatementEnd
