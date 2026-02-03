package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ridehail/payment/internal/domain"
)

type PaymentRepo struct {
	pool *pgxpool.Pool
}

func NewPaymentRepo(pool *pgxpool.Pool) *PaymentRepo {
	return &PaymentRepo{pool: pool}
}

func (r *PaymentRepo) Create(ctx context.Context, p *domain.Payment) error {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO payments (ride_id, amount, currency, method, status, external_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, now(), now())
		 ON CONFLICT (ride_id) DO NOTHING
		 RETURNING id, created_at, updated_at`,
		p.RideID, p.Amount, p.Currency, p.Method, p.Status, nullStr(p.ExternalID),
	)
	var id string
	var createdAt, updatedAt interface{}
	err := row.Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPaymentExists
		}
		return err
	}
	p.ID = id
	return nil
}

var ErrPaymentExists = errors.New("payment already exists for this ride")

func (r *PaymentRepo) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, ride_id, amount, currency, method, status, external_id, created_at, updated_at
		 FROM payments WHERE id = $1`,
		id,
	)
	return scanPayment(row)
}

func (r *PaymentRepo) GetByRideID(ctx context.Context, rideID string) (*domain.Payment, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, ride_id, amount, currency, method, status, external_id, created_at, updated_at
		 FROM payments WHERE ride_id = $1`,
		rideID,
	)
	return scanPayment(row)
}

func (r *PaymentRepo) UpdateStatus(ctx context.Context, id, status, externalID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payments SET status = $1, external_id = COALESCE($2, external_id), updated_at = now() WHERE id = $3`,
		status, nullStr(externalID), id,
	)
	return err
}

func scanPayment(row pgx.Row) (*domain.Payment, error) {
	var p domain.Payment
	var extID *string
	err := row.Scan(&p.ID, &p.RideID, &p.Amount, &p.Currency, &p.Method, &p.Status, &extID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if extID != nil {
		p.ExternalID = *extID
	}
	return &p, nil
}

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
