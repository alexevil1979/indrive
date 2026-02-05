// Package main — RideHail Auth Service (2026)
// Echo chosen: production-grade, middleware, OpenTelemetry-friendly.
// DDD: delivery (HTTP) -> usecase -> domain; infra (pg, redis) separate.
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

	httphandler "github.com/ridehail/auth/internal/delivery/http"
	"github.com/ridehail/auth/internal/infra/bcrypt"
	"github.com/ridehail/auth/internal/infra/jwt"
	"github.com/ridehail/auth/internal/infra/oauth"
	"github.com/ridehail/auth/internal/infra/pg"
	"github.com/ridehail/auth/internal/infra/redis"
	"github.com/ridehail/auth/internal/usecase"
)

const serviceName = "auth"

func main() {
	// Initialize structured JSON logger
	log := logger.Default(serviceName)
	log.Info("starting auth service")

	// Configuration
	port := getEnv("PORT", "8080")
	pgDSN := getEnv("PG_DSN", "postgres://ridehail:ridehail_secret@localhost:5432/ridehail?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-in-production")
	redisAddr := getEnv("REDIS_ADDR", "")
	otlpEndpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	// OAuth configuration
	baseURL := getEnv("BASE_URL", "http://localhost:8080")
	frontendURL := getEnv("FRONTEND_URL", "") // Where to redirect after OAuth

	oauthCfg := oauth.Config{
		Google: oauth.GoogleConfig{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  baseURL + "/auth/oauth/google/callback",
		},
		Yandex: oauth.YandexConfig{
			ClientID:     getEnv("YANDEX_CLIENT_ID", ""),
			ClientSecret: getEnv("YANDEX_CLIENT_SECRET", ""),
			RedirectURL:  baseURL + "/auth/oauth/yandex/callback",
		},
		VK: oauth.VKConfig{
			ClientID:     getEnv("VK_CLIENT_ID", ""),
			ClientSecret: getEnv("VK_CLIENT_SECRET", ""),
			RedirectURL:  baseURL + "/auth/oauth/vk/callback",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize OpenTelemetry tracing
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

	// Initialize Prometheus metrics
	m := metrics.New(metrics.Config{ServiceName: serviceName})

	// Connect to PostgreSQL
	pool, err := pg.Connect(ctx, pgDSN)
	if err != nil {
		log.Error("postgres connect", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pg.EnsurePostGIS(ctx, pool); err != nil {
		log.Error("postgis check", "error", err)
		os.Exit(1)
	}
	if err := pg.Migrate(ctx, pool); err != nil {
		log.Error("migrate", "error", err)
		os.Exit(1)
	}
	log.Info("postgres + PostGIS + migrations ready")

	// Connect to Redis (optional)
	rdb, err := redis.New(redisAddr)
	if err != nil {
		log.Error("redis connect", "error", err)
		os.Exit(1)
	}
	if rdb != nil {
		defer rdb.Close()
		log.Info("redis ready")
	}

	// Initialize repositories
	userRepo := pg.NewUserRepo(pool)
	oauthRepo := pg.NewOAuthRepo(pool)

	// Initialize use cases
	jwtUtil := jwt.New(jwtSecret, 24*time.Hour, 7*24*time.Hour)
	hasher := bcrypt.New(12)
	authUC := usecase.NewAuthUseCase(userRepo, jwtUtil, hasher)
	oauthUC := usecase.NewOAuthUseCase(userRepo, oauthRepo, jwtUtil, hasher)

	// Initialize OAuth manager
	oauthMgr := oauth.NewManager(oauthCfg)
	stateMgr := oauth.NewStateManager()
	oauthHandler := httphandler.NewOAuthHandler(oauthMgr, stateMgr, oauthUC, frontendURL)

	// Log configured OAuth providers
	providers := oauthMgr.ListProviders()
	if len(providers) > 0 {
		log.Info("oauth providers configured", "providers", providers)
	} else {
		log.Warn("no oauth providers configured")
	}

	// Setup Echo with observability middleware
	e := echo.New()
	e.HideBanner = true
	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(echoObservability(log, m))

	// Health & metrics
	e.GET("/health", httphandler.Health)
	e.GET("/ready", httphandler.Ready(pool))
	e.GET("/metrics", echo.WrapHandler(m.Handler()))

	// Auth routes
	e.POST("/auth/register", httphandler.Register(authUC))
	e.POST("/auth/login", httphandler.Login(authUC))
	e.POST("/auth/refresh", httphandler.Refresh(authUC))

	// OAuth routes
	e.GET("/auth/oauth/providers", oauthHandler.ListProviders())
	e.GET("/auth/oauth/:provider", oauthHandler.Redirect())
	e.GET("/auth/oauth/:provider/callback", oauthHandler.Callback())

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
	log.Info("auth service stopped")
}

// echoObservability adds tracing, metrics and logging to Echo requests.
func echoObservability(log *logger.Logger, m *metrics.Metrics) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			path := req.URL.Path
			method := req.Method

			// Track active requests
			m.IncActiveRequests(method, path)
			defer m.DecActiveRequests(method, path)

			// Execute handler
			err := next(c)

			// Record metrics
			duration := time.Since(start)
			status := c.Response().Status
			m.RecordRequest(method, path, string(rune(status/100)+'0')+"xx", duration)

			// Log request (skip noisy endpoints)
			if path != "/health" && path != "/ready" && path != "/metrics" {
				log.InfoContext(req.Context(), "http_request",
					"method", method,
					"path", path,
					"status", status,
					"duration_ms", duration.Milliseconds(),
					"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
				)
			}

			// Record errors
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
