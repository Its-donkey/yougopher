package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Google OAuth 2.0 endpoints.
const (
	DefaultAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	DefaultTokenURL = "https://oauth2.googleapis.com/token"
)

// YouTube API scopes.
const (
	// ScopeLiveChat grants read access to live chat messages.
	ScopeLiveChat = "https://www.googleapis.com/auth/youtube"

	// ScopeLiveChatModerate grants moderator access to live chat.
	ScopeLiveChatModerate = "https://www.googleapis.com/auth/youtube.force-ssl"

	// ScopeReadOnly grants read-only access to YouTube account.
	ScopeReadOnly = "https://www.googleapis.com/auth/youtube.readonly"

	// ScopeUpload grants access to upload videos and manage playlists.
	ScopeUpload = "https://www.googleapis.com/auth/youtube.upload"

	// ScopePartner grants access to YouTube Analytics.
	ScopePartner = "https://www.googleapis.com/auth/youtubepartner"

	// ScopePartnerChannelAudit grants access to YouTube Analytics monetary reports.
	ScopePartnerChannelAudit = "https://www.googleapis.com/auth/youtubepartner-channel-audit"
)

// Config holds OAuth 2.0 configuration.
type Config struct {
	// ClientID is the OAuth client ID.
	ClientID string

	// ClientSecret is the OAuth client secret.
	ClientSecret string

	// RedirectURL is the callback URL for the OAuth flow.
	RedirectURL string

	// Scopes are the requested OAuth scopes.
	Scopes []string

	// AuthURL is the authorization endpoint (defaults to Google's).
	AuthURL string

	// TokenURL is the token endpoint (defaults to Google's).
	TokenURL string
}

// Lifecycle states.
const (
	stateStopped  int32 = 0
	stateRunning  int32 = 1
	stateStopping int32 = 2
)

// AuthClient handles OAuth 2.0 authentication for YouTube APIs.
type AuthClient struct {
	config     Config
	httpClient *http.Client

	// Token management
	mu    sync.RWMutex
	token *Token

	// Auto-refresh
	state        atomic.Int32
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	refreshEarly time.Duration // How early to refresh before expiry

	// Callbacks
	onTokenRefresh func(*Token)
	onRefreshError func(error)
}

// AuthClientOption configures an AuthClient.
type AuthClientOption func(*AuthClient)

