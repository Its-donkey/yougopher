---
layout: default
title: LiveChatMessages.streamList
description: Establishes a Server-Sent Events (SSE) connection to receive live chat messages in real-time
---

Establishes a Server-Sent Events (SSE) connection to receive live chat messages in real-time.

This method provides lower latency than polling by maintaining a long-lived HTTP connection that pushes messages as they arrive.

## Request

### HTTP Request

```
GET https://www.googleapis.com/youtube/v3/liveChat/messages/stream
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `liveChatId` | Yes | string | The ID of the live chat to stream messages from. |
| `part` | Yes | string | Comma-separated list of resource parts to include. Valid values: `id`, `snippet`, `authorDetails`. |
| `hl` | No | string | Language code for localized Super Chat currency display. |
| `maxResults` | No | integer | Maximum messages per response. Acceptable values: 200-2000. Default: 500. |
| `pageToken` | No | string | Token to resume from a specific position in the chat history. |
| `profileImageSize` | No | integer | Profile image size in pixels. Acceptable values: 16-720. Default: 88. |

### Authorization

Requires OAuth 2.0 authorization with the following scope:

- `https://www.googleapis.com/auth/youtube.readonly`

### Request Body

Do not provide a request body when calling this method.

## Response

If successful, this method returns a stream of SSE events. Each `data:` event contains a JSON object with the following structure:

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
  ],
  "activePollItem": liveChatPollItem Resource
}
```

### Properties

| Property | Type | Description |
|----------|------|-------------|
| `kind` | string | Identifies the resource type. Value: `"youtube#liveChatMessageListResponse"`. |
| `etag` | string | The ETag of the response. |
| `nextPageToken` | string | Token to resume the stream from this position if reconnecting. |
| `pollingIntervalMillis` | integer | Minimum interval before requesting new messages (for polling fallback). |
| `offlineAt` | datetime | Timestamp when the live stream ended. Present only if chat has ended. |
| `pageInfo.totalResults` | integer | Total number of results in the result set. |
| `pageInfo.resultsPerPage` | integer | Number of results included in the response. |
| `items[]` | list | List of live chat messages. See [liveChatMessage resource](https://developers.google.com/youtube/v3/live/docs/liveChatMessages#resource). |
| `activePollItem` | object | Currently active poll, if any. |

### Message Types

The `items[].snippet.type` field indicates the message type:

| Type | Description |
|------|-------------|
| `textMessageEvent` | Regular text message |
| `superChatEvent` | Super Chat donation |
| `superStickerEvent` | Super Sticker |
| `newSponsorEvent` | New channel membership |
| `memberMilestoneChatEvent` | Membership milestone celebration |
| `membershipGiftingEvent` | Gifted memberships (sender) |
| `giftMembershipReceivedEvent` | Gifted membership (recipient) |
| `messageDeletedEvent` | Message was deleted |
| `userBannedEvent` | User was banned/timed out |
| `pollEvent` | Poll-related event |

## Errors

### gRPC Errors

| Code | Description |
|------|-------------|
| 3 (INVALID_ARGUMENT) | The `liveChatId` parameter is missing or invalid. |
| 5 (NOT_FOUND) | The live chat specified by `liveChatId` cannot be found. |
| 7 (PERMISSION_DENIED) | The request is not authorized to access the specified live chat. |
| 8 (RESOURCE_EXHAUSTED) | API quota has been exceeded. |
| 9 (FAILED_PRECONDITION) | The live chat has ended or is disabled. |

### HTTP Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `badRequest` | Required parameter is missing or invalid. |
| 403 | `forbidden` | Access to the live chat is forbidden. |
| 403 | `liveChatDisabled` | Live chat is disabled for this broadcast. |
| 403 | `liveChatEnded` | The live chat has ended. |
| 403 | `quotaExceeded` | Daily API quota has been exceeded. |
| 404 | `liveChatNotFound` | The specified live chat does not exist. |

## Quota Cost

This method consumes **5 quota units** per connection.

SSE streaming is more quota-efficient than polling for active chats since you pay once per connection rather than per poll interval.

---

## Yougopher Implementation

### LiveChatStream

Create and manage an SSE connection using `LiveChatStream`:

```go
stream := streaming.NewLiveChatStream(client, liveChatID,
    streaming.WithStreamAccessToken(accessToken),
    streaming.WithStreamParts("id", "snippet", "authorDetails"),
    streaming.WithStreamMaxResults(1000),
)
```

### Options

| Option | Parameter | Description |
|--------|-----------|-------------|
| `WithStreamAccessToken(token)` | `Authorization` header | OAuth access token |
| `WithStreamParts(parts...)` | `part` | Resource parts to include |
| `WithStreamHL(lang)` | `hl` | Localized currency language |
| `WithStreamMaxResults(n)` | `maxResults` | Messages per response (200-2000) |
| `WithStreamProfileImageSize(px)` | `profileImageSize` | Profile image size (16-720) |
| `WithStreamReconnectDelay(d)` | - | Initial reconnect delay |
| `WithStreamMaxReconnectDelay(d)` | - | Maximum reconnect delay |
| `WithStreamBackoff(cfg)` | - | Custom backoff configuration |
| `WithStreamHTTPClient(hc)` | - | Custom HTTP client |

### Lifecycle

```go
// Start the stream
err := stream.Start(ctx)

// Check if running
if stream.IsRunning() { ... }

// Stop the stream
stream.Stop()
```

### Event Handlers

All handlers return an unsubscribe function:

```go
// Handle messages
unsub := stream.OnMessage(func(msg *streaming.LiveChatMessage) {
    fmt.Printf("[%s] %s\n", msg.AuthorDetails.DisplayName, msg.Snippet.DisplayMessage)
})

// Handle full responses (includes metadata)
stream.OnResponse(func(resp *streaming.LiveChatMessageListResponse) {
    if resp.IsChatEnded() {
        fmt.Println("Chat ended")
    }
})

// Handle deletions
stream.OnDelete(func(messageID string) { ... })

// Handle bans
stream.OnBan(func(details *streaming.UserBannedDetails) { ... })

// Handle connection events
stream.OnConnect(func() { ... })
stream.OnDisconnect(func() { ... })

// Handle errors
stream.OnError(func(err error) { ... })

// Unsubscribe when done
unsub()
```

### Stream Resumption

Use page tokens to resume from a specific position:

```go
// Save token before stopping
token := stream.PageToken()

// Resume from saved position
stream.SetPageToken(token)
stream.Start(ctx)

// Clear token to start fresh
stream.ResetPageToken()
```

### Error Handling

```go
stream.OnError(func(err error) {
    var apiErr *core.APIError
    if errors.As(err, &apiErr) {
        switch {
        case apiErr.IsChatEnded():
            log.Println("Chat ended")
        case apiErr.IsChatDisabled():
            log.Println("Chat disabled")
        case apiErr.IsQuotaExceeded():
            log.Println("Quota exceeded")
        }
    }
})
```

### Automatic Reconnection

The stream automatically reconnects with exponential backoff:

- Initial delay: 1 second
- Multiplier: 2x per attempt
- Maximum delay: 30 seconds
- Jitter: ±20%

Configure with `WithStreamBackoff`:

```go
stream := streaming.NewLiveChatStream(client, liveChatID,
    streaming.WithStreamBackoff(core.NewBackoffConfig(
        core.WithBaseDelay(500*time.Millisecond),
        core.WithMaxDelay(60*time.Second),
        core.WithMultiplier(1.5),
        core.WithJitter(0.1),
    )),
)
```
