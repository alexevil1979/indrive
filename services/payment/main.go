// Package main — RideHail Payment Service (2026)
// Checkout flow, cash/card stub (later: Tinkoff, Sber, YooMoney)
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	httphandler "github.com/ridehail/payment/internal/delivery/http"
	"github.com/ridehail/payment/internal/infra/jwt"
	"github.com/ridehail/payment/internal/infra/pg"
	"github.com/ridehail/payment/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := getEnv("PORT", "8084")
	pgDSN := getEnv("PG_DSN", "postgres://ridehail:ridehail_secret@localhost:5432/ridehail?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-in-production")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pg.Connect(ctx, pgDSN)
	if err != nil {
		slog.Error("postgres connect", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pg.Migrate(ctx, pool); err != nil {
		slog.Error("migrate", "error", err)
		os.Exit(1)
	}
	slog.Info("postgres + migrations ready")

	jwtValidator := jwt.NewValidator(jwtSecret)
	paymentRepo := pg.NewPaymentRepo(pool)
	paymentUC := usecase.NewPaymentUseCase(paymentRepo)

	e := echo.New()
	e.Use(middleware.Recover(), middleware.Logger(), middleware.RequestID())
	e.GET("/health", httphandler.Health)
	e.GET("/ready", httphandler.Ready(pool))
	api := e.Group("/api/v1")
	api.Use(httphandler.JWTAuth(jwtValidator))
	api.POST("/payments", httphandler.CreatePayment(paymentUC))
	api.GET("/payments/ride/:rideId", httphandler.GetPaymentByRide(paymentUC))
	api.GET("/payments/:id", httphandler.GetPayment(paymentUC))
	api.POST("/payments/:id/confirm", httphandler.ConfirmPayment(paymentUC))

	go func() {
		if err := e.Start(":" + port); err != nil && err != echo.ErrServerClosed {
			slog.Error("server", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down...")
	graceCtx, graceCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer graceCancel()
	if err := e.Shutdown(graceCtx); err != nil {
		slog.Error("shutdown", "error", err)
	}
	slog.Info("payment service stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
