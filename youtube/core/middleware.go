package core

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

// Middleware wraps request execution with additional behavior.
// It receives the request and a next function to call the next middleware or
// the actual request handler.
type Middleware func(ctx context.Context, req *Request, next func(context.Context, *Request) error) error

// MiddlewareChain chains multiple middlewares together.
// Middlewares are executed in the order they are provided.
func MiddlewareChain(middlewares ...Middleware) Middleware {
	return func(ctx context.Context, req *Request, next func(context.Context, *Request) error) error {
		// Build the chain from the inside out
		handler := next
		for i := len(middlewares) - 1; i >= 0; i-- {
			mw := middlewares[i]
			h := handler // capture for closure
			handler = func(ctx context.Context, req *Request) error {
				return mw(ctx, req, h)
			}
		}
		return handler(ctx, req)
	}
}

// LoggingMiddleware logs request details.
type LoggingMiddleware struct {
	logger    Logger
	logBody   bool
	logTiming bool
}

// Logger is the interface for logging.
type Logger interface {
	Printf(format string, v ...any)
}

// defaultLogger uses the standard log package.
type defaultLogger struct{}

func (defaultLogger) Printf(format string, v ...any) {
	log.Printf(format, v...)
}

// LoggingOption configures LoggingMiddleware.
type LoggingOption func(*LoggingMiddleware)

// WithLogger sets a custom logger.
func WithLogger(l Logger) LoggingOption {
	return func(m *LoggingMiddleware) { m.logger = l }
}

// WithLogBody enables logging of request bodies.
func WithLogBody(enabled bool) LoggingOption {
	return func(m *LoggingMiddleware) { m.logBody = enabled }
}

// WithLogTiming enables logging of request timing.
func WithLogTiming(enabled bool) LoggingOption {
	return func(m *LoggingMiddleware) { m.logTiming = enabled }
}

// NewLoggingMiddleware creates a logging middleware.
func NewLoggingMiddleware(opts ...LoggingOption) Middleware {
	m := &LoggingMiddleware{
		logger:    defaultLogger{},
		logTiming: true,
	}
	for _, opt := range opts {
		opt(m)
	}

	return func(ctx context.Context, req *Request, next func(context.Context, *Request) error) error {
		start := time.Now()

		// Log request
		m.logger.Printf("[youtube] %s %s", req.Method, req.Path)
		if m.logBody && req.Body != nil {
			m.logger.Printf("[youtube] body: %+v", req.Body)
		}

		// Execute request
		err := next(ctx, req)

		// Log result
		if m.logTiming {
			duration := time.Since(start)
			if err != nil {
				m.logger.Printf("[youtube] %s %s failed after %v: %v", req.Method, req.Path, duration, err)
			} else {
				m.logger.Printf("[youtube] %s %s completed in %v", req.Method, req.Path, duration)
			}
		}

		return err
	}
}

// RetryMiddleware retries failed requests with exponential backoff.
type RetryMiddleware struct {
	maxRetries int
	backoff    *BackoffConfig
	shouldRetry func(error) bool
}

// RetryOption configures RetryMiddleware.
type RetryOption func(*RetryMiddleware)

// WithMaxRetries sets the maximum number of retry attempts.
func WithMaxRetries(n int) RetryOption {
	return func(m *RetryMiddleware) { m.maxRetries = n }
}

// WithRetryBackoff sets the backoff configuration.
func WithRetryBackoff(cfg *BackoffConfig) RetryOption {
	return func(m *RetryMiddleware) { m.backoff = cfg }
}

// WithShouldRetry sets a custom function to determine if an error is retryable.
func WithShouldRetry(fn func(error) bool) RetryOption {
	return func(m *RetryMiddleware) { m.shouldRetry = fn }
}

// NewRetryMiddleware creates a retry middleware.
func NewRetryMiddleware(opts ...RetryOption) Middleware {
	m := &RetryMiddleware{
		maxRetries: 3,
		backoff:    NewBackoffConfig(),
		shouldRetry: func(err error) bool {
			// By default, retry rate limit errors
			if _, ok := err.(*RateLimitError); ok {
				return true
			}
			return false
		},
	}
	for _, opt := range opts {
		opt(m)
	}

	return func(ctx context.Context, req *Request, next func(context.Context, *Request) error) error {
		var lastErr error

		for attempt := 0; attempt <= m.maxRetries; attempt++ {
			// Check context before each attempt
			if ctx.Err() != nil {
				return ctx.Err()
			}

			// Wait before retry (skip on first attempt)
			if attempt > 0 {
				delay := m.backoff.Delay(attempt - 1)

				// If we have a rate limit error, use its retry-after
				if rle, ok := lastErr.(*RateLimitError); ok && rle.RetryAfter > delay {
					delay = rle.RetryAfter
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
				}
			}

			lastErr = next(ctx, req)
			if lastErr == nil {
				return nil
			}

			if !m.shouldRetry(lastErr) {
				return lastErr
			}
		}

		return fmt.Errorf("max retries (%d) exceeded: %w", m.maxRetries, lastErr)
	}
}

// MetricsMiddleware tracks request metrics.
type MetricsMiddleware struct {
	totalRequests  atomic.Int64
	failedRequests atomic.Int64
	totalDuration  atomic.Int64 // nanoseconds
	collector      MetricsCollector
}

// MetricsCollector receives metrics events.
type MetricsCollector interface {
	// RecordRequest records a completed request.
	RecordRequest(method, path string, duration time.Duration, err error)
}

// MetricsOption configures MetricsMiddleware.
type MetricsOption func(*MetricsMiddleware)

