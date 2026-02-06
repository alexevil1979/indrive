package pg

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ridehail/user/internal/domain"
)

type SettingsRepo struct {
	pool *pgxpool.Pool
}

func NewSettingsRepo(pool *pgxpool.Pool) *SettingsRepo {
	return &SettingsRepo{pool: pool}
}

// GetSettings retrieves current app settings
func (r *SettingsRepo) GetSettings(ctx context.Context) (*domain.AppSettings, error) {
	query := `
		SELECT id, map_provider, google_maps_api_key, yandex_maps_api_key,
		       default_language, default_currency, updated_at, updated_by
		FROM app_settings
		WHERE id = 'default'
	`

	var s domain.AppSettings
	var googleKey, yandexKey, updatedBy *string

	err := r.pool.QueryRow(ctx, query).Scan(
		&s.ID,
		&s.MapProvider,
		&googleKey,
		&yandexKey,
		&s.DefaultLanguage,
		&s.DefaultCurrency,
		&s.UpdatedAt,
		&updatedBy,
	)
	if err != nil {
		return nil, err
	}

	if googleKey != nil {
		s.GoogleMapsAPIKey = *googleKey
	}
	if yandexKey != nil {
		s.YandexMapsAPIKey = *yandexKey
	}
	if updatedBy != nil {
		s.UpdatedBy = *updatedBy
	}

	return &s, nil
}

// UpdateSettings updates app settings
func (r *SettingsRepo) UpdateSettings(ctx context.Context, settings *domain.AppSettings, userID string) error {
	query := `
		UPDATE app_settings
		SET map_provider = $1,
		    google_maps_api_key = $2,
		    yandex_maps_api_key = $3,
		    default_language = $4,
		    default_currency = $5,
		    updated_at = $6,
		    updated_by = $7
		WHERE id = 'default'
	`

	settings.UpdatedAt = time.Now()
	settings.UpdatedBy = userID

	_, err := r.pool.Exec(ctx, query,
		settings.MapProvider,
		nullString(settings.GoogleMapsAPIKey),
		nullString(settings.YandexMapsAPIKey),
		settings.DefaultLanguage,
		settings.DefaultCurrency,
		settings.UpdatedAt,
		userID,
	)
	return err
}

// GetMapSettings returns map-specific settings for mobile apps (without exposing all API keys)
func (r *SettingsRepo) GetMapSettings(ctx context.Context) (*domain.MapSettings, error) {
	query := `
		SELECT map_provider, google_maps_api_key, yandex_maps_api_key
		FROM app_settings
		WHERE id = 'default'
	`

	var provider domain.MapProvider
	var googleKey, yandexKey *string

	err := r.pool.QueryRow(ctx, query).Scan(&provider, &googleKey, &yandexKey)
	if err != nil {
		return nil, err
	}

	ms := &domain.MapSettings{
		Provider: provider,
	}

	// Return only the active provider's API key
	switch provider {
	case domain.MapProviderGoogle:
		if googleKey != nil {
			ms.APIKey = *googleKey
		}
	case domain.MapProviderYandex:
		if yandexKey != nil {
			ms.APIKey = *yandexKey
		}
	}

	return ms, nil
}

func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
