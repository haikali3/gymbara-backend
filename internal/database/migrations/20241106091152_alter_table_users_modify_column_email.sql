-- +goose Up
-- +goose StatementBegin
ALTER TABLE Users
ADD CONSTRAINT unique_email UNIQUE (email);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE Users
DROP CONSTRAINT IF EXISTS unique_email;
-- +goose StatementEnd

