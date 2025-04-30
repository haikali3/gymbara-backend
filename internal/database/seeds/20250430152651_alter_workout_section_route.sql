-- file: internal/database/seeds/20250426132152_init_seed.sql

-- +goose Up
-- +goose StatementBegin

UPDATE WorkoutSections SET route = REPLACE(route, '_', '-');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

UPDATE WorkoutSections SET route = REPLACE(route, '-', '_');

-- +goose StatementEnd


