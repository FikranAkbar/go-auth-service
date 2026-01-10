.PHONY: help dev build run test clean docker-up docker-down migrate-up migrate-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Run application with hot reload (Air)
	air

build: ## Build the application
	go build -o bin/api cmd/api/main.go

run: build ## Build and run the application
	./bin/api

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf tmp/
	rm -f build-errors.log
	rm -f coverage.out coverage.html

docker-up: ## Start Docker services (PostgreSQL & Redis)
	docker-compose up -d

docker-down: ## Stop Docker services
	docker-compose down

docker-logs: ## View Docker services logs
	docker-compose logs -f

install-tools: ## Install development tools (Air, migrate, etc)
	go install github.com/air-verse/air@latest
	@echo "Tools installed successfully!"

tidy: ## Tidy go modules
	go mod tidy

fmt: ## Format code
	go fmt ./...

lint: ## Run linter (requires golangci-lint)
	golangci-lint run

# Add your database connection string here
DB_URL ?= postgres://user:password@localhost:5432/dbname?sslmode=disable

migrate-up: ## Run database migrations up
	migrate -path migrations -database "$(DB_URL)" up

migrate-down: ## Rollback database migrations
	migrate -path migrations -database "$(DB_URL)" down 1

migrate-create: ## Create new migration file (usage: make migrate-create NAME=create_table_name)
	migrate create -ext sql -dir migrations -seq $(NAME)

