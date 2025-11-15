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

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

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
