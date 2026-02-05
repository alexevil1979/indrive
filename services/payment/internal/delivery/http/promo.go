package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/alexevil1979/indrive/services/payment/internal/domain"
	"github.com/alexevil1979/indrive/services/payment/internal/usecase"
)

// PromoHandler handles promo HTTP endpoints
type PromoHandler struct {
	uc *usecase.PromoUseCase
}

// NewPromoHandler creates a new promo handler
func NewPromoHandler(uc *usecase.PromoUseCase) *PromoHandler {
	return &PromoHandler{uc: uc}
}

// ValidatePromoRequest is the request for validating a promo code
type ValidatePromoRequest struct {
	Code        string  `json:"code"`
	OrderAmount float64 `json:"order_amount"`
}

// ValidatePromo handles POST /api/v1/promos/validate
func (h *PromoHandler) ValidatePromo(c echo.Context) error {
	userID, ok := c.Get(UserIDKey).(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, errorResponse("unauthorized"))
	}

	var req ValidatePromoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("invalid request"))
	}

	result, err := h.uc.ValidatePromo(c.Request().Context(), req.Code, userID, req.OrderAmount)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusOK, result)
}

// ApplyPromoRequest is the request for applying a promo code
type ApplyPromoRequest struct {
	Code        string  `json:"code"`
	RideID      string  `json:"ride_id"`
	OrderAmount float64 `json:"order_amount"`
}

// ApplyPromo handles POST /api/v1/promos/apply
func (h *PromoHandler) ApplyPromo(c echo.Context) error {
	userID, ok := c.Get(UserIDKey).(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, errorResponse("unauthorized"))
	}

	var req ApplyPromoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("invalid request"))
	}

	result, err := h.uc.ApplyPromo(c.Request().Context(), req.Code, userID, req.RideID, req.OrderAmount)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusOK, result)
}

// GetMyPromos handles GET /api/v1/promos/my
func (h *PromoHandler) GetMyPromos(c echo.Context) error {
	userID, ok := c.Get(UserIDKey).(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, errorResponse("unauthorized"))
	}

	limit := 20
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	promos, err := h.uc.GetUserPromos(c.Request().Context(), userID, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"promos": promos})
}

// ============ Admin endpoints ============

// CreatePromoRequest is the request for creating a promo
type CreatePromoRequest struct {
	Code          string  `json:"code"`
	Description   string  `json:"description"`
	Type          string  `json:"type"` // percent or fixed
	Value         float64 `json:"value"`
	MinOrderValue float64 `json:"min_order_value"`
	MaxDiscount   float64 `json:"max_discount"`
	UsageLimit    int     `json:"usage_limit"`
	PerUserLimit  int     `json:"per_user_limit"`
	StartsAt      string  `json:"starts_at,omitempty"`  // RFC3339
	ExpiresAt     string  `json:"expires_at,omitempty"` // RFC3339
}

// CreatePromo handles POST /api/v1/admin/promos
func (h *PromoHandler) CreatePromo(c echo.Context) error {
	role, ok := c.Get(UserRoleKey).(string)
	if !ok || role != "admin" {
		return c.JSON(http.StatusForbidden, errorResponse("admin access required"))
	}

	var req CreatePromoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("invalid request"))
	}

	promo := &domain.Promo{
		Code:          req.Code,
		Description:   req.Description,
		Type:          domain.PromoType(req.Type),
		Value:         req.Value,
		MinOrderValue: req.MinOrderValue,
		MaxDiscount:   req.MaxDiscount,
		UsageLimit:    req.UsageLimit,
		PerUserLimit:  req.PerUserLimit,
		IsActive:      true,
		StartsAt:      time.Now(),
	}

	if req.StartsAt != "" {
		if t, err := time.Parse(time.RFC3339, req.StartsAt); err == nil {
			promo.StartsAt = t
		}
	}
	if req.ExpiresAt != "" {
		if t, err := time.Parse(time.RFC3339, req.ExpiresAt); err == nil {
			promo.ExpiresAt = t
		}
	}

	if err := h.uc.CreatePromo(c.Request().Context(), promo); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusCreated, promo)
}

