// Package domain — Promo codes for discounts
package domain

import "time"

// PromoType defines the discount type
type PromoType string

const (
	PromoTypePercent PromoType = "percent" // e.g. 10% off
	PromoTypeFixed   PromoType = "fixed"   // e.g. 100 RUB off
)

// Promo represents a promotional code
type Promo struct {
	ID            string    `json:"id"`
	Code          string    `json:"code"`           // Unique promo code (e.g. "WELCOME10")
	Description   string    `json:"description"`    // Human-readable description
	Type          PromoType `json:"type"`           // percent or fixed
	Value         float64   `json:"value"`          // Discount value (10 for 10% or 100 for 100 RUB)
	MinOrderValue float64   `json:"min_order_value"` // Minimum order to apply (0 = no minimum)
	MaxDiscount   float64   `json:"max_discount"`   // Max discount for percent type (0 = no limit)
	UsageLimit    int       `json:"usage_limit"`    // Total uses allowed (0 = unlimited)
	UsageCount    int       `json:"usage_count"`    // Current usage count
	PerUserLimit  int       `json:"per_user_limit"` // Uses per user (0 = unlimited)
	IsActive      bool      `json:"is_active"`
	StartsAt      time.Time `json:"starts_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserPromo tracks user's promo usage
type UserPromo struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	PromoID   string    `json:"promo_id"`
	RideID    string    `json:"ride_id,omitempty"` // Ride where promo was applied
	Discount  float64   `json:"discount"`          // Actual discount amount applied
	UsedAt    time.Time `json:"used_at"`
}

// PromoResult is the result of applying a promo code
type PromoResult struct {
	Valid       bool    `json:"valid"`
	Promo       *Promo  `json:"promo,omitempty"`
	Discount    float64 `json:"discount"`     // Calculated discount amount
	FinalPrice  float64 `json:"final_price"`  // Price after discount
	Error       string  `json:"error,omitempty"`
}

// IsValid checks if promo can be used
func (p *Promo) IsValid() bool {
	now := time.Now()
	
	if !p.IsActive {
		return false
	}
	
	if now.Before(p.StartsAt) {
		return false
	}
	
	if !p.ExpiresAt.IsZero() && now.After(p.ExpiresAt) {
		return false
	}
	
	if p.UsageLimit > 0 && p.UsageCount >= p.UsageLimit {
		return false
	}
	
	return true
}

// CalculateDiscount calculates the discount for a given order amount
func (p *Promo) CalculateDiscount(orderAmount float64) float64 {
	if orderAmount < p.MinOrderValue {
		return 0
	}
	
	var discount float64
	
	switch p.Type {
	case PromoTypePercent:
		discount = orderAmount * (p.Value / 100)
		if p.MaxDiscount > 0 && discount > p.MaxDiscount {
			discount = p.MaxDiscount
		}
	case PromoTypeFixed:
		discount = p.Value
	}
	
	// Discount cannot exceed order amount
	if discount > orderAmount {
		discount = orderAmount
	}
	
	return discount
}
