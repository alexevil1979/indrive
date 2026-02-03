package usecase

import (
	"context"
	"testing"

	"github.com/ridehail/ride/internal/domain"
)

func TestRideUseCase_CreateRide_InvalidCoords(t *testing.T) {
	uc := &RideUseCase{}
	_, err := uc.CreateRide(context.Background(), "user1", domain.Point{Lat: 100, Lng: 0}, domain.Point{Lat: 55, Lng: 37})
	if err != ErrInvalidStatus {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestRideUseCase_PlaceBid_InvalidPrice(t *testing.T) {
	uc := &RideUseCase{}
	_, err := uc.PlaceBid(context.Background(), "ride1", "driver1", 0)
	if err != ErrInvalidStatus {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}
