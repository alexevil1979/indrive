// Package main — RideHail User Service (2026)
// DDD: delivery -> usecase -> domain; infra pg + Redis
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

	httphandler "github.com/ridehail/user/internal/delivery/http"
	"github.com/ridehail/user/internal/infra/jwt"
	"github.com/ridehail/user/internal/infra/pg"
	"github.com/ridehail/user/internal/infra/redis"
	"github.com/ridehail/user/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := getEnv("PORT", "8081")
	pgDSN := getEnv("PG_DSN", "postgres://ridehail:ridehail_secret@localhost:5432/ridehail?sslmode=disable")
	redisAddr := getEnv("REDIS_ADDR", "")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-in-production")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pg.Connect(ctx, pgDSN)
	if err != nil {
		slog.Error("postgres connect", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("postgres ready")

	rdb, err := redis.New(redisAddr)
	if err != nil {
		slog.Error("redis connect", "error", err)
		os.Exit(1)
	}
	if rdb != nil {
		defer rdb.Close()
		slog.Info("redis ready")
	}

	profileRepo := pg.NewProfileRepo(pool)
	jwtValidator := jwt.NewValidator(jwtSecret)
	profileUC := usecase.NewProfileUseCase(profileRepo)

	e := echo.New()
	e.Use(middleware.Recover(), middleware.Logger(), middleware.RequestID())
	e.GET("/health", httphandler.Health)
	e.GET("/ready", httphandler.Ready(pool))
	api := e.Group("/api/v1")
	api.Use(httphandler.JWTAuth(jwtValidator))
	api.GET("/users/me", httphandler.GetProfile(profileUC))
	api.PATCH("/users/me", httphandler.UpdateProfile(profileUC))
	api.POST("/users/me/driver", httphandler.CreateDriverProfile(profileUC))

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
	slog.Info("user service stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
