// Package oauth — OAuth2 provider clients (Google, Yandex, VK)
package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ridehail/auth/internal/domain"
)

var (
	ErrInvalidState = errors.New("invalid oauth state")
	ErrTokenFailed  = errors.New("failed to exchange token")
	ErrUserInfo     = errors.New("failed to get user info")
)

// Config holds OAuth provider configuration.
type Config struct {
	Google GoogleConfig
	Yandex YandexConfig
	VK     VKConfig
}

// GoogleConfig for Google OAuth2.
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// YandexConfig for Yandex OAuth2.
type YandexConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// VKConfig for VK OAuth2.
type VKConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// TokenResponse from OAuth provider.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Provider interface for OAuth providers.
type Provider interface {
	AuthURL(state string) string
	Exchange(ctx context.Context, code string) (*TokenResponse, error)
	UserInfo(ctx context.Context, accessToken string) (*domain.OAuthUserInfo, error)
	Name() string
}

// Manager manages multiple OAuth providers.
type Manager struct {
	providers map[string]Provider
}

// NewManager creates OAuth manager with configured providers.
func NewManager(cfg Config) *Manager {
	m := &Manager{
		providers: make(map[string]Provider),
	}

	if cfg.Google.ClientID != "" {
		m.providers[domain.ProviderGoogle] = NewGoogleProvider(cfg.Google)
	}
	if cfg.Yandex.ClientID != "" {
		m.providers[domain.ProviderYandex] = NewYandexProvider(cfg.Yandex)
	}
	if cfg.VK.ClientID != "" {
		m.providers[domain.ProviderVK] = NewVKProvider(cfg.VK)
	}

	return m
}

// GetProvider returns provider by name or nil.
func (m *Manager) GetProvider(name string) Provider {
	return m.providers[name]
}

// HasProvider checks if provider is configured.
func (m *Manager) HasProvider(name string) bool {
	_, ok := m.providers[name]
	return ok
}

// ListProviders returns list of configured provider names.
func (m *Manager) ListProviders() []string {
	var names []string
	for name := range m.providers {
		names = append(names, name)
	}
	return names
}

// --- Google ---

type googleProvider struct {
	cfg GoogleConfig
}

func NewGoogleProvider(cfg GoogleConfig) Provider {
	return &googleProvider{cfg: cfg}
}

func (p *googleProvider) Name() string { return domain.ProviderGoogle }

func (p *googleProvider) AuthURL(state string) string {
	params := url.Values{
		"client_id":     {p.cfg.ClientID},
		"redirect_uri":  {p.cfg.RedirectURL},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
		"access_type":   {"offline"},
		"prompt":        {"consent"},
	}
	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

func (p *googleProvider) Exchange(ctx context.Context, code string) (*TokenResponse, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {p.cfg.ClientID},
		"client_secret": {p.cfg.ClientSecret},
		"redirect_uri":  {p.cfg.RedirectURL},
		"grant_type":    {"authorization_code"},
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: %s", ErrTokenFailed, string(body))
	}

	var tr TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

func (p *googleProvider) UserInfo(ctx context.Context, accessToken string) (*domain.OAuthUserInfo, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrUserInfo
	}

	var info struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &domain.OAuthUserInfo{
		ProviderUserID: info.ID,
		Email:          info.Email,
		Name:           info.Name,
		AvatarURL:      info.Picture,
		Provider:       domain.ProviderGoogle,
	}, nil
}

// --- Yandex ---

type yandexProvider struct {
	cfg YandexConfig
}

func NewYandexProvider(cfg YandexConfig) Provider {
	return &yandexProvider{cfg: cfg}
}

func (p *yandexProvider) Name() string { return domain.ProviderYandex }

func (p *yandexProvider) AuthURL(state string) string {
	params := url.Values{
		"client_id":     {p.cfg.ClientID},
		"redirect_uri":  {p.cfg.RedirectURL},
		"response_type": {"code"},
		"state":         {state},
	}
	return "https://oauth.yandex.ru/authorize?" + params.Encode()
}

func (p *yandexProvider) Exchange(ctx context.Context, code string) (*TokenResponse, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {p.cfg.ClientID},
		"client_secret": {p.cfg.ClientSecret},
		"grant_type":    {"authorization_code"},
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://oauth.yandex.ru/token", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: %s", ErrTokenFailed, string(body))
	}

	var tr TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

