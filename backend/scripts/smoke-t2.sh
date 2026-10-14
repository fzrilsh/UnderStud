#!/usr/bin/env bash
# Smoke T2 — Data Siswa: list scoped, CRUD, observasi, red flag, scoping.
# (Import leger dipindah ke smoke-t4.sh.)
set -euo pipefail
BASE_URL="${1:-http://localhost:8080}"
JAR_WALI="$(mktemp)"; JAR_BK="$(mktemp)"
trap 'rm -f "$JAR_WALI" "$JAR_BK"' EXIT

pass() { printf '  \033[32m✓\033[0m %s\n' "$1"; }
fail() { printf '  \033[31m✗ %s\033[0m\n' "$1"; exit 1; }

login() { # $1 jar, $2 email
  curl -sf -c "$1" -X POST "$BASE_URL/api/v1/auth/login" -H 'Content-Type: application/json' \
    -d "{\"email\":\"$2\",\"password\":\"password123\"}" > /dev/null
}

echo "==> [1/8] login Wali Kelas + BK"
login "$JAR_WALI" "wali@demo.local"
login "$JAR_BK" "bk@demo.local"
pass "login ok"

echo "==> [2/8] daftar siswa (WK, kelasnya)"
COUNT=$(curl -sf -b "$JAR_WALI" "$BASE_URL/api/v1/students" | jq '.data | length')
[ "$COUNT" -ge 4 ] || fail "harap seed demo dulu (got $COUNT)"
pass "WK lihat $COUNT siswa"

echo "==> [3/8] katalog checklist = 28 item"
N=$(curl -sf -b "$JAR_WALI" "$BASE_URL/api/v1/checklist" | jq '.data | length')
[ "$N" = "28" ] || fail "checklist expected 28, got $N"
pass "checklist 28 item"

echo "==> [4/8] tambah siswa manual (WK)"
NISN="99$RANDOM$RANDOM"
NEW_ID=$(curl -sf -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students" -H 'Content-Type: application/json' \
  -d "{\"nisn\":\"$NISN\",\"name\":\"Siswa Uji\",\"kelas\":\"X TJKT 1\"}" | jq -r '.data.id')
[ -n "$NEW_ID" ] && [ "$NEW_ID" != "null" ] || fail "gagal tambah siswa"
pass "siswa dibuat: $NEW_ID"

echo "==> [5/8] BK read-only: coba tambah -> 403"
STATUS=$(curl -s -o /dev/null -w '%{http_code}' -b "$JAR_BK" -X POST "$BASE_URL/api/v1/students" \
  -H 'Content-Type: application/json' -d "{\"nisn\":\"1$RANDOM$RANDOM\",\"name\":\"X\",\"kelas\":\"X TJKT 1\"}")
[ "$STATUS" = "403" ] || fail "BK create expected 403, got $STATUS"
pass "BK ditolak (read-only)"

echo "==> [6/8] simpan profil + absensi (WK) + detail memuatnya"
curl -sf -b "$JAR_WALI" -X PUT "$BASE_URL/api/v1/students/$NEW_ID/profile" -H 'Content-Type: application/json' \
  -d '{"nilai_erapor":[{"mapel":"Matematika (Umum)","nilai":88,"periode":"2025/2026 Genap"}],"minat_bakat":"Melukis","sakit":2,"izin":1,"alpa":0}' > /dev/null
curl -sf -b "$JAR_WALI" "$BASE_URL/api/v1/students/$NEW_ID" \
  | jq -e '.data.profile.minat_bakat == "Melukis" and .data.profile.sakit == 2' > /dev/null
pass "profil + absensi tersimpan"

echo "==> [7/8] observasi biasa + red flag (WK) + revisi"
curl -sf -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/$NEW_ID/observations" -H 'Content-Type: application/json' \
  -d "{\"items\":[\"emosi:murung\"],\"tanggal\":\"$(date +%Y-%m-%d)\"}" \
  | jq -e '.data.red_flag == false' > /dev/null
# red flag untuk diuji edit+revisi (bukan untuk di-handle — edit mengubahnya jadi non-red-flag)
EDIT_ID=$(curl -sf -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/$NEW_ID/observations" -H 'Content-Type: application/json' \
  -d "{\"items\":[\"red_flag:menyakiti_diri\"],\"tanggal\":\"$(date +%Y-%m-%d)\"}" | jq -r '.data.id')
curl -sf -b "$JAR_WALI" -X PUT "$BASE_URL/api/v1/observations/$EDIT_ID" -H 'Content-Type: application/json' \
  -d "{\"items\":[\"emosi:murung\"],\"tanggal\":\"$(date +%Y-%m-%d)\"}" > /dev/null
curl -sf -b "$JAR_WALI" "$BASE_URL/api/v1/observations/$EDIT_ID/revisions" | jq -e '.data | length >= 1' > /dev/null
pass "observasi + red flag + revisi ok"

echo "==> [8/8] BK tandai red flag ditangani -> hilang dari panel"
OBS_ID=$(curl -sf -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/$NEW_ID/observations" -H 'Content-Type: application/json' \
  -d "{\"items\":[\"red_flag:menyakiti_diri\"],\"tanggal\":\"$(date +%Y-%m-%d)\"}" | jq -r '.data.id')
curl -sf -b "$JAR_BK" "$BASE_URL/api/v1/red-flags" | jq -e --arg id "$NEW_ID" '.data | map(.student_id) | index($id)' > /dev/null
curl -sf -b "$JAR_BK" -X POST "$BASE_URL/api/v1/observations/$OBS_ID/handle" > /dev/null
curl -sf -b "$JAR_BK" "$BASE_URL/api/v1/red-flags" | jq -e --arg id "$NEW_ID" '.data | map(.student_id) | index($id) == null' > /dev/null
pass "red flag tampil + ditandai ditangani"

echo
printf '\033[32mALL T2 SMOKE TESTS PASSED\033[0m\n'
