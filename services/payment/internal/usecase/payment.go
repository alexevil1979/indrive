package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ridehail/payment/internal/domain"
	"github.com/ridehail/payment/internal/infra/gateway"
	"github.com/ridehail/payment/internal/infra/pg"
)

// PaymentRepository interface
type PaymentRepository interface {
	Create(ctx context.Context, p *domain.Payment) error
	GetByID(ctx context.Context, id string) (*domain.Payment, error)
	GetByRideID(ctx context.Context, rideID string) (*domain.Payment, error)
	GetByExternalID(ctx context.Context, externalID string) (*domain.Payment, error)
	UpdateStatus(ctx context.Context, id, status, externalID string) error
	UpdatePayment(ctx context.Context, p *domain.Payment) error
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Payment, error)

	CreatePaymentMethod(ctx context.Context, pm *domain.PaymentMethod) error
	GetPaymentMethod(ctx context.Context, id string) (*domain.PaymentMethod, error)
	ListPaymentMethods(ctx context.Context, userID string) ([]*domain.PaymentMethod, error)
	DeletePaymentMethod(ctx context.Context, id, userID string) error
	SetDefaultPaymentMethod(ctx context.Context, id, userID string) error

	CreateRefund(ctx context.Context, paymentID string, amount float64, reason, externalID string) (string, error)
	UpdateRefundStatus(ctx context.Context, id, status string) error
}

// PaymentUseCase — payment business logic
type PaymentUseCase struct {
	repo     PaymentRepository
	gateways *gateway.Manager
}

// NewPaymentUseCase creates payment use case
func NewPaymentUseCase(repo PaymentRepository, gateways *gateway.Manager) *PaymentUseCase {
	return &PaymentUseCase{
		repo:     repo,
		gateways: gateways,
	}
}

// CreatePaymentInput — input for creating payment
type CreatePaymentInput struct {
	RideID      string
	UserID      string
	Amount      float64
	Currency    string
	Method      string  // cash | card
	Provider    string  // cash | tinkoff | yoomoney | sber
	Description string
	ReturnURL   string
	SaveCard    bool
	TokenID     string // Use saved card
	UserEmail   string
	UserPhone   string
}

// CreatePayment creates a new payment
func (uc *PaymentUseCase) CreatePayment(ctx context.Context, input CreatePaymentInput) (*domain.PaymentIntent, error) {
	// Validate
	if input.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	if !domain.IsValidMethod(input.Method) {
		return nil, domain.ErrInvalidMethod
	}

	// Default provider based on method
	if input.Provider == "" {
		if input.Method == domain.MethodCash {
			input.Provider = domain.ProviderCash
		} else {
			input.Provider = domain.ProviderTinkoff // Default card provider
		}
	}

	if !domain.IsValidProvider(input.Provider) {
		return nil, domain.ErrInvalidProvider
	}

	// Currency default
	if input.Currency == "" {
		input.Currency = "RUB"
	}

	// Create payment record
	metadata := map[string]string{
		"user_id": input.UserID,
		"ride_id": input.RideID,
	}
	metadataJSON, _ := json.Marshal(metadata)

	p := &domain.Payment{
		RideID:      input.RideID,
		UserID:      input.UserID,
		Amount:      input.Amount,
		Currency:    input.Currency,
		Method:      input.Method,
		Provider:    input.Provider,
		Status:      domain.PaymentStatusPending,
		Description: input.Description,
		Metadata:    string(metadataJSON),
	}

	if err := uc.repo.Create(ctx, p); err != nil {
		if errors.Is(err, pg.ErrPaymentExists) {
			return nil, domain.ErrPaymentExists
		}
		return nil, err
	}

	// Cash payment — no gateway needed
	if input.Provider == domain.ProviderCash {
		return &domain.PaymentIntent{
			PaymentID:   p.ID,
			Provider:    input.Provider,
			Amount:      input.Amount,
			Currency:    input.Currency,
			Description: input.Description,
		}, nil
	}

	// Get gateway
	gw, ok := uc.gateways.Get(input.Provider)
	if !ok {
		return nil, domain.ErrInvalidProvider
	}

	// Create payment via gateway
	gwInput := gateway.CreatePaymentInput{
		PaymentID:   p.ID,
		Amount:      input.Amount,
		Currency:    input.Currency,
		Description: input.Description,
		ReturnURL:   input.ReturnURL,
		UserEmail:   input.UserEmail,
		UserPhone:   input.UserPhone,
		Metadata:    metadata,
		SaveCard:    input.SaveCard,
		TokenID:     input.TokenID,
	}

	result, err := gw.CreatePayment(ctx, gwInput)
	if err != nil {
		// Mark as failed
		p.Status = domain.PaymentStatusFailed
		p.FailReason = err.Error()
		uc.repo.UpdatePayment(ctx, p)
		return nil, err
	}

	// Update payment with gateway response
	p.ExternalID = result.ExternalID
	p.ConfirmURL = result.ConfirmURL
	p.Status = result.Status
	uc.repo.UpdatePayment(ctx, p)

	return &domain.PaymentIntent{
		PaymentID:   p.ID,
		ConfirmURL:  result.ConfirmURL,
		Provider:    input.Provider,
		Amount:      input.Amount,
		Currency:    input.Currency,
		Description: input.Description,
	}, nil
}

// GetByID returns payment by ID
func (uc *PaymentUseCase) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	return uc.repo.GetByID(ctx, id)
}

// GetByRideID returns payment by ride ID
func (uc *PaymentUseCase) GetByRideID(ctx context.Context, rideID string) (*domain.Payment, error) {
	return uc.repo.GetByRideID(ctx, rideID)
}

