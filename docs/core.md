---
layout: default
title: Core API
description: HTTP client, quota tracking, caching, and middleware for YouTube API.
---

## Overview

Foundation components for YouTube API interactions:

**Client:** HTTP client with authentication
- OAuth access tokens or API keys
- Automatic request formatting
- Response parsing and error handling

**Quota Tracking:** Monitor API usage
- Track usage by operation type
- Set limits and receive alerts
- Automatic daily reset (Pacific midnight)

**Cache:** In-memory caching with TTL
- Configurable TTL and max items
- Atomic get-or-set operations
- Automatic expired entry eviction

**Middleware:** Request/response pipeline
- Logging with timing
- Retry with exponential backoff
- Metrics collection
- Rate limiting

**Errors:** Typed error responses
- Quota exceeded detection
- Rate limit handling
- Auth error identification

## Client

### NewClient

Create a new YouTube API client.

```go
client := core.NewClient()
```

### Options

```go
client := core.NewClient(
    core.WithAccessToken("your-access-token"),
    core.WithAPIKey("your-api-key"),
    core.WithHTTPClient(customHTTPClient),
    core.WithBaseURL("https://custom-url"),
    core.WithUserAgent("MyApp/1.0"),
    core.WithQuotaTracker(quotaTracker),
)
```

### SetAccessToken

Update the access token (for token refresh).

```go
client.SetAccessToken(newToken)
```

### Do

Execute a raw API request.

```go
var result MyResponseType
err := client.Do(ctx, &core.Request{
    Method:    "GET",
    Path:      "videos",
    Query:     url.Values{"id": {"VIDEO_ID"}, "part": {"snippet"}},
    Operation: "videos.list", // For quota tracking
}, &result)
```

### Convenience Methods

```go
// GET request
err := client.Get(ctx, "videos", query, "videos.list", &result)

// POST request
err := client.Post(ctx, "videos", query, body, "videos.insert", &result)

// DELETE request
err := client.Delete(ctx, "videos", query, "videos.delete")
```

## Quota Tracking

YouTube API has a daily quota of 10,000 units by default. Different operations cost different amounts.

### Quota Costs

```go
var QuotaCosts = map[string]int{
    // Live Chat
    "liveChatMessages.list":     5,
    "liveChatMessages.insert":   50,
    "liveChatMessages.delete":   50,
    "liveChatBans.insert":       50,
    "liveChatBans.delete":       50,
    "liveChatModerators.insert": 50,
    "liveChatModerators.delete": 50,

    // Data API - Read
    "videos.list":        1,
    "channels.list":      1,
    "playlists.list":     1,
    "search.list":        100, // Expensive!

    // Data API - Write
    "videos.insert":        1600, // Video upload
    "videos.update":        50,
    "playlists.insert":     50,
    "comments.insert":      50,
}
```

### NewQuotaTracker

Create a quota tracker with a custom limit.

```go
tracker := core.NewQuotaTracker(10000) // Default daily quota
```

### Usage

```go
// Add quota manually
tracker.Add("videos.list", 1)

// Get current usage
used, limit := tracker.Used(), tracker.Limit()
fmt.Printf("Quota: %d/%d\n", used, limit)

// Check remaining
remaining := tracker.Remaining()

// Reset time
resetAt := tracker.ResetAt()
```

### OnQuotaUpdate

Register a callback for quota changes.

```go
unsub := tracker.OnQuotaUpdate(func(used, limit int) {
    pct := float64(used) / float64(limit) * 100
    if pct > 80 {
        log.Printf("Warning: Quota at %.1f%%", pct)
    }
})
// Later: unsub() to unregister
```

### Integration with Client

```go
tracker := core.NewQuotaTracker(10000)
client := core.NewClient(core.WithQuotaTracker(tracker))

// Quota is tracked automatically for all requests
err := client.Get(ctx, "videos", query, "videos.list", &result)

// Check usage
fmt.Printf("Used: %d/%d\n", tracker.Used(), tracker.Limit())
```

## Error Types

### APIError

General API error with status code and message.

```go
var apiErr *core.APIError
if errors.As(err, &apiErr) {
    fmt.Printf("Status: %d, Code: %s\n", apiErr.StatusCode, apiErr.Code)

    if apiErr.IsQuotaExceeded() {
        log.Println("Daily quota exceeded!")
    }
    if apiErr.IsRateLimited() {
        log.Println("Rate limited, slow down!")
    }
    if apiErr.IsNotFound() {
        log.Println("Resource not found")
    }
    if apiErr.IsForbidden() {
        log.Println("Access forbidden")
    }
}
```

### QuotaError

Indicates daily quota has been exceeded.

```go
var quotaErr *core.QuotaError
if errors.As(err, &quotaErr) {
    fmt.Printf("Quota: %d/%d\n", quotaErr.Used, quotaErr.Limit)
    fmt.Printf("Resets at: %s\n", quotaErr.ResetAt)
}
```

