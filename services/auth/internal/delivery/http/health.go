// Package http — REST delivery for auth service (Echo)
package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Health — liveness probe (no dependencies)
func Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok", "service": "auth"})
}
