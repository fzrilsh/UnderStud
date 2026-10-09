# Hackathon Starter — Go Backend

Go 1.26 + chi + pgx/sqlc + goose + Postgres 17.
Auth: long-lived JWT via httpOnly cookie (no refresh token). Logout = clear cookie.

## Untuk Juri/Panitia (quickstart)

Prasyarat: **Docker saja**.

```bash
docker-compose up -d
```

- API: http://localhost:8080 — health di `GET /healthz`
- Postgres di `localhost:5432` (db `hackathon`, user/pass `postgres`/`postgres`)
- Migrations **otomatis** dijalankan saat service start.

### Endpoint Auth

| Method | Path | Body | Response |
|---|---|---|---|
| POST | `/api/v1/auth/register` | `{email, password, name}` | `201` + cookie + user |
| POST | `/api/v1/auth/login` | `{email, password}` | `200` + cookie + user |
| POST | `/api/v1/auth/logout` | — | `200`, cookie cleared |
| GET | `/api/v1/auth/me` | — (protected) | `200` + user |

Contoh curl:

```bash
# register (simpan cookie)
curl -c cookies.txt -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"juri@example.com","password":"password123","name":"Juri"}'

# me (pakai cookie)
curl -b cookies.txt http://localhost:8080/api/v1/auth/me

# bearer fallback untuk Postman: header Authorization: Bearer <token> juga diterima
```

## Dev Guide

### Struktur

```text
cmd/api/                    entry point
internal/config/            env config (all defaults)
internal/server/            router + middleware chain
internal/handler/           HTTP layer
internal/service/           business logic
internal/db/                sqlc generated + pool + migrate runner
internal/db/migrations/     goose SQL migrations
internal/middleware/        RequireAuth, request logger
internal/tokens/            JWT manager
internal/utils/             app errors, password, respond helpers
db/queries/                 sqlc query definitions (source)
api/openapi.yaml            OpenAPI 3.1 spec (source of truth untuk FE types)
scripts/smoke.sh            smoke test end-to-end
```

### Perintah

```bash
make dev            # hot reload (butuh: go install github.com/air-verse/air@latest)
make run            # run tanpa reload
make sqlc           # regenerate db layer setelah edit db/queries/*.sql
make migrate-new name=add_posts   # migration baru
make migrate-up     # apply migrations manual (startup sudah otomatis)
make smoke          # smoke test (server harus jalan)
make compose-up     # docker: api + postgres
```

### Environment Variables (semua punya default)

| Env | Default | Keterangan |
|---|---|---|
| PORT | 8080 | port API |
| DATABASE_URL | postgres://postgres:postgres@localhost:5432/hackathon?sslmode=disable | koneksi DB |
| JWT_SECRET | dev-secret-change-me-in-prod | HS256 secret |
| AUTH_TOKEN_TTL | 87600h | TTL token & cookie (10 tahun) |
| CORS_ORIGINS | http://localhost:5173,http://localhost:3000 | origins yang diizinkan (credentials) |
| DB_MAX_CONNS | 10 | pool max |
| DB_MIN_CONNS | 2 | pool min |
| LOG_LEVEL | info | debug/info/warn/error |

### Cara Nambah Fitur (end-to-end)

Contoh: resource `posts`.

1. Migration: `make migrate-new name=add_posts` → edit file SQL baru di `internal/db/migrations/`
2. Query: tulis `-- name: CreatePost :one` dsb. di `db/queries/posts.sql`
3. Generate: `make sqlc`
4. Service: `internal/service/post.go` — business logic + validasi
5. Handler: `internal/handler/post.go` — Bind + OK/Created/Error
6. Route: mount di `internal/server/server.go` (tambah `r.Route("/posts", ...)`)
7. Dokumentasi: update `api/openapi.yaml`
8. FE types: di repo frontend, `pnpm gen:api`
9. Verifikasi: `make run` + `bash scripts/smoke.sh` (tambah step baru di smoke bila relevan)
