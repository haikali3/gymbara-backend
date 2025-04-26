-- +goose Up
-- +goose StatementBegin
-- 1. Create the Sections table (Upper body, Lower, Full)
CREATE TABLE WorkoutSections (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL
);

-- 2. Create the Exercises table
CREATE TABLE Exercises (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    workoutsection_id INT REFERENCES WorkoutSections(id) ON DELETE CASCADE,
    notes TEXT,
    substitution_1 VARCHAR(100),
    substitution_2 VARCHAR(100)
);

-- 3. Create the ExerciseDetails table
CREATE TABLE ExerciseDetails (
    id SERIAL PRIMARY KEY,
    exercise_id INT REFERENCES Exercises(id) ON DELETE CASCADE,
    week_start INT NOT NULL,
    week_end INT NOT NULL,
    warmup_sets INT,
    working_sets INT,
    reps VARCHAR(20),
    load INT,
    rpe INT,
    rest_time VARCHAR(20)
);

-- 4. Create the Instructions table
CREATE TABLE Instructions (
    id SERIAL PRIMARY KEY,
    exercise_id INT REFERENCES Exercises(id) ON DELETE CASCADE,
    instruction TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE Instructions;
DROP TABLE ExerciseDetails;
DROP TABLE Exercises;
DROP TABLE WorkoutSections;
-- +goose StatementEnd
