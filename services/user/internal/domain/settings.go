package domain

import "time"

// MapProvider represents map service provider
type MapProvider string

const (
	MapProviderGoogle MapProvider = "google"
	MapProviderYandex MapProvider = "yandex"
)

// AppSettings represents application-wide settings
type AppSettings struct {
	ID          string      `json:"id"`
	MapProvider MapProvider `json:"map_provider"`
	// Google Maps settings
	GoogleMapsAPIKey string `json:"google_maps_api_key,omitempty"`
	// Yandex Maps settings
	YandexMapsAPIKey string `json:"yandex_maps_api_key,omitempty"`
	// Additional settings
	DefaultLanguage string    `json:"default_language"`
	DefaultCurrency string    `json:"default_currency"`
	UpdatedAt       time.Time `json:"updated_at"`
	UpdatedBy       string    `json:"updated_by,omitempty"`
}

// MapSettings represents map-specific settings for mobile apps
type MapSettings struct {
	Provider MapProvider `json:"provider"`
	APIKey   string      `json:"api_key"`
}

// Validate validates MapProvider value
func (p MapProvider) Validate() bool {
	return p == MapProviderGoogle || p == MapProviderYandex
}
