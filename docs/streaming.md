---
layout: default
title: Streaming API
description: Build YouTube Live chat bots with real-time message handling and moderation.
---

## Overview

Build interactive chat bots for YouTube Live streams:

**ChatBotClient (Recommended):** High-level bot with semantic events
- Receive typed events (messages, Super Chats, memberships)
- Automatic token refresh and reconnection
- Built-in panic recovery for handlers

**LiveChatPoller (Advanced):** Low-level polling control
- Direct access to raw API responses
- Custom polling intervals and backoff
- Fine-grained control over message processing

**LiveChatStream (SSE):** Server-Sent Events streaming
- Lower latency than polling
- Real-time message delivery
- Automatic reconnection with resumption

**Moderation:** Full chat moderation support
- Ban/timeout users
- Delete messages
- Add/remove moderators

**Broadcasts:** Live broadcast management
- Get broadcast information
- Retrieve live chat IDs
- Check stream status (live, upcoming, complete)

## Prerequisites

- **Read chat:** `youtube` scope
- **Moderate chat:** `youtube.force-ssl` scope

## ChatBotClient

### NewChatBotClient

Create a new high-level chat bot client.

```go
bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
if err != nil {
    log.Fatal(err)
}
```

### Connect

Start the bot and begin listening for messages.

```go
err := bot.Connect(ctx)
if err != nil {
    log.Fatal(err)
}
defer bot.Close()
```

### OnMessage

Register a handler for chat messages.

```go
unsub := bot.OnMessage(func(msg *streaming.ChatMessage) {
    fmt.Printf("[%s] %s\n", msg.Author.DisplayName, msg.Message)
})
// Later: unsub() to unregister
```

**ChatMessage fields:**
```go
type ChatMessage struct {
    ID          string
    Message     string
    Author      *Author
    PublishedAt time.Time
    Raw         *LiveChatMessage
}
```

### OnSuperChat

Register a handler for Super Chat donations.

```go
bot.OnSuperChat(func(event *streaming.SuperChatEvent) {
    fmt.Printf("Super Chat from %s: %s - %s\n",
        event.Author.DisplayName, event.Amount, event.Message)
})
```

**SuperChatEvent fields:**
```go
type SuperChatEvent struct {
    ID           string
    Author       *Author
    Message      string
    Amount       string    // e.g., "$5.00"
    AmountMicros int64     // e.g., 5000000
    Currency     string    // e.g., "USD"
    Tier         int       // 1-7
    Raw          *LiveChatMessage
}
```

### OnSuperSticker

Register a handler for Super Sticker events.

```go
bot.OnSuperSticker(func(event *streaming.SuperStickerEvent) {
    fmt.Printf("Super Sticker from %s: %s (%s)\n",
        event.Author.DisplayName, event.AltText, event.Amount)
})
```

### OnMembership

Register a handler for new channel memberships.

```go
bot.OnMembership(func(event *streaming.MembershipEvent) {
    fmt.Printf("New member: %s joined at %s level\n",
        event.Author.DisplayName, event.LevelName)
})
```

### OnMemberMilestone

Register a handler for member milestone celebrations.

```go
bot.OnMemberMilestone(func(event *streaming.MemberMilestoneEvent) {
    fmt.Printf("%s has been a member for %d months!\n",
        event.Author.DisplayName, event.Months)
})
```

### OnGiftMembership

Register a handler for gifted memberships.

```go
bot.OnGiftMembership(func(event *streaming.GiftMembershipEvent) {
    fmt.Printf("%s gifted %d memberships!\n",
        event.Author.DisplayName, event.Count)
})
```

### OnMessageDeleted

Register a handler for message deletions.

```go
bot.OnMessageDeleted(func(messageID string) {
    fmt.Printf("Message deleted: %s\n", messageID)
})
```

### OnUserBanned

Register a handler for user bans.

```go
bot.OnUserBanned(func(event *streaming.BanEvent) {
    fmt.Printf("User banned: %s (%s)\n",
        event.BannedUser.DisplayName, event.BanType)
})
```

### OnConnect / OnDisconnect

Register handlers for connection state changes.

```go
bot.OnConnect(func() {
    fmt.Println("Bot connected!")
})

bot.OnDisconnect(func() {
    fmt.Println("Bot disconnected")
})
```

### OnError

Register a handler for errors.

```go
bot.OnError(func(err error) {
    log.Printf("Bot error: %v", err)
})
```

## Sending Messages

### Say

Send a message to the chat.

**Requires:** youtube.force-ssl scope

```go
err := bot.Say(ctx, "Hello, chat!")
if err != nil {
    log.Printf("Failed to send message: %v", err)
}
```

