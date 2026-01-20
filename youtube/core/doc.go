// Package core provides the base HTTP client, error types, and quota tracking
// for interacting with YouTube APIs.
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
//
// # Quota Tracking
//
// YouTube API uses a quota system where different operations cost different
// amounts. The QuotaTracker helps monitor usage:
//
//	tracker := core.NewQuotaTracker(10000) // 10,000 daily limit
//	tracker.Add("liveChatMessages.list", 5)
//	remaining := tracker.Remaining()
package core
