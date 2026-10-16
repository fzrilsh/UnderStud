# Understud — React Frontend

React 19 + TypeScript + Vite + TanStack Query + zustand + shadcn/Tailwind v4.
Auth via httpOnly cookie (dikirim backend) — tidak ada token di JS.

## Untuk Juri/Panitia (quickstart)

> **Cara termudah:** dari root `project/`, jalankan `docker-compose up -d` — FE sudah
> **ter-embed** di image backend, jadi cukup satu container. Buka **http://localhost:8080**.

FE **tidak** punya container sendiri: hasil build Vite di-embed ke binary Go
(`backend/internal/web/dist`, lihat `make build-web` di `backend/`). Bagian ini untuk
menjalankan FE saja saat development.

Prasyarat: **Node 20+ & pnpm**, dan **backend sudah jalan** (API di `localhost:8080`).

```bash
pnpm install
pnpm dev     # dev server + HMR di http://localhost:5173 (proxy /api → :8080)
```

Buka http://localhost:5173 → login dengan akun staf (`bk@demo.local` / `password123`).

## Dev Guide

### Struktur

```text
src/routes/                 router config + ProtectedRoute + RoleGuard
src/components/layout/      RootLayout (public), AppLayout (protected: sidebar+topbar), Sidebar, Topbar
src/components/ui/          shadcn components
src/pages/                  login, lapor, dashboard, dashboard-wk, students, student-detail, reports, not-found
src/features/               auth, dashboard, students, reports, lapor (hooks + komponen per-fitur)
src/stores/                 zustand (auth state only)
src/lib/api.ts              axios instance + 401 interceptor
src/lib/format.ts           format tanggal/waktu bahasa Indonesia
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
