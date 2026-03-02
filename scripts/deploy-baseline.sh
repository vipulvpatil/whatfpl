#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "Deploying baseline..."

COMMIT=$(git -C "$ROOT" rev-parse --short=8 HEAD)
TS=$(date +%Y%m%d-%H%M%S)
docker build -t whatfpl:baseline -t "whatfpl:${COMMIT}-${TS}" "$ROOT"
docker compose -f "$ROOT/docker-compose.yml" up -d --no-deps --force-recreate baseline
docker compose -f "$ROOT/docker-compose.yml" up -d checker

echo "Done. Baseline running on :8080, checker running."
