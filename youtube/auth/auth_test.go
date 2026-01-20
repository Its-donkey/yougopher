package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewAuthClient_Defaults(t *testing.T) {
	config := Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost/callback",
		Scopes:       []string{ScopeLiveChat},
	}

	client := NewAuthClient(config)

	if client.config.AuthURL != DefaultAuthURL {
		t.Errorf("AuthURL = %q, want %q", client.config.AuthURL, DefaultAuthURL)
	}
	if client.config.TokenURL != DefaultTokenURL {
		t.Errorf("TokenURL = %q, want %q", client.config.TokenURL, DefaultTokenURL)
	}
	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestNewAuthClient_WithOptions(t *testing.T) {
	config := Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost/callback",
	}

	customHTTP := &http.Client{Timeout: 60 * time.Second}
	token := &Token{AccessToken: "initial-token"}
	var refreshCalled bool
	var errorCalled bool

	client := NewAuthClient(config,
		WithHTTPClient(customHTTP),
		WithToken(token),
		WithRefreshEarly(10*time.Minute),
		WithOnTokenRefresh(func(*Token) { refreshCalled = true }),
		WithOnRefreshError(func(error) { errorCalled = true }),
	)

	if client.httpClient != customHTTP {
		t.Error("httpClient not set correctly")
	}
	if client.token != token {
		t.Error("token not set correctly")
	}
	if client.refreshEarly != 10*time.Minute {
		t.Errorf("refreshEarly = %v, want 10m", client.refreshEarly)
	}

	// Test callbacks are set (just verify not nil)
	if client.onTokenRefresh == nil {
		t.Error("onTokenRefresh should be set")
	}
	if client.onRefreshError == nil {
		t.Error("onRefreshError should be set")
	}

	// Silence unused warnings
	_ = refreshCalled
	_ = errorCalled
}

func TestAuthClient_AuthorizationURL(t *testing.T) {
	config := Config{
		ClientID:    "client-id",
		RedirectURL: "http://localhost/callback",
		Scopes:      []string{ScopeLiveChat, ScopeReadOnly},
		AuthURL:     "https://auth.example.com/authorize",
	}

	client := NewAuthClient(config)

	url := client.AuthorizationURL("state123")

	// Check URL contains required parameters
	if !strings.Contains(url, "client_id=client-id") {
		t.Error("URL missing client_id")
	}
	if !strings.Contains(url, "redirect_uri=http") {
		t.Error("URL missing redirect_uri")
	}
	if !strings.Contains(url, "response_type=code") {
		t.Error("URL missing response_type")
	}
	if !strings.Contains(url, "state=state123") {
		t.Error("URL missing state")
	}
	if !strings.Contains(url, "access_type=offline") {
		t.Error("URL missing access_type")
	}
	if !strings.HasPrefix(url, "https://auth.example.com/authorize?") {
		t.Errorf("URL has wrong prefix: %s", url)
	}
}

func TestAuthClient_AuthorizationURL_WithOptions(t *testing.T) {
	config := Config{
		ClientID:    "client-id",
		RedirectURL: "http://localhost/callback",
		Scopes:      []string{ScopeLiveChat},
	}

	client := NewAuthClient(config)

	url := client.AuthorizationURL("state",
		WithPrompt("consent"),
		WithLoginHint("user@example.com"),
	)

	if !strings.Contains(url, "prompt=consent") {
		t.Error("URL missing prompt parameter")
	}
	if !strings.Contains(url, "login_hint=user%40example.com") {
		t.Error("URL missing login_hint parameter")
	}
}

