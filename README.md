# Gymbara Backend API

A comprehensive Golang-based REST API for managing workout programs, built with native Go HTTP server and PostgreSQL.

## Table of Contents

- [Project Overview](#project-overview)
- [Directory Structure](#directory-structure)
- [Prerequisites](#prerequisites)
- [Local Development Setup](#local-development-setup)
- [Database Setup](#database-setup)
- [API Endpoints](#api-endpoints)
- [Deployment](#deployment)
- [Environment Variables](#environment-variables)
- [ERD Diagram](#erd-diagram)

## Project Overview

This backend service provides APIs for:

- Managing workout sections (Full Body, Upper Body, Lower Body)
- Exercise management with detailed attributes
- Exercise details including sets, reps, and rest times
- Instructions and substitution exercises

## Directory Structure

```plaintext
.gymbara-backend/
├── cmd/                      # Entry points for the application
│   └── main.go               # Main application entry point
├── config/                   # Configuration management
│   └── config.go
├── internal/                 # Internal application logic
│   ├── auth/                 # Authentication logic
│   │   ├── oauth.go
│   ├── controllers/          # HTTP handlers
│   │   ├── user_controller.go
│   │   └── exercise_controller.go
│   ├── database/             # Database connection and queries
│   │   ├── migrations/       # Database migrations 
│   │   ├── script/           # Script for creating tables 
│   │   └── db.go             # Database connection
│   │   └── statements.go     # SQL statements
│   ├── middleware/           # Middleware functions
│   │   ├── authmiddleware.go
│   │   ├── cors.go
│   │   └── rate_limit.go
│   └── routes/               # Route definitions
│       └── routes.go
├── pkg/                      # Shared utilities and helpers 
│   ├── models/               # Data models
│   │   ├── google_user.go
│   │   └── exercise.go
│   └── utils/                # Utility functions
│       ├── helpers.go        # Helper functions(handle http errors, response, etc)
│       └── logger.go         # Added zap logger
├── .env.example              # Example environment variables file
├── go.mod                    # Go module definition
├── go.sum                    # Go module dependencies
├── README.md                 # Project documentation
└── Dockerfile                # Docker configuration (if needed)
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

5. Set the environment for development or production on your server

```bash
# For development
export APP_ENV=development

# For production
export APP_ENV=production

```

6. Run the server:

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

| Endpoint                                | Method | Description                                                                                              | Example Request                                                                                            | Example Response                                                                                                  |
| --------------------------------------- | ------ | -------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| **Get Workout Sections**                | GET    | Retrieves all workout sections (e.g., Full Body, Upper Body, etc.).                                      | `/workout-sections`                                                                                        | `[{"id": 1, "name": "Full Body"}, {"id": 2, "name": "Upper Body"}, ...]`                                          |
| **Get Exercise List**                   | GET    | Retrieves a list of exercises based on the workout section ID.                                           | `/workout-sections/list?workout_section_id=1`                                                              | `[{"name": "Exercise Name", "reps": "8-10"}, ...]`                                                                |
| **Get Exercise Details**                | GET    | Retrieves detailed information about exercises, including warmup sets, work sets, reps, etc.             | `/workout-sections/details?workout_section_id=1`                                                           | `[{"name": "Exercise Name", "warmup_sets": 2, "work_sets": 3, ...}, ...]`                                         |
| **Login with Google**                   | GET    | Redirects users to Google’s OAuth login page for authentication.                                         | `/oauth/login`                                                                                             | Redirects to Google OAuth login page                                                                              |
| **Google OAuth Callback**               | GET    | Handles the callback from Google after successful authentication. Retrieves and stores user information. | `/oauth/callback`                                                                                          | Redirects to frontend with auth token (set as cookie)                                                             |
| **Get Workout Sections With Exercises** | GET    | Retrieves workout sections along with their associated exercises.                                        | `/workout-sections/with-exercises?workout_section_ids=1,2`                                                 | `[{"id": 1, "name": "Full Body", "route": "/full-body", "exercises": [...]}, {"id": 2, ...}]`                     |
| **User Submit Exercises**               | POST   | Submits or updates user-specific exercise details, including reps and load.                              | `/workout-sections/user-exercise-details` Body: `{ "section_id": 1, "user_email": "test@gmail.com", ... }` | `{ "message": "user exercise details submitted successfully", "user_workout_id": 1, "updated_exercises": [...] }` |

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

### 4. Database Migrations and Development Tools

The project uses a Makefile to automate common development tasks and database migrations using Goose.

#### 4.1 Creating New Migrations
Set up Goose for migrations:

1. Create script

```bash
goose create add_users_table sql
goose create <script_name> sql
```


2. Apply goose migrations:
```bash
export DATABASE_URL="postgres://user:password@host:5432/dbname?sslmode=disable"
goose up                    # Apply migrations
goose down                  # Rollback last migration
goose status               # Check migration status
```


#### 4.2 Using the Makefile

Run the following commands to automate tasks:

* **Run the app**:
```bash
make run
```

* **Build the app**:
```bash
make build
```

* **Apply migrations**:
```bash
make migrate-up
```

* **Rollback migrations**:
```bash
make migrate-down
```

* **Check migration status**:
```bash
make migrate-status
```

* **Run tests**:
```bash
make test
```

* **Format code**:
```bash
make fmt
```

* **Run full development cycle** (format, lint, test, run):
```bash
make dev
```

#### 5. Testing
To run the tests with benchmarks and memory allocation statistics, use the following command:

```bash
go test -bench . -benchmem
```
This command will execute all benchmarks in the current package and provide detailed memory allocation information.

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

## ERD Diagram

![alt text](image.png)
