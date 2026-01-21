---
layout: default
title: LiveBroadcasts.delete
description: Deletes a YouTube live broadcast
---

Deletes a YouTube live broadcast.

## Request

### HTTP Request

```
DELETE https://www.googleapis.com/youtube/v3/liveBroadcasts
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `id` | Yes | string | The ID of the broadcast to delete. |

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
| 400 | `idRequired` | The broadcast ID is required. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | The user cannot delete this broadcast. |
| 403 | `liveBroadcastCannotBeDeleted` | The broadcast is live and cannot be deleted. |
| 404 | `liveBroadcastNotFound` | The broadcast does not exist. |

## Quota Cost

This method consumes **50 quota units**.

---

## Yougopher Implementation

### DeleteBroadcast

Delete a broadcast by ID.

```go
err := streaming.DeleteBroadcast(ctx, client, "broadcast-id")
if err != nil {
    var apiErr *core.APIError
    if errors.As(err, &apiErr) {
        if apiErr.IsNotFound() {
            log.Println("Broadcast already deleted")
        }
    }
    log.Fatal(err)
}

fmt.Println("Broadcast deleted successfully")
```

### Deletion Rules

- **Cannot delete live broadcasts**: You must end the broadcast first using `TransitionBroadcast` with `TransitionComplete`.
- **Deleting a bound stream**: Deleting a broadcast does not delete the associated stream. Use `DeleteStream` separately.
- **Permanent action**: This action cannot be undone.

### Example: End and Delete

```go
// First, end the broadcast if it's live
broadcast, _ := streaming.GetBroadcast(ctx, client, broadcastID, "status")
if broadcast.IsLive() {
    _, err := streaming.TransitionBroadcast(ctx, client, broadcastID, streaming.TransitionComplete)
    if err != nil {
        log.Fatal(err)
    }
}

// Now delete it
err := streaming.DeleteBroadcast(ctx, client, broadcastID)
```
