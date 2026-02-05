// Package main — RideHail Geolocation Service (2026)
// Driver tracking (Redis GEO), nearest search, WebSocket stub
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/alexevil1979/indrive/packages/otel-go/logger"
	"github.com/alexevil1979/indrive/packages/otel-go/metrics"
	"github.com/alexevil1979/indrive/packages/otel-go/tracing"

	httphandler "github.com/ridehail/geolocation/internal/delivery/http"
	"github.com/ridehail/geolocation/internal/delivery/ws"
	"github.com/ridehail/geolocation/internal/infra/redis"
	"github.com/ridehail/geolocation/internal/usecase"
)

const serviceName = "geolocation"

func main() {
	log := logger.Default(serviceName)
	log.Info("starting geolocation service")

	port := getEnv("PORT", "8082")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	otlpEndpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize tracing
	tracerProvider, err := tracing.Init(ctx, tracing.Config{
		ServiceName:    serviceName,
		ServiceVersion: "0.1.0",
		Environment:    getEnv("ENV", "development"),
		OTLPEndpoint:   otlpEndpoint,
		Enabled:        otlpEndpoint != "",
	})
	if err != nil {
		log.Error("tracing init", "error", err)
		os.Exit(1)
	}
	defer tracerProvider.Shutdown(context.Background())

	// Initialize metrics
	m := metrics.New(metrics.Config{ServiceName: serviceName})

	// Connect to Redis
	rdb, err := redis.New(redisAddr)
	if err != nil {
		log.Error("redis connect", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()
	log.Info("redis ready")

	// Initialize use cases
	geoStore := redis.NewGeoStore(rdb)
	locUC := usecase.NewLocationUseCase(geoStore)
	hub := ws.NewHub()
	go hub.Run()

	// Setup Echo
	e := echo.New()
	e.HideBanner = true
	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(echoObservability(log, m))

	// Routes
	e.GET("/health", httphandler.Health)
	e.GET("/metrics", echo.WrapHandler(m.Handler()))
	e.POST("/api/v1/drivers/:id/location", httphandler.UpdateDriverLocation(locUC))
	e.GET("/api/v1/drivers/nearest", httphandler.NearestDrivers(locUC))
	e.GET("/ws/tracking", ws.HandleTracking(hub))

	// Start server
	go func() {
		log.Info("listening", "port", port)
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Error("server", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down...")
	graceCtx, graceCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer graceCancel()
	if err := e.Shutdown(graceCtx); err != nil {
		log.Error("shutdown", "error", err)
	}
	log.Info("geolocation service stopped")
}

func echoObservability(log *logger.Logger, m *metrics.Metrics) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			path := req.URL.Path
			method := req.Method

			m.IncActiveRequests(method, path)
			defer m.DecActiveRequests(method, path)

			err := next(c)

			duration := time.Since(start)
			status := c.Response().Status
			m.RecordRequest(method, path, string(rune(status/100)+'0')+"xx", duration)

			if path != "/health" && path != "/metrics" {
				log.InfoContext(req.Context(), "http_request",
					"method", method, "path", path, "status", status,
					"duration_ms", duration.Milliseconds(),
					"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
				)
			}

			if status >= 500 {
				m.RecordError("http_5xx")
			}

			return err
		}
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
