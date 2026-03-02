#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "Deploying p99-faulty canary..."
echo "3% of requests spike to 3000ms — p95 stays within threshold, p99 does not."
echo "goodtogo will pass this. Roll back manually after promotion to generate a training example."
echo ""

COMMIT=$(git -C "$ROOT" rev-parse --short=8 HEAD)
TS=$(date +%Y%m%d-%H%M%S)
docker build -t whatfpl:canary -t "whatfpl:${COMMIT}-${TS}" "$ROOT"
FAULT_LATENCY_SPIKE_RATE=0.03 \
  docker compose -f "$ROOT/docker-compose.yml" --profile canary up -d --no-deps --force-recreate canary
docker compose -f "$ROOT/docker-compose.yml" up -d checker

echo "Done. p99-faulty canary on :8081 (3% requests at 3000ms)."
