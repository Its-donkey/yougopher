package auth

import (
	"encoding/json"
	"sync"
	"time"
)

// Token represents an OAuth 2.0 token.
type Token struct {
	mu sync.RWMutex

	// AccessToken is the token used to authenticate API requests.
	AccessToken string `json:"access_token"`

	// TokenType is the type of token (usually "Bearer").
	TokenType string `json:"token_type,omitempty"`

	// RefreshToken is used to obtain a new access token.
	RefreshToken string `json:"refresh_token,omitempty"`

	// Expiry is the time when the access token expires.
	Expiry time.Time `json:"expiry,omitempty"`

	// Scopes are the granted OAuth scopes.
	Scopes []string `json:"scopes,omitempty"`
}

// expiryDelta is how early a token is considered expired to account for clock skew.
const expiryDelta = 10 * time.Second

// Valid reports whether t is non-nil, has an AccessToken, and is not expired.
func (t *Token) Valid() bool {
	if t == nil {
		return false
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.valid()
}

// valid reports whether the token is valid (must hold lock).
func (t *Token) valid() bool {
	return t.AccessToken != "" && !t.expired()
}

// expired reports whether the token is expired (must hold lock).
func (t *Token) expired() bool {
	if t.Expiry.IsZero() {
		return false // No expiry set means it doesn't expire
	}
	return t.Expiry.Add(-expiryDelta).Before(time.Now())
}

// Expired reports whether the token has expired.
func (t *Token) Expired() bool {
	if t == nil {
		return true
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.expired()
}

// SetAccessToken updates the access token and expiry.
func (t *Token) SetAccessToken(accessToken string, expiry time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.AccessToken = accessToken
	t.Expiry = expiry
}

// GetAccessToken returns the current access token.
func (t *Token) GetAccessToken() string {
	if t == nil {
		return ""
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.AccessToken
}

// GetRefreshToken returns the refresh token.
func (t *Token) GetRefreshToken() string {
	if t == nil {
		return ""
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.RefreshToken
}

// ExpiresIn returns the duration until the token expires.
// Returns 0 if the token has no expiry or is already expired.
func (t *Token) ExpiresIn() time.Duration {
	if t == nil {
		return 0
	}
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.Expiry.IsZero() {
		return 0
	}

	d := time.Until(t.Expiry)
	if d < 0 {
		return 0
	}
	return d
}

// Clone returns a deep copy of the token.
func (t *Token) Clone() *Token {
	if t == nil {
		return nil
	}
	t.mu.RLock()
	defer t.mu.RUnlock()

	scopes := make([]string, len(t.Scopes))
	copy(scopes, t.Scopes)

	return &Token{
		AccessToken:  t.AccessToken,
		TokenType:    t.TokenType,
		RefreshToken: t.RefreshToken,
		Expiry:       t.Expiry,
		Scopes:       scopes,
	}
}

// tokenJSON is used for JSON marshaling/unmarshaling.
type tokenJSON struct {
	AccessToken  string   `json:"access_token"`
	TokenType    string   `json:"token_type,omitempty"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	ExpiresIn    int64    `json:"expires_in,omitempty"`
	Expiry       string   `json:"expiry,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (t *Token) MarshalJSON() ([]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	tj := tokenJSON{
		AccessToken:  t.AccessToken,
		TokenType:    t.TokenType,
		RefreshToken: t.RefreshToken,
		Scopes:       t.Scopes,
	}

	if !t.Expiry.IsZero() {
		tj.Expiry = t.Expiry.Format(time.RFC3339)
	}

	return json.Marshal(tj)
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Token) UnmarshalJSON(data []byte) error {
	var tj tokenJSON
	if err := json.Unmarshal(data, &tj); err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.AccessToken = tj.AccessToken
	t.TokenType = tj.TokenType
	t.RefreshToken = tj.RefreshToken
	t.Scopes = tj.Scopes

	// Parse expiry from either expires_in or expiry field
	if tj.Expiry != "" {
		expiry, err := time.Parse(time.RFC3339, tj.Expiry)
		if err == nil {
			t.Expiry = expiry
		}
	} else if tj.ExpiresIn > 0 {
		t.Expiry = time.Now().Add(time.Duration(tj.ExpiresIn) * time.Second)
	}

	return nil
}

// tokenResponse represents the JSON response from Google's token endpoint.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// toToken converts a tokenResponse to a Token.
func (tr *tokenResponse) toToken() *Token {
	var expiry time.Time
	if tr.ExpiresIn > 0 {
		expiry = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	}

	return &Token{
		AccessToken:  tr.AccessToken,
		TokenType:    tr.TokenType,
		RefreshToken: tr.RefreshToken,
		Expiry:       expiry,
	}
}
