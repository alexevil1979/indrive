package usecase

import (
	"context"

	"github.com/ridehail/user/internal/domain"
)

type ProfileRepository interface {
	GetByUserID(ctx context.Context, userID string) (*domain.Profile, error)
	Upsert(ctx context.Context, p *domain.Profile) error
	GetDriverByUserID(ctx context.Context, userID string) (*domain.DriverProfile, error)
	CreateDriver(ctx context.Context, d *domain.DriverProfile) error
}

type ProfileUseCase struct {
	repo ProfileRepository
}

func NewProfileUseCase(repo ProfileRepository) *ProfileUseCase {
	return &ProfileUseCase{repo: repo}
}

func (uc *ProfileUseCase) GetProfile(ctx context.Context, userID string) (*domain.Profile, *domain.DriverProfile, error) {
	p, err := uc.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	d, _ := uc.repo.GetDriverByUserID(ctx, userID)
	return p, d, nil
}

func (uc *ProfileUseCase) UpdateProfile(ctx context.Context, userID string, displayName, phone, avatarURL *string) (*domain.Profile, error) {
	p, err := uc.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		p = &domain.Profile{UserID: userID}
	}
	if displayName != nil {
		p.DisplayName = *displayName
	}
	if phone != nil {
		p.Phone = *phone
	}
	if avatarURL != nil {
		p.AvatarURL = *avatarURL
	}
	if err := uc.repo.Upsert(ctx, p); err != nil {
		return nil, err
	}
	return uc.repo.GetByUserID(ctx, userID)
}

// CreateDriverProfile — driver verification stub (doc upload via MinIO later)
func (uc *ProfileUseCase) CreateDriverProfile(ctx context.Context, userID, licenseNumber string) (*domain.DriverProfile, error) {
	existing, _ := uc.repo.GetDriverByUserID(ctx, userID)
	if existing != nil {
		return existing, nil
	}
	d := &domain.DriverProfile{
		UserID:        userID,
		Verified:      false,
		LicenseNumber: licenseNumber,
	}
	if err := uc.repo.CreateDriver(ctx, d); err != nil {
		return nil, err
	}
	return uc.repo.GetDriverByUserID(ctx, userID)
}
