#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

body="$(curl -sf "${BASE_URL}/")"
echo "$body" | grep -qE '"status"[[:space:]]*:[[:space:]]*"ok"' || { echo "expected status ok"; exit 1; }
echo "$body" | grep -q '"version"' || { echo "expected version field"; exit 1; }
echo "$body" | grep -qE '"cpu"[[:space:]]*:[[:space:]]*false' || { echo "expected idle cpu load"; exit 1; }

echo "01-health: OK (${BASE_URL}/)"
