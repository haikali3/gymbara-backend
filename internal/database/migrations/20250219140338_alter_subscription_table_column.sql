-- +goose Up
-- +goose StatementBegin
ALTER TABLE Subscriptions ADD COLUMN stripe_customer_id VARCHAR(100);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE Subscriptions DROP COLUMN stripe_customer_id;
-- +goose StatementEnd
