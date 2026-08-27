#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
MAX_WAIT="${MAX_WAIT:-15}"

curl -sf -X POST "${BASE_URL}/load/cpu/start" >/dev/null

deadline=$((SECONDS + MAX_WAIT))
while (( SECONDS < deadline )); do
  body="$(curl -sf "${BASE_URL}/")"
  if echo "$body" | grep -qE '"cpu"[[:space:]]*:[[:space:]]*true'; then
    break
  fi
  sleep 1
done

body="$(curl -sf "${BASE_URL}/")"
echo "$body" | grep -qE '"cpu"[[:space:]]*:[[:space:]]*true' || { echo "CPU load did not start within ${MAX_WAIT}s"; exit 1; }

curl -sf -X POST "${BASE_URL}/load/cpu/stop" >/dev/null

deadline=$((SECONDS + MAX_WAIT))
while (( SECONDS < deadline )); do
  body="$(curl -sf "${BASE_URL}/")"
  if echo "$body" | grep -qE '"cpu"[[:space:]]*:[[:space:]]*false'; then
    echo "03-load-cpu: OK (started and stopped CPU load)"
    exit 0
  fi
  sleep 1
done

echo "CPU load did not stop within ${MAX_WAIT}s"
exit 1
