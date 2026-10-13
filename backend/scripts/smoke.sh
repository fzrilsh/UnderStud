#!/usr/bin/env bash
# Smoke test — verifies the auth vertical slice end-to-end (T0).
# Usage: bash scripts/smoke.sh [BASE_URL]
# Requires: curl, jq. Server must be running, dan seed-demo sudah dijalankan.
set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"
JAR="$(mktemp)"
trap 'rm -f "$JAR"' EXIT

echo "==> [1/6] healthz"
curl -sf "$BASE_URL/healthz" | jq -e '.status == "ok"' > /dev/null

echo "==> [2/6] login staf seed (BK)"
curl -sf -X POST "$BASE_URL/api/v1/auth/login" \
  -H 'Content-Type: application/json' -c "$JAR" \
  -d '{"email":"bk@demo.local","password":"password123"}' \
  | jq -e '.data.user.email == "bk@demo.local" and .data.user.role == "BK"' > /dev/null

echo "==> [3/6] me (with cookie) — role ikut"
curl -sf -b "$JAR" "$BASE_URL/api/v1/auth/me" | jq -e '.data.user.role == "BK"' > /dev/null

echo "==> [4/6] login wali kelas — kelas ikut"
JAR2="$(mktemp)"
curl -sf -X POST "$BASE_URL/api/v1/auth/login" \
  -H 'Content-Type: application/json' -c "$JAR2" \
  -d '{"email":"wali@demo.local","password":"password123"}' \
  | jq -e '.data.user.role == "WALI_KELAS" and .data.user.kelas == "XI-IPA-1"' > /dev/null
rm -f "$JAR2"

echo "==> [5/6] login salah -> 401"
STATUS=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$BASE_URL/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"bk@demo.local","password":"salahbanget"}')
[ "$STATUS" = "401" ] || { echo "FAIL: expected 401, got $STATUS"; exit 1; }

echo "==> [6/6] logout -> 200, me after logout -> 401"
STATUS=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$BASE_URL/api/v1/auth/logout" -c "$JAR")
[ "$STATUS" = "200" ] || { echo "FAIL: logout expected 200, got $STATUS"; exit 1; }
STATUS=$(curl -s -o /dev/null -w '%{http_code}' -b "$JAR" "$BASE_URL/api/v1/auth/me")
[ "$STATUS" = "401" ] || { echo "FAIL: expected 401 after logout, got $STATUS"; exit 1; }

echo "ALL SMOKE TESTS PASSED"
