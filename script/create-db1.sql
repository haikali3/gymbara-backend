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
    load INT(10),
    rpe INT(10),
    rest_time VARCHAR(20)
);

-- 4. Create the Instructions table
CREATE TABLE Instructions (
    id SERIAL PRIMARY KEY,
    exercise_id INT REFERENCES Exercises(id) ON DELETE CASCADE,
    instruction TEXT NOT NULL
);

-- Sample INSERT queries for data

-- Insert into Sections
INSERT INTO WorkoutSections (name) VALUES 
('Full Body'), 
('Upper Body'), 
('Lower Body');

-- Insert into Exercises
INSERT INTO Exercises (name, workoutsection_id, notes, substitution_1, substitution_2) VALUES 
('Incline Machine Press', 1, '45° incline, focus on squeezing chest', 'Incline Smith Machine Press', 'Incline DB Press'),
('Single-Leg Leg Press (Heavy)', 1, 'High and wide foot positioning, start with weaker leg', 'Machine Squat', 'Hack Squat');

-- Insert into ExerciseDetails
INSERT INTO ExerciseDetails (exercise_id, week_start, week_end, warmup_sets, working_sets, reps, load, rpe, rest_time) VALUES 
(1, 5, 8, 1, 2, '8-10', '63.6', '9-10', '~2 MINS'),
(2, 5, 8, 2, 1, '6-8 per leg', '54.5kg', '8-9', '~3 MINS');

-- Insert into Instructions
INSERT INTO Instructions (exercise_id, instruction) VALUES 
(1, '45° incline, focus on squeezing your chest'),
(2, 'High and wide foot positioning, start with weaker leg');
