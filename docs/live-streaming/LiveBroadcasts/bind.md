---
layout: default
title: BindBroadcast
description: Bind or unbind a stream to a broadcast
---

Binds a YouTube broadcast to a stream or removes an existing binding.

**Quota Cost:** 50 units

## BindBroadcast

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

## Unbind a Stream

Omit the `StreamID` parameter to unbind:

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

## Check Binding Status

```go
broadcast, _ := streaming.GetBroadcast(ctx, client, broadcastID, "contentDetails")

if broadcast.HasBoundStream() {
    fmt.Printf("Bound to stream: %s\n", broadcast.BoundStreamID())
} else {
    fmt.Println("No stream bound")
}
```

## Requirements

1. **Stream must exist**: The stream ID must refer to an existing stream owned by the same channel
2. **One-to-one relationship**: A stream can only be bound to one broadcast at a time
3. **Required for going live**: A broadcast must have a bound stream to transition to testing or live
4. **Stream must be active**: The bound stream must be receiving video before transitioning to testing

## Common Errors

| Error | Description |
|-------|-------------|
| `liveStreamNotFound` | Stream doesn't exist |
| `liveBroadcastAlreadyBound` | Broadcast already has a stream |
| `streamAlreadyBound` | Stream is bound to another broadcast |
