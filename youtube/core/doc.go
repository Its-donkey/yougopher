// Package core provides the base HTTP client, error types, quota tracking,
// caching, and middleware for interacting with YouTube APIs.
//
// # Client
//
// The Client type is the foundation for all API calls. It handles HTTP
// communication, authentication headers, and error parsing.
//
//	client := core.NewClient(
//		core.WithHTTPClient(customClient),
//		core.WithBaseURL("https://www.googleapis.com/youtube/v3"),
//	)
//
// # Error Types
//
// The package defines several error types for different failure scenarios:
//
//   - APIError: General YouTube API errors with status codes and messages
//   - QuotaError: Daily quota exceeded (resets at midnight Pacific Time)
//   - RateLimitError: Per-second rate limit exceeded
//   - AuthError: Authentication and authorization failures
//   - NotFoundError: Resource not found
//
// # Quota Tracking
//
// YouTube API uses a quota system where different operations cost different
// amounts. The QuotaTracker helps monitor usage:
//
//	tracker := core.NewQuotaTracker(10000) // 10,000 daily limit
//	tracker.Add("liveChatMessages.list", 5)
//	remaining := tracker.Remaining()
//
// # Cache
//
// The Cache provides in-memory caching with TTL support:
//
//	cache := core.NewCache(
//		core.WithDefaultTTL(5 * time.Minute),
//		core.WithMaxItems(1000),
//	)
//
//	// Simple set/get
//	cache.Set("key", value)
//	val, ok := cache.Get("key")
//
//	// Atomic get-or-set (computes only if missing/expired)
//	val, err := cache.GetOrSet("key", func() (any, error) {
//		return computeExpensiveValue()
//	})
//
// # Middleware
//
// Middleware wraps request execution with additional behavior. Chain multiple
// middlewares together:
//
//	chain := core.MiddlewareChain(
//		core.NewLoggingMiddleware(),
//		core.NewRetryMiddleware(core.WithMaxRetries(3)),
//	)
//
// Available middleware:
//
//   - LoggingMiddleware: Logs requests and response times
//   - RetryMiddleware: Retries failed requests with exponential backoff
//   - MetricsMiddleware: Tracks request counts and durations
//   - RateLimitingMiddleware: Limits requests per second
//   - CachingMiddleware: Provides cache key generation
//
// Example with retry and logging:
//
//	loggingMW := core.NewLoggingMiddleware(
//		core.WithLogger(myLogger),
//		core.WithLogTiming(true),
//	)
//
//	retryMW := core.NewRetryMiddleware(
//		core.WithMaxRetries(3),
//		core.WithRetryBackoff(&core.BackoffConfig{
//			BaseDelay:  time.Second,
//			MaxDelay:   30 * time.Second,
//			Multiplier: 2.0,
//		}),
//	)
//
// Example with metrics:
//
//	metrics, metricsMW := core.NewMetricsMiddleware()
//	// ... use middleware ...
//	fmt.Printf("Total requests: %d\n", metrics.TotalRequests())
//	fmt.Printf("Average duration: %v\n", metrics.AverageDuration())
package core
