
-- +goose Up
-- +goose StatementBegin

-- Insert Exercises (Upper Body)
INSERT INTO Exercises (name, workout_section_id, notes, substitution_1, substitution_2) VALUES
('2-Grip Pullup', 2, 'First set 1.5x shoulder width grip. Second set 1.0x shoulder width grip.', 'Machine Pulldown', '2-Grip Lat Pulldown'),
('Weighted Dip(Heavy)', 2, 'Tuck your elbows at 45°, lean your torso forward 15°, shoulder width or slightly wider grip.', 'Machine Chest Press', 'Flat DB Press'),
('Weighted Dip(Back off)', 2, 'Tuck your elbows at 45°, lean your torso forward 15°, shoulder width or slightly wider grip.', 'Machine Chest Press', 'Flat DB Press'),
('Incline Chest-Supported DB Row', 2, 'Keep elbows at ~30° angle from torso. Pull the weight towards your navel.', 'Chest-Supported T-Bar Row', 'Seated Cable Row'),
('Standing DB Arnold Press', 2, 'Start with your elbows in front of you and palms facing in. Rotate the dumbbells so that your palms face forward as you press.', 'Machine Shoulder Press', 'Seated DB Shoulder Press'),
('A1: DB Incline Curl', 2, 'Brace upper back against bench 45° incline, keep shoulder back as you curl.', 'Cable EZ Curl', 'EZ Bar Curl'),
('A2: DB French Press', 2, 'Can perform seated or standing. Press the dumbbell straight up and down behind your head.', 'Overhead Cable Triceps Extension', 'EZ Bar Skull Crusher');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Delete Upper Body Exercises
DELETE FROM Exercises
WHERE name IN (
  '2-Grip Pullup',
  'Weighted Dip(Heavy)',
  'Weighted Dip(Back off)',
  'Incline Chest-Supported DB Row',
  'Standing DB Arnold Press',
  'A1: DB Incline Curl',
  'A2: DB French Press'
);


-- +goose StatementEnd
