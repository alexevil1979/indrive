// Package usecase — OAuth use cases: OAuthLogin (create/link account)
package usecase

import (
	"context"
	"time"

	"github.com/ridehail/auth/internal/domain"
)

// OAuthRepository — persistence for OAuth accounts
type OAuthRepository interface {
	FindByProvider(ctx context.Context, provider, providerUserID string) (*domain.OAuthAccount, error)
	FindByUserAndProvider(ctx context.Context, userID, provider string) (*domain.OAuthAccount, error)
	Create(ctx context.Context, oa *domain.OAuthAccount) error
	Update(ctx context.Context, oa *domain.OAuthAccount) error
	ListByUser(ctx context.Context, userID string) ([]*domain.OAuthAccount, error)
	Delete(ctx context.Context, id string) error
}

// OAuthUseCase handles OAuth authentication flow.
type OAuthUseCase struct {
	userRepo  UserRepository
	oauthRepo OAuthRepository
	jwt       JWTIssuer
	hasher    PasswordHasher
}

// NewOAuthUseCase creates OAuth use case.
func NewOAuthUseCase(userRepo UserRepository, oauthRepo OAuthRepository, jwt JWTIssuer, hasher PasswordHasher) *OAuthUseCase {
	return &OAuthUseCase{
		userRepo:  userRepo,
		oauthRepo: oauthRepo,
		jwt:       jwt,
		hasher:    hasher,
	}
}

// OAuthLoginResult contains login result.
type OAuthLoginResult struct {
	User       *domain.User
	TokenPair  *TokenPair
	IsNewUser  bool
	OAuthEmail string
}

// Login handles OAuth login/registration flow.
// If OAuth account exists → login.
// If email exists but no OAuth → link and login.
// If neither → create new user and OAuth account.
func (uc *OAuthUseCase) Login(ctx context.Context, info *domain.OAuthUserInfo, accessToken, refreshToken string, expiresIn int) (*OAuthLoginResult, error) {
	// Check if OAuth account already exists
	existing, err := uc.oauthRepo.FindByProvider(ctx, info.Provider, info.ProviderUserID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		// OAuth account exists — update tokens and login
		existing.AccessToken = accessToken
		existing.RefreshToken = refreshToken
		existing.Email = info.Email
		existing.Name = info.Name
		existing.AvatarURL = info.AvatarURL
		if expiresIn > 0 {
			existing.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
		}
		if err := uc.oauthRepo.Update(ctx, existing); err != nil {
			return nil, err
		}

		// Get user and issue tokens
		user, err := uc.userRepo.ByID(ctx, existing.UserID)
		if err != nil || user == nil {
			return nil, domain.ErrOAuthFailed
		}

		tp, err := uc.issueTokenPair(user.ID, user.Role)
		if err != nil {
			return nil, err
		}

		return &OAuthLoginResult{
			User:       user,
			TokenPair:  tp,
			IsNewUser:  false,
			OAuthEmail: info.Email,
		}, nil
	}

	// OAuth account doesn't exist — check if user exists by email
	var user *domain.User
	var isNewUser bool

	if info.Email != "" {
		user, err = uc.userRepo.ByEmail(ctx, info.Email)
		if err != nil {
			return nil, err
		}
	}

	if user != nil {
		// User exists by email — check if already has this provider
		existingOAuth, _ := uc.oauthRepo.FindByUserAndProvider(ctx, user.ID, info.Provider)
		if existingOAuth != nil {
			// This shouldn't happen (different provider_user_id same user+provider)
			return nil, domain.ErrOAuthAccountExists
		}
		// Link OAuth account to existing user
		isNewUser = false
	} else {
		// Create new user
		randomPassword, _ := uc.hasher.Hash(info.ProviderUserID + info.Provider + time.Now().String())
		user = &domain.User{
			Email:        info.Email,
			PasswordHash: randomPassword, // OAuth users can't login with password
			Role:         "passenger",    // Default role
		}
		if err := uc.userRepo.Create(ctx, user); err != nil {
			// If email taken (race condition), try to get existing user
			if err == domain.ErrEmailTaken && info.Email != "" {
				user, err = uc.userRepo.ByEmail(ctx, info.Email)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
		isNewUser = true
	}

	// Create OAuth account link
	oauthAccount := &domain.OAuthAccount{
		UserID:         user.ID,
		Provider:       info.Provider,
		ProviderUserID: info.ProviderUserID,
		Email:          info.Email,
		Name:           info.Name,
		AvatarURL:      info.AvatarURL,
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
	}
	if expiresIn > 0 {
		oauthAccount.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	}

	if err := uc.oauthRepo.Create(ctx, oauthAccount); err != nil {
		return nil, err
	}

	// Issue JWT tokens
	tp, err := uc.issueTokenPair(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &OAuthLoginResult{
		User:       user,
		TokenPair:  tp,
		IsNewUser:  isNewUser,
		OAuthEmail: info.Email,
	}, nil
}

// ListLinkedAccounts returns OAuth accounts linked to user.
func (uc *OAuthUseCase) ListLinkedAccounts(ctx context.Context, userID string) ([]*domain.OAuthAccount, error) {
	return uc.oauthRepo.ListByUser(ctx, userID)
}

// UnlinkAccount removes OAuth account link.
func (uc *OAuthUseCase) UnlinkAccount(ctx context.Context, userID, oauthAccountID string) error {
	oa, err := uc.oauthRepo.FindByProvider(ctx, "", "")
	if err != nil {
		return err
	}
	// Verify ownership
	accounts, err := uc.oauthRepo.ListByUser(ctx, userID)
	if err != nil {
		return err
	}

	var found bool
	for _, acc := range accounts {
		if acc.ID == oauthAccountID {
			found = true
			break
		}
	}
	if !found {
		return domain.ErrOAuthFailed
	}

	// Don't allow unlinking if it's the only auth method and no password
	user, err := uc.userRepo.ByID(ctx, userID)
	if err != nil {
		return err
	}
	if len(accounts) <= 1 && user.PasswordHash == "" {
		return domain.ErrOAuthFailed // Can't remove last auth method
	}

	_ = oa // unused
	return uc.oauthRepo.Delete(ctx, oauthAccountID)
}

func (uc *OAuthUseCase) issueTokenPair(userID, role string) (*TokenPair, error) {
	access, err := uc.jwt.IssueAccess(userID, role)
	if err != nil {
		return nil, err
	}
	refresh, err := uc.jwt.IssueRefresh(userID)
	if err != nil {
		return nil, err
	}
	expiresIn := int64(24 * 3600)
	return &TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: expiresIn}, nil
}
