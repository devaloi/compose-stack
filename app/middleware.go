package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests by method, path, and status.",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency distribution.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	httpActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_active_connections",
		Help: "Currently active HTTP connections.",
	})

	dbConnectionsActive = promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "db_connections_active",
		Help: "Active PostgreSQL connections.",
	}, func() float64 { return 0 })

	redisCommandsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "redis_commands_total",
		Help: "Total Redis commands executed.",
	})
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		httpActiveConnections.Inc()
		defer httpActiveConnections.Dec()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		duration := time.Since(start).Seconds()
		path := normalizePath(r.URL.Path)
		httpRequestsTotal.WithLabelValues(r.Method, path, strconv.Itoa(rec.status)).Inc()
		httpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
	})
}

func normalizePath(path string) string {
	switch {
	case path == "/health":
		return "/health"
	case path == "/api/items":
		return "/api/items"
	case len(path) > 11 && path[:11] == "/api/items/":
		return "/api/items/{id}"
	case len(path) > 11 && path[:11] == "/api/cache/":
		return "/api/cache/{key}"
	default:
		return path
	}
}

func registerDBStatsCollector(s *server) {
	prometheus.Unregister(dbConnectionsActive)
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "db_connections_active",
		Help: "Active PostgreSQL connections.",
	}, func() float64 {
		if s.store == nil {
			return 0
		}
		return float64(s.store.DBStats().InUse)
	})
}
