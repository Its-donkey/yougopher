package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewDeviceClient(t *testing.T) {
	config := DeviceConfig{
		ClientID: "test-client-id",
		Scopes:   []string{ScopeLiveChat, ScopeReadOnly},
	}

	client := NewDeviceClient(config)

	if client.config.ClientID != "test-client-id" {
		t.Errorf("expected ClientID %q, got %q", "test-client-id", client.config.ClientID)
	}
	if client.config.DeviceCodeURL != DefaultDeviceCodeURL {
		t.Errorf("expected DeviceCodeURL %q, got %q", DefaultDeviceCodeURL, client.config.DeviceCodeURL)
	}
	if client.config.TokenURL != DefaultTokenURL {
		t.Errorf("expected TokenURL %q, got %q", DefaultTokenURL, client.config.TokenURL)
	}
}

func TestNewDeviceClient_WithOptions(t *testing.T) {
	customHTTP := &http.Client{Timeout: 60 * time.Second}

	config := DeviceConfig{
		ClientID:      "test-client-id",
		Scopes:        []string{ScopeLiveChat},
		DeviceCodeURL: "https://custom.example.com/device",
		TokenURL:      "https://custom.example.com/token",
	}

	client := NewDeviceClient(config, WithDeviceHTTPClient(customHTTP))

	if client.httpClient != customHTTP {
		t.Error("expected custom HTTP client to be set")
	}
	if client.config.DeviceCodeURL != "https://custom.example.com/device" {
		t.Errorf("expected custom DeviceCodeURL, got %q", client.config.DeviceCodeURL)
	}
}

