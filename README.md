# UnderStud — Monorepo

Monorepo berisi dua service: **backend** (Go) dan **frontend** (React).
Docker terpisah per service, plus compose root untuk menjalankan semuanya sekaligus.

> Skeleton ini berisi vertical slice auth yang sudah jalan end-to-end
> (register → login → `/me` → logout → protected dashboard) sebagai fondasi & contoh pola.

## Arsitektur

```
project/
├── backend/     Go 1.26 + chi + pgx/sqlc + goose + Postgres 17
├── frontend/    React 19 + TS + Vite + TanStack Query + zustand + shadcn/Tailwind v4
└── docker-compose.yml   ← root: db + api + web sekaligus
```

**Alur request** (semua same-origin via proxy — cookie auth jalan tanpa isu CORS):
- **Dev**: browser → Vite (`:5173`) → proxy `/api` → backend (`:8080`) → Postgres.
- **Docker**: browser → nginx (`:3000`) → proxy `/api` → `api:8080` (service name) → Postgres.

**Kontrak API**: `backend/api/openapi.yaml` (source of truth). FE generate types via
`pnpm gen:api` (di `frontend/`).

## Untuk Juri/Panitia (quickstart)

Prasyarat: **Docker saja**.

```bash
docker-compose up -d
```

Buka http://localhost:3000.

> Port 3000 bentrok? `WEB_PORT=3100 docker-compose up -d` → buka http://localhost:3100.

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
