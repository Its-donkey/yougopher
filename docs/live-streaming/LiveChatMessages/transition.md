---
layout: default
title: TransitionChatMode
description: Change the chat mode
---

Changes the chat mode (subscribers only, members only, slow mode, etc.).

**Quota Cost:** 50 units

## TransitionChatMode

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

// Enable subscribers only mode
err := poller.TransitionChatMode(ctx, streaming.ChatModeSubscribersOnly)
if err != nil {
    log.Fatal(err)
}
```

## Chat Mode Constants

```go
streaming.ChatModeSubscribersOnly // Only subscribers can chat
streaming.ChatModeMembersOnly     // Only channel members can chat
streaming.ChatModeSlowMode        // Slow mode enabled
streaming.ChatModeNormal          // Normal chat mode
```

## Enable Slow Mode

Use `TransitionChatModeWithDelay` for slow mode with a specific delay:

```go
// Enable slow mode with 30 second delay
err := poller.TransitionChatModeWithDelay(ctx, streaming.ChatModeSlowMode, 30000)
```

### Delay Values

| Milliseconds | Delay |
|--------------|-------|
| 1000 | 1 second |
| 5000 | 5 seconds |
| 30000 | 30 seconds |
| 60000 | 1 minute |
| 120000 | 2 minutes |

## Disable Slow Mode

Return to normal mode:

```go
err := poller.TransitionChatMode(ctx, streaming.ChatModeNormal)
```

## Enable Members Only

Restrict chat to channel members:

```go
err := poller.TransitionChatMode(ctx, streaming.ChatModeMembersOnly)
```

## Enable Subscribers Only

Restrict chat to subscribers:

```go
err := poller.TransitionChatMode(ctx, streaming.ChatModeSubscribersOnly)
```

## Permission Requirements

Only the broadcast owner can change chat modes. Moderators cannot change chat modes.

## Use Cases

| Scenario | Mode |
|----------|------|
| Spam attack | Subscribers/members only (temporary) |
| Giveaways | Slow mode for ordered participation |
| Q&A sessions | Slow mode for ordered questions |
| VIP events | Members only for exclusive streams |

## Common Errors

| Error | Description |
|-------|-------------|
| `invalidChatMode` | Invalid mode type |
| `ForbiddenError` | Not the broadcast owner |
| `liveChatEnded` | Chat has ended |
