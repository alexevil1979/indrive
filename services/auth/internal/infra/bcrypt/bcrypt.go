// Package bcrypt — password hashing (OWASP: bcrypt cost 12)
package bcrypt

import (
	"golang.org/x/crypto/bcrypt"
)

const DefaultCost = 12

type Hasher struct {
	cost int
}

func New(cost int) *Hasher {
	if cost <= 0 {
		cost = DefaultCost
	}
	return &Hasher{cost: cost}
}

func (h *Hasher) Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (h *Hasher) Verify(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
