-- +goose Up
-- +goose StatementBegin

-- Users table to store OAuth users
CREATE TABLE Users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50),
    email VARCHAR(100),
    oauth_provider VARCHAR(50),
    oauth_id VARCHAR(100),
    is_premium BOOLEAN DEFAULT FALSE
);

-- Subscriptions table
CREATE TABLE Subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES Users(id),
    paid_date DATE,
    expiration_date DATE,
    stripe_subscription_id VARCHAR(100)
);

-- UserWorkouts table to link users with workout sections
CREATE TABLE UserWorkouts (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES Users(id),
    section_id INT REFERENCES WorkoutSections(id),
    day_of_week INT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop the UserWorkouts table
DROP TABLE IF EXISTS UserWorkouts;

-- Drop the Subscriptions table
DROP TABLE IF EXISTS Subscriptions;

-- Drop the Users table
DROP TABLE IF EXISTS Users;

-- +goose StatementEnd
