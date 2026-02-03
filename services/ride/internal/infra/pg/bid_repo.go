package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ridehail/ride/internal/domain"
)

const BidStatusPending = "pending"
const BidStatusAccepted = "accepted"
const BidStatusRejected = "rejected"

type BidRepo struct {
	pool *pgxpool.Pool
}

func NewBidRepo(pool *pgxpool.Pool) *BidRepo {
	return &BidRepo{pool: pool}
}

func (r *BidRepo) Create(ctx context.Context, bid *domain.Bid) error {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO bids (ride_id, driver_id, price, status, created_at)
		 VALUES ($1, $2, $3, $4, now())
		 RETURNING id, created_at`,
		bid.RideID, bid.DriverID, bid.Price, BidStatusPending,
	)
	var id string
	var createdAt interface{}
	if err := row.Scan(&id, &createdAt); err != nil {
		return err
	}
	bid.ID = id
	bid.Status = BidStatusPending
	return nil
}

func (r *BidRepo) GetByID(ctx context.Context, id string) (*domain.Bid, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, ride_id, driver_id, price, status, created_at FROM bids WHERE id = $1`,
		id,
	)
	return scanBid(row)
}

func (r *BidRepo) ListByRideID(ctx context.Context, rideID string) ([]*domain.Bid, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, ride_id, driver_id, price, status, created_at FROM bids WHERE ride_id = $1 ORDER BY created_at ASC`,
		rideID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Bid
	for rows.Next() {
		var bid domain.Bid
		var createdAt interface{}
		err := rows.Scan(&bid.ID, &bid.RideID, &bid.DriverID, &bid.Price, &bid.Status, &createdAt)
		if err != nil {
			return nil, err
		}
		out = append(out, &bid)
	}
	return out, rows.Err()
}

func (r *BidRepo) AcceptBid(ctx context.Context, bidID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE bids SET status = $1 WHERE id = $2`,
		BidStatusAccepted, bidID,
	)
	return err
}

func (r *BidRepo) RejectOtherBidsForRide(ctx context.Context, rideID, exceptBidID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE bids SET status = $1 WHERE ride_id = $2 AND id != $3`,
		BidStatusRejected, rideID, exceptBidID,
	)
	return err
}

func scanBid(row pgx.Row) (*domain.Bid, error) {
	var bid domain.Bid
	var createdAt interface{}
	err := row.Scan(&bid.ID, &bid.RideID, &bid.DriverID, &bid.Price, &bid.Status, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &bid, nil
}
