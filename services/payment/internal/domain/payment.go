// Package domain — Payment bounded context: checkout, providers (Tinkoff, YooMoney, Sber)
package domain

import (
	"errors"
	"time"
)

// Payment methods
const (
	MethodCash = "cash"
	MethodCard = "card"
)

// Payment providers
const (
	ProviderCash    = "cash"    // Cash payment (no gateway)
	ProviderTinkoff = "tinkoff" // Tinkoff Acquiring
	ProviderYooMoney = "yoomoney" // YooMoney (ЮКасса)
	ProviderSber    = "sber"    // SberPay / Sberbank Acquiring
)

// Payment statuses
const (
	PaymentStatusPending    = "pending"    // Created, awaiting payment
	PaymentStatusProcessing = "processing" // Payment in progress
	PaymentStatusCompleted  = "completed"  // Successfully paid
	PaymentStatusFailed     = "failed"     // Payment failed
	PaymentStatusCancelled  = "cancelled"  // Cancelled by user/system
	PaymentStatusRefunded   = "refunded"   // Refunded to user
)

// Domain errors
var (
	ErrPaymentNotFound    = errors.New("payment not found")
	ErrInvalidAmount      = errors.New("amount must be positive")
	ErrInvalidMethod      = errors.New("invalid payment method")
	ErrInvalidProvider    = errors.New("invalid payment provider")
	ErrPaymentExists      = errors.New("payment already exists for this ride")
	ErrPaymentNotPending  = errors.New("payment is not in pending status")
	ErrRefundNotAllowed   = errors.New("refund not allowed for this payment")
	ErrProviderError      = errors.New("payment provider error")
)

// Payment — main payment entity
type Payment struct {
	ID          string    `json:"id"`
	RideID      string    `json:"ride_id"`
	UserID      string    `json:"user_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Method      string    `json:"method"`   // cash | card
	Provider    string    `json:"provider"` // cash | tinkoff | yoomoney | sber
	Status      string    `json:"status"`
	ExternalID  string    `json:"external_id,omitempty"`  // Provider's payment ID
	ConfirmURL  string    `json:"confirm_url,omitempty"`  // URL for 3DS / redirect
	Description string    `json:"description,omitempty"`
	Metadata    string    `json:"metadata,omitempty"`     // JSON metadata
	FailReason  string    `json:"fail_reason,omitempty"`
	RefundedAt  *time.Time `json:"refunded_at,omitempty"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PaymentMethod — saved payment method (card token)
type PaymentMethod struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Provider   string    `json:"provider"`
	Type       string    `json:"type"`             // card
	Last4      string    `json:"last4,omitempty"`  // Last 4 digits
	Brand      string    `json:"brand,omitempty"`  // visa, mastercard, mir
	ExpiryMonth int      `json:"expiry_month,omitempty"`
	ExpiryYear  int      `json:"expiry_year,omitempty"`
	TokenID    string    `json:"token_id"`         // Provider token for recurring
	IsDefault  bool      `json:"is_default"`
	CreatedAt  time.Time `json:"created_at"`
}

// PaymentIntent — intent to create payment (for frontend redirect flow)
type PaymentIntent struct {
	PaymentID   string `json:"payment_id"`
	ConfirmURL  string `json:"confirm_url"`  // URL to redirect user
	Provider    string `json:"provider"`
	Amount      float64 `json:"amount"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
}

// RefundRequest — refund parameters
type RefundRequest struct {
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount,omitempty"` // Partial refund amount (0 = full)
	Reason    string  `json:"reason,omitempty"`
}

// RefundResult — refund response
type RefundResult struct {
	RefundID   string    `json:"refund_id"`
	PaymentID  string    `json:"payment_id"`
	Amount     float64   `json:"amount"`
	Status     string    `json:"status"`
	RefundedAt time.Time `json:"refunded_at"`
}

// WebhookEvent — incoming webhook from provider
type WebhookEvent struct {
	Provider    string `json:"provider"`
	EventType   string `json:"event_type"`   // payment.succeeded, payment.failed, refund.succeeded
	PaymentID   string `json:"payment_id"`   // Our payment ID (from metadata)
	ExternalID  string `json:"external_id"`  // Provider's payment ID
	Status      string `json:"status"`
	Amount      float64 `json:"amount"`
	RawPayload  string `json:"raw_payload"`  // Original JSON
}

// IsValidProvider checks if provider is valid
func IsValidProvider(p string) bool {
	switch p {
	case ProviderCash, ProviderTinkoff, ProviderYooMoney, ProviderSber:
		return true
	}
	return false
}

// IsValidMethod checks if method is valid
func IsValidMethod(m string) bool {
	return m == MethodCash || m == MethodCard
}

// RequiresGateway returns true if payment needs external gateway
func RequiresGateway(provider string) bool {
	return provider != ProviderCash
}
