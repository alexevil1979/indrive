package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ridehail/ride/internal/domain"
)

type RideRepo struct {
	pool *pgxpool.Pool
}

func NewRideRepo(pool *pgxpool.Pool) *RideRepo {
	return &RideRepo{pool: pool}
}

func (r *RideRepo) Create(ctx context.Context, ride *domain.Ride) error {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO rides (passenger_id, status, from_lat, from_lng, from_address, to_lat, to_lng, to_address, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), now())
		 RETURNING id, created_at, updated_at`,
		ride.PassengerID, domain.StatusRequested,
		ride.From.Lat, ride.From.Lng, nullStr(ride.From.Address),
		ride.To.Lat, ride.To.Lng, nullStr(ride.To.Address),
	)
	var id string
	var createdAt, updatedAt interface{}
	if err := row.Scan(&id, &createdAt, &updatedAt); err != nil {
		return err
	}
	ride.ID = id
	ride.Status = domain.StatusRequested
	return nil
}

func (r *RideRepo) GetByID(ctx context.Context, id string) (*domain.Ride, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, passenger_id, driver_id, status, from_lat, from_lng, from_address, to_lat, to_lng, to_address, price, created_at, updated_at
		 FROM rides WHERE id = $1`,
		id,
	)
	return scanRide(row)
}

func (r *RideRepo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE rides SET status = $1, updated_at = now() WHERE id = $2`,
		status, id,
	)
	return err
}

func (r *RideRepo) SetDriverAndPrice(ctx context.Context, id, driverID string, price float64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE rides SET driver_id = $1, price = $2, status = $3, updated_at = now() WHERE id = $4`,
		driverID, price, domain.StatusMatched, id,
	)
	return err
}

func (r *RideRepo) ListByPassenger(ctx context.Context, passengerID string, limit int) ([]*domain.Ride, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, passenger_id, driver_id, status, from_lat, from_lng, from_address, to_lat, to_lng, to_address, price, created_at, updated_at
		 FROM rides WHERE passenger_id = $1 ORDER BY created_at DESC LIMIT $2`,
		passengerID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRides(rows)
}

func (r *RideRepo) ListByDriver(ctx context.Context, driverID string, limit int) ([]*domain.Ride, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, passenger_id, driver_id, status, from_lat, from_lng, from_address, to_lat, to_lng, to_address, price, created_at, updated_at
		 FROM rides WHERE driver_id = $1 ORDER BY created_at DESC LIMIT $2`,
		driverID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRides(rows)
}

// ListOpenRides — rides with status requested or bidding (for drivers to bid)
func (r *RideRepo) ListOpenRides(ctx context.Context, limit int) ([]*domain.Ride, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, passenger_id, driver_id, status, from_lat, from_lng, from_address, to_lat, to_lng, to_address, price, created_at, updated_at
		 FROM rides WHERE status IN ('requested', 'bidding') ORDER BY created_at DESC LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRides(rows)
}

// ListAll — admin: all rides (for dashboard/monitoring)
func (r *RideRepo) ListAll(ctx context.Context, limit int) ([]*domain.Ride, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, passenger_id, driver_id, status, from_lat, from_lng, from_address, to_lat, to_lng, to_address, price, created_at, updated_at
		 FROM rides ORDER BY created_at DESC LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRides(rows)
}

func scanRide(row pgx.Row) (*domain.Ride, error) {
	var ride domain.Ride
	var driverID, fromAddr, toAddr interface{}
	var price *float64
	err := row.Scan(&ride.ID, &ride.PassengerID, &driverID, &ride.Status,
		&ride.From.Lat, &ride.From.Lng, &fromAddr, &ride.To.Lat, &ride.To.Lng, &toAddr,
		&price, &ride.CreatedAt, &ride.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if driverID != nil {
		ride.DriverID = driverID.(string)
	}
	if fromAddr != nil {
		ride.From.Address = fromAddr.(string)
	}
	if toAddr != nil {
		ride.To.Address = toAddr.(string)
	}
	ride.Price = price
	return &ride, nil
}

func scanRides(rows pgx.Rows) ([]*domain.Ride, error) {
	var out []*domain.Ride
	for rows.Next() {
		var ride domain.Ride
		var driverID, fromAddr, toAddr interface{}
		var price *float64
		err := rows.Scan(&ride.ID, &ride.PassengerID, &driverID, &ride.Status,
			&ride.From.Lat, &ride.From.Lng, &fromAddr, &ride.To.Lat, &ride.To.Lng, &toAddr,
			&price, &ride.CreatedAt, &ride.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if driverID != nil {
			ride.DriverID = driverID.(string)
		}
		if fromAddr != nil {
			if s, ok := fromAddr.(string); ok {
				ride.From.Address = s
			}
		}
		if toAddr != nil {
			if s, ok := toAddr.(string); ok {
				ride.To.Address = s
			}
		}
		ride.Price = price
		out = append(out, &ride)
	}
	return out, rows.Err()
}

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
