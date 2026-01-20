# Yougopher

A YouTube API toolkit in Go focused on live chat bot functionality.

## Installation

```bash
go get github.com/Its-donkey/yougopher
```

## Features

- **Live Chat Bot** - Monitor and interact with YouTube live chat
- **Chat Monitoring** - Read-only logging, moderation alerts
- **OAuth 2.0** - Full authentication flow support
- **Quota Tracking** - Built-in YouTube API quota management

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
    bot := streaming.NewChatBotClient(client, authClient, "live-chat-id")

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
| `youtube/core` | Base HTTP client, error types, quota tracking |
| `youtube/auth` | OAuth 2.0 flows, token management |
| `youtube/streaming` | Live chat bot, polling, moderation |
| `youtube/data` | Videos, channels, playlists, search |
| `youtube/analytics` | YouTube Analytics API |

## Requirements

- Go 1.23 or later
- YouTube Data API v3 credentials

## License

MIT License - see [LICENSE](LICENSE) for details.
