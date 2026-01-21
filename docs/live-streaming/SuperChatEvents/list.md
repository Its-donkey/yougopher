---
layout: default
title: SuperChatEvents.list
description: Lists Super Chat and Super Sticker events for the authenticated user's channel
---

Lists Super Chat and Super Sticker events for the authenticated user's channel.

This provides historical data about Super Chats received, not real-time events. For real-time Super Chat notifications during a live stream, use the `LiveChatPoller` or `ChatBotClient` event handlers.

## Request

### HTTP Request

```
GET https://www.googleapis.com/youtube/v3/superChatEvents
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `part` | Yes | string | Must include `id` and/or `snippet`. |
| `hl` | No | string | Language code for localized resource properties. |
| `maxResults` | No | integer | Maximum number of items to return (1-50). Default: 5. |
| `pageToken` | No | string | Token for pagination. |

### Authorization

Requires OAuth 2.0 authorization with one of the following scopes:

- `https://www.googleapis.com/auth/youtube.readonly`
- `https://www.googleapis.com/auth/youtube.force-ssl`

### Request Body

Do not provide a request body when calling this method.

## Response

```json
{
  "kind": "youtube#superChatEventListResponse",
  "etag": "string",
  "nextPageToken": "string",
  "pageInfo": {
    "totalResults": integer,
    "resultsPerPage": integer
  },
  "items": [
    {
      "kind": "youtube#superChatEvent",
      "etag": "string",
      "id": "string",
      "snippet": {
        "channelId": "string",
        "supporterDetails": {
          "channelId": "string",
          "channelUrl": "string",
          "displayName": "string",
          "profileImageUrl": "string"
        },
        "commentText": "string",
        "createdAt": "datetime",
        "amountMicros": long,
        "currency": "string",
        "displayString": "string",
        "messageType": integer,
        "isSuperStickerEvent": boolean,
        "superStickerMetadata": {
          "stickerId": "string",
          "altText": "string",
          "language": "string"
        }
      }
    }
  ]
}
```

### Properties

| Property | Type | Description |
|----------|------|-------------|
| `id` | string | The Super Chat event ID. |
| `snippet.channelId` | string | Your channel ID (recipient). |
| `snippet.supporterDetails` | object | Details about the supporter. |
| `snippet.supporterDetails.displayName` | string | Supporter's display name. |
| `snippet.commentText` | string | The message sent with the Super Chat. |
| `snippet.createdAt` | datetime | When the Super Chat was sent. |
| `snippet.amountMicros` | long | Amount in micros (divide by 1,000,000 for dollars). |
| `snippet.currency` | string | Currency code (e.g., "USD", "EUR"). |
| `snippet.displayString` | string | Formatted amount (e.g., "$5.00"). |
| `snippet.messageType` | integer | 1 for Super Chat, 2 for Super Sticker. |
| `snippet.isSuperStickerEvent` | boolean | True if this is a Super Sticker. |
| `snippet.superStickerMetadata` | object | Sticker details (if Super Sticker). |

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | The user cannot access Super Chat events. |
| 403 | `superChatNotEnabled` | Super Chat is not enabled for this channel. |

## Quota Cost

This method consumes **5 quota units**.

---

## Yougopher Implementation

### ListSuperChatEvents

List historical Super Chat events.

```go
resp, err := streaming.ListSuperChatEvents(ctx, client, &streaming.ListSuperChatEventsParams{
    MaxResults: 50,
})
if err != nil {
    log.Fatal(err)
}

for _, event := range resp.Items {
    fmt.Printf("Super Chat from %s: %s - \"%s\"\n",
        event.Snippet.SupporterDetails.DisplayName,
        event.Snippet.DisplayString,
        event.Snippet.CommentText)
}
```

### Calculate Total Revenue

```go
var totalMicros int64

pageToken := ""
for {
    resp, err := streaming.ListSuperChatEvents(ctx, client, &streaming.ListSuperChatEventsParams{
        MaxResults: 50,
        PageToken:  pageToken,
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, event := range resp.Items {
        totalMicros += event.Snippet.AmountMicros
    }

    if resp.NextPageToken == "" {
        break
    }
    pageToken = resp.NextPageToken
}

totalDollars := float64(totalMicros) / 1_000_000
fmt.Printf("Total Super Chat revenue: $%.2f\n", totalDollars)
```

### Filter Super Stickers

```go
resp, err := streaming.ListSuperChatEvents(ctx, client, &streaming.ListSuperChatEventsParams{
    MaxResults: 50,
})

for _, event := range resp.Items {
    if event.Snippet.IsSuperStickerEvent {
        fmt.Printf("Super Sticker: %s\n", event.Snippet.SuperStickerMetadata.AltText)
    } else {
        fmt.Printf("Super Chat: %s\n", event.Snippet.DisplayString)
    }
}
```

### Real-Time Super Chats

For real-time Super Chat notifications during a live stream, use event handlers:

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

poller.OnMessage(func(msg *streaming.LiveChatMessage) {
    if msg.Type() == streaming.MessageTypeSuperChat {
        details := msg.Snippet.SuperChatDetails
        fmt.Printf("SUPER CHAT! %s sent %s: %s\n",
            msg.AuthorDetails.DisplayName,
            details.AmountDisplayString,
            msg.Snippet.DisplayMessage)
    }

    if msg.Type() == streaming.MessageTypeSuperSticker {
        details := msg.Snippet.SuperStickerDetails
        fmt.Printf("SUPER STICKER! %s sent %s\n",
            msg.AuthorDetails.DisplayName,
            details.AmountDisplayString)
    }
})
```

### Notes

- `superChatEvents.list` returns historical data, not real-time events.
- For real-time Super Chats, use `liveChatMessages.list` or SSE streaming.
- Super Chat amounts are in "micros" (1 dollar = 1,000,000 micros).
- Super Chats may be filtered by YouTube before reaching the list.
