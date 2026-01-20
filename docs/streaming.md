---
layout: default
title: Streaming API
description: Build YouTube Live chat bots with real-time message handling and moderation.
---

## Overview

Build interactive chat bots for YouTube Live streams:

**ChatBotClient (Recommended):** High-level bot with semantic events
- Receive typed events (messages, Super Chats, memberships)
- Automatic token refresh and reconnection
- Built-in panic recovery for handlers

**LiveChatPoller (Advanced):** Low-level polling control
- Direct access to raw API responses
- Custom polling intervals and backoff
- Fine-grained control over message processing

**Moderation:** Full chat moderation support
- Ban/timeout users
- Delete messages
- Add/remove moderators

## Prerequisites

- **Read chat:** `youtube` scope
- **Moderate chat:** `youtube.force-ssl` scope

## ChatBotClient

### NewChatBotClient

Create a new high-level chat bot client.

```go
bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
if err != nil {
    log.Fatal(err)
}
```

### Connect

Start the bot and begin listening for messages.

```go
err := bot.Connect(ctx)
if err != nil {
    log.Fatal(err)
}
defer bot.Close()
```

### OnMessage

Register a handler for chat messages.

```go
unsub := bot.OnMessage(func(msg *streaming.ChatMessage) {
    fmt.Printf("[%s] %s\n", msg.Author.DisplayName, msg.Message)
})
// Later: unsub() to unregister
```

**ChatMessage fields:**
```go
type ChatMessage struct {
    ID          string
    Message     string
    Author      *Author
    PublishedAt time.Time
    Raw         *LiveChatMessage
}
```

### OnSuperChat

Register a handler for Super Chat donations.

```go
bot.OnSuperChat(func(event *streaming.SuperChatEvent) {
    fmt.Printf("Super Chat from %s: %s - %s\n",
        event.Author.DisplayName, event.Amount, event.Message)
})
```

**SuperChatEvent fields:**
```go
type SuperChatEvent struct {
    ID           string
    Author       *Author
    Message      string
    Amount       string    // e.g., "$5.00"
    AmountMicros int64     // e.g., 5000000
    Currency     string    // e.g., "USD"
    Tier         int       // 1-7
    Raw          *LiveChatMessage
}
```

### OnSuperSticker

Register a handler for Super Sticker events.

```go
bot.OnSuperSticker(func(event *streaming.SuperStickerEvent) {
    fmt.Printf("Super Sticker from %s: %s (%s)\n",
        event.Author.DisplayName, event.AltText, event.Amount)
})
```

### OnMembership

Register a handler for new channel memberships.

```go
bot.OnMembership(func(event *streaming.MembershipEvent) {
    fmt.Printf("New member: %s joined at %s level\n",
        event.Author.DisplayName, event.LevelName)
})
```

### OnMemberMilestone

Register a handler for member milestone celebrations.

```go
bot.OnMemberMilestone(func(event *streaming.MemberMilestoneEvent) {
    fmt.Printf("%s has been a member for %d months!\n",
        event.Author.DisplayName, event.Months)
})
```

### OnGiftMembership

Register a handler for gifted memberships.

```go
bot.OnGiftMembership(func(event *streaming.GiftMembershipEvent) {
    fmt.Printf("%s gifted %d memberships!\n",
        event.Author.DisplayName, event.Count)
})
```

### OnMessageDeleted

Register a handler for message deletions.

```go
bot.OnMessageDeleted(func(messageID string) {
    fmt.Printf("Message deleted: %s\n", messageID)
})
```

### OnUserBanned

Register a handler for user bans.

```go
bot.OnUserBanned(func(event *streaming.BanEvent) {
    fmt.Printf("User banned: %s (%s)\n",
        event.BannedUser.DisplayName, event.BanType)
})
```

### OnConnect / OnDisconnect

Register handlers for connection state changes.

```go
bot.OnConnect(func() {
    fmt.Println("Bot connected!")
})

bot.OnDisconnect(func() {
    fmt.Println("Bot disconnected")
})
```

### OnError

Register a handler for errors.

```go
bot.OnError(func(err error) {
    log.Printf("Bot error: %v", err)
})
```

## Sending Messages

### Say

Send a message to the chat.

**Requires:** youtube.force-ssl scope

```go
err := bot.Say(ctx, "Hello, chat!")
if err != nil {
    log.Printf("Failed to send message: %v", err)
}
```

## Moderation

### Delete

Delete a message from the chat.

**Requires:** youtube.force-ssl scope

```go
err := bot.Delete(ctx, messageID)
if err != nil {
    log.Printf("Failed to delete message: %v", err)
}
```

### Ban

Permanently ban a user from the chat.

**Requires:** youtube.force-ssl scope

```go
err := bot.Ban(ctx, channelID)
if err != nil {
    log.Printf("Failed to ban user: %v", err)
}
```

### Timeout

Temporarily ban a user from the chat.

**Requires:** youtube.force-ssl scope

```go
err := bot.Timeout(ctx, channelID, 300) // 5 minute timeout
if err != nil {
    log.Printf("Failed to timeout user: %v", err)
}
```

### Unban

Remove a ban from a user.

**Requires:** youtube.force-ssl scope

```go
err := bot.Unban(ctx, banID)
if err != nil {
    log.Printf("Failed to unban user: %v", err)
}
```

### AddModerator

Add a moderator to the chat.

**Requires:** youtube.force-ssl scope

```go
err := bot.AddModerator(ctx, channelID)
if err != nil {
    log.Printf("Failed to add moderator: %v", err)
}
```

### RemoveModerator

Remove a moderator from the chat.

**Requires:** youtube.force-ssl scope

```go
err := bot.RemoveModerator(ctx, moderatorID)
if err != nil {
    log.Printf("Failed to remove moderator: %v", err)
}
```

## LiveChatPoller (Advanced)

For more control over polling behavior, use LiveChatPoller directly.

### NewLiveChatPoller

Create a new poller with optional configuration.

```go
poller := streaming.NewLiveChatPoller(client, liveChatID,
    streaming.WithMinPollInterval(2*time.Second),
    streaming.WithMaxPollInterval(30*time.Second),
    streaming.WithBackoff(&core.BackoffConfig{
        InitialDelay: time.Second,
        MaxDelay:     30 * time.Second,
        Multiplier:   2.0,
    }),
)
```

### Start / Stop

Control the polling lifecycle.

```go
err := poller.Start(ctx)
if err != nil {
    log.Fatal(err)
}

// Later...
poller.Stop()
```

### OnMessage (Poller)

Handle raw message events with full API data.

```go
poller.OnMessage(func(msg *streaming.LiveChatMessage) {
    fmt.Printf("Type: %s, Message: %s\n", msg.Type(), msg.Message())
})
```

### Reset

Reset polling state for reuse (must be stopped).

```go
err := poller.Reset()
if err != nil {
    log.Printf("Cannot reset: %v", err) // Returns ErrAlreadyRunning if running
}
```

## Author Information

All event types include author details:

```go
type Author struct {
    ChannelID       string
    DisplayName     string
    ProfileImageURL string
    IsModerator     bool
    IsOwner         bool
    IsMember        bool
    IsVerified      bool
}
```

## Handler Unsubscription

All `On*` methods return an unsubscribe function:

```go
unsub := bot.OnMessage(handler)

// Later, to stop receiving events:
unsub() // Safe to call multiple times
```

## Thread Safety

Both `ChatBotClient` and `LiveChatPoller` are safe for concurrent use. Handler registration and unsubscription can be done from any goroutine.

## Error Handling

Handlers are protected with panic recovery. If a handler panics, the error is dispatched to error handlers and polling continues normally.
