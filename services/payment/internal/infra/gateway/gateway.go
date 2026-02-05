// Package gateway — payment gateway interface and implementations
package gateway

import (
	"context"
	"errors"

	"github.com/ridehail/payment/internal/domain"
)

// Common gateway errors
var (
	ErrProviderUnavailable = errors.New("payment provider unavailable")
	ErrInvalidCredentials  = errors.New("invalid provider credentials")
	ErrPaymentRejected     = errors.New("payment rejected by provider")
	ErrInvalidWebhook      = errors.New("invalid webhook signature")
)

// CreatePaymentInput — input for creating payment
type CreatePaymentInput struct {
	PaymentID   string  // Our internal payment ID
	Amount      float64
	Currency    string
	Description string
	ReturnURL   string // URL to redirect after payment
	UserEmail   string // Optional: for receipts
	UserPhone   string // Optional: for receipts
	Metadata    map[string]string
	SaveCard    bool   // Request to save card for future payments
	TokenID     string // Use saved card token
}

// CreatePaymentResult — result from provider
type CreatePaymentResult struct {
	ExternalID  string // Provider's payment ID
	ConfirmURL  string // URL for redirect (3DS, payment form)
	Status      string // pending, processing, completed, failed
	RequiresRedirect bool
}

// ConfirmPaymentInput — input for confirming payment (3DS, etc.)
type ConfirmPaymentInput struct {
	PaymentID  string
	ExternalID string
}

// RefundInput — input for refund
type RefundInput struct {
	PaymentID  string
	ExternalID string
	Amount     float64 // 0 = full refund
	Reason     string
}

// RefundResult — result from provider
type RefundResult struct {
	RefundID   string
	ExternalID string
	Amount     float64
	Status     string
}

// CardInfo — saved card information from provider
type CardInfo struct {
	TokenID     string
	Last4       string
	Brand       string
	ExpiryMonth int
	ExpiryYear  int
}

// Gateway — payment gateway interface
type Gateway interface {
	// Provider returns the provider name
	Provider() string

	// CreatePayment initiates a payment
	CreatePayment(ctx context.Context, input CreatePaymentInput) (*CreatePaymentResult, error)

	// GetPaymentStatus gets current payment status
	GetPaymentStatus(ctx context.Context, externalID string) (string, error)

	// Refund processes a refund
	Refund(ctx context.Context, input RefundInput) (*RefundResult, error)

	// ParseWebhook parses and validates incoming webhook
	ParseWebhook(ctx context.Context, body []byte, signature string) (*domain.WebhookEvent, error)

	// GetSavedCard returns saved card info from successful payment
	GetSavedCard(ctx context.Context, externalID string) (*CardInfo, error)
}

// Manager manages multiple payment gateways
type Manager struct {
	gateways map[string]Gateway
}

// NewManager creates a gateway manager
func NewManager() *Manager {
	return &Manager{
		gateways: make(map[string]Gateway),
	}
}

// Register registers a gateway
func (m *Manager) Register(g Gateway) {
	m.gateways[g.Provider()] = g
}

// Get returns gateway by provider name
func (m *Manager) Get(provider string) (Gateway, bool) {
	g, ok := m.gateways[provider]
	return g, ok
}

// Available returns list of available providers
func (m *Manager) Available() []string {
	providers := make([]string, 0, len(m.gateways))
	for p := range m.gateways {
		providers = append(providers, p)
	}
	return providers
}
