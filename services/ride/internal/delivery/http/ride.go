package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/ridehail/ride/internal/domain"
	"github.com/ridehail/ride/internal/usecase"
)

type RideUseCase interface {
	CreateRide(ctx context.Context, passengerID string, from, to domain.Point) (*domain.Ride, error)
	GetRide(ctx context.Context, id string) (*domain.Ride, error)
	PlaceBid(ctx context.Context, rideID, driverID string, price float64) (*domain.Bid, error)
	ListBids(ctx context.Context, rideID string) ([]*domain.Bid, error)
	AcceptBid(ctx context.Context, rideID, bidID, passengerID string) (*domain.Ride, error)
	UpdateStatus(ctx context.Context, rideID, status, userID, userRole string) (*domain.Ride, error)
	ListRidesByPassenger(ctx context.Context, passengerID string, limit int) ([]*domain.Ride, error)
	ListRidesByDriver(ctx context.Context, driverID string, limit int) ([]*domain.Ride, error)
	ListOpenRides(ctx context.Context, limit int) ([]*domain.Ride, error)
	ListAllRides(ctx context.Context, limit int) ([]*domain.Ride, error)
}

// CreateRideRequest — POST /api/v1/rides
type CreateRideRequest struct {
	From domain.Point `json:"from"`
	To   domain.Point `json:"to"`
}

func CreateRide(uc RideUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get(UserIDKey).(string)
		var req CreateRideRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
		}
		ride, err := uc.CreateRide(c.Request().Context(), userID, req.From, req.To)
		if err != nil {
			if err == usecase.ErrInvalidStatus {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid coordinates"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create ride"})
		}
		return c.JSON(http.StatusCreated, ride)
	}
}

func GetRide(uc RideUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		rideID := c.Param("id")
		ride, err := uc.GetRide(c.Request().Context(), rideID)
		if err != nil || ride == nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "ride not found"})
		}
		return c.JSON(http.StatusOK, ride)
	}
}

// PlaceBidRequest — POST /api/v1/rides/:id/bids
type PlaceBidRequest struct {
	Price float64 `json:"price"`
}

func PlaceBid(uc RideUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		rideID := c.Param("id")
		driverID := c.Get(UserIDKey).(string)
		var req PlaceBidRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
		}
		bid, err := uc.PlaceBid(c.Request().Context(), rideID, driverID, req.Price)
		if err != nil {
			if err == usecase.ErrRideNotFound {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "ride not found"})
			}
			if err == usecase.ErrRideNotBidding {
				return c.JSON(http.StatusConflict, map[string]string{"error": "ride is not accepting bids"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to place bid"})
		}
		return c.JSON(http.StatusCreated, bid)
	}
}

func ListBids(uc RideUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		rideID := c.Param("id")
		bids, err := uc.ListBids(c.Request().Context(), rideID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list bids"})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"bids": bids})
	}
}

// AcceptBidRequest — POST /api/v1/rides/:id/accept
type AcceptBidRequest struct {
	BidID string `json:"bid_id"`
}

func AcceptBid(uc RideUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		rideID := c.Param("id")
		passengerID := c.Get(UserIDKey).(string)
		var req AcceptBidRequest
		if err := c.Bind(&req); err != nil || req.BidID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "bid_id required"})
		}
		ride, err := uc.AcceptBid(c.Request().Context(), rideID, req.BidID, passengerID)
		if err != nil {
			if err == usecase.ErrRideNotFound || err == usecase.ErrBidNotFound {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "ride or bid not found"})
			}
			if err == usecase.ErrNotPassenger {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "not the ride passenger"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to accept bid"})
		}
		return c.JSON(http.StatusOK, ride)
	}
}

// UpdateStatusRequest — PATCH /api/v1/rides/:id/status
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

func UpdateRideStatus(uc RideUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		rideID := c.Param("id")
		userID := c.Get(UserIDKey).(string)
		userRole := c.Get(UserRoleKey).(string)
		var req UpdateStatusRequest
		if err := c.Bind(&req); err != nil || req.Status == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "status required"})
		}
		ride, err := uc.UpdateStatus(c.Request().Context(), rideID, req.Status, userID, userRole)
		if err != nil {
			if err == usecase.ErrRideNotFound {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "ride not found"})
			}
			if err == usecase.ErrNotPassenger || err == usecase.ErrNotDriver {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
			}
			if err == usecase.ErrInvalidStatus {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid status"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update status"})
		}
		return c.JSON(http.StatusOK, ride)
	}
}

func ListMyRides(uc RideUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get(UserIDKey).(string)
		userRole := c.Get(UserRoleKey).(string)
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		var rides []*domain.Ride
		var err error
		if userRole == "driver" {
			rides, err = uc.ListRidesByDriver(c.Request().Context(), userID, limit)
		} else {
			rides, err = uc.ListRidesByPassenger(c.Request().Context(), userID, limit)
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list rides"})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"rides": rides})
	}
}

// ListAvailableRides — GET /api/v1/rides/available (driver only: open rides to bid)
func ListAvailableRides(uc RideUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		userRole := c.Get(UserRoleKey).(string)
		if userRole != "driver" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "driver only"})
		}
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		rides, err := uc.ListOpenRides(c.Request().Context(), limit)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list available rides"})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"rides": rides})
	}
}

// ListAllRides — GET /api/v1/admin/rides (admin only: all rides for dashboard)
func ListAllRides(uc RideUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		userRole := c.Get(UserRoleKey).(string)
		if userRole != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "admin only"})
		}
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		rides, err := uc.ListAllRides(c.Request().Context(), limit)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list rides"})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"rides": rides})
	}
}
