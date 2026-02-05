// Package main — RideHail User Service (2026)
// DDD: delivery -> usecase -> domain; infra pg + Redis
package main

import (
	"context"
	"io"
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

	httphandler "github.com/ridehail/user/internal/delivery/http"
	"github.com/ridehail/user/internal/infra/jwt"
	"github.com/ridehail/user/internal/infra/pg"
	"github.com/ridehail/user/internal/infra/redis"
	"github.com/ridehail/user/internal/infra/storage"
	"github.com/ridehail/user/internal/usecase"
)

const serviceName = "user"

func main() {
	log := logger.Default(serviceName)
	log.Info("starting user service")

	port := getEnv("PORT", "8081")
	pgDSN := getEnv("PG_DSN", "postgres://ridehail:ridehail_secret@localhost:5432/ridehail?sslmode=disable")
	redisAddr := getEnv("REDIS_ADDR", "")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-in-production")
	otlpEndpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	// MinIO configuration
	minioEndpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "ridehail_minio")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "ridehail_minio_secret")
	minioBucket := getEnv("MINIO_BUCKET", "ridehail-documents")
	minioPublicURL := getEnv("MINIO_PUBLIC_URL", "")

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
	log.Info("postgres ready")

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

	// Initialize MinIO storage client
	storageClient, err := storage.New(storage.Config{
		Endpoint:        minioEndpoint,
		AccessKeyID:     minioAccessKey,
		SecretAccessKey: minioSecretKey,
		Bucket:          minioBucket,
		UseSSL:          false,
		PublicURL:       minioPublicURL,
	})
	if err != nil {
		log.Error("minio connect", "error", err)
		os.Exit(1)
	}
	if err := storageClient.EnsureBucket(ctx); err != nil {
		log.Error("minio bucket", "error", err)
		os.Exit(1)
	}
	log.Info("minio ready", "bucket", minioBucket)

	// Initialize repositories
	profileRepo := pg.NewProfileRepo(pool)
	verificationRepo := pg.NewVerificationRepo(pool)

	// Initialize use cases
	jwtValidator := jwt.NewValidator(jwtSecret)
	profileUC := usecase.NewProfileUseCase(profileRepo)
	verificationUC := usecase.NewVerificationUseCase(verificationRepo, &storageAdapter{client: storageClient})

	// Verification handler
	verificationHandler := httphandler.NewVerificationHandler(verificationUC)

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
	api.GET("/users/me", httphandler.GetProfile(profileUC))
	api.PATCH("/users/me", httphandler.UpdateProfile(profileUC))
	api.POST("/users/me/driver", httphandler.CreateDriverProfile(profileUC))

	// Driver verification routes
	api.POST("/verification", verificationHandler.StartVerification())
	api.GET("/verification", verificationHandler.GetVerificationStatus())
	api.POST("/verification/documents", verificationHandler.UploadDocument())
	api.GET("/verification/documents", verificationHandler.ListDocuments())
	api.GET("/verification/documents/:id", verificationHandler.GetDocument())
	api.DELETE("/verification/documents/:id", verificationHandler.DeleteDocument())

	// Admin verification routes
	admin := e.Group("/api/v1/admin")
	admin.Use(httphandler.JWTAuth(jwtValidator))
	admin.GET("/verifications", verificationHandler.ListPendingVerifications())
	admin.GET("/verifications/:id", verificationHandler.GetVerificationByID())
	admin.POST("/verifications/:id/review", verificationHandler.ReviewVerification())
	admin.POST("/documents/:id/review", verificationHandler.ReviewDocument())

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
	log.Info("user service stopped")
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

// storageAdapter adapts storage.Client to usecase.StorageClient interface
type storageAdapter struct {
	client *storage.Client
}

func (a *storageAdapter) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (*usecase.UploadResult, error) {
	result, err := a.client.Upload(ctx, key, reader, size, contentType)
	if err != nil {
		return nil, err
	}
	return &usecase.UploadResult{
		Key:         result.Key,
		Size:        result.Size,
		ContentType: result.ContentType,
		URL:         result.URL,
	}, nil
}

func (a *storageAdapter) Delete(ctx context.Context, key string) error {
	return a.client.Delete(ctx, key)
}

func (a *storageAdapter) GetPresignedURL(ctx context.Context, key string, expires interface{}) (string, error) {
	if d, ok := expires.(time.Duration); ok {
		return a.client.GetPresignedURL(ctx, key, d)
	}
	return a.client.GetPresignedURL(ctx, key, 15*time.Minute)
}
