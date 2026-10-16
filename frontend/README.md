# Understud — React Frontend

React 19 + TypeScript + Vite + TanStack Query + zustand + shadcn/Tailwind v4.
Auth via httpOnly cookie (dikirim backend) — tidak ada token di JS.

## Untuk Juri/Panitia (quickstart)

> **Cara termudah:** dari root `project/`, jalankan `docker-compose up -d` — itu menyalakan
> db + api + web sekaligus (lihat `project/README.md`). Bagian ini untuk menjalankan
> frontend saja.

Prasyarat: **Docker**, dan **backend sudah jalan** (`backend/`, `docker-compose up -d` — API di `localhost:8080`).

```bash
docker-compose up -d
```

Buka http://localhost:3000 → Register → login otomatis → Dashboard.

> Port 3000 bentrok dengan proses lain? Jalankan dengan port lain:
> `WEB_PORT=3100 docker-compose up -d` → buka http://localhost:3100.

Tanpa Docker (butuh Node 20+ & pnpm): `pnpm install && pnpm build && pnpm preview` — atau untuk dev: `pnpm dev` (proxy /api ke localhost:8080, buka http://localhost:5173).

## Dev Guide

### Struktur

```text
src/routes/                 router config + ProtectedRoute
src/components/layout/      RootLayout (public), AppLayout (protected: sidebar+navbar)
src/components/ui/          shadcn components
src/pages/                  login, register, dashboard, not-found
src/features/auth/          hooks (use-me/login/register/logout) + zod schemas
src/stores/                 zustand (auth state only)
src/lib/api.ts              axios instance + 401 interceptor
src/types/api.gen.ts        generated dari backend/api/openapi.yaml (committed)
```

### Perintah

```bash
pnpm dev         # dev server + HMR (http://localhost:5173, proxy /api → :8080)
pnpm build       # tsc + vite build
pnpm gen:api     # regen types dari ../backend/api/openapi.yaml (jalan di repo frontend)
pnpm dlx shadcn@latest add <component>   # tambah komponen shadcn
```

### Konvensi

- Semua request lewat `@/lib/api` (axios, `withCredentials`). Jangan fetch manual.
- Server-state → TanStack Query (`useQuery`/`useMutation`). UI/auth state → zustand.
- Tipe API dari `@/types/api.gen` — setelah backend nambah endpoint: update `api/openapi.yaml` di repo backend, lalu `pnpm gen:api` di sini.
- Form: react-hook-form + zod schema di `features/*/schemas.ts`.

### Cara Nambah Fitur (end-to-end)

Contoh: fitur `posts`.

1. Backend: ikuti "Cara Nambah Fitur" di README backend (migration → query → service → handler → route → openapi.yaml).
2. Di sini: `pnpm gen:api` (regen types).
3. Buat `src/features/posts/` — `schemas.ts` (zod) + `hooks.ts` (use-queries/use-mutations).
4. Buat `src/pages/posts.tsx` (atau beberapa page), compose dari `components/ui`.
5. Tambah route di `src/routes/index.tsx` (di dalam ProtectedRoute → AppLayout kalau perlu login).
6. Tambah menu di `src/components/layout/app-layout.tsx` (`navItems`).
