package http

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// Ready — readiness probe (checks Postgres)
func Ready(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := pool.Ping(c.Request().Context()); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "not ready", "error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
	}
}
