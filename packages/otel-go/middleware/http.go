// Package middleware provides HTTP middleware for observability.
package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/alexevil1979/indrive/packages/otel-go/logger"
	"github.com/alexevil1979/indrive/packages/otel-go/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Observability returns middleware that adds tracing, metrics, and logging.
func Observability(serviceName string, log *logger.Logger, m *metrics.Metrics) func(http.Handler) http.Handler {
	tracer := otel.Tracer(serviceName)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := r.URL.Path
			method := r.Method

			// Extract trace context from incoming request
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Start span
			ctx, span := tracer.Start(ctx, fmt.Sprintf("%s %s", method, path),
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					semconv.HTTPMethodKey.String(method),
					semconv.HTTPTargetKey.String(r.RequestURI),
					semconv.HTTPSchemeKey.String(r.URL.Scheme),
					attribute.String("http.user_agent", r.UserAgent()),
				),
			)
			defer span.End()

			// Track active requests
			if m != nil {
				m.IncActiveRequests(method, path)
				defer m.DecActiveRequests(method, path)
			}

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Inject trace context into response headers (for debugging)
			otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))

			// Call next handler
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Record metrics and log
			duration := time.Since(start)
			status := fmt.Sprintf("%d", wrapped.statusCode)

			span.SetAttributes(semconv.HTTPStatusCodeKey.Int(wrapped.statusCode))

			if m != nil {
				m.RecordRequest(method, path, status, duration)
			}

			// Log request (skip health/ready/metrics to reduce noise)
			if log != nil && path != "/health" && path != "/ready" && path != "/metrics" {
				log.InfoContext(ctx, "http_request",
					"method", method,
					"path", path,
					"status", wrapped.statusCode,
					"duration_ms", duration.Milliseconds(),
					"remote_addr", r.RemoteAddr,
				)
			}

			// Record errors
			if wrapped.statusCode >= 500 && m != nil {
				m.RecordError("http_5xx")
			}
		})
	}
}

// Recovery returns middleware that recovers from panics.
func Recovery(log *logger.Logger, m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					if m != nil {
						m.RecordError("panic")
					}
					if log != nil {
						log.ErrorContext(r.Context(), "panic_recovered",
							"error", fmt.Sprintf("%v", err),
							"path", r.URL.Path,
						)
					}
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
