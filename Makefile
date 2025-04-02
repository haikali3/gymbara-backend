# Variables
DB_URL := postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable
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