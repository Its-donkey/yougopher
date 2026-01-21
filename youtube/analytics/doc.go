// Package analytics provides access to the YouTube Analytics API.
//
// # Reports
//
// Query analytics data for your channel:
//
//	client := analytics.NewClient(
//		analytics.WithTokenProvider(authClient.AccessToken),
//	)
//
//	report, err := client.Query(ctx, &analytics.QueryParams{
//		IDs:        "channel==MINE",
//		StartDate:  "2025-01-01",
//		EndDate:    "2025-01-31",
//		Metrics:    "views,estimatedMinutesWatched,averageViewDuration",
//		Dimensions: "day",
//	})
//
// # Convenience Methods
//
// Common queries have helper methods:
//
//	// Channel overview
//	report, err := client.QueryChannelViews(ctx, "2025-01-01", "2025-01-31")
//
//	// Daily breakdown
//	report, err := client.QueryDailyViews(ctx, "2025-01-01", "2025-01-31")
//
//	// Top videos
//	report, err := client.QueryTopVideos(ctx, "2025-01-01", "2025-01-31", 10)
//
//	// Geographic breakdown
//	report, err := client.QueryCountryBreakdown(ctx, "2025-01-01", "2025-01-31")
//
//	// Device breakdown
//	report, err := client.QueryDeviceBreakdown(ctx, "2025-01-01", "2025-01-31")
//
// # Working with Reports
//
// Access report data using typed accessors:
//
//	for _, row := range report.Rows() {
//		day := row.GetString("day")
//		views := row.GetInt("views")
//		minutes := row.GetFloat("estimatedMinutesWatched")
//		fmt.Printf("%s: %d views, %.1f minutes\n", day, views, minutes)
//	}
//
//	// Aggregate totals
//	totalViews := report.TotalViews()
//	totalMinutes := report.TotalMinutesWatched()
//
// # Common Metrics
//
//   - views: Number of video views
//   - estimatedMinutesWatched: Total watch time in minutes
//   - averageViewDuration: Average view duration in seconds
//   - subscribersGained: New subscribers
//   - subscribersLost: Lost subscribers
//   - likes, dislikes, comments, shares: Engagement metrics
//   - estimatedRevenue: Total estimated revenue (if monetized)
//
// # Dimensions
//
// Group data by:
//
//   - day, month: Time-based breakdown
//   - video: Per-video breakdown
//   - country, province, city: Geographic breakdown
//   - deviceType: Device breakdown (MOBILE, DESKTOP, TV, etc.)
//   - operatingSystem: OS breakdown
//   - ageGroup, gender: Demographics (if available)
//   - trafficSourceType: Where views came from
//
// # Error Handling
//
// Handle analytics-specific errors:
//
//	var analyticsErr *analytics.AnalyticsError
//	if errors.As(err, &analyticsErr) {
//		if analyticsErr.IsPermissionDenied() {
//			// Need proper OAuth scope or channel access
//		}
//		if analyticsErr.IsQuotaExceeded() {
//			// Rate limit hit
//		}
//	}
package analytics
