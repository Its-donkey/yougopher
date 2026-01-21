package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
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

// ServiceAccountConfig holds configuration for service account authentication.
type ServiceAccountConfig struct {
	// Email is the service account email address.
	Email string `json:"client_email"`

	// PrivateKey is the PEM-encoded RSA private key.
	PrivateKey string `json:"private_key"`

	// PrivateKeyID is the private key ID (optional).
	PrivateKeyID string `json:"private_key_id,omitempty"`

	// Scopes are the requested OAuth scopes.
	Scopes []string `json:"-"`

	// TokenURL is the token endpoint (defaults to Google's).
	TokenURL string `json:"token_uri,omitempty"`

	// Subject is the email of the user to impersonate (domain-wide delegation).
	Subject string `json:"-"`
}

// ServiceAccountClient handles authentication using service account credentials.
type ServiceAccountClient struct {
	config     ServiceAccountConfig
	httpClient *http.Client
	privateKey *rsa.PrivateKey

	// Token management
	mu    sync.RWMutex
	token *Token

	// Auto-refresh
	state        atomic.Int32
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	refreshEarly time.Duration

	// Callbacks
	onTokenRefresh func(*Token)
	onRefreshError func(error)

	// For testing
	now func() time.Time
}

// ServiceAccountOption configures a ServiceAccountClient.
type ServiceAccountOption func(*ServiceAccountClient)

// NewServiceAccountClient creates a new service account client.
func NewServiceAccountClient(config ServiceAccountConfig, opts ...ServiceAccountOption) (*ServiceAccountClient, error) {
	if config.TokenURL == "" {
		config.TokenURL = DefaultTokenURL
	}

	// Parse the private key
	privateKey, err := parsePrivateKey(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}

	c := &ServiceAccountClient{
		config:       config,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		privateKey:   privateKey,
		refreshEarly: 5 * time.Minute,
		now:          time.Now,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// NewServiceAccountClientFromJSON creates a client from Google's JSON credentials file.
func NewServiceAccountClientFromJSON(jsonData []byte, scopes []string, opts ...ServiceAccountOption) (*ServiceAccountClient, error) {
	var config ServiceAccountConfig
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, fmt.Errorf("parsing credentials JSON: %w", err)
	}
	config.Scopes = scopes

	return NewServiceAccountClient(config, opts...)
}

// WithServiceAccountHTTPClient sets a custom HTTP client.
func WithServiceAccountHTTPClient(hc *http.Client) ServiceAccountOption {
	return func(c *ServiceAccountClient) { c.httpClient = hc }
}

// WithServiceAccountRefreshEarly sets how early to refresh tokens before expiry.
func WithServiceAccountRefreshEarly(d time.Duration) ServiceAccountOption {
	return func(c *ServiceAccountClient) { c.refreshEarly = d }
}

// WithServiceAccountOnTokenRefresh sets a callback for successful token refreshes.
func WithServiceAccountOnTokenRefresh(fn func(*Token)) ServiceAccountOption {
	return func(c *ServiceAccountClient) { c.onTokenRefresh = fn }
}

// WithServiceAccountOnRefreshError sets a callback for refresh errors.
func WithServiceAccountOnRefreshError(fn func(error)) ServiceAccountOption {
	return func(c *ServiceAccountClient) { c.onRefreshError = fn }
}

// WithSubject sets the email of the user to impersonate (domain-wide delegation).
func WithSubject(email string) ServiceAccountOption {
	return func(c *ServiceAccountClient) { c.config.Subject = email }
}

// Token returns the current token (thread-safe clone).
func (c *ServiceAccountClient) Token() *Token {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.token == nil {
		return nil
	}
	return c.token.Clone()
}

// AccessToken returns a valid access token, fetching a new one if necessary.
func (c *ServiceAccountClient) AccessToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	token := c.token
	c.mu.RUnlock()

	if token != nil && token.Valid() {
		return token.GetAccessToken(), nil
	}

	// Token is nil or expired, fetch a new one
	newToken, err := c.FetchToken(ctx)
	if err != nil {
		return "", err
	}

	return newToken.AccessToken, nil
}

