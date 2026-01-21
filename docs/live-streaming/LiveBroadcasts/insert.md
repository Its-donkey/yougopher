---
layout: default
title: LiveBroadcasts.insert
description: Creates a new YouTube live broadcast
---

Creates a new YouTube live broadcast.

## Request

### HTTP Request

```
POST https://www.googleapis.com/youtube/v3/liveBroadcasts
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `part` | Yes | string | Comma-separated list of resource parts to include in the response. At minimum, must include the parts being set in the request body. |

### Authorization

Requires OAuth 2.0 authorization with the following scope:

- `https://www.googleapis.com/auth/youtube.force-ssl`

### Request Body

Provide a liveBroadcast resource in the request body.

```json
{
  "snippet": {
    "title": "string",
    "description": "string",
    "scheduledStartTime": "datetime",
    "scheduledEndTime": "datetime"
  },
  "status": {
    "privacyStatus": "string",
    "selfDeclaredMadeForKids": boolean
  },
  "contentDetails": {
    "enableDvr": boolean,
    "enableEmbed": boolean,
    "enableAutoStart": boolean,
    "enableAutoStop": boolean,
    "enableClosedCaptions": boolean,
    "latencyPreference": "string",
    "projection": "string"
  }
}
```

### Required Fields

| Field | Description |
|-------|-------------|
| `snippet.title` | The broadcast's title (required). |
| `snippet.scheduledStartTime` | When the broadcast is scheduled to start (required). |
| `status.privacyStatus` | Privacy setting: `private`, `public`, or `unlisted`. |

### Optional Fields

| Field | Description |
|-------|-------------|
| `snippet.description` | The broadcast's description. |
| `snippet.scheduledEndTime` | When the broadcast is scheduled to end. |
| `status.selfDeclaredMadeForKids` | Whether the broadcast is made for kids. |
| `contentDetails.enableDvr` | Enable DVR for the broadcast. Default: true. |
| `contentDetails.enableEmbed` | Allow embedding. Default: true. |
| `contentDetails.enableAutoStart` | Auto-start when stream is active. |
| `contentDetails.enableAutoStop` | Auto-stop when stream ends. |
| `contentDetails.latencyPreference` | `normal`, `low`, or `ultraLow`. |
| `contentDetails.projection` | `rectangular` or `360`. |

## Response

If successful, this method returns the created liveBroadcast resource.

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `invalidScheduledStartTime` | The scheduled start time is invalid. |
| 400 | `invalidScheduledEndTime` | The scheduled end time must be after start time. |
| 400 | `titleRequired` | A title is required. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | The user is not allowed to create broadcasts. |
| 403 | `liveStreamingNotEnabled` | The channel does not have live streaming enabled. |

## Quota Cost

This method consumes **50 quota units**.

---

## Yougopher Implementation

### InsertBroadcast

Create a new live broadcast.

```go
startTime := time.Now().Add(1 * time.Hour)

broadcast := &streaming.LiveBroadcast{
    Snippet: &streaming.BroadcastSnippet{
        Title:              "My Live Stream",
        Description:        "Streaming live from my channel",
        ScheduledStartTime: &startTime,
    },
    Status: &streaming.BroadcastStatus{
        PrivacyStatus: "unlisted",
    },
    ContentDetails: &streaming.BroadcastContentDetails{
        EnableDvr:         true,
        EnableEmbed:       true,
        EnableAutoStart:   true,
        EnableAutoStop:    true,
        LatencyPreference: "normal",
    },
}

created, err := streaming.InsertBroadcast(ctx, client, broadcast, "snippet", "status", "contentDetails")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created broadcast: %s\n", created.ID)
fmt.Printf("Live chat ID: %s\n", created.LiveChatID())
```

### Privacy Options

| Value | Description |
|-------|-------------|
| `private` | Only you can view the broadcast. |
| `unlisted` | Anyone with the link can view. |
| `public` | Anyone can find and view. |

### Latency Options

| Value | Description |
|-------|-------------|
| `normal` | Standard latency (15-30 seconds). Highest quality. |
| `low` | Low latency (5-10 seconds). Good balance. |
| `ultraLow` | Ultra-low latency (2-4 seconds). Reduced quality. |
