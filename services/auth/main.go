// Package main — RideHail Auth Service (2026)
// Echo chosen: production-grade, middleware, OpenTelemetry-friendly.
// DDD: delivery (HTTP) -> usecase -> domain; infra (pg, redis) separate.
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

	httphandler "github.com/ridehail/auth/internal/delivery/http"
	"github.com/ridehail/auth/internal/infra/bcrypt"
	"github.com/ridehail/auth/internal/infra/jwt"
	"github.com/ridehail/auth/internal/infra/pg"
	"github.com/ridehail/auth/internal/infra/redis"
	"github.com/ridehail/auth/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := getEnv("PORT", "8080")
	pgDSN := getEnv("PG_DSN", "postgres://ridehail:ridehail_secret@localhost:5432/ridehail?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-in-production")
	redisAddr := getEnv("REDIS_ADDR", "")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pg.Connect(ctx, pgDSN)
	if err != nil {
		slog.Error("postgres connect", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pg.EnsurePostGIS(ctx, pool); err != nil {
		slog.Error("postgis check", "error", err)
		os.Exit(1)
	}
	if err := pg.Migrate(ctx, pool); err != nil {
		slog.Error("migrate", "error", err)
		os.Exit(1)
	}
	slog.Info("postgres + PostGIS + migrations ready")

	rdb, err := redis.New(redisAddr)
	if err != nil {
		slog.Error("redis connect", "error", err)
		os.Exit(1)
	}
	if rdb != nil {
		defer rdb.Close()
		slog.Info("redis ready")
	}

	jwtUtil := jwt.New(jwtSecret, 24*time.Hour, 7*24*time.Hour)
	userRepo := pg.NewUserRepo(pool)
	hasher := bcrypt.New(12)
	authUC := usecase.NewAuthUseCase(userRepo, jwtUtil, hasher)

	e := echo.New()
	e.Use(middleware.Recover(), middleware.Logger(), middleware.RequestID())
	e.GET("/health", httphandler.Health)
	e.GET("/ready", httphandler.Ready(pool))
	e.POST("/auth/register", httphandler.Register(authUC))
	e.POST("/auth/login", httphandler.Login(authUC))
	e.POST("/auth/refresh", httphandler.Refresh(authUC))
	e.GET("/auth/oauth/google", httphandler.OAuthStub("google"))
	e.GET("/auth/oauth/yandex", httphandler.OAuthStub("yandex"))
	e.GET("/auth/oauth/vk", httphandler.OAuthStub("vk"))

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
	slog.Info("auth service stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
