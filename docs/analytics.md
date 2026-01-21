---
layout: default
title: Analytics API
description: Query YouTube Analytics data for channel statistics, demographics, and revenue.
---

## Overview

Query YouTube Analytics data for your channel:

**Reports:** Flexible analytics queries
- Custom date ranges
- Multiple metrics and dimensions
- Filtering and sorting
- Pagination support

**Convenience Methods:** Common queries
- Channel overview statistics
- Daily view breakdown
- Top videos by views
- Geographic and device breakdown
- Revenue reports (if monetized)

**Report Processing:** Easy data access
- Typed accessors for row data
- Aggregate totals
- Column header metadata

## NewClient

Create a new Analytics API client.

```go
client := analytics.NewClient(
    analytics.WithAccessToken("your-access-token"),
)
```

### With Token Provider

Integrate with AuthClient for automatic token refresh.

```go
client := analytics.NewClient(
    analytics.WithTokenProvider(authClient.AccessToken),
)
```

### Options

```go
client := analytics.NewClient(
    analytics.WithHTTPClient(customHTTPClient),
    analytics.WithAccessToken("static-token"),
    analytics.WithTokenProvider(tokenFunc),
)
```

## Query

Execute a custom analytics query.

```go
report, err := client.Query(ctx, &analytics.QueryParams{
    IDs:        "channel==MINE",
    StartDate:  "2025-01-01",
    EndDate:    "2025-01-31",
    Metrics:    "views,estimatedMinutesWatched,averageViewDuration",
    Dimensions: "day",
    Sort:       "day",
})
if err != nil {
    log.Fatal(err)
}
```

### Query Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `IDs` | Yes | Channel or content owner (`channel==MINE` or `channel==UC1234`) |
| `StartDate` | Yes | Start date in `YYYY-MM-DD` format |
| `EndDate` | Yes | End date in `YYYY-MM-DD` format |
| `Metrics` | Yes | Comma-separated metrics to retrieve |
| `Dimensions` | No | Comma-separated dimensions to group by |
| `Filters` | No | Semicolon-separated dimension filters |
| `Sort` | No | Comma-separated sort fields (prefix `-` for descending) |
| `MaxResults` | No | Maximum number of rows to return |
| `StartIndex` | No | 1-based index of first row |
| `Currency` | No | Currency code for revenue metrics (`USD`, `EUR`, etc.) |

### Filtering

Filter results by dimension values.

```go
report, err := client.Query(ctx, &analytics.QueryParams{
    IDs:       "channel==MINE",
    StartDate: "2025-01-01",
    EndDate:   "2025-01-31",
    Metrics:   "views,estimatedMinutesWatched",
    Filters:   "video==dQw4w9WgXcQ;country==US",
})
```

## Convenience Methods

### QueryChannelViews

Get overall channel statistics.

```go
report, err := client.QueryChannelViews(ctx, "2025-01-01", "2025-01-31")
// Returns: views, estimatedMinutesWatched, averageViewDuration,
//          subscribersGained, subscribersLost
```

### QueryDailyViews

Get daily breakdown of views.

```go
report, err := client.QueryDailyViews(ctx, "2025-01-01", "2025-01-31")
// Returns: day, views, estimatedMinutesWatched
```

### QueryTopVideos

Get top videos by view count.

```go
report, err := client.QueryTopVideos(ctx, "2025-01-01", "2025-01-31", 10)
// Returns: video, views, estimatedMinutesWatched, likes, comments
```

### QueryCountryBreakdown

Get views by country.

```go
report, err := client.QueryCountryBreakdown(ctx, "2025-01-01", "2025-01-31")
// Returns: country, views, estimatedMinutesWatched
```

### QueryDeviceBreakdown

Get views by device type.

```go
report, err := client.QueryDeviceBreakdown(ctx, "2025-01-01", "2025-01-31")
// Returns: deviceType, views, estimatedMinutesWatched
```

### QueryRevenueReport

Get revenue metrics (requires monetization).

```go
report, err := client.QueryRevenueReport(ctx, "2025-01-01", "2025-01-31")
// Returns: day, estimatedRevenue, estimatedAdRevenue, monetizedPlaybacks, cpm
```

## Working with Reports

### Report Structure

