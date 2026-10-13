#!/usr/bin/env bash
# Smoke T2 — Data Siswa: list scoped, CRUD, observasi, red flag, import, scoping.
set -euo pipefail
BASE_URL="${1:-http://localhost:8080}"
JAR_WALI="$(mktemp)"; JAR_BK="$(mktemp)"; JAR_OTHER="$(mktemp)"
CSV="$(mktemp)"
trap 'rm -f "$JAR_WALI" "$JAR_BK" "$JAR_OTHER" "$CSV"' EXIT

pass() { printf '  \033[32m✓\033[0m %s\n' "$1"; }
fail() { printf '  \033[31m✗ %s\033[0m\n' "$1"; exit 1; }

login() { # $1 jar, $2 email
  curl -sf -c "$1" -X POST "$BASE_URL/api/v1/auth/login" -H 'Content-Type: application/json' \
    -d "{\"email\":\"$2\",\"password\":\"password123\"}" > /dev/null
}

echo "==> [1/10] login Wali Kelas + BK"
login "$JAR_WALI" "wali@demo.local"
login "$JAR_BK" "bk@demo.local"
pass "login ok"

echo "==> [2/10] daftar siswa (WK, kelasnya)"
COUNT=$(curl -sf -b "$JAR_WALI" "$BASE_URL/api/v1/students" | jq '.data | length')
[ "$COUNT" -ge 4 ] || fail "harap seed demo dulu (got $COUNT)"
pass "WK lihat $COUNT siswa"

echo "==> [3/10] katalog checklist = 28 item"
N=$(curl -sf -b "$JAR_WALI" "$BASE_URL/api/v1/checklist" | jq '.data | length')
[ "$N" = "28" ] || fail "checklist expected 28, got $N"
pass "checklist 28 item"

echo "==> [4/10] tambah siswa manual (WK)"
NIS="99$RANDOM"
NEW_ID=$(curl -sf -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students" -H 'Content-Type: application/json' \
  -d "{\"nis\":\"$NIS\",\"name\":\"Siswa Uji\",\"kelas\":\"XI-IPA-1\"}" | jq -r '.data.id')
[ -n "$NEW_ID" ] && [ "$NEW_ID" != "null" ] || fail "gagal tambah siswa"
pass "siswa dibuat: $NEW_ID"

echo "==> [5/10] BK read-only: coba tambah -> 403"
STATUS=$(curl -s -o /dev/null -w '%{http_code}' -b "$JAR_BK" -X POST "$BASE_URL/api/v1/students" \
  -H 'Content-Type: application/json' -d "{\"nis\":\"1$RANDOM\",\"name\":\"X\",\"kelas\":\"XI-IPA-1\"}")
[ "$STATUS" = "403" ] || fail "BK create expected 403, got $STATUS"
pass "BK ditolak (read-only)"

echo "==> [6/10] simpan profil (WK) + detail memuatnya"
curl -sf -b "$JAR_WALI" -X PUT "$BASE_URL/api/v1/students/$NEW_ID/profile" -H 'Content-Type: application/json' \
  -d '{"nilai_erapor":[{"mapel":"Matematika","nilai":88,"periode":"2026-Ganjil"}],"minat_bakat":"Melukis"}' > /dev/null
curl -sf -b "$JAR_WALI" "$BASE_URL/api/v1/students/$NEW_ID" | jq -e '.data.profile.minat_bakat == "Melukis"' > /dev/null
pass "profil tersimpan"

echo "==> [7/10] observasi biasa (WK) — bukan red flag"
curl -sf -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/$NEW_ID/observations" -H 'Content-Type: application/json' \
  -d "{\"items\":[\"emosi:murung\"],\"tanggal\":\"$(date +%Y-%m-%d)\"}" \
  | jq -e '.data.red_flag == false' > /dev/null
pass "observasi biasa ok"

echo "==> [8/10] observasi red flag (WK) -> tampil di /red-flags (BK); edit simpan revisi"
OBS_ID=$(curl -sf -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/$NEW_ID/observations" -H 'Content-Type: application/json' \
  -d "{\"items\":[\"red_flag:menyakiti_diri\"],\"tanggal\":\"$(date +%Y-%m-%d)\"}" | jq -r '.data.id')
curl -sf -b "$JAR_BK" "$BASE_URL/api/v1/red-flags" | jq -e --arg id "$NEW_ID" '.data | map(.student_id) | index($id)' > /dev/null
curl -sf -b "$JAR_WALI" -X PUT "$BASE_URL/api/v1/observations/$OBS_ID" -H 'Content-Type: application/json' \
  -d "{\"items\":[\"emosi:murung\"],\"tanggal\":\"$(date +%Y-%m-%d)\"}" > /dev/null
curl -sf -b "$JAR_WALI" "$BASE_URL/api/v1/observations/$OBS_ID/revisions" | jq -e '.data | length >= 1' > /dev/null
pass "red flag tampil + revisi tersimpan"

echo "==> [9/10] import CSV: preview (baru) + commit"
cat > "$CSV" <<EOF
nis,nama,kelas,mapel,nilai,periode
$NIS,Siswa Uji,XI-IPA-1,Biologi,80,2026-Ganjil
77$RANDOM,Siswa Impor,XI-IPA-1,Kimia,75,2026-Ganjil
EOF
TOKEN=$(curl -sf -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/import/preview" -F "file=@$CSV" | jq -r '.data.preview_token')
[ -n "$TOKEN" ] && [ "$TOKEN" != "null" ] || fail "preview gagal"
RES=$(curl -sf -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/import/commit" -H 'Content-Type: application/json' -d "{\"preview_token\":\"$TOKEN\"}")
echo "$RES" | jq -e '.data.created >= 1' > /dev/null || fail "commit tidak membuat siswa baru"
pass "import preview+commit ok"

echo "==> [10/10] import scoping: kelas luar cakupan -> error di preview"
cat > "$CSV" <<EOF
nis,nama,kelas,mapel,nilai,periode
88$RANDOM,Siswa Luar,XI-IPA-9,Kimia,75,2026-Ganjil
EOF
ERR=$(curl -sf -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/import/preview" -F "file=@$CSV" | jq '.data.errors | length')
[ "$ERR" -ge 1 ] || fail "kelas luar cakupan harus error"
pass "kelas luar cakupan ditolak"

echo
printf '\033[32mALL T2 SMOKE TESTS PASSED\033[0m\n'
