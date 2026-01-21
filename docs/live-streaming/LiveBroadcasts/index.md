---
layout: default
title: LiveBroadcasts
description: Live broadcast management for YouTube streaming
---

A `LiveBroadcast` represents an event that will be streamed, via live video, on YouTube.

## Functions

| Function | Quota | Description |
|----------|-------|-------------|
| [GetBroadcasts](list) | 5 | Returns broadcasts matching parameters |
| [GetBroadcast](list) | 5 | Returns a single broadcast by ID |
| [GetMyActiveBroadcast](list) | 5 | Returns your currently active broadcast |
| [InsertBroadcast](insert) | 50 | Creates a new broadcast |
| [UpdateBroadcast](update) | 50 | Updates an existing broadcast |
| [DeleteBroadcast](delete) | 50 | Deletes a broadcast |
| [BindBroadcast](bind) | 50 | Binds/unbinds a stream to a broadcast |
| [TransitionBroadcast](transition) | 50 | Changes broadcast lifecycle status |
| [InsertCuepoint](cuepoint) | 50 | Inserts an ad break |

## Type Definition

```go
type LiveBroadcast struct {
    Kind           string                   `json:"kind,omitempty"`
    ETag           string                   `json:"etag,omitempty"`
    ID             string                   `json:"id,omitempty"`
    Snippet        *BroadcastSnippet        `json:"snippet,omitempty"`
    Status         *BroadcastStatus         `json:"status,omitempty"`
    ContentDetails *BroadcastContentDetails `json:"contentDetails,omitempty"`
    Statistics     *BroadcastStatistics     `json:"statistics,omitempty"`
    MonetizationDetails *MonetizationDetails `json:"monetizationDetails,omitempty"`
}
```

## Helper Methods

```go
broadcast.IsLive()         // Returns true if currently live
broadcast.IsComplete()     // Returns true if broadcast has ended
broadcast.IsUpcoming()     // Returns true if not started yet
broadcast.IsTesting()      // Returns true if in testing state
broadcast.LiveChatID()     // Returns the live chat ID
broadcast.BoundStreamID()  // Returns the bound stream ID
broadcast.HasBoundStream() // Returns true if a stream is bound
```

## Broadcast Lifecycle

```
created → ready → testing → live → complete
```

| Status | Description |
|--------|-------------|
| `created` | Broadcast was created but not ready |
| `ready` | Ready for testing (stream bound) |
| `testing` | Preview mode - only you can see |
| `testStarting` | Transitioning to testing |
| `live` | Public broadcast - viewers can watch |
| `liveStarting` | Transitioning to live |
| `complete` | Broadcast has ended |

## Quick Example

```go
import "github.com/Its-donkey/yougopher/youtube/streaming"

// Create a broadcast
startTime := time.Now().Add(1 * time.Hour)
broadcast, err := streaming.InsertBroadcast(ctx, client, &streaming.LiveBroadcast{
    Snippet: &streaming.BroadcastSnippet{
        Title:              "My Live Stream",
        ScheduledStartTime: &startTime,
    },
    Status: &streaming.BroadcastStatus{
        PrivacyStatus: "unlisted",
    },
}, "snippet", "status")

// Bind a stream
bound, err := streaming.BindBroadcast(ctx, client, &streaming.BindBroadcastParams{
    BroadcastID: broadcast.ID,
    StreamID:    "stream-id",
})

// Go live
live, err := streaming.TransitionBroadcast(ctx, client, broadcast.ID, streaming.TransitionLive)
if live.IsLive() {
    fmt.Println("You are now live!")
}
```
