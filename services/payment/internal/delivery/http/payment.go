package http

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ridehail/payment/internal/domain"
	"github.com/ridehail/payment/internal/usecase"
)

type PaymentUseCase interface {
	CreatePayment(ctx context.Context, rideID string, amount float64, method string) (*domain.Payment, error)
	GetByRideID(ctx context.Context, rideID string) (*domain.Payment, error)
	GetByID(ctx context.Context, id string) (*domain.Payment, error)
	ConfirmPayment(ctx context.Context, id string) (*domain.Payment, error)
}

// CreatePaymentRequest — POST /api/v1/payments (checkout)
type CreatePaymentRequest struct {
	RideID string  `json:"ride_id"`
	Amount float64 `json:"amount"`
	Method string  `json:"method"` // cash | card
}

func CreatePayment(uc PaymentUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req CreatePaymentRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
		}
		if req.RideID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "ride_id required"})
		}
		if req.Method == "" {
			req.Method = domain.MethodCash
		}
		p, err := uc.CreatePayment(c.Request().Context(), req.RideID, req.Amount, req.Method)
		if err != nil {
			if err == usecase.ErrInvalidAmount || err == usecase.ErrInvalidMethod {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			}
			if err == usecase.ErrPaymentExists {
				return c.JSON(http.StatusConflict, map[string]string{"error": "payment already exists for this ride"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create payment"})
		}
		return c.JSON(http.StatusCreated, p)
	}
}

func GetPaymentByRide(uc PaymentUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		rideID := c.Param("rideId")
		p, err := uc.GetByRideID(c.Request().Context(), rideID)
		if err != nil || p == nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "payment not found"})
		}
		return c.JSON(http.StatusOK, p)
	}
}

func GetPayment(uc PaymentUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		p, err := uc.GetByID(c.Request().Context(), id)
		if err != nil || p == nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "payment not found"})
		}
		return c.JSON(http.StatusOK, p)
	}
}

func ConfirmPayment(uc PaymentUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		p, err := uc.ConfirmPayment(c.Request().Context(), id)
		if err != nil {
			if err == usecase.ErrPaymentNotFound {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "payment not found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to confirm"})
		}
		return c.JSON(http.StatusOK, p)
	}
}
