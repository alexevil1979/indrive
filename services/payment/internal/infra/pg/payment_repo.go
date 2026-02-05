package pg

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ridehail/payment/internal/domain"
)

var ErrPaymentExists = errors.New("payment already exists for this ride")

// PaymentRepo — payment repository
type PaymentRepo struct {
	pool *pgxpool.Pool
}

// NewPaymentRepo creates payment repository
func NewPaymentRepo(pool *pgxpool.Pool) *PaymentRepo {
	return &PaymentRepo{pool: pool}
}

// Create creates a new payment
func (r *PaymentRepo) Create(ctx context.Context, p *domain.Payment) error {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO payments (ride_id, user_id, amount, currency, method, provider, status, external_id, 
		                       confirm_url, description, metadata, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now(), now())
		 ON CONFLICT (ride_id) DO NOTHING
		 RETURNING id, created_at, updated_at`,
		p.RideID, nullStr(p.UserID), p.Amount, p.Currency, p.Method, p.Provider, p.Status,
		nullStr(p.ExternalID), nullStr(p.ConfirmURL), nullStr(p.Description), nullStr(p.Metadata),
	)
	err := row.Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPaymentExists
		}
		return err
	}
	return nil
}

// GetByID returns payment by ID
func (r *PaymentRepo) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, ride_id, user_id, amount, currency, method, provider, status, external_id,
		        confirm_url, description, metadata, fail_reason, refunded_at, paid_at, created_at, updated_at
		 FROM payments WHERE id = $1`,
		id,
	)
	return scanPayment(row)
}

// GetByRideID returns payment by ride ID
func (r *PaymentRepo) GetByRideID(ctx context.Context, rideID string) (*domain.Payment, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, ride_id, user_id, amount, currency, method, provider, status, external_id,
		        confirm_url, description, metadata, fail_reason, refunded_at, paid_at, created_at, updated_at
		 FROM payments WHERE ride_id = $1`,
		rideID,
	)
	return scanPayment(row)
}

// GetByExternalID returns payment by external provider ID
func (r *PaymentRepo) GetByExternalID(ctx context.Context, externalID string) (*domain.Payment, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, ride_id, user_id, amount, currency, method, provider, status, external_id,
		        confirm_url, description, metadata, fail_reason, refunded_at, paid_at, created_at, updated_at
		 FROM payments WHERE external_id = $1`,
		externalID,
	)
	return scanPayment(row)
}

// UpdateStatus updates payment status
func (r *PaymentRepo) UpdateStatus(ctx context.Context, id, status, externalID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payments SET status = $1, external_id = COALESCE($2, external_id), updated_at = now() WHERE id = $3`,
		status, nullStr(externalID), id,
	)
	return err
}

// UpdatePayment updates payment fields
func (r *PaymentRepo) UpdatePayment(ctx context.Context, p *domain.Payment) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payments SET status = $1, external_id = $2, confirm_url = $3, fail_reason = $4, 
		        paid_at = $5, refunded_at = $6, updated_at = now()
		 WHERE id = $7`,
		p.Status, nullStr(p.ExternalID), nullStr(p.ConfirmURL), nullStr(p.FailReason),
		p.PaidAt, p.RefundedAt, p.ID,
	)
	return err
}

// ListByUser returns user's payments
func (r *PaymentRepo) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Payment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, ride_id, user_id, amount, currency, method, provider, status, external_id,
		        confirm_url, description, metadata, fail_reason, refunded_at, paid_at, created_at, updated_at
		 FROM payments WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		p, err := scanPaymentRow(rows)
		if err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}

// --- Payment Methods ---

// CreatePaymentMethod saves a payment method
func (r *PaymentRepo) CreatePaymentMethod(ctx context.Context, pm *domain.PaymentMethod) error {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO payment_methods (user_id, provider, type, last4, brand, expiry_month, expiry_year, token_id, is_default)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (user_id, provider, token_id) DO UPDATE SET 
		     last4 = EXCLUDED.last4, brand = EXCLUDED.brand, 
		     expiry_month = EXCLUDED.expiry_month, expiry_year = EXCLUDED.expiry_year
		 RETURNING id, created_at`,
		pm.UserID, pm.Provider, pm.Type, pm.Last4, pm.Brand, pm.ExpiryMonth, pm.ExpiryYear, pm.TokenID, pm.IsDefault,
	)
	return row.Scan(&pm.ID, &pm.CreatedAt)
}

// GetPaymentMethod returns payment method by ID
func (r *PaymentRepo) GetPaymentMethod(ctx context.Context, id string) (*domain.PaymentMethod, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, provider, type, last4, brand, expiry_month, expiry_year, token_id, is_default, created_at
		 FROM payment_methods WHERE id = $1`,
		id,
	)
	return scanPaymentMethod(row)
}