### RateLimitError

Indicates rate limit (requests per second) exceeded.

```go
var rateErr *core.RateLimitError
if errors.As(err, &rateErr) {
    fmt.Printf("Rate limited, retry after: %s\n", rateErr.RetryAfter)
    time.Sleep(rateErr.RetryAfter)
}
```

### AuthError

Indicates authentication failure.

```go
var authErr *core.AuthError
if errors.As(err, &authErr) {
    fmt.Printf("Auth failed: %s - %s\n", authErr.Code, authErr.Message)

    if authErr.IsExpired() {
        // Refresh token and retry
    }
    if authErr.IsRevoked() {
        // Re-authenticate user
    }
}
```

## Backoff Configuration

Configure exponential backoff for retries.

```go
backoff := &core.BackoffConfig{
    InitialDelay: 1 * time.Second,
    MaxDelay:     30 * time.Second,
    Multiplier:   2.0,
    Jitter:       0.1, // 10% random jitter
}

// Calculate delay for attempt N
delay := backoff.Delay(attemptNumber)
```

## Sentinel Errors

```go
var (
    core.ErrAlreadyRunning // Operation already in progress
    core.ErrNotRunning     // Operation not running
)
```

## Cache

In-memory caching with TTL (time-to-live) support.

### NewCache

Create a cache with optional configuration.

```go
cache := core.NewCache(
    core.WithDefaultTTL(5 * time.Minute),
    core.WithMaxItems(1000),
)
```

### Basic Operations

```go
// Set with default TTL
cache.Set("video:abc123", videoData)

// Set with custom TTL
cache.SetWithTTL("search:query", results, 1*time.Hour)

// Get
val, ok := cache.Get("video:abc123")
if ok {
    video := val.(*Video)
}

// Delete
cache.Delete("video:abc123")

// Clear all
cache.Clear()
```

### GetOrSet

Atomically get a value or compute it if missing/expired.

```go
val, err := cache.GetOrSet("expensive:key", func() (any, error) {
    // Only called if key is missing or expired
    return computeExpensiveValue()
})
```

### Cache Statistics

```go
stats := cache.Stats()
fmt.Printf("Total: %d, Active: %d, Expired: %d\n",
    stats.Items, stats.Active, stats.Expired)
```

### Cleanup

Remove expired entries manually (also happens automatically on insert).

```go
removed := cache.Cleanup()
fmt.Printf("Removed %d expired entries\n", removed)
```

## Middleware

Middleware wraps request execution with additional behavior.

### MiddlewareChain

Combine multiple middlewares.

```go
chain := core.MiddlewareChain(
    core.NewLoggingMiddleware(),
    core.NewRetryMiddleware(),
    metricsMW,
)
```

### LoggingMiddleware

Log requests and response times.

```go
loggingMW := core.NewLoggingMiddleware(
    core.WithLogger(myLogger),    // Custom logger
    core.WithLogTiming(true),     // Log durations
    core.WithLogBody(false),      // Don't log request bodies
)
```

### RetryMiddleware

Retry failed requests with exponential backoff.

```go
retryMW := core.NewRetryMiddleware(
    core.WithMaxRetries(3),
    core.WithRetryBackoff(&core.BackoffConfig{
        BaseDelay:  1 * time.Second,
        MaxDelay:   30 * time.Second,
        Multiplier: 2.0,
        Jitter:     0.1,
    }),
    core.WithShouldRetry(func(err error) bool {
        // Custom retry logic
        var rle *core.RateLimitError
        return errors.As(err, &rle)
    }),
)
```

### MetricsMiddleware

Track request counts and durations.

```go
metrics, metricsMW := core.NewMetricsMiddleware()

// Use middleware...

// Check stats
fmt.Printf("Total requests: %d\n", metrics.TotalRequests())
fmt.Printf("Failed requests: %d\n", metrics.FailedRequests())
fmt.Printf("Successful: %d\n", metrics.SuccessfulRequests())
fmt.Printf("Average duration: %v\n", metrics.AverageDuration())

// Reset stats
metrics.Reset()
```

### RateLimitingMiddleware

Limit outgoing request rate.

```go
rateLimitMW := core.NewRateLimitingMiddleware(
    core.WithRequestsPerSecond(10), // Max 10 requests/second
)
```

### Custom Middleware

Create your own middleware:

```go
func MyMiddleware() core.Middleware {
    return func(ctx context.Context, req *core.Request, next func(context.Context, *core.Request) error) error {
        // Before request
        log.Printf("Starting %s %s", req.Method, req.Path)

        // Execute request
        err := next(ctx, req)

        // After request
        if err != nil {
            log.Printf("Request failed: %v", err)
        }

        return err
    }
}
```

## Thread Safety

All types in the core package are safe for concurrent use.
