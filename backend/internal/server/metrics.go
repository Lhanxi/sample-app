package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const unmatchedRoute = "unmatched"

type HTTPMetrics struct {
	registry         *prometheus.Registry
	requestsTotal    *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	inFlightRequests prometheus.Gauge
}

func NewHTTPMetrics() *HTTPMetrics {
	metrics := &HTTPMetrics{
		registry: prometheus.NewRegistry(),
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "sample_backend",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests processed.",
			},
			[]string{"method", "route", "status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "sample_backend",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "Duration of HTTP requests in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "route", "status"},
		),
		inFlightRequests: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "sample_backend",
				Subsystem: "http",
				Name:      "in_flight_requests",
				Help:      "Current number of HTTP requests being processed.",
			},
		),
	}

	metrics.registry.MustRegister(
		metrics.requestsTotal,
		metrics.requestDuration,
		metrics.inFlightRequests,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return metrics
}

func (m *HTTPMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *HTTPMetrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		m.inFlightRequests.Inc()
		defer m.inFlightRequests.Dec()

		recorder := &responseRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		route := r.Pattern
		if route == "" {
			route = unmatchedRoute
		}

		labels := []string{r.Method, route, strconv.Itoa(recorder.status)}
		m.requestsTotal.WithLabelValues(labels...).Inc()
		m.requestDuration.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
	})
}
