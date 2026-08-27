#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

status="$(curl -s -o /dev/null -w '%{http_code}' -X POST "${BASE_URL}/log/error")"
if [[ "$status" != "500" ]]; then
  echo "expected HTTP 500, got ${status}"
  exit 1
fi

echo "05-observe-logs: OK (POST /log/error → 500; check worker logs / Observe)"
