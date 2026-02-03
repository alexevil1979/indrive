package http

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ridehail/user/internal/domain"
)

type ProfileUseCase interface {
	GetProfile(ctx context.Context, userID string) (*domain.Profile, *domain.DriverProfile, error)
	UpdateProfile(ctx context.Context, userID string, displayName, phone, avatarURL *string) (*domain.Profile, error)
	CreateDriverProfile(ctx context.Context, userID, licenseNumber string) (*domain.DriverProfile, error)
}

// GetProfile — GET /api/v1/users/me
func GetProfile(uc ProfileUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get(UserIDKey).(string)
		p, d, err := uc.GetProfile(c.Request().Context(), userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get profile"})
		}
		resp := map[string]interface{}{
			"user_id": userID,
			"profile": p,
		}
		if d != nil {
			resp["driver_profile"] = d
		}
		return c.JSON(http.StatusOK, resp)
	}
}

// UpdateProfileRequest — PATCH /api/v1/users/me
type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name"`
	Phone       *string `json:"phone"`
	AvatarURL   *string `json:"avatar_url"`
}

func UpdateProfile(uc ProfileUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get(UserIDKey).(string)
		var req UpdateProfileRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
		}
		p, err := uc.UpdateProfile(c.Request().Context(), userID, req.DisplayName, req.Phone, req.AvatarURL)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update profile"})
		}
		return c.JSON(http.StatusOK, p)
	}
}

// CreateDriverRequest — POST /api/v1/users/me/driver (verification stub)
type CreateDriverRequest struct {
	LicenseNumber string `json:"license_number"`
}

func CreateDriverProfile(uc ProfileUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get(UserIDKey).(string)
		var req CreateDriverRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
		}
		d, err := uc.CreateDriverProfile(c.Request().Context(), userID, req.LicenseNumber)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create driver profile"})
		}
		return c.JSON(http.StatusCreated, d)
	}
}
