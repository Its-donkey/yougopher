package core

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMiddlewareChain(t *testing.T) {
	var order []string

	mw1 := func(ctx context.Context, req *Request, next func(context.Context, *Request) error) error {
		order = append(order, "mw1-before")
		err := next(ctx, req)
		order = append(order, "mw1-after")
		return err
	}

	mw2 := func(ctx context.Context, req *Request, next func(context.Context, *Request) error) error {
		order = append(order, "mw2-before")
		err := next(ctx, req)
		order = append(order, "mw2-after")
		return err
	}

	chain := MiddlewareChain(mw1, mw2)

	handler := func(ctx context.Context, req *Request) error {
		order = append(order, "handler")
		return nil
	}

	err := chain(context.Background(), &Request{}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(order) != len(expected) {
		t.Fatalf("order = %v, want %v", order, expected)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("order[%d] = %q, want %q", i, order[i], v)
		}
	}
}

type testLogger struct {
	mu   sync.Mutex
	logs []string
}

func (l *testLogger) Printf(format string, v ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, strings.TrimSpace(log.Default().Prefix()+format))
}

func TestLoggingMiddleware(t *testing.T) {
	t.Run("basic logging", func(t *testing.T) {
		logger := &testLogger{}
		mw := NewLoggingMiddleware(
			WithLogger(logger),
			WithLogTiming(true),
		)

		handler := func(ctx context.Context, req *Request) error {
			return nil
		}

		req := &Request{Method: "GET", Path: "/videos"}
		err := mw(context.Background(), req, handler)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(logger.logs) != 2 {
			t.Fatalf("expected 2 log entries, got %d: %v", len(logger.logs), logger.logs)
		}
	})

	t.Run("logs error", func(t *testing.T) {
		logger := &testLogger{}
		mw := NewLoggingMiddleware(WithLogger(logger))

		handler := func(ctx context.Context, req *Request) error {
			return errors.New("request failed")
		}

		req := &Request{Method: "POST", Path: "/messages"}
		err := mw(context.Background(), req, handler)
		if err == nil {
			t.Fatal("expected error")
		}

		// Should log the error
		found := false
		for _, log := range logger.logs {
			if strings.Contains(log, "failed") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected error to be logged")
		}
	})

	t.Run("logs body", func(t *testing.T) {
		logger := &testLogger{}
		mw := NewLoggingMiddleware(
			WithLogger(logger),
			WithLogBody(true),
		)

		handler := func(ctx context.Context, req *Request) error {
			return nil
		}

		req := &Request{Method: "POST", Path: "/messages", Body: map[string]string{"text": "hello"}}
		err := mw(context.Background(), req, handler)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should log the body
		found := false
		for _, log := range logger.logs {
			if strings.Contains(log, "body") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected body to be logged")
		}
	})
}

