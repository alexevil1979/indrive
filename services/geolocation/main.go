// Package main — RideHail Geolocation Service (2026)
// Driver tracking (Redis GEO), nearest search, WebSocket stub
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

	httphandler "github.com/ridehail/geolocation/internal/delivery/http"
	"github.com/ridehail/geolocation/internal/delivery/ws"
	"github.com/ridehail/geolocation/internal/infra/redis"
	"github.com/ridehail/geolocation/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := getEnv("PORT", "8082")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rdb, err := redis.New(redisAddr)
	if err != nil {
		slog.Error("redis connect", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()
	slog.Info("redis ready")

	geoStore := redis.NewGeoStore(rdb)
	locUC := usecase.NewLocationUseCase(geoStore)
	hub := ws.NewHub()
	go hub.Run()

	e := echo.New()
	e.Use(middleware.Recover(), middleware.Logger(), middleware.RequestID())
	e.GET("/health", httphandler.Health)
	e.POST("/api/v1/drivers/:id/location", httphandler.UpdateDriverLocation(locUC))
	e.GET("/api/v1/drivers/nearest", httphandler.NearestDrivers(locUC))
	e.GET("/ws/tracking", ws.HandleTracking(hub))

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
	slog.Info("geolocation service stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
