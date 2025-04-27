-- +goose Up
-- +goose StatementBegin

INSERT INTO WorkoutSections (name, route) VALUES
  ('Full Body', 'full_body'),
  ('Upper Body', 'upper_body'),
  ('Lower Body', 'lower_body');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM WorkoutSections
  WHERE route IN ('full_body', 'upper_body', 'lower_body');

-- rewind the auto-inc back to 1
ALTER SEQUENCE workoutsections_id_seq RESTART WITH 1;

-- +goose StatementEnd
