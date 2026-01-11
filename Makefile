.PHONY: help migrate-up migrate-down migrate-status migrate-create migrate-reset build run test

# Add Go bin to PATH
export PATH := $(shell go env GOPATH)/bin:$(PATH)

# Load environment variables from .env file if it exists
ifneq (,$(wildcard .env))
    include .env
    export
endif

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# Database migration commands
migrate-up: ## Apply all pending migrations
	@echo "Running migrations..."
	goose -dir migrations postgres "${DB_URL}" up

migrate-down: ## Rollback the last migration
	@echo "Rolling back last migration..."
	goose -dir migrations postgres "${DB_URL}" down

migrate-status: ## Check migration status
	@echo "Checking migration status..."
	goose -dir migrations postgres "${DB_URL}" status

migrate-create: ## Create a new migration (use name=your_migration_name)
	@if [ -z "$(name)" ]; then \
		echo "Error: Please provide a migration name. Usage: make migrate-create name=your_migration_name"; \
		exit 1; \
	fi
	goose -dir migrations create $(name) sql

migrate-reset: ## Rollback all migrations
	@echo "Resetting all migrations..."
	goose -dir migrations postgres "${DB_URL}" reset

# Build and run
build: ## Build the application
	@echo "Building application..."
	go build -o bin/api ./cmd/api

run: ## Run the application
	@echo "Running application..."
	go run ./cmd/api

dev: ## Run the application with air (hot reload)
	@echo "Running application with hot reload..."
	air

# Testing
test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Docker
docker-up: ## Start docker services
	docker-compose up -d

docker-down: ## Stop docker services
	docker-compose down

docker-logs: ## Show docker logs
	docker-compose logs -f

