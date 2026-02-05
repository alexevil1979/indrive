package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/ridehail/payment/internal/domain"
	"github.com/ridehail/payment/internal/usecase"
)

// PaymentUseCase interface
type PaymentUseCase interface {
	CreatePayment(ctx context.Context, input usecase.CreatePaymentInput) (*domain.PaymentIntent, error)
	GetByRideID(ctx context.Context, rideID string) (*domain.Payment, error)
	GetByID(ctx context.Context, id string) (*domain.Payment, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Payment, error)
	ConfirmPayment(ctx context.Context, id string) (*domain.Payment, error)
	ProcessWebhook(ctx context.Context, provider string, body []byte, signature string) error
	RefundPayment(ctx context.Context, req domain.RefundRequest) (*domain.RefundResult, error)
	ListPaymentMethods(ctx context.Context, userID string) ([]*domain.PaymentMethod, error)
	DeletePaymentMethod(ctx context.Context, id, userID string) error
	SetDefaultPaymentMethod(ctx context.Context, id, userID string) error
	GetAvailableProviders() []string
}

// CreatePaymentRequest — POST /api/v1/payments (checkout)
type CreatePaymentRequest struct {
	RideID      string  `json:"ride_id"`
	Amount      float64 `json:"amount"`
	Method      string  `json:"method"`      // cash | card
	Provider    string  `json:"provider"`    // cash | tinkoff | yoomoney | sber
	Description string  `json:"description"`
	ReturnURL   string  `json:"return_url"`
	SaveCard    bool    `json:"save_card"`
	TokenID     string  `json:"token_id"` // Use saved card
	UserEmail   string  `json:"user_email"`
	UserPhone   string  `json:"user_phone"`
}

// CreatePayment creates a new payment
func CreatePayment(uc PaymentUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get(UserIDKey).(string)

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

		input := usecase.CreatePaymentInput{
			RideID:      req.RideID,
			UserID:      userID,
			Amount:      req.Amount,
			Method:      req.Method,
			Provider:    req.Provider,
			Description: req.Description,
			ReturnURL:   req.ReturnURL,
			SaveCard:    req.SaveCard,
			TokenID:     req.TokenID,
			UserEmail:   req.UserEmail,
			UserPhone:   req.UserPhone,
		}

		intent, err := uc.CreatePayment(c.Request().Context(), input)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidAmount) || errors.Is(err, domain.ErrInvalidMethod) || errors.Is(err, domain.ErrInvalidProvider) {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			}
			if errors.Is(err, domain.ErrPaymentExists) {
				return c.JSON(http.StatusConflict, map[string]string{"error": "payment already exists for this ride"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusCreated, intent)
	}
}

// GetPayment returns payment by ID
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

// GetPaymentByRide returns payment by ride ID
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

// ListPayments returns user's payments
func ListPayments(uc PaymentUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get(UserIDKey).(string)
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		offset, _ := strconv.Atoi(c.QueryParam("offset"))

		payments, err := uc.ListByUser(c.Request().Context(), userID, limit, offset)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list payments"})
		}
		return c.JSON(http.StatusOK, payments)
	}
}

// ConfirmPayment confirms cash payment
func ConfirmPayment(uc PaymentUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		p, err := uc.ConfirmPayment(c.Request().Context(), id)
		if err != nil {
			if errors.Is(err, domain.ErrPaymentNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "payment not found"})
			}
			if errors.Is(err, domain.ErrPaymentNotPending) {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "payment is not pending"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to confirm"})
		}
		return c.JSON(http.StatusOK, p)
	}
}

// RefundRequest — POST /api/v1/payments/:id/refund
type RefundRequest struct {
	Amount float64 `json:"amount"` // 0 = full refund
	Reason string  `json:"reason"`
}

// RefundPayment processes a refund
func RefundPayment(uc PaymentUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		var req RefundRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
		}

		result, err := uc.RefundPayment(c.Request().Context(), domain.RefundRequest{
			PaymentID: id,
			Amount:    req.Amount,
			Reason:    req.Reason,
		})
		if err != nil {
			if errors.Is(err, domain.ErrPaymentNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "payment not found"})
			}
			if errors.Is(err, domain.ErrRefundNotAllowed) {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "refund not allowed"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, result)
	}
}

// --- Webhooks ---

// Webhook handles payment provider callbacks
func Webhook(uc PaymentUseCase, provider string) echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to read body"})
		}

		// Get signature header (varies by provider)
		signature := ""
		switch provider {
		case domain.ProviderTinkoff:
			// Tinkoff uses Token in body
		case domain.ProviderYooMoney:
			signature = c.Request().Header.Get("X-YooKassa-Signature")
		case domain.ProviderSber:
			signature = c.Request().Header.Get("X-Signature")
		}

		if err := uc.ProcessWebhook(c.Request().Context(), provider, body, signature); err != nil {
			// Log error but return OK to prevent retries
			c.Logger().Error("webhook processing error", "provider", provider, "error", err)
		}

		// Always return OK for webhooks
		return c.String(http.StatusOK, "OK")
	}
}

// --- Payment Methods ---

// ListPaymentMethods returns saved cards
func ListPaymentMethods(uc PaymentUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get(UserIDKey).(string)
		methods, err := uc.ListPaymentMethods(c.Request().Context(), userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list methods"})
		}
		return c.JSON(http.StatusOK, methods)
	}
}

// DeletePaymentMethod removes a saved card
func DeletePaymentMethod(uc PaymentUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get(UserIDKey).(string)
		id := c.Param("id")
		if err := uc.DeletePaymentMethod(c.Request().Context(), id, userID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
	}
}

// SetDefaultPaymentMethod sets a card as default
func SetDefaultPaymentMethod(uc PaymentUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get(UserIDKey).(string)
		id := c.Param("id")
		if err := uc.SetDefaultPaymentMethod(c.Request().Context(), id, userID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to set default"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "updated"})
	}
}

// --- Providers ---

// GetProviders returns available payment providers
func GetProviders(uc PaymentUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		providers := uc.GetAvailableProviders()
		return c.JSON(http.StatusOK, map[string][]string{"providers": providers})
	}
}
