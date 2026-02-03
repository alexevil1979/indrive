package usecase

import (
	"context"
	"errors"

	"github.com/ridehail/geolocation/internal/domain"
)

var ErrInvalidCoordinates = errors.New("invalid coordinates: lat in [-90,90], lng in [-180,180]")

type GeoStore interface {
	Set(ctx context.Context, driverID string, lat, lng float64) error
	Nearest(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]domain.DriverLocation, error)
	Remove(ctx context.Context, driverID string) error
}

type LocationUseCase struct {
	store GeoStore
}

func NewLocationUseCase(store GeoStore) *LocationUseCase {
	return &LocationUseCase{store: store}
}

func (uc *LocationUseCase) UpdateDriverLocation(ctx context.Context, driverID string, lat, lng float64) error {
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return ErrInvalidCoordinates
	}
	return uc.store.Set(ctx, driverID, lat, lng)
}

func (uc *LocationUseCase) FindNearestDrivers(ctx context.Context, q domain.NearestQuery) ([]domain.DriverLocation, error) {
	if q.Limit <= 0 {
		q.Limit = 10
	}
	if q.RadiusKM <= 0 {
		q.RadiusKM = 10
	}
	return uc.store.Nearest(ctx, q.Lat, q.Lng, q.RadiusKM, q.Limit)
}

func (uc *LocationUseCase) RemoveDriver(ctx context.Context, driverID string) error {
	return uc.store.Remove(ctx, driverID)
}
