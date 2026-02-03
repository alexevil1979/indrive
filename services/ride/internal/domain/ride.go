// Package domain — Ride bounded context: request, bidding, matching, status
package domain

import "time"

// RideStatus — inDrive-style flow
const (
	StatusRequested  = "requested"
	StatusBidding    = "bidding"
	StatusMatched    = "matched"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusCancelled  = "cancelled"
)

type Point struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Address string  `json:"address,omitempty"`
}

type Ride struct {
	ID          string    `json:"id"`
	PassengerID string    `json:"passenger_id"`
	DriverID    string    `json:"driver_id,omitempty"`
	Status      string    `json:"status"`
	From        Point     `json:"from"`
	To          Point     `json:"to"`
	Price       *float64  `json:"price,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Bid struct {
	ID        string    `json:"id"`
	RideID    string    `json:"ride_id"`
	DriverID  string    `json:"driver_id"`
	Price     float64   `json:"price"`
	Status    string    `json:"status"` // pending, accepted, rejected
	CreatedAt time.Time `json:"created_at"`
}
