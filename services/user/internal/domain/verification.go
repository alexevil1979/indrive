// Package domain — Driver verification entities
package domain

import (
	"errors"
	"time"
)

// Verification statuses
const (
	VerificationStatusPending  = "pending"
	VerificationStatusApproved = "approved"
	VerificationStatusRejected = "rejected"
)

// Document types
const (
	DocTypeLicense      = "license"       // Driver's license
	DocTypePassport     = "passport"      // Passport/ID
	DocTypeVehicleReg   = "vehicle_reg"   // Vehicle registration
	DocTypeInsurance    = "insurance"     // Insurance document
	DocTypePhoto        = "photo"         // Driver photo
	DocTypeVehiclePhoto = "vehicle_photo" // Vehicle photo
)

var (
	ErrDocumentNotFound    = errors.New("document not found")
	ErrInvalidDocumentType = errors.New("invalid document type")
	ErrVerificationPending = errors.New("verification already pending")
	ErrAlreadyVerified     = errors.New("already verified")
	ErrNotDriver           = errors.New("user is not a driver")
)

// DriverDocument — uploaded document for driver verification
type DriverDocument struct {
	ID          string
	UserID      string
	DocType     string // license, passport, vehicle_reg, insurance, photo, vehicle_photo
	FileName    string
	FileSize    int64
	ContentType string
	StorageKey  string // S3/MinIO object key
	StorageURL  string // Public or presigned URL
	Status      string // pending, approved, rejected
	RejectReason string
	UploadedAt  time.Time
	ReviewedAt  *time.Time
	ReviewedBy  *string // Admin user ID
}

// DriverVerification — verification request aggregate
type DriverVerification struct {
	ID            string
	UserID        string
	Status        string // pending, approved, rejected
	LicenseNumber string
	VehicleModel  string
	VehiclePlate  string
	VehicleYear   int
	Documents     []*DriverDocument
	SubmittedAt   time.Time
	ReviewedAt    *time.Time
	ReviewedBy    *string
	RejectReason  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// IsValidDocType checks if document type is valid
func IsValidDocType(docType string) bool {
	switch docType {
	case DocTypeLicense, DocTypePassport, DocTypeVehicleReg, DocTypeInsurance, DocTypePhoto, DocTypeVehiclePhoto:
		return true
	}
	return false
}

// RequiredDocTypes returns required document types for verification
func RequiredDocTypes() []string {
	return []string{DocTypeLicense, DocTypePhoto}
}

// AllDocTypes returns all supported document types
func AllDocTypes() []string {
	return []string{DocTypeLicense, DocTypePassport, DocTypeVehicleReg, DocTypeInsurance, DocTypePhoto, DocTypeVehiclePhoto}
}
