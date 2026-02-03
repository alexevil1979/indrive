// Package main — RideHail Ride Service (2026)
// Request, bidding, matching, status + Kafka events
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	httphandler "github.com/ridehail/ride/internal/delivery/http"
	"github.com/ridehail/ride/internal/infra/jwt"
	"github.com/ridehail/ride/internal/infra/kafka"
	"github.com/ridehail/ride/internal/infra/pg"
	"github.com/ridehail/ride/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := getEnv("PORT", "8083")
	pgDSN := getEnv("PG_DSN", "postgres://ridehail:ridehail_secret@localhost:5432/ridehail?sslmode=disable")
	kafkaBrokers := getEnv("KAFKA_BROKERS", "")
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

	var pub usecase.EventPublisher = &kafka.NoopProducer{}
	if kafkaBrokers != "" {
		brokers := strings.Split(kafkaBrokers, ",")
		for i := range brokers {
			brokers[i] = strings.TrimSpace(brokers[i])
		}
		kp, err := kafka.NewProducer(brokers)
		if err != nil {
			slog.Warn("kafka connect failed, using noop", "error", err)
		} else {
			defer kp.Close()
			pub = kp
			slog.Info("kafka ready")
		}
	}

	jwtValidator := jwt.NewValidator(jwtSecret)
	rideRepo := pg.NewRideRepo(pool)
	bidRepo := pg.NewBidRepo(pool)
	rideUC := usecase.NewRideUseCase(rideRepo, bidRepo, pub)

	e := echo.New()
	e.Use(middleware.Recover(), middleware.Logger(), middleware.RequestID())
	e.GET("/health", httphandler.Health)
	e.GET("/ready", httphandler.Ready(pool))
	api := e.Group("/api/v1")
	api.Use(httphandler.JWTAuth(jwtValidator))
	api.POST("/rides", httphandler.CreateRide(rideUC))
	api.GET("/rides", httphandler.ListMyRides(rideUC))
	api.GET("/rides/available", httphandler.ListAvailableRides(rideUC))
	api.GET("/admin/rides", httphandler.ListAllRides(rideUC))
	api.GET("/rides/:id", httphandler.GetRide(rideUC))
	api.POST("/rides/:id/bids", httphandler.PlaceBid(rideUC))
	api.GET("/rides/:id/bids", httphandler.ListBids(rideUC))
	api.POST("/rides/:id/accept", httphandler.AcceptBid(rideUC))
	api.PATCH("/rides/:id/status", httphandler.UpdateRideStatus(rideUC))

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
	slog.Info("ride service stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
