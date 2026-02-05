// Package main — RideHail Payment Service (2026)
// Checkout flow, cash/card stub (later: Tinkoff, Sber, YooMoney)
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

	httphandler "github.com/ridehail/payment/internal/delivery/http"
	"github.com/ridehail/payment/internal/infra/jwt"
	"github.com/ridehail/payment/internal/infra/pg"
	"github.com/ridehail/payment/internal/usecase"
)

const serviceName = "payment"

func main() {
	log := logger.Default(serviceName)
	log.Info("starting payment service")

	port := getEnv("PORT", "8084")
	pgDSN := getEnv("PG_DSN", "postgres://ridehail:ridehail_secret@localhost:5432/ridehail?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-in-production")
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

	// Connect to PostgreSQL
	pool, err := pg.Connect(ctx, pgDSN)
	if err != nil {
		log.Error("postgres connect", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pg.Migrate(ctx, pool); err != nil {
		log.Error("migrate", "error", err)
		os.Exit(1)
	}
	log.Info("postgres + migrations ready")

	// Initialize use cases
	jwtValidator := jwt.NewValidator(jwtSecret)
	paymentRepo := pg.NewPaymentRepo(pool)
	paymentUC := usecase.NewPaymentUseCase(paymentRepo)

	// Setup Echo
	e := echo.New()
	e.HideBanner = true
	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(echoObservability(log, m))

	// Routes
	e.GET("/health", httphandler.Health)
	e.GET("/ready", httphandler.Ready(pool))
	e.GET("/metrics", echo.WrapHandler(m.Handler()))

	api := e.Group("/api/v1")
	api.Use(httphandler.JWTAuth(jwtValidator))
	api.POST("/payments", httphandler.CreatePayment(paymentUC))
	api.GET("/payments/ride/:rideId", httphandler.GetPaymentByRide(paymentUC))
	api.GET("/payments/:id", httphandler.GetPayment(paymentUC))
	api.POST("/payments/:id/confirm", httphandler.ConfirmPayment(paymentUC))

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
	log.Info("payment service stopped")
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

			if path != "/health" && path != "/ready" && path != "/metrics" {
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