## Moderation

### Delete

Delete a message from the chat.

**Requires:** youtube.force-ssl scope

```go
err := bot.Delete(ctx, messageID)
if err != nil {
    log.Printf("Failed to delete message: %v", err)
}
```

### Ban

Permanently ban a user from the chat.

**Requires:** youtube.force-ssl scope

```go
err := bot.Ban(ctx, channelID)
if err != nil {
    log.Printf("Failed to ban user: %v", err)
}
```

### Timeout

Temporarily ban a user from the chat.

**Requires:** youtube.force-ssl scope

```go
err := bot.Timeout(ctx, channelID, 300) // 5 minute timeout
if err != nil {
    log.Printf("Failed to timeout user: %v", err)
}
```

### Unban

Remove a ban from a user.

**Requires:** youtube.force-ssl scope

```go
err := bot.Unban(ctx, banID)
if err != nil {
    log.Printf("Failed to unban user: %v", err)
}
```

### AddModerator

Add a moderator to the chat.

**Requires:** youtube.force-ssl scope

```go
err := bot.AddModerator(ctx, channelID)
if err != nil {
    log.Printf("Failed to add moderator: %v", err)
}
```

### RemoveModerator

Remove a moderator from the chat.

**Requires:** youtube.force-ssl scope

```go
err := bot.RemoveModerator(ctx, moderatorID)
if err != nil {
    log.Printf("Failed to remove moderator: %v", err)
}
```

## LiveChatPoller (Advanced)

For more control over polling behavior, use LiveChatPoller directly.

### NewLiveChatPoller

Create a new poller with optional configuration.

```go
poller := streaming.NewLiveChatPoller(client, liveChatID,
    streaming.WithMinPollInterval(2*time.Second),
    streaming.WithMaxPollInterval(30*time.Second),
    streaming.WithBackoff(&core.BackoffConfig{
        InitialDelay: time.Second,
        MaxDelay:     30 * time.Second,
        Multiplier:   2.0,
    }),
)
```

### Start / Stop

Control the polling lifecycle.

```go
err := poller.Start(ctx)
if err != nil {
    log.Fatal(err)
}

// Later...
poller.Stop()
```

### OnMessage (Poller)

Handle raw message events with full API data.

```go
poller.OnMessage(func(msg *streaming.LiveChatMessage) {
    fmt.Printf("Type: %s, Message: %s\n", msg.Type(), msg.Message())
})
```

### Reset

Reset polling state for reuse (must be stopped).

```go
err := poller.Reset()
if err != nil {
    log.Printf("Cannot reset: %v", err) // Returns ErrAlreadyRunning if running
}
```

## LiveChatStream (SSE Streaming)

For real-time chat messages with lower latency than polling, use SSE streaming via `liveChatMessages.streamList`.

### NewLiveChatStream

Create a new SSE-based live chat stream.

```go
stream := streaming.NewLiveChatStream(client, liveChatID,
    streaming.WithStreamAccessToken(accessToken),
    streaming.WithStreamMaxResults(1000),      // 200-2000
    streaming.WithStreamProfileImageSize(240), // 16-720 pixels
)
```

### Start / Stop

Control the stream lifecycle.

```go
err := stream.Start(ctx)
if err != nil {
    log.Fatal(err)
}

// Later...
stream.Stop()
```

### OnMessage (Stream)

Handle real-time messages from the SSE stream.

```go
stream.OnMessage(func(msg *streaming.LiveChatMessage) {
    fmt.Printf("[%s] %s\n",
        msg.AuthorDetails.DisplayName,
        msg.Snippet.DisplayMessage)
})
```

### OnResponse

Handle full API responses (includes metadata like offlineAt).

```go
stream.OnResponse(func(resp *streaming.LiveChatMessageListResponse) {
    if resp.IsChatEnded() {
        fmt.Println("Chat has ended")
    }
    fmt.Printf("Received %d messages\n", len(resp.Items))
})
```

### OnConnect / OnDisconnect (Stream)

Handle connection state changes.

```go
stream.OnConnect(func() {
    fmt.Println("Stream connected")
})

stream.OnDisconnect(func() {
    fmt.Println("Stream disconnected")
})
```

### OnError (Stream)

Handle stream errors.

```go
stream.OnError(func(err error) {
    log.Printf("Stream error: %v", err)
})
```

### Stream Resumption

Use page tokens to resume from where you left off after disconnection.

```go
// Save the page token before stopping
token := stream.PageToken()

// Later, resume from that position
stream.SetPageToken(token)
stream.Start(ctx)
```

### SSE vs Polling

