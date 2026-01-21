---
layout: default
title: LiveChatBans.insert
description: Bans a user from a live chat (permanent or temporary timeout)
---

Bans a user from a live chat (permanent or temporary timeout).

## Request

### HTTP Request

```
POST https://www.googleapis.com/youtube/v3/liveChat/bans
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `part` | Yes | string | Must include `snippet`. |

### Authorization

Requires OAuth 2.0 authorization with the following scope:

- `https://www.googleapis.com/auth/youtube.force-ssl`

The authenticated user must be:
- The broadcast owner, OR
- A chat moderator

### Request Body

```json
{
  "snippet": {
    "liveChatId": "string",
    "type": "string",
    "bannedUserDetails": {
      "channelId": "string"
    },
    "banDurationSeconds": long
  }
}
```

### Required Fields

| Field | Description |
|-------|-------------|
| `snippet.liveChatId` | The live chat ID. |
| `snippet.type` | Ban type: `permanent` or `temporary`. |
| `snippet.bannedUserDetails.channelId` | The channel ID of the user to ban. |

### Optional Fields

| Field | Description |
|-------|-------------|
| `snippet.banDurationSeconds` | Duration for temporary bans (in seconds). Required when type is `temporary`. |

### Ban Types

| Type | Description |
|------|-------------|
| `permanent` | User is permanently banned from the chat. |
| `temporary` | User is temporarily banned (timeout). Must specify duration. |

## Response

If successful, this method returns a liveChatBan resource:

```json
{
  "kind": "youtube#liveChatBan",
  "etag": "string",
  "id": "string",
  "snippet": {
    "liveChatId": "string",
    "type": "string",
    "banDurationSeconds": long,
    "bannedUserDetails": {
      "channelId": "string",
      "channelUrl": "string",
      "displayName": "string",
      "profileImageUrl": "string"
    }
  }
}
```

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `liveChatIdRequired` | The liveChatId is required. |
| 400 | `channelIdRequired` | The banned user's channel ID is required. |
| 400 | `invalidBanDuration` | Ban duration is invalid (for temporary bans). |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | The user cannot ban in this chat. |
| 403 | `liveChatEnded` | The live chat has ended. |
| 403 | `cannotBanSelf` | Cannot ban yourself. |
| 403 | `cannotBanOwner` | Cannot ban the broadcast owner. |
| 404 | `liveChatNotFound` | The live chat does not exist. |

## Quota Cost

This method consumes **50 quota units**.

---

## Yougopher Implementation

### BanUser (Permanent)

Permanently ban a user from the live chat.

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

ban, err := poller.BanUser(ctx, "channel-id-to-ban")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("User banned: %s\n", ban.ID)
```

### TimeoutUser (Temporary)

Temporarily ban a user (timeout).

```go
// Timeout for 5 minutes (300 seconds)
ban, err := poller.TimeoutUser(ctx, "channel-id", 300)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("User timed out for %d seconds\n", ban.Snippet.BanDurationSeconds)
```

### Ban via ChatBotClient

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

### Ban Constants

```go
streaming.BanTypePermanent // "permanent"
streaming.BanTypeTemporary // "temporary"
```

### Common Timeout Durations

| Duration | Seconds | Use Case |
|----------|---------|----------|
| 1 minute | 60 | Minor warning |
| 5 minutes | 300 | Standard timeout |
| 10 minutes | 600 | Moderate offense |
| 1 hour | 3600 | Serious offense |
| 24 hours | 86400 | Severe offense |

### Auto-Moderation Example

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

### Ban Event Notification

When a user is banned, an event is sent to all listeners:

```go
poller.OnBan(func(details *streaming.UserBannedDetails) {
    fmt.Printf("User %s was %s\n",
        details.BannedUserDetails.DisplayName,
        details.BanType)
})
```
