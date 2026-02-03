// Package pg — User repository implementation
package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ridehail/auth/internal/domain"
)

// UserRepo — persistence for users table
type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3)
		 RETURNING id, created_at`,
		u.Email, u.PasswordHash, u.Role,
	)
	var id string
	var createdAt interface{}
	if err := row.Scan(&id, &createdAt); err != nil {
		if isUniqueViolation(err) {
			return domain.ErrEmailTaken
		}
		return err
	}
	u.ID = id
	return nil
}

func (r *UserRepo) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, role, created_at FROM users WHERE email = $1`,
		email,
	)
	return scanUser(row)
}

func (r *UserRepo) ByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, role, created_at FROM users WHERE id = $1`,
		id,
	)
	return scanUser(row)
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	var id string
	err := row.Scan(&id, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	u.ID = id
	return &u, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // unique_violation
	}
	return false
}
