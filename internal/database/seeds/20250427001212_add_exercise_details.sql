-- file: internal/database/seeds/20250427001212_add_exercise_details.sql
-- +goose Up
-- +goose StatementBegin
INSERT INTO ExerciseDetails (exercise_id, week_start, week_end, warmup_sets, working_sets, reps, load, rpe, rest_time) VALUES 
(1, 5, 8, 1, 2, '8-10', NULL, '9-10', '~2 MINS'),
(2, 5, 8, 2, 1, '6-8 per leg', NULL, '8-9', '~3 MINS'),
(3, 5, 8, 0, 1, '10-12 per leg', NULL, '8-9', '~3 MINS'),
(4, 5, 8, 2, 2, '8-10', NULL, '9-10', '~2 MINS'),
(5, 5, 8, 1, 1, '10-12', NULL, '10', '~1.5 MINS'),
(6, 5, 8, 1, 1, '12-15 (dropset)', NULL, '10', '~1.5 MINS'),
(7, 5, 8, 1, 1, '12-15 (dropset)', NULL, '10', '~1.5 MINS'),
(8, 5, 8, 0, 1, '12-15', NULL, '10', '~1.5 MINS');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM ExerciseDetails
WHERE (exercise_id, week_start, week_end) IN (
  (1, 5, 8),
  (2, 5, 8),
  (3, 5, 8),
  (4, 5, 8),
  (5, 5, 8),
  (6, 5, 8),
  (7, 5, 8),
  (8, 5, 8)
);

-- +goose StatementEnd
