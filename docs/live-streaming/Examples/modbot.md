---
layout: default
title: Moderation Bot
description: Auto-moderation bot with spam detection
---

An auto-moderation bot with spam detection and bad word filtering.

**Directory:** `examples/modbot/`

## Run

**Requirements:** Active YouTube live stream + moderator permissions

1. Complete the [setup steps](./index.md#setup) (clone, credentials, env vars)
2. Navigate to the example:
   ```bash
   cd examples/modbot
   ```
3. Run the bot:
   ```bash
   go run main.go
   ```
4. Open http://localhost:8080/login and complete OAuth
5. The bot will connect to your active broadcast and start moderating

## Features

- Bad word filtering with auto-delete and timeout
- Spam detection (messages per minute tracking)
- Role-based permissions (mods and owner bypass filters)
- Ban event logging

## Commands

| Command | Description |
|---------|-------------|
| `!ban @user` | Permanently ban a user |
| `!timeout @user` | Timeout a user (5 minutes) |
| `!unban @user` | Remove a user's ban |
| `!stats` | Show moderation statistics |

## Configuration

```go
type ModerationConfig struct {
    BadWords         []string
    SpamThreshold    int           // Messages per window
    SpamWindow       time.Duration // Time window for spam detection
    TimeoutDuration  time.Duration
    EnableAutoDelete bool
    EnableAutoTimeout bool
}

config := ModerationConfig{
    BadWords:         []string{"badword1", "badword2"},
    SpamThreshold:    5,
    SpamWindow:       time.Minute,
    TimeoutDuration:  5 * time.Minute,
    EnableAutoDelete: true,
    EnableAutoTimeout: true,
}
```

## Moderation Actions

```go
// Delete a message
err := streaming.DeleteMessage(ctx, client, messageID)

// Timeout a user (temporary ban)
ban, err := streaming.InsertBan(ctx, client, &streaming.LiveChatBan{
    Snippet: &streaming.BanSnippet{
        LiveChatID: liveChatID,
        Type:       "temporary",
        BanDurationSeconds: 300, // 5 minutes
        BannedUserDetails: &streaming.BannedUserDetails{
            ChannelID: userChannelID,
        },
    },
})

// Permanently ban a user
ban, err := streaming.InsertBan(ctx, client, &streaming.LiveChatBan{
    Snippet: &streaming.BanSnippet{
        LiveChatID: liveChatID,
        Type:       "permanent",
        BannedUserDetails: &streaming.BannedUserDetails{
            ChannelID: userChannelID,
        },
    },
})

// Unban a user
err := streaming.DeleteBan(ctx, client, banID)
```

## Spam Detection

```go
type SpamTracker struct {
    messages map[string][]time.Time // channelID -> message timestamps
    mu       sync.Mutex
}

func (t *SpamTracker) RecordMessage(channelID string) bool {
    t.mu.Lock()
    defer t.mu.Unlock()

    now := time.Now()
    cutoff := now.Add(-config.SpamWindow)

    // Filter old messages
    var recent []time.Time
    for _, ts := range t.messages[channelID] {
        if ts.After(cutoff) {
            recent = append(recent, ts)
        }
    }
    recent = append(recent, now)
    t.messages[channelID] = recent

    return len(recent) > config.SpamThreshold
}
```

## Role Checking

```go
func shouldModerate(msg *streaming.LiveChatMessage) bool {
    // Skip owner and moderators
    if msg.AuthorDetails.IsChatOwner || msg.AuthorDetails.IsChatModerator {
        return false
    }
    return true
}
```
