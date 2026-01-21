---
layout: default
title: InsertBroadcast
description: Create a new YouTube live broadcast
---

Creates a new YouTube live broadcast.

**Quota Cost:** 50 units

## InsertBroadcast

```go
startTime := time.Now().Add(1 * time.Hour)

broadcast := &streaming.LiveBroadcast{
    Snippet: &streaming.BroadcastSnippet{
        Title:              "My Live Stream",
        Description:        "Streaming live from my channel",
        ScheduledStartTime: &startTime,
    },
    Status: &streaming.BroadcastStatus{
        PrivacyStatus: "unlisted",
    },
    ContentDetails: &streaming.BroadcastContentDetails{
        EnableDVR:         true,
        EnableEmbed:       true,
        EnableAutoStart:   true,
        EnableAutoStop:    true,
        LatencyPreference: "normal",
    },
}

created, err := streaming.InsertBroadcast(ctx, client, broadcast, "snippet", "status", "contentDetails")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created broadcast: %s\n", created.ID)
fmt.Printf("Live chat ID: %s\n", created.LiveChatID())
```

## Required Fields

| Field | Description |
|-------|-------------|
| `Snippet.Title` | The broadcast's title (max 100 chars) |
| `Snippet.ScheduledStartTime` | When the broadcast is scheduled to start |
| `Status.PrivacyStatus` | `private`, `public`, or `unlisted` |

## Privacy Options

| Value | Description |
|-------|-------------|
| `private` | Only you can view the broadcast |
| `unlisted` | Anyone with the link can view |
| `public` | Anyone can find and view |

## Content Details Options

| Field | Default | Description |
|-------|---------|-------------|
| `EnableDVR` | true | Viewers can pause/rewind |
| `EnableEmbed` | true | Allow embedding in other sites |
| `EnableAutoStart` | false | Auto-start when stream is active |
| `EnableAutoStop` | false | Auto-stop when stream ends |
| `EnableClosedCaptions` | false | Enable closed captions |
| `LatencyPreference` | `normal` | `normal`, `low`, or `ultraLow` |
| `Projection` | `rectangular` | `rectangular` or `360` |

## Latency Options

| Value | Delay | Notes |
|-------|-------|-------|
| `normal` | 15-30 seconds | Highest quality, most reliable |
| `low` | 5-10 seconds | Good balance of quality and interaction |
| `ultraLow` | 2-4 seconds | Best for real-time interaction, reduced quality |

## Made for Kids

```go
broadcast := &streaming.LiveBroadcast{
    Snippet: &streaming.BroadcastSnippet{
        Title:              "Kids Show",
        ScheduledStartTime: &startTime,
    },
    Status: &streaming.BroadcastStatus{
        PrivacyStatus:           "public",
        SelfDeclaredMadeForKids: true,
    },
}
```

## Common Errors

| Error | Description |
|-------|-------------|
| `invalidScheduledStartTime` | Start time is in the past |
| `liveStreamingNotEnabled` | Channel doesn't have live streaming enabled |
| `titleRequired` | Title is missing or empty |
