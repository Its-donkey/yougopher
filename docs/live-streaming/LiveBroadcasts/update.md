---
layout: default
title: UpdateBroadcast
description: Update an existing YouTube live broadcast
---

Updates an existing YouTube live broadcast.

**Quota Cost:** 50 units

## UpdateBroadcast

```go
// First, retrieve the existing broadcast
broadcast, err := streaming.GetBroadcast(ctx, client, "broadcast-id", "snippet", "status")
if err != nil {
    log.Fatal(err)
}

// Modify the fields you want to update
broadcast.Snippet.Title = "Updated Stream Title"
broadcast.Snippet.Description = "New description for the stream"

// Update the broadcast (only specify parts you're updating)
updated, err := streaming.UpdateBroadcast(ctx, client, broadcast, "snippet")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Updated broadcast title: %s\n", updated.Snippet.Title)
```

## Update Privacy

```go
broadcast, _ := streaming.GetBroadcast(ctx, client, broadcastID, "status")
broadcast.Status.PrivacyStatus = "public"

updated, err := streaming.UpdateBroadcast(ctx, client, broadcast, "status")
```

## Update Content Details

```go
broadcast, _ := streaming.GetBroadcast(ctx, client, broadcastID, "contentDetails")
broadcast.ContentDetails.EnableDVR = false
broadcast.ContentDetails.EnableEmbed = true

updated, err := streaming.UpdateBroadcast(ctx, client, broadcast, "contentDetails")
```

## Updatable Fields by Part

| Part | Updatable Fields |
|------|------------------|
| `snippet` | `Title`, `Description`, `ScheduledStartTime`, `ScheduledEndTime` |
| `status` | `PrivacyStatus`, `SelfDeclaredMadeForKids` |
| `contentDetails` | `EnableDVR`, `EnableEmbed`, `EnableAutoStart`, `EnableAutoStop`, `EnableClosedCaptions` |

## Restrictions

Some fields cannot be updated after the broadcast has started:

| Field | Restriction |
|-------|-------------|
| `LatencyPreference` | Must be set before going live |
| `Projection` | Must be set at creation time |
| `ScheduledStartTime` | Cannot change after broadcast starts |

## Common Errors

| Error | Description |
|-------|-------------|
| `NotFoundError` | Broadcast doesn't exist |
| `liveBroadcastCannotBeUpdated` | Field cannot be updated in current state |
| `invalidScheduledStartTime` | Invalid start time value |
