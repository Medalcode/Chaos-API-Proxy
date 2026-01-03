.PHONY: help build run dev test clean docker-build docker-up docker-down

# Variables
APP_NAME=chaos-api-proxy
DOCKER_IMAGE=chaos-api-proxy:latest

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Install dependencies
	npm install

build: ## Build the application (TS -> JS)
	npm run build

run: ## Run the application locally (Production mode)
	npm start

dev: ## Run in development mode with hot reload
	npm run dev

test: ## Run tests
	npm test

clean: ## Clean build artifacts
	rm -rf dist/
	rm -rf node_modules/

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-up: ## Start services with Docker Compose
	docker-compose up -d --build

docker-down: ## Stop services
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f

docker-restart: docker-down docker-up ## Restart Docker services

# Utilities
examples: ## Print example usage
	@echo "See README.md and examples/ directory for usage."
