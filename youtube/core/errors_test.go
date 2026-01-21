package core

import (
	"math/rand"
	"testing"
	"time"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *APIError
		want string
	}{
		{
			name: "with code",
			err:  &APIError{StatusCode: 403, Code: "forbidden", Message: "access denied"},
			want: "youtube api: forbidden (403): access denied",
		},
		{
			name: "without code",
			err:  &APIError{StatusCode: 500, Message: "internal error"},
			want: "youtube api: status 500: internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("APIError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAPIError_Helpers(t *testing.T) {
	tests := []struct {
		name           string
		err            *APIError
		isQuota        bool
		isNotFound     bool
		isForbidden    bool
	}{
		{
			name:      "quota exceeded",
			err:       &APIError{Code: "quotaExceeded"},
			isQuota:   true,
		},
		{
			name:       "not found by code",
			err:        &APIError{Code: "notFound"},
			isNotFound: true,
		},
		{
			name:       "not found by status",
			err:        &APIError{StatusCode: 404},
			isNotFound: true,
		},
		{
			name:        "forbidden by code",
			err:         &APIError{Code: "forbidden"},
			isForbidden: true,
		},
		{
			name:        "forbidden by status",
			err:         &APIError{StatusCode: 403},
			isForbidden: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.IsQuotaExceeded(); got != tt.isQuota {
				t.Errorf("IsQuotaExceeded() = %v, want %v", got, tt.isQuota)
			}
			if got := tt.err.IsNotFound(); got != tt.isNotFound {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.isNotFound)
			}
			if got := tt.err.IsForbidden(); got != tt.isForbidden {
				t.Errorf("IsForbidden() = %v, want %v", got, tt.isForbidden)
			}
		})
	}
}

func TestQuotaError_Error(t *testing.T) {
	resetAt := time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC)
	err := &QuotaError{Used: 10000, Limit: 10000, ResetAt: resetAt}

	got := err.Error()
	if got == "" {
		t.Error("QuotaError.Error() returned empty string")
	}
	// Should contain key info
	if !contains(got, "quota exceeded") {
		t.Errorf("QuotaError.Error() should contain 'quota exceeded', got %q", got)
	}
	if !contains(got, "10000/10000") {
		t.Errorf("QuotaError.Error() should contain usage info, got %q", got)
	}
}

func TestRateLimitError_Error(t *testing.T) {
	err := &RateLimitError{RetryAfter: 5 * time.Second}
	got := err.Error()

	if !contains(got, "rate limited") {
		t.Errorf("RateLimitError.Error() should contain 'rate limited', got %q", got)
	}
	if !contains(got, "5s") {
		t.Errorf("RateLimitError.Error() should contain retry duration, got %q", got)
	}
}

