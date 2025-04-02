-- +goose Up
-- +goose StatementBegin
ALTER TABLE subscriptions DROP COLUMN stripe_customer_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE subscriptions ADD COLUMN stripe_customer_id VARCHAR(255);
-- +goose StatementEnd
