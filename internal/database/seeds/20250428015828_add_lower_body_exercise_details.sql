-- file: internal/database/seeds/20250428015828_add_lower_body_exercise_details.sql
-- +goose Up
-- +goose StatementBegin
INSERT INTO ExerciseDetails (
  exercise_id,
  week_start,
  week_end,
  warmup_sets,
  working_sets,
  reps,
  load,
  rpe,
  rest_time
) VALUES
  -- DB Bulgarian Split Squat
  (16, 5, 8, 2, 3, '10-12', NULL, '8-9',  '~2 MINS'),
  -- DB Romanian Deadlift
  (17, 5, 8, 2, 2, '10-12', NULL, '8-9',  '~2 MINS'),
  -- Goblet Squat
  (18, 5, 8, 1, 1, '12-15', NULL, '9-10', '~1.5 MINS'),
  -- A1: Leg Press Toe Press
  (19, 5, 8, 1, 2, '15-20', NULL, '10',   '0 MINS'),
  -- A2: Machine Crunch
  (20, 5, 8, 1, 2, '10-12', NULL, '1-9',  '~1.5 MINS');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM ExerciseDetails
WHERE (exercise_id, week_start, week_end) IN (
  (16, 5, 8),
  (17, 5, 8),
  (18, 5, 8),
  (19, 5, 8),
  (20, 5, 8)
);
-- +goose StatementEnd