func TestAuthClient_Exchange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %q, want application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm error: %v", err)
		}

		if r.Form.Get("grant_type") != "authorization_code" {
			t.Errorf("grant_type = %q, want authorization_code", r.Form.Get("grant_type"))
		}
		if r.Form.Get("code") != "auth-code" {
			t.Errorf("code = %q, want auth-code", r.Form.Get("code"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "access-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": "refresh-token",
		})
	}))
	defer server.Close()

	config := Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost/callback",
		TokenURL:     server.URL,
	}

	client := NewAuthClient(config)
	token, err := client.Exchange(context.Background(), "auth-code")

	if err != nil {
		t.Fatalf("Exchange() error = %v", err)
	}

	if token.AccessToken != "access-token" {
		t.Errorf("AccessToken = %q, want 'access-token'", token.AccessToken)
	}
	if token.RefreshToken != "refresh-token" {
		t.Errorf("RefreshToken = %q, want 'refresh-token'", token.RefreshToken)
	}

	// Verify internal token is set
	internalToken := client.Token()
	if internalToken.AccessToken != "access-token" {
		t.Error("Internal token not set correctly")
	}
}

func TestAuthClient_Exchange_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":             "invalid_grant",
			"error_description": "Code expired",
		})
	}))
	defer server.Close()

	config := Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost/callback",
		TokenURL:     server.URL,
	}

	client := NewAuthClient(config)
	_, err := client.Exchange(context.Background(), "bad-code")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("Expected *AuthError, got %T", err)
	}
	if authErr.Code != "invalid_grant" {
		t.Errorf("Code = %q, want 'invalid_grant'", authErr.Code)
	}
}

func TestAuthClient_Refresh(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm error: %v", err)
		}

		if r.Form.Get("grant_type") != "refresh_token" {
			t.Errorf("grant_type = %q, want refresh_token", r.Form.Get("grant_type"))
		}
		if r.Form.Get("refresh_token") != "old-refresh-token" {
			t.Errorf("refresh_token = %q, want old-refresh-token", r.Form.Get("refresh_token"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "new-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
			// Note: refresh_token not returned - should preserve old one
		})
	}))
	defer server.Close()

	config := Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		TokenURL:     server.URL,
	}

	var refreshedToken *Token
	client := NewAuthClient(config,
		WithToken(&Token{
			AccessToken:  "old-access-token",
			RefreshToken: "old-refresh-token",
		}),
		WithOnTokenRefresh(func(t *Token) {
			refreshedToken = t
		}),
	)

	token, err := client.Refresh(context.Background())
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	if token.AccessToken != "new-access-token" {
		t.Errorf("AccessToken = %q, want 'new-access-token'", token.AccessToken)
	}
	// Refresh token should be preserved
	if token.RefreshToken != "old-refresh-token" {
		t.Errorf("RefreshToken = %q, want 'old-refresh-token' (preserved)", token.RefreshToken)
	}

	// Callback should have been called
	if refreshedToken == nil {
		t.Error("onTokenRefresh callback not called")
	}
}

func TestAuthClient_Refresh_NoRefreshToken(t *testing.T) {
	config := Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}

	client := NewAuthClient(config)
	_, err := client.Refresh(context.Background())

	if err == nil {
		t.Fatal("Expected error for missing refresh token")
	}
	if !strings.Contains(err.Error(), "no refresh token") {
		t.Errorf("Error = %q, want to contain 'no refresh token'", err.Error())
	}
}

func TestAuthClient_Token(t *testing.T) {
	config := Config{ClientID: "id"}

	t.Run("nil token", func(t *testing.T) {
		client := NewAuthClient(config)
		got := client.Token()
		if got != nil {
			t.Errorf("Token() = %v, want nil", got)
		}
	})

	t.Run("returns clone", func(t *testing.T) {
		token := &Token{AccessToken: "test"}
		client := NewAuthClient(config, WithToken(token))

		got := client.Token()
		got.AccessToken = "modified"

		// Original should not be modified
		internal := client.Token()
		if internal.AccessToken != "test" {
			t.Error("Token() did not return a clone")
		}
	})
}

func TestAuthClient_SetToken(t *testing.T) {
	config := Config{ClientID: "id"}
	client := NewAuthClient(config)

	token := &Token{AccessToken: "new-token"}
	client.SetToken(token)

	got := client.Token()
	if got.AccessToken != "new-token" {
		t.Errorf("Token().AccessToken = %q, want 'new-token'", got.AccessToken)
	}
}

