#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

start="$(curl -sf "${BASE_URL}/count")"
echo "$start" | grep -qE '"count"[[:space:]]*:[[:space:]]*0' || { echo "expected count 0 at start: $start"; exit 1; }

curl -sf -X POST "${BASE_URL}/count/incr" >/dev/null
curl -sf -X POST "${BASE_URL}/count/incr" >/dev/null

end="$(curl -sf "${BASE_URL}/count")"
echo "$end" | grep -qE '"count"[[:space:]]*:[[:space:]]*2' || { echo "expected count 2: $end"; exit 1; }

echo "02-counter: OK (memory counter 0 → 2)"
