// Package domain — Geolocation bounded context: driver position, nearest search
package domain

// Location — lat/lng (WGS84)
type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// DriverLocation — driver id + position (Redis GEO member)
type DriverLocation struct {
	DriverID string   `json:"driver_id"`
	Location Location `json:"location"`
	Distance float64  `json:"distance,omitempty"` // km, filled on nearest search
}

// NearestQuery — input for search
type NearestQuery struct {
	Lat      float64
	Lng      float64
	RadiusKM float64
	Limit    int
}
