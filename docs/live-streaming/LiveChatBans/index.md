---
layout: default
title: LiveChatBans
description: Ban and timeout users in YouTube live chat
---

A `LiveChatBan` represents a ban that prevents a user from posting messages in a live chat. Bans can be permanent or temporary (timeout).

## Functions

| Function | Quota | Description |
|----------|-------|-------------|
| [BanUser](insert) | 50 | Permanently ban a user |
| [TimeoutUser](insert) | 50 | Temporarily ban a user |
| [UnbanUser](delete) | 50 | Remove a ban |

## Type Definition

```go
type LiveChatBan struct {
    Kind    string      `json:"kind,omitempty"`
    ETag    string      `json:"etag,omitempty"`
    ID      string      `json:"id,omitempty"`
    Snippet *BanSnippet `json:"snippet,omitempty"`
}
```

## Ban Type Constants

```go
streaming.BanTypePermanent // Permanent ban
streaming.BanTypeTemporary // Temporary ban (timeout)
```

## Common Timeout Durations

| Duration | Seconds | Use Case |
|----------|---------|----------|
| 1 minute | 60 | Minor warning |
| 5 minutes | 300 | Standard timeout |
| 10 minutes | 600 | Moderate offense |
| 1 hour | 3600 | Serious offense |
| 24 hours | 86400 | Severe offense |

## Who Can Ban

| User | Can Ban |
|------|---------|
| Broadcast owner | Yes |
| Moderator | Yes |
| Regular viewer | No |

## Quick Example

```go
import "github.com/Its-donkey/yougopher/youtube/streaming"

poller := streaming.NewLiveChatPoller(client, liveChatID)

// Permanent ban
ban, err := poller.BanUser(ctx, "channel-id-to-ban")

// Timeout for 5 minutes
ban, err := poller.TimeoutUser(ctx, "channel-id", 300)

// Unban
err := poller.UnbanUser(ctx, ban.ID)
```
