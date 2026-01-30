# Yougopher

[![Go Reference](https://pkg.go.dev/badge/github.com/Its-donkey/yougopher.svg)](https://pkg.go.dev/github.com/Its-donkey/yougopher)
[![codecov](https://codecov.io/gh/Its-donkey/yougopher/graph/badge.svg)](https://codecov.io/gh/Its-donkey/yougopher)
[![Go Report Card](https://goreportcard.com/badge/github.com/Its-donkey/yougopher)](https://goreportcard.com/report/github.com/Its-donkey/yougopher)

A YouTube API toolkit in Go focused on live chat bot functionality.

## Installation

```bash
go get github.com/Its-donkey/yougopher
```

## Features

- **OAuth 2.0** - Authorization code flow, device flow, service accounts
- **Quota Tracking** - Built-in YouTube API quota management
- **Data API** - Videos, channels, playlists, search, comments, subscriptions
- **Live Streaming** - Live chat polling, moderation, broadcasts, stream management
- **Analytics** - Channel statistics, views, demographics, revenue
- **Caching** - In-memory cache with TTL support
- **Middleware** - Logging, retry, metrics, rate limiting

## Requirements

- Go 1.23 or later
- YouTube Data API v3 credentials

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

    select {} // Keep running
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

## Documentation

- [Core API](docs/core.md) - Client, errors, quota, cache, middleware
- [Authentication](docs/auth.md) - OAuth 2.0, device flow, service accounts
- [Streaming](docs/streaming.md) - Chat bot, polling, broadcasts
- [Data API](docs/data.md) - Videos, channels, playlists, search
- [Analytics API](docs/analytics.md) - Channel statistics, demographics, revenue
- [Test Results](docs/test-results.md) - Coverage and mutation testing

## Example Projects

See [docs/examples/](docs/examples/) for complete working applications:

- **[Chat Bot](docs/examples/chatbot/)** - Live chat bot with commands and message handling
- **[Moderation Bot](docs/examples/modbot/)** - Auto-moderation with spam detection and alerts
- **[Chat Monitor](docs/examples/monitor/)** - Read-only chat logging and statistics
- **[Analytics Dashboard](docs/examples/analytics/)** - CLI dashboard with channel stats

## License

MIT License - see [LICENSE](LICENSE) for details.
