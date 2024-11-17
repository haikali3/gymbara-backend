-- +goose Up
-- +goose StatementBegin
-- Add the `route` column to the `workoutsections` table and update its values
ALTER TABLE workoutsections ADD COLUMN route VARCHAR(50);
UPDATE workoutsections SET route = 'full-body' WHERE id = 1;
UPDATE workoutsections SET route = 'upper-body' WHERE id = 2;
UPDATE workoutsections SET route = 'lower-body' WHERE id = 3;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove the `route` column from the `workoutsections` table
ALTER TABLE workoutsections DROP COLUMN route;
-- +goose StatementEnd
