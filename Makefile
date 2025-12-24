.PHONY: build test clean run-server run-client docker-build docker-up docker-down docker-logs fmt lint

## build: Build server and client binaries
build:
	@echo "Building server..."
	@go build -o bin/server ./cmd/server
	@echo "Building client..."
	@go build -o bin/client ./cmd/client
	@echo "✓ Build complete"

## test: Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

## test-cover: Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	@go test -v -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

## test-integration: Run integration tests only
test-integration:
	@echo "Running integration tests..."
	@go test -v ./internal/tests

## run-server: Run server locally
run-server:
	@echo "Starting server..."
	@go run ./cmd/server

## run-client: Run client locally
run-client:
	@echo "Starting client..."
	@go run ./cmd/client

## docker-build: Build Docker images
docker-build:
	@echo "Building Docker images..."
	@docker-compose build

## docker-up: Start Docker containers
docker-up:
	@echo "Starting Docker containers..."
	@docker-compose up --build

## docker-up-d: Start Docker containers in background
docker-up-d:
	@echo "Starting Docker containers in background..."
	@docker-compose up --build -d

## docker-down: Stop Docker containers
docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose down

## docker-logs: Show Docker logs
docker-logs:
	@docker-compose logs -f

## docker-clean: Remove Docker containers and images
docker-clean:
	@echo "Cleaning Docker resources..."
	@docker-compose down -v --rmi all

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

## lint: Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -f server client
	@echo "✓ Clean complete"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify
	@echo "✓ Dependencies ready"

## tidy: Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy
	@echo "✓ Dependencies tidied"
