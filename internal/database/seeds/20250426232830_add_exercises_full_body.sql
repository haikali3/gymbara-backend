-- file: internal/database/seeds/20250426232830_add_exercises_full_body.sql
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
    'Incline Machine Press',
    1,
    '45° incline, focus on squeezing your chest.',
    'Incline Smith Machine Press',
    'Incline DB Press'
  ),
  (
    'Single-Leg Leg Press(Heavy)',
    1,
    'High and wide foot positioning, start with weaker leg.',
    'Machine Squat',
    'Hack Squat'
  ),
  (
    'Single-Leg Leg Press(Back off)',
    1,
    'High and wide foot positioning, start with weaker leg.',
    'Machine Squat',
    'Hack Squat'
  ),
  (
    'Pendlay Row',
    1,
    'Initiate the movement by squeezing your shoulder blades together, pull to your lower chest, avoid using momentum.',
    'T-Bar Row',
    'Seated Cable Row'
  ),
  (
    'Glute-Ham Raise',
    1,
    'Keep your hips straight, do Nordic ham curls if no GHR machine.',
    'Nordic Ham Curl',
    'Lying Leg Curl'
  ),
  (
    'Spider Curl',
    1,
    'Dropset: perform 12–15 reps, drop the weight by ~50%, perform an additional 12–15 reps. Brace your chest against an incline bench, curl with your elbows slightly in front of you.',
    'DB Preacher Curl',
    'Bayesian Cable Curl'
  ),
  (
    'Cable Lateral Raise',
    1,
    'Dropset: perform 12–15 reps, drop the weight by ~50%, perform an additional 12–15 reps. Lean away from the cable. Focus on squeezing your delts.',
    'Machine Lateral Raise',
    'DB Lateral Raise'
  ),
  (
    'Hanging Leg Raise',
    1,
    'Knees to chest, controlled reps, straighten legs more to increase difficulty.',
    'Roman Chair Crunch',
    'Reverse Crunch'
  );

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DELETE FROM Exercises
WHERE
  name IN (
    'Incline Machine Press',
    'Single-Leg Leg Press(Heavy)',
    'Single-Leg Leg Press(Back off)',
    'Pendlay Row',
    'Glute-Ham Raise',
    'Spider Curl',
    'Cable Lateral Raise',
    'Hanging Leg Raise'
  );

-- rewind the auto-increment back to 1
ALTER SEQUENCE exercises_id_seq
RESTART WITH 1;

-- +goose StatementEnd