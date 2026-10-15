#!/usr/bin/env bash
# ============================================================================
# UnderStud: script deploy ke server (VPS).
#
# Yang dilakukan:
#   1. Cek prasyarat (docker, docker-compose, openssl)
#   2. Siapkan .env.prod dari template kalau belum ada, lalu isi JWT_SECRET &
#      POSTGRES_PASSWORD otomatis (kalau masih placeholder)
#   3. Tarik image terbaru dari GHCR
#   4. Jalankan stack
#   5. Tunggu sampai API sehat, lalu tampilkan ringkasan
#
# Pakai:
#   bash deploy/deploy.sh              # setup + jalan
#   bash deploy/deploy.sh --update     # hanya tarik image baru + restart
#   bash deploy/deploy.sh --seed       # isi data demo setelah jalan
#   bash deploy/deploy.sh --logs       # lihat log
#   bash deploy/deploy.sh --down       # hentikan (data tetap aman)
# ============================================================================
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$DIR"

COMPOSE_FILE="docker-compose.prod.yml"
ENV_FILE=".env.prod"
ENV_TEMPLATE="deploy/.env.prod.example"

# Warna (kalau terminal mendukung)
if [ -t 1 ]; then
  B=$'\033[1m'; G=$'\033[32m'; Y=$'\033[33m'; R=$'\033[31m'; N=$'\033[0m'
else
  B=""; G=""; Y=""; R=""; N=""
fi
step() { printf '%s==>%s %s\n' "$B" "$N" "$1"; }
ok()   { printf '    %s✓%s %s\n' "$G" "$N" "$1"; }
warn() { printf '    %s!%s %s\n' "$Y" "$N" "$1"; }
die()  { printf '%sERROR:%s %s\n' "$R" "$N" "$1" >&2; exit 1; }

# --- Mode operasi -----------------------------------------------------------
MODE="deploy"
case "${1:-}" in
  --update) MODE="update" ;;
  --seed)   MODE="seed" ;;
  --logs)   MODE="logs" ;;
  --down)   MODE="down" ;;
  --help|-h)
    sed -n '2,20p' "$0" | sed 's/^# \{0,1\}//'
    exit 0 ;;
  "") ;;
  *) die "Argumen tidak dikenal: $1 (pakai --help)" ;;
esac

# --- 1. Prasyarat -----------------------------------------------------------
step "Cek prasyarat"
command -v docker >/dev/null 2>&1 || die "docker tidak terpasang. Pasang dulu: https://docs.docker.com/engine/install/"
ok "docker $(docker --version | awk '{print $3}' | tr -d ,)"

if docker-compose version >/dev/null 2>&1; then
  DC="docker-compose"
elif docker compose version >/dev/null 2>&1; then
  DC="docker compose"
else
  die "docker-compose tidak terpasang (butuh binary docker-compose atau plugin 'docker compose')."
fi
ok "$($DC version | head -1)"

# --- Mode yang tidak butuh .env (log & down) --------------------------------
if [ "$MODE" = "logs" ]; then
  exec $DC -f "$COMPOSE_FILE" --env-file "$ENV_FILE" logs -f --tail=100
fi
if [ "$MODE" = "down" ]; then
  step "Hentikan stack (volume data tetap aman)"
  $DC -f "$COMPOSE_FILE" --env-file "$ENV_FILE" down
  ok "Stack dihentikan. Jalankan ulang: bash deploy/deploy.sh"
  exit 0
fi

