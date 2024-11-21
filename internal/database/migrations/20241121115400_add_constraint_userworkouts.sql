-- +goose Up
-- +goose StatementBegin
ALTER TABLE UserWorkouts
ADD CONSTRAINT user_section_unique UNIQUE (user_id, section_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE UserWorkouts
DROP CONSTRAINT user_section_unique;
-- +goose StatementEnd