// NewAuthClient creates a new OAuth client.
func NewAuthClient(config Config, opts ...AuthClientOption) *AuthClient {
	if config.AuthURL == "" {
		config.AuthURL = DefaultAuthURL
	}
	if config.TokenURL == "" {
		config.TokenURL = DefaultTokenURL
	}

	c := &AuthClient{
		config:       config,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		refreshEarly: 5 * time.Minute,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) AuthClientOption {
	return func(c *AuthClient) { c.httpClient = hc }
}

// WithToken sets an initial token.
func WithToken(token *Token) AuthClientOption {
	return func(c *AuthClient) { c.token = token }
}

// WithRefreshEarly sets how early to refresh tokens before expiry.
func WithRefreshEarly(d time.Duration) AuthClientOption {
	return func(c *AuthClient) { c.refreshEarly = d }
}

// WithOnTokenRefresh sets a callback for successful token refreshes.
func WithOnTokenRefresh(fn func(*Token)) AuthClientOption {
	return func(c *AuthClient) { c.onTokenRefresh = fn }
}

// WithOnRefreshError sets a callback for refresh errors.
func WithOnRefreshError(fn func(error)) AuthClientOption {
	return func(c *AuthClient) { c.onRefreshError = fn }
}

// AuthorizationURL returns the URL to redirect the user to for authorization.
func (c *AuthClient) AuthorizationURL(state string, opts ...AuthURLOption) string {
	params := url.Values{
		"client_id":     {c.config.ClientID},
		"redirect_uri":  {c.config.RedirectURL},
		"response_type": {"code"},
		"scope":         {strings.Join(c.config.Scopes, " ")},
		"state":         {state},
		"access_type":   {"offline"},
	}

	for _, opt := range opts {
		opt(params)
	}

	return c.config.AuthURL + "?" + params.Encode()
}

// AuthURLOption configures the authorization URL.
type AuthURLOption func(url.Values)

// WithPrompt sets the prompt parameter (e.g., "consent", "select_account").
func WithPrompt(prompt string) AuthURLOption {
	return func(v url.Values) { v.Set("prompt", prompt) }
}

// WithLoginHint sets a hint for which account to use.
func WithLoginHint(hint string) AuthURLOption {
	return func(v url.Values) { v.Set("login_hint", hint) }
}

// Exchange exchanges an authorization code for a token.
func (c *AuthClient) Exchange(ctx context.Context, code string) (*Token, error) {
	data := url.Values{
		"client_id":     {c.config.ClientID},
		"client_secret": {c.config.ClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {c.config.RedirectURL},
	}

	token, err := c.doTokenRequest(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("exchange: %w", err)
	}

	c.mu.Lock()
	c.token = token
	c.mu.Unlock()

	return token.Clone(), nil
}

// Refresh refreshes the access token using the refresh token.
func (c *AuthClient) Refresh(ctx context.Context) (*Token, error) {
	c.mu.RLock()
	refreshToken := ""
	if c.token != nil {
		refreshToken = c.token.RefreshToken
	}
	c.mu.RUnlock()

	if refreshToken == "" {
		return nil, errors.New("no refresh token available")
	}

	data := url.Values{
		"client_id":     {c.config.ClientID},
		"client_secret": {c.config.ClientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}

	newToken, err := c.doTokenRequest(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("refresh: %w", err)
	}

	// Preserve the original refresh token if not returned
	if newToken.RefreshToken == "" {
		newToken.RefreshToken = refreshToken
	}

	c.mu.Lock()
	c.token = newToken
	c.mu.Unlock()

	if c.onTokenRefresh != nil {
		c.onTokenRefresh(newToken.Clone())
	}

	return newToken.Clone(), nil
}

// Token returns the current token (thread-safe clone).
func (c *AuthClient) Token() *Token {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.token == nil {
		return nil
	}
	return c.token.Clone()
}

// SetToken sets the token.
func (c *AuthClient) SetToken(token *Token) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
}

// AccessToken returns a valid access token, refreshing if necessary.
func (c *AuthClient) AccessToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	token := c.token
	c.mu.RUnlock()

	if token == nil {
		return "", errors.New("no token available")
	}

	if token.Valid() {
		return token.GetAccessToken(), nil
	}

	// Token expired, try to refresh
	newToken, err := c.Refresh(ctx)
	if err != nil {
		return "", err
	}

	return newToken.AccessToken, nil
}

// StartAutoRefresh starts a goroutine that automatically refreshes the token.
func (c *AuthClient) StartAutoRefresh(ctx context.Context) error {
	if !c.state.CompareAndSwap(stateStopped, stateRunning) {
		return errors.New("auto-refresh already running")
	}

	refreshCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	c.wg.Add(1)
	go c.refreshLoop(refreshCtx)

	return nil
}

// StopAutoRefresh stops the auto-refresh goroutine.
func (c *AuthClient) StopAutoRefresh() {
	if !c.state.CompareAndSwap(stateRunning, stateStopping) {
		return // Not running
	}

	if c.cancel != nil {
		c.cancel()
	}

	c.wg.Wait()
	c.state.Store(stateStopped)
}

// refreshLoop periodically refreshes the token.
func (c *AuthClient) refreshLoop(ctx context.Context) {
	defer c.wg.Done()
	defer c.state.Store(stateStopped) // Ensure state is stopped on exit

	for {
		c.mu.RLock()
		token := c.token
		c.mu.RUnlock()

		var sleepDuration time.Duration
		if token == nil || token.Expiry.IsZero() {
			// No token or no expiry, check again later
			sleepDuration = 1 * time.Minute
		} else {
			// Calculate when to refresh
			refreshAt := token.Expiry.Add(-c.refreshEarly)
			sleepDuration = max(time.Until(refreshAt), 0)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(sleepDuration):
		}

		// Check if we should refresh
		c.mu.RLock()
		token = c.token
		c.mu.RUnlock()

		if token == nil {
			continue
		}

		// Refresh if within the early refresh window
		if !token.Expiry.IsZero() && time.Until(token.Expiry) <= c.refreshEarly {
			_, err := c.Refresh(ctx)
			if err != nil && c.onRefreshError != nil {
				c.onRefreshError(err)
			}
		}
	}
}

// doTokenRequest performs a token request to the OAuth server.
func (c *AuthClient) doTokenRequest(ctx context.Context, data url.Values) (*Token, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	// Limit response body to 1MB to prevent memory exhaustion
	const maxResponseSize = 1024 * 1024
	limitedReader := io.LimitReader(resp.Body, maxResponseSize+1)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if len(body) > maxResponseSize {
		return nil, fmt.Errorf("response body exceeds maximum size of %d bytes", maxResponseSize)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
			return nil, &AuthError{
				Code:    errResp.Error,
				Message: errResp.ErrorDescription,
			}
		}
		return nil, fmt.Errorf("token request failed: %s", string(body))
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}

	return tr.toToken(), nil
}

// AuthError represents an OAuth authentication error.
type AuthError struct {
	Code    string // e.g., "invalid_grant", "expired_token"
	Message string
}

// Error implements the error interface.
func (e *AuthError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("auth error: %s - %s", e.Code, e.Message)
	}
	return fmt.Sprintf("auth error: %s", e.Code)
}

// IsExpiredToken returns true if this is an expired token error.
func (e *AuthError) IsExpiredToken() bool {
	return e.Code == "invalid_grant" || e.Code == "expired_token"
}

// IsInvalidGrant returns true if this is an invalid grant error.
func (e *AuthError) IsInvalidGrant() bool {
	return e.Code == "invalid_grant"
}
