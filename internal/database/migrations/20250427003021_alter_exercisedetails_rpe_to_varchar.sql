-- +goose Up
-- +goose StatementBegin
ALTER TABLE exercisedetails ALTER COLUMN rpe TYPE VARCHAR;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE exercisedetails ALTER COLUMN rpe TYPE INTEGER; -- Assuming the original type was INTEGER
-- +goose StatementEnd
