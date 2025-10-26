.PHONY: help build run test test-unit test-integration clean docker-up docker-down

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building application..."
	@go build -o bin/server cmd/server/main.go

run: ## Run the application
	@echo "Running application..."
	@go run cmd/server/main.go

test: ## Run all tests
	@echo "Running all tests..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	@go test -v -race -short ./...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -race -run Integration ./...

coverage: test ## Generate coverage report
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/"; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	@go mod tidy

docker-up: ## Start MongoDB in Docker
	@echo "Starting MongoDB..."
	@docker run -d --name mongo-news -p 27017:27017 mongo:7

docker-down: ## Stop MongoDB Docker container
	@echo "Stopping MongoDB..."
	@docker stop mongo-news || true
	@docker rm mongo-news || true

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.txt coverage.html
	@go clean

install-deps: ## Install dependencies
	@echo "Installing dependencies..."
	@go mod download

all: clean tidy fmt vet build test ## Run all build steps
