---
layout: default
title: SuperChatEvents
description: Super Chat and Super Sticker monetary contributions
---

A `SuperChatEventResource` represents a Super Chat or Super Sticker purchase. This resource provides historical data about Super Chats received on your channel.

## Functions

| Function | Quota | Description |
|----------|-------|-------------|
| [ListSuperChatEvents](list) | 5 | List historical Super Chat events |

## Type Definition

```go
type SuperChatEventResource struct {
    Kind    string                       `json:"kind,omitempty"`
    ETag    string                       `json:"etag,omitempty"`
    ID      string                       `json:"id,omitempty"`
    Snippet *SuperChatEventResourceSnippet `json:"snippet,omitempty"`
}
```

## Amount Conversion

Super Chat amounts are in "micros" (millionths):

```go
// Convert micros to dollars
dollars := float64(event.Snippet.AmountMicros) / 1_000_000

// Example: 5000000 micros = $5.00
```

## Real-Time vs Historical

| Use Case | Method |
|----------|--------|
| Historical Super Chats | `ListSuperChatEvents` |
| Real-time during stream | `LiveChatPoller` or `LiveChatStream` |

## Super Chat Tiers

| Tier | Amount (USD) | Color |
|------|-------------|-------|
| 1 | $1-1.99 | Blue |
| 2 | $2-4.99 | Cyan |
| 3 | $5-9.99 | Green |
| 4 | $10-19.99 | Yellow |
| 5 | $20-49.99 | Orange |
| 6 | $50-99.99 | Magenta |
| 7 | $100+ | Red |

## Quick Example

```go
import "github.com/Its-donkey/yougopher/youtube/streaming"

// Historical Super Chats
resp, err := streaming.ListSuperChatEvents(ctx, client, &streaming.ListSuperChatEventsParams{
    MaxResults: 50,
})

for _, event := range resp.Items {
    fmt.Printf("Super Chat from %s: %s\n",
        event.Snippet.SupporterDetails.DisplayName,
        event.Snippet.DisplayString)
}

// Real-time Super Chats
poller := streaming.NewLiveChatPoller(client, liveChatID)
poller.OnMessage(func(msg *streaming.LiveChatMessage) {
    if msg.IsSuperChat() {
        fmt.Printf("SUPER CHAT! %s sent %s\n",
            msg.AuthorDetails.DisplayName,
            msg.Snippet.SuperChatDetails.AmountDisplayString)
    }
})
```
