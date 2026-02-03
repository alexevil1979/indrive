package usecase

import (
	"context"
	"errors"

	"github.com/ridehail/ride/internal/domain"
)

var (
	ErrRideNotFound   = errors.New("ride not found")
	ErrBidNotFound    = errors.New("bid not found")
	ErrInvalidStatus  = errors.New("invalid status transition")
	ErrNotPassenger   = errors.New("not the ride passenger")
	ErrNotDriver      = errors.New("not the ride driver")
	ErrRideNotBidding = errors.New("ride is not in bidding status")
)

type RideRepository interface {
	Create(ctx context.Context, ride *domain.Ride) error
	GetByID(ctx context.Context, id string) (*domain.Ride, error)
	UpdateStatus(ctx context.Context, id, status string) error
	SetDriverAndPrice(ctx context.Context, id, driverID string, price float64) error
	ListByPassenger(ctx context.Context, passengerID string, limit int) ([]*domain.Ride, error)
	ListByDriver(ctx context.Context, driverID string, limit int) ([]*domain.Ride, error)
	ListOpenRides(ctx context.Context, limit int) ([]*domain.Ride, error)
	ListAll(ctx context.Context, limit int) ([]*domain.Ride, error)
}

type BidRepository interface {
	Create(ctx context.Context, bid *domain.Bid) error
	GetByID(ctx context.Context, id string) (*domain.Bid, error)
	ListByRideID(ctx context.Context, rideID string) ([]*domain.Bid, error)
	AcceptBid(ctx context.Context, bidID string) error
	RejectOtherBidsForRide(ctx context.Context, rideID, exceptBidID string) error
}

type EventPublisher interface {
	SendRideRequested(ctx context.Context, rideID, passengerID string, payload interface{}) error
	SendRideBidPlaced(ctx context.Context, rideID, bidID, driverID string, price float64) error
	SendRideMatched(ctx context.Context, rideID, driverID string, price float64) error
	SendRideStatusChanged(ctx context.Context, rideID, status string) error
}

type RideUseCase struct {
	rideRepo RideRepository
	bidRepo  BidRepository
	pub      EventPublisher
}

func NewRideUseCase(rideRepo RideRepository, bidRepo BidRepository, pub EventPublisher) *RideUseCase {
	return &RideUseCase{rideRepo: rideRepo, bidRepo: bidRepo, pub: pub}
}

func (uc *RideUseCase) CreateRide(ctx context.Context, passengerID string, from, to domain.Point) (*domain.Ride, error) {
	if from.Lat < -90 || from.Lat > 90 || from.Lng < -180 || from.Lng > 180 {
		return nil, ErrInvalidStatus
	}
	if to.Lat < -90 || to.Lat > 90 || to.Lng < -180 || to.Lng > 180 {
		return nil, ErrInvalidStatus
	}
	ride := &domain.Ride{
		PassengerID: passengerID,
		Status:      domain.StatusRequested,
		From:        from,
		To:          to,
	}
	if err := uc.rideRepo.Create(ctx, ride); err != nil {
		return nil, err
	}
	ride.Status = domain.StatusBidding
	_ = uc.pub.SendRideRequested(ctx, ride.ID, passengerID, ride)
	return ride, nil
}

func (uc *RideUseCase) GetRide(ctx context.Context, id string) (*domain.Ride, error) {
	return uc.rideRepo.GetByID(ctx, id)
}

func (uc *RideUseCase) PlaceBid(ctx context.Context, rideID, driverID string, price float64) (*domain.Bid, error) {
	if price <= 0 {
		return nil, ErrInvalidStatus
	}
	ride, err := uc.rideRepo.GetByID(ctx, rideID)
	if err != nil || ride == nil {
		return nil, ErrRideNotFound
	}
	if ride.Status != domain.StatusRequested && ride.Status != domain.StatusBidding {
		return nil, ErrRideNotBidding
	}
	bid := &domain.Bid{RideID: rideID, DriverID: driverID, Price: price}
	if err := uc.bidRepo.Create(ctx, bid); err != nil {
		return nil, err
	}
	_ = uc.pub.SendRideBidPlaced(ctx, rideID, bid.ID, driverID, price)
	return bid, nil
}

func (uc *RideUseCase) ListBids(ctx context.Context, rideID string) ([]*domain.Bid, error) {
	return uc.bidRepo.ListByRideID(ctx, rideID)
}

func (uc *RideUseCase) AcceptBid(ctx context.Context, rideID, bidID, passengerID string) (*domain.Ride, error) {
	ride, err := uc.rideRepo.GetByID(ctx, rideID)
	if err != nil || ride == nil {
		return nil, ErrRideNotFound
	}
	if ride.PassengerID != passengerID {
		return nil, ErrNotPassenger
	}
	if ride.Status != domain.StatusRequested && ride.Status != domain.StatusBidding {
		return nil, ErrInvalidStatus
	}
	bid, err := uc.bidRepo.GetByID(ctx, bidID)
	if err != nil || bid == nil {
		return nil, ErrBidNotFound
	}
	if bid.RideID != rideID {
		return nil, ErrBidNotFound
	}
	if err := uc.bidRepo.AcceptBid(ctx, bidID); err != nil {
		return nil, err
	}
	if err := uc.bidRepo.RejectOtherBidsForRide(ctx, rideID, bidID); err != nil {
		return nil, err
	}
	if err := uc.rideRepo.SetDriverAndPrice(ctx, rideID, bid.DriverID, bid.Price); err != nil {
		return nil, err
	}
	_ = uc.pub.SendRideMatched(ctx, rideID, bid.DriverID, bid.Price)
	return uc.rideRepo.GetByID(ctx, rideID)
}

func (uc *RideUseCase) UpdateStatus(ctx context.Context, rideID, status, userID, userRole string) (*domain.Ride, error) {
	valid := map[string]bool{
		domain.StatusInProgress: true,
		domain.StatusCompleted:   true,
		domain.StatusCancelled:   true,
	}
	if !valid[status] {
		return nil, ErrInvalidStatus
	}
	ride, err := uc.rideRepo.GetByID(ctx, rideID)
	if err != nil || ride == nil {
		return nil, ErrRideNotFound
	}
	if userRole == "passenger" && ride.PassengerID != userID {
		return nil, ErrNotPassenger
	}
	if userRole == "driver" && ride.DriverID != userID {
		return nil, ErrNotDriver
	}
	if err := uc.rideRepo.UpdateStatus(ctx, rideID, status); err != nil {
		return nil, err
	}
	_ = uc.pub.SendRideStatusChanged(ctx, rideID, status)
	return uc.rideRepo.GetByID(ctx, rideID)
}

func (uc *RideUseCase) ListRidesByPassenger(ctx context.Context, passengerID string, limit int) ([]*domain.Ride, error) {
	return uc.rideRepo.ListByPassenger(ctx, passengerID, limit)
}

func (uc *RideUseCase) ListRidesByDriver(ctx context.Context, driverID string, limit int) ([]*domain.Ride, error) {
	return uc.rideRepo.ListByDriver(ctx, driverID, limit)
}

// ListOpenRides — rides in requested/bidding for drivers to bid
func (uc *RideUseCase) ListOpenRides(ctx context.Context, limit int) ([]*domain.Ride, error) {
	return uc.rideRepo.ListOpenRides(ctx, limit)
}

// ListAllRides — admin: all rides
func (uc *RideUseCase) ListAllRides(ctx context.Context, limit int) ([]*domain.Ride, error) {
	return uc.rideRepo.ListAll(ctx, limit)
}
