---
layout: default
title: LiveChatMessages.transition
description: Changes the chat mode for a live chat
---

Changes the chat mode for a live chat (subscribers only, members only, slow mode, etc.).

## Request

### HTTP Request

```
POST https://www.googleapis.com/youtube/v3/liveChat/messages/transition
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `part` | Yes | string | Must include `snippet`. |

### Authorization

Requires OAuth 2.0 authorization with the following scope:

- `https://www.googleapis.com/auth/youtube.force-ssl`

The authenticated user must be the broadcast owner.

### Request Body

```json
{
  "snippet": {
    "liveChatId": "string",
    "type": "string",
    "slowModeDelayMs": long
  }
}
```

### Chat Mode Types

| Type | Description |
|------|-------------|
| `subscribersOnlyModeEvent` | Only subscribers can chat. |
| `membersOnlyModeEvent` | Only channel members can chat. |
| `slowModeEvent` | Slow mode - delay between messages. |
| `normalChatEvent` | Normal mode - everyone can chat freely. |

### Slow Mode Delay

When setting slow mode, specify the delay in milliseconds:

| Value | Delay |
|-------|-------|
| 1000 | 1 second |
| 5000 | 5 seconds |
| 30000 | 30 seconds |
| 60000 | 1 minute |
| 120000 | 2 minutes |

## Response

If successful, this method returns an empty response body with HTTP status code 204.

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `liveChatIdRequired` | The liveChatId is required. |
| 400 | `invalidChatMode` | The chat mode type is invalid. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | Only the broadcast owner can change chat mode. |
| 403 | `liveChatEnded` | The live chat has ended. |
| 404 | `liveChatNotFound` | The live chat does not exist. |

## Quota Cost

This method consumes **50 quota units**.

---

## Yougopher Implementation

### TransitionChatMode

Change the chat mode.

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

// Enable subscribers only mode
err := poller.TransitionChatMode(ctx, streaming.ChatModeSubscribersOnly)
if err != nil {
    log.Fatal(err)
}
```

### Chat Mode Constants

```go
streaming.ChatModeSubscribersOnly // Only subscribers can chat
streaming.ChatModeMembersOnly     // Only channel members can chat
streaming.ChatModeSlowMode        // Slow mode enabled
streaming.ChatModeNormal          // Normal chat mode
```

### Enable Slow Mode

Use `TransitionChatModeWithDelay` for slow mode with a specific delay.

```go
// Enable slow mode with 30 second delay
err := poller.TransitionChatModeWithDelay(ctx, streaming.ChatModeSlowMode, 30000)
```

### Disable Slow Mode

Return to normal mode to disable slow mode.

```go
err := poller.TransitionChatMode(ctx, streaming.ChatModeNormal)
```

### Enable Members Only

Restrict chat to channel members.

```go
err := poller.TransitionChatMode(ctx, streaming.ChatModeMembersOnly)
```

### Enable Subscribers Only

Restrict chat to subscribers.

```go
err := poller.TransitionChatMode(ctx, streaming.ChatModeSubscribersOnly)
```

### Permission Requirements

Only the broadcast owner can change chat modes. Moderators cannot change chat modes.

### Use Cases

1. **Spam attack**: Enable subscribers/members only mode temporarily.
2. **Giveaways**: Enable slow mode to manage participation.
3. **Q&A sessions**: Enable slow mode for ordered questions.
4. **VIP events**: Use members only mode for exclusive streams.
