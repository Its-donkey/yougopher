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
	"time"
)

// Google Device Code endpoint.
const DefaultDeviceCodeURL = "https://oauth2.googleapis.com/device/code"

// Device flow error codes.
const (
	errAuthorizationPending = "authorization_pending"
	errSlowDown             = "slow_down"
	errAccessDenied         = "access_denied"
	errExpiredToken         = "expired_token"
)

// DeviceConfig holds configuration for Device Code Flow.
type DeviceConfig struct {
	// ClientID is the OAuth client ID.
	ClientID string

	// Scopes are the requested OAuth scopes.
	Scopes []string

	// DeviceCodeURL is the device authorization endpoint (defaults to Google's).
	DeviceCodeURL string

	// TokenURL is the token endpoint (defaults to Google's).
	TokenURL string
}

// DeviceAuthResponse contains the response from the device authorization request.
type DeviceAuthResponse struct {
	// DeviceCode is the device verification code.
	DeviceCode string `json:"device_code"`

	// UserCode is the code the user enters at the verification URL.
	UserCode string `json:"user_code"`

	// VerificationURL is where the user should go to enter the code.
	VerificationURL string `json:"verification_uri"`

	// VerificationURLComplete is a URL that includes the user code (optional).
	VerificationURLComplete string `json:"verification_uri_complete,omitempty"`

	// ExpiresIn is how long the codes are valid for in seconds.
	ExpiresIn int `json:"expires_in"`

	// Interval is the minimum polling interval in seconds.
	Interval int `json:"interval"`

	// Internal tracking
	expiry time.Time
}

// Expired returns true if the device authorization has expired.
func (d *DeviceAuthResponse) Expired() bool {
	return time.Now().After(d.expiry)
}

// DeviceClient handles OAuth 2.0 Device Code Flow for limited-input devices.
type DeviceClient struct {
	config     DeviceConfig
	httpClient *http.Client

	// For testing
	now func() time.Time
}

// DeviceClientOption configures a DeviceClient.
type DeviceClientOption func(*DeviceClient)

// NewDeviceClient creates a new Device Code Flow client.
func NewDeviceClient(config DeviceConfig, opts ...DeviceClientOption) *DeviceClient {
	if config.DeviceCodeURL == "" {
		config.DeviceCodeURL = DefaultDeviceCodeURL
	}
	if config.TokenURL == "" {
		config.TokenURL = DefaultTokenURL
	}

	c := &DeviceClient{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		now:        time.Now,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithDeviceHTTPClient sets a custom HTTP client.
func WithDeviceHTTPClient(hc *http.Client) DeviceClientOption {
	return func(c *DeviceClient) { c.httpClient = hc }
}

// RequestDeviceCode initiates the device authorization flow.
// Returns the device authorization response containing the user code and verification URL.
func (c *DeviceClient) RequestDeviceCode(ctx context.Context) (*DeviceAuthResponse, error) {
	data := url.Values{
		"client_id": {c.config.ClientID},
		"scope":     {strings.Join(c.config.Scopes, " ")},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.DeviceCodeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating device code request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("device code request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := c.readResponseBody(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(body, "device code request failed")
	}

	var authResp DeviceAuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, fmt.Errorf("parsing device code response: %w", err)
	}

	// Set expiry time
	authResp.expiry = c.now().Add(time.Duration(authResp.ExpiresIn) * time.Second)

	// Default interval if not provided
	if authResp.Interval == 0 {
		authResp.Interval = 5
	}

	return &authResp, nil
}

// PollForToken polls the token endpoint until the user authorizes or the code expires.
// This is a blocking call that handles the polling loop internally.
func (c *DeviceClient) PollForToken(ctx context.Context, authResp *DeviceAuthResponse) (*Token, error) {
	if authResp == nil {
		return nil, errors.New("device auth response is nil")
	}

	interval := time.Duration(authResp.Interval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if authResp.Expired() {
				return nil, &DeviceAuthError{Code: errExpiredToken, Message: "device code has expired"}
			}

			token, err := c.exchangeDeviceCode(ctx, authResp.DeviceCode)
			if err == nil {
				return token, nil
			}

			var deviceErr *DeviceAuthError
			if errors.As(err, &deviceErr) {
				switch deviceErr.Code {
				case errAuthorizationPending:
					// User hasn't authorized yet, continue polling
					continue
				case errSlowDown:
					// Increase polling interval
					interval += 5 * time.Second
					ticker.Reset(interval)
					continue
				case errAccessDenied:
					return nil, deviceErr
				case errExpiredToken:
					return nil, deviceErr
				}
			}

			// Other error, return it
			return nil, err
		}
	}
}

// PollForTokenAsync starts polling in a goroutine and returns channels for the result.
// The token channel receives the token when authorized.
// The error channel receives any error that occurs.
// The cancel function stops the polling.
func (c *DeviceClient) PollForTokenAsync(ctx context.Context, authResp *DeviceAuthResponse) (<-chan *Token, <-chan error, context.CancelFunc) {
	tokenCh := make(chan *Token, 1)
	errCh := make(chan error, 1)

	pollCtx, cancel := context.WithCancel(ctx)

	go func() {
		defer close(tokenCh)
		defer close(errCh)

		token, err := c.PollForToken(pollCtx, authResp)
		if err != nil {
			errCh <- err
			return
		}
		tokenCh <- token
	}()

	return tokenCh, errCh, cancel
}

// exchangeDeviceCode exchanges the device code for an access token.
func (c *DeviceClient) exchangeDeviceCode(ctx context.Context, deviceCode string) (*Token, error) {
	data := url.Values{
		"client_id":   {c.config.ClientID},
		"device_code": {deviceCode},
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
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
		return nil, c.parseDeviceError(body)
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}

	return tr.toToken(), nil
}

// readResponseBody reads and limits response body size.
func (c *DeviceClient) readResponseBody(body io.Reader) ([]byte, error) {
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

// parseError parses a generic error response.
func (c *DeviceClient) parseError(body []byte, defaultMsg string) error {
	var errResp struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return &DeviceAuthError{
			Code:    errResp.Error,
			Message: errResp.ErrorDescription,
		}
	}
	return fmt.Errorf("%s: %s", defaultMsg, string(body))
}

// parseDeviceError parses device authorization error responses.
func (c *DeviceClient) parseDeviceError(body []byte) error {
	var errResp struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return &DeviceAuthError{
			Code:    errResp.Error,
			Message: errResp.ErrorDescription,
		}
	}
	return fmt.Errorf("device token exchange failed: %s", string(body))
}

// DeviceAuthError represents a device authorization error.
type DeviceAuthError struct {
	Code    string // e.g., "authorization_pending", "slow_down", "access_denied", "expired_token"
	Message string
}

// Error implements the error interface.
func (e *DeviceAuthError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("device auth error: %s - %s", e.Code, e.Message)
	}
	return fmt.Sprintf("device auth error: %s", e.Code)
}

// IsAuthorizationPending returns true if the user hasn't authorized yet.
func (e *DeviceAuthError) IsAuthorizationPending() bool {
	return e.Code == errAuthorizationPending
}

// IsSlowDown returns true if polling should slow down.
func (e *DeviceAuthError) IsSlowDown() bool {
	return e.Code == errSlowDown
}

// IsAccessDenied returns true if the user denied access.
func (e *DeviceAuthError) IsAccessDenied() bool {
	return e.Code == errAccessDenied
}

// IsExpired returns true if the device code has expired.
func (e *DeviceAuthError) IsExpired() bool {
	return e.Code == errExpiredToken
}