func TestAuthError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *AuthError
		want string
	}{
		{
			name: "with code",
			err:  &AuthError{Code: "invalid_grant", Message: "token expired"},
			want: "youtube auth: invalid_grant: token expired",
		},
		{
			name: "without code",
			err:  &AuthError{Message: "authentication failed"},
			want: "youtube auth: authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("AuthError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAuthError_IsExpiredToken(t *testing.T) {
	tests := []struct {
		code    string
		expired bool
	}{
		{"expired_token", true},
		{"invalid_grant", true},
		{"invalid_client", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			err := &AuthError{Code: tt.code}
			if got := err.IsExpiredToken(); got != tt.expired {
				t.Errorf("IsExpiredToken() = %v, want %v", got, tt.expired)
			}
		})
	}
}

func TestChatEndedError_Error(t *testing.T) {
	err := &ChatEndedError{LiveChatID: "abc123"}
	got := err.Error()

	if !contains(got, "live chat ended") {
		t.Errorf("ChatEndedError.Error() should contain 'live chat ended', got %q", got)
	}
	if !contains(got, "abc123") {
		t.Errorf("ChatEndedError.Error() should contain chat ID, got %q", got)
	}
}

func TestNotFoundError_Error(t *testing.T) {
	err := &NotFoundError{ResourceType: "video", ResourceID: "vid123"}
	got := err.Error()

	if !contains(got, "not found") {
		t.Errorf("NotFoundError.Error() should contain 'not found', got %q", got)
	}
	if !contains(got, "video") {
		t.Errorf("NotFoundError.Error() should contain resource type, got %q", got)
	}
	if !contains(got, "vid123") {
		t.Errorf("NotFoundError.Error() should contain resource ID, got %q", got)
	}
}

func TestBackoffConfig_Delay(t *testing.T) {
	// Use fixed random for deterministic tests
	backoff := &BackoffConfig{
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
		Multiplier: 2.0,
		Jitter:     0.0, // No jitter for exact testing
		RandFloat:  func() float64 { return 0.5 },
	}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 16 * time.Second},
		{5, 30 * time.Second}, // Capped at max
		{6, 30 * time.Second}, // Still capped
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := backoff.Delay(tt.attempt)
			if got != tt.want {
				t.Errorf("Delay(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

func TestBackoffConfig_DelayWithJitter(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	backoff := NewBackoffConfig(WithRandSource(rng))

	// With jitter, delays should vary but stay within bounds
	for range 10 {
		delay := backoff.Delay(0)
		minDelay := time.Duration(float64(backoff.BaseDelay) * (1 - backoff.Jitter))
		maxDelay := time.Duration(float64(backoff.BaseDelay) * (1 + backoff.Jitter))

		if delay < minDelay || delay > maxDelay {
			t.Errorf("Delay with jitter out of bounds: %v not in [%v, %v]", delay, minDelay, maxDelay)
		}
	}
}

func TestNewBackoffConfig(t *testing.T) {
	// Test defaults
	b := NewBackoffConfig()
	if b.BaseDelay != 1*time.Second {
		t.Errorf("BaseDelay = %v, want 1s", b.BaseDelay)
	}
	if b.MaxDelay != 30*time.Second {
		t.Errorf("MaxDelay = %v, want 30s", b.MaxDelay)
	}
	if b.Multiplier != 2.0 {
		t.Errorf("Multiplier = %v, want 2.0", b.Multiplier)
	}
	if b.Jitter != 0.2 {
		t.Errorf("Jitter = %v, want 0.2", b.Jitter)
	}
}

func TestBackoffOptions(t *testing.T) {
	b := NewBackoffConfig(
		WithBaseDelay(2*time.Second),
		WithMaxDelay(60*time.Second),
		WithMultiplier(3.0),
		WithJitter(0.5),
	)

	if b.BaseDelay != 2*time.Second {
		t.Errorf("WithBaseDelay failed: got %v", b.BaseDelay)
	}
	if b.MaxDelay != 60*time.Second {
		t.Errorf("WithMaxDelay failed: got %v", b.MaxDelay)
	}
	if b.Multiplier != 3.0 {
		t.Errorf("WithMultiplier failed: got %v", b.Multiplier)
	}
	if b.Jitter != 0.5 {
		t.Errorf("WithJitter failed: got %v", b.Jitter)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestInvalidTransitionError(t *testing.T) {
	err := &InvalidTransitionError{
		BroadcastID:    "broadcast123",
		CurrentState:   "created",
		RequestedState: "live",
	}

	msg := err.Error()
	if !contains(msg, "broadcast123") {
		t.Errorf("error message should contain broadcast ID")
	}
	if !contains(msg, "created") {
		t.Errorf("error message should contain current state")
	}
	if !contains(msg, "live") {
		t.Errorf("error message should contain requested state")
	}
	if !contains(msg, "invalid transition") {
		t.Errorf("error message should contain 'invalid transition'")
	}
}

func TestStreamNotHealthyError(t *testing.T) {
	t.Run("with issues", func(t *testing.T) {
		err := &StreamNotHealthyError{
			StreamID: "stream123",
			Status:   "bad",
			Issues:   []string{"no video", "no audio"},
		}

		msg := err.Error()
		if !contains(msg, "stream123") {
			t.Errorf("error message should contain stream ID")
		}
		if !contains(msg, "bad") {
			t.Errorf("error message should contain status")
		}
		if !contains(msg, "no video") {
			t.Errorf("error message should contain issues")
		}
	})

	t.Run("without issues", func(t *testing.T) {
		err := &StreamNotHealthyError{
			StreamID: "stream123",
			Status:   "bad",
		}

		msg := err.Error()
		if !contains(msg, "stream123") {
			t.Errorf("error message should contain stream ID")
		}
		if !contains(msg, "bad") {
			t.Errorf("error message should contain status")
		}
	})
}

func TestStreamNotBoundError(t *testing.T) {
	err := &StreamNotBoundError{
		BroadcastID: "broadcast123",
	}

	msg := err.Error()
	if !contains(msg, "broadcast123") {
		t.Errorf("error message should contain broadcast ID")
	}
	if !contains(msg, "no bound stream") {
		t.Errorf("error message should contain 'no bound stream'")
	}
}
