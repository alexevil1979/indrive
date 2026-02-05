package pg

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ridehail/payment/internal/domain"
)

// PromoRepo handles promo persistence
type PromoRepo struct {
	pool *pgxpool.Pool
}

// NewPromoRepo creates a new promo repository
func NewPromoRepo(pool *pgxpool.Pool) *PromoRepo {
	return &PromoRepo{pool: pool}
}

// Create inserts a new promo code
func (r *PromoRepo) Create(ctx context.Context, promo *domain.Promo) error {
	query := `
		INSERT INTO promos (code, description, type, value, min_order_value, max_discount, 
		                    usage_limit, per_user_limit, is_active, starts_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`
	var expiresAt *time.Time
	if !promo.ExpiresAt.IsZero() {
		expiresAt = &promo.ExpiresAt
	}

	return r.pool.QueryRow(ctx, query,
		strings.ToUpper(promo.Code),
		promo.Description,
		promo.Type,
		promo.Value,
		promo.MinOrderValue,
		promo.MaxDiscount,
		promo.UsageLimit,
		promo.PerUserLimit,
		promo.IsActive,
		promo.StartsAt,
		expiresAt,
	).Scan(&promo.ID, &promo.CreatedAt, &promo.UpdatedAt)
}

// GetByID retrieves a promo by ID
func (r *PromoRepo) GetByID(ctx context.Context, id string) (*domain.Promo, error) {
	query := `
		SELECT id, code, description, type, value, min_order_value, max_discount,
		       usage_limit, usage_count, per_user_limit, is_active, starts_at, 
		       COALESCE(expires_at, '0001-01-01'::timestamptz), created_at, updated_at
		FROM promos WHERE id = $1
	`
	var promo domain.Promo
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&promo.ID, &promo.Code, &promo.Description, &promo.Type, &promo.Value,
		&promo.MinOrderValue, &promo.MaxDiscount, &promo.UsageLimit, &promo.UsageCount,
		&promo.PerUserLimit, &promo.IsActive, &promo.StartsAt, &promo.ExpiresAt,
		&promo.CreatedAt, &promo.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &promo, nil
}

// GetByCode retrieves a promo by code (case-insensitive)
func (r *PromoRepo) GetByCode(ctx context.Context, code string) (*domain.Promo, error) {
	query := `
		SELECT id, code, description, type, value, min_order_value, max_discount,
		       usage_limit, usage_count, per_user_limit, is_active, starts_at, 
		       COALESCE(expires_at, '0001-01-01'::timestamptz), created_at, updated_at
		FROM promos WHERE UPPER(code) = UPPER($1)
	`
	var promo domain.Promo
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&promo.ID, &promo.Code, &promo.Description, &promo.Type, &promo.Value,
		&promo.MinOrderValue, &promo.MaxDiscount, &promo.UsageLimit, &promo.UsageCount,
		&promo.PerUserLimit, &promo.IsActive, &promo.StartsAt, &promo.ExpiresAt,
		&promo.CreatedAt, &promo.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &promo, nil
}

// Update updates a promo
func (r *PromoRepo) Update(ctx context.Context, promo *domain.Promo) error {
	query := `
		UPDATE promos SET
			description = $2,
			type = $3,
			value = $4,
			min_order_value = $5,
			max_discount = $6,
			usage_limit = $7,
			per_user_limit = $8,
			is_active = $9,
			starts_at = $10,
			expires_at = $11,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`
	var expiresAt *time.Time
	if !promo.ExpiresAt.IsZero() {
		expiresAt = &promo.ExpiresAt
	}

	return r.pool.QueryRow(ctx, query,
		promo.ID,
		promo.Description,
		promo.Type,
		promo.Value,
		promo.MinOrderValue,
		promo.MaxDiscount,
		promo.UsageLimit,
		promo.PerUserLimit,
		promo.IsActive,
		promo.StartsAt,
		expiresAt,
	).Scan(&promo.UpdatedAt)
}

// Delete removes a promo
func (r *PromoRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM promos WHERE id = $1", id)
	return err
}

// List returns all promos with pagination
func (r *PromoRepo) List(ctx context.Context, limit, offset int, activeOnly bool) ([]domain.Promo, int, error) {
	// Count total
	countQuery := "SELECT COUNT(*) FROM promos"
	if activeOnly {
		countQuery += " WHERE is_active = true"
	}
	var total int
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get promos
	query := `
		SELECT id, code, description, type, value, min_order_value, max_discount,
		       usage_limit, usage_count, per_user_limit, is_active, starts_at, 
		       COALESCE(expires_at, '0001-01-01'::timestamptz), created_at, updated_at
		FROM promos
	`
	if activeOnly {
		query += " WHERE is_active = true"
	}
	query += " ORDER BY created_at DESC LIMIT $1 OFFSET $2"

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var promos []domain.Promo
	for rows.Next() {
		var promo domain.Promo
		if err := rows.Scan(
			&promo.ID, &promo.Code, &promo.Description, &promo.Type, &promo.Value,
			&promo.MinOrderValue, &promo.MaxDiscount, &promo.UsageLimit, &promo.UsageCount,
			&promo.PerUserLimit, &promo.IsActive, &promo.StartsAt, &promo.ExpiresAt,
			&promo.CreatedAt, &promo.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		promos = append(promos, promo)
	}
	return promos, total, rows.Err()
}

// GetUserPromoCount returns how many times a user used a specific promo
func (r *PromoRepo) GetUserPromoCount(ctx context.Context, userID, promoID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM user_promos WHERE user_id = $1 AND promo_id = $2`
	err := r.pool.QueryRow(ctx, query, userID, promoID).Scan(&count)
	return count, err
}

// RecordUsage records promo usage by a user
func (r *PromoRepo) RecordUsage(ctx context.Context, usage *domain.UserPromo) error {
	query := `
		INSERT INTO user_promos (user_id, promo_id, ride_id, discount)
		VALUES ($1, $2, $3, $4)
		RETURNING id, used_at
	`
	var rideID *string
	if usage.RideID != "" {
		rideID = &usage.RideID
	}

	return r.pool.QueryRow(ctx, query,
		usage.UserID,
		usage.PromoID,
		rideID,
		usage.Discount,
	).Scan(&usage.ID, &usage.UsedAt)
}

// GetUserPromos returns user's promo usage history
func (r *PromoRepo) GetUserPromos(ctx context.Context, userID string, limit int) ([]domain.UserPromo, error) {
	query := `
		SELECT up.id, up.user_id, up.promo_id, COALESCE(up.ride_id::text, ''), up.discount, up.used_at
		FROM user_promos up
		WHERE up.user_id = $1
		ORDER BY up.used_at DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usages []domain.UserPromo
	for rows.Next() {
		var u domain.UserPromo
		if err := rows.Scan(&u.ID, &u.UserID, &u.PromoID, &u.RideID, &u.Discount, &u.UsedAt); err != nil {
			return nil, err
		}
		usages = append(usages, u)
	}
	return usages, rows.Err()
}

// CheckPromoUsedForRide checks if promo was already used for a ride
func (r *PromoRepo) CheckPromoUsedForRide(ctx context.Context, rideID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM user_promos WHERE ride_id = $1)`
	err := r.pool.QueryRow(ctx, query, rideID).Scan(&exists)
	return exists, err
}
