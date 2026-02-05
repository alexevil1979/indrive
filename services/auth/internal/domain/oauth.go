// Package domain — OAuth entities
package domain

import (
	"errors"
	"time"
)

// Supported OAuth providers
const (
	ProviderGoogle = "google"
	ProviderYandex = "yandex"
	ProviderVK     = "vk"
)

var (
	ErrUnsupportedProvider = errors.New("unsupported oauth provider")
	ErrOAuthFailed         = errors.New("oauth authentication failed")
	ErrOAuthAccountExists  = errors.New("oauth account already linked to another user")
)

// OAuthAccount — link between OAuth provider and internal user
type OAuthAccount struct {
	ID             string
	UserID         string
	Provider       string // google, yandex, vk
	ProviderUserID string // external user ID from provider
	Email          string // email from provider (may differ from User.Email)
	Name           string // display name from provider
	AvatarURL      string
	AccessToken    string // encrypted; for API calls if needed
	RefreshToken   string // encrypted; for token refresh
	ExpiresAt      time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// OAuthUserInfo — user info returned by OAuth provider
type OAuthUserInfo struct {
	ProviderUserID string
	Email          string
	Name           string
	AvatarURL      string
	Provider       string
}

// IsValidProvider checks if provider is supported
func IsValidProvider(provider string) bool {
	switch provider {
	case ProviderGoogle, ProviderYandex, ProviderVK:
		return true
	}
	return false
}
