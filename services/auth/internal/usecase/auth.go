// Package usecase — Auth use cases: Register, Login, Refresh
package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/ridehail/auth/internal/domain"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

// AuthUseCase — register, login, refresh
type AuthUseCase struct {
	repo   UserRepository
	jwt    JWTIssuer
	hasher PasswordHasher
}

// UserRepository — persistence for identity
type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	ByEmail(ctx context.Context, email string) (*domain.User, error)
	ByID(ctx context.Context, id string) (*domain.User, error)
}

// JWTIssuer — issue access/refresh tokens
type JWTIssuer interface {
	IssueAccess(userID, role string) (string, error)
	IssueRefresh(userID string) (string, error)
	ValidateRefresh(tokenString string) (userID string, err error)
}

// PasswordHasher — hash and verify passwords (bcrypt)
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) bool
}

// TokenPair — access + refresh
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
}

func NewAuthUseCase(repo UserRepository, jwt JWTIssuer, hasher PasswordHasher) *AuthUseCase {
	return &AuthUseCase{repo: repo, jwt: jwt, hasher: hasher}
}

// Register — validate, hash password, create user, return tokens
func (uc *AuthUseCase) Register(ctx context.Context, email, password, role string) (*domain.User, *TokenPair, error) {
	if role == "" {
		role = "passenger"
	}
	if role != "passenger" && role != "driver" && role != "admin" {
		return nil, nil, fmt.Errorf("invalid role: %s", role)
	}
	hash, err := uc.hasher.Hash(password)
	if err != nil {
		return nil, nil, err
	}
	u := &domain.User{
		Email:        email,
		PasswordHash: hash,
		Role:         role,
	}
	if err := uc.repo.Create(ctx, u); err != nil {
		if errors.Is(err, domain.ErrEmailTaken) {
			return nil, nil, domain.ErrEmailTaken
		}
		return nil, nil, err
	}
	tp, err := uc.issueTokenPair(u.ID, u.Role)
	if err != nil {
		return u, nil, err
	}
	return u, tp, nil
}

// Login — verify credentials, return tokens
func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	u, err := uc.repo.ByEmail(ctx, email)
	if err != nil || u == nil {
		return nil, ErrInvalidCredentials
	}
	if !uc.hasher.Verify(u.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}
	return uc.issueTokenPair(u.ID, u.Role)
}

// Refresh — validate refresh token, return new access token
func (uc *AuthUseCase) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	userID, err := uc.jwt.ValidateRefresh(refreshToken)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	u, err := uc.repo.ByID(ctx, userID)
	if err != nil || u == nil {
		return nil, ErrInvalidCredentials
	}
	return uc.issueTokenPair(u.ID, u.Role)
}

func (uc *AuthUseCase) issueTokenPair(userID, role string) (*TokenPair, error) {
	access, err := uc.jwt.IssueAccess(userID, role)
	if err != nil {
		return nil, err
	}
	refresh, err := uc.jwt.IssueRefresh(userID)
	if err != nil {
		return nil, err
	}
	// ExpiresIn: 24h in seconds
	expiresIn := int64(24 * 3600)
	return &TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: expiresIn}, nil
}
