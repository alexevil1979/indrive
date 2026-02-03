// Package http — Auth REST handlers: register, login, refresh, OAuth stubs
package http

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/ridehail/auth/internal/domain"
	"github.com/ridehail/auth/internal/usecase"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// RegisterRequest — POST /auth/register
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // passenger | driver, default passenger
}

// LoginRequest — POST /auth/login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest — POST /auth/refresh
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// TokenResponse — access + refresh
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	UserID       string `json:"user_id,omitempty"`
	Role         string `json:"role,omitempty"`
}

func Register(uc *usecase.AuthUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req RegisterRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
		}
		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		if !emailRegex.MatchString(req.Email) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid email"})
		}
		if len(req.Password) < 8 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		}
		if req.Role == "" {
			req.Role = "passenger"
		}
		u, tp, err := uc.Register(c.Request().Context(), req.Email, req.Password, req.Role)
		if err != nil {
			if err == domain.ErrEmailTaken {
				return c.JSON(http.StatusConflict, map[string]string{"error": "email already taken"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "registration failed"})
		}
		return c.JSON(http.StatusCreated, TokenResponse{
			AccessToken:  tp.AccessToken,
			RefreshToken: tp.RefreshToken,
			ExpiresIn:    tp.ExpiresIn,
			UserID:       u.ID,
			Role:         u.Role,
		})
	}
}

func Login(uc *usecase.AuthUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req LoginRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
		}
		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		if req.Email == "" || req.Password == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "email and password required"})
		}
		tp, err := uc.Login(c.Request().Context(), req.Email, req.Password)
		if err != nil {
			if err == usecase.ErrInvalidCredentials {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "login failed"})
		}
		return c.JSON(http.StatusOK, TokenResponse{
			AccessToken:  tp.AccessToken,
			RefreshToken: tp.RefreshToken,
			ExpiresIn:    tp.ExpiresIn,
		})
	}
}

func Refresh(uc *usecase.AuthUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req RefreshRequest
		if err := c.Bind(&req); err != nil || req.RefreshToken == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "refresh_token required"})
		}
		tp, err := uc.Refresh(c.Request().Context(), req.RefreshToken)
		if err != nil {
			if err == usecase.ErrInvalidCredentials {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired refresh token"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "refresh failed"})
		}
		return c.JSON(http.StatusOK, TokenResponse{
			AccessToken:  tp.AccessToken,
			RefreshToken: tp.RefreshToken,
			ExpiresIn:    tp.ExpiresIn,
		})
	}
}

// OAuthStub — GET /auth/oauth/:provider (Google/Yandex/VK) — 501 stub
func OAuthStub(provider string) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusNotImplemented, map[string]string{
			"error":   "OAuth not implemented",
			"message": "OAuth2 for " + provider + " will be added later",
		})
	}
}
