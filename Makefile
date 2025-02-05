# Variables
DB_URL := postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable
APP_NAME := gymbara-backend

# Default target
.PHONY: all
all: run

# Run the app in development
run:
	echo 1 | go run cmd/main.go

# Run the app in production
run-prod:
	echo 2 | go run cmd/main.go

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
