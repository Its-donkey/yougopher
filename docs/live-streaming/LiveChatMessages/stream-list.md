---
layout: default
title: LiveChatStream
description: SSE streaming for real-time chat messages
---

Establishes a Server-Sent Events (SSE) connection for real-time chat messages with lower latency than polling.

**Quota Cost:** 5 units per connection

## LiveChatStream

```go
stream := streaming.NewLiveChatStream(client, liveChatID,
    streaming.WithStreamAccessToken(accessToken),
    streaming.WithStreamParts("id", "snippet", "authorDetails"),
    streaming.WithStreamMaxResults(1000),
)
```

## Options

| Option | Description |
|--------|-------------|
| `WithStreamAccessToken(token)` | OAuth access token |
| `WithStreamParts(parts...)` | Resource parts to include |
| `WithStreamHL(lang)` | Localized currency language |
| `WithStreamMaxResults(n)` | Messages per response (200-2000) |
| `WithStreamProfileImageSize(px)` | Profile image size (16-720) |
| `WithStreamReconnectDelay(d)` | Initial reconnect delay |
| `WithStreamMaxReconnectDelay(d)` | Maximum reconnect delay |
| `WithStreamBackoff(cfg)` | Custom backoff configuration |
| `WithStreamHTTPClient(hc)` | Custom HTTP client |

## Lifecycle

```go
// Start the stream
err := stream.Start(ctx)

// Check if running
if stream.IsRunning() { ... }

// Stop the stream
stream.Stop()
```

## Event Handlers

All handlers return an unsubscribe function:

```go
// Handle messages
unsub := stream.OnMessage(func(msg *streaming.LiveChatMessage) {
    fmt.Printf("[%s] %s\n", msg.AuthorDetails.DisplayName, msg.Message())
})

// Handle full responses (includes metadata)
stream.OnResponse(func(resp *streaming.LiveChatMessageListResponse) {
    if resp.IsChatEnded() {
        fmt.Println("Chat ended")
    }
})

// Handle deletions
stream.OnDelete(func(messageID string) { ... })

// Handle bans
stream.OnBan(func(details *streaming.UserBannedDetails) { ... })

// Handle connection events
stream.OnConnect(func() { ... })
stream.OnDisconnect(func() { ... })

// Handle errors
stream.OnError(func(err error) { ... })

// Unsubscribe when done
unsub()
```

## Stream Resumption

```go
// Save token before stopping
token := stream.PageToken()

// Resume from saved position
stream.SetPageToken(token)
stream.Start(ctx)

// Clear token to start fresh
stream.ResetPageToken()
```

## Error Handling

```go
stream.OnError(func(err error) {
    var apiErr *core.APIError
    if errors.As(err, &apiErr) {
        switch {
        case apiErr.IsChatEnded():
            log.Println("Chat ended")
        case apiErr.IsChatDisabled():
            log.Println("Chat disabled")
        case apiErr.IsQuotaExceeded():
            log.Println("Quota exceeded")
        }
    }
})
```

## Automatic Reconnection

The stream automatically reconnects with exponential backoff:

- Initial delay: 1 second
- Multiplier: 2x per attempt
- Maximum delay: 30 seconds
- Jitter: ±20%

Configure with `WithStreamBackoff`:

```go
stream := streaming.NewLiveChatStream(client, liveChatID,
    streaming.WithStreamBackoff(core.NewBackoffConfig(
        core.WithBaseDelay(500*time.Millisecond),
        core.WithMaxDelay(60*time.Second),
        core.WithMultiplier(1.5),
        core.WithJitter(0.1),
    )),
)
```

## SSE vs Polling

| Feature | SSE (this) | Polling |
|---------|------------|---------|
| Latency | ~instant | Poll interval |
| Quota | 5/connection | 5/poll |
| Best for | Real-time bots | Custom logic |