// ListPaymentMethods returns user's saved payment methods
func (r *PaymentRepo) ListPaymentMethods(ctx context.Context, userID string) ([]*domain.PaymentMethod, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, provider, type, last4, brand, expiry_month, expiry_year, token_id, is_default, created_at
		 FROM payment_methods WHERE user_id = $1 ORDER BY is_default DESC, created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []*domain.PaymentMethod
	for rows.Next() {
		pm := &domain.PaymentMethod{}
		err := rows.Scan(&pm.ID, &pm.UserID, &pm.Provider, &pm.Type, &pm.Last4, &pm.Brand,
			&pm.ExpiryMonth, &pm.ExpiryYear, &pm.TokenID, &pm.IsDefault, &pm.CreatedAt)
		if err != nil {
			return nil, err
		}
		methods = append(methods, pm)
	}
	return methods, nil
}

// DeletePaymentMethod removes a saved payment method
func (r *PaymentRepo) DeletePaymentMethod(ctx context.Context, id, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM payment_methods WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

// SetDefaultPaymentMethod sets a payment method as default
func (r *PaymentRepo) SetDefaultPaymentMethod(ctx context.Context, id, userID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Clear current default
	_, err = tx.Exec(ctx, `UPDATE payment_methods SET is_default = FALSE WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	// Set new default
	_, err = tx.Exec(ctx, `UPDATE payment_methods SET is_default = TRUE WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// --- Refunds ---

// CreateRefund creates a refund record
func (r *PaymentRepo) CreateRefund(ctx context.Context, paymentID string, amount float64, reason, externalID string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO refunds (payment_id, amount, reason, external_id, status)
		 VALUES ($1, $2, $3, $4, 'pending')
		 RETURNING id`,
		paymentID, amount, reason, nullStr(externalID),
	).Scan(&id)
	return id, err
}

// UpdateRefundStatus updates refund status
func (r *PaymentRepo) UpdateRefundStatus(ctx context.Context, id, status string) error {
	processedAt := (*time.Time)(nil)
	if status == "succeeded" || status == "failed" {
		now := time.Now()
		processedAt = &now
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE refunds SET status = $1, processed_at = $2 WHERE id = $3`,
		status, processedAt, id,
	)
	return err
}

// --- Helpers ---

func scanPayment(row pgx.Row) (*domain.Payment, error) {
	var p domain.Payment
	var userID, extID, confirmURL, desc, metadata, failReason *string
	var refundedAt, paidAt *time.Time

	err := row.Scan(&p.ID, &p.RideID, &userID, &p.Amount, &p.Currency, &p.Method, &p.Provider, &p.Status, &extID,
		&confirmURL, &desc, &metadata, &failReason, &refundedAt, &paidAt, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if userID != nil {
		p.UserID = *userID
	}
	if extID != nil {
		p.ExternalID = *extID
	}
	if confirmURL != nil {
		p.ConfirmURL = *confirmURL
	}
	if desc != nil {
		p.Description = *desc
	}
	if metadata != nil {
		p.Metadata = *metadata
	}
	if failReason != nil {
		p.FailReason = *failReason
	}
	p.RefundedAt = refundedAt
	p.PaidAt = paidAt

	return &p, nil
}

func scanPaymentRow(rows pgx.Rows) (*domain.Payment, error) {
	var p domain.Payment
	var userID, extID, confirmURL, desc, metadata, failReason *string
	var refundedAt, paidAt *time.Time

	err := rows.Scan(&p.ID, &p.RideID, &userID, &p.Amount, &p.Currency, &p.Method, &p.Provider, &p.Status, &extID,
		&confirmURL, &desc, &metadata, &failReason, &refundedAt, &paidAt, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if userID != nil {
		p.UserID = *userID
	}
	if extID != nil {
		p.ExternalID = *extID
	}
	if confirmURL != nil {
		p.ConfirmURL = *confirmURL
	}
	if desc != nil {
		p.Description = *desc
	}
	if metadata != nil {
		p.Metadata = *metadata
	}
	if failReason != nil {
		p.FailReason = *failReason
	}
	p.RefundedAt = refundedAt
	p.PaidAt = paidAt

	return &p, nil
}

func scanPaymentMethod(row pgx.Row) (*domain.PaymentMethod, error) {
	pm := &domain.PaymentMethod{}
	err := row.Scan(&pm.ID, &pm.UserID, &pm.Provider, &pm.Type, &pm.Last4, &pm.Brand,
		&pm.ExpiryMonth, &pm.ExpiryYear, &pm.TokenID, &pm.IsDefault, &pm.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return pm, nil
}

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// MetadataToJSON converts map to JSON string
func MetadataToJSON(m map[string]string) string {
	if m == nil {
		return ""
	}
	b, _ := json.Marshal(m)
	return string(b)
}