// ListByUser returns user's payments
func (uc *PaymentUseCase) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Payment, error) {
	if limit <= 0 {
		limit = 20
	}
	return uc.repo.ListByUser(ctx, userID, limit, offset)
}

// ConfirmPayment confirms cash payment (driver marks as received)
func (uc *PaymentUseCase) ConfirmPayment(ctx context.Context, id string) (*domain.Payment, error) {
	p, err := uc.repo.GetByID(ctx, id)
	if err != nil || p == nil {
		return nil, domain.ErrPaymentNotFound
	}
	if p.Status == domain.PaymentStatusCompleted {
		return p, nil
	}
	if p.Status != domain.PaymentStatusPending {
		return nil, domain.ErrPaymentNotPending
	}

	now := time.Now()
	p.Status = domain.PaymentStatusCompleted
	p.PaidAt = &now
	if err := uc.repo.UpdatePayment(ctx, p); err != nil {
		return nil, err
	}
	return uc.repo.GetByID(ctx, id)
}

// ProcessWebhook handles webhook from payment provider
func (uc *PaymentUseCase) ProcessWebhook(ctx context.Context, provider string, body []byte, signature string) error {
	gw, ok := uc.gateways.Get(provider)
	if !ok {
		return domain.ErrInvalidProvider
	}

	event, err := gw.ParseWebhook(ctx, body, signature)
	if err != nil {
		return err
	}

	// Find payment
	var p *domain.Payment
	if event.PaymentID != "" {
		p, _ = uc.repo.GetByID(ctx, event.PaymentID)
	}
	if p == nil && event.ExternalID != "" {
		p, _ = uc.repo.GetByExternalID(ctx, event.ExternalID)
	}
	if p == nil {
		return domain.ErrPaymentNotFound
	}

	// Update payment based on event
	switch event.EventType {
	case "payment.succeeded":
		now := time.Now()
		p.Status = domain.PaymentStatusCompleted
		p.PaidAt = &now

		// Try to save card if available
		if cardInfo, err := gw.GetSavedCard(ctx, event.ExternalID); err == nil && cardInfo != nil {
			pm := &domain.PaymentMethod{
				UserID:      p.UserID,
				Provider:    provider,
				Type:        "card",
				Last4:       cardInfo.Last4,
				Brand:       cardInfo.Brand,
				ExpiryMonth: cardInfo.ExpiryMonth,
				ExpiryYear:  cardInfo.ExpiryYear,
				TokenID:     cardInfo.TokenID,
			}
			uc.repo.CreatePaymentMethod(ctx, pm)
		}

	case "payment.failed":
		p.Status = domain.PaymentStatusFailed

	case "payment.cancelled":
		p.Status = domain.PaymentStatusCancelled

	case "refund.succeeded":
		now := time.Now()
		p.Status = domain.PaymentStatusRefunded
		p.RefundedAt = &now
	}

	return uc.repo.UpdatePayment(ctx, p)
}

// RefundPayment processes a refund
func (uc *PaymentUseCase) RefundPayment(ctx context.Context, req domain.RefundRequest) (*domain.RefundResult, error) {
	p, err := uc.repo.GetByID(ctx, req.PaymentID)
	if err != nil || p == nil {
		return nil, domain.ErrPaymentNotFound
	}

	// Can only refund completed payments
	if p.Status != domain.PaymentStatusCompleted {
		return nil, domain.ErrRefundNotAllowed
	}

	// Cash payments cannot be refunded via gateway
	if p.Provider == domain.ProviderCash {
		return nil, domain.ErrRefundNotAllowed
	}

	// Refund amount
	amount := req.Amount
	if amount <= 0 {
		amount = p.Amount // Full refund
	}

	// Get gateway
	gw, ok := uc.gateways.Get(p.Provider)
	if !ok {
		return nil, domain.ErrInvalidProvider
	}

	// Process refund
	refundInput := gateway.RefundInput{
		PaymentID:  p.ID,
		ExternalID: p.ExternalID,
		Amount:     amount,
		Reason:     req.Reason,
	}

	result, err := gw.Refund(ctx, refundInput)
	if err != nil {
		return nil, err
	}

	// Create refund record
	refundID, err := uc.repo.CreateRefund(ctx, p.ID, amount, req.Reason, result.RefundID)
	if err != nil {
		return nil, err
	}

	// Update refund status
	uc.repo.UpdateRefundStatus(ctx, refundID, result.Status)

	// Update payment if fully refunded
	if amount >= p.Amount {
		now := time.Now()
		p.Status = domain.PaymentStatusRefunded
		p.RefundedAt = &now
		uc.repo.UpdatePayment(ctx, p)
	}

	return &domain.RefundResult{
		RefundID:   refundID,
		PaymentID:  p.ID,
		Amount:     amount,
		Status:     result.Status,
		RefundedAt: time.Now(),
	}, nil
}

// --- Payment Methods ---

// ListPaymentMethods returns user's saved cards
func (uc *PaymentUseCase) ListPaymentMethods(ctx context.Context, userID string) ([]*domain.PaymentMethod, error) {
	return uc.repo.ListPaymentMethods(ctx, userID)
}

// DeletePaymentMethod removes a saved card
func (uc *PaymentUseCase) DeletePaymentMethod(ctx context.Context, id, userID string) error {
	return uc.repo.DeletePaymentMethod(ctx, id, userID)
}

// SetDefaultPaymentMethod sets a card as default
func (uc *PaymentUseCase) SetDefaultPaymentMethod(ctx context.Context, id, userID string) error {
	return uc.repo.SetDefaultPaymentMethod(ctx, id, userID)
}

// GetAvailableProviders returns list of available payment providers
func (uc *PaymentUseCase) GetAvailableProviders() []string {
	providers := []string{domain.ProviderCash}
	providers = append(providers, uc.gateways.Available()...)
	return providers
}
