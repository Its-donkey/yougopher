---
layout: default
title: API Reference
description: Complete API documentation for all Yougopher packages
---

## Packages

Yougopher is organized into several packages, each handling a specific area of the YouTube API.

### Core

| Package | Description |
|---------|-------------|
| [core](core) | HTTP client, caching, middleware, error handling |
| [auth](auth) | OAuth 2.0, device flow, service accounts |

### Data API

| Package | Description |
|---------|-------------|
| [data](data) | Videos, channels, playlists, comments, subscriptions, search |
| [analytics](analytics) | Channel analytics and reporting |

### Live Streaming

| Package | Description |
|---------|-------------|
| [streaming](streaming) | Broadcasts, streams, live chat |
| [Live Streaming API Reference](live-streaming/) | Detailed method documentation |

---

## Package: youtube/core

The core package provides the HTTP client and foundational utilities.

### Client

```go
// Create a new client
client := core.NewClient(ctx, httpClient, opts...)

// With caching
client := core.NewClient(ctx, httpClient,
    core.WithCache(cache, 5*time.Minute),
)

// With middleware
client := core.NewClient(ctx, httpClient,
    core.WithMiddleware(loggingMiddleware),
)
```

### Error Handling

```go
resp, err := data.GetVideos(ctx, client, params)
if err != nil {
    var apiErr *core.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API Error %d: %s\n", apiErr.Code, apiErr.Message)
    }
}
```

[Full core documentation →](core)

---

## Package: youtube/auth

Authentication methods for the YouTube API.

### OAuth 2.0 Scopes

| Scope | Description |
|-------|-------------|
| `ScopeYouTube` | Full access to manage your YouTube account |
| `ScopeYouTubeReadOnly` | View your YouTube account |
| `ScopeYouTubeForceSsl` | View and manage YouTube resources (SSL required) |
| `ScopeYouTubeUpload` | Upload videos |

### Methods

| Function | Description |
|----------|-------------|
| `NewOAuth2Config` | Create OAuth2 configuration |
| `StartDeviceFlow` | Begin device authorization flow |
| `PollDeviceToken` | Poll for device flow token |
| `LoadToken` | Load token from file |
| `SaveToken` | Save token to file |

[Full auth documentation →](auth)

---

## Package: youtube/data

Data API for videos, channels, playlists, and more.

### Videos

```go
// Get videos by ID
resp, err := data.GetVideos(ctx, client, &data.GetVideosParams{
    IDs:  []string{"dQw4w9WgXcQ"},
    Part: []string{"snippet", "statistics"},
})

// Search videos
resp, err := data.Search(ctx, client, &data.SearchParams{
    Query:      "golang tutorial",
    Type:       "video",
    MaxResults: 25,
})
```

### Channels

```go
// Get your channel
resp, err := data.GetChannels(ctx, client, &data.GetChannelsParams{
    Mine: true,
    Part: []string{"snippet", "statistics"},
})

// Get channel by ID
resp, err := data.GetChannels(ctx, client, &data.GetChannelsParams{
    IDs:  []string{"UC_x5XG1OV2P6uZZ5FSM9Ttw"},
    Part: []string{"snippet"},
})
```

[Full data documentation →](data)

---

## Package: youtube/streaming

Live streaming, broadcasts, and chat.

### Broadcasts

```go
// List active broadcasts
resp, err := streaming.GetBroadcasts(ctx, client, &streaming.GetBroadcastsParams{
    BroadcastStatus: "active",
    Part:            []string{"snippet", "status"},
})
```

### Live Chat

```go
// Poll for messages
poller := streaming.NewLiveChatPoller(client, liveChatID)
poller.OnMessage(func(msg *streaming.LiveChatMessage) {
    fmt.Printf("%s: %s\n", msg.AuthorDetails.DisplayName, msg.Snippet.DisplayMessage)
})
poller.Start(ctx)

// SSE streaming
stream, err := streaming.StreamLiveChatMessages(ctx, client, &streaming.StreamParams{
    LiveChatID: liveChatID,
})
for msg := range stream.Messages() {
    fmt.Println(msg.Snippet.DisplayMessage)
}
```

### Send Messages

```go
err := streaming.InsertLiveChatMessage(ctx, client, &streaming.InsertMessageParams{
    LiveChatID: liveChatID,
    Message:    "Hello from Yougopher!",
})
```

[Full streaming documentation →](streaming)

---

## Live Streaming API Reference

Detailed documentation for each YouTube Live Streaming API resource and method.

| Resource | Description |
|----------|-------------|
| [liveBroadcasts](live-streaming/liveBroadcasts/) | Live broadcast management |
| [liveStreams](live-streaming/liveStreams/) | Stream ingestion settings |
| [liveChatMessages](live-streaming/liveChatMessages/) | Chat message operations |
| [liveChatBans](live-streaming/liveChatBans/) | Chat moderation bans |
| [liveChatModerators](live-streaming/liveChatModerators/) | Moderator management |
| [superChatEvents](live-streaming/superChatEvents/) | Super Chat history |

[Full Live Streaming API Reference →](live-streaming/)
