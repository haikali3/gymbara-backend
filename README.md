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
- [Stripe](#stripe)


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
│   ├── payment/              # Stripe
│   ├── middleware/
│   │   ├── authmiddleware.go
│   │   ├── cors.go
│   │   └── rate_limit.go
│   └── routes/               # Route definitions
│   │   └── routes.go
│   └── webhook/
├── pkg/                      # Shared utilities and helpers 
│   ├── cache/                # Cache management
│   ├── proto/                # Protocol Buffers files
│   ├── models/               # Data models
│   │   ├── google_user.go
│   │   └── exercise.go
│   └── utils/                #
│       ├── helpers.go        # Helper functions(handle http errors, response, etc)
│       └── logger.go         # Zap logger integration
├── proto/                    # Protocol Buffers files
│   └── workout.proto         # gRPC service definition for workout history
├── .env.example              # Example environment variables file
├── README.md                 # Project documentation
├── Makefile                  # Makefile for automating tasks
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

5. Run the server:

```bash
# Run with live reload using Air (recommended for development)
make run

# OR run without live reload
go run cmd/main.go
```

### Development with Air (Live Reload)

The project uses Air for live reload during development. The `make run` command automatically uses Air to watch for file changes and restart the server.

If you want to use Air directly:

1. Install Air:
```bash
go install github.com/cosmtrek/air@latest
```

2. For less typing, add this alias to your shell configuration:
```bash
# Add to ~/.bashrc or ~/.zshrc
alias air='~/.air'
```

3. Run the server with Air:
```bash
air
```

## Database Setup

### 1. PostgreSQL Installation

#### Option 1: Docker Installation

Before running PostgreSQL with Docker, you need to have either Colima, Docker Desktop, or Podman installed:

##### For macOS:
```bash
# Install Colima (recommended for macOS)
# Documentation: https://github.com/abiosoft/colima
brew install colima
colima start
# Check if Colima is running
colima status
# Stop Colima
colima stop


# OR install Docker Desktop
# Documentation: https://docs.docker.com/desktop/install/mac-install/
brew install --cask docker

# OR install Podman
# Documentation: https://podman.io/getting-started/installation
brew install podman
podman machine init
podman machine start
```

After installing either Colima, Docker Desktop, or Podman, you can proceed with running PostgreSQL:

```bash
# Pull the PostgreSQL image
# Documentation: https://hub.docker.com/_/postgres
docker pull postgres:latest

# Run PostgreSQL container
docker run --name gymbara-postgres \
    -e POSTGRES_USER=youruser \
    -e POSTGRES_PASSWORD=yourpassword \
    -e POSTGRES_DB=gymbara \
    -p 5432:5432 \
    -d postgres:latest
```

To stop and remove the container:
```bash
docker stop gymbara-postgres
docker rm gymbara-postgres
```

To view logs:
```bash
docker logs gymbara-postgres
```

#### Option 2: Native Installation

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

3. Set up database environment variables:

You can set these variables in your `.env` file or export them in your shell:

```bash
# Database configuration
DB_USER=youruser
DB_PASSWORD=yourpassword
DB_HOST=localhost
DB_PORT=5432
DB_NAME=gymbara
```

Or export them directly in your shell:

```bash
export DB_USER=youruser
export DB_PASSWORD=yourpassword
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=gymbara
```

The Makefile will use these environment variables for all database operations. If not set, it will use the default values:
- DB_USER: postgres
- DB_PASSWORD: postgres
- DB_HOST: localhost
- DB_PORT: 5432
- DB_NAME: gymbara

### 3. Schema Setup and Data Seeding

The project uses Goose for database migrations and seeding. Follow these steps to set up your database:

1. Run migrations to create the schema:
```bash
make migrate-up
```

2. Seed the database with initial data:
```bash
make seed-up
```

3. If you need to reset everything and start fresh:
```bash
make reset-db    # Drops and recreates the database
make migrate-up  # Creates the schema
make seed-up     # Seeds the data
```

4. To check the status of migrations and seeds:
```bash
make migrate-status  # Check migration status
make seed-status     # Check seed status
```

The migrations will create the following tables:
- WorkoutSections
- Exercises
- ExerciseDetails
- Instructions
- and more...

And the seeds will populate these tables with initial workout data including:
- Workout sections (Full Body, Upper Body, Lower Body)
- Sample exercises with their details
- Exercise instructions and substitutions

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
![alt text](image-1.png)

## Additional Workout Features with gRPC

### 6. Group Workouts

Enable users to participate in group workouts with friends or other users. This can be implemented using gRPC bidirectional streaming.

#### Client-Server Interaction

- **Client**: Sends and receives workout data to/from the server in real-time.
- **Server**: Manages the group workout session and broadcasts data to all participants.

## Stripe

### Webhook Setup

To handle Stripe webhooks locally, use the Stripe CLI to forward events to your local server.

1. Start listening for events and forward them to your local server:

```bash
stripe listen --forward-to http://localhost:8080/webhook/stripe
```

2. New Terimanl: Trigger a `checkout.session.completed` event to test your webhook endpoint:

```bash
stripe trigger checkout.session.completed
```

You should see output similar to the following:

```plaintext
Setting up fixture for: product
Running fixture for: product
Setting up fixture for: price
Running fixture for: price
Setting up fixture for: checkout_session
Running fixture for: checkout_session
Setting up fixture for: payment_page
Running fixture for: payment_page
Setting up fixture for: payment_method
Running fixture for: payment_method
Setting up fixture for: payment_page_confirm
Running fixture for: payment_page_confirm
Trigger succeeded! Check dashboard for event details.
```

Check your Stripe dashboard to verify that the event was received and processed correctly.

3. Cancel subscription (check the subscription id from the stripe dashboard)

```bash
stripe cancel subscription sub_123456
```