package api

import (
	"encoding/json"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vipulvpatil/whatfpl/internal/fpl"
	"github.com/vipulvpatil/whatfpl/internal/metrics"
)


func NewHandler(dm *fpl.DataManager) http.Handler {
	m := metrics.New()
	mux := http.NewServeMux()
	mux.Handle("GET /players", withMetrics(m, handlePlayers(dm)))
	mux.HandleFunc("GET /metrics", handleMetrics(m))
	return mux
}

func handlePlayers(dm *fpl.DataManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw := r.URL.Query().Get("ids")
		if raw == "" {
			http.Error(w, "missing ids", http.StatusBadRequest)
			return
		}

		parts := strings.Split(raw, ",")
		ids := make([]int, 0, len(parts))
		for _, part := range parts {
			id, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil {
				http.Error(w, "invalid id: "+part, http.StatusBadRequest)
				return
			}
			ids = append(ids, id)
		}

		time.Sleep(simulatedLatency())

		store := dm.Store()

		if err := store.ValidateStartingTeam(ids); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"total_points": store.TeamEventPoints(ids)})
	}
}

func handleMetrics(m *metrics.Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := m.Snapshot()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"uptime_seconds":  s.UptimeSeconds,
			"requests_total":  s.RequestsTotal,
			"errors_4xx_total": s.Errors4xxTotal,
			"errors_5xx_total": s.Errors5xxTotal,
			"inflight":        s.Inflight,
			"latency_ms": map[string]int64{
				"p50": s.P50Ms,
				"p95": s.P95Ms,
			},
		})
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
// mean=200ms, stddev=150ms, clamped to [100, 1000]ms.
func simulatedLatency() time.Duration {
	ms := rand.NormFloat64()*150 + 200
	ms = math.Max(100, math.Min(1000, ms))
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
