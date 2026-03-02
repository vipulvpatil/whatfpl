# whatfpl

Demo environment for AI-assisted deployment gates. Runs an HTTP server that validates FPL starting XIs and returns total points. Metrics are exposed in Prometheus format and used by [goodtogo](https://github.com/vipulvpatil/goodtogo) to gate canary promotions.

## Architecture

```
checker ──► baseline :8080  ──► /metrics ──┐
        └──► canary   :8081  ──► /metrics ──┴──► Prometheus :9090
                                                       │
                                               goodtogo check
```

Each instance exposes its own `GET /metrics`. Prometheus scrapes both and adds `job="baseline"` / `job="canary"` labels, so you can compare them in PromQL and with goodtogo.

---

## Deploy scripts

All scripts live in `scripts/` and handle building, tagging, and starting containers. Each build produces two tags: the role tag (`whatfpl:baseline` or `whatfpl:canary`) and a unique content tag (`whatfpl:<8-char-commit>-<YYYYMMDD-HHMMSS>`), which goodtogo uses to identify the build in its decision log.

### Healthy baseline

```bash
scripts/deploy-baseline.sh
```

Builds and starts `whatfpl:baseline` on `:8080` with the checker running against it.

### Healthy canary

```bash
scripts/deploy-canary.sh
```

Builds and starts `whatfpl:canary` on `:8081` alongside the running baseline.

### Faulty canary (rule-based catches it)

```bash
scripts/deploy-faulty-canary.sh
```

Injects high error rates and latency — goodtogo's rule-based checks will block promotion.

| Fault | Value |
|---|---|
| `FAULT_5XX_RATE` | 30% of requests return 5xx |
| `FAULT_4XX_RATE` | 15% of requests return 4xx |
| `FAULT_LATENCY_MEAN_MS` | Mean latency 800ms (vs ~200ms healthy) |

### p99-faulty canary (rule-based misses it — LLM blind spot demo)

```bash
scripts/deploy-p99-canary.sh
```

Injects a latency spike on 3% of requests (→ 3000ms). The p95 check passes (most traffic is unaffected) but p99 spikes. goodtogo approves; roll back manually to generate a `promoted`→`rolled-back` training example.

| Fault | Value |
|---|---|
| `FAULT_LATENCY_SPIKE_RATE` | 3% of requests → 3000ms |

### Promote canary to baseline (via goodtogo)

```bash
scripts/promote.sh
```

Runs `goodtogo` against the live metrics. If it passes, prompts for confirmation, then retags `whatfpl:canary` as `whatfpl:baseline` and does a rolling restart. Passes `GOODTOGO_DATA_DIR` so the decision is recorded next to the goodtogo repo.

---

## Fault injection environment variables

All fault variables default to `0` (disabled). Set them when starting the canary container:

| Variable | Type | Effect |
|---|---|---|
| `FAULT_5XX_RATE` | float [0,1] | Fraction of requests that return 500 |
| `FAULT_4XX_RATE` | float [0,1] | Fraction of requests that return 400 |
| `FAULT_LATENCY_MEAN_MS` | int (ms) | Shifts the mean of the simulated latency distribution |
| `FAULT_LATENCY_SPIKE_RATE` | float [0,1] | Fraction of requests that jump to 3000ms |

---

## Getting metrics

### Raw Prometheus text (per instance)

```
GET http://localhost:8080/metrics   # baseline
GET http://localhost:8081/metrics   # canary
```

### Prometheus UI

Open [http://localhost:9090](http://localhost:9090) and query across both instances.

**Useful queries:**

```promql
# request rate per job
sum by (job) (rate(whatfpl_requests_total[1m]))

# error rate (4xx + 5xx) per job
sum by (job) (rate(whatfpl_errors_4xx_total[1m])) + sum by (job) (rate(whatfpl_errors_5xx_total[1m]))

# p95 latency per job
histogram_quantile(0.95, sum by (le, job) (rate(whatfpl_request_duration_ms_bucket[1m])))

# p99 latency per job
histogram_quantile(0.99, sum by (le, job) (rate(whatfpl_request_duration_ms_bucket[1m])))

# compare p95: canary vs baseline ratio
histogram_quantile(0.95, sum by (le) (rate(whatfpl_request_duration_ms_bucket{job="canary"}[1m])))
  /
histogram_quantile(0.95, sum by (le) (rate(whatfpl_request_duration_ms_bucket{job="baseline"}[1m])))

# inflight requests
whatfpl_inflight_requests
```

---

## Typical workflow

```
1. scripts/deploy-baseline.sh        # stable baseline running
2. scripts/deploy-canary.sh          # canary alongside it
3. (wait 5 min for goodtogo to stabilise)
4. scripts/promote.sh                # goodtogo check → promote or abort
5. (if promoted, label the decision) cd ../goodtogo && ./label
```

To demo a false negative (p99 blind spot):

```
1. scripts/deploy-baseline.sh
2. scripts/deploy-p99-canary.sh      # 3% spike → high p99, passes p95 check
3. scripts/promote.sh                # goodtogo approves
4. (manually roll back)
5. ./label → mark as rolled-back
6. ./eval  → see rule-based FP, LLM may catch it
```

---

## Metrics reference

| Metric | Type | Description |
|---|---|---|
| `whatfpl_requests_total` | Counter | Total requests handled |
| `whatfpl_errors_4xx_total` | Counter | Client errors (invalid team, bad input) |
| `whatfpl_errors_5xx_total` | Counter | Server errors |
| `whatfpl_inflight_requests` | Gauge | Requests currently in flight |
| `whatfpl_request_duration_ms` | Histogram | Latency in ms (buckets: 100–3000ms) |
| `go_*` / `process_*` | — | Go runtime and process metrics (automatic) |

All metrics are labeled with `job` (`baseline` or `canary`) and `instance` by Prometheus at scrape time.

---

## API

### `GET /players?ids=1,2,3,...`

Validates a starting XI and returns total points.

**Rules:**
- Exactly 11 player IDs
- No duplicate IDs
- All IDs must exist in the FPL dataset
- Max 3 players from the same club
- Formation: 1 GK, 3–5 DEF, 2–5 MID, 1–3 FWD

**Response:**
```json
{"total_points": 42}
```

**Error responses:** `400 Bad Request` (invalid team), `500 Internal Server Error`.
