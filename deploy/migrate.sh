#!/usr/bin/env bash
# ============================================================================
# UnderStud: jalankan migrasi database di server.
#
# Dipakai karena `MIGRATE_ON_START=false` (app tidak migrasi sendiri saat start,
# supaya aman di balik pgbouncer). Migrasi dijalankan terpisah, LANGSUNG ke
# Postgres (bukan pgbouncer), karena goose butuh sesi yang stabil.
#
# Pakai:
#   bash deploy/migrate.sh                 # baca .env di ~/services/understud
#   ENV_FILE=/path/.env bash deploy/migrate.sh
#   DIRECT_DATABASE_URL='postgres://...' bash deploy/migrate.sh   # override
#
# Cara kerja: menjalankan container `understud-api` sekali pakai (bukan yang
# sedang jalan) dengan perintah `cli migrate`, di network yang sama dengan
# database, lalu langsung dihapus.
# ============================================================================
set -euo pipefail

ENV_FILE="${ENV_FILE:-$HOME/services/understud/.env}"
IMAGE="${IMAGE:-ghcr.io/fzrilsh/understud-api:latest}"
NETWORK="${NETWORK:-internal}"

if [ -t 1 ]; then B=$'\033[1m'; G=$'\033[32m'; Y=$'\033[33m'; N=$'\033[0m'; else B=""; G=""; Y=""; N=""; fi
step() { printf '%s==>%s %s\n' "$B" "$N" "$1"; }
ok()   { printf '    %s✓%s %s\n' "$G" "$N" "$1"; }
warn() { printf '    %s!%s %s\n' "$Y" "$N" "$1"; }

[ -f "$ENV_FILE" ] || { echo "ERROR: $ENV_FILE tidak ditemukan. Isi dari deploy/.env.prod.example dulu." >&2; exit 1; }

# Ambil DATABASE_URL dari .env
DB_URL="$(grep -E '^DATABASE_URL=' "$ENV_FILE" | head -1 | cut -d= -f2-)"
[ -n "$DB_URL" ] || { echo "ERROR: DATABASE_URL tidak ada di $ENV_FILE" >&2; exit 1; }

# Migrasi harus langsung ke Postgres, bukan pgbouncer. Kalau DATABASE_URL masih
# menunjuk pgbouncer, pakai DIRECT_DATABASE_URL (kalau diisi).
DIRECT_URL="${DIRECT_DATABASE_URL:-}"
if [ -n "$DIRECT_URL" ]; then
  DB_URL="$DIRECT_URL"
  ok "memakai DIRECT_DATABASE_URL (langsung ke Postgres)"
elif printf '%s' "$DB_URL" | grep -qi pgbouncer; then
  warn "DATABASE_URL menunjuk pgbouncer. Migrasi sebaiknya langsung ke Postgres."
  warn "Set DIRECT_DATABASE_URL (mis. postgres://user:pass@postgres:5432/understud) lalu ulangi."
  echo "ERROR: butuh DIRECT_DATABASE_URL untuk migrasi." >&2
  exit 1
fi

# Image dipublikasikan untuk linux/amd64. Kalau script dijalankan di mesin
# dengan arsitektur lain (mis. laptop ARM untuk uji coba), paksa platform
# linux/amd64. Di server x86_64 flag ini tidak dipakai.
PLATFORM_FLAG=""
if [ "$(uname -m)" != "x86_64" ]; then
  PLATFORM_FLAG="--platform linux/amd64"
  warn "arsitektur $(uname -m) bukan x86_64; memakai emulasi linux/amd64"
fi

step "Siapkan image migrasi"
if docker image inspect "$IMAGE" >/dev/null 2>&1; then
  # Image sudah ada di mesin ini (mis. saat uji coba lokal sebelum di-push).
  # Jangan pakai --platform: image lokal bisa beda arsitektur dan justru bikin
  # docker mencoba pull dari registry lalu gagal.
  ok "memakai image lokal yang sudah ada"
  PLATFORM_FLAG=""
else
  # shellcheck disable=SC2086
  docker pull $PLATFORM_FLAG "$IMAGE" >/dev/null
  ok "image ditarik dari registry"
fi

step "Jalankan migrasi"
# shellcheck disable=SC2086
docker run --rm $PLATFORM_FLAG \
  --network "$NETWORK" \
  -e DATABASE_URL="$DB_URL" \
  -e MIGRATE_ON_START=false \
  "$IMAGE" /app/cli migrate

ok "Migrasi selesai."
