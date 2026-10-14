#!/usr/bin/env bash
# Smoke T1 — alur chat anonim → antrean BK → pin → balas → reply → requeue.
set -euo pipefail
BASE_URL="${1:-http://localhost:8080}"
JAR_ANON="$(mktemp)"; JAR_BK="$(mktemp)"
trap 'rm -f "$JAR_ANON" "$JAR_BK"' EXIT

pass() { printf '  \033[32m✓\033[0m %s\n' "$1"; }
fail() { printf '  \033[31m✗ %s\033[0m\n' "$1"; exit 1; }

echo "==> [1/10] bootstrap sesi anonim"
curl -sf -c "$JAR_ANON" "$BASE_URL/api/v1/lapor" | jq -e '.data.profile' > /dev/null
pass "sesi anonim dibuat"

echo "==> [2/10] siswa kirim laporan (anonim) + lampiran"
printf '\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89\x00\x00\x00\nIDATx\x9cc\x00\x01\x00\x00\x05\x00\x01\r\n-\xb4\x00\x00\x00\x00IEND\xaeB`\x82' > "$JAR_ANON.png"
CHAT_ID=$(curl -sf -b "$JAR_ANON" -X POST "$BASE_URL/api/v1/lapor/chats" \
  -F 'body=Laporan smoke T1' -F 'identity_mode=anonim' \
  -F "files=@$JAR_ANON.png;type=image/png;filename=bukti.png" | jq -r '.data.id')
rm -f "$JAR_ANON.png"
[ -n "$CHAT_ID" ] && [ "$CHAT_ID" != "null" ] || fail "chat id kosong"
pass "chat dibuat: $CHAT_ID"

echo "==> [3/10] lampiran muncul di detail siswa"
ATT_ID=$(curl -sf -b "$JAR_ANON" "$BASE_URL/api/v1/lapor/chats/$CHAT_ID" \
  | jq -r '.data.messages[0].attachments[0].id')
[ -n "$ATT_ID" ] && [ "$ATT_ID" != "null" ] || fail "lampiran tidak terdaftar"
pass "lampiran terdaftar: $ATT_ID"

echo "==> [4/10] login BK"
curl -sf -c "$JAR_BK" -X POST "$BASE_URL/api/v1/auth/login" -H 'Content-Type: application/json' \
  -d '{"email":"bk@demo.local","password":"password123"}' > /dev/null
pass "login BK ok"

echo "==> [5/10] antrean BK memuat laporan"
curl -sf -b "$JAR_BK" "$BASE_URL/api/v1/reports" | jq -e --arg id "$CHAT_ID" '.data | map(.id) | index($id)' > /dev/null
pass "laporan ada di antrean"

echo "==> [6/10] serve lampiran: staf 200, sesi lain 404"
S=$(curl -s -o /dev/null -w '%{http_code}' -b "$JAR_BK" "$BASE_URL/api/v1/attachments/$ATT_ID")
[ "$S" = "200" ] || fail "staf serve expected 200, got $S"
JAR_OTHER="$(mktemp)"
curl -s -c "$JAR_OTHER" "$BASE_URL/api/v1/lapor" > /dev/null
S=$(curl -s -o /dev/null -w '%{http_code}' -b "$JAR_OTHER" "$BASE_URL/api/v1/lapor/attachments/$ATT_ID")
rm -f "$JAR_OTHER"
[ "$S" = "404" ] || fail "sesi lain expected 404, got $S"
pass "otorisasi lampiran benar"

echo "==> [7/10] Next sebelum balas -> 400"
STATUS=$(curl -s -o /dev/null -w '%{http_code}' -b "$JAR_BK" -X POST "$BASE_URL/api/v1/reports/$CHAT_ID/next")
[ "$STATUS" = "400" ] || fail "expected 400, got $STATUS"
pass "Next ditolak sebelum balas"

echo "==> [8/10] BK pin + balas"
curl -sf -b "$JAR_BK" -X POST "$BASE_URL/api/v1/reports/$CHAT_ID/pin" > /dev/null
curl -sf -b "$JAR_BK" -X POST "$BASE_URL/api/v1/reports/$CHAT_ID/messages" \
  -H 'Content-Type: application/json' -d '{"body":"Terima kasih, kami tindak lanjuti."}' \
  | jq -e '.data.sender == "BK"' > /dev/null
pass "pin + balas ok"

echo "==> [9/10] Next setelah balas -> 200; keluar antrean; siswa reply -> requeue 'baru'"
STATUS=$(curl -s -o /dev/null -w '%{http_code}' -b "$JAR_BK" -X POST "$BASE_URL/api/v1/reports/$CHAT_ID/next")
[ "$STATUS" = "200" ] || fail "next expected 200, got $STATUS"
# 'dilewati' harus keluar dari sidebar (spec §7.2) — regresi T1 yang sempat lolos.
curl -sf -b "$JAR_BK" "$BASE_URL/api/v1/reports" | jq -e --arg id "$CHAT_ID" '.data | map(.id) | index($id) == null' > /dev/null \
  || fail "chat 'dilewati' masih muncul di sidebar"
curl -sf -b "$JAR_ANON" -X POST "$BASE_URL/api/v1/lapor/chats/$CHAT_ID/messages" \
  -F 'body=Baik, terima kasih.' > /dev/null
# requeue → status balik aktif ('baru'; 'dibuka' bila sudah dibuka BK) + masuk sidebar lagi.
curl -sf -b "$JAR_BK" "$BASE_URL/api/v1/reports" | jq -e --arg id "$CHAT_ID" '.data | map(.id) | index($id) != null' > /dev/null \
  || fail "chat yang di-requeue harus balik ke sidebar"
pass "Next keluar antrean + requeue ke antrean ok"

echo "==> [10/10] BK balas chat yang di-hide siswa -> 403"
curl -sf -b "$JAR_ANON" -X DELETE "$BASE_URL/api/v1/lapor/chats/$CHAT_ID" > /dev/null
STATUS=$(curl -s -o /dev/null -w '%{http_code}' -b "$JAR_BK" -X POST "$BASE_URL/api/v1/reports/$CHAT_ID/messages" \
  -H 'Content-Type: application/json' -d '{"body":"tes"}')
[ "$STATUS" = "403" ] || fail "expected 403, got $STATUS"
pass "chat tersembunyi menolak balasan"

echo
printf '\033[32mALL T1 SMOKE TESTS PASSED\033[0m\n'
