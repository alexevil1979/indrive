package usecase

import (
	"context"
	"errors"

	"github.com/ridehail/user/internal/domain"
)

var (
	ErrInvalidMapProvider = errors.New("invalid map provider, must be 'google' or 'yandex'")
	ErrUnauthorized       = errors.New("unauthorized to modify settings")
)

// SettingsRepository defines the interface for settings persistence
type SettingsRepository interface {
	GetSettings(ctx context.Context) (*domain.AppSettings, error)
	UpdateSettings(ctx context.Context, settings *domain.AppSettings, userID string) error
	GetMapSettings(ctx context.Context) (*domain.MapSettings, error)
}

// SettingsUseCase handles settings business logic
type SettingsUseCase struct {
	repo SettingsRepository
}

// NewSettingsUseCase creates new settings use case
func NewSettingsUseCase(repo SettingsRepository) *SettingsUseCase {
	return &SettingsUseCase{repo: repo}
}

// GetSettings retrieves current app settings (admin only)
func (uc *SettingsUseCase) GetSettings(ctx context.Context) (*domain.AppSettings, error) {
	return uc.repo.GetSettings(ctx)
}

// UpdateMapProvider updates the active map provider
func (uc *SettingsUseCase) UpdateMapProvider(ctx context.Context, provider domain.MapProvider, userID string) error {
	if !provider.Validate() {
		return ErrInvalidMapProvider
	}

	settings, err := uc.repo.GetSettings(ctx)
	if err != nil {
		return err
	}

	settings.MapProvider = provider
	return uc.repo.UpdateSettings(ctx, settings, userID)
}

// UpdateSettings updates all settings (admin only)
func (uc *SettingsUseCase) UpdateSettings(ctx context.Context, settings *domain.AppSettings, userID string) error {
	if !settings.MapProvider.Validate() {
		return ErrInvalidMapProvider
	}
	return uc.repo.UpdateSettings(ctx, settings, userID)
}

// GetMapSettings returns map settings for mobile apps
func (uc *SettingsUseCase) GetMapSettings(ctx context.Context) (*domain.MapSettings, error) {
	return uc.repo.GetMapSettings(ctx)
}
