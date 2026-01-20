// Package analytics provides access to the YouTube Analytics API.
//
// # Reports
//
// Query analytics data:
//
//	report, err := analytics.Query(ctx, client, &analytics.QueryParams{
//		IDs:        "channel==MINE",
//		StartDate:  "2025-01-01",
//		EndDate:    "2025-01-31",
//		Metrics:    "views,estimatedMinutesWatched,averageViewDuration",
//		Dimensions: "day",
//	})
//
// # Common Metrics
//
//   - views: Number of video views
//   - estimatedMinutesWatched: Total watch time in minutes
//   - averageViewDuration: Average view duration in seconds
//   - subscribersGained: New subscribers
//   - subscribersLost: Lost subscribers
//
// # Dimensions
//
// Group data by:
//
//   - day: Daily breakdown
//   - video: Per-video breakdown
//   - country: Geographic breakdown
//   - deviceType: Device breakdown (MOBILE, DESKTOP, etc.)
package analytics
