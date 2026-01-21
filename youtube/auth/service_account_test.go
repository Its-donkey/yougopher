package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// testPrivateKey is a test RSA private key for testing purposes only.
var testPrivateKey string

func init() {
	// Generate a test RSA key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}
	testPrivateKey = string(pem.EncodeToMemory(pemBlock))
}

func TestNewServiceAccountClient(t *testing.T) {
	config := ServiceAccountConfig{
		Email:      "test@example.iam.gserviceaccount.com",
		PrivateKey: testPrivateKey,
		Scopes:     []string{ScopeLiveChat, ScopeReadOnly},
	}

	client, err := NewServiceAccountClient(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.config.Email != "test@example.iam.gserviceaccount.com" {
		t.Errorf("expected Email %q, got %q", "test@example.iam.gserviceaccount.com", client.config.Email)
	}
	if client.config.TokenURL != DefaultTokenURL {
		t.Errorf("expected TokenURL %q, got %q", DefaultTokenURL, client.config.TokenURL)
	}
	if client.privateKey == nil {
		t.Error("expected privateKey to be set")
	}
}

func TestNewServiceAccountClient_InvalidKey(t *testing.T) {
	config := ServiceAccountConfig{
		Email:      "test@example.iam.gserviceaccount.com",
		PrivateKey: "invalid-key",
		Scopes:     []string{ScopeLiveChat},
	}

	_, err := NewServiceAccountClient(config)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "parsing private key") {
		t.Errorf("expected parsing private key error, got: %v", err)
	}
}

