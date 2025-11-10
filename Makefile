.PHONY: build run clean test migrate help

# Load environment variables from .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Build the application
build:
	@echo "Building application..."
	@go build -o bin/supreme-broccoli cmd/server/main.go
	@echo "Build complete: bin/supreme-broccoli"

# Run the application (loads .env automatically)
run:
	@echo "Starting application..."
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Please create it with required variables."; \
		exit 1; \
	fi
	@go run cmd/server/main.go

# Run the built binary (loads .env automatically)
start:
	@echo "Starting application from binary..."
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Please create it with required variables."; \
		exit 1; \
	fi
	@if [ ! -f bin/supreme-broccoli ]; then \
		echo "Error: Binary not found. Run 'make build' first."; \
		exit 1; \
	fi
	@./bin/supreme-broccoli

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f supreme-broccoli migrate
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Build migration tool
migrate-build:
	@echo "Building migration tool..."
	@go build -o bin/migrate migrate.go
	@echo "Migration tool built: bin/migrate"

# Run migration
migrate-run: migrate-build
	@echo "Running migration..."
	@./bin/migrate

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted"

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Show help
help:
	@echo "Available commands:"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the application (loads .env)"
	@echo "  make start         - Run the built binary (loads .env)"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make test          - Run tests"
	@echo "  make migrate-build - Build migration tool"
	@echo "  make migrate-run   - Run migration"
	@echo "  make deps          - Install dependencies"
	@echo "  make fmt           - Format code"
	@echo "  make lint          - Run linter"
	@echo "  make help          - Show this help message"
