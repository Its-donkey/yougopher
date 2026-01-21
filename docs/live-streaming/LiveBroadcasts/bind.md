---
layout: default
title: LiveBroadcasts.bind
description: Binds a YouTube broadcast to a stream or removes an existing binding
---

Binds a YouTube broadcast to a stream or removes an existing binding between a broadcast and a stream.

## Request

### HTTP Request

```
POST https://www.googleapis.com/youtube/v3/liveBroadcasts/bind
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `id` | Yes | string | The ID of the broadcast to bind. |
| `part` | Yes | string | Comma-separated list of resource parts to include in the response. |
| `streamId` | No | string | The ID of the stream to bind. Omit to unbind the current stream. |

### Authorization

Requires OAuth 2.0 authorization with the following scope:

- `https://www.googleapis.com/auth/youtube.force-ssl`

### Request Body

Do not provide a request body when calling this method.

## Response

If successful, this method returns the updated liveBroadcast resource with the binding information.

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `idRequired` | The broadcast ID is required. |
| 400 | `invalidStreamId` | The stream ID is invalid. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | The user cannot bind this broadcast. |
| 403 | `liveBroadcastAlreadyBound` | The broadcast is already bound to a stream. |
| 403 | `streamAlreadyBound` | The stream is already bound to another broadcast. |
| 404 | `liveBroadcastNotFound` | The broadcast does not exist. |
| 404 | `liveStreamNotFound` | The stream does not exist. |

## Quota Cost

This method consumes **50 quota units**.

---

## Yougopher Implementation

### BindBroadcast

Bind a stream to a broadcast.

```go
bound, err := streaming.BindBroadcast(ctx, client, &streaming.BindBroadcastParams{
    BroadcastID: "broadcast-id",
    StreamID:    "stream-id",
    Parts:       []string{"snippet", "contentDetails"},
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Bound stream %s to broadcast %s\n", bound.BoundStreamID(), bound.ID)
```

### Unbind a Stream

To unbind a stream, omit the `StreamID` parameter.

```go
unbound, err := streaming.BindBroadcast(ctx, client, &streaming.BindBroadcastParams{
    BroadcastID: "broadcast-id",
    // StreamID omitted = unbind
    Parts:       []string{"snippet", "contentDetails"},
})
if err != nil {
    log.Fatal(err)
}

if !unbound.HasBoundStream() {
    fmt.Println("Stream unbound successfully")
}
```

### Check Binding Status

```go
broadcast, _ := streaming.GetBroadcast(ctx, client, broadcastID, "contentDetails")

if broadcast.HasBoundStream() {
    fmt.Printf("Bound to stream: %s\n", broadcast.BoundStreamID())
} else {
    fmt.Println("No stream bound")
}
```

### Binding Requirements

1. **Stream must exist**: The stream ID must refer to an existing stream owned by the same channel.
2. **One-to-one relationship**: A stream can only be bound to one broadcast at a time.
3. **Required for going live**: A broadcast must have a bound stream to transition to testing or live state.
4. **Stream must be active**: The bound stream must be actively receiving video before transitioning to testing.
