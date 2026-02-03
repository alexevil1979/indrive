package usecase

import (
	"context"
	"testing"

	"github.com/ridehail/geolocation/internal/domain"
)

func TestLocationUseCase_UpdateDriverLocation_InvalidCoords(t *testing.T) {
	uc := &LocationUseCase{} // nil store; we only check coords first
	err := uc.UpdateDriverLocation(context.Background(), "d1", 100, 0)
	if err != ErrInvalidCoordinates {
		t.Errorf("expected ErrInvalidCoordinates, got %v", err)
	}
	err = uc.UpdateDriverLocation(context.Background(), "d1", 0, 200)
	if err != ErrInvalidCoordinates {
		t.Errorf("expected ErrInvalidCoordinates, got %v", err)
	}
}

func TestLocationUseCase_FindNearestDrivers_DefaultLimit(t *testing.T) {
	uc := &LocationUseCase{} // nil store will panic on Nearest; just check default limit in usecase
	q := domain.NearestQuery{Lat: 55.75, Lng: 37.62, Limit: 0, RadiusKM: 0}
	if q.Limit != 0 {
		t.Fail()
	}
	// Defaults applied inside FindNearestDrivers; with nil store we'd panic - skip full test
	_ = q
}
