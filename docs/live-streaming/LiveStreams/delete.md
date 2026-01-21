---
layout: default
title: LiveStreams.delete
description: Deletes a video stream
---

Deletes a video stream.

## Request

### HTTP Request

```
DELETE https://www.googleapis.com/youtube/v3/liveStreams
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `id` | Yes | string | The ID of the stream to delete. |

### Authorization

Requires OAuth 2.0 authorization with the following scope:

- `https://www.googleapis.com/auth/youtube.force-ssl`

### Request Body

Do not provide a request body when calling this method.

## Response

If successful, this method returns an empty response body with HTTP status code 204.

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `idRequired` | The stream ID is required. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | The user cannot delete this stream. |
| 403 | `liveStreamBound` | The stream is bound to a broadcast and cannot be deleted. |
| 403 | `liveStreamActive` | The stream is active and cannot be deleted. |
| 404 | `liveStreamNotFound` | The stream does not exist. |

## Quota Cost

This method consumes **50 quota units**.

---

## Yougopher Implementation

### DeleteStream

Delete a stream by ID.

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

### Deletion Rules

- **Cannot delete bound streams**: Unbind from the broadcast first using `BindBroadcast` with no stream ID.
- **Cannot delete active streams**: Stop sending video to the stream first.
- **Permanent action**: This action cannot be undone. The stream key will be invalidated.

### Example: Unbind and Delete

```go
// Check if stream is bound
broadcast, _ := streaming.GetBroadcasts(ctx, client, &streaming.GetBroadcastsParams{
    Mine:  true,
    Parts: []string{"contentDetails"},
})

for _, b := range broadcast.Items {
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

### When to Delete Streams

- **Compromised stream key**: If your stream key was leaked, delete the stream and create a new one.
- **Cleanup unused streams**: Delete streams that are no longer needed.
- **Organization**: Remove test or temporary streams.

### Preserving Stream Keys

If you want to keep using the same stream key, don't delete the stream. Instead:
- Create reusable streams (`IsReusable: true`)
- Unbind from broadcasts when done instead of deleting
