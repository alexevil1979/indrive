// Package domain — Auth bounded context: identity entities
package domain

import (
	"errors"
	"time"
)

var ErrEmailTaken = errors.New("email already taken")

// User — identity aggregate (email, password_hash, role)
type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string // passenger | driver | admin
	CreatedAt    time.Time
}
