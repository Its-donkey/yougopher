# Yougopher

A YouTube API toolkit in Go focused on live chat bot functionality.

## Installation

```bash
go get github.com/Its-donkey/yougopher
```

## Features

- **Live Chat Bot** - Monitor and interact with YouTube live chat
- **Chat Monitoring** - Read-only logging, moderation alerts
- **OAuth 2.0** - Authorization code flow, device flow, service accounts
- **Quota Tracking** - Built-in YouTube API quota management
- **Data API** - Videos, channels, playlists, search, comments, subscriptions
- **Live Streaming** - Broadcasts, live chat ID retrieval
- **Analytics** - Channel statistics, views, demographics, revenue
- **Caching** - In-memory cache with TTL support
- **Middleware** - Logging, retry, metrics, rate limiting

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/Its-donkey/yougopher/youtube/auth"
    "github.com/Its-donkey/yougopher/youtube/core"
    "github.com/Its-donkey/yougopher/youtube/streaming"
)

func main() {
    ctx := context.Background()

    // Initialize clients
    client := core.NewClient()
    authClient := auth.NewAuthClient(auth.Config{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        RedirectURL:  "http://localhost:8080/callback",
    })

    // Create chat bot
    bot, err := streaming.NewChatBotClient(client, authClient, "live-chat-id")
    if err != nil {
        log.Fatal(err)
    }

    // Register message handler
    bot.OnMessage(func(msg *streaming.ChatMessage) {
        log.Printf("[%s] %s", msg.Author.DisplayName, msg.Message)
    })

    // Connect and start listening
    if err := bot.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer bot.Close()

    // Keep running...
    select {}
}
```

## Packages

| Package | Description |
|---------|-------------|
| `youtube/core` | HTTP client, errors, quota tracking, cache, middleware |
| `youtube/auth` | OAuth 2.0 flows, device flow, service accounts, token management |
| `youtube/streaming` | Live chat bot, polling, moderation, broadcasts |
| `youtube/data` | Videos, channels, playlists, search, comments, subscriptions |
| `youtube/analytics` | YouTube Analytics API (channel stats, demographics, revenue) |

## Data API Examples

```go
// Get video information
video, err := data.GetVideo(ctx, client, "video-id")
if video.IsLive() {
    liveChatID := video.LiveStreamingDetails.ActiveLiveChatID
}

// Search for live streams (costs 100 quota units!)
results, err := data.SearchLiveStreams(ctx, client, "gaming", 10)

// Get channel subscriptions
subs, err := data.GetMySubscriptions(ctx, client, 50)

// Get video comments
comments, err := data.GetVideoComments(ctx, client, "video-id", 20)

// Get playlist items
items, err := data.GetPlaylistItems(ctx, client, &data.GetPlaylistItemsParams{
    PlaylistID: "playlist-id",
    MaxResults: 50,
})
```

## Core Examples

```go
// In-memory cache with TTL
cache := core.NewCache(core.WithDefaultTTL(5 * time.Minute))
cache.Set("key", value)
val, ok := cache.Get("key")

// Atomic get-or-set
video, err := cache.GetOrSet("video:abc", func() (any, error) {
    return data.GetVideo(ctx, client, "abc")
})

// Middleware chain
metrics, metricsMW := core.NewMetricsMiddleware()
retryMW := core.NewRetryMiddleware(core.WithMaxRetries(3))
loggingMW := core.NewLoggingMiddleware()

// Check metrics
fmt.Printf("Requests: %d, Avg: %v\n",
    metrics.TotalRequests(), metrics.AverageDuration())
```

## Analytics Examples

```go
// Create analytics client
client := analytics.NewClient(
    analytics.WithTokenProvider(authClient.AccessToken),
)

// Channel statistics for last 30 days
report, err := client.QueryChannelViews(ctx, "2025-01-01", "2025-01-31")
fmt.Printf("Total views: %d\n", report.TotalViews())

// Top videos by views
topVideos, err := client.QueryTopVideos(ctx, "2025-01-01", "2025-01-31", 10)
for _, row := range topVideos.Rows() {
    fmt.Printf("Video %s: %d views\n", row.GetString("video"), row.GetInt("views"))
}

// Daily breakdown
daily, err := client.QueryDailyViews(ctx, "2025-01-01", "2025-01-31")

// Geographic breakdown
countries, err := client.QueryCountryBreakdown(ctx, "2025-01-01", "2025-01-31")
```

## Documentation

- [Core API](docs/core.md) - Client, errors, quota, cache, middleware
- [Authentication](docs/auth.md) - OAuth 2.0, device flow, service accounts
- [Streaming](docs/streaming.md) - Chat bot, polling, broadcasts
- [Data API](docs/data.md) - Videos, channels, playlists, search
- [Analytics API](docs/analytics.md) - Channel statistics, demographics, revenue

## Examples

See [docs/examples/](docs/examples/) for complete working applications:

- **[Chat Bot](docs/examples/chatbot/)** - Basic live chat bot with commands
- **[Moderation Bot](docs/examples/modbot/)** - Auto-moderation with spam detection
- **[Analytics Dashboard](docs/examples/analytics/)** - CLI dashboard with channel stats

## Requirements

- Go 1.23 or later
- YouTube Data API v3 credentials

## License

MIT License - see [LICENSE](LICENSE) for details.
