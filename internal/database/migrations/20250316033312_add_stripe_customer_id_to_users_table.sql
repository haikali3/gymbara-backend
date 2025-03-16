-- +goose Up
-- +goose StatementBegin
ALTER TABLE Users ADD COLUMN stripe_customer_id VARCHAR(100);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE Users DROP COLUMN stripe_customer_id;
-- +goose StatementEnd
