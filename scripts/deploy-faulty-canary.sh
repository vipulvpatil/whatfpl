#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "Deploying faulty canary..."

COMMIT=$(git -C "$ROOT" rev-parse --short=8 HEAD)
TS=$(date +%Y%m%d-%H%M%S)
docker build -t whatfpl:canary -t "whatfpl:${COMMIT}-${TS}" "$ROOT"
FAULT_5XX_RATE=0.3 FAULT_4XX_RATE=0.15 FAULT_LATENCY_MEAN_MS=800 \
  docker compose -f "$ROOT/docker-compose.yml" --profile canary up -d --no-deps --force-recreate canary
docker compose -f "$ROOT/docker-compose.yml" up -d checker

echo "Done. Faulty canary on :8081 (30% 5xx, 15% 4xx, 800ms latency mean)."