func TestRetryMiddleware(t *testing.T) {
	t.Run("no retry on success", func(t *testing.T) {
		callCount := 0
		mw := NewRetryMiddleware(WithMaxRetries(3))

		handler := func(ctx context.Context, req *Request) error {
			callCount++
			return nil
		}

		err := mw(context.Background(), &Request{}, handler)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if callCount != 1 {
			t.Errorf("callCount = %d, want 1", callCount)
		}
	})

	t.Run("retries on rate limit error", func(t *testing.T) {
		callCount := 0
		mw := NewRetryMiddleware(
			WithMaxRetries(3),
			WithRetryBackoff(&BackoffConfig{
				BaseDelay:  1 * time.Millisecond,
				MaxDelay:   10 * time.Millisecond,
				Multiplier: 2.0,
				Jitter:     0,
				RandFloat:  func() float64 { return 0.5 },
			}),
		)

		handler := func(ctx context.Context, req *Request) error {
			callCount++
			if callCount < 3 {
				return &RateLimitError{RetryAfter: 1 * time.Millisecond}
			}
			return nil
		}

		err := mw(context.Background(), &Request{}, handler)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if callCount != 3 {
			t.Errorf("callCount = %d, want 3", callCount)
		}
	})

	t.Run("no retry on non-retryable error", func(t *testing.T) {
		callCount := 0
		mw := NewRetryMiddleware(WithMaxRetries(3))

		handler := func(ctx context.Context, req *Request) error {
			callCount++
			return errors.New("permanent error")
		}

		err := mw(context.Background(), &Request{}, handler)
		if err == nil {
			t.Fatal("expected error")
		}
		if callCount != 1 {
			t.Errorf("callCount = %d, want 1", callCount)
		}
	})

	t.Run("max retries exceeded", func(t *testing.T) {
		callCount := 0
		mw := NewRetryMiddleware(
			WithMaxRetries(2),
			WithRetryBackoff(&BackoffConfig{
				BaseDelay: 1 * time.Millisecond,
				MaxDelay:  10 * time.Millisecond,
				Jitter:    0,
				RandFloat: func() float64 { return 0.5 },
			}),
		)

		handler := func(ctx context.Context, req *Request) error {
			callCount++
			return &RateLimitError{RetryAfter: 1 * time.Millisecond}
		}

		err := mw(context.Background(), &Request{}, handler)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "max retries") {
			t.Errorf("error = %v, want max retries message", err)
		}
		if callCount != 3 { // 1 initial + 2 retries
			t.Errorf("callCount = %d, want 3", callCount)
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		mw := NewRetryMiddleware(WithMaxRetries(3))

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		handler := func(ctx context.Context, req *Request) error {
			return &RateLimitError{RetryAfter: 1 * time.Second}
		}

		err := mw(ctx, &Request{}, handler)
		if err != context.Canceled {
			t.Errorf("error = %v, want %v", err, context.Canceled)
		}
	})

	t.Run("custom should retry", func(t *testing.T) {
		callCount := 0
		mw := NewRetryMiddleware(
			WithMaxRetries(3),
			WithShouldRetry(func(err error) bool {
				return strings.Contains(err.Error(), "retry me")
			}),
			WithRetryBackoff(&BackoffConfig{
				BaseDelay: 1 * time.Millisecond,
				Jitter:    0,
				RandFloat: func() float64 { return 0.5 },
			}),
		)

		handler := func(ctx context.Context, req *Request) error {
			callCount++
			if callCount < 2 {
				return errors.New("retry me")
			}
			return nil
		}

		err := mw(context.Background(), &Request{}, handler)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if callCount != 2 {
			t.Errorf("callCount = %d, want 2", callCount)
		}
	})
}

func TestMetricsMiddleware(t *testing.T) {
	t.Run("tracks successful requests", func(t *testing.T) {
		metrics, mw := NewMetricsMiddleware()

		handler := func(ctx context.Context, req *Request) error {
			time.Sleep(1 * time.Millisecond)
			return nil
		}

		for i := 0; i < 5; i++ {
			err := mw(context.Background(), &Request{Method: "GET", Path: "/videos"}, handler)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}

		if metrics.TotalRequests() != 5 {
			t.Errorf("TotalRequests() = %d, want 5", metrics.TotalRequests())
		}
		if metrics.FailedRequests() != 0 {
			t.Errorf("FailedRequests() = %d, want 0", metrics.FailedRequests())
		}
		if metrics.SuccessfulRequests() != 5 {
			t.Errorf("SuccessfulRequests() = %d, want 5", metrics.SuccessfulRequests())
		}
		if metrics.TotalDuration() == 0 {
			t.Error("TotalDuration() should be > 0")
		}
		if metrics.AverageDuration() == 0 {
			t.Error("AverageDuration() should be > 0")
		}
	})

	t.Run("tracks failed requests", func(t *testing.T) {
		metrics, mw := NewMetricsMiddleware()

		handler := func(ctx context.Context, req *Request) error {
			return errors.New("request failed")
		}

		for i := 0; i < 3; i++ {
			_ = mw(context.Background(), &Request{}, handler)
		}

		if metrics.TotalRequests() != 3 {
			t.Errorf("TotalRequests() = %d, want 3", metrics.TotalRequests())
		}
		if metrics.FailedRequests() != 3 {
			t.Errorf("FailedRequests() = %d, want 3", metrics.FailedRequests())
		}
	})

	t.Run("reset", func(t *testing.T) {
		metrics, mw := NewMetricsMiddleware()

		handler := func(ctx context.Context, req *Request) error {
			return nil
		}

		_ = mw(context.Background(), &Request{}, handler)
		metrics.Reset()

		if metrics.TotalRequests() != 0 {
			t.Errorf("TotalRequests() = %d, want 0", metrics.TotalRequests())
		}
	})

	t.Run("average duration with no requests", func(t *testing.T) {
		metrics, _ := NewMetricsMiddleware()

		if metrics.AverageDuration() != 0 {
			t.Errorf("AverageDuration() = %v, want 0", metrics.AverageDuration())
		}
	})

	t.Run("with collector", func(t *testing.T) {
		var collected []struct {
			method   string
			path     string
			duration time.Duration
			err      error
		}

		collector := &testCollector{
			record: func(method, path string, duration time.Duration, err error) {
				collected = append(collected, struct {
					method   string
					path     string
					duration time.Duration
					err      error
				}{method, path, duration, err})
			},
		}

		_, mw := NewMetricsMiddleware(WithMetricsCollector(collector))

		handler := func(ctx context.Context, req *Request) error {
			return nil
		}

		_ = mw(context.Background(), &Request{Method: "GET", Path: "/test"}, handler)

		if len(collected) != 1 {
			t.Fatalf("expected 1 collected, got %d", len(collected))
		}
		if collected[0].method != "GET" || collected[0].path != "/test" {
			t.Error("collected wrong request info")
		}
	})
}

