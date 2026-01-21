---
layout: default
title: LiveStreams.update
description: Updates an existing video stream
---

Updates an existing video stream.

## Request

### HTTP Request

```
PUT https://www.googleapis.com/youtube/v3/liveStreams
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `part` | Yes | string | Comma-separated list of resource parts being updated. |

### Authorization

Requires OAuth 2.0 authorization with the following scope:

- `https://www.googleapis.com/auth/youtube.force-ssl`

### Request Body

Provide a liveStream resource in the request body. Must include the `id` field.

```json
{
  "id": "string",
  "snippet": {
    "title": "string",
    "description": "string"
  },
  "cdn": {
    "resolution": "string",
    "frameRate": "string"
  },
  "contentDetails": {
    "isReusable": boolean
  }
}
```

### Required Fields

| Field | Description |
|-------|-------------|
| `id` | The stream ID to update (required). |

### Updatable Fields

| Part | Fields |
|------|--------|
| `snippet` | `title`, `description` |
| `cdn` | `resolution`, `frameRate` (only when not active) |
| `contentDetails` | `isReusable` |

## Response

If successful, this method returns the updated liveStream resource.

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `idRequired` | The stream ID is required. |
| 400 | `invalidResolution` | The resolution value is invalid. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | The user cannot update this stream. |
| 403 | `liveStreamCannotBeUpdated` | CDN settings cannot be changed while stream is active. |
| 404 | `liveStreamNotFound` | The stream does not exist. |

## Quota Cost

This method consumes **50 quota units**.

---

## Yougopher Implementation

### UpdateStream

Update an existing stream.

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

### Update CDN Settings

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

### Make Stream Reusable

```go
stream, _ := streaming.GetStream(ctx, client, streamID, "contentDetails")
stream.ContentDetails.IsReusable = true

updated, err := streaming.UpdateStream(ctx, client, stream, "contentDetails")
```

### Restrictions

- **CDN settings**: Cannot be changed while the stream is actively receiving video.
- **Ingestion type**: Cannot be changed after creation. Create a new stream instead.
- **Stream key**: Cannot be changed. Delete and recreate the stream to get a new key.
