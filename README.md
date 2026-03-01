# whatfpl

Demo environment for AI-assisted deployment gates. Runs an HTTP server that validates FPL starting XIs and returns total points. Metrics are exposed in Prometheus format.

## Architecture

```
checker ──► baseline :8080  ──► /metrics
        └──► canary   :8081  ──► /metrics
                                    ▲
                              Prometheus :9090
                           (scrapes both separately)
```

Each instance exposes its **own** metrics at `GET /metrics`. `/metrics` on the baseline returns only baseline data; `/metrics` on the canary returns only canary data. Prometheus scrapes both and adds `job="baseline"` / `job="canary"` labels, so you can compare them in PromQL.

---

## Deploying

### 1. Build images

```bash
# baseline — current production image
docker build -t whatfpl:baseline .

# canary — new version being tested (build from a different branch/commit if needed)
docker build -t whatfpl:canary .
```

### 2. Start baseline + Prometheus

```bash
docker compose up -d
```

This starts:
- `baseline` on `:8080`
- `prometheus` on `:9090`

### 3. Add canary alongside baseline

```bash
docker compose --profile canary up -d
```

This starts `canary` on `:8081` while baseline keeps running. Prometheus is already configured to scrape both.

### 4. Generate load (optional)

The checker fires random valid and invalid teams at the server(s):

```bash
# against baseline only
go run ./cmd/checker

# split across baseline and canary
go run ./cmd/checker -targets http://localhost:8080,http://localhost:8081
```

---

## Getting metrics

### Raw Prometheus text (per instance)

```
GET http://localhost:8080/metrics   # baseline
GET http://localhost:8081/metrics   # canary
```

These return standard Prometheus text format for that instance only.

### Prometheus UI

Open [http://localhost:9090](http://localhost:9090) and query across both instances.

**Useful queries:**

```promql
# request rate per job
rate(whatfpl_requests_total[1m])

# error rate (4xx + 5xx) per job
rate(whatfpl_errors_4xx_total[1m]) + rate(whatfpl_errors_5xx_total[1m])

# p95 latency per job
histogram_quantile(0.95, rate(whatfpl_request_duration_ms_bucket[1m]))

# compare p95: canary vs baseline
histogram_quantile(0.95, rate(whatfpl_request_duration_ms_bucket{job="canary"}[1m]))
  /
histogram_quantile(0.95, rate(whatfpl_request_duration_ms_bucket{job="baseline"}[1m]))

# inflight requests
whatfpl_inflight_requests
```

---

## Promoting canary to baseline

Once the canary looks good, retag it and do a rolling restart:

```bash
# 1. stop canary profile
docker compose --profile canary stop canary

# 2. retag as baseline
docker tag whatfpl:canary whatfpl:baseline

# 3. restart baseline with the new image
docker compose up -d --force-recreate baseline
```

The old baseline image is replaced. Run `docker compose --profile canary down` to clean up the canary service entirely.

---

## Metrics reference

| Metric | Type | Description |
|---|---|---|
| `whatfpl_requests_total` | Counter | Total requests handled |
| `whatfpl_errors_4xx_total` | Counter | Client errors (invalid team, bad input) |
| `whatfpl_errors_5xx_total` | Counter | Server errors |
| `whatfpl_inflight_requests` | Gauge | Requests currently in flight |
| `whatfpl_request_duration_ms` | Histogram | Latency in ms (buckets: 100–1000ms) |
| `go_*` / `process_*` | — | Go runtime and process metrics (automatic) |

All metrics are labeled with `job` (`baseline` or `canary`) and `instance` by Prometheus at scrape time.
