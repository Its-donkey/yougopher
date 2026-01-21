---
layout: default
title: Live Streaming
description: YouTube Live Streaming API with Yougopher
---

The `youtube/streaming` package provides a complete implementation for YouTube Live Streaming.

## Resources

| Resource | Description |
|----------|-------------|
| [LiveBroadcasts](LiveBroadcasts/) | Manage broadcasts - create, update, delete, bind streams, go live |
| [LiveStreams](LiveStreams/) | Manage video streams - create, configure, get stream keys |
| [LiveChatMessages](LiveChatMessages/) | Read and send chat messages with polling or SSE |
| [LiveChatBans](LiveChatBans/) | Ban and timeout users in live chat |
| [LiveChatModerators](LiveChatModerators/) | Add and remove chat moderators |
| [SuperChatEvents](SuperChatEvents/) | Access Super Chat and Super Sticker data |

## Key Types

| Type | Description |
|------|-------------|
| `LiveBroadcast` | A live streaming event |
| `LiveStream` | Video stream with RTMP credentials |
| `LiveChatMessage` | Chat message with author details |
| `LiveChatPoller` | Automatic chat polling client |
| `LiveChatStream` | SSE streaming chat client |
| `ChatBotClient` | High-level chat bot |
| `StreamController` | High-level broadcast manager |

## Quick Start

```go
import (
    "github.com/Its-donkey/yougopher/youtube/core"
    "github.com/Its-donkey/yougopher/youtube/streaming"
)

// Create authenticated client
client := core.NewClient()
client.SetAccessToken("your-oauth-token")
```

## Complete Broadcast Example

```go
// 1. Create a broadcast
startTime := time.Now().Add(1 * time.Hour)
broadcast, _ := streaming.InsertBroadcast(ctx, client, &streaming.LiveBroadcast{
    Snippet: &streaming.BroadcastSnippet{
        Title:              "My Live Stream",
        ScheduledStartTime: &startTime,
    },
    Status: &streaming.BroadcastStatus{
        PrivacyStatus: "unlisted",
    },
}, "snippet", "status")

// 2. Create a stream
stream, _ := streaming.InsertStream(ctx, client, &streaming.LiveStream{
    Snippet: &streaming.StreamSnippet{Title: "Primary Stream"},
    CDN: &streaming.StreamCDN{
        IngestionType: "rtmp",
        Resolution:    "1080p",
        FrameRate:     "30fps",
    },
}, "snippet", "cdn")

fmt.Printf("OBS Server: %s\n", stream.RTMPSUrl())
fmt.Printf("Stream Key: %s\n", stream.StreamKey())

// 3. Bind stream to broadcast
streaming.BindBroadcast(ctx, client, &streaming.BindBroadcastParams{
    BroadcastID: broadcast.ID,
    StreamID:    stream.ID,
})

// 4. Start streaming video in OBS, then go live
streaming.TransitionBroadcast(ctx, client, broadcast.ID, streaming.TransitionTesting)
streaming.TransitionBroadcast(ctx, client, broadcast.ID, streaming.TransitionLive)
```

## Chat Bot Example

```go
// Create a poller for the live chat
poller := streaming.NewLiveChatPoller(client, broadcast.LiveChatID())

// Handle messages
poller.OnMessage(func(msg *streaming.LiveChatMessage) {
    fmt.Printf("[%s] %s\n", msg.AuthorDetails.DisplayName, msg.Message())

    // Handle commands
    if strings.HasPrefix(msg.Message(), "!hello") {
        poller.SendMessage(ctx, fmt.Sprintf("Hello, %s!", msg.AuthorDetails.DisplayName))
    }

    // Handle Super Chats
    if msg.IsSuperChat() {
        fmt.Printf("SUPER CHAT: %s\n", msg.Snippet.SuperChatDetails.AmountDisplayString)
    }
})

// Handle bans
poller.OnBan(func(details *streaming.UserBannedDetails) {
    fmt.Printf("User banned: %s\n", details.BannedUserDetails.DisplayName)
})

// Start polling
poller.Start(ctx)
defer poller.Stop()
```

## Quota Costs

| Operation | Cost |
|-----------|------|
| List broadcasts/streams/chat | 5 |
| Insert/update/delete broadcast | 50 |
| Insert/update/delete stream | 50 |
| Bind/transition broadcast | 50 |
| Send/delete chat message | 50 |
| Ban/unban user | 50 |
| Add/remove moderator | 50 |
