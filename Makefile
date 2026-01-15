.PHONY: help migrate-up migrate-down migrate-status migrate-create migrate-reset build run test test-coverage test-coverage-html test-coverage-func test-coverage-summary

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
	@awk 'BEGIN {FS = ":"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-30s\033[0m", $$1; $$1=""; sub(/^[^#]*## */, ""); print}' $(MAKEFILE_LIST)

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

test-coverage: ## Run tests and generate coverage report
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./... || true
	@echo ""
	@echo "=== COVERAGE SUMMARY ==="
	@go tool cover -func=coverage.out 2>/dev/null | grep total: || echo "No coverage data available"
	@echo ""
	@echo "Run 'make test-coverage-html' to view detailed HTML report"
	@echo "Run 'make test-coverage-func' to view coverage by function"

test-coverage-html: ## Generate and open HTML coverage report
	@echo "Generating HTML coverage report..."
	@go test -coverprofile=coverage.out ./... || true
	@go tool cover -html=coverage.out
	@echo "Coverage report opened in browser"

test-coverage-func: ## Show coverage by function
	@echo "Coverage by function:"
	@go test -coverprofile=coverage.out ./... > /dev/null 2>&1 || true
	@go tool cover -func=coverage.out 2>/dev/null || echo "No coverage data available"

test-coverage-summary: ## Show coverage summary by package
	@echo "=== TEST COVERAGE SUMMARY ==="
	@echo ""
	@go test ./... -cover 2>&1 | grep -E "coverage:" | sort -t: -k2 -rn || true
	@echo ""
	@go test -coverprofile=coverage.out ./... > /dev/null 2>&1 || true
	@echo "Total coverage:"
	@go tool cover -func=coverage.out 2>/dev/null | grep total: || echo "No coverage data available"

# Docker
docker-up: ## Start docker services
	docker-compose up -d

docker-down: ## Stop docker services
	docker-compose down

docker-logs: ## Show docker logs
	docker-compose logs -f

