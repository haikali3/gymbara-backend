# Variables
DB_URL := postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable
APP_NAME := gymbara-backend

# Default target
.PHONY: all
all: run

# Run the app in development
run:
	echo 1 | go run cmd/main.go

# Build the app
build:
	go build -o $(APP_NAME) cmd/main.go

# Database migrations
migrate-up:
	goose -dir internal/database/migrations postgres "$(DB_URL)" up

migrate-down:
	goose -dir internal/database/migrations postgres "$(DB_URL)" down

migrate-status:
	goose -dir internal/database/migrations postgres "$(DB_URL)" status

migrate-up-to:
	goose -dir internal/database/migrations postgres "$(DB_URL)" up-to $(VERSION)

migrate-down-to:
	goose -dir internal/database/migrations postgres "$(DB_URL)" down-to $(VERSION)

# Test the app
test:
	go test ./... -v

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Run everything
dev: fmt lint test run