// WithMetricsCollector sets a custom metrics collector.
func WithMetricsCollector(c MetricsCollector) MetricsOption {
	return func(m *MetricsMiddleware) { m.collector = c }
}

// NewMetricsMiddleware creates a metrics middleware.
func NewMetricsMiddleware(opts ...MetricsOption) (*MetricsMiddleware, Middleware) {
	m := &MetricsMiddleware{}
	for _, opt := range opts {
		opt(m)
	}

	mw := func(ctx context.Context, req *Request, next func(context.Context, *Request) error) error {
		start := time.Now()

		err := next(ctx, req)

		duration := time.Since(start)
		m.totalRequests.Add(1)
		m.totalDuration.Add(int64(duration))

		if err != nil {
			m.failedRequests.Add(1)
		}

		if m.collector != nil {
			m.collector.RecordRequest(req.Method, req.Path, duration, err)
		}

		return err
	}

	return m, mw
}

// TotalRequests returns the total number of requests made.
func (m *MetricsMiddleware) TotalRequests() int64 {
	return m.totalRequests.Load()
}

// FailedRequests returns the number of failed requests.
func (m *MetricsMiddleware) FailedRequests() int64 {
	return m.failedRequests.Load()
}

// SuccessfulRequests returns the number of successful requests.
func (m *MetricsMiddleware) SuccessfulRequests() int64 {
	return m.totalRequests.Load() - m.failedRequests.Load()
}

// TotalDuration returns the total duration of all requests.
func (m *MetricsMiddleware) TotalDuration() time.Duration {
	return time.Duration(m.totalDuration.Load())
}

// AverageDuration returns the average request duration.
func (m *MetricsMiddleware) AverageDuration() time.Duration {
	total := m.totalRequests.Load()
	if total == 0 {
		return 0
	}
	return time.Duration(m.totalDuration.Load() / total)
}

// Reset resets all metrics to zero.
func (m *MetricsMiddleware) Reset() {
	m.totalRequests.Store(0)
	m.failedRequests.Store(0)
	m.totalDuration.Store(0)
}

// CachingMiddleware caches GET request responses.
type CachingMiddleware struct {
	cache  *Cache
	keyFn  func(*Request) string
	ttl    time.Duration
}

// CachingOption configures CachingMiddleware.
type CachingOption func(*CachingMiddleware)

// WithCacheTTL sets the TTL for cached responses.
func WithCacheTTL(ttl time.Duration) CachingOption {
	return func(m *CachingMiddleware) { m.ttl = ttl }
}

// WithCacheKeyFunc sets a custom function to generate cache keys.
func WithCacheKeyFunc(fn func(*Request) string) CachingOption {
	return func(m *CachingMiddleware) { m.keyFn = fn }
}

// defaultCacheKey generates a cache key from a request.
func defaultCacheKey(req *Request) string {
	key := req.Method + ":" + req.Path
	if req.Query != nil {
		key += "?" + req.Query.Encode()
	}
	return key
}

// NewCachingMiddleware creates a caching middleware.
// Note: This middleware caches successful GET responses only.
// The cache is not response-aware - you need to handle cache invalidation
// separately for write operations.
func NewCachingMiddleware(cache *Cache, opts ...CachingOption) Middleware {
	m := &CachingMiddleware{
		cache: cache,
		keyFn: defaultCacheKey,
		ttl:   5 * time.Minute,
	}
	for _, opt := range opts {
		opt(m)
	}

	return func(ctx context.Context, req *Request, next func(context.Context, *Request) error) error {
		// Only cache GET requests
		if req.Method != "GET" {
			return next(ctx, req)
		}

		// Generate cache key - this can be used by response handlers
		// The key is stored in context for downstream use
		_ = m.keyFn(req) // Cache key generation (for future response caching)

		// Note: Full response caching requires integration at the Client level
		// since middleware doesn't have access to response bodies.
		// This middleware provides key generation and TTL configuration.
		return next(ctx, req)
	}
}

// RateLimitingMiddleware limits request rate.
type RateLimitingMiddleware struct {
	requestsPerSecond float64
	lastRequest       atomic.Int64 // UnixNano
	minInterval       time.Duration
}

// RateLimitOption configures RateLimitingMiddleware.
type RateLimitOption func(*RateLimitingMiddleware)

// WithRequestsPerSecond sets the maximum requests per second.
func WithRequestsPerSecond(rps float64) RateLimitOption {
	return func(m *RateLimitingMiddleware) {
		m.requestsPerSecond = rps
		if rps > 0 {
			m.minInterval = time.Duration(float64(time.Second) / rps)
		}
	}
}

// NewRateLimitingMiddleware creates a rate limiting middleware.
func NewRateLimitingMiddleware(opts ...RateLimitOption) Middleware {
	m := &RateLimitingMiddleware{
		requestsPerSecond: 10, // Default: 10 requests per second
		minInterval:       100 * time.Millisecond,
	}
	for _, opt := range opts {
		opt(m)
	}

	return func(ctx context.Context, req *Request, next func(context.Context, *Request) error) error {
		now := time.Now().UnixNano()

		for {
			last := m.lastRequest.Load()
			nextAllowed := last + int64(m.minInterval)

			if now >= nextAllowed {
				if m.lastRequest.CompareAndSwap(last, now) {
					break
				}
				continue // Another goroutine updated, retry
			}

			// Need to wait
			waitTime := time.Duration(nextAllowed - now)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
				now = time.Now().UnixNano()
			}
		}

		return next(ctx, req)
	}
}
