# Variables
# DB_URL := postgres://postgres:postgres@localhost:5432/gymbara?sslmode=disable
DB_USER ?= youruser
DB_PASSWORD ?= yourpassword
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= gymbara
DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

APP_NAME := gymbara-backend

# Run the app in development
run:
	air

# Run the app in production( add variables)
run-prod:
	@echo "Starting Gymbara Backend in Production Mode..."
	APP_ENV=production ./bin/$(APP_NAME)

# Build & Run the app in Production Mode
build-prod:
	@echo "Building production binary..."
	GOOS=linux GOARCH=amd64 go build -o bin/$(APP_NAME) ./cmd/main.go

# Create a new migration file with a specific name
create-migration:
	@read -p "Enter migration name (use underscores instead of spaces, e.g., 'add_users_table'): " NAME; \
	goose -dir internal/database/migrations create $$NAME sql

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

# Create a new seed file
create-seed:
	@read -p "Enter seed name (use underscores, e.g., 'add_workout_seed'): " NAME; \
	goose -dir internal/database/seeds \
      -table goose_seed_version \
      create $$NAME sql

# Run only seeds
seed-up:
	goose -dir internal/database/seeds \
      -table goose_seed_version \
      postgres "$(DB_URL)" up

# Roll back the last seed batch
seed-down:
	goose -dir internal/database/seeds \
      -table goose_seed_version \
      postgres "$(DB_URL)" down

seed-status:
	goose -dir internal/database/seeds \
      -table goose_seed_version \
      postgres "$(DB_URL)" status

# apply seeds up to a specific version
seed-up-to:
	goose -dir internal/database/seeds \
      -table goose_seed_version \
      postgres "$(DB_URL)" up-to $(VERSION)

# roll back seeds down to a specific version
seed-down-to:
	goose -dir internal/database/seeds \
      -table goose_seed_version \
      postgres "$(DB_URL)" down-to $(VERSION)

seed-reset:
	@echo "🧹 wiping out data & schemas…"
	psql $(DB_URL) \
	  -c "DELETE FROM exercisedetails; DELETE FROM exercises; DELETE FROM workoutsections;"
	psql $(DB_URL) \
	  -c "ALTER SEQUENCE exercisedetails_id_seq RESTART WITH 1; ALTER SEQUENCE exercises_id_seq RESTART WITH 1; ALTER SEQUENCE workoutsections_id_seq RESTART WITH 1;"
	psql $(DB_URL) \
	  -c "TRUNCATE TABLE goose_seed_version;"
	@echo "🚀 re-seeding from zero…"
	goose -dir internal/database/seeds \
	  -table goose_seed_version \
	  postgres "$(DB_URL)" up


# Migrations and Seed
migrate-and-seed: migrate-up
	goose -dir internal/database/seeds -table goose_seed_version \
      postgres "$(DB_URL)" up

# reset gymbara DB
reset-db:
	@echo "👉 Dropping $(DB_NAME) DB…"
	dropdb --if-exists $(DB_NAME)
	@echo "👉 Recreating $(DB_NAME) DB…"
	createdb $(DB_NAME)

# Lint code
lint:
	golangci-lint run

# Run everything
dev: fmt lint test run

# Protobuf
# Directories
PROTO_DIR=proto
OUT_DIR=pkg

# generate gRPC code in pkg/proto folder
generate-proto:
	protoc --go_out=$(OUT_DIR) \
					--go_opt=paths=source_relative \
					--go-grpc_out=$(OUT_DIR) \
					--go-grpc_opt=paths=source_relative \
					$(PROTO_DIR)/workout.proto

# Clean Generated Files
clean-proto:
# rm -f $(OUT_DIR)/workout.pb.go $(OUT_DIR)/workout_grpc.pb.go
	rm -f pkg/proto/workout.pb.go pkg/proto/workout_grpc.pb.go

# Regenerate Protobuf Files
regen-proto: clean-proto generate-proto

run-server:
	go run cmd/grpc_server/workout_server.go

run-client:
	go run cmd/grpc_client/workout_client.go

run-both:
	@echo "Starting gRPC Server and Client..."
	@go run cmd/grpc_server/workout_server.go & \
	sleep 2 && \
	go run cmd/grpc_client/workout_client.go