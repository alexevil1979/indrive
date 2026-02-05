// Package logger provides structured JSON logging with OpenTelemetry trace context.
// Uses Go 1.21+ slog for structured logging with JSON output.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// Config holds logger configuration.
type Config struct {
	ServiceName string
	Level       slog.Level
	Output      io.Writer // defaults to os.Stdout
}

// Logger wraps slog.Logger with service context.
type Logger struct {
	*slog.Logger
	serviceName string
}

// New creates a new structured JSON logger.
func New(cfg Config) *Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level: cfg.Level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Rename time to timestamp for consistency
			if a.Key == slog.TimeKey {
				a.Key = "timestamp"
				a.Value = slog.StringValue(a.Value.Time().Format(time.RFC3339Nano))
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(cfg.Output, opts)
	baseLogger := slog.New(handler).With(
		slog.String("service", cfg.ServiceName),
	)

	return &Logger{
		Logger:      baseLogger,
		serviceName: cfg.ServiceName,
	}
}

// WithContext returns a logger with trace context (trace_id, span_id) if present.
func (l *Logger) WithContext(ctx context.Context) *slog.Logger {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return l.Logger
	}

	return l.Logger.With(
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
	)
}

// Info logs at INFO level with context.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.WithContext(ctx).Info(msg, args...)
}

// Error logs at ERROR level with context.
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.WithContext(ctx).Error(msg, args...)
}

// Warn logs at WARN level with context.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.WithContext(ctx).Warn(msg, args...)
}

// Debug logs at DEBUG level with context.
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.WithContext(ctx).Debug(msg, args...)
}

// Default returns a default logger for the given service.
func Default(serviceName string) *Logger {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}
	return New(Config{
		ServiceName: serviceName,
		Level:       level,
	})
}
