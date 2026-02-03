package usecase

import (
	"context"
	"errors"

	"github.com/ridehail/payment/internal/domain"
	"github.com/ridehail/payment/internal/infra/pg"
)

var (
	ErrInvalidAmount  = errors.New("amount must be positive")
	ErrInvalidMethod  = errors.New("method must be cash or card")
	ErrPaymentExists  = errors.New("payment already exists for this ride")
	ErrPaymentNotFound = errors.New("payment not found")
)

type PaymentRepository interface {
	Create(ctx context.Context, p *domain.Payment) error
	GetByID(ctx context.Context, id string) (*domain.Payment, error)
	GetByRideID(ctx context.Context, rideID string) (*domain.Payment, error)
	UpdateStatus(ctx context.Context, id, status, externalID string) error
}

type PaymentUseCase struct {
	repo PaymentRepository
}

func NewPaymentUseCase(repo PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
}

// CreatePayment — checkout: create payment for ride (stub cash/card)
// Later: verify ride completed via Ride service
func (uc *PaymentUseCase) CreatePayment(ctx context.Context, rideID string, amount float64, method string) (*domain.Payment, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if method != domain.MethodCash && method != domain.MethodCard {
		return nil, ErrInvalidMethod
	}
	p := &domain.Payment{
		RideID:   rideID,
		Amount:   amount,
		Currency: "RUB",
		Method:   method,
		Status:   domain.PaymentStatusPending,
	}
	if err := uc.repo.Create(ctx, p); err != nil {
		if errors.Is(err, pg.ErrPaymentExists) {
			return nil, ErrPaymentExists
		}
		return nil, err
	}
	// Stub: immediately complete (cash/card stub success)
	p.Status = domain.PaymentStatusCompleted
	extID := "stub_" + method + "_" + rideID
	_ = uc.repo.UpdateStatus(ctx, p.ID, domain.PaymentStatusCompleted, extID)
	p.ExternalID = extID
	return uc.repo.GetByID(ctx, p.ID)
}

func (uc *PaymentUseCase) GetByRideID(ctx context.Context, rideID string) (*domain.Payment, error) {
	return uc.repo.GetByRideID(ctx, rideID)
}

func (uc *PaymentUseCase) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	return uc.repo.GetByID(ctx, id)
}

// ConfirmPayment — stub: mark completed (for cash on delivery flow)
func (uc *PaymentUseCase) ConfirmPayment(ctx context.Context, id string) (*domain.Payment, error) {
	p, err := uc.repo.GetByID(ctx, id)
	if err != nil || p == nil {
		return nil, ErrPaymentNotFound
	}
	if p.Status == domain.PaymentStatusCompleted {
		return p, nil
	}
	if err := uc.repo.UpdateStatus(ctx, id, domain.PaymentStatusCompleted, ""); err != nil {
		return nil, err
	}
	return uc.repo.GetByID(ctx, id)
}
