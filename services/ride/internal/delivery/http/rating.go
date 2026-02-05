package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/ridehail/ride/internal/domain"
	"github.com/ridehail/ride/internal/usecase"
)

// RatingHandler handles rating HTTP endpoints
type RatingHandler struct {
	uc *usecase.RatingUseCase
}

// NewRatingHandler creates a new rating handler
func NewRatingHandler(uc *usecase.RatingUseCase) *RatingHandler {
	return &RatingHandler{uc: uc}
}

// SubmitRatingRequest is the request body for submitting a rating
type SubmitRatingRequest struct {
	Score   int      `json:"score" validate:"required,min=1,max=5"`
	Comment string   `json:"comment,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

// SubmitRating handles POST /api/v1/rides/:id/rating
func (h *RatingHandler) SubmitRating(c echo.Context) error {
	rideID := c.Param("id")
	userID, ok := c.Get(UserIDKey).(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, errorResponse("unauthorized"))
	}

	var req SubmitRatingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("invalid request body"))
	}

	rating, err := h.uc.SubmitRating(c.Request().Context(), usecase.SubmitRatingInput{
		RideID:     rideID,
		FromUserID: userID,
		Score:      req.Score,
		Comment:    req.Comment,
		Tags:       req.Tags,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusCreated, rating)
}

// GetRideRatings handles GET /api/v1/rides/:id/ratings
func (h *RatingHandler) GetRideRatings(c echo.Context) error {
	rideID := c.Param("id")
	ratings, err := h.uc.GetRideRatings(c.Request().Context(), rideID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"ratings": ratings})
}

// GetUserRating handles GET /api/v1/users/:id/rating
func (h *RatingHandler) GetUserRating(c echo.Context) error {
	userID := c.Param("id")
	role := c.QueryParam("role")
	if role == "" {
		role = "driver" // default to driver ratings
	}

	rating, err := h.uc.GetUserRating(c.Request().Context(), userID, role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}
	return c.JSON(http.StatusOK, rating)
}

// GetUserRatings handles GET /api/v1/users/:id/ratings
func (h *RatingHandler) GetUserRatings(c echo.Context) error {
	userID := c.Param("id")
	role := c.QueryParam("role")
	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	limit := 20
	offset := 0
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	ratings, err := h.uc.GetUserRatings(c.Request().Context(), userID, role, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"ratings": ratings})
}

// GetMyRating handles GET /api/v1/me/rating
func (h *RatingHandler) GetMyRating(c echo.Context) error {
	userID, ok := c.Get(UserIDKey).(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, errorResponse("unauthorized"))
	}

	role := c.QueryParam("role")
	if role == "" {
		role = "driver"
	}

	rating, err := h.uc.GetUserRating(c.Request().Context(), userID, role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}
	return c.JSON(http.StatusOK, rating)
}

// GetMyRatings handles GET /api/v1/me/ratings
func (h *RatingHandler) GetMyRatings(c echo.Context) error {
	userID, ok := c.Get(UserIDKey).(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, errorResponse("unauthorized"))
	}

	role := c.QueryParam("role")
	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	limit := 20
	offset := 0
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	ratings, err := h.uc.GetUserRatings(c.Request().Context(), userID, role, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"ratings": ratings})
}

// GetRatingTags handles GET /api/v1/ratings/tags
func (h *RatingHandler) GetRatingTags(c echo.Context) error {
	role := c.QueryParam("role")
	if role == "" {
		role = "driver"
	}

	tags := h.uc.GetRatingTags(role)
	
	// Return tags with labels
	result := make([]map[string]string, len(tags))
	for i, tag := range tags {
		result[i] = map[string]string{
			"id":    tag,
			"label": domain.TagLabels[tag],
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"tags": result})
}

// ListAllRatings handles GET /api/v1/admin/ratings (admin only)
func (h *RatingHandler) ListAllRatings(c echo.Context) error {
	role, ok := c.Get(UserRoleKey).(string)
	if !ok || role != "admin" {
		return c.JSON(http.StatusForbidden, errorResponse("admin access required"))
	}

	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	limit := 20
	offset := 0
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	ratings, total, err := h.uc.ListAllRatings(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"ratings": ratings,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}
