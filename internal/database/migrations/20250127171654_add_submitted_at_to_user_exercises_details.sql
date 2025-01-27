-- +goose Up
-- +goose StatementBegin
ALTER TABLE UserExercisesDetails
ADD COLUMN submitted_at DATE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE UserExercisesDetails
DROP COLUMN submitted_at;
-- +goose StatementEnd
