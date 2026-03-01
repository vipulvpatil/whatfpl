package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	reqTotal      prometheus.Counter
	err4xxTotal   prometheus.Counter
	err5xxTotal   prometheus.Counter
	inflightGauge prometheus.Gauge
	latency       prometheus.Histogram
}

func New() *Metrics {
	m := &Metrics{
		reqTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "whatfpl_requests_total",
			Help: "Total number of requests handled",
		}),
		err4xxTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "whatfpl_errors_4xx_total",
			Help: "Total number of 4xx errors",
		}),
		err5xxTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "whatfpl_errors_5xx_total",
			Help: "Total number of 5xx errors",
		}),
		inflightGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "whatfpl_inflight_requests",
			Help: "Number of requests currently in flight",
		}),
		latency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "whatfpl_request_duration_ms",
			Help:    "Request latency in milliseconds",
			Buckets: []float64{100, 150, 200, 250, 300, 400, 500, 750, 1000},
		}),
	}
	prometheus.MustRegister(m.reqTotal, m.err4xxTotal, m.err5xxTotal, m.inflightGauge, m.latency)
	return m
}

func (m *Metrics) RecordRequest(statusCode int, latencyMs int64) {
	m.reqTotal.Inc()
	m.latency.Observe(float64(latencyMs))
	switch {
	case statusCode >= 500:
		m.err5xxTotal.Inc()
	case statusCode >= 400:
		m.err4xxTotal.Inc()
	}
}

func (m *Metrics) InflightAdd(delta int64) {
	m.inflightGauge.Add(float64(delta))
}

func (m *Metrics) PrometheusHandler() http.Handler {
	return promhttp.Handler()
}
