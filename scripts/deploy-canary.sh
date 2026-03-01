#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "Deploying canary..."

docker build -t whatfpl:canary "$ROOT"
docker-compose -f "$ROOT/docker-compose.yml" --profile canary up -d --no-deps --force-recreate canary

echo "Done. Canary running on :8081."
