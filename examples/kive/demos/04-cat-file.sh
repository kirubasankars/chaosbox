#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

curl -sf "${BASE_URL}/_cat/file" >/dev/null
echo "04-cat-file: OK (/_cat/file returned 200)"
