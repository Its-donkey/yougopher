---
layout: default
title: TransitionBroadcast
description: Transition a broadcast to a different lifecycle status
---

Transitions a YouTube live broadcast to a different lifecycle status.

**Quota Cost:** 50 units

## TransitionBroadcast

```go
broadcast, err := streaming.TransitionBroadcast(ctx, client,
    "broadcast-id",
    streaming.TransitionLive,
    "snippet", "status",
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Broadcast status: %s\n", broadcast.Status.LifeCycleStatus)
```

## Transition Constants

```go
streaming.TransitionTesting  // Start preview mode
streaming.TransitionLive     // Go public
streaming.TransitionComplete // End broadcast
```

## Start Testing

Preview mode - only you can see the stream. Requires:
- A stream bound to the broadcast
- The stream actively receiving video

```go
broadcast, err := streaming.TransitionBroadcast(ctx, client,
    broadcastID,
    streaming.TransitionTesting,
)
if err != nil {
    var apiErr *core.APIError
    if errors.As(err, &apiErr) {
        switch {
        case apiErr.Code == "streamNotActive":
            log.Println("Start streaming video before testing")
        case apiErr.Code == "streamNotBound":
            log.Println("Bind a stream first")
        }
    }
    log.Fatal(err)
}
```

## Go Live

Transition from testing to public broadcast.

```go
broadcast, err := streaming.TransitionBroadcast(ctx, client,
    broadcastID,
    streaming.TransitionLive,
)
if err != nil {
    log.Fatal(err)
}

if broadcast.IsLive() {
    fmt.Println("You are now live!")
}
```

## End Broadcast

```go
broadcast, err := streaming.TransitionBroadcast(ctx, client,
    broadcastID,
    streaming.TransitionComplete,
)
if err != nil {
    log.Fatal(err)
}

if broadcast.IsComplete() {
    fmt.Println("Broadcast ended")
}
```

## Valid Transitions

| From | To | Description |
|------|----|-------------|
| `ready` | `testing` | Start preview mode |
| `testing` | `live` | Go public |
| `testing` | `complete` | End without going live |
| `live` | `complete` | End the broadcast |

Invalid transitions (e.g., `live` → `testing`) return an error.

## Handling Asynchronous Transitions

Transitions may not be instantaneous. Poll for completion:

```go
broadcast, _ := streaming.TransitionBroadcast(ctx, client, broadcastID, streaming.TransitionLive)

// Poll until transition completes
for broadcast.Status.LifeCycleStatus == "liveStarting" {
    time.Sleep(2 * time.Second)
    broadcast, _ = streaming.GetBroadcast(ctx, client, broadcastID, "status")
}

if broadcast.IsLive() {
    fmt.Println("Transition complete - you are live!")
}
```

## Common Errors

| Error | Description |
|-------|-------------|
| `invalidTransition` | Invalid state transition |
| `redundantTransition` | Already in requested state |
| `streamNotActive` | No active video stream |
| `streamNotBound` | No stream bound to broadcast |