```go
type Report struct {
    Kind          string         // "youtubeAnalytics#resultTable"
    ColumnHeaders []ColumnHeader // Column metadata
    RawRows       [][]any        // Raw row data
}

type ColumnHeader struct {
    Name       string // Column name
    ColumnType string // "DIMENSION" or "METRIC"
    DataType   string // "STRING", "INTEGER", "FLOAT", etc.
}
```

### Iterating Rows

Use `Rows()` for typed access to report data.

```go
for _, row := range report.Rows() {
    day := row.GetString("day")
    views := row.GetInt("views")
    minutes := row.GetFloat("estimatedMinutesWatched")
    fmt.Printf("%s: %d views, %.1f minutes watched\n", day, views, minutes)
}
```

### Row Accessors

```go
// Get string value
day := row.GetString("day")

// Get integer value (converts from float64)
views := row.GetInt("views")

// Get float value
minutes := row.GetFloat("estimatedMinutesWatched")
```

### Aggregate Totals

```go
totalViews := report.TotalViews()
totalMinutes := report.TotalMinutesWatched()

fmt.Printf("Total: %d views, %.0f minutes\n", totalViews, totalMinutes)
```

## Common Metrics

| Metric | Description |
|--------|-------------|
| `views` | Number of video views |
| `estimatedMinutesWatched` | Total watch time in minutes |
| `averageViewDuration` | Average view duration in seconds |
| `subscribersGained` | New subscribers |
| `subscribersLost` | Lost subscribers |
| `likes` | Number of likes |
| `dislikes` | Number of dislikes |
| `comments` | Number of comments |
| `shares` | Number of shares |
| `estimatedRevenue` | Total estimated revenue |
| `estimatedAdRevenue` | Ad revenue |
| `cpm` | Cost per thousand impressions |
| `monetizedPlaybacks` | Monetized playback count |

## Common Dimensions

| Dimension | Description |
|-----------|-------------|
| `day` | Daily breakdown |
| `month` | Monthly breakdown |
| `video` | Per-video breakdown |
| `country` | Country code (US, GB, etc.) |
| `province` | Province/state |
| `city` | City |
| `deviceType` | MOBILE, DESKTOP, TV, GAME_CONSOLE, etc. |
| `operatingSystem` | Operating system |
| `ageGroup` | Age demographic |
| `gender` | Gender demographic |
| `trafficSourceType` | Traffic source |

## Error Handling

```go
var analyticsErr *analytics.AnalyticsError
if errors.As(err, &analyticsErr) {
    fmt.Printf("Status: %d, Code: %s\n", analyticsErr.StatusCode, analyticsErr.Code)

    if analyticsErr.IsPermissionDenied() {
        log.Println("Need proper OAuth scope or channel access")
    }
    if analyticsErr.IsInvalidArgument() {
        log.Println("Invalid query parameter")
    }
    if analyticsErr.IsQuotaExceeded() {
        log.Println("Rate limit exceeded")
    }
}
```

## Example: Channel Dashboard

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/Its-donkey/yougopher/youtube/analytics"
    "github.com/Its-donkey/yougopher/youtube/auth"
)

func main() {
    ctx := context.Background()

    // Setup authentication
    authClient := auth.NewAuthClient(auth.Config{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        RedirectURL:  "http://localhost:8080/callback",
        Scopes:       []string{auth.ScopeReadOnly, auth.ScopePartner},
    })

    // ... complete OAuth flow ...

    // Create analytics client
    client := analytics.NewClient(
        analytics.WithTokenProvider(authClient.AccessToken),
    )

    // Get date range (last 30 days)
    endDate := time.Now().Format("2006-01-02")
    startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

    // Channel overview
    overview, err := client.QueryChannelViews(ctx, startDate, endDate)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Total views: %d\n", overview.TotalViews())
    fmt.Printf("Total watch time: %.0f minutes\n", overview.TotalMinutesWatched())

    // Top videos
    topVideos, err := client.QueryTopVideos(ctx, startDate, endDate, 5)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("\nTop 5 Videos:")
    for i, row := range topVideos.Rows() {
        fmt.Printf("%d. Video %s: %d views\n", i+1, row.GetString("video"), row.GetInt("views"))
    }

    // Daily trend
    daily, err := client.QueryDailyViews(ctx, startDate, endDate)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("\nDaily Views:")
    for _, row := range daily.Rows() {
        fmt.Printf("  %s: %d views\n", row.GetString("day"), row.GetInt("views"))
    }
}
```

## Thread Safety

`Client` is safe for concurrent use.
