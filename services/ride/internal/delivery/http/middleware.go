package http

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/ridehail/ride/internal/infra/jwt"
)

const UserIDKey = "user_id"
const UserRoleKey = "user_role"

type JWTValidator interface {
	Validate(tokenString string) (*jwt.Claims, error)
}

func JWTAuth(v JWTValidator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing Authorization header"})
			}
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid Authorization header"})
			}
			claims, err := v.Validate(parts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
			}
			c.Set(UserIDKey, claims.UserID)
			c.Set(UserRoleKey, claims.Role)
			return next(c)
		}
	}
}
