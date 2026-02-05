// Package pg — Driver verification repository
package pg

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ridehail/user/internal/domain"
)

// VerificationRepo manages driver verification persistence.
type VerificationRepo struct {
	pool *pgxpool.Pool
}

// NewVerificationRepo creates verification repository.
func NewVerificationRepo(pool *pgxpool.Pool) *VerificationRepo {
	return &VerificationRepo{pool: pool}
}

// --- Driver Verification ---

// CreateVerification creates a new verification request.
func (r *VerificationRepo) CreateVerification(ctx context.Context, v *domain.DriverVerification) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO driver_verifications (user_id, status, license_number, vehicle_model, vehicle_plate, vehicle_year)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`, v.UserID, domain.VerificationStatusPending, v.LicenseNumber, v.VehicleModel, v.VehiclePlate, v.VehicleYear)
	return err
}

// GetVerification returns verification by user ID.
func (r *VerificationRepo) GetVerification(ctx context.Context, userID string) (*domain.DriverVerification, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, status, license_number, vehicle_model, vehicle_plate, vehicle_year,
		       reject_reason, submitted_at, reviewed_at, reviewed_by, created_at, updated_at
		FROM driver_verifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, userID)

	var v domain.DriverVerification
	var vehicleYear *int
	err := row.Scan(
		&v.ID, &v.UserID, &v.Status, &v.LicenseNumber, &v.VehicleModel, &v.VehiclePlate, &vehicleYear,
		&v.RejectReason, &v.SubmittedAt, &v.ReviewedAt, &v.ReviewedBy, &v.CreatedAt, &v.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if vehicleYear != nil {
		v.VehicleYear = *vehicleYear
	}
	return &v, nil
}

// GetVerificationByID returns verification by ID.
func (r *VerificationRepo) GetVerificationByID(ctx context.Context, id string) (*domain.DriverVerification, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, status, license_number, vehicle_model, vehicle_plate, vehicle_year,
		       reject_reason, submitted_at, reviewed_at, reviewed_by, created_at, updated_at
		FROM driver_verifications
		WHERE id = $1
	`, id)

	var v domain.DriverVerification
	var vehicleYear *int
	err := row.Scan(
		&v.ID, &v.UserID, &v.Status, &v.LicenseNumber, &v.VehicleModel, &v.VehiclePlate, &vehicleYear,
		&v.RejectReason, &v.SubmittedAt, &v.ReviewedAt, &v.ReviewedBy, &v.CreatedAt, &v.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrDocumentNotFound
	}
	if err != nil {
		return nil, err
	}
	if vehicleYear != nil {
		v.VehicleYear = *vehicleYear
	}
	return &v, nil
}

