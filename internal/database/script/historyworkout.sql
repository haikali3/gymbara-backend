
SELECT ued.exercise_id, e.name, ued.custom_load, ued.custom_reps, ued.submitted_at
FROM UserExercisesDetails ued
JOIN Exercises e ON ued.exercise_id = e.id
JOIN UserWorkouts uw ON ued.user_workout_id = uw.id
WHERE uw.user_id = 1
ORDER BY ued.submitted_at ASC

SELECT * FROM UserExercisesDetails;

SELECT ued.exercise_id, e.name, ued.custom_reps, ued.custom_load, ued.submitted_at
FROM UserExercisesDetails ued
JOIN Exercises e ON ued.exercise_id = e.id
JOIN UserWorkouts uw ON ued.user_workout_id = uw.id
WHERE uw.user_id = 1 AND ued.submitted_at BETWEEN '2025-01-01' AND '2025-03-01';

SELECT ued.submitted_at
FROM UserExercisesDetails ued
JOIN UserWorkouts uw ON ued.user_workout_id = uw.id
WHERE uw.user_id = 1
ORDER BY ued.submitted_at ASC;
