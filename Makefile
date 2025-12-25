.PHONY: help build run test clean docker-build docker-up docker-down deps lint

# Variables
APP_NAME=chaos-api-proxy
DOCKER_IMAGE=chaos-api-proxy:latest

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download Go dependencies
	go mod download
	go mod tidy

build: ## Build the application
	go build -o bin/$(APP_NAME) ./cmd/server

run: ## Run the application locally
	go run ./cmd/server/main.go

test: ## Run tests
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

test-coverage: test ## Run tests and show coverage
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.txt coverage.html

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-up: ## Start services with Docker Compose
	docker-compose up -d

docker-down: ## Stop services
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f

docker-restart: docker-down docker-up ## Restart Docker services

dev: ## Run in development mode with hot reload
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

# Example configurations
create-example-config: ## Create an example configuration
	@echo "Creating example configuration..."
	@curl -X POST http://localhost:8080/api/v1/configs \
		-H "Content-Type: application/json" \
		-d '{
			"name": "Stripe API Chaos Test",
			"description": "Test configuration for Stripe API",
			"target": "https://api.stripe.com",
			"enabled": true,
			"rules": {
				"latency_ms": 500,
				"jitter": 200,
				"inject_failure_rate": 0.1,
				"error_code": 503
			}
		}'

list-configs: ## List all configurations
	@curl -s http://localhost:8080/api/v1/configs | jq .

health-check: ## Check service health
	@curl -s http://localhost:8080/health | jq .
