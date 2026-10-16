# Backend Makefile — UnderStud
DATABASE_URL ?= postgres://postgres:postgres@127.0.0.1:5433/hackathon?sslmode=disable
GOOSE_DIR    ?= internal/db/migrations

.PHONY: dev run sqlc migrate-up migrate-down migrate-new migrate-status smoke compose-up compose-down compose-logs tidy help create-staff deactivate-staff check-location seed-demo build-web

dev: ## Run with hot reload (requires air: go install github.com/air-verse/air@latest)
	air

run: ## Run the API server
	go run ./cmd/api

build-web: ## Build FE (Vite) lalu salin ke internal/web/dist untuk di-embed
	rm -rf internal/web/dist
	mkdir -p internal/web/dist
	cd ../frontend && pnpm install --frozen-lockfile && pnpm build
	cp -R ../frontend/dist/. internal/web/dist/

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

create-staff: ## make create-staff email=... password=... name="..." role=BK kelas=XI-IPA-1
	go run ./cmd/cli create-staff --email "$(email)" --password "$(password)" --name "$(name)" --role "$(role)" --kelas "$(kelas)"

deactivate-staff: ## make deactivate-staff email=...
	go run ./cmd/cli deactivate-staff --email "$(email)"

check-location: ## make check-location lat=... lng=... radius=200
	go run ./cmd/cli check-location --lat $(lat) --lng $(lng) --radius $(radius)

seed-demo: ## Isi data demo
	go run ./cmd/cli seed-demo

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
