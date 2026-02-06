package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ridehail/user/internal/domain"
	"github.com/ridehail/user/internal/usecase"
)

// SettingsHandler handles settings HTTP requests
type SettingsHandler struct {
	uc *usecase.SettingsUseCase
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(uc *usecase.SettingsUseCase) *SettingsHandler {
	return &SettingsHandler{uc: uc}
}

// GetSettings returns all app settings (admin only)
func (h *SettingsHandler) GetSettings() echo.HandlerFunc {
	return func(c echo.Context) error {
		settings, err := h.uc.GetSettings(c.Request().Context())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, settings)
	}
}

// UpdateSettingsRequest represents the update settings request body
type UpdateSettingsRequest struct {
	MapProvider      string `json:"map_provider"`
	GoogleMapsAPIKey string `json:"google_maps_api_key,omitempty"`
	YandexMapsAPIKey string `json:"yandex_maps_api_key,omitempty"`
	DefaultLanguage  string `json:"default_language,omitempty"`
	DefaultCurrency  string `json:"default_currency,omitempty"`
}

// UpdateSettings updates app settings (admin only)
func (h *SettingsHandler) UpdateSettings() echo.HandlerFunc {
	return func(c echo.Context) error {
		var req UpdateSettingsRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		}

		userID := c.Get(UserIDKey).(string)

		settings := &domain.AppSettings{
			ID:               "default",
			MapProvider:      domain.MapProvider(req.MapProvider),
			GoogleMapsAPIKey: req.GoogleMapsAPIKey,
			YandexMapsAPIKey: req.YandexMapsAPIKey,
			DefaultLanguage:  req.DefaultLanguage,
			DefaultCurrency:  req.DefaultCurrency,
		}

		if settings.DefaultLanguage == "" {
			settings.DefaultLanguage = "ru"
		}
		if settings.DefaultCurrency == "" {
			settings.DefaultCurrency = "RUB"
		}

		if err := h.uc.UpdateSettings(c.Request().Context(), settings, userID); err != nil {
			if err == usecase.ErrInvalidMapProvider {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "updated"})
	}
}

// GetMapSettings returns map settings for mobile apps (public endpoint)
func (h *SettingsHandler) GetMapSettings() echo.HandlerFunc {
	return func(c echo.Context) error {
		settings, err := h.uc.GetMapSettings(c.Request().Context())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, settings)
	}
}
