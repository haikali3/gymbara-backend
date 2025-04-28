-- file: internal/database/seeds/20250426132152_init_seed.sql

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

-- +goose StatementEnd
