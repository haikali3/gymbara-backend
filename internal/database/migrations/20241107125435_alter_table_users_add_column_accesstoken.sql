-- +goose Up
-- +goose StatementBegin
ALTER TABLE Users ADD COLUMN access_token TEXT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE Users DROP COLUMN access_token;
-- +goose StatementEnd
