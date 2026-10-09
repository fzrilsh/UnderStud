#!/usr/bin/env bash
# Smoke test — verifies the auth vertical slice end-to-end.
# Usage: bash scripts/smoke.sh [BASE_URL]
# Requires: curl, jq. Server must be running (make run / make dev / docker-compose up).
set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"
JAR="$(mktemp)"
trap 'rm -f "$JAR"' EXIT
EMAIL="smoke-$(date +%s)@example.com"
PASS="smoke123"

echo "==> [1/8] healthz"
curl -sf "$BASE_URL/healthz" | jq -e '.status == "ok"' > /dev/null

echo "==> [2/8] register ($EMAIL)"
curl -sf -X POST "$BASE_URL/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -c "$JAR" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\",\"name\":\"Smoke Tester\"}" \
  | jq -e '.user.id and (.user.email == "'"$EMAIL"'")' > /dev/null

echo "==> [3/8] me (with cookie)"
curl -sf -b "$JAR" "$BASE_URL/api/v1/auth/me" | jq -e '.user.email == "'"$EMAIL"'"' > /dev/null

echo "==> [4/8] duplicate register -> 409"
STATUS=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$BASE_URL/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\",\"name\":\"Smoke Tester\"}")
[ "$STATUS" = "409" ] || { echo "FAIL: expected 409, got $STATUS"; exit 1; }

echo "==> [5/8] login"
curl -sf -X POST "$BASE_URL/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -c "$JAR" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}" \
  | jq -e '.user.email == "'"$EMAIL"'"' > /dev/null

echo "==> [6/8] me again (after login)"
curl -sf -b "$JAR" "$BASE_URL/api/v1/auth/me" | jq -e '.user.email == "'"$EMAIL"'"' > /dev/null

echo "==> [7/8] logout -> 200"
STATUS=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$BASE_URL/api/v1/auth/logout" -c "$JAR")
[ "$STATUS" = "200" ] || { echo "FAIL: logout expected 200, got $STATUS"; exit 1; }

echo "==> [8/8] me after logout -> 401"
STATUS=$(curl -s -o /dev/null -w '%{http_code}' -b "$JAR" "$BASE_URL/api/v1/auth/me")
[ "$STATUS" = "401" ] || { echo "FAIL: expected 401 after logout, got $STATUS"; exit 1; }

echo "ALL SMOKE TESTS PASSED"
