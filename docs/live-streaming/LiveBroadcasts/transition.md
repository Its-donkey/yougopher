---
layout: default
title: LiveBroadcasts.transition
description: Transitions a YouTube live broadcast to a different lifecycle status
---

Transitions a YouTube live broadcast to a different lifecycle status.

## Request

### HTTP Request

```
POST https://www.googleapis.com/youtube/v3/liveBroadcasts/transition
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `id` | Yes | string | The ID of the broadcast to transition. |
| `broadcastStatus` | Yes | string | The target status: `testing`, `live`, or `complete`. |
| `part` | Yes | string | Comma-separated list of resource parts to include in the response. |

### Authorization

Requires OAuth 2.0 authorization with the following scope:

- `https://www.googleapis.com/auth/youtube.force-ssl`

### Request Body

Do not provide a request body when calling this method.

## Broadcast Lifecycle

```
created → ready → testing → live → complete
```

| Status | Description |
|--------|-------------|
| `created` | Broadcast was created but is not ready. |
| `ready` | Broadcast is ready for testing. |
| `testing` | Preview mode - only you can see the stream. |
| `testStarting` | Transitioning to testing (asynchronous). |
| `live` | Public broadcast - viewers can watch. |
| `liveStarting` | Transitioning to live (asynchronous). |
| `complete` | Broadcast has ended. |

## Response

If successful, this method returns the updated liveBroadcast resource.

Note: The transition may be asynchronous. The response may show `testStarting` or `liveStarting` status. Poll the broadcast status until it reaches the target state.

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `idRequired` | The broadcast ID is required. |
| 400 | `invalidTransition` | The requested transition is not valid from the current state. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | The user cannot transition this broadcast. |
| 403 | `redundantTransition` | The broadcast is already in the requested state. |
| 403 | `streamNotActive` | Cannot start testing without an active stream. |
| 403 | `streamNotBound` | No stream is bound to this broadcast. |
| 404 | `liveBroadcastNotFound` | The broadcast does not exist. |

## Quota Cost

This method consumes **50 quota units**.

---

## Yougopher Implementation

### TransitionBroadcast

Transition a broadcast to a new lifecycle status.

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

### Transition Constants

```go
streaming.TransitionTesting  // "testing" - Start preview mode
streaming.TransitionLive     // "live" - Go public
streaming.TransitionComplete // "complete" - End broadcast
```

### Start Testing

Transition to preview mode. Requires:
- A stream bound to the broadcast
- The stream actively receiving video data

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
}
```

### Go Live

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

### End Broadcast

End the broadcast and stop streaming.

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

### Valid Transitions

| From | To | Description |
|------|----|-------------|
| `ready` | `testing` | Start preview mode |
| `testing` | `live` | Go public |
| `testing` | `complete` | End without going live |
| `live` | `complete` | End the broadcast |

Invalid transitions (e.g., `live` → `testing`) will return an error.

### Handling Asynchronous Transitions

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
