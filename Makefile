.PHONY: help dev dev-docker build build-docker test test-integration test-coverage clean docker-up docker-down docker-logs docker-clean install-deps setup-env setup-env-prod db-setup db-reset db-backup lint format migrate seed health-check deploy-prod deploy-staging

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Start development environment (manual)
	docker-compose up -d postgres
	@echo "Waiting for database..."
	./scripts/db-manage.sh wait bodda_dev
	@echo "Starting backend with hot reloading..."
	air -c .air.toml &
	@echo "Starting frontend..."
	cd frontend && npm run dev

dev-docker: ## Start development environment with Docker
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml up

build: ## Build the application
	go build -o bin/bodda main.go
	cd frontend && npm run build

build-docker: ## Build Docker images
	docker build -f Dockerfile.backend --target production -t bodda-backend .
	docker build -f frontend/Dockerfile --target production -t bodda-frontend ./frontend

test: ## Run all tests
	go test ./...
	cd frontend && npm test -- --run

test-integration: ## Run integration tests
	./scripts/db-manage.sh setup-test
	TEST_DATABASE_URL="postgres://postgres:password@localhost:5433/bodda_test?sslmode=disable" go test -tags=integration ./...

test-coverage: ## Run tests with coverage
	go test -cover ./...
	cd frontend && npm run test:coverage

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf tmp/
	rm -rf frontend/dist/
	rm -rf frontend/node_modules/.cache/

docker-up: ## Start all services with Docker
	docker-compose up -d

docker-down: ## Stop all Docker services
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f

docker-clean: ## Clean Docker resources
	docker-compose down -v
	docker system prune -f

install-deps: ## Install dependencies
	go mod tidy
	cd frontend && npm install

setup-env: ## Set up development environment
	./scripts/setup-env.sh development

setup-env-prod: ## Set up production environment
	./scripts/setup-env.sh production

db-setup: ## Set up development database
	./scripts/db-manage.sh setup-dev

db-reset: ## Reset development database
	./scripts/db-manage.sh reset-dev

db-backup: ## Backup development database
	./scripts/db-manage.sh backup bodda_dev

lint: ## Run linters
	golangci-lint run || echo "golangci-lint not installed, skipping Go linting"
	cd frontend && npm run lint

format: ## Format code
	go fmt ./...
	cd frontend && npm run format

migrate: ## Run database migrations
	./scripts/db-manage.sh migrate bodda_dev

seed: ## Seed development database
	./scripts/db-manage.sh seed bodda_dev seed-dev-data.sql

health-check: ## Check application health
	curl -f http://localhost:8080/monitoring/health || exit 1

# Production deployment targets
deploy-prod: ## Deploy to production
	docker-compose -f docker-compose.prod.yml up -d

deploy-staging: ## Deploy to staging
	BUILD_TARGET=production docker-compose up -d