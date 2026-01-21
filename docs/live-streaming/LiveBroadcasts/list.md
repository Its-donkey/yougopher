---
layout: default
title: GetBroadcasts
description: Retrieve YouTube live broadcasts
---

Retrieve broadcasts matching filter parameters.

**Quota Cost:** 5 units

## GetBroadcasts

Retrieve multiple broadcasts with filtering options.

```go
resp, err := streaming.GetBroadcasts(ctx, client, &streaming.GetBroadcastsParams{
    Mine:            true,
    BroadcastStatus: "active",
    Parts:           []string{"snippet", "status", "contentDetails"},
    MaxResults:      10,
})
if err != nil {
    log.Fatal(err)
}

for _, broadcast := range resp.Items {
    fmt.Printf("%s: %s\n", broadcast.ID, broadcast.Snippet.Title)
}
```

### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `Mine` | bool | Retrieve the authenticated user's broadcasts |
| `IDs` | []string | Specific broadcast IDs to retrieve |
| `BroadcastStatus` | string | Filter: `active`, `all`, `completed`, `upcoming` |
| `BroadcastType` | string | Filter: `all`, `event`, `persistent` |
| `Parts` | []string | Resource parts: `snippet`, `status`, `contentDetails`, `statistics` |
| `MaxResults` | int | Maximum results (1-50, default 5) |
| `PageToken` | string | Pagination token |

### Pagination

```go
var allBroadcasts []*streaming.LiveBroadcast
pageToken := ""

for {
    resp, err := streaming.GetBroadcasts(ctx, client, &streaming.GetBroadcastsParams{
        Mine:      true,
        Parts:     []string{"snippet"},
        PageToken: pageToken,
    })
    if err != nil {
        log.Fatal(err)
    }

    allBroadcasts = append(allBroadcasts, resp.Items...)

    if resp.NextPageToken == "" {
        break
    }
    pageToken = resp.NextPageToken
}
```

## GetBroadcast

Retrieve a single broadcast by ID.

```go
broadcast, err := streaming.GetBroadcast(ctx, client, "broadcast-id", "snippet", "status")
if err != nil {
    var notFound *core.NotFoundError
    if errors.As(err, &notFound) {
        log.Println("Broadcast not found")
        return
    }
    log.Fatal(err)
}

fmt.Printf("Title: %s\n", broadcast.Snippet.Title)
fmt.Printf("Status: %s\n", broadcast.Status.LifeCycleStatus)
```

## GetMyActiveBroadcast

Retrieve the authenticated user's currently active broadcast.

```go
broadcast, err := streaming.GetMyActiveBroadcast(ctx, client, "snippet", "status")
if err != nil {
    log.Fatal(err)
}

if broadcast.IsLive() {
    fmt.Printf("Currently streaming: %s\n", broadcast.Snippet.Title)
}
```

## GetBroadcastLiveChatID

Get the live chat ID for a broadcast (useful for chat operations).

```go
liveChatID, err := streaming.GetBroadcastLiveChatID(ctx, client, "broadcast-id")
if err != nil {
    log.Fatal(err)
}

// Use liveChatID with LiveChatPoller or ChatBotClient
```

## Common Errors

| Error | Description |
|-------|-------------|
| `NotFoundError` | The broadcast doesn't exist |
| `AuthError` | Missing or invalid OAuth token |
| `ForbiddenError` | No access to this broadcast |
