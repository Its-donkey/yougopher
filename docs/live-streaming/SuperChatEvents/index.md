---
layout: default
title: SuperChatEvents
description: Super Chat and Super Sticker monetary contributions
---

A `superChatEvent` resource represents a Super Chat or Super Sticker purchase.

This resource provides historical data about Super Chats received on your channel. For real-time Super Chat notifications during a live stream, use `liveChatMessages.list` or SSE streaming and filter for `superChatEvent` and `superStickerEvent` message types.

## Methods

| Method | HTTP Request | Description |
|--------|--------------|-------------|
| [list](list) | `GET /superChatEvents` | Lists Super Chat events for the authenticated user's channel. |

## Resource Representation

```json
{
  "kind": "youtube#superChatEvent",
  "etag": string,
  "id": string,
  "snippet": {
    "channelId": string,
    "supporterDetails": {
      "channelId": string,
      "channelUrl": string,
      "displayName": string,
      "profileImageUrl": string
    },
    "commentText": string,
    "createdAt": datetime,
    "amountMicros": unsigned long,
    "currency": string,
    "displayString": string,
    "messageType": unsigned integer,
    "isSuperStickerEvent": boolean,
    "superStickerMetadata": {
      "stickerId": string,
      "altText": string,
      "language": string
    }
  }
}
```

## Properties

| Property | Type | Description |
|----------|------|-------------|
| `kind` | string | Identifies the resource type. Value: `youtube#superChatEvent`. |
| `etag` | string | The ETag of the resource. |
| `id` | string | The ID that YouTube assigns to uniquely identify the Super Chat event. |

### snippet

| Property | Type | Description |
|----------|------|-------------|
| `snippet.channelId` | string | The channel ID that received the Super Chat (your channel). |
| `snippet.supporterDetails.channelId` | string | The supporter's channel ID. |
| `snippet.supporterDetails.channelUrl` | string | The URL of the supporter's channel. |
| `snippet.supporterDetails.displayName` | string | The supporter's display name. |
| `snippet.supporterDetails.profileImageUrl` | string | The URL of the supporter's profile image. |
| `snippet.commentText` | string | The message sent with the Super Chat. |
| `snippet.createdAt` | datetime | The date and time the Super Chat was sent. |
| `snippet.amountMicros` | long | The purchase amount in micros (1 dollar = 1,000,000 micros). |
| `snippet.currency` | string | The currency code (e.g., `USD`, `EUR`, `JPY`). |
| `snippet.displayString` | string | The formatted amount (e.g., `$5.00`). |
| `snippet.messageType` | integer | The message type: 1 for Super Chat, 2 for Super Sticker. |
| `snippet.isSuperStickerEvent` | boolean | True if this is a Super Sticker rather than Super Chat. |
| `snippet.superStickerMetadata.stickerId` | string | The Super Sticker's ID. |
| `snippet.superStickerMetadata.altText` | string | Alt text describing the sticker. |
| `snippet.superStickerMetadata.language` | string | The language of the alt text. |

---

## Yougopher Type

```go
type SuperChatEventResource struct {
    Kind    string                    `json:"kind,omitempty"`
    ETag    string                    `json:"etag,omitempty"`
    ID      string                    `json:"id,omitempty"`
    Snippet *SuperChatEventSnippet    `json:"snippet,omitempty"`
}
```

### Amount Conversion

Super Chat amounts are in "micros" (millionths):

```go
// Convert micros to dollars
dollars := float64(event.Snippet.AmountMicros) / 1_000_000

// Example: 5000000 micros = $5.00
```

### Real-Time vs Historical

| Use Case | Method |
|----------|--------|
| Historical Super Chats | `superChatEvents.list` |
| Real-time during stream | `liveChatMessages.list` or SSE streaming |

### Super Chat Tiers

Super Chats have different tiers based on the amount, which affects:
- Message highlight duration
- Message pinning duration
- Maximum message length

| Tier | Amount (USD) | Highlight |
|------|-------------|-----------|
| 1 | $1-1.99 | Blue |
| 2 | $2-4.99 | Cyan |
| 3 | $5-9.99 | Green |
| 4 | $10-19.99 | Yellow |
| 5 | $20-49.99 | Orange |
| 6 | $50-99.99 | Magenta |
| 7 | $100-199.99 | Red |
| 8 | $200-299.99 | Red |
| 9 | $300-399.99 | Red |
| 10 | $400-500 | Red |
