#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "Deploying canary..."

COMMIT=$(git -C "$ROOT" rev-parse --short=8 HEAD)
TS=$(date +%Y%m%d-%H%M%S)
docker build -t whatfpl:canary -t "whatfpl:${COMMIT}-${TS}" "$ROOT"
docker compose -f "$ROOT/docker-compose.yml" --profile canary up -d --no-deps --force-recreate canary
docker compose -f "$ROOT/docker-compose.yml" up -d checker

echo "Done. Canary running on :8081, checker running."
