package core

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Sentinel errors for common conditions.
var (
	ErrAlreadyRunning = errors.New("already running")
	ErrNotRunning     = errors.New("not running")
)

// ErrorDetail contains additional information about an API error.
type ErrorDetail struct {
	Type     string `json:"@type,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// APIError represents an error returned by the YouTube API.
type APIError struct {
	StatusCode int
	Code       string // e.g., "quotaExceeded", "forbidden"
	Message    string
	Details    []ErrorDetail
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("youtube api: %s (%d): %s", e.Code, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("youtube api: status %d: %s", e.StatusCode, e.Message)
}

// IsQuotaExceeded returns true if this error indicates quota exhaustion.
func (e *APIError) IsQuotaExceeded() bool {
	return e.Code == "quotaExceeded"
}

// IsNotFound returns true if this error indicates a resource was not found.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404 || e.Code == "notFound"
}

// IsForbidden returns true if this error indicates access is forbidden.
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == 403 || e.Code == "forbidden"
}

// QuotaError indicates the daily API quota has been exceeded.
// YouTube API quotas reset at midnight Pacific Time.
type QuotaError struct {
	Used    int
	Limit   int
	ResetAt time.Time // Pacific midnight as per Google official documentation
}

func (e *QuotaError) Error() string {
	return fmt.Sprintf("youtube api: quota exceeded (%d/%d), resets at %s",
		e.Used, e.Limit, e.ResetAt.Format(time.RFC3339))
}

// RateLimitError indicates a per-second rate limit was exceeded.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("youtube api: rate limited, retry after %s", e.RetryAfter)
}

// AuthError indicates an authentication or authorization failure.
type AuthError struct {
	Code    string // e.g., "invalid_grant", "expired_token"
	Message string
}

func (e *AuthError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("youtube auth: %s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("youtube auth: %s", e.Message)
}

// IsExpiredToken returns true if the token has expired.
func (e *AuthError) IsExpiredToken() bool {
	return e.Code == "expired_token" || e.Code == "invalid_grant"
}

// ChatEndedError indicates the live chat has ended and is no longer available.
type ChatEndedError struct {
	LiveChatID string
}

func (e *ChatEndedError) Error() string {
	return fmt.Sprintf("youtube api: live chat ended: %s", e.LiveChatID)
}

// BackoffConfig configures exponential backoff with jitter for retry logic.
type BackoffConfig struct {
	BaseDelay  time.Duration  // Initial delay (default: 1s)
	MaxDelay   time.Duration  // Maximum delay cap (default: 30s)
	Multiplier float64        // Exponential multiplier (default: 2.0)
	Jitter     float64        // Jitter factor 0-1 (default: 0.2 = ±20%)
	RandFloat  func() float64 // Random source [0,1) - injectable for testing
}

// Delay calculates the backoff delay for the given attempt number (0-indexed).
func (b *BackoffConfig) Delay(attempt int) time.Duration {
	delay := float64(b.BaseDelay) * math.Pow(b.Multiplier, float64(attempt))
	if delay > float64(b.MaxDelay) {
		delay = float64(b.MaxDelay)
	}
	// Add jitter: delay * (1 ± jitter)
	jitterRange := delay * b.Jitter
	randFn := b.RandFloat
	if randFn == nil {
		randFn = rand.Float64
	}
	jitter := (randFn()*2 - 1) * jitterRange
	return time.Duration(delay + jitter)
}

// BackoffOption configures a BackoffConfig.
type BackoffOption func(*BackoffConfig)

// NewBackoffConfig creates a BackoffConfig with sensible defaults.
func NewBackoffConfig(opts ...BackoffOption) *BackoffConfig {
	b := &BackoffConfig{
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
		Multiplier: 2.0,
		Jitter:     0.2,
		RandFloat:  rand.Float64,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// WithRandSource sets a custom random source for deterministic testing.
func WithRandSource(rng *rand.Rand) BackoffOption {
	return func(b *BackoffConfig) { b.RandFloat = rng.Float64 }
}

// WithBaseDelay sets the initial backoff delay.
func WithBaseDelay(d time.Duration) BackoffOption {
	return func(b *BackoffConfig) { b.BaseDelay = d }
}

// WithMaxDelay sets the maximum backoff delay cap.
func WithMaxDelay(d time.Duration) BackoffOption {
	return func(b *BackoffConfig) { b.MaxDelay = d }
}

// WithMultiplier sets the exponential multiplier.
func WithMultiplier(m float64) BackoffOption {
	return func(b *BackoffConfig) { b.Multiplier = m }
}

// WithJitter sets the jitter factor (0-1).
func WithJitter(j float64) BackoffOption {
	return func(b *BackoffConfig) { b.Jitter = j }
}
