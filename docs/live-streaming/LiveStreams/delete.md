---
layout: default
title: DeleteStream
description: Delete a video stream
---

Deletes a video stream.

**Quota Cost:** 50 units

## DeleteStream

```go
err := streaming.DeleteStream(ctx, client, "stream-id")
if err != nil {
    var apiErr *core.APIError
    if errors.As(err, &apiErr) {
        switch apiErr.Code {
        case "liveStreamBound":
            log.Println("Unbind the stream from its broadcast first")
        case "liveStreamActive":
            log.Println("Stop streaming video first")
        default:
            log.Printf("Error: %v", err)
        }
    }
    return
}

fmt.Println("Stream deleted successfully")
```

## Unbind and Delete

Streams bound to broadcasts cannot be deleted. Unbind first:

```go
// Check if stream is bound to any broadcast
broadcasts, _ := streaming.GetBroadcasts(ctx, client, &streaming.GetBroadcastsParams{
    Mine:  true,
    Parts: []string{"contentDetails"},
})

for _, b := range broadcasts.Items {
    if b.BoundStreamID() == streamID {
        // Unbind the stream
        _, err := streaming.BindBroadcast(ctx, client, &streaming.BindBroadcastParams{
            BroadcastID: b.ID,
            // StreamID omitted = unbind
        })
        if err != nil {
            log.Fatal(err)
        }
    }
}

// Now delete the stream
err := streaming.DeleteStream(ctx, client, streamID)
```

## Rules

- **Cannot delete bound streams**: Unbind from broadcasts first
- **Cannot delete active streams**: Stop sending video first
- **Permanent**: Stream key will be invalidated

## When to Delete Streams

- **Compromised stream key**: If your stream key was leaked
- **Cleanup**: Remove unused or test streams
- **Organization**: Clean up temporary streams

## Preserving Stream Keys

To keep using the same stream key, don't delete the stream:

- Create reusable streams (`IsReusable: true`)
- Unbind from broadcasts when done instead of deleting

## Common Errors

| Error | Description |
|-------|-------------|
| `NotFoundError` | Stream doesn't exist |
| `liveStreamBound` | Stream is bound to a broadcast |
| `liveStreamActive` | Stream is actively receiving video |
