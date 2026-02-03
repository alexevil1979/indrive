// Package domain — Payment bounded context: checkout, cash/card stub
package domain

import "time"

const (
	MethodCash = "cash"
	MethodCard = "card"
)

const (
	PaymentStatusPending   = "pending"
	PaymentStatusCompleted = "completed"
	PaymentStatusFailed     = "failed"
)

type Payment struct {
	ID         string    `json:"id"`
	RideID     string    `json:"ride_id"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	Method     string    `json:"method"` // cash | card
	Status     string    `json:"status"`
	ExternalID string    `json:"external_id,omitempty"` // stub gateway id
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
