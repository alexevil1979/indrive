// Package usecase — Driver verification use cases
package usecase

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/ridehail/user/internal/domain"
)

// VerificationRepository interface for verification persistence.
type VerificationRepository interface {
	CreateVerification(ctx context.Context, v *domain.DriverVerification) error
	GetVerification(ctx context.Context, userID string) (*domain.DriverVerification, error)
	GetVerificationByID(ctx context.Context, id string) (*domain.DriverVerification, error)
	UpdateVerificationStatus(ctx context.Context, id, status, reviewerID, rejectReason string) error
	ListPendingVerifications(ctx context.Context, limit, offset int) ([]*domain.DriverVerification, error)

	CreateDocument(ctx context.Context, d *domain.DriverDocument) error
	GetDocument(ctx context.Context, id string) (*domain.DriverDocument, error)
	ListDocumentsByUser(ctx context.Context, userID string) ([]*domain.DriverDocument, error)
	UpdateDocumentStatus(ctx context.Context, id, status, reviewerID, rejectReason string) error
	DeleteDocument(ctx context.Context, id string) error
	GetDocumentByType(ctx context.Context, userID, docType string) (*domain.DriverDocument, error)
	SetUserVerified(ctx context.Context, userID string, verified bool, verificationID string) error
}

// StorageClient interface for file storage.
type StorageClient interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (*UploadResult, error)
	Delete(ctx context.Context, key string) error
	GetPresignedURL(ctx context.Context, key string, expires interface{}) (string, error)
}

// UploadResult from storage.
type UploadResult struct {
	Key         string
	Size        int64
	ContentType string
	URL         string
}

// VerificationUseCase handles driver verification logic.
type VerificationUseCase struct {
	repo    VerificationRepository
	storage StorageClient
}

// NewVerificationUseCase creates verification use case.
func NewVerificationUseCase(repo VerificationRepository, storage StorageClient) *VerificationUseCase {
	return &VerificationUseCase{
		repo:    repo,
		storage: storage,
	}
}

// StartVerification creates a new verification request.
func (uc *VerificationUseCase) StartVerification(ctx context.Context, userID, licenseNumber, vehicleModel, vehiclePlate string, vehicleYear int) (*domain.DriverVerification, error) {
	// Check if already has pending/approved verification
	existing, err := uc.repo.GetVerification(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if existing.Status == domain.VerificationStatusPending {
			return nil, domain.ErrVerificationPending
		}
		if existing.Status == domain.VerificationStatusApproved {
			return nil, domain.ErrAlreadyVerified
		}
		// If rejected, allow resubmission
	}

	v := &domain.DriverVerification{
		UserID:        userID,
		Status:        domain.VerificationStatusPending,
		LicenseNumber: licenseNumber,
		VehicleModel:  vehicleModel,
		VehiclePlate:  vehiclePlate,
		VehicleYear:   vehicleYear,
	}

	if err := uc.repo.CreateVerification(ctx, v); err != nil {
		return nil, err
	}

	return uc.repo.GetVerification(ctx, userID)
}

// UploadDocumentInput for document upload.
type UploadDocumentInput struct {
	UserID      string
	DocType     string
	FileName    string
	FileSize    int64
	ContentType string
	Reader      io.Reader
}

// UploadDocument uploads a verification document.
func (uc *VerificationUseCase) UploadDocument(ctx context.Context, input UploadDocumentInput) (*domain.DriverDocument, error) {
	// Validate document type
	if !domain.IsValidDocType(input.DocType) {
		return nil, domain.ErrInvalidDocumentType
	}

	// Validate file size (max 10MB)
	if input.FileSize > 10*1024*1024 {
		return nil, fmt.Errorf("file too large, max 10MB")
	}

	// Validate content type
	if !isValidContentType(input.ContentType) {
		return nil, fmt.Errorf("invalid file type, allowed: jpg, png, pdf")
	}

	// Generate storage key
	key := generateStorageKey(input.UserID, input.DocType, input.FileName)

	// Upload to storage
	result, err := uc.storage.Upload(ctx, key, input.Reader, input.FileSize, input.ContentType)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	// Create document record
	doc := &domain.DriverDocument{
		UserID:      input.UserID,
		DocType:     input.DocType,
		FileName:    input.FileName,
		FileSize:    result.Size,
		ContentType: input.ContentType,
		StorageKey:  result.Key,
		StorageURL:  result.URL,
		Status:      domain.VerificationStatusPending,
	}

	if err := uc.repo.CreateDocument(ctx, doc); err != nil {
		// Cleanup uploaded file
		_ = uc.storage.Delete(ctx, key)
		return nil, err
	}

	return doc, nil
}