// UpdatePromoRequest is the request for updating a promo
type UpdatePromoRequest struct {
	Description   *string  `json:"description,omitempty"`
	Type          *string  `json:"type,omitempty"`
	Value         *float64 `json:"value,omitempty"`
	MinOrderValue *float64 `json:"min_order_value,omitempty"`
	MaxDiscount   *float64 `json:"max_discount,omitempty"`
	UsageLimit    *int     `json:"usage_limit,omitempty"`
	PerUserLimit  *int     `json:"per_user_limit,omitempty"`
	IsActive      *bool    `json:"is_active,omitempty"`
	StartsAt      *string  `json:"starts_at,omitempty"`
	ExpiresAt     *string  `json:"expires_at,omitempty"`
}

// UpdatePromo handles PUT /api/v1/admin/promos/:id
func (h *PromoHandler) UpdatePromo(c echo.Context) error {
	role, ok := c.Get(UserRoleKey).(string)
	if !ok || role != "admin" {
		return c.JSON(http.StatusForbidden, errorResponse("admin access required"))
	}

	id := c.Param("id")
	promo, err := h.uc.GetPromo(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, errorResponse("promo not found"))
	}

	var req UpdatePromoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("invalid request"))
	}

	// Apply updates
	if req.Description != nil {
		promo.Description = *req.Description
	}
	if req.Type != nil {
		promo.Type = domain.PromoType(*req.Type)
	}
	if req.Value != nil {
		promo.Value = *req.Value
	}
	if req.MinOrderValue != nil {
		promo.MinOrderValue = *req.MinOrderValue
	}
	if req.MaxDiscount != nil {
		promo.MaxDiscount = *req.MaxDiscount
	}
	if req.UsageLimit != nil {
		promo.UsageLimit = *req.UsageLimit
	}
	if req.PerUserLimit != nil {
		promo.PerUserLimit = *req.PerUserLimit
	}
	if req.IsActive != nil {
		promo.IsActive = *req.IsActive
	}
	if req.StartsAt != nil {
		if t, err := time.Parse(time.RFC3339, *req.StartsAt); err == nil {
			promo.StartsAt = t
		}
	}
	if req.ExpiresAt != nil {
		if t, err := time.Parse(time.RFC3339, *req.ExpiresAt); err == nil {
			promo.ExpiresAt = t
		}
	}

	if err := h.uc.UpdatePromo(c.Request().Context(), promo); err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusOK, promo)
}

// DeletePromo handles DELETE /api/v1/admin/promos/:id
func (h *PromoHandler) DeletePromo(c echo.Context) error {
	role, ok := c.Get(UserRoleKey).(string)
	if !ok || role != "admin" {
		return c.JSON(http.StatusForbidden, errorResponse("admin access required"))
	}

	id := c.Param("id")
	if err := h.uc.DeletePromo(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}

// GetPromo handles GET /api/v1/admin/promos/:id
func (h *PromoHandler) GetPromo(c echo.Context) error {
	role, ok := c.Get(UserRoleKey).(string)
	if !ok || role != "admin" {
		return c.JSON(http.StatusForbidden, errorResponse("admin access required"))
	}

	id := c.Param("id")
	promo, err := h.uc.GetPromo(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, errorResponse("promo not found"))
	}

	return c.JSON(http.StatusOK, promo)
}

// ListPromos handles GET /api/v1/admin/promos
func (h *PromoHandler) ListPromos(c echo.Context) error {
	role, ok := c.Get(UserRoleKey).(string)
	if !ok || role != "admin" {
		return c.JSON(http.StatusForbidden, errorResponse("admin access required"))
	}

	limit := 20
	offset := 0
	activeOnly := false

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}
	if a := c.QueryParam("active_only"); a == "true" {
		activeOnly = true
	}

	promos, total, err := h.uc.ListPromos(c.Request().Context(), limit, offset, activeOnly)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"promos": promos,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// ListActivePromos handles GET /api/v1/promos (public list of active promos)
func (h *PromoHandler) ListActivePromos(c echo.Context) error {
	promos, _, err := h.uc.ListPromos(c.Request().Context(), 50, 0, true)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	// Return only public fields
	type PublicPromo struct {
		Code          string  `json:"code"`
		Description   string  `json:"description"`
		Type          string  `json:"type"`
		Value         float64 `json:"value"`
		MinOrderValue float64 `json:"min_order_value"`
		MaxDiscount   float64 `json:"max_discount"`
	}

	public := make([]PublicPromo, 0, len(promos))
	for _, p := range promos {
		if p.IsValid() {
			public = append(public, PublicPromo{
				Code:          p.Code,
				Description:   p.Description,
				Type:          string(p.Type),
				Value:         p.Value,
				MinOrderValue: p.MinOrderValue,
				MaxDiscount:   p.MaxDiscount,
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"promos": public})
}
