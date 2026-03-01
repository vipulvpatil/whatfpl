#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "Promoting canary → baseline..."

docker tag whatfpl:canary whatfpl:baseline
docker compose -f "$ROOT/docker-compose.yml" up -d --no-deps --force-recreate baseline
docker compose -f "$ROOT/docker-compose.yml" --profile canary stop canary

echo "Done. Baseline updated on :8080, canary stopped."
