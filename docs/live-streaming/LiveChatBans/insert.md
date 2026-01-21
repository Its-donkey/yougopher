---
layout: default
title: BanUser / TimeoutUser
description: Ban or timeout a user in live chat
---

Ban or timeout a user from the live chat.

**Quota Cost:** 50 units

## BanUser (Permanent)

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

ban, err := poller.BanUser(ctx, "channel-id-to-ban")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("User banned: %s\n", ban.ID)
```

## TimeoutUser (Temporary)

```go
// Timeout for 5 minutes (300 seconds)
ban, err := poller.TimeoutUser(ctx, "channel-id", 300)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("User timed out for %d seconds\n", ban.Snippet.BanDurationSeconds)
```

## Via ChatBotClient

```go
bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
if err != nil {
    log.Fatal(err)
}

// Permanent ban
err = bot.Ban(ctx, "channel-id")

// Timeout for 5 minutes
err = bot.Timeout(ctx, "channel-id", 300)
```

## Common Timeout Durations

| Duration | Seconds | Use Case |
|----------|---------|----------|
| 1 minute | 60 | Minor warning |
| 5 minutes | 300 | Standard timeout |
| 10 minutes | 600 | Moderate offense |
| 1 hour | 3600 | Serious offense |
| 24 hours | 86400 | Severe offense |

## Auto-Moderation Example

Timeout users who spam:

```go
messageCount := make(map[string]int)
var mu sync.Mutex

poller.OnMessage(func(msg *streaming.LiveChatMessage) {
    channelID := msg.AuthorDetails.ChannelID

    mu.Lock()
    messageCount[channelID]++
    count := messageCount[channelID]
    mu.Unlock()

    // If more than 5 messages in quick succession, timeout
    if count > 5 {
        _, err := poller.TimeoutUser(ctx, channelID, 60)
        if err != nil {
            log.Printf("Failed to timeout: %v", err)
        }
        mu.Lock()
        messageCount[channelID] = 0
        mu.Unlock()
    }
})

// Reset counts periodically
go func() {
    for range time.Tick(10 * time.Second) {
        mu.Lock()
        messageCount = make(map[string]int)
        mu.Unlock()
    }
}()
```

## Ban Event Notification

When a user is banned, all listeners receive an event:

```go
poller.OnBan(func(details *streaming.UserBannedDetails) {
    fmt.Printf("User %s was %s\n",
        details.BannedUserDetails.DisplayName,
        details.BanType)
})
```

## Common Errors

| Error | Description |
|-------|-------------|
| `invalidBanDuration` | Invalid timeout duration |
| `cannotBanSelf` | Cannot ban yourself |
| `cannotBanOwner` | Cannot ban the broadcast owner |
| `ForbiddenError` | No permission to ban |
| `liveChatEnded` | Chat has ended |
