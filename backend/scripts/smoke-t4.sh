#!/usr/bin/env bash
# Smoke T4 — impor leger E-Rapor (.xlsx/.csv), NISN, absensi, template.
set -euo pipefail
BASE_URL="${1:-http://localhost:8080}"
JAR_WALI="$(mktemp)"; JAR_BK="$(mktemp)"; CSV="$(mktemp)"
trap 'rm -f "$JAR_WALI" "$JAR_BK" "$CSV"' EXIT

pass() { printf '  \033[32m✓\033[0m %s\n' "$1"; }
fail() { printf '  \033[31m✗ %s\033[0m\n' "$1"; exit 1; }
login() { curl -sf -c "$1" -X POST "$BASE_URL/api/v1/auth/login" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$2\",\"password\":\"password123\"}" > /dev/null; }

echo "==> [1/6] login WK + BK"
login "$JAR_WALI" "wali@demo.local"
login "$JAR_BK" "bk@demo.local"
pass "login ok"

echo "==> [2/6] unduh template xlsx (WK)"
CT=$(curl -s -D - -o /tmp/tpl.xlsx -b "$JAR_WALI" "$BASE_URL/api/v1/students/import/template?format=xlsx" | grep -i '^content-type')
echo "$CT" | grep -qi 'spreadsheetml' || fail "template xlsx content-type salah: $CT"
head -c 2 /tmp/tpl.xlsx | grep -q 'PK' || fail "template bukan xlsx"
pass "template xlsx ok"

echo "==> [3/6] template ditolak untuk BK (403)"
STATUS=$(curl -s -o /dev/null -w '%{http_code}' -b "$JAR_BK" "$BASE_URL/api/v1/students/import/template")
[ "$STATUS" = "403" ] || fail "BK template expected 403, got $STATUS"
pass "BK ditolak"

echo "==> [4/6] impor leger CSV (kelas benar) → preview baru"
NISN="77$RANDOM$RANDOM"
cat > "$CSV" <<EOF
Sekolah,,SMKN 1 TANGERANG
NPSN,,20606915
Kelas,,X TJKT 1
Tahun Pelajaran,,2025/2026 Genap

No,NAMA PESERTA DIDIK,NISN,Matematika (Umum),Bahasa Indonesia,S,I,A
1,Siswa Leger,$NISN,85,90,3,1,0
EOF
PREV=$(curl -sf -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/import/preview" -F "file=@$CSV;filename=leger.csv")
echo "$PREV" | jq -e '.data.ringkasan.baru >= 1' > /dev/null || fail "preview tidak mendeteksi siswa baru"
TOKEN=$(echo "$PREV" | jq -r '.data.preview_token')
pass "preview baru ok"

echo "==> [5/6] commit → siswa masuk + absensi tersimpan"
curl -sf -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/import/commit" -H 'Content-Type: application/json' \
  -d "{\"preview_token\":\"$TOKEN\"}" | jq -e '.data.created >= 1' > /dev/null
SID=$(curl -sf -b "$JAR_WALI" "$BASE_URL/api/v1/students?q=Siswa%20Leger" | jq -r '.data[0].id')
curl -sf -b "$JAR_WALI" "$BASE_URL/api/v1/students/$SID" | jq -e '.data.profile.sakit == 3 and .data.profile.izin == 1' > /dev/null \
  || fail "absensi tidak tersimpan"
pass "commit + absensi ok"

echo "==> [6/6] kelas beda di metadata → tolak file"
cat > "$CSV" <<EOF
Kelas,,X TJKT 9
No,NAMA PESERTA DIDIK,NISN,Matematika (Umum)
1,Anak Salah Kelas,9999999999,80
EOF
STATUS=$(curl -s -o /dev/null -w '%{http_code}' -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/import/preview" -F "file=@$CSV;filename=leger.csv")
[ "$STATUS" = "400" ] || fail "kelas mismatch expected 400, got $STATUS"
pass "kelas mismatch ditolak"

echo
printf '\033[32mALL T4 SMOKE TESTS PASSED\033[0m\n'
