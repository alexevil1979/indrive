package usecase

import (
	"context"
	"testing"
)

func TestAuthUseCase_Register_InvalidRole(t *testing.T) {
	// Stub: full test would use mock repo/jwt/hasher
	uc := &AuthUseCase{}
	_, _, err := uc.Register(context.Background(), "a@b.com", "password123", "invalid_role")
	if err == nil {
		t.Error("expected error for invalid role")
	}
}