| Feature | SSE (LiveChatStream) | Polling (LiveChatPoller) |
|---------|---------------------|-------------------------|
| Latency | Lower (~instant) | Higher (poll interval) |
| Complexity | Simple connection | Interval management |
| Reconnection | Automatic with token | Manual token handling |
| Use case | Real-time bots | Custom polling logic |

## Author Information

All event types include author details:

```go
type Author struct {
    ChannelID       string
    DisplayName     string
    ProfileImageURL string
    IsModerator     bool
    IsOwner         bool
    IsMember        bool
    IsVerified      bool
}
```

## Handler Unsubscription

All `On*` methods return an unsubscribe function:

```go
unsub := bot.OnMessage(handler)

// Later, to stop receiving events:
unsub() // Safe to call multiple times
```

## Thread Safety

Both `ChatBotClient` and `LiveChatPoller` are safe for concurrent use. Handler registration and unsubscription can be done from any goroutine.

## Error Handling

Handlers are protected with panic recovery. If a handler panics, the error is dispatched to error handlers and polling continues normally.

## Broadcasts

Retrieve live broadcast information to get live chat IDs and stream status.

### GetBroadcast

Get a specific broadcast by ID.

```go
broadcast, err := streaming.GetBroadcast(ctx, client, "broadcast-id")
if err != nil {
    log.Fatal(err)
}

if broadcast.IsLive() {
    fmt.Println("Currently streaming!")
    liveChatID := broadcast.LiveChatID()
}
```

### GetMyActiveBroadcast

Get the authenticated user's currently active broadcast.

**Requires:** youtube.readonly or youtube scope

```go
myBroadcast, err := streaming.GetMyActiveBroadcast(ctx, client)
if err != nil {
    var notFound *core.NotFoundError
    if errors.As(err, &notFound) {
        fmt.Println("No active broadcast")
        return
    }
    log.Fatal(err)
}

fmt.Printf("Streaming: %s\n", myBroadcast.Snippet.Title)
```

### GetBroadcastLiveChatID

Get the live chat ID directly from a broadcast.

```go
liveChatID, err := streaming.GetBroadcastLiveChatID(ctx, client, "broadcast-id")
if err != nil {
    log.Fatal(err)
}

// Use the live chat ID to create a bot
bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
```

### Broadcast Status

Check broadcast status with helper methods:

```go
if broadcast.IsLive() {
    // Currently streaming
}

if broadcast.IsUpcoming() {
    // Scheduled but not started
}

if broadcast.IsComplete() {
    // Stream has ended
}
```

### GetBroadcasts

Get multiple broadcasts with filtering options.

```go
resp, err := streaming.GetBroadcasts(ctx, client, &streaming.GetBroadcastsParams{
    BroadcastStatus: "active",  // "active", "all", "completed", "upcoming"
    Mine:            true,
    MaxResults:      10,
})

for _, broadcast := range resp.Items {
    fmt.Printf("%s (%s)\n", broadcast.Snippet.Title, broadcast.Status.LifeCycleStatus)
}
```

## Broadcast Management

Create and manage live broadcasts programmatically.

### InsertBroadcast

Create a new live broadcast.

**Requires:** youtube.force-ssl scope
**Quota cost:** 50 units

```go
broadcast := &streaming.LiveBroadcast{
    Snippet: &streaming.BroadcastSnippet{
        Title:              "My Live Stream",
        Description:        "Stream description",
        ScheduledStartTime: &startTime,
    },
    Status: &streaming.BroadcastStatus{
        PrivacyStatus: "unlisted",
    },
    ContentDetails: &streaming.BroadcastContentDetails{
        EnableDvr:       true,
        EnableEmbed:     true,
        RecordFromStart: true,
    },
}

created, err := streaming.InsertBroadcast(ctx, client, broadcast,
    "snippet", "status", "contentDetails")
```

### UpdateBroadcast

Update an existing broadcast.

**Quota cost:** 50 units

```go
broadcast.Snippet.Title = "Updated Title"
updated, err := streaming.UpdateBroadcast(ctx, client, broadcast, "snippet")
```

### DeleteBroadcast

Delete a broadcast.

**Quota cost:** 50 units

```go
err := streaming.DeleteBroadcast(ctx, client, broadcastID)
```

### BindBroadcast

Bind a stream to a broadcast.

**Quota cost:** 50 units

```go
bound, err := streaming.BindBroadcast(ctx, client, &streaming.BindBroadcastParams{
    BroadcastID: broadcastID,
    StreamID:    streamID,
})
```

### TransitionBroadcast

Transition a broadcast to a new state.

**Quota cost:** 50 units

