-- +goose Up
-- +goose StatementBegin
ALTER TABLE Sections RENAME TO WorkoutSections;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE WorkoutSections RENAME TO Sections;
-- +goose StatementEnd
