---
layout: default
title: InsertStream
description: Create a new video stream
---

Creates a new video stream.

**Quota Cost:** 50 units

## InsertStream

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

## Required Fields

| Field | Description |
|-------|-------------|
| `Snippet.Title` | The stream's title |
| `CDN.IngestionType` | `rtmp`, `dash`, or `hls` |
| `CDN.Resolution` | Video resolution |
| `CDN.FrameRate` | Video frame rate |

## Ingestion Types

| Value | Description |
|-------|-------------|
| `rtmp` | Real-Time Messaging Protocol (most common) |
| `dash` | Dynamic Adaptive Streaming over HTTP |
| `hls` | HTTP Live Streaming |

## Resolution Options

| Value | Size |
|-------|------|
| `240p` | 426x240 |
| `360p` | 640x360 |
| `480p` | 854x480 |
| `720p` | 1280x720 (HD) |
| `1080p` | 1920x1080 (Full HD) |
| `1440p` | 2560x1440 (2K) |
| `2160p` | 3840x2160 (4K) |
| `variable` | Auto-detect |

## Frame Rate Options

| Value | Description |
|-------|-------------|
| `30fps` | 30 frames per second |
| `60fps` | 60 frames per second |
| `variable` | Auto-detect |

## OBS Configuration

After creating a stream, configure OBS:

```go
stream, _ := streaming.InsertStream(ctx, client, stream, "cdn")

fmt.Println("=== OBS Configuration ===")
fmt.Printf("Server: %s\n", stream.RTMPSUrl()) // Use RTMPS for security
fmt.Printf("Stream Key: %s\n", stream.StreamKey())
```

## Reusable Streams

Create a reusable stream for multiple broadcasts:

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

## Best Practices

1. **Match encoder settings**: Set resolution and frame rate to match your streaming software
2. **Use RTMPS**: Prefer the secure RTMPS URL over plain RTMP
3. **Keep stream key secret**: The stream key acts like a password
4. **Use reusable streams**: Create one stream and reuse it for multiple broadcasts

## Common Errors

| Error | Description |
|-------|-------------|
| `titleRequired` | Title is missing |
| `cdnRequired` | CDN configuration is missing |
| `invalidResolution` | Invalid resolution value |
| `invalidFrameRate` | Invalid frame rate value |
| `liveStreamingNotEnabled` | Channel doesn't have live streaming |
