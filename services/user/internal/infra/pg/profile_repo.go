package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ridehail/user/internal/domain"
)

type ProfileRepo struct {
	pool *pgxpool.Pool
}

func NewProfileRepo(pool *pgxpool.Pool) *ProfileRepo {
	return &ProfileRepo{pool: pool}
}

func (r *ProfileRepo) GetByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT user_id, display_name, phone, avatar_url, created_at, updated_at
		 FROM profiles WHERE user_id = $1`,
		userID,
	)
	var p domain.Profile
	err := row.Scan(&p.UserID, &p.DisplayName, &p.Phone, &p.AvatarURL, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *ProfileRepo) Upsert(ctx context.Context, p *domain.Profile) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO profiles (user_id, display_name, phone, avatar_url, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, now(), now())
		 ON CONFLICT (user_id) DO UPDATE SET
		   display_name = COALESCE(EXCLUDED.display_name, profiles.display_name),
		   phone = COALESCE(EXCLUDED.phone, profiles.phone),
		   avatar_url = COALESCE(EXCLUDED.avatar_url, profiles.avatar_url),
		   updated_at = now()`,
		p.UserID, p.DisplayName, p.Phone, p.AvatarURL,
	)
	return err
}

func (r *ProfileRepo) GetDriverByUserID(ctx context.Context, userID string) (*domain.DriverProfile, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT user_id, verified, license_number, doc_license_url, doc_photo_url, created_at, updated_at
		 FROM driver_profiles WHERE user_id = $1`,
		userID,
	)
	var d domain.DriverProfile
	err := row.Scan(&d.UserID, &d.Verified, &d.LicenseNumber, &d.DocLicenseURL, &d.DocPhotoURL, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *ProfileRepo) CreateDriver(ctx context.Context, d *domain.DriverProfile) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO driver_profiles (user_id, verified, license_number, doc_license_url, doc_photo_url, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, now(), now())
		 ON CONFLICT (user_id) DO NOTHING`,
		d.UserID, d.Verified, d.LicenseNumber, d.DocLicenseURL, d.DocPhotoURL,
	)
	return err
}
