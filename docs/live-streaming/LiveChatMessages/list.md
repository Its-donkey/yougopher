---
layout: default
title: LiveChatMessages.list
description: Lists live chat messages for a specified chat
---

Lists live chat messages for a specified chat.

## Request

### HTTP Request

```
GET https://www.googleapis.com/youtube/v3/liveChat/messages
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `liveChatId` | Yes | string | The ID of the live chat to retrieve messages from. |
| `part` | Yes | string | Comma-separated list of resource parts. Valid values: `id`, `snippet`, `authorDetails`. |
| `hl` | No | string | Language code for localized Super Chat currency display. |
| `maxResults` | No | integer | Maximum messages to return (200-2000). Default: 500. |
| `pageToken` | No | string | Token for pagination/resumption. |
| `profileImageSize` | No | string | Profile image size: `default` (88px), `medium` (240px), `high` (800px). |

### Authorization

Requires OAuth 2.0 authorization with one of the following scopes:

- `https://www.googleapis.com/auth/youtube.readonly`
- `https://www.googleapis.com/auth/youtube`
- `https://www.googleapis.com/auth/youtube.force-ssl`

### Request Body

Do not provide a request body when calling this method.

## Response

```json
{
  "kind": "youtube#liveChatMessageListResponse",
  "etag": "string",
  "nextPageToken": "string",
  "pollingIntervalMillis": integer,
  "offlineAt": "datetime",
  "pageInfo": {
    "totalResults": integer,
    "resultsPerPage": integer
  },
  "items": [
    liveChatMessage Resource
  ]
}
```

### Properties

| Property | Type | Description |
|----------|------|-------------|
| `nextPageToken` | string | Token to use for the next request to get new messages. |
| `pollingIntervalMillis` | integer | Minimum time to wait before polling again (ms). |
| `offlineAt` | datetime | When the chat ended. Present only if chat has ended. |
| `items[]` | list | List of chat messages. |

### liveChatMessage Resource

| Property | Type | Description |
|----------|------|-------------|
| `id` | string | The message's unique identifier. |
| `snippet.type` | string | Message type (see Message Types below). |
| `snippet.liveChatId` | string | The live chat ID. |
| `snippet.authorChannelId` | string | The author's channel ID. |
| `snippet.publishedAt` | datetime | When the message was sent. |
| `snippet.displayMessage` | string | The message text. |
| `authorDetails.channelId` | string | Author's channel ID. |
| `authorDetails.displayName` | string | Author's display name. |
| `authorDetails.profileImageUrl` | string | Author's profile image URL. |
| `authorDetails.isChatOwner` | boolean | Is the broadcast owner. |
| `authorDetails.isChatModerator` | boolean | Is a moderator. |
| `authorDetails.isChatSponsor` | boolean | Is a channel member. |
| `authorDetails.isVerified` | boolean | Is a verified channel. |

### Message Types

| Type | Description |
|------|-------------|
| `textMessageEvent` | Regular text message. |
| `superChatEvent` | Super Chat donation. |
| `superStickerEvent` | Super Sticker. |
| `newSponsorEvent` | New channel membership. |
| `memberMilestoneChatEvent` | Membership milestone. |
| `membershipGiftingEvent` | Gifted memberships (sender). |
| `giftMembershipReceivedEvent` | Received gift (recipient). |
| `messageDeletedEvent` | Message was deleted. |
| `userBannedEvent` | User was banned/timed out. |
| `pollEvent` | Poll-related event. |

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `liveChatIdRequired` | The liveChatId is required. |
| 403 | `forbidden` | Access to the live chat is forbidden. |
| 403 | `liveChatDisabled` | Live chat is disabled. |
| 403 | `liveChatEnded` | The live chat has ended. |
| 404 | `liveChatNotFound` | The live chat does not exist. |

## Quota Cost

This method consumes **5 quota units** per call.

---

## Yougopher Implementation

### LiveChatPoller

Use `LiveChatPoller` for automatic polling with handlers.

```go
poller := streaming.NewLiveChatPoller(client, liveChatID,
    streaming.WithMinPollInterval(2*time.Second),
    streaming.WithProfileImageSize(streaming.ProfileImageMedium),
)

// Register message handler
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

### Event Handlers

```go
// Handle regular messages
poller.OnMessage(func(msg *streaming.LiveChatMessage) {
    switch msg.Type() {
    case streaming.MessageTypeText:
        fmt.Printf("Text: %s\n", msg.Snippet.DisplayMessage)
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

### Polling Configuration

```go
poller := streaming.NewLiveChatPoller(client, liveChatID,
    streaming.WithMinPollInterval(1*time.Second),  // Minimum time between polls
    streaming.WithMaxPollInterval(30*time.Second), // Maximum time between polls
    streaming.WithProfileImageSize(streaming.ProfileImageHigh), // 800px images
    streaming.WithBackoff(core.NewBackoffConfig(
        core.WithBaseDelay(1*time.Second),
        core.WithMaxDelay(30*time.Second),
    )),
)
```

### Page Token Management

```go
// Save token before stopping
token := poller.PageToken()

// Later, resume from that position
poller.SetPageToken(token)
poller.Start(ctx)

// Clear token to start fresh
poller.ResetPageToken()
```

### Poll Interval

The API returns `pollingIntervalMillis` indicating how long to wait. The poller respects this within your configured bounds:

```go
// Get current poll interval
interval := poller.PollInterval()
```

### Alternative: SSE Streaming

For lower latency, use [streamList](stream-list.md) (Server-Sent Events) instead of polling.