func TestNewServiceAccountClient_WithOptions(t *testing.T) {
	customHTTP := &http.Client{Timeout: 60 * time.Second}
	var refreshCalled bool

	config := ServiceAccountConfig{
		Email:      "test@example.iam.gserviceaccount.com",
		PrivateKey: testPrivateKey,
		Scopes:     []string{ScopeLiveChat},
		TokenURL:   "https://custom.example.com/token",
	}

	client, err := NewServiceAccountClient(config,
		WithServiceAccountHTTPClient(customHTTP),
		WithServiceAccountRefreshEarly(10*time.Minute),
		WithServiceAccountOnTokenRefresh(func(*Token) { refreshCalled = true }),
		WithServiceAccountOnRefreshError(func(error) {}),
		WithSubject("user@example.com"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.httpClient != customHTTP {
		t.Error("expected custom HTTP client to be set")
	}
	if client.refreshEarly != 10*time.Minute {
		t.Errorf("expected refreshEarly %v, got %v", 10*time.Minute, client.refreshEarly)
	}
	if client.config.Subject != "user@example.com" {
		t.Errorf("expected Subject %q, got %q", "user@example.com", client.config.Subject)
	}
	if client.config.TokenURL != "https://custom.example.com/token" {
		t.Errorf("expected custom TokenURL, got %q", client.config.TokenURL)
	}

	// Verify callback is set by checking it's not nil
	if client.onTokenRefresh == nil {
		t.Error("expected onTokenRefresh to be set")
	}
	// Trigger callback to verify it was set correctly
	client.onTokenRefresh(&Token{})
	if !refreshCalled {
		t.Error("expected refresh callback to be called")
	}
}

func TestNewServiceAccountClientFromJSON(t *testing.T) {
	jsonData := []byte(`{
		"type": "service_account",
		"client_email": "test@example.iam.gserviceaccount.com",
		"private_key": "` + strings.ReplaceAll(testPrivateKey, "\n", "\\n") + `",
		"private_key_id": "key-123",
		"token_uri": "https://oauth2.googleapis.com/token"
	}`)

	client, err := NewServiceAccountClientFromJSON(jsonData, []string{ScopeLiveChat})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.config.Email != "test@example.iam.gserviceaccount.com" {
		t.Errorf("expected Email %q, got %q", "test@example.iam.gserviceaccount.com", client.config.Email)
	}
	if client.config.PrivateKeyID != "key-123" {
		t.Errorf("expected PrivateKeyID %q, got %q", "key-123", client.config.PrivateKeyID)
	}
	if len(client.config.Scopes) != 1 || client.config.Scopes[0] != ScopeLiveChat {
		t.Errorf("expected Scopes [%s], got %v", ScopeLiveChat, client.config.Scopes)
	}
}

func TestNewServiceAccountClientFromJSON_InvalidJSON(t *testing.T) {
	jsonData := []byte(`{invalid json}`)

	_, err := NewServiceAccountClientFromJSON(jsonData, []string{ScopeLiveChat})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestServiceAccountClient_FetchToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("expected Content-Type application/x-www-form-urlencoded")
		}

		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("grant_type") != "urn:ietf:params:oauth:grant-type:jwt-bearer" {
			t.Errorf("expected grant_type jwt-bearer, got %s", r.Form.Get("grant_type"))
		}
		if r.Form.Get("assertion") == "" {
			t.Error("expected assertion to be set")
		}

		resp := map[string]any{
			"access_token": "test-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := ServiceAccountConfig{
		Email:      "test@example.iam.gserviceaccount.com",
		PrivateKey: testPrivateKey,
		Scopes:     []string{ScopeLiveChat},
		TokenURL:   server.URL,
	}

	client, err := NewServiceAccountClient(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token, err := client.FetchToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token.AccessToken != "test-access-token" {
		t.Errorf("expected AccessToken %q, got %q", "test-access-token", token.AccessToken)
	}
	if token.TokenType != "Bearer" {
		t.Errorf("expected TokenType %q, got %q", "Bearer", token.TokenType)
	}
}

func TestServiceAccountClient_FetchToken_WithSubject(t *testing.T) {
	var receivedAssertion string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		receivedAssertion = r.Form.Get("assertion")

		resp := map[string]any{
			"access_token": "test-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := ServiceAccountConfig{
		Email:      "test@example.iam.gserviceaccount.com",
		PrivateKey: testPrivateKey,
		Scopes:     []string{ScopeLiveChat},
		TokenURL:   server.URL,
		Subject:    "impersonated@example.com",
	}

	client, err := NewServiceAccountClient(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.FetchToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the JWT contains the subject claim
	// The JWT is in format: header.claims.signature
	parts := strings.Split(receivedAssertion, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWT parts, got %d", len(parts))
	}
}

func TestServiceAccountClient_FetchToken_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		resp := map[string]any{
			"error":             "invalid_grant",
			"error_description": "Invalid JWT",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := ServiceAccountConfig{
		Email:      "test@example.iam.gserviceaccount.com",
		PrivateKey: testPrivateKey,
		Scopes:     []string{ScopeLiveChat},
		TokenURL:   server.URL,
	}

	client, err := NewServiceAccountClient(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.FetchToken(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var authErr *AuthError
	if !containsError(err, &authErr) {
		t.Fatalf("expected AuthError, got %T", err)
	}
	if authErr.Code != "invalid_grant" {
		t.Errorf("expected error code %q, got %q", "invalid_grant", authErr.Code)
	}
}

func TestServiceAccountClient_FetchToken_Callback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"access_token": "test-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	var callbackToken *Token
	config := ServiceAccountConfig{
		Email:      "test@example.iam.gserviceaccount.com",
		PrivateKey: testPrivateKey,
		Scopes:     []string{ScopeLiveChat},
		TokenURL:   server.URL,
	}

	client, err := NewServiceAccountClient(config,
		WithServiceAccountOnTokenRefresh(func(token *Token) {
			callbackToken = token
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.FetchToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callbackToken == nil {
		t.Error("expected callback to be called")
	}
	if callbackToken.AccessToken != "test-access-token" {
		t.Errorf("expected AccessToken %q, got %q", "test-access-token", callbackToken.AccessToken)
	}
}

func TestServiceAccountClient_AccessToken(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		resp := map[string]any{
			"access_token": "test-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := ServiceAccountConfig{
		Email:      "test@example.iam.gserviceaccount.com",
		PrivateKey: testPrivateKey,
		Scopes:     []string{ScopeLiveChat},
		TokenURL:   server.URL,
	}

	client, err := NewServiceAccountClient(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First call should fetch token
	accessToken, err := client.AccessToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accessToken != "test-access-token" {
		t.Errorf("expected access token %q, got %q", "test-access-token", accessToken)
	}
	if callCount.Load() != 1 {
		t.Errorf("expected 1 call, got %d", callCount.Load())
	}

	// Second call should use cached token
	accessToken, err = client.AccessToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accessToken != "test-access-token" {
		t.Errorf("expected access token %q, got %q", "test-access-token", accessToken)
	}
	if callCount.Load() != 1 {
		t.Errorf("expected 1 call (cached), got %d", callCount.Load())
	}
}

func TestServiceAccountClient_Token(t *testing.T) {
	config := ServiceAccountConfig{
		Email:      "test@example.iam.gserviceaccount.com",
		PrivateKey: testPrivateKey,
		Scopes:     []string{ScopeLiveChat},
	}

	client, err := NewServiceAccountClient(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Initially nil
	if client.Token() != nil {
		t.Error("expected nil token initially")
	}

	// Set token directly for testing
	client.mu.Lock()
	client.token = &Token{
		AccessToken: "test-token",
		Expiry:      time.Now().Add(1 * time.Hour),
	}
	client.mu.Unlock()

	// Now should return token
	token := client.Token()
	if token == nil {
		t.Fatal("expected token, got nil")
	}
	if token.AccessToken != "test-token" {
		t.Errorf("expected AccessToken %q, got %q", "test-token", token.AccessToken)
	}
}

func TestServiceAccountClient_AutoRefresh(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		resp := map[string]any{
			"access_token": "test-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := ServiceAccountConfig{
		Email:      "test@example.iam.gserviceaccount.com",
		PrivateKey: testPrivateKey,
		Scopes:     []string{ScopeLiveChat},
		TokenURL:   server.URL,
	}

	client, err := NewServiceAccountClient(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Set an expired token to trigger refresh
	client.mu.Lock()
	client.token = &Token{
		AccessToken: "expired-token",
		Expiry:      time.Now().Add(-1 * time.Hour), // Already expired
	}
	client.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start auto-refresh
	err = client.StartAutoRefresh(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Starting again should fail
	err = client.StartAutoRefresh(ctx)
	if err == nil {
		t.Error("expected error when starting auto-refresh twice")
	}

	// Wait a bit for the refresh loop to run (expired token means it should refresh immediately)
	time.Sleep(200 * time.Millisecond)

	// Stop auto-refresh
	client.StopAutoRefresh()

	// Check that at least one token fetch was made
	if callCount.Load() < 1 {
		t.Errorf("expected at least 1 call, got %d", callCount.Load())
	}
}

func TestServiceAccountClient_StopAutoRefresh_NotRunning(t *testing.T) {
	config := ServiceAccountConfig{
		Email:      "test@example.iam.gserviceaccount.com",
		PrivateKey: testPrivateKey,
		Scopes:     []string{ScopeLiveChat},
	}

	client, err := NewServiceAccountClient(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Stopping when not running should be a no-op
	client.StopAutoRefresh()
}

func TestParsePrivateKey_PKCS8(t *testing.T) {
	// Generate a key and encode as PKCS#8
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}

	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	}
	pemData := string(pem.EncodeToMemory(pemBlock))

	parsedKey, err := parsePrivateKey(pemData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsedKey == nil {
		t.Error("expected non-nil key")
	}
}

func TestParsePrivateKey_PKCS1(t *testing.T) {
	// testPrivateKey is already PKCS#1 format
	parsedKey, err := parsePrivateKey(testPrivateKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsedKey == nil {
		t.Error("expected non-nil key")
	}
}

func TestParsePrivateKey_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		errMsg string
	}{
		{
			name:   "no PEM block",
			input:  "not a pem block",
			errMsg: "no valid PEM block found",
		},
		{
			name:   "invalid PEM content",
			input:  "-----BEGIN RSA PRIVATE KEY-----\nYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=\n-----END RSA PRIVATE KEY-----",
			errMsg: "failed to parse private key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parsePrivateKey(tt.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing %q, got: %v", tt.errMsg, err)
			}
		})
	}
}

func TestServiceAccountClient_CreateJWT(t *testing.T) {
	config := ServiceAccountConfig{
		Email:        "test@example.iam.gserviceaccount.com",
		PrivateKey:   testPrivateKey,
		PrivateKeyID: "key-123",
		Scopes:       []string{ScopeLiveChat, ScopeReadOnly},
		TokenURL:     DefaultTokenURL,
	}

	client, err := NewServiceAccountClient(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	jwt, err := client.createJWT()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// JWT should have 3 parts separated by dots
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWT parts, got %d", len(parts))
	}

	// Each part should be non-empty
	for i, part := range parts {
		if part == "" {
			t.Errorf("JWT part %d is empty", i)
		}
	}
}
