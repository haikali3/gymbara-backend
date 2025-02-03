SELECT * FROM users;

SELECT * FROM ExerciseDetails;
SELECT * FROM Exercises;
SELECT * FROM UserExercisesDetails;
SELECT * FROM UserWorkouts;
SELECT * FROM WorkoutSections;

ALTER TABLE UserExercisesDetails
ADD COLUMN submitted_at DATE;

SELECT 
    UserExercisesDetails.*, 
    Exercises.name AS exercise_name
FROM 
    UserExercisesDetails
JOIN 
    Exercises ON UserExercisesDetails.exercise_id = Exercises.id;


DELETE FROM UserExercisesDetails 
WHERE submitted_at >= '2025-02-01';


