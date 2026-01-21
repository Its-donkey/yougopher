---
layout: default
title: LiveBroadcasts.update
description: Updates an existing YouTube live broadcast
---

Updates an existing YouTube live broadcast.

## Request

### HTTP Request

```
PUT https://www.googleapis.com/youtube/v3/liveBroadcasts
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `part` | Yes | string | Comma-separated list of resource parts being updated. |

### Authorization

Requires OAuth 2.0 authorization with the following scope:

- `https://www.googleapis.com/auth/youtube.force-ssl`

### Request Body

Provide a liveBroadcast resource in the request body. Must include the `id` field.

```json
{
  "id": "string",
  "snippet": {
    "title": "string",
    "description": "string",
    "scheduledStartTime": "datetime"
  },
  "status": {
    "privacyStatus": "string"
  },
  "contentDetails": {
    "enableDvr": boolean,
    "enableEmbed": boolean
  }
}
```

### Required Fields

| Field | Description |
|-------|-------------|
| `id` | The broadcast ID to update (required). |

### Updatable Fields

Only include the parts you want to update. Fields not included will retain their current values.

| Part | Fields |
|------|--------|
| `snippet` | `title`, `description`, `scheduledStartTime`, `scheduledEndTime` |
| `status` | `privacyStatus`, `selfDeclaredMadeForKids` |
| `contentDetails` | `enableDvr`, `enableEmbed`, `enableAutoStart`, `enableAutoStop`, `enableClosedCaptions`, `latencyPreference` |

## Response

If successful, this method returns the updated liveBroadcast resource.

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `idRequired` | The broadcast ID is required. |
| 400 | `invalidScheduledStartTime` | The scheduled start time is invalid. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | The user cannot update this broadcast. |
| 403 | `liveBroadcastCannotBeUpdated` | Certain fields cannot be updated after broadcast starts. |
| 404 | `liveBroadcastNotFound` | The broadcast does not exist. |

## Quota Cost

This method consumes **50 quota units**.

---

## Yougopher Implementation

### UpdateBroadcast

Update an existing broadcast.

```go
// First, retrieve the existing broadcast
broadcast, err := streaming.GetBroadcast(ctx, client, "broadcast-id", "snippet", "status")
if err != nil {
    log.Fatal(err)
}

// Modify the fields you want to update
broadcast.Snippet.Title = "Updated Stream Title"
broadcast.Snippet.Description = "New description for the stream"

// Update the broadcast
updated, err := streaming.UpdateBroadcast(ctx, client, broadcast, "snippet")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Updated broadcast title: %s\n", updated.Snippet.Title)
```

### Update Privacy

```go
broadcast, _ := streaming.GetBroadcast(ctx, client, broadcastID, "status")
broadcast.Status.PrivacyStatus = "public"

updated, err := streaming.UpdateBroadcast(ctx, client, broadcast, "status")
```

### Update Content Details

```go
broadcast, _ := streaming.GetBroadcast(ctx, client, broadcastID, "contentDetails")
broadcast.ContentDetails.EnableDvr = false
broadcast.ContentDetails.EnableEmbed = true

updated, err := streaming.UpdateBroadcast(ctx, client, broadcast, "contentDetails")
```

### Restrictions

Some fields cannot be updated after the broadcast has started:

- `latencyPreference` - Must be set before going live
- `projection` - Must be set at creation time
- `scheduledStartTime` - Cannot change after broadcast starts
