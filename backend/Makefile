# Backend Makefile — hackathon starter kit
DATABASE_URL ?= postgres://postgres:postgres@127.0.0.1:5433/hackathon?sslmode=disable
GOOSE_DIR    ?= internal/db/migrations

.PHONY: dev run sqlc migrate-up migrate-down migrate-new migrate-status smoke compose-up compose-down compose-logs tidy help

dev: ## Run with hot reload (requires air: go install github.com/air-verse/air@latest)
	air

run: ## Run the API server
	go run ./cmd/api

sqlc: ## Generate type-safe db code from db/queries/*.sql
	sqlc generate

migrate-up: ## Apply migrations (goose CLI)
	goose -dir $(GOOSE_DIR) postgres "$(DATABASE_URL)" up

migrate-down: ## Rollback last migration (goose CLI)
	goose -dir $(GOOSE_DIR) postgres "$(DATABASE_URL)" down

migrate-new: ## Create new migration: make migrate-new name=add_posts
	goose -dir $(GOOSE_DIR) postgres "$(DATABASE_URL)" create $(name) sql

migrate-status: ## Show migration status
	goose -dir $(GOOSE_DIR) postgres "$(DATABASE_URL)" status

smoke: ## Run API smoke test (server must be running)
	bash scripts/smoke.sh

compose-up: ## Start api + postgres via docker
	docker-compose up -d --build

compose-down: ## Stop and remove containers
	docker-compose down

compose-logs: ## Tail api logs
	docker-compose logs -f api

tidy: ## Tidy go modules
	go mod tidy

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