func TestDeviceClient_RequestDeviceCode(t *testing.T) {
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
		if r.Form.Get("client_id") != "test-client-id" {
			t.Errorf("expected client_id test-client-id, got %s", r.Form.Get("client_id"))
		}

		resp := map[string]any{
			"device_code":               "test-device-code",
			"user_code":                 "ABCD-1234",
			"verification_uri":          "https://www.google.com/device",
			"verification_uri_complete": "https://www.google.com/device?user_code=ABCD-1234",
			"expires_in":                1800,
			"interval":                  5,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewDeviceClient(DeviceConfig{
		ClientID:      "test-client-id",
		Scopes:        []string{ScopeLiveChat},
		DeviceCodeURL: server.URL,
	})

	authResp, err := client.RequestDeviceCode(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if authResp.DeviceCode != "test-device-code" {
		t.Errorf("expected DeviceCode %q, got %q", "test-device-code", authResp.DeviceCode)
	}
	if authResp.UserCode != "ABCD-1234" {
		t.Errorf("expected UserCode %q, got %q", "ABCD-1234", authResp.UserCode)
	}
	if authResp.VerificationURL != "https://www.google.com/device" {
		t.Errorf("expected VerificationURL %q, got %q", "https://www.google.com/device", authResp.VerificationURL)
	}
	if authResp.ExpiresIn != 1800 {
		t.Errorf("expected ExpiresIn %d, got %d", 1800, authResp.ExpiresIn)
	}
	if authResp.Interval != 5 {
		t.Errorf("expected Interval %d, got %d", 5, authResp.Interval)
	}
}

func TestDeviceClient_RequestDeviceCode_DefaultInterval(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"device_code":      "test-device-code",
			"user_code":        "ABCD-1234",
			"verification_uri": "https://www.google.com/device",
			"expires_in":       1800,
			// No interval field
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewDeviceClient(DeviceConfig{
		ClientID:      "test-client-id",
		Scopes:        []string{ScopeLiveChat},
		DeviceCodeURL: server.URL,
	})

	authResp, err := client.RequestDeviceCode(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if authResp.Interval != 5 {
		t.Errorf("expected default Interval %d, got %d", 5, authResp.Interval)
	}
}

func TestDeviceClient_RequestDeviceCode_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]any{
			"error":             "invalid_client",
			"error_description": "The OAuth client was not found.",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewDeviceClient(DeviceConfig{
		ClientID:      "invalid-client",
		Scopes:        []string{ScopeLiveChat},
		DeviceCodeURL: server.URL,
	})

	_, err := client.RequestDeviceCode(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var deviceErr *DeviceAuthError
	if !containsError(err, &deviceErr) {
		t.Fatalf("expected DeviceAuthError, got %T", err)
	}
	if deviceErr.Code != "invalid_client" {
		t.Errorf("expected error code %q, got %q", "invalid_client", deviceErr.Code)
	}
}

func TestDeviceClient_PollForToken_Success(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := callCount.Add(1)

		if count < 3 {
			// First two calls return authorization_pending
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]any{
				"error":             "authorization_pending",
				"error_description": "The authorization request is still pending.",
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Third call returns token
		resp := map[string]any{
			"access_token":  "test-access-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": "test-refresh-token",
			"scope":         "https://www.googleapis.com/auth/youtube",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewDeviceClient(DeviceConfig{
		ClientID: "test-client-id",
		Scopes:   []string{ScopeLiveChat},
		TokenURL: server.URL,
	})

	authResp := &DeviceAuthResponse{
		DeviceCode: "test-device-code",
		UserCode:   "ABCD-1234",
		ExpiresIn:  1800,
		Interval:   1, // 1 second for faster test
		expiry:     time.Now().Add(30 * time.Minute),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	token, err := client.PollForToken(ctx, authResp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token.AccessToken != "test-access-token" {
		t.Errorf("expected AccessToken %q, got %q", "test-access-token", token.AccessToken)
	}
	if token.RefreshToken != "test-refresh-token" {
		t.Errorf("expected RefreshToken %q, got %q", "test-refresh-token", token.RefreshToken)
	}
}

func TestDeviceClient_PollForToken_SlowDown(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := callCount.Add(1)

		if count == 1 {
			// First call returns slow_down
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]any{
				"error":             "slow_down",
				"error_description": "Slow down the polling.",
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Second call returns token
		resp := map[string]any{
			"access_token": "test-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewDeviceClient(DeviceConfig{
		ClientID: "test-client-id",
		Scopes:   []string{ScopeLiveChat},
		TokenURL: server.URL,
	})

	authResp := &DeviceAuthResponse{
		DeviceCode: "test-device-code",
		Interval:   1,
		expiry:     time.Now().Add(30 * time.Minute),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	token, err := client.PollForToken(ctx, authResp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token.AccessToken != "test-access-token" {
		t.Errorf("expected AccessToken %q, got %q", "test-access-token", token.AccessToken)
	}

	if callCount.Load() != 2 {
		t.Errorf("expected 2 calls, got %d", callCount.Load())
	}
}

func TestDeviceClient_PollForToken_AccessDenied(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]any{
			"error":             "access_denied",
			"error_description": "The user denied the authorization request.",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewDeviceClient(DeviceConfig{
		ClientID: "test-client-id",
		Scopes:   []string{ScopeLiveChat},
		TokenURL: server.URL,
	})

	authResp := &DeviceAuthResponse{
		DeviceCode: "test-device-code",
		Interval:   1,
		expiry:     time.Now().Add(30 * time.Minute),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.PollForToken(ctx, authResp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var deviceErr *DeviceAuthError
	if !containsError(err, &deviceErr) {
		t.Fatalf("expected DeviceAuthError, got %T", err)
	}
	if !deviceErr.IsAccessDenied() {
		t.Errorf("expected IsAccessDenied() true, got false")
	}
}

func TestDeviceClient_PollForToken_Expired(t *testing.T) {
	client := NewDeviceClient(DeviceConfig{
		ClientID: "test-client-id",
		Scopes:   []string{ScopeLiveChat},
	})

	// Set expiry in the past
	authResp := &DeviceAuthResponse{
		DeviceCode: "test-device-code",
		Interval:   1,
		expiry:     time.Now().Add(-1 * time.Minute),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.PollForToken(ctx, authResp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var deviceErr *DeviceAuthError
	if !containsError(err, &deviceErr) {
		t.Fatalf("expected DeviceAuthError, got %T", err)
	}
	if !deviceErr.IsExpired() {
		t.Errorf("expected IsExpired() true, got false")
	}
}

func TestDeviceClient_PollForToken_NilAuthResp(t *testing.T) {
	client := NewDeviceClient(DeviceConfig{
		ClientID: "test-client-id",
		Scopes:   []string{ScopeLiveChat},
	})

	_, err := client.PollForToken(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeviceClient_PollForToken_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]any{
			"error": "authorization_pending",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewDeviceClient(DeviceConfig{
		ClientID: "test-client-id",
		Scopes:   []string{ScopeLiveChat},
		TokenURL: server.URL,
	})

	authResp := &DeviceAuthResponse{
		DeviceCode: "test-device-code",
		Interval:   1,
		expiry:     time.Now().Add(30 * time.Minute),
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, err := client.PollForToken(ctx, authResp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestDeviceClient_PollForTokenAsync(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := callCount.Add(1)

		if count < 2 {
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]any{
				"error": "authorization_pending",
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
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

	client := NewDeviceClient(DeviceConfig{
		ClientID: "test-client-id",
		Scopes:   []string{ScopeLiveChat},
		TokenURL: server.URL,
	})

	authResp := &DeviceAuthResponse{
		DeviceCode: "test-device-code",
		Interval:   1,
		expiry:     time.Now().Add(30 * time.Minute),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tokenCh, errCh, cancelPoll := client.PollForTokenAsync(ctx, authResp)
	defer cancelPoll()

	select {
	case token := <-tokenCh:
		if token.AccessToken != "test-access-token" {
			t.Errorf("expected AccessToken %q, got %q", "test-access-token", token.AccessToken)
		}
	case err := <-errCh:
		t.Fatalf("unexpected error: %v", err)
	case <-time.After(15 * time.Second):
		t.Fatal("timeout waiting for token")
	}
}

func TestDeviceClient_PollForTokenAsync_Cancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]any{
			"error": "authorization_pending",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewDeviceClient(DeviceConfig{
		ClientID: "test-client-id",
		Scopes:   []string{ScopeLiveChat},
		TokenURL: server.URL,
	})

	authResp := &DeviceAuthResponse{
		DeviceCode: "test-device-code",
		Interval:   1,
		expiry:     time.Now().Add(30 * time.Minute),
	}

	ctx := context.Background()
	tokenCh, errCh, cancelPoll := client.PollForTokenAsync(ctx, authResp)

	// Cancel after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancelPoll()
	}()

	select {
	case <-tokenCh:
		t.Fatal("expected no token")
	case err := <-errCh:
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for cancellation")
	}
}

func TestDeviceAuthResponse_Expired(t *testing.T) {
	tests := []struct {
		name    string
		expiry  time.Time
		expired bool
	}{
		{
			name:    "not expired",
			expiry:  time.Now().Add(30 * time.Minute),
			expired: false,
		},
		{
			name:    "expired",
			expiry:  time.Now().Add(-1 * time.Minute),
			expired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authResp := &DeviceAuthResponse{
				expiry: tt.expiry,
			}
			if authResp.Expired() != tt.expired {
				t.Errorf("expected Expired() = %v, got %v", tt.expired, authResp.Expired())
			}
		})
	}
}

func TestDeviceAuthError(t *testing.T) {
	tests := []struct {
		name                 string
		code                 string
		message              string
		expectedError        string
		isAuthorizationPend  bool
		isSlowDown           bool
		isAccessDenied       bool
		isExpired            bool
	}{
		{
			name:                 "authorization_pending",
			code:                 "authorization_pending",
			message:              "Waiting for user",
			expectedError:        "device auth error: authorization_pending - Waiting for user",
			isAuthorizationPend:  true,
		},
		{
			name:          "slow_down",
			code:          "slow_down",
			message:       "Too many requests",
			expectedError: "device auth error: slow_down - Too many requests",
			isSlowDown:    true,
		},
		{
			name:           "access_denied",
			code:           "access_denied",
			message:        "User denied",
			expectedError:  "device auth error: access_denied - User denied",
			isAccessDenied: true,
		},
		{
			name:          "expired_token",
			code:          "expired_token",
			message:       "Code expired",
			expectedError: "device auth error: expired_token - Code expired",
			isExpired:     true,
		},
		{
			name:          "no message",
			code:          "unknown_error",
			message:       "",
			expectedError: "device auth error: unknown_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &DeviceAuthError{Code: tt.code, Message: tt.message}

			if err.Error() != tt.expectedError {
				t.Errorf("expected Error() = %q, got %q", tt.expectedError, err.Error())
			}
			if err.IsAuthorizationPending() != tt.isAuthorizationPend {
				t.Errorf("expected IsAuthorizationPending() = %v, got %v", tt.isAuthorizationPend, err.IsAuthorizationPending())
			}
			if err.IsSlowDown() != tt.isSlowDown {
				t.Errorf("expected IsSlowDown() = %v, got %v", tt.isSlowDown, err.IsSlowDown())
			}
			if err.IsAccessDenied() != tt.isAccessDenied {
				t.Errorf("expected IsAccessDenied() = %v, got %v", tt.isAccessDenied, err.IsAccessDenied())
			}
			if err.IsExpired() != tt.isExpired {
				t.Errorf("expected IsExpired() = %v, got %v", tt.isExpired, err.IsExpired())
			}
		})
	}
}

// containsError checks if err contains an error of type T.
func containsError[T error](err error, target *T) bool {
	return errors.As(err, target)
}
