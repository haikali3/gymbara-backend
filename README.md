# Gymbara Backend API

A comprehensive Golang-based REST API for managing workout programs, built with native Go HTTP server and PostgreSQL.

## Table of Contents
- [Project Overview](#project-overview)
- [Directory Structure](#directory-structure)
- [Prerequisites](#prerequisites)
- [Local Development Setup](#local-development-setup)
- [Database Setup](#database-setup)
- [API Endpoints](#api-endpoints)
- [Code Structure](#code-structure)
- [Deployment](#deployment)
- [Environment Variables](#environment-variables)

## Project Overview

This backend service provides APIs for:
- Managing workout sections (Full Body, Upper Body, Lower Body)
- Exercise management with detailed attributes
- Exercise details including sets, reps, and rest times
- Instructions and substitution exercises

## Directory Structure

```plaintext
.
├── config/
│   └── config.go           # Environment configuration
├── controllers/
│   └── exercises.go        # HTTP request handlers
├── models/
│   └── models.go          # Data structures
├── routes/
│   └── routes.go          # Route definitions and CORS
├── database/
│   └── db.go              # Database connection
├── migrations/
│   └── *.sql              # SQL migration files
├── .env                    # Environment variables
├── main.go                # Application entry point
└── README.md
```

## Prerequisites

- Go 1.18+
- PostgreSQL 12+
- [godotenv](https://github.com/joho/godotenv)
- [goose](https://github.com/pressly/goose) (optional, for migrations)

## Local Development Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd gymbara-backend
```

2. Install dependencies:
```bash
go mod tidy
```

3. Copy the example environment file:
```bash
cp .env.example .env
```

4. Configure your local `.env`:
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gymbara
SERVER_PORT=8080
```

5. Run the server:
```bash
go run main.go
```

## Database Setup

### 1. PostgreSQL Installation

#### Ubuntu/Debian:
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
```

#### macOS (using Homebrew):
```bash
brew install postgresql
brew services start postgresql
```

### 2. Database Creation

1. Access PostgreSQL:
```bash
sudo -u postgres psql
```

2. Create database and user:
```sql
CREATE DATABASE gymbara;
CREATE USER youruser WITH PASSWORD 'yourpassword';
GRANT ALL PRIVILEGES ON DATABASE gymbara TO youruser;
```

### 3. Schema Setup

Run the following SQL scripts:

```sql
-- 1. Create the Sections table
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
```

### 4. Sample Data

```sql
-- Insert workout sections
INSERT INTO WorkoutSections (name) VALUES 
('Full Body'), 
('Upper Body'), 
('Lower Body');

-- Insert exercises
INSERT INTO Exercises (name, workoutsection_id, notes, substitution_1, substitution_2) VALUES 
('Incline Machine Press', 1, '45° incline, focus on squeezing chest', 'Incline Smith Machine Press', 'Incline DB Press'),
('Single-Leg Leg Press (Heavy)', 1, 'High and wide foot positioning, start with weaker leg', 'Machine Squat', 'Hack Squat');
```

## API Endpoints

### Get Workout Sections
```
GET /workout-sections
Response: [{"id": 1, "name": "Full Body"}, ...]
```

### Get Exercise List
```
GET /workout-sections/list?workout_section_id=1
Response: [{"name": "Exercise Name", "reps": "8-10", ...}, ...]
```

### Get Exercise Details
```
GET /workout-sections/details?workout_section_id=1
Response: [{"name": "Exercise Name", "warmup_sets": 2, ...}, ...]
```

## Code Structure

### Configuration (config/config.go)
```go
type Config struct {
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string
    ServerPort string
}

func LoadConfig() *Config {
    // Loads environment variables
    // Returns configuration struct
}
```

### Models (models/models.go)
```go
type Exercise struct {
    ID           int    `json:"id"`
    ExerciseName string `json:"name"`
    // ... other fields
}

type WorkoutSection struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}
```

### Controllers (controllers/exercises.go)
```go
func GetWorkoutSections(w http.ResponseWriter, r *http.Request) {
    // Handles workout sections endpoint
}

func GetExercisesList(w http.ResponseWriter, r *http.Request) {
    // Handles exercise list endpoint
}
```

## Deployment

### 1. Building for Production

Create a production build:
```bash
go build -o gymbara-backend
```

### 2. Docker Deployment

1. Create Dockerfile:
```dockerfile
FROM golang:1.18-alpine

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o gymbara-backend

EXPOSE 8080

CMD ["./gymbara-backend"]
```

2. Build and run:
```bash
docker build -t gymbara-backend .
docker run -p 8080:8080 --env-file .env gymbara-backend
```

### 3. Production Environment Variables

Create a production `.env`:
```bash
DB_HOST=production-db-host
DB_PORT=5432
DB_USER=production-user
DB_PASSWORD=secure-password
DB_NAME=gymbara
SERVER_PORT=8080
```

### 4. Database Migrations (Optional)

Set up Goose for migrations:
```bash
export DATABASE_URL="postgres://user:password@host:5432/dbname?sslmode=disable"
goose up                    # Apply migrations
goose down                  # Rollback last migration
goose status               # Check migration status
```

## Security Considerations

1. CORS Configuration (routes/routes.go):
```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
        // ... other CORS headers
    })
}
```

2. Environment Variables:
- Never commit `.env` files
- Use different `.env` files for development and production
- Use secure passwords in production
- Enable SSL in production database connections

## Monitoring and Logging

1. Basic logging is implemented using the standard `log` package
2. Consider implementing structured logging for production
3. Add monitoring metrics for:
   - Request duration
   - Database connection pool stats
   - Error rates
   - Endpoint usage