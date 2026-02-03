// Package jwt — JWT issue/validate (access + refresh). 2026: golang-jwt/jwt/v5.
// Production: use RS256 or store secrets in vault; short access TTL.
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"uid"`
	Role   string `json:"role"`
}

type Manager struct {
	secret        []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// New returns a JWT manager. accessTTL/refreshTTL for token expiry.
func New(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// IssueAccess creates an access token for user.
func (m *Manager) IssueAccess(userID, role string) (string, error) {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ridehail-auth",
		},
		UserID: userID,
		Role:   role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// IssueRefresh creates a refresh token (opaque or JWT; we use JWT for simplicity).
func (m *Manager) IssueRefresh(userID string) (string, error) {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ridehail-auth",
		},
		UserID: userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ValidateAccess parses and validates access token; returns claims or ErrInvalidToken.
func (m *Manager) ValidateAccess(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// ValidateRefresh parses refresh token and returns userID or error.
func (m *Manager) ValidateRefresh(tokenString string) (userID string, err error) {
	claims, err := m.ValidateAccess(tokenString) // same structure, no role required
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}