// FetchToken fetches a new access token using the service account credentials.
func (c *ServiceAccountClient) FetchToken(ctx context.Context) (*Token, error) {
	jwt, err := c.createJWT()
	if err != nil {
		return nil, fmt.Errorf("creating JWT: %w", err)
	}

	data := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {jwt},
	}

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

	body, err := c.readResponseBody(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(body)
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}

	token := tr.toToken()
	token.Scopes = c.config.Scopes // Service account tokens use the requested scopes

	c.mu.Lock()
	c.token = token
	c.mu.Unlock()

	if c.onTokenRefresh != nil {
		c.onTokenRefresh(token.Clone())
	}

	return token.Clone(), nil
}

// StartAutoRefresh starts a goroutine that automatically refreshes the token.
func (c *ServiceAccountClient) StartAutoRefresh(ctx context.Context) error {
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
func (c *ServiceAccountClient) StopAutoRefresh() {
	if !c.state.CompareAndSwap(stateRunning, stateStopping) {
		return
	}

	if c.cancel != nil {
		c.cancel()
	}

	c.wg.Wait()
	c.state.Store(stateStopped)
}

// refreshLoop periodically refreshes the token.
func (c *ServiceAccountClient) refreshLoop(ctx context.Context) {
	defer c.wg.Done()
	defer c.state.Store(stateStopped)

	for {
		c.mu.RLock()
		token := c.token
		c.mu.RUnlock()

		var sleepDuration time.Duration
		if token == nil || token.Expiry.IsZero() {
			sleepDuration = 1 * time.Minute
		} else {
			refreshAt := token.Expiry.Add(-c.refreshEarly)
			sleepDuration = max(time.Until(refreshAt), 0)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(sleepDuration):
		}

		c.mu.RLock()
		token = c.token
		c.mu.RUnlock()

		if token == nil {
			_, err := c.FetchToken(ctx)
			if err != nil && c.onRefreshError != nil {
				c.onRefreshError(err)
			}
			continue
		}

		if !token.Expiry.IsZero() && time.Until(token.Expiry) <= c.refreshEarly {
			_, err := c.FetchToken(ctx)
			if err != nil && c.onRefreshError != nil {
				c.onRefreshError(err)
			}
		}
	}
}

// jwtHeader represents the JWT header.
type jwtHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid,omitempty"`
}

// jwtClaimSet represents the JWT claim set.
type jwtClaimSet struct {
	Issuer    string `json:"iss"`
	Subject   string `json:"sub,omitempty"`
	Audience  string `json:"aud"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	Scope     string `json:"scope,omitempty"`
}

// createJWT creates a signed JWT assertion for the service account.
func (c *ServiceAccountClient) createJWT() (string, error) {
	now := c.now()

	header := jwtHeader{
		Algorithm: "RS256",
		Type:      "JWT",
		KeyID:     c.config.PrivateKeyID,
	}

	claims := jwtClaimSet{
		Issuer:    c.config.Email,
		Audience:  c.config.TokenURL,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(1 * time.Hour).Unix(),
		Scope:     strings.Join(c.config.Scopes, " "),
	}

	if c.config.Subject != "" {
		claims.Subject = c.config.Subject
	}

	// Encode header
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Encode claims
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signing input
	signingInput := headerEncoded + "." + claimsEncoded

	// Sign with RS256
	hash := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}
	signatureEncoded := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + signatureEncoded, nil
}

// readResponseBody reads and limits response body size.
func (c *ServiceAccountClient) readResponseBody(body io.Reader) ([]byte, error) {
	const maxResponseSize = 1024 * 1024 // 1MB
	limitedReader := io.LimitReader(body, maxResponseSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if len(data) > maxResponseSize {
		return nil, fmt.Errorf("response body exceeds maximum size of %d bytes", maxResponseSize)
	}
	return data, nil
}

// parseError parses an error response.
func (c *ServiceAccountClient) parseError(body []byte) error {
	var errResp struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return &AuthError{
			Code:    errResp.Error,
			Message: errResp.ErrorDescription,
		}
	}
	return fmt.Errorf("service account token request failed: %s", string(body))
}

// parsePrivateKey parses a PEM-encoded RSA private key.
func parsePrivateKey(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("no valid PEM block found")
	}

	// Try PKCS#8 first (newer format)
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not RSA")
		}
		return rsaKey, nil
	}

	// Fall back to PKCS#1 (older format)
	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return rsaKey, nil
}
