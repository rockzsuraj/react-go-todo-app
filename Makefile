.PHONY: help build up down restart logs clean deploy health test lint format

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Development commands
build: ## Build all containers
	docker-compose build

up: ## Start all services
	docker-compose up -d

down: ## Stop all services
	docker-compose down

restart: ## Restart all services
	docker-compose restart

logs: ## Show logs for all services
	docker-compose logs -f

logs-backend: ## Show backend logs
	docker-compose logs -f backend

logs-frontend: ## Show frontend logs
	docker-compose logs -f frontend

logs-db: ## Show database logs
	docker-compose logs -f db

# Production commands
deploy: ## Deploy to production with health checks
	chmod +x deploy.sh
	./deploy.sh deploy

rollback: ## Rollback to previous version
	./deploy.sh rollback

health: ## Check service health
	./deploy.sh health

# Development tools
clean: ## Clean up containers, images, and volumes
	docker-compose down -v
	docker system prune -f
	docker volume prune -f

rebuild: ## Rebuild and restart everything
	make down
	docker-compose build --no-cache
	make up

# Testing
test-backend: ## Run backend tests
	docker-compose exec backend go test ./...

test-frontend: ## Run frontend tests
	docker-compose exec frontend npm test

# Code quality
lint-backend: ## Lint backend code
	docker-compose exec backend go vet ./...
	docker-compose exec backend golangci-lint run

lint-frontend: ## Lint frontend code
	docker-compose exec frontend npm run lint

format-backend: ## Format backend code
	docker-compose exec backend go fmt ./...

format-frontend: ## Format frontend code
	docker-compose exec frontend npm run format

# Database operations
db-migrate: ## Run database migrations
	docker-compose exec backend ./migrate

db-seed: ## Seed database with sample data
	docker-compose exec db psql -U postgres -d todos -f /docker-entrypoint-initdb.d/seed.sql

db-backup: ## Backup database
	docker-compose exec db pg_dump -U postgres todos > backup_$(shell date +%Y%m%d_%H%M%S).sql

db-restore: ## Restore database (usage: make db-restore FILE=backup.sql)
	docker-compose exec -T db psql -U postgres todos < $(FILE)

# Monitoring
ps: ## Show running containers
	docker-compose ps

stats: ## Show container resource usage
	docker stats

# Quick development setup
dev-setup: ## Quick development environment setup
	make build
	make up
	@echo "Waiting for services to start..."
	sleep 10
	make health
	@echo "Development environment ready!"
	@echo "Frontend: http://localhost:3000"
	@echo "Backend: http://localhost:8080"
	@echo "pgAdmin: http://localhost:8081"