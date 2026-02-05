// Package gateway — Cash gateway (no external provider)
package gateway

import (
	"context"
	"fmt"

	"github.com/ridehail/payment/internal/domain"
)

// CashGateway — cash payment (no external gateway)
type CashGateway struct{}

// NewCashGateway creates cash gateway
func NewCashGateway() *CashGateway {
	return &CashGateway{}
}

// Provider returns provider name
func (g *CashGateway) Provider() string {
	return domain.ProviderCash
}

// CreatePayment for cash — immediately completed
func (g *CashGateway) CreatePayment(ctx context.Context, input CreatePaymentInput) (*CreatePaymentResult, error) {
	return &CreatePaymentResult{
		ExternalID:       fmt.Sprintf("cash_%s", input.PaymentID),
		Status:           domain.PaymentStatusPending, // Will be confirmed by driver
		RequiresRedirect: false,
	}, nil
}

// GetPaymentStatus for cash
func (g *CashGateway) GetPaymentStatus(ctx context.Context, externalID string) (string, error) {
	// Cash payments are always pending until driver confirms
	return domain.PaymentStatusPending, nil
}

// Refund for cash — not supported
func (g *CashGateway) Refund(ctx context.Context, input RefundInput) (*RefundResult, error) {
	return nil, fmt.Errorf("refund not supported for cash payments")
}

// ParseWebhook for cash — not applicable
func (g *CashGateway) ParseWebhook(ctx context.Context, body []byte, signature string) (*domain.WebhookEvent, error) {
	return nil, fmt.Errorf("webhooks not supported for cash payments")
}

// GetSavedCard for cash — not applicable
func (g *CashGateway) GetSavedCard(ctx context.Context, externalID string) (*CardInfo, error) {
	return nil, nil
}