func TestAuthClient_AccessToken(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		config := Config{ClientID: "id"}
		token := &Token{
			AccessToken: "valid-token",
			Expiry:      time.Now().Add(1 * time.Hour),
		}
		client := NewAuthClient(config, WithToken(token))

		got, err := client.AccessToken(context.Background())
		if err != nil {
			t.Fatalf("AccessToken() error = %v", err)
		}
		if got != "valid-token" {
			t.Errorf("AccessToken() = %q, want 'valid-token'", got)
		}
	})

	t.Run("no token", func(t *testing.T) {
		config := Config{ClientID: "id"}
		client := NewAuthClient(config)

		_, err := client.AccessToken(context.Background())
		if err == nil {
			t.Fatal("Expected error for no token")
		}
	})

	t.Run("expired token refreshes", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "refreshed-token",
				"expires_in":   3600,
			})
		}))
		defer server.Close()

		config := Config{
			ClientID:     "id",
			ClientSecret: "secret",
			TokenURL:     server.URL,
		}
		token := &Token{
			AccessToken:  "expired-token",
			RefreshToken: "refresh-token",
			Expiry:       time.Now().Add(-1 * time.Hour), // Expired
		}
		client := NewAuthClient(config, WithToken(token))

		got, err := client.AccessToken(context.Background())
		if err != nil {
			t.Fatalf("AccessToken() error = %v", err)
		}
		if got != "refreshed-token" {
			t.Errorf("AccessToken() = %q, want 'refreshed-token'", got)
		}
	})
}

func TestAuthClient_StartStopAutoRefresh(t *testing.T) {
	config := Config{ClientID: "id"}
	client := NewAuthClient(config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start auto-refresh
	err := client.StartAutoRefresh(ctx)
	if err != nil {
		t.Fatalf("StartAutoRefresh() error = %v", err)
	}

	// Starting again should fail
	err = client.StartAutoRefresh(ctx)
	if err == nil {
		t.Error("StartAutoRefresh() should fail when already running")
	}

	// Stop
	client.StopAutoRefresh()

	// Stopping again should be safe (idempotent)
	client.StopAutoRefresh()

	// Should be able to start again after stop
	err = client.StartAutoRefresh(ctx)
	if err != nil {
		t.Fatalf("StartAutoRefresh() after stop error = %v", err)
	}

	client.StopAutoRefresh()
}

func TestAuthError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *AuthError
		wantStr string
	}{
		{
			name:    "with message",
			err:     &AuthError{Code: "invalid_grant", Message: "Token expired"},
			wantStr: "auth error: invalid_grant - Token expired",
		},
		{
			name:    "without message",
			err:     &AuthError{Code: "invalid_grant"},
			wantStr: "auth error: invalid_grant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantStr {
				t.Errorf("Error() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}

func TestAuthError_IsExpiredToken(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{"invalid_grant", true},
		{"expired_token", true},
		{"access_denied", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			err := &AuthError{Code: tt.code}
			got := err.IsExpiredToken()
			if got != tt.want {
				t.Errorf("IsExpiredToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthError_IsInvalidGrant(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{"invalid_grant", true},
		{"expired_token", false},
		{"access_denied", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			err := &AuthError{Code: tt.code}
			got := err.IsInvalidGrant()
			if got != tt.want {
				t.Errorf("IsInvalidGrant() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScopes(t *testing.T) {
	// Just verify the scope constants are defined correctly
	scopes := []string{
		ScopeLiveChat,
		ScopeLiveChatModerate,
		ScopeReadOnly,
		ScopeUpload,
		ScopePartner,
		ScopePartnerChannelAudit,
	}

	for _, s := range scopes {
		if !strings.HasPrefix(s, "https://www.googleapis.com/auth/") {
			t.Errorf("Scope %q should start with Google auth prefix", s)
		}
	}
}
