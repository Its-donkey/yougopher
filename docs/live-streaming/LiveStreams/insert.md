---
layout: default
title: LiveStreams.insert
description: Creates a new video stream
---

Creates a new video stream.

## Request

### HTTP Request

```
POST https://www.googleapis.com/youtube/v3/liveStreams
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `part` | Yes | string | Comma-separated list of resource parts to include in the response. |

### Authorization

Requires OAuth 2.0 authorization with the following scope:

- `https://www.googleapis.com/auth/youtube.force-ssl`

### Request Body

```json
{
  "snippet": {
    "title": "string",
    "description": "string"
  },
  "cdn": {
    "ingestionType": "string",
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
| `snippet.title` | The stream's title (required). |
| `cdn.ingestionType` | Ingest method: `rtmp`, `dash`, or `hls` (required). |
| `cdn.resolution` | Video resolution (required). |
| `cdn.frameRate` | Video frame rate (required). |

### CDN Configuration Options

**Ingestion Type:**

| Value | Description |
|-------|-------------|
| `rtmp` | Real-Time Messaging Protocol (most common). |
| `dash` | Dynamic Adaptive Streaming over HTTP. |
| `hls` | HTTP Live Streaming. |

**Resolution:**

| Value | Description |
|-------|-------------|
| `240p` | 426x240 |
| `360p` | 640x360 |
| `480p` | 854x480 |
| `720p` | 1280x720 (HD) |
| `1080p` | 1920x1080 (Full HD) |
| `1440p` | 2560x1440 (2K) |
| `2160p` | 3840x2160 (4K) |
| `variable` | Auto-detect resolution |

**Frame Rate:**

| Value | Description |
|-------|-------------|
| `30fps` | 30 frames per second |
| `60fps` | 60 frames per second |
| `variable` | Auto-detect frame rate |

## Response

If successful, this method returns the created liveStream resource, including the generated stream key and ingest URLs.

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `titleRequired` | A title is required. |
| 400 | `cdnRequired` | CDN configuration is required. |
| 400 | `invalidResolution` | The resolution value is invalid. |
| 400 | `invalidFrameRate` | The frame rate value is invalid. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | The user cannot create streams. |
| 403 | `liveStreamingNotEnabled` | Live streaming is not enabled for this channel. |

## Quota Cost

This method consumes **50 quota units**.

---

## Yougopher Implementation

### InsertStream

Create a new live stream.

```go
stream := &streaming.LiveStream{
    Snippet: &streaming.StreamSnippet{
        Title:       "Primary Stream",
        Description: "My main streaming source",
    },
    CDN: &streaming.StreamCDN{
        IngestionType: "rtmp",
        Resolution:    "1080p",
        FrameRate:     "30fps",
    },
    ContentDetails: &streaming.StreamContentDetails{
        IsReusable: true,
    },
}

created, err := streaming.InsertStream(ctx, client, stream, "snippet", "cdn", "status")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Stream created: %s\n", created.ID)
fmt.Printf("Stream key: %s\n", created.StreamKey())
fmt.Printf("RTMP URL: %s\n", created.RTMPUrl())
fmt.Printf("RTMPS URL: %s\n", created.RTMPSUrl())
```

### OBS Configuration

After creating a stream, configure OBS:

```go
stream, _ := streaming.InsertStream(ctx, client, stream, "cdn")

fmt.Println("=== OBS Configuration ===")
fmt.Printf("Server: %s\n", stream.RTMPSUrl()) // Use RTMPS for security
fmt.Printf("Stream Key: %s\n", stream.StreamKey())
```

### Reusable Streams

Create a reusable stream that can be bound to multiple broadcasts:

```go
stream := &streaming.LiveStream{
    Snippet: &streaming.StreamSnippet{
        Title: "Reusable Stream",
    },
    CDN: &streaming.StreamCDN{
        IngestionType: "rtmp",
        Resolution:    "1080p",
        FrameRate:     "60fps",
    },
    ContentDetails: &streaming.StreamContentDetails{
        IsReusable: true, // Can be reused for future broadcasts
    },
}
```

### Best Practices

1. **Match encoder settings**: Set resolution and frame rate to match your streaming software.
2. **Use RTMPS**: Prefer the secure RTMPS URL over plain RTMP.
3. **Keep stream key secret**: The stream key acts like a password.
4. **Use reusable streams**: Create one stream and reuse it for multiple broadcasts.
