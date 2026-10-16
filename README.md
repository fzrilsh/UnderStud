# Understud — Monorepo

Monorepo berisi **backend** (Go) dan **frontend** (React). Frontend di-build Vite lalu
**di-embed ke binary Go** (`//go:embed`) — jadi satu image saja yang menyajikan UI + API.

## Arsitektur

```
project/
├── backend/     Go 1.26 + chi + pgx/sqlc + goose + Postgres 17 (+ FE ter-embed)
├── frontend/    React 19 + TS + Vite + TanStack Query + zustand + shadcn/Tailwind v4
└── docker-compose.yml   ← root: db + api (FE ter-embed) sekaligus
```

**Alur request** (semua same-origin — cookie auth jalan tanpa isu CORS):
- **Dev**: browser → Vite (`:5173`) → proxy `/api` → backend (`:8080`) → Postgres.
- **Docker**: browser → `api:8080` (serve aset FE + `/api/*`) → Postgres.

**Kontrak API**: `backend/api/openapi.yaml` (source of truth). FE generate types via
`pnpm gen:api` (di `frontend/`).

**Embed FE**: `make build-web` (di `backend/`) mem-build FE lalu menyalin `dist/` ke
`backend/internal/web/dist/` untuk di-embed. Dockerfile melakukannya otomatis
(build context = folder `project/`).

## Untuk Juri/Panitia (quickstart)

Prasyarat: **Docker saja**.

```bash
docker-compose up -d
```

Buka http://localhost:8080.

> Port 8080 bentrok? `API_PORT=8081 docker-compose up -d` → buka http://localhost:8081.

**Isi data demo** (sekali saja — DB baru belum punya akun):

```bash
docker-compose exec api /app/cli seed-demo
```

Akun demo: `bk@demo.local` (Guru BK) dan `wali@demo.local` (Wali Kelas XI-IPA-1),
password `password123` untuk keduanya.

Perintah lain:

```bash
docker-compose logs -f api     # tail log backend
docker-compose down            # stop semua (data Postgres + lampiran persist di volume)
```

## Dev Guide

Jalanin salah satu service sendiri (hot reload):

```bash
# Backend (butuh Postgres; paling gampang: `docker-compose up -d db` dulu)
cd backend && make dev          # air hot reload, API di :8080

# Frontend (proxy /api → :8080)
cd frontend && pnpm dev         # Vite di :5173
```

Detail lengkap (env, struktur folder, dan **cara nambah fitur end-to-end**) ada di
`backend/README.md` dan `frontend/README.md`.

## Struktur repo

- Repo **ini** (`project/`) = monorepo kode (backend + frontend).
- Folder parent workspace = repo terpisah untuk **docs/arsitektur & config** (bukan kode).
