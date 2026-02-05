// Package metrics provides Prometheus metrics and OpenTelemetry metrics bridge.
package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds common HTTP metrics for a service.
type Metrics struct {
	serviceName string
	registry    *prometheus.Registry

	// HTTP metrics
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	RequestsActive  *prometheus.GaugeVec
	ErrorsTotal     *prometheus.CounterVec
}

// Config holds metrics configuration.
type Config struct {
	ServiceName string
	Namespace   string // optional, defaults to "ridehail"
}

// New creates a new Metrics instance with Prometheus collectors.
func New(cfg Config) *Metrics {
	if cfg.Namespace == "" {
		cfg.Namespace = "ridehail"
	}

	registry := prometheus.NewRegistry()
	// Register default Go collectors
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	factory := promauto.With(registry)

	m := &Metrics{
		serviceName: cfg.ServiceName,
		registry:    registry,

		RequestsTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: cfg.Namespace,
			Subsystem: cfg.ServiceName,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		}, []string{"method", "path", "status"}),

		RequestDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: cfg.Namespace,
			Subsystem: cfg.ServiceName,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"method", "path", "status"}),

		RequestsActive: factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: cfg.Namespace,
			Subsystem: cfg.ServiceName,
			Name:      "http_requests_active",
			Help:      "Number of active HTTP requests",
		}, []string{"method", "path"}),

		ErrorsTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: cfg.Namespace,
			Subsystem: cfg.ServiceName,
			Name:      "errors_total",
			Help:      "Total number of errors",
		}, []string{"type"}),
	}

	return m
}

// Handler returns an http.Handler for the /metrics endpoint.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// RecordRequest records metrics for an HTTP request.
func (m *Metrics) RecordRequest(method, path, status string, duration time.Duration) {
	m.RequestsTotal.WithLabelValues(method, path, status).Inc()
	m.RequestDuration.WithLabelValues(method, path, status).Observe(duration.Seconds())
}

// IncActiveRequests increments active requests gauge.
func (m *Metrics) IncActiveRequests(method, path string) {
	m.RequestsActive.WithLabelValues(method, path).Inc()
}

// DecActiveRequests decrements active requests gauge.
func (m *Metrics) DecActiveRequests(method, path string) {
	m.RequestsActive.WithLabelValues(method, path).Dec()
}

// RecordError records an error by type.
func (m *Metrics) RecordError(errType string) {
	m.ErrorsTotal.WithLabelValues(errType).Inc()
}

// Registry returns the Prometheus registry for custom collectors.
func (m *Metrics) Registry() *prometheus.Registry {
	return m.registry
}
