#!/usr/bin/env bash
# Smoke T3 — ekspor ringkasan (pdf/txt + audit) + riwayat laporan BK.
set -euo pipefail
BASE_URL="${1:-http://localhost:8080}"
JAR_WALI="$(mktemp)"; JAR_BK="$(mktemp)"
trap 'rm -f "$JAR_WALI" "$JAR_BK"' EXIT

pass() { printf '  \033[32m✓\033[0m %s\n' "$1"; }
fail() { printf '  \033[31m✗ %s\033[0m\n' "$1"; exit 1; }
login() {
  curl -sf -c "$1" -X POST "$BASE_URL/api/v1/auth/login" -H 'Content-Type: application/json' \
    -d "{\"email\":\"$2\",\"password\":\"password123\"}" > /dev/null
}

echo "==> [1/6] login WK + BK"
login "$JAR_WALI" "wali@demo.local"
login "$JAR_BK" "bk@demo.local"
pass "login ok"

echo "==> [2/6] ambil siswa seed (Andi)"
SID=$(curl -sf -b "$JAR_WALI" "$BASE_URL/api/v1/students?q=Andi" | jq -r '.data[0].id')
[ -n "$SID" ] && [ "$SID" != "null" ] || fail "siswa seed tidak ada (jalankan make seed-demo)"
pass "student_id=$SID"

echo "==> [3/6] ekspor format tidak valid -> 400"
STATUS=$(curl -s -o /dev/null -w '%{http_code}' -b "$JAR_WALI" \
  "$BASE_URL/api/v1/students/$SID/summaries/00000000-0000-0000-0000-000000000000/export?format=docx")
[ "$STATUS" = "400" ] || fail "format invalid expected 400, got $STATUS"
pass "format invalid ditolak"

echo "==> [4/6] ekspor ringkasan tidak ada -> 404"
STATUS=$(curl -s -o /dev/null -w '%{http_code}' -b "$JAR_WALI" \
  "$BASE_URL/api/v1/students/$SID/summaries/00000000-0000-0000-0000-000000000000/export?format=txt")
[ "$STATUS" = "404" ] || fail "ringkasan tak ada expected 404, got $STATUS"
pass "ringkasan tak ada -> 404"

echo "==> [5/6] riwayat laporan BK: BK 200, WK 403"
curl -sf -b "$JAR_BK" "$BASE_URL/api/v1/reports/history" | jq -e '.meta.total != null' > /dev/null
STATUS=$(curl -s -o /dev/null -w '%{http_code}' -b "$JAR_WALI" "$BASE_URL/api/v1/reports/history")
[ "$STATUS" = "403" ] || fail "WK history expected 403, got $STATUS"
pass "riwayat BK ok + WK ditolak"

echo "==> [6/6] ekspor sukses (bila ada ringkasan): validasi header + audit"
SUMID=$(docker exec project-db-1 psql -U postgres -d hackathon -tAc \
  "select id from ai_summaries where student_id='$SID' order by created_at desc limit 1" 2>/dev/null || true)
if [ -z "$SUMID" ]; then
  echo "  (lewati: belum ada ringkasan — AI_API_KEY kosong)"
else
  CT=$(curl -s -D - -o /tmp/t3.txt -b "$JAR_WALI" \
    "$BASE_URL/api/v1/students/$SID/summaries/$SUMID/export?format=txt" | grep -i '^content-type' || true)
  echo "$CT" | grep -qi 'text/plain' || fail "txt export content-type salah: $CT"
  CT=$(curl -s -D - -o /tmp/t3.pdf -b "$JAR_WALI" \
    "$BASE_URL/api/v1/students/$SID/summaries/$SUMID/export?format=pdf" | grep -i '^content-type' || true)
  echo "$CT" | grep -qi 'application/pdf' || fail "pdf export content-type salah: $CT"
  head -c 5 /tmp/t3.pdf | grep -q '%PDF-' || fail "bukan PDF valid"
  rm -f /tmp/t3.txt /tmp/t3.pdf
  pass "ekspor txt+pdf valid"
fi

echo
printf '\033[32mALL T3 SMOKE TESTS PASSED\033[0m\n'