// GetVerificationStatus returns current verification status for user.
func (uc *VerificationUseCase) GetVerificationStatus(ctx context.Context, userID string) (*domain.DriverVerification, []*domain.DriverDocument, error) {
	v, err := uc.repo.GetVerification(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	docs, err := uc.repo.ListDocumentsByUser(ctx, userID)
	if err != nil {
		return v, nil, err
	}

	if v != nil {
		v.Documents = docs
	}

	return v, docs, nil
}

// GetDocument returns document with presigned URL for download.
func (uc *VerificationUseCase) GetDocument(ctx context.Context, userID, docID string) (*domain.DriverDocument, string, error) {
	doc, err := uc.repo.GetDocument(ctx, docID)
	if err != nil {
		return nil, "", err
	}

	// Verify ownership (unless admin)
	if doc.UserID != userID {
		return nil, "", domain.ErrDocumentNotFound
	}

	// Generate presigned URL for download
	// presignedURL, err := uc.storage.GetPresignedURL(ctx, doc.StorageKey, 15*time.Minute)
	// For now, return storage URL directly
	return doc, doc.StorageURL, nil
}

// DeleteDocument removes a document.
func (uc *VerificationUseCase) DeleteDocument(ctx context.Context, userID, docID string) error {
	doc, err := uc.repo.GetDocument(ctx, docID)
	if err != nil {
		return err
	}

	// Verify ownership
	if doc.UserID != userID {
		return domain.ErrDocumentNotFound
	}

	// Only allow deletion of pending documents
	if doc.Status != domain.VerificationStatusPending {
		return fmt.Errorf("cannot delete reviewed document")
	}

	// Delete from storage
	if err := uc.storage.Delete(ctx, doc.StorageKey); err != nil {
		// Log but continue with DB deletion
	}

	return uc.repo.DeleteDocument(ctx, docID)
}

// --- Admin methods ---

// ListPendingVerifications returns verifications awaiting review (admin).
func (uc *VerificationUseCase) ListPendingVerifications(ctx context.Context, limit, offset int) ([]*domain.DriverVerification, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return uc.repo.ListPendingVerifications(ctx, limit, offset)
}

// AdminReviewInput for admin review.
type AdminReviewInput struct {
	VerificationID string
	AdminUserID    string
	Approved       bool
	RejectReason   string
}

// AdminReviewVerification approves or rejects verification (admin).
func (uc *VerificationUseCase) AdminReviewVerification(ctx context.Context, input AdminReviewInput) error {
	v, err := uc.repo.GetVerificationByID(ctx, input.VerificationID)
	if err != nil {
		return err
	}

	if v.Status != domain.VerificationStatusPending {
		return fmt.Errorf("verification already reviewed")
	}

	status := domain.VerificationStatusRejected
	if input.Approved {
		status = domain.VerificationStatusApproved
	}

	// Update verification status
	if err := uc.repo.UpdateVerificationStatus(ctx, input.VerificationID, status, input.AdminUserID, input.RejectReason); err != nil {
		return err
	}

	// Update user's verified status
	if input.Approved {
		if err := uc.repo.SetUserVerified(ctx, v.UserID, true, input.VerificationID); err != nil {
			return err
		}
	}

	return nil
}

// AdminReviewDocumentInput for document review.
type AdminReviewDocumentInput struct {
	DocumentID   string
	AdminUserID  string
	Approved     bool
	RejectReason string
}

// AdminReviewDocument approves or rejects a specific document (admin).
func (uc *VerificationUseCase) AdminReviewDocument(ctx context.Context, input AdminReviewDocumentInput) error {
	doc, err := uc.repo.GetDocument(ctx, input.DocumentID)
	if err != nil {
		return err
	}

	if doc.Status != domain.VerificationStatusPending {
		return fmt.Errorf("document already reviewed")
	}

	status := domain.VerificationStatusRejected
	if input.Approved {
		status = domain.VerificationStatusApproved
	}

	return uc.repo.UpdateDocumentStatus(ctx, input.DocumentID, status, input.AdminUserID, input.RejectReason)
}

// GetVerificationByID returns verification by ID (admin).
func (uc *VerificationUseCase) GetVerificationByID(ctx context.Context, id string) (*domain.DriverVerification, error) {
	v, err := uc.repo.GetVerificationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	docs, err := uc.repo.ListDocumentsByUser(ctx, v.UserID)
	if err != nil {
		return v, nil
	}
	v.Documents = docs

	return v, nil
}

// Helpers

func isValidContentType(ct string) bool {
	ct = strings.ToLower(ct)
	validTypes := []string{
		"image/jpeg", "image/jpg", "image/png", "image/webp",
		"application/pdf",
	}
	for _, vt := range validTypes {
		if ct == vt {
			return true
		}
	}
	return false
}

func generateStorageKey(userID, docType, filename string) string {
	// Extract extension
	ext := ""
	if idx := strings.LastIndex(filename, "."); idx >= 0 {
		ext = filename[idx:]
	}
	return fmt.Sprintf("drivers/%s/%s%s", userID, docType, ext)
}
