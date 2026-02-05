// Package pg — OAuth accounts repository
package pg

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ridehail/auth/internal/domain"
)

// OAuthRepo manages OAuth account persistence.
type OAuthRepo struct {
	pool *pgxpool.Pool
}

// NewOAuthRepo creates a new OAuth repository.
func NewOAuthRepo(pool *pgxpool.Pool) *OAuthRepo {
	return &OAuthRepo{pool: pool}
}

// FindByProvider finds OAuth account by provider and provider user ID.
func (r *OAuthRepo) FindByProvider(ctx context.Context, provider, providerUserID string) (*domain.OAuthAccount, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, provider, provider_user_id, email, name, avatar_url,
		       access_token, refresh_token, expires_at, created_at, updated_at
		FROM oauth_accounts
		WHERE provider = $1 AND provider_user_id = $2
	`, provider, providerUserID)

	var oa domain.OAuthAccount
	var expiresAt *time.Time
	err := row.Scan(
		&oa.ID, &oa.UserID, &oa.Provider, &oa.ProviderUserID,
		&oa.Email, &oa.Name, &oa.AvatarURL,
		&oa.AccessToken, &oa.RefreshToken, &expiresAt,
		&oa.CreatedAt, &oa.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if expiresAt != nil {
		oa.ExpiresAt = *expiresAt
	}
	return &oa, nil
}

// FindByUserAndProvider finds OAuth account for a specific user and provider.
func (r *OAuthRepo) FindByUserAndProvider(ctx context.Context, userID, provider string) (*domain.OAuthAccount, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, provider, provider_user_id, email, name, avatar_url,
		       access_token, refresh_token, expires_at, created_at, updated_at
		FROM oauth_accounts
		WHERE user_id = $1 AND provider = $2
	`, userID, provider)

	var oa domain.OAuthAccount
	var expiresAt *time.Time
	err := row.Scan(
		&oa.ID, &oa.UserID, &oa.Provider, &oa.ProviderUserID,
		&oa.Email, &oa.Name, &oa.AvatarURL,
		&oa.AccessToken, &oa.RefreshToken, &expiresAt,
		&oa.CreatedAt, &oa.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if expiresAt != nil {
		oa.ExpiresAt = *expiresAt
	}
	return &oa, nil
}

// Create inserts a new OAuth account.
func (r *OAuthRepo) Create(ctx context.Context, oa *domain.OAuthAccount) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO oauth_accounts (user_id, provider, provider_user_id, email, name, avatar_url, access_token, refresh_token, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, oa.UserID, oa.Provider, oa.ProviderUserID, oa.Email, oa.Name, oa.AvatarURL, oa.AccessToken, oa.RefreshToken, nullTime(oa.ExpiresAt))
	return err
}

// Update updates OAuth account tokens.
func (r *OAuthRepo) Update(ctx context.Context, oa *domain.OAuthAccount) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE oauth_accounts
		SET email = $1, name = $2, avatar_url = $3, access_token = $4, refresh_token = $5, expires_at = $6, updated_at = NOW()
		WHERE id = $7
	`, oa.Email, oa.Name, oa.AvatarURL, oa.AccessToken, oa.RefreshToken, nullTime(oa.ExpiresAt), oa.ID)
	return err
}

// ListByUser returns all OAuth accounts for a user.
func (r *OAuthRepo) ListByUser(ctx context.Context, userID string) ([]*domain.OAuthAccount, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, provider, provider_user_id, email, name, avatar_url, created_at
		FROM oauth_accounts
		WHERE user_id = $1
		ORDER BY created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*domain.OAuthAccount
	for rows.Next() {
		var oa domain.OAuthAccount
		if err := rows.Scan(&oa.ID, &oa.UserID, &oa.Provider, &oa.ProviderUserID, &oa.Email, &oa.Name, &oa.AvatarURL, &oa.CreatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, &oa)
	}
	return accounts, rows.Err()
}

// Delete removes an OAuth account link.
func (r *OAuthRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM oauth_accounts WHERE id = $1`, id)
	return err
}

func nullTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}
