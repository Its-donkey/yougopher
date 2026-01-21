---
layout: default
title: LiveChatMessages
description: Live chat message operations for YouTube streaming
---

A `LiveChatMessage` represents a chat message in a YouTube live chat. Messages can be text, Super Chats, memberships, and other event types.

## Functions

| Function | Quota | Description |
|----------|-------|-------------|
| [LiveChatPoller](list) | 5/poll | Automatic polling with handlers |
| [LiveChatStream](stream-list) | 5/conn | SSE streaming for real-time messages |
| [SendMessage](insert) | 50 | Send a message to chat |
| [DeleteMessage](delete) | 50 | Delete a message from chat |
| [TransitionChatMode](transition) | 50 | Change chat mode (slow, members only, etc.) |

## Type Definition

```go
type LiveChatMessage struct {
    Kind          string         `json:"kind,omitempty"`
    ETag          string         `json:"etag,omitempty"`
    ID            string         `json:"id,omitempty"`
    Snippet       *MessageSnippet `json:"snippet,omitempty"`
    AuthorDetails *AuthorDetails  `json:"authorDetails,omitempty"`
}
```

## Helper Methods

```go
msg.Type()              // Returns the message type
msg.Message()           // Returns the display message text
msg.IsTextMessage()     // True if regular text
msg.IsSuperChat()       // True if Super Chat
msg.IsSuperSticker()    // True if Super Sticker
msg.IsMembership()      // True if new membership
msg.IsMemberMilestone() // True if membership milestone
msg.IsGiftMembership()  // True if gift membership
msg.IsPoll()            // True if poll event
msg.HasActivePoll()     // True if has active poll
msg.ActivePoll()        // Returns active poll item
```

## Message Types

| Type | Constant | Description |
|------|----------|-------------|
| `textMessageEvent` | `MessageTypeText` | Regular text message |
| `superChatEvent` | `MessageTypeSuperChat` | Super Chat donation |
| `superStickerEvent` | `MessageTypeSuperSticker` | Super Sticker |
| `newSponsorEvent` | `MessageTypeMembership` | New membership |
| `memberMilestoneChatEvent` | `MessageTypeMemberMilestone` | Milestone |
| `membershipGiftingEvent` | `MessageTypeMembershipGifting` | Gift (sender) |
| `giftMembershipReceivedEvent` | `MessageTypeGiftMembershipReceived` | Gift (recipient) |
| `messageDeletedEvent` | `MessageTypeMessageDeleted` | Deleted |
| `userBannedEvent` | `MessageTypeUserBanned` | Ban/timeout |
| `pollEvent` | `MessageTypePoll` | Poll event |

## Polling vs SSE Streaming

| Feature | [Polling](list) | [SSE Streaming](stream-list) |
|---------|-----------------|------------------------------|
| Latency | Higher (poll interval) | Lower (~instant) |
| Connection | Repeated requests | Long-lived HTTP |
| Quota | 5 units per poll | 5 units per connection |
| Reconnection | Manual | Automatic with backoff |
| Use case | Custom polling logic | Real-time bots |

## Quick Example

```go
import "github.com/Its-donkey/yougopher/youtube/streaming"

// Create a poller
poller := streaming.NewLiveChatPoller(client, liveChatID)

// Handle messages
poller.OnMessage(func(msg *streaming.LiveChatMessage) {
    fmt.Printf("[%s] %s\n",
        msg.AuthorDetails.DisplayName,
        msg.Message())
})

// Handle Super Chats
poller.OnMessage(func(msg *streaming.LiveChatMessage) {
    if msg.IsSuperChat() {
        sc := msg.Snippet.SuperChatDetails
        fmt.Printf("Super Chat: %s from %s\n",
            sc.AmountDisplayString,
            msg.AuthorDetails.DisplayName)
    }
})

// Start polling
err := poller.Start(ctx)
defer poller.Stop()
```
