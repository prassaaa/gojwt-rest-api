.PHONY: help build run clean test

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building application..."
	@go build -o bin/api cmd/api/main.go
	@echo "Build completed: bin/api"

run: ## Run the application
	@echo "Running application..."
	@go run cmd/api/main.go

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@echo "Clean completed"

install: ## Install dependencies
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

generate-secret: ## Generate JWT secret
	@go run cmd/tools/generate_secret.go

test: ## Run all tests
	@echo "Running all tests..."
	@go test -v ./test/...

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@go test -v ./test/unit/...

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	@go test -v ./test/integration/...

test-e2e: ## Run E2E tests only
	@echo "Running E2E tests..."
	@go test -v ./test/e2e/...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test ./test/... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out
	@echo "\nCoverage report saved to coverage.out"
	@echo "To view HTML coverage report, run: make test-coverage-html"

test-coverage-html: ## Generate HTML coverage report
	@echo "Generating HTML coverage report..."
	@go test ./test/... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@echo "HTML coverage report generated: coverage.html"

test-watch: ## Run tests in watch mode (requires air)
	@air -c .air.test.toml

test-bench: ## Run benchmark tests
	@echo "Running benchmark tests..."
	@go test -bench=. -benchmem ./test/unit/...

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format completed"

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run
	@echo "Lint completed"

migrate: ## Run database migrations
	@echo "Running migrations..."
	@go run cmd/api/main.go
	@echo "Migrations completed"

dev: ## Run with hot reload (requires air)
	@air

docker-build: ## Build Docker image
	@docker build -t gojwt-rest-api .

docker-run: ## Run Docker container
	@docker-compose up -d

docker-down: ## Stop Docker container
	@docker-compose down
