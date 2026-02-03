package domain

import "time"

// Profile — passenger/driver profile (user service aggregate)
type Profile struct {
	UserID      string
	DisplayName string
	Phone       string
	AvatarURL   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// DriverProfile — driver verification stub (doc upload later)
type DriverProfile struct {
	UserID         string
	Verified       bool
	LicenseNumber  string
	DocLicenseURL  string
	DocPhotoURL    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