func (p *yandexProvider) UserInfo(ctx context.Context, accessToken string) (*domain.OAuthUserInfo, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://login.yandex.ru/info?format=json", nil)
	req.Header.Set("Authorization", "OAuth "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrUserInfo
	}

	var info struct {
		ID            string `json:"id"`
		DefaultEmail  string `json:"default_email"`
		DisplayName   string `json:"display_name"`
		RealName      string `json:"real_name"`
		DefaultAvatar string `json:"default_avatar_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	name := info.DisplayName
	if name == "" {
		name = info.RealName
	}

	avatarURL := ""
	if info.DefaultAvatar != "" {
		avatarURL = fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-200", info.DefaultAvatar)
	}

	return &domain.OAuthUserInfo{
		ProviderUserID: info.ID,
		Email:          info.DefaultEmail,
		Name:           name,
		AvatarURL:      avatarURL,
		Provider:       domain.ProviderYandex,
	}, nil
}

// --- VK ---

type vkProvider struct {
	cfg VKConfig
}

func NewVKProvider(cfg VKConfig) Provider {
	return &vkProvider{cfg: cfg}
}

func (p *vkProvider) Name() string { return domain.ProviderVK }

func (p *vkProvider) AuthURL(state string) string {
	params := url.Values{
		"client_id":     {p.cfg.ClientID},
		"redirect_uri":  {p.cfg.RedirectURL},
		"response_type": {"code"},
		"scope":         {"email"},
		"state":         {state},
		"v":             {"5.131"},
	}
	return "https://oauth.vk.com/authorize?" + params.Encode()
}

func (p *vkProvider) Exchange(ctx context.Context, code string) (*TokenResponse, error) {
	params := url.Values{
		"code":          {code},
		"client_id":     {p.cfg.ClientID},
		"client_secret": {p.cfg.ClientSecret},
		"redirect_uri":  {p.cfg.RedirectURL},
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://oauth.vk.com/access_token?"+params.Encode(), nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: %s", ErrTokenFailed, string(body))
	}

	var vkResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		UserID      int    `json:"user_id"`
		Email       string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&vkResp); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken: vkResp.AccessToken,
		ExpiresIn:   vkResp.ExpiresIn,
	}, nil
}

func (p *vkProvider) UserInfo(ctx context.Context, accessToken string) (*domain.OAuthUserInfo, error) {
	params := url.Values{
		"access_token": {accessToken},
		"fields":       {"photo_200,email"},
		"v":            {"5.131"},
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.vk.com/method/users.get?"+params.Encode(), nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrUserInfo
	}

	var vkResp struct {
		Response []struct {
			ID        int    `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Photo200  string `json:"photo_200"`
		} `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&vkResp); err != nil {
		return nil, err
	}

	if len(vkResp.Response) == 0 {
		return nil, ErrUserInfo
	}

	user := vkResp.Response[0]
	return &domain.OAuthUserInfo{
		ProviderUserID: fmt.Sprintf("%d", user.ID),
		Email:          "", // VK doesn't always return email in users.get
		Name:           strings.TrimSpace(user.FirstName + " " + user.LastName),
		AvatarURL:      user.Photo200,
		Provider:       domain.ProviderVK,
	}, nil
}

// StateManager handles OAuth state generation and validation.
type StateManager struct {
	// In production, use Redis or signed tokens
	// For now, simple in-memory with TTL
	states map[string]stateEntry
}

type stateEntry struct {
	provider  string
	expiresAt time.Time
}

func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[string]stateEntry),
	}
}

// Generate creates a new state token.
func (sm *StateManager) Generate(provider string) string {
	// Simple random state; in production use crypto/rand
	state := fmt.Sprintf("%s_%d", provider, time.Now().UnixNano())
	sm.states[state] = stateEntry{
		provider:  provider,
		expiresAt: time.Now().Add(10 * time.Minute),
	}
	return state
}

// Validate checks state and returns provider name.
func (sm *StateManager) Validate(state string) (string, error) {
	entry, ok := sm.states[state]
	if !ok {
		return "", ErrInvalidState
	}
	delete(sm.states, state)
	if time.Now().After(entry.expiresAt) {
		return "", ErrInvalidState
	}
	return entry.provider, nil
}
