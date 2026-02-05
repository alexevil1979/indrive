// Package http — OAuth HTTP handlers: redirect to provider, callback
package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ridehail/auth/internal/domain"
	"github.com/ridehail/auth/internal/infra/oauth"
	"github.com/ridehail/auth/internal/usecase"
)

// OAuthHandler handles OAuth routes.
type OAuthHandler struct {
	oauthMgr     *oauth.Manager
	stateMgr     *oauth.StateManager
	oauthUC      *usecase.OAuthUseCase
	frontendURL  string // Where to redirect after OAuth (with token in query)
}

// NewOAuthHandler creates OAuth handler.
func NewOAuthHandler(oauthMgr *oauth.Manager, stateMgr *oauth.StateManager, oauthUC *usecase.OAuthUseCase, frontendURL string) *OAuthHandler {
	return &OAuthHandler{
		oauthMgr:    oauthMgr,
		stateMgr:    stateMgr,
		oauthUC:     oauthUC,
		frontendURL: frontendURL,
	}
}

// ListProviders returns configured OAuth providers.
// GET /auth/oauth/providers
func (h *OAuthHandler) ListProviders() echo.HandlerFunc {
	return func(c echo.Context) error {
		providers := h.oauthMgr.ListProviders()
		return c.JSON(http.StatusOK, map[string]interface{}{
			"providers": providers,
		})
	}
}

// Redirect redirects user to OAuth provider.
// GET /auth/oauth/:provider
func (h *OAuthHandler) Redirect() echo.HandlerFunc {
	return func(c echo.Context) error {
		providerName := c.Param("provider")

		if !domain.IsValidProvider(providerName) {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "unsupported provider",
			})
		}

		provider := h.oauthMgr.GetProvider(providerName)
		if provider == nil {
			return c.JSON(http.StatusNotImplemented, map[string]string{
				"error":   "provider not configured",
				"message": "OAuth for " + providerName + " is not configured. Set environment variables.",
			})
		}

		// Generate state token
		state := h.stateMgr.Generate(providerName)

		// Redirect to provider
		authURL := provider.AuthURL(state)
		return c.Redirect(http.StatusTemporaryRedirect, authURL)
	}
}

// Callback handles OAuth callback from provider.
// GET /auth/oauth/:provider/callback?code=...&state=...
func (h *OAuthHandler) Callback() echo.HandlerFunc {
	return func(c echo.Context) error {
		providerName := c.Param("provider")
		code := c.QueryParam("code")
		state := c.QueryParam("state")
		errorParam := c.QueryParam("error")

		// Handle OAuth error from provider
		if errorParam != "" {
			errorDesc := c.QueryParam("error_description")
			return h.redirectWithError(c, "oauth_error", errorDesc)
		}

		if code == "" {
			return h.redirectWithError(c, "missing_code", "Authorization code not provided")
		}

		// Validate state
		validatedProvider, err := h.stateMgr.Validate(state)
		if err != nil {
			return h.redirectWithError(c, "invalid_state", "Invalid or expired state")
		}

		if validatedProvider != providerName {
			return h.redirectWithError(c, "provider_mismatch", "Provider mismatch")
		}

		provider := h.oauthMgr.GetProvider(providerName)
		if provider == nil {
			return h.redirectWithError(c, "provider_not_configured", "Provider not configured")
		}

		ctx := c.Request().Context()

		// Exchange code for tokens
		tokenResp, err := provider.Exchange(ctx, code)
		if err != nil {
			return h.redirectWithError(c, "token_exchange_failed", err.Error())
		}

		// Get user info from provider
		userInfo, err := provider.UserInfo(ctx, tokenResp.AccessToken)
		if err != nil {
			return h.redirectWithError(c, "userinfo_failed", err.Error())
		}

		// Process OAuth login (create/link account, issue JWT)
		result, err := h.oauthUC.Login(ctx, userInfo, tokenResp.AccessToken, tokenResp.RefreshToken, tokenResp.ExpiresIn)
		if err != nil {
			return h.redirectWithError(c, "login_failed", err.Error())
		}

		// Redirect to frontend with tokens
		return h.redirectWithTokens(c, result)
	}
}

// redirectWithError redirects to frontend with error.
func (h *OAuthHandler) redirectWithError(c echo.Context, errCode, errMsg string) error {
	if h.frontendURL == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":       errCode,
			"description": errMsg,
		})
	}
	return c.Redirect(http.StatusTemporaryRedirect,
		h.frontendURL+"?error="+errCode+"&error_description="+errMsg)
}

// redirectWithTokens redirects to frontend with tokens.
func (h *OAuthHandler) redirectWithTokens(c echo.Context, result *usecase.OAuthLoginResult) error {
	if h.frontendURL == "" {
		// No frontend URL configured — return JSON
		return c.JSON(http.StatusOK, TokenResponse{
			AccessToken:  result.TokenPair.AccessToken,
			RefreshToken: result.TokenPair.RefreshToken,
			ExpiresIn:    result.TokenPair.ExpiresIn,
			UserID:       result.User.ID,
			Role:         result.User.Role,
		})
	}

	// Redirect to frontend with tokens in fragment (safer than query params)
	// Frontend should extract tokens from URL hash
	redirectURL := h.frontendURL +
		"#access_token=" + result.TokenPair.AccessToken +
		"&refresh_token=" + result.TokenPair.RefreshToken +
		"&expires_in=" + string(rune(result.TokenPair.ExpiresIn)) +
		"&user_id=" + result.User.ID +
		"&is_new=" + boolStr(result.IsNewUser)

	return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