```go
// Start testing
broadcast, err := streaming.TransitionBroadcast(ctx, client, broadcastID, streaming.TransitionTesting)

// Go live
broadcast, err = streaming.TransitionBroadcast(ctx, client, broadcastID, streaming.TransitionLive)

// End broadcast
broadcast, err = streaming.TransitionBroadcast(ctx, client, broadcastID, streaming.TransitionComplete)
```

**Valid transitions:**
- `testing`: Start testing (requires bound and active stream)
- `live`: Go live (from testing state)
- `complete`: End the broadcast

## Stream Management

Create and manage live streams (the video feed sent to YouTube).

### InsertStream

Create a new live stream.

**Requires:** youtube.force-ssl scope
**Quota cost:** 50 units

```go
stream := &streaming.LiveStream{
    Snippet: &streaming.StreamSnippet{
        Title:       "Primary Stream",
        Description: "Main RTMP ingest",
    },
    CDN: &streaming.StreamCDN{
        Resolution:    "1080p",
        FrameRate:     "30fps",
        IngestionType: "rtmp",
    },
}

created, err := streaming.InsertStream(ctx, client, stream, "snippet", "cdn")
```

### GetStream

Retrieve stream information.

**Quota cost:** 5 units

```go
stream, err := streaming.GetStream(ctx, client, streamID, "snippet", "cdn", "status")
```

### GetStreamKey

Get the stream key for OBS/streaming software.

```go
stream, err := streaming.GetStream(ctx, client, streamID, "cdn")
if err != nil {
    log.Fatal(err)
}

streamKey := stream.StreamKey()      // "stream-key-123"
rtmpURL := stream.RTMPUrl()          // "rtmp://a.rtmp.youtube.com/live2"
rtmpsURL := stream.RTMPSUrl()        // "rtmps://a.rtmps.youtube.com/live2"
```

### UpdateStream

Update an existing stream.

**Quota cost:** 50 units

```go
stream.Snippet.Title = "Updated Stream"
updated, err := streaming.UpdateStream(ctx, client, stream, "snippet")
```

### DeleteStream

Delete a stream.

**Quota cost:** 50 units

```go
err := streaming.DeleteStream(ctx, client, streamID)
```

### Stream Health

Check stream health status.

```go
stream, err := streaming.GetStream(ctx, client, streamID, "status")

if stream.IsActive() {
    fmt.Println("Stream is receiving data")
}

if stream.IsHealthy() {
    fmt.Println("Stream health is good")
}

if stream.HasConfigurationIssues() {
    for _, issue := range stream.Status.HealthStatus.ConfigurationIssues {
        fmt.Printf("Issue: %s - %s\n", issue.Type, issue.Description)
    }
}
```

## StreamController (High-Level)

For common streaming workflows, use StreamController.

### NewStreamController

Create a high-level stream controller.

```go
controller, err := streaming.NewStreamController(client, authClient)
```

### CreateBroadcastWithStream

Create a broadcast and stream in one call, automatically binding them together.

**Quota cost:** 150 units (insert broadcast + insert stream + bind)

```go
result, err := controller.CreateBroadcastWithStream(ctx, &streaming.CreateBroadcastParams{
    Title:              "My Stream",
    Description:        "Stream description",
    PrivacyStatus:      "unlisted",
    Resolution:         "1080p",
    FrameRate:          "30fps",
    EnableDVR:          true,
    EnableAutoStart:    true,
    LatencyPreference:  "low",
})

// Get stream key for OBS
fmt.Printf("RTMP URL: %s\n", result.Stream.RTMPUrl())
fmt.Printf("Stream Key: %s\n", result.Stream.StreamKey())
```

### Complete Streaming Workflow

```go
// 1. Create broadcast with stream
result, err := controller.CreateBroadcastWithStream(ctx, &streaming.CreateBroadcastParams{
    Title:         "My Live Stream",
    PrivacyStatus: "unlisted",
})

// 2. Configure OBS/encoder with stream key
fmt.Printf("RTMP URL: %s\n", result.Stream.RTMPUrl())
fmt.Printf("Stream Key: %s\n", result.Stream.StreamKey())

// 3. Start streaming in OBS, then start testing
broadcast, err := controller.StartTesting(ctx, result.Broadcast.ID)

// 4. Verify everything looks good, then go live
broadcast, err = controller.GoLive(ctx, result.Broadcast.ID)

// 5. When done, end the broadcast
broadcast, err = controller.EndBroadcast(ctx, result.Broadcast.ID)
```

### GetStreamHealth

Check stream health before going live.

```go
health, err := controller.GetStreamHealth(ctx, streamID)
if health.Status == streaming.StreamHealthGood {
    fmt.Println("Stream is healthy, ready to go live")
}
```
