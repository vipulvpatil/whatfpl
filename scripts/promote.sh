#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GOODTOGO_BIN="${GOODTOGO_BIN:-$ROOT/../goodtogo/goodtogo}"
PROMETHEUS_URL="${PROMETHEUS_URL:-http://localhost:9090}"

if [[ ! -x "$GOODTOGO_BIN" ]]; then
  echo "goodtogo binary not found at $GOODTOGO_BIN"
  echo "Build it with: go build -o goodtogo ./cmd/check  (in the goodtogo repo)"
  echo "Or set GOODTOGO_BIN=/path/to/goodtogo"
  exit 1
fi

echo "Running goodtogo check..."
echo ""

set +e
PROMETHEUS_URL="$PROMETHEUS_URL" "$GOODTOGO_BIN"
GOODTOGO_EXIT=$?
set -e

echo ""

if [[ $GOODTOGO_EXIT -eq 0 ]]; then
  read -r -p "Proceed with promotion? [y/N] " REPLY
else
  read -r -p "NOT GOOD TO GO — promote anyway? [y/N] " REPLY
fi
echo ""

if [[ ! "$REPLY" =~ ^[Yy]$ ]]; then
  echo "Promotion cancelled."
  exit 1
fi

echo "Promoting canary → baseline..."

docker tag whatfpl:canary whatfpl:baseline
docker compose -f "$ROOT/docker-compose.yml" up -d --no-deps --force-recreate baseline
docker compose -f "$ROOT/docker-compose.yml" --profile canary stop canary
docker compose -f "$ROOT/docker-compose.yml" restart checker

echo "Done. Baseline updated on :8080, canary stopped, checker restarted (baseline-only traffic)."
