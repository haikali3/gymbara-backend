-- +goose Up
-- +goose StatementBegin
CREATE TABLE UserExerciseDetails (
    id SERIAL PRIMARY KEY,
    user_workout_id INT REFERENCES UserWorkouts(id),
    exercise_id INT REFERENCES Exercises(id),
    custom_reps INT,
    custom_load INT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS UserExerciseDetails;
-- +goose StatementEnd
