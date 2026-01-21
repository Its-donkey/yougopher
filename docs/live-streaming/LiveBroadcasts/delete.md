---
layout: default
title: DeleteBroadcast
description: Delete a YouTube live broadcast
---

Deletes a YouTube live broadcast.

**Quota Cost:** 50 units

## DeleteBroadcast

```go
err := streaming.DeleteBroadcast(ctx, client, "broadcast-id")
if err != nil {
    var apiErr *core.APIError
    if errors.As(err, &apiErr) {
        if apiErr.IsNotFound() {
            log.Println("Broadcast already deleted")
            return
        }
    }
    log.Fatal(err)
}

fmt.Println("Broadcast deleted successfully")
```

## End and Delete

Live broadcasts cannot be deleted. End it first:

```go
// Check if broadcast is live
broadcast, _ := streaming.GetBroadcast(ctx, client, broadcastID, "status")

if broadcast.IsLive() {
    // End the broadcast first
    _, err := streaming.TransitionBroadcast(ctx, client, broadcastID, streaming.TransitionComplete)
    if err != nil {
        log.Fatal(err)
    }
}

// Now delete it
err := streaming.DeleteBroadcast(ctx, client, broadcastID)
if err != nil {
    log.Fatal(err)
}
```

## Rules

- **Cannot delete live broadcasts**: Must end the broadcast first
- **Stream not deleted**: Deleting a broadcast doesn't delete the bound stream
- **Permanent**: This action cannot be undone

## Common Errors

| Error | Description |
|-------|-------------|
| `NotFoundError` | Broadcast doesn't exist |
| `liveBroadcastCannotBeDeleted` | Broadcast is live |
| `ForbiddenError` | No permission to delete |
