-- file: internal/database/seeds/20250428014215_add_lower_body_exercises.sql
-- +goose Up
-- +goose StatementBegin
INSERT INTO
  Exercises (
    name,
    workoutsection_id,
    notes,
    substitution_1,
    substitution_2
  )
VALUES
  (
    'DB Bulgarian Split Squat',
    3,
    'Start with your weaker leg. Squat deep',
    'Goblet Squat',
    'Leg Press Toe Press'
  ),
  (
    'DB Romanian Deadlift',
    3,
    'Emphasize the stretch in your hamstrings, prevent your lower back from rounding',
    'Romanian Deadlift',
    '45 deg Hyperextension'
  ),
  (
    'Goblet Squat',
    3,
    'Hold the dumbbell underneath your chin, sit back and down, push your knees out laterally',
    'Leg Extension',
    'Step Up'
  ),
  (
    'Leg Press Toe Press',
    3,
    'Press all the way up to your toes, stretch your calves at the bottom, don''t bounce.',
    'Standing Calf Raise',
    'Seated Calf Raise'
  ),
  (
    'Machine Crunch',
    3,
    'Squeeze your abs to move the weight, don''t use your arms to help',
    'Plate-Weighted Crunch',
    'Cable Crunch'
  );

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DELETE FROM Exercises
WHERE
  name IN (
    'DB Bulgarian Split Squat',
    'DB Romanian Deadlift',
    'Goblet Squat',
    'Leg Press Toe Press',
    'Machine Crunch'
  );

-- +goose StatementEnd