-- +goose Up
-- +goose StatementBegin
ALTER TABLE Exercises
  RENAME COLUMN workoutsection_id TO workout_section_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE Exercises
  RENAME COLUMN workout_section_id TO workoutsection_id;
-- +goose StatementEnd
