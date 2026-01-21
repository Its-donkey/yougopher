---
layout: default
title: Home
description: A YouTube API toolkit in Go focused on live chat bot functionality
---

<div style="text-align: center; margin-bottom: 2rem;">
  <img src="{{ '/assets/images/logo.png' | relative_url }}" alt="Yougopher" style="width: 180px; height: 180px;">
  <h1 style="margin-top: 1rem;">Yougopher</h1>
  <p style="font-size: 1.125rem; color: #53535F;">A comprehensive Go wrapper for the YouTube Data API with full endpoint coverage, multiple authentication flows, and real-time streaming support.</p>
</div>

## What is Yougopher?

Yougopher is a Go library for interacting with the YouTube Data API v3 and YouTube Live Streaming API. It provides a clean, idiomatic Go interface for building YouTube integrations, with a focus on live chat and streaming functionality.

## Getting Started

1. [Quick Start Guide](quickstart) - Set up authentication and make your first API call
2. [API Reference](api-reference) - Complete API documentation
3. [Cookbook](cookbook) - Common recipes and patterns
   
## Features

- **Authentication** - OAuth 2.0, device flow, and service account support
- **Data API** - Videos, channels, playlists, comments, subscriptions, search
- **Live Streaming** - Broadcasts, streams, live chat, Super Chats
- **Real-time Chat** - Polling and SSE streaming for live chat messages
- **Built-in Caching** - Configurable response caching with TTL
- **Rate Limiting** - Automatic quota management and retry logic

## Packages

| Package | Description |
|---------|-------------|
| `youtube/auth` | OAuth 2.0 authentication flows |
| `youtube/core` | HTTP client, caching, middleware |
| `youtube/data` | Data API (videos, channels, etc.) |
| `youtube/analytics` | Analytics and reporting API |
| `youtube/streaming` | Live streaming and chat API |

## Requirements

- Go 1.21 or later
- YouTube Data API v3 credentials
- OAuth 2.0 client ID (for user authentication)

## Installation

```bash
go get github.com/Its-donkey/yougopher
```

## Quick Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/yougopher/youtube/auth"
    "github.com/Its-donkey/yougopher/youtube/core"
    "github.com/Its-donkey/yougopher/youtube/data"
)

func main() {
    ctx := context.Background()

    // Create OAuth2 config
    config := auth.NewOAuth2Config(
        "YOUR_CLIENT_ID",
        "YOUR_CLIENT_SECRET",
        "http://localhost:8080/callback",
        auth.ScopeYouTubeReadOnly,
    )

    // Get token (implement your own flow)
    token, err := auth.LoadToken("token.json")
    if err != nil {
        log.Fatal(err)
    }

    // Create authenticated client
    client := core.NewClient(ctx, config.Client(ctx, token))

    // Get channel info
    resp, err := data.GetChannels(ctx, client, &data.GetChannelsParams{
        Mine: true,
        Part: []string{"snippet", "statistics"},
    })
    if err != nil {
        log.Fatal(err)
    }

    channel := resp.Items[0]
    fmt.Printf("Channel: %s\n", channel.Snippet.Title)
    fmt.Printf("Subscribers: %d\n", channel.Statistics.SubscriberCount)
}
```

## License

Yougopher is released under the MIT License.
