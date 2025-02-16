-- +goose Up
-- +goose StatementBegin
ALTER TABLE ExerciseDetails ALTER COLUMN load TYPE DOUBLE PRECISION USING load::double precision;
ALTER TABLE UserExercisesDetails ALTER COLUMN custom_load TYPE DOUBLE PRECISION USING custom_load::double precision;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ExerciseDetails ALTER COLUMN load TYPE INTEGER;
ALTER TABLE UserExercisesDetails ALTER COLUMN custom_load TYPE INTEGER;
-- +goose StatementEnd
