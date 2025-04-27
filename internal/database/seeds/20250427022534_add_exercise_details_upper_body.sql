-- file: internal/database/seeds/20250427022316_add_upper_body_exercise_details.sql
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
  -- 2-Grip Pullup
  (9, 5, 8, 1, 2, '8-10', '63kgx10', '9-10', '~3 MINS'),
  -- Weighted Dip (Heavy)
  (10, 5, 8, 2, 1, '6-8', '72kg x 6', '8-9', '~3 MINS'),
  -- Weighted Dip (Back off)
  (11, 5, 8, 0, 1, '10-12', '59kg x 10', '9-10', '~3 MINS'),
  -- Incline Chest-Supported DB Row
  (12, 5, 8, 1, 2, '8-10', '59kgx10', '9-10', '~2 MINS'),
  -- Standing DB Arnold Press
  (13, 5, 8, 1, 2, '8-10', '50kgx10', '9-10', '~2 MINS'),
  -- A1: DB Incline Curl
  (14, 5, 8, 1, 2, '15-20', '12.5kgx15', '10', '0 MINS'),
  -- A2: DB French Press
  (15, 5, 8, 1, 2, '15-20', NULL, '10', '~1.5 MINS');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM ExerciseDetails
WHERE (exercise_id, week_start, week_end) IN (
  (9, 5, 8),
  (10, 5, 8),
  (11, 5, 8),
  (12, 5, 8),
  (13, 5, 8),
  (14, 5, 8),
  (15, 5, 8)
);
-- +goose StatementEnd
