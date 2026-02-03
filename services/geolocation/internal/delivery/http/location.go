package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/ridehail/geolocation/internal/domain"
	"github.com/ridehail/geolocation/internal/usecase"
)

type LocationUseCase interface {
	UpdateDriverLocation(ctx context.Context, driverID string, lat, lng float64) error
	FindNearestDrivers(ctx context.Context, q domain.NearestQuery) ([]domain.DriverLocation, error)
}

// UpdateLocationRequest — POST /api/v1/drivers/:id/location
type UpdateLocationRequest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func UpdateDriverLocation(uc LocationUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		driverID := c.Param("id")
		if driverID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "driver id required"})
		}
		var req UpdateLocationRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
		}
		if err := uc.UpdateDriverLocation(c.Request().Context(), driverID, req.Lat, req.Lng); err != nil {
			if err == usecase.ErrInvalidCoordinates {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update location"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}
}

// Nearest — GET /api/v1/drivers/nearest?lat=55.75&lng=37.62&radius_km=5&limit=10
func NearestDrivers(uc LocationUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		lat, _ := strconv.ParseFloat(c.QueryParam("lat"), 64)
		lng, _ := strconv.ParseFloat(c.QueryParam("lng"), 64)
		radiusKm, _ := strconv.ParseFloat(c.QueryParam("radius_km"), 64)
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		if lat == 0 && lng == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "lat and lng required"})
		}
		if radiusKm <= 0 {
			radiusKm = 10
		}
		if limit <= 0 {
			limit = 10
		}
		q := domain.NearestQuery{Lat: lat, Lng: lng, RadiusKM: radiusKm, Limit: limit}
		drivers, err := uc.FindNearestDrivers(c.Request().Context(), q)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "search failed"})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"drivers": drivers})
	}
}
