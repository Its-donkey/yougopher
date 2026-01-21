---
layout: default
title: UpdateStream
description: Update an existing video stream
---

Updates an existing video stream.

**Quota Cost:** 50 units

## UpdateStream

```go
// Retrieve the existing stream
stream, err := streaming.GetStream(ctx, client, "stream-id", "snippet", "cdn")
if err != nil {
    log.Fatal(err)
}

// Modify fields
stream.Snippet.Title = "Updated Stream Title"
stream.Snippet.Description = "New description"

// Update the stream
updated, err := streaming.UpdateStream(ctx, client, stream, "snippet")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Updated stream: %s\n", updated.Snippet.Title)
```

## Update CDN Settings

CDN settings can only be updated when the stream is not active.

```go
stream, _ := streaming.GetStream(ctx, client, streamID, "cdn", "status")

if stream.IsActive() {
    log.Fatal("Cannot update CDN while stream is active")
}

stream.CDN.Resolution = "1080p"
stream.CDN.FrameRate = "60fps"

updated, err := streaming.UpdateStream(ctx, client, stream, "cdn")
```

## Make Stream Reusable

```go
stream, _ := streaming.GetStream(ctx, client, streamID, "contentDetails")
stream.ContentDetails.IsReusable = true

updated, err := streaming.UpdateStream(ctx, client, stream, "contentDetails")
```

## Updatable Fields by Part

| Part | Updatable Fields |
|------|------------------|
| `snippet` | `Title`, `Description` |
| `cdn` | `Resolution`, `FrameRate` (only when not active) |
| `contentDetails` | `IsReusable` |

## Restrictions

| Field | Restriction |
|-------|-------------|
| `IngestionType` | Cannot be changed after creation |
| `StreamKey` | Cannot be changed - delete and recreate stream |
| CDN settings | Cannot be changed while actively streaming |

## Common Errors

| Error | Description |
|-------|-------------|
| `NotFoundError` | Stream doesn't exist |
| `liveStreamCannotBeUpdated` | CDN settings cannot be changed while active |
| `invalidResolution` | Invalid resolution value |
