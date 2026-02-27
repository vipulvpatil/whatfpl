package metrics

import (
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ringSize         = 1000
	percentileInterval = 5 * time.Second
)

type Metrics struct {
	start    time.Time
	total    atomic.Int64
	err4xx   atomic.Int64
	err5xx   atomic.Int64
	inflight atomic.Int64

	mu      sync.Mutex
	ring    [ringSize]int64
	ringPos int
	full    bool

	p50 atomic.Int64
	p95 atomic.Int64
}

func New() *Metrics {
	m := &Metrics{start: time.Now()}
	go m.loop()
	return m
}

func (m *Metrics) RecordRequest(statusCode int, latencyMs int64) {
	m.total.Add(1)
	switch {
	case statusCode >= 500:
		m.err5xx.Add(1)
	case statusCode >= 400:
		m.err4xx.Add(1)
	}

	m.mu.Lock()
	m.ring[m.ringPos] = latencyMs
	m.ringPos++
	if m.ringPos == ringSize {
		m.ringPos = 0
		m.full = true
	}
	m.mu.Unlock()
}

func (m *Metrics) InflightAdd(delta int64) {
	m.inflight.Add(delta)
}

func (m *Metrics) loop() {
	ticker := time.NewTicker(percentileInterval)
	defer ticker.Stop()
	for range ticker.C {
		m.recompute()
	}
}

func (m *Metrics) recompute() {
	m.mu.Lock()
	var raw []int64
	if m.full {
		raw = make([]int64, ringSize)
		copy(raw, m.ring[:])
	} else {
		raw = make([]int64, m.ringPos)
		copy(raw, m.ring[:m.ringPos])
	}
	m.mu.Unlock()

	p50, p95 := percentiles(raw)
	m.p50.Store(p50)
	m.p95.Store(p95)
}

type Snapshot struct {
	UptimeSeconds  int64
	RequestsTotal  int64
	Errors4xxTotal int64
	Errors5xxTotal int64
	Inflight       int64
	P50Ms          int64
	P95Ms          int64
}

func (m *Metrics) Snapshot() Snapshot {
	return Snapshot{
		UptimeSeconds:  int64(time.Since(m.start).Seconds()),
		RequestsTotal:  m.total.Load(),
		Errors4xxTotal: m.err4xx.Load(),
		Errors5xxTotal: m.err5xx.Load(),
		Inflight:       m.inflight.Load(),
		P50Ms:          m.p50.Load(),
		P95Ms:          m.p95.Load(),
	}
}

func percentiles(samples []int64) (p50, p95 int64) {
	if len(samples) == 0 {
		return 0, 0
	}
	s := make([]int64, len(samples))
	copy(s, samples)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	p50 = s[int(math.Floor(float64(len(s)-1)*0.50))]
	p95 = s[int(math.Floor(float64(len(s)-1)*0.95))]
	return
}
