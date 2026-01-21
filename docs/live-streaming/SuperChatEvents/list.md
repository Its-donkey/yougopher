---
layout: default
title: ListSuperChatEvents
description: List historical Super Chat events
---

Lists historical Super Chat and Super Sticker events for the authenticated user's channel.

**Quota Cost:** 5 units

## ListSuperChatEvents

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

## Calculate Total Revenue

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

## Filter Super Stickers

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

## Real-Time Super Chats

For real-time Super Chat notifications during a live stream, use event handlers:

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

poller.OnMessage(func(msg *streaming.LiveChatMessage) {
    if msg.IsSuperChat() {
        details := msg.Snippet.SuperChatDetails
        fmt.Printf("SUPER CHAT! %s sent %s: %s\n",
            msg.AuthorDetails.DisplayName,
            details.AmountDisplayString,
            msg.Message())
    }

    if msg.IsSuperSticker() {
        details := msg.Snippet.SuperStickerDetails
        fmt.Printf("SUPER STICKER! %s sent %s\n",
            msg.AuthorDetails.DisplayName,
            details.AmountDisplayString)
    }
})
```

## Notes

- `ListSuperChatEvents` returns historical data, not real-time events
- For real-time Super Chats, use `LiveChatPoller` or `LiveChatStream`
- Super Chat amounts are in "micros" (1 dollar = 1,000,000 micros)

## Common Errors

| Error | Description |
|-------|-------------|
| `superChatNotEnabled` | Super Chat not enabled on channel |
| `ForbiddenError` | No access to Super Chat events |
