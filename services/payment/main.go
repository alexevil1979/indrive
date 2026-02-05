// Package main — RideHail Payment Service (2026)
// Checkout flow with Tinkoff, YooMoney, Sber integrations
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
	"github.com/ridehail/payment/internal/domain"
	"github.com/ridehail/payment/internal/infra/gateway"
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

	// Payment provider configuration
	baseURL := getEnv("BASE_URL", "http://localhost:8084")
	tinkoffTerminalKey := getEnv("TINKOFF_TERMINAL_KEY", "")
	tinkoffPassword := getEnv("TINKOFF_PASSWORD", "")
	tinkoffTestMode := getEnv("TINKOFF_TEST_MODE", "true") == "true"
	yooMoneyShopID := getEnv("YOOMONEY_SHOP_ID", "")
	yooMoneySecretKey := getEnv("YOOMONEY_SECRET_KEY", "")
	yooMoneyWebhookSecret := getEnv("YOOMONEY_WEBHOOK_SECRET", "")
	sberUserName := getEnv("SBER_USERNAME", "")
	sberPassword := getEnv("SBER_PASSWORD", "")
	sberToken := getEnv("SBER_TOKEN", "")
	sberTestMode := getEnv("SBER_TEST_MODE", "true") == "true"

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

	// Initialize payment gateways
	gwManager := gateway.NewManager()

	// Register cash gateway (always available)
	gwManager.Register(gateway.NewCashGateway())

	// Register Tinkoff gateway
	if tinkoffTerminalKey != "" {
		tinkoffGW := gateway.NewTinkoffGateway(gateway.TinkoffConfig{
			TerminalKey: tinkoffTerminalKey,
			Password:    tinkoffPassword,
			TestMode:    tinkoffTestMode,
			NotifyURL:   baseURL + "/webhooks/tinkoff",
		})
		gwManager.Register(tinkoffGW)
		log.Info("tinkoff gateway registered", "test_mode", tinkoffTestMode)
	}

	// Register YooMoney gateway
	if yooMoneyShopID != "" {
		yooMoneyGW := gateway.NewYooMoneyGateway(gateway.YooMoneyConfig{
			ShopID:        yooMoneyShopID,
			SecretKey:     yooMoneySecretKey,
			WebhookSecret: yooMoneyWebhookSecret,
			ReturnURL:     baseURL + "/payment/success",
		})
		gwManager.Register(yooMoneyGW)
		log.Info("yoomoney gateway registered")
	}

	// Register Sber gateway
	if sberUserName != "" || sberToken != "" {
		sberGW := gateway.NewSberGateway(gateway.SberConfig{
			UserName:  sberUserName,
			Password:  sberPassword,
			Token:     sberToken,
			TestMode:  sberTestMode,
			ReturnURL: baseURL + "/payment/success",
			FailURL:   baseURL + "/payment/fail",
		})
		gwManager.Register(sberGW)
		log.Info("sber gateway registered", "test_mode", sberTestMode)
	}

	// Initialize use cases
	jwtValidator := jwt.NewValidator(jwtSecret)
	paymentRepo := pg.NewPaymentRepo(pool)
	promoRepo := pg.NewPromoRepo(pool)
	paymentUC := usecase.NewPaymentUseCase(paymentRepo, gwManager)
	promoUC := usecase.NewPromoUseCase(promoRepo)
	promoHandler := httphandler.NewPromoHandler(promoUC)

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

	// Public routes (webhooks)
	e.POST("/webhooks/tinkoff", httphandler.Webhook(paymentUC, domain.ProviderTinkoff))
	e.POST("/webhooks/yoomoney", httphandler.Webhook(paymentUC, domain.ProviderYooMoney))
	e.POST("/webhooks/sber", httphandler.Webhook(paymentUC, domain.ProviderSber))

	// Providers info (public)
	e.GET("/api/v1/payments/providers", httphandler.GetProviders(paymentUC))

	api := e.Group("/api/v1")
	api.Use(httphandler.JWTAuth(jwtValidator))

	// Payment operations
	api.POST("/payments", httphandler.CreatePayment(paymentUC))
	api.GET("/payments", httphandler.ListPayments(paymentUC))
	api.GET("/payments/ride/:rideId", httphandler.GetPaymentByRide(paymentUC))
	api.GET("/payments/:id", httphandler.GetPayment(paymentUC))
	api.POST("/payments/:id/confirm", httphandler.ConfirmPayment(paymentUC))
	api.POST("/payments/:id/refund", httphandler.RefundPayment(paymentUC))

	// Payment methods (saved cards)
	api.GET("/payment-methods", httphandler.ListPaymentMethods(paymentUC))
	api.DELETE("/payment-methods/:id", httphandler.DeletePaymentMethod(paymentUC))
	api.POST("/payment-methods/:id/default", httphandler.SetDefaultPaymentMethod(paymentUC))

	// Promo codes
	api.GET("/promos", promoHandler.ListActivePromos)
	api.POST("/promos/validate", promoHandler.ValidatePromo)
	api.POST("/promos/apply", promoHandler.ApplyPromo)
	api.GET("/promos/my", promoHandler.GetMyPromos)

	// Admin promo management
	api.GET("/admin/promos", promoHandler.ListPromos)
	api.POST("/admin/promos", promoHandler.CreatePromo)
	api.GET("/admin/promos/:id", promoHandler.GetPromo)
	api.PUT("/admin/promos/:id", promoHandler.UpdatePromo)
	api.DELETE("/admin/promos/:id", promoHandler.DeletePromo)

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
