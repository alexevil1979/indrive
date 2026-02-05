// Package http — Driver verification HTTP handlers
package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/ridehail/user/internal/domain"
	"github.com/ridehail/user/internal/usecase"
)

// VerificationHandler handles verification endpoints.
type VerificationHandler struct {
	uc *usecase.VerificationUseCase
}

// NewVerificationHandler creates verification handler.
func NewVerificationHandler(uc *usecase.VerificationUseCase) *VerificationHandler {
	return &VerificationHandler{uc: uc}
}

// StartVerificationRequest — POST /api/v1/verification
type StartVerificationRequest struct {
	LicenseNumber string `json:"license_number"`
	VehicleModel  string `json:"vehicle_model"`
	VehiclePlate  string `json:"vehicle_plate"`
	VehicleYear   int    `json:"vehicle_year"`
}

// StartVerification creates a new verification request.
// POST /api/v1/verification
func (h *VerificationHandler) StartVerification() echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get(UserIDKey).(string)

		var req StartVerificationRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
		}

		if req.LicenseNumber == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "license_number required"})
		}

		v, err := h.uc.StartVerification(c.Request().Context(), userID, req.LicenseNumber, req.VehicleModel, req.VehiclePlate, req.VehicleYear)
		if err != nil {
			if err == domain.ErrVerificationPending {
				return c.JSON(http.StatusConflict, map[string]string{"error": "verification already pending"})
			}
			if err == domain.ErrAlreadyVerified {
				return c.JSON(http.StatusConflict, map[string]string{"error": "already verified"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to start verification"})
		}

		return c.JSON(http.StatusCreated, v)
	}
}

// GetVerificationStatus returns current verification status.
// GET /api/v1/verification
func (h *VerificationHandler) GetVerificationStatus() echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get(UserIDKey).(string)

		v, docs, err := h.uc.GetVerificationStatus(c.Request().Context(), userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get status"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"verification": v,
			"documents":    docs,
		})
	}
}

// UploadDocument uploads a verification document.
// POST /api/v1/verification/documents
func (h *VerificationHandler) UploadDocument() echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get(UserIDKey).(string)

		// Get document type from form
		docType := c.FormValue("doc_type")
		if docType == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "doc_type required"})
		}
		if !domain.IsValidDocType(docType) {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error":       "invalid doc_type",
				"valid_types": "license, passport, vehicle_reg, insurance, photo, vehicle_photo",
			})
		}

		// Get file from multipart form
		file, err := c.FormFile("file")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "file required"})
		}

		// Open file
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to read file"})
		}
		defer src.Close()

		// Upload document
		input := usecase.UploadDocumentInput{
			UserID:      userID,
			DocType:     docType,
			FileName:    file.Filename,
			FileSize:    file.Size,
			ContentType: file.Header.Get("Content-Type"),
			Reader:      src,
		}

		doc, err := h.uc.UploadDocument(c.Request().Context(), input)
		if err != nil {
			if err == domain.ErrInvalidDocumentType {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusCreated, doc)
	}
}

// ListDocuments returns user's uploaded documents.
// GET /api/v1/verification/documents
func (h *VerificationHandler) ListDocuments() echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get(UserIDKey).(string)

		_, docs, err := h.uc.GetVerificationStatus(c.Request().Context(), userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list documents"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"documents": docs,
		})
	}
}

// GetDocument returns a specific document.
// GET /api/v1/verification/documents/:id
func (h *VerificationHandler) GetDocument() echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get(UserIDKey).(string)
		docID := c.Param("id")

		doc, url, err := h.uc.GetDocument(c.Request().Context(), userID, docID)
		if err != nil {
			if err == domain.ErrDocumentNotFound {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "document not found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get document"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"document":     doc,
			"download_url": url,
		})
	}
}

// DeleteDocument removes a document.
// DELETE /api/v1/verification/documents/:id
func (h *VerificationHandler) DeleteDocument() echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get(UserIDKey).(string)
		docID := c.Param("id")

		if err := h.uc.DeleteDocument(c.Request().Context(), userID, docID); err != nil {
			if err == domain.ErrDocumentNotFound {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "document not found"})
			}
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		return c.NoContent(http.StatusNoContent)
	}
}

// --- Admin endpoints ---

// ListPendingVerifications returns pending verifications (admin only).
// GET /api/v1/admin/verifications
func (h *VerificationHandler) ListPendingVerifications() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Check admin role
		role := c.Get(UserRoleKey)
		if role != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "admin only"})
		}

		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		offset, _ := strconv.Atoi(c.QueryParam("offset"))

		verifications, err := h.uc.ListPendingVerifications(c.Request().Context(), limit, offset)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list verifications"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"verifications": verifications,
		})
	}
}

// GetVerificationByID returns a verification by ID (admin only).
// GET /api/v1/admin/verifications/:id
func (h *VerificationHandler) GetVerificationByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		role := c.Get(UserRoleKey)
		if role != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "admin only"})
		}

		id := c.Param("id")
		v, err := h.uc.GetVerificationByID(c.Request().Context(), id)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "verification not found"})
		}

		return c.JSON(http.StatusOK, v)
	}
}

// ReviewVerificationRequest — POST /api/v1/admin/verifications/:id/review
type ReviewVerificationRequest struct {
	Approved     bool   `json:"approved"`
	RejectReason string `json:"reject_reason"`
}

// ReviewVerification approves or rejects a verification (admin only).
// POST /api/v1/admin/verifications/:id/review
func (h *VerificationHandler) ReviewVerification() echo.HandlerFunc {
	return func(c echo.Context) error {
		role := c.Get(UserRoleKey)
		if role != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "admin only"})
		}

		adminUserID := c.Get(UserIDKey).(string)
		verificationID := c.Param("id")

		var req ReviewVerificationRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
		}

		input := usecase.AdminReviewInput{
			VerificationID: verificationID,
			AdminUserID:    adminUserID,
			Approved:       req.Approved,
			RejectReason:   req.RejectReason,
		}

		if err := h.uc.AdminReviewVerification(c.Request().Context(), input); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "reviewed"})
	}
}

// ReviewDocumentRequest — POST /api/v1/admin/documents/:id/review
type ReviewDocumentRequest struct {
	Approved     bool   `json:"approved"`
	RejectReason string `json:"reject_reason"`
}

// ReviewDocument approves or rejects a document (admin only).
// POST /api/v1/admin/documents/:id/review
func (h *VerificationHandler) ReviewDocument() echo.HandlerFunc {
	return func(c echo.Context) error {
		role := c.Get(UserRoleKey)
		if role != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "admin only"})
		}

		adminUserID := c.Get(UserIDKey).(string)
		documentID := c.Param("id")

		var req ReviewDocumentRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
		}

		input := usecase.AdminReviewDocumentInput{
			DocumentID:   documentID,
			AdminUserID:  adminUserID,
			Approved:     req.Approved,
			RejectReason: req.RejectReason,
		}

		if err := h.uc.AdminReviewDocument(c.Request().Context(), input); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "reviewed"})
	}
}