// UpdateVerificationStatus updates verification status (admin review).
func (r *VerificationRepo) UpdateVerificationStatus(ctx context.Context, id, status, reviewerID, rejectReason string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE driver_verifications
		SET status = $1, reviewed_at = $2, reviewed_by = $3, reject_reason = $4, updated_at = $2
		WHERE id = $5
	`, status, now, reviewerID, rejectReason, id)
	return err
}

// ListPendingVerifications returns pending verifications (for admin).
func (r *VerificationRepo) ListPendingVerifications(ctx context.Context, limit, offset int) ([]*domain.DriverVerification, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, status, license_number, vehicle_model, vehicle_plate, vehicle_year,
		       reject_reason, submitted_at, reviewed_at, reviewed_by, created_at, updated_at
		FROM driver_verifications
		WHERE status = $1
		ORDER BY submitted_at ASC
		LIMIT $2 OFFSET $3
	`, domain.VerificationStatusPending, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var verifications []*domain.DriverVerification
	for rows.Next() {
		var v domain.DriverVerification
		var vehicleYear *int
		if err := rows.Scan(
			&v.ID, &v.UserID, &v.Status, &v.LicenseNumber, &v.VehicleModel, &v.VehiclePlate, &vehicleYear,
			&v.RejectReason, &v.SubmittedAt, &v.ReviewedAt, &v.ReviewedBy, &v.CreatedAt, &v.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if vehicleYear != nil {
			v.VehicleYear = *vehicleYear
		}
		verifications = append(verifications, &v)
	}
	return verifications, rows.Err()
}

// --- Driver Documents ---

// CreateDocument creates a new document record.
func (r *VerificationRepo) CreateDocument(ctx context.Context, d *domain.DriverDocument) error {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO driver_documents (user_id, verification_id, doc_type, file_name, file_size, content_type, storage_key, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, uploaded_at
	`, d.UserID, nullString(d.ID), d.DocType, d.FileName, d.FileSize, d.ContentType, d.StorageKey, domain.VerificationStatusPending)
	return row.Scan(&d.ID, &d.UploadedAt)
}

// GetDocument returns document by ID.
func (r *VerificationRepo) GetDocument(ctx context.Context, id string) (*domain.DriverDocument, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, doc_type, file_name, file_size, content_type, storage_key, status,
		       reject_reason, uploaded_at, reviewed_at, reviewed_by
		FROM driver_documents
		WHERE id = $1
	`, id)

	var d domain.DriverDocument
	err := row.Scan(
		&d.ID, &d.UserID, &d.DocType, &d.FileName, &d.FileSize, &d.ContentType, &d.StorageKey, &d.Status,
		&d.RejectReason, &d.UploadedAt, &d.ReviewedAt, &d.ReviewedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrDocumentNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// ListDocumentsByUser returns all documents for a user.
func (r *VerificationRepo) ListDocumentsByUser(ctx context.Context, userID string) ([]*domain.DriverDocument, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, doc_type, file_name, file_size, content_type, storage_key, status,
		       reject_reason, uploaded_at, reviewed_at, reviewed_by
		FROM driver_documents
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*domain.DriverDocument
	for rows.Next() {
		var d domain.DriverDocument
		if err := rows.Scan(
			&d.ID, &d.UserID, &d.DocType, &d.FileName, &d.FileSize, &d.ContentType, &d.StorageKey, &d.Status,
			&d.RejectReason, &d.UploadedAt, &d.ReviewedAt, &d.ReviewedBy,
		); err != nil {
			return nil, err
		}
		docs = append(docs, &d)
	}
	return docs, rows.Err()
}

// UpdateDocumentStatus updates document status (admin review).
func (r *VerificationRepo) UpdateDocumentStatus(ctx context.Context, id, status, reviewerID, rejectReason string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE driver_documents
		SET status = $1, reviewed_at = $2, reviewed_by = $3, reject_reason = $4
		WHERE id = $5
	`, status, now, reviewerID, rejectReason, id)
	return err
}

// DeleteDocument removes a document record.
func (r *VerificationRepo) DeleteDocument(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM driver_documents WHERE id = $1`, id)
	return err
}

// GetDocumentByType returns latest document of given type for user.
func (r *VerificationRepo) GetDocumentByType(ctx context.Context, userID, docType string) (*domain.DriverDocument, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, doc_type, file_name, file_size, content_type, storage_key, status,
		       reject_reason, uploaded_at, reviewed_at, reviewed_by
		FROM driver_documents
		WHERE user_id = $1 AND doc_type = $2
		ORDER BY uploaded_at DESC
		LIMIT 1
	`, userID, docType)

	var d domain.DriverDocument
	err := row.Scan(
		&d.ID, &d.UserID, &d.DocType, &d.FileName, &d.FileSize, &d.ContentType, &d.StorageKey, &d.Status,
		&d.RejectReason, &d.UploadedAt, &d.ReviewedAt, &d.ReviewedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// SetUserVerified updates user's verified status in profiles.
func (r *VerificationRepo) SetUserVerified(ctx context.Context, userID string, verified bool, verificationID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE profiles
		SET verified = $1, verification_id = $2, updated_at = NOW()
		WHERE user_id = $3
	`, verified, nullString(verificationID), userID)
	return err
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
