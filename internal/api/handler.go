package api

import (
	"encoding/json"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vipulvpatil/whatfpl/internal/fpl"
	"github.com/vipulvpatil/whatfpl/internal/metrics"
)

type Config struct {
	FaultRate5xx  float64 // fraction of requests → 500
	FaultRate4xx  float64 // fraction of requests → 400
	LatencyMeanMs float64 // latency normal distribution mean (default 200)
}

func configFromEnv() Config {
	return Config{
		FaultRate5xx:  parseEnvFloat("FAULT_5XX_RATE", 0),
		FaultRate4xx:  parseEnvFloat("FAULT_4XX_RATE", 0),
		LatencyMeanMs: parseEnvFloat("FAULT_LATENCY_MEAN_MS", 200),
	}
}

func parseEnvFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func NewHandler(dm *fpl.DataManager) http.Handler {
	cfg := configFromEnv()
	m := metrics.New()
	mux := http.NewServeMux()
	mux.Handle("GET /players", withMetrics(m, handlePlayers(dm, cfg)))
	mux.Handle("GET /metrics", m.PrometheusHandler())
	return mux
}

func handlePlayers(dm *fpl.DataManager, cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.FaultRate5xx > 0 && rand.Float64() < cfg.FaultRate5xx {
			writeError(w, "internal fault injected", http.StatusInternalServerError)
			return
		}
		if cfg.FaultRate4xx > 0 && rand.Float64() < cfg.FaultRate4xx {
			writeError(w, "fault injected", http.StatusBadRequest)
			return
		}

		raw := r.URL.Query().Get("ids")
		if raw == "" {
			writeError(w, "missing ids", http.StatusBadRequest)
			return
		}

		parts := strings.Split(raw, ",")
		ids := make([]int, 0, len(parts))
		for _, part := range parts {
			id, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil {
				writeError(w, "invalid id: "+part, http.StatusBadRequest)
				return
			}
			ids = append(ids, id)
		}

		time.Sleep(simulatedLatency(cfg.LatencyMeanMs))

		store := dm.Store()

		if err := store.ValidateStartingTeam(ids); err != nil {
			writeError(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"total_points": store.TeamEventPoints(ids)})
	}
}

// withMetrics wraps a handler to track inflight count, status codes, and latency.
func withMetrics(m *metrics.Metrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.InflightAdd(1)
		defer m.InflightAdd(-1)

		rec := &statusRecorder{ResponseWriter: w}
		start := time.Now()
		next.ServeHTTP(rec, r)
		m.RecordRequest(rec.status, time.Since(start).Milliseconds())
	})
}

// simulatedLatency returns a duration drawn from a normal distribution.
// stddev=150ms, clamped to [100, 3000]ms.
func simulatedLatency(meanMs float64) time.Duration {
	ms := rand.NormFloat64()*150 + meanMs
	ms = math.Max(100, math.Min(3000, ms))
	return time.Duration(ms) * time.Millisecond
}

type statusRecorder struct {
	http.ResponseWriter
	status  int
	written bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if r.written {
		return
	}
	r.written = true
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.written {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(b)
}

func writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
