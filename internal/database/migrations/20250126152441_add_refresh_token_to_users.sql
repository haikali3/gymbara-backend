-- +goose Up
-- +goose StatementBegin
ALTER TABLE Users
ADD COLUMN refresh_token TEXT DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE Users
DROP COLUMN refresh_token;
-- +goose StatementEnd
