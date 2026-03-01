#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "Deploying baseline..."

docker build -t whatfpl:baseline "$ROOT"
docker compose -f "$ROOT/docker-compose.yml" up -d --no-deps --force-recreate baseline

echo "Done. Baseline running on :8080."