# --- 2. Siapkan .env.prod ---------------------------------------------------
if [ "$MODE" != "seed" ]; then
  step "Siapkan konfigurasi ($ENV_FILE)"
  if [ ! -f "$ENV_FILE" ]; then
    [ -f "$ENV_TEMPLATE" ] || die "Template $ENV_TEMPLATE tidak ditemukan."
    cp "$ENV_TEMPLATE" "$ENV_FILE"
    ok "Dibuat dari template"
  else
    ok "Sudah ada, dipakai apa adanya"
  fi

  # Isi otomatis nilai yang masih placeholder.
  command -v openssl >/dev/null 2>&1 || warn "openssl tidak ada, isi JWT_SECRET & POSTGRES_PASSWORD manual"
  if command -v openssl >/dev/null 2>&1; then
    if grep -q '^JWT_SECRET=ganti-dengan-hasil-openssl-rand-hex-32' "$ENV_FILE"; then
      NEW="$(openssl rand -hex 32)"
      # `#` dipakai sebagai pemisah supaya aman untuk nilai heksadesimal.
      sed -i.bak "s|^JWT_SECRET=.*|JWT_SECRET=$NEW|" "$ENV_FILE" && rm -f "$ENV_FILE.bak"
      ok "JWT_SECRET dibuat otomatis (acak, 64 karakter)"
    fi
    if grep -q '^POSTGRES_PASSWORD=ganti-password-database-yang-kuat' "$ENV_FILE"; then
      NEW="$(openssl rand -hex 16)"
      sed -i.bak "s|^POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD=$NEW|" "$ENV_FILE" && rm -f "$ENV_FILE.bak"
      ok "POSTGRES_PASSWORD dibuat otomatis (acak)"
    fi
  fi

  # Validasi cepat: pastikan tidak ada placeholder yang tertinggal.
  if grep -qE '^(JWT_SECRET|POSTGRES_PASSWORD)=ganti-' "$ENV_FILE"; then
    die "Masih ada nilai placeholder di $ENV_FILE. Isi manual lalu jalankan lagi."
  fi
  JWT_LEN=$(grep '^JWT_SECRET=' "$ENV_FILE" | head -1 | cut -d= -f2- | wc -c | tr -d ' ')
  if [ "$JWT_LEN" -lt 33 ]; then
    die "JWT_SECRET terlalu pendek ($((JWT_LEN-1)) karakter). Minimal 32."
  fi
  ok "JWT_SECRET valid ($((JWT_LEN-1)) karakter)"

  CORS_LINE=$(grep '^CORS_ORIGINS=' "$ENV_FILE" | head -1 | cut -d= -f2-)
  ok "Domain diizinkan: $CORS_LINE"
fi

[ -f "$ENV_FILE" ] || die "$ENV_FILE tidak ada. Jalankan tanpa argumen dulu untuk setup."

# --- 3. Tarik image ---------------------------------------------------------
if [ "$MODE" != "seed" ]; then
  step "Tarik image terbaru dari GHCR"
  if ! $DC -f "$COMPOSE_FILE" --env-file "$ENV_FILE" pull; then
    die "Gagal menarik image. Pastikan paket GHCR publik, atau login dulu: docker login ghcr.io"
  fi
  ok "Image terbaru ditarik"
fi

# --- 4. Jalankan stack ------------------------------------------------------
if [ "$MODE" != "seed" ]; then
  step "Jalankan stack"
  $DC -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d
  ok "Container dijalankan"
fi

# --- 5. Tunggu API sehat ----------------------------------------------------
if [ "$MODE" != "seed" ]; then
  step "Tunggu API siap (migrasi database ikut jalan otomatis)"
  WEB_PORT_V="$(grep '^WEB_PORT=' "$ENV_FILE" | head -1 | cut -d= -f2-)"
  WEB_PORT_V="${WEB_PORT_V:-3000}"
  BIND_ADDR_V="$(grep '^BIND_ADDR=' "$ENV_FILE" | head -1 | cut -d= -f2-)"
  BIND_ADDR_V="${BIND_ADDR_V:-127.0.0.1}"
  URL="http://${BIND_ADDR_V}:${WEB_PORT_V}"
  # /healthz ada di root (bukan di bawah /api/v1); lewat nginx web pun diteruskan.
  for i in $(seq 1 60); do
    if curl -fsS -o /dev/null "${URL}/healthz" 2>/dev/null; then
      ok "API sehat setelah ${i} detik"
      break
    fi
    if [ "$i" -eq 60 ]; then
      warn "API belum sehat setelah 60 detik. Cek log:"
      $DC -f "$COMPOSE_FILE" --env-file "$ENV_FILE" logs --tail=40 api || true
      die "Deployment gagal: API tidak sehat."
    fi
    sleep 1
  done
  ok "Aplikasi siap di ${URL}"
fi

# --- 6. Seed data demo (opsional) -------------------------------------------
if [ "$MODE" = "seed" ]; then
  step "Isi data demo"
  $DC -f "$COMPOSE_FILE" --env-file "$ENV_FILE" exec -T api /app/cli seed-demo
  ok "Data demo terisi (bk@demo.local / wali@demo.local, password: password123)"
fi

# --- Ringkasan --------------------------------------------------------------
if [ "$MODE" != "seed" ]; then
  printf '\n%sDeployment selesai.%s\n\n' "$B$G" "$N"
  printf 'Langkah berikutnya:\n'
  printf '  1. Arahkan domain ke server ini (Cloudflare Tunnel / A record + proxy).\n'
  printf '  2. Pastikan proxy meneruskan ke: %s\n' "$URL"
  printf '  3. Isi data demo (opsional):  bash deploy/deploy.sh --seed\n'
  printf '  4. Bikin akun staf asli:      %s -f %s --env-file %s exec api /app/cli create-staff --help\n' "$DC" "$COMPOSE_FILE" "$ENV_FILE"
  printf '  5. Lihat log:                 bash deploy/deploy.sh --logs\n\n'
fi
