---
layout: default
title: LiveChatPoller
description: Poll for live chat messages
---

Poll for live chat messages with automatic handling.

**Quota Cost:** 5 units per poll

## LiveChatPoller

```go
poller := streaming.NewLiveChatPoller(client, liveChatID,
    streaming.WithMinPollInterval(2*time.Second),
    streaming.WithProfileImageSize(streaming.ProfileImageMedium),
)

// Register handlers
poller.OnMessage(func(msg *streaming.LiveChatMessage) {
    fmt.Printf("[%s] %s\n",
        msg.AuthorDetails.DisplayName,
        msg.Snippet.DisplayMessage)
})

// Start polling
err := poller.Start(ctx)
if err != nil {
    log.Fatal(err)
}
defer poller.Stop()
```

## Options

| Option | Description |
|--------|-------------|
| `WithMinPollInterval(d)` | Minimum time between polls |
| `WithMaxPollInterval(d)` | Maximum time between polls |
| `WithProfileImageSize(size)` | Image size: `ProfileImageDefault`, `ProfileImageMedium`, `ProfileImageHigh` |
| `WithBackoff(cfg)` | Custom backoff configuration |

## Event Handlers

```go
// Handle regular messages
poller.OnMessage(func(msg *streaming.LiveChatMessage) {
    switch msg.Type() {
    case streaming.MessageTypeText:
        fmt.Printf("Text: %s\n", msg.Message())
    case streaming.MessageTypeSuperChat:
        fmt.Printf("Super Chat: %s\n", msg.Snippet.SuperChatDetails.AmountDisplayString)
    case streaming.MessageTypeMembership:
        fmt.Println("New member!")
    }
})

// Handle deletions
poller.OnDelete(func(messageID string) {
    fmt.Printf("Message deleted: %s\n", messageID)
})

// Handle bans
poller.OnBan(func(details *streaming.UserBannedDetails) {
    fmt.Printf("User banned: %s\n", details.BannedUserDetails.DisplayName)
})

// Handle errors
poller.OnError(func(err error) {
    log.Printf("Error: %v", err)
})

// Handle connection events
poller.OnConnect(func() { log.Println("Connected") })
poller.OnDisconnect(func() { log.Println("Disconnected") })

// Handle poll completion
poller.OnPollComplete(func(count int, nextPoll time.Duration) {
    fmt.Printf("Received %d messages, next poll in %v\n", count, nextPoll)
})
```

## Configuration

```go
poller := streaming.NewLiveChatPoller(client, liveChatID,
    streaming.WithMinPollInterval(1*time.Second),
    streaming.WithMaxPollInterval(30*time.Second),
    streaming.WithProfileImageSize(streaming.ProfileImageHigh),
    streaming.WithBackoff(core.NewBackoffConfig(
        core.WithBaseDelay(1*time.Second),
        core.WithMaxDelay(30*time.Second),
    )),
)
```

## Page Token Management

```go
// Save token before stopping
token := poller.PageToken()

// Resume from that position later
poller.SetPageToken(token)
poller.Start(ctx)

// Clear token to start fresh
poller.ResetPageToken()
```

## Poll Interval

The API returns `pollingIntervalMillis` indicating how long to wait. The poller respects this within your configured bounds:

```go
interval := poller.PollInterval()
```

## Alternative

For lower latency, use [LiveChatStream](stream-list) (SSE) instead of polling.

## Common Errors

| Error | Description |
|-------|-------------|
| `liveChatNotFound` | Chat doesn't exist |
| `liveChatEnded` | Chat has ended |
| `liveChatDisabled` | Chat is disabled |