type testCollector struct {
	record func(method, path string, duration time.Duration, err error)
}

func (c *testCollector) RecordRequest(method, path string, duration time.Duration, err error) {
	c.record(method, path, duration, err)
}

func TestRateLimitingMiddleware(t *testing.T) {
	t.Run("limits rate", func(t *testing.T) {
		// 100 requests per second = 10ms between requests
		mw := NewRateLimitingMiddleware(WithRequestsPerSecond(100))

		callCount := 0
		handler := func(ctx context.Context, req *Request) error {
			callCount++
			return nil
		}

		start := time.Now()
		for i := 0; i < 5; i++ {
			err := mw(context.Background(), &Request{}, handler)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}
		elapsed := time.Since(start)

		// Should take at least ~40ms for 5 requests at 100/sec
		// (first is immediate, then 4 * 10ms)
		if elapsed < 30*time.Millisecond {
			t.Errorf("elapsed = %v, expected >= 30ms", elapsed)
		}

		if callCount != 5 {
			t.Errorf("callCount = %d, want 5", callCount)
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		mw := NewRateLimitingMiddleware(WithRequestsPerSecond(1)) // Very slow

		// Make one request to start the rate limiter
		_ = mw(context.Background(), &Request{}, func(ctx context.Context, req *Request) error {
			return nil
		})

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		// This should timeout waiting for rate limit
		err := mw(ctx, &Request{}, func(ctx context.Context, req *Request) error {
			return nil
		})

		if err != context.DeadlineExceeded {
			t.Errorf("error = %v, want %v", err, context.DeadlineExceeded)
		}
	})
}

func TestCachingMiddleware(t *testing.T) {
	cache := NewCache()
	mw := NewCachingMiddleware(cache,
		WithCacheTTL(5*time.Minute),
		WithCacheKeyFunc(defaultCacheKey),
	)

	t.Run("passes through GET requests", func(t *testing.T) {
		called := false
		handler := func(ctx context.Context, req *Request) error {
			called = true
			return nil
		}

		err := mw(context.Background(), &Request{Method: "GET", Path: "/videos"}, handler)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("expected handler to be called")
		}
	})

	t.Run("passes through non-GET requests", func(t *testing.T) {
		called := false
		handler := func(ctx context.Context, req *Request) error {
			called = true
			return nil
		}

		err := mw(context.Background(), &Request{Method: "POST", Path: "/messages"}, handler)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("expected handler to be called")
		}
	})
}

func TestDefaultCacheKey(t *testing.T) {
	tests := []struct {
		name string
		req  *Request
		want string
	}{
		{
			name: "simple",
			req:  &Request{Method: "GET", Path: "/videos"},
			want: "GET:/videos",
		},
		{
			name: "with query",
			req: &Request{
				Method: "GET",
				Path:   "/videos",
				Query:  map[string][]string{"id": {"abc123"}},
			},
			want: "GET:/videos?id=abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := defaultCacheKey(tt.req)
			if got != tt.want {
				t.Errorf("defaultCacheKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultLogger(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	logger := defaultLogger{}
	logger.Printf("test message %d", 123)

	if !strings.Contains(buf.String(), "test message 123") {
		t.Errorf("log output = %q, want to contain 'test message 123'", buf.String())
	}
}
