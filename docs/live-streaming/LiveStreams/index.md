---
layout: default
title: LiveStreams
description: Video stream management for YouTube live streaming
---

A `LiveStream` contains information about the video stream you're transmitting to YouTube. Once created, a stream can be bound to one or more broadcasts.

## Functions

| Function | Quota | Description |
|----------|-------|-------------|
| [GetStreams](list) | 5 | Returns streams matching parameters |
| [GetStream](list) | 5 | Returns a single stream by ID |
| [GetMyStreams](list) | 5 | Returns your streams |
| [InsertStream](insert) | 50 | Creates a new stream |
| [UpdateStream](update) | 50 | Updates an existing stream |
| [DeleteStream](delete) | 50 | Deletes a stream |

## Type Definition

```go
type LiveStream struct {
    Kind           string                `json:"kind,omitempty"`
    ETag           string                `json:"etag,omitempty"`
    ID             string                `json:"id,omitempty"`
    Snippet        *StreamSnippet        `json:"snippet,omitempty"`
    CDN            *StreamCDN            `json:"cdn,omitempty"`
    Status         *StreamStatus         `json:"status,omitempty"`
    ContentDetails *StreamContentDetails `json:"contentDetails,omitempty"`
}
```

## Helper Methods

```go
stream.IsActive()               // Receiving video data
stream.IsReady()                // Ready to go live
stream.IsHealthy()              // Health is good or ok
stream.StreamKey()              // Stream key for OBS
stream.RTMPUrl()                // Primary RTMP URL
stream.RTMPSUrl()               // Secure RTMPS URL
stream.HasConfigurationIssues() // Configuration problems exist
```

## Status Values

| Value | Description |
|-------|-------------|
| `created` | Stream was created but never used |
| `ready` | Ready to receive data |
| `active` | Actively receiving video |
| `inactive` | Was active but stopped receiving |
| `error` | Stream encountered an error |

## Health Values

| Value | Description |
|-------|-------------|
| `good` | Stream is healthy |
| `ok` | Minor issues but working |
| `bad` | Problems affecting quality |
| `noData` | No health data available yet |

## Quick Example

```go
import "github.com/Its-donkey/yougopher/youtube/streaming"

// Create a stream
stream, err := streaming.InsertStream(ctx, client, &streaming.LiveStream{
    Snippet: &streaming.StreamSnippet{
        Title: "Primary Stream",
    },
    CDN: &streaming.StreamCDN{
        IngestionType: "rtmp",
        Resolution:    "1080p",
        FrameRate:     "30fps",
    },
    ContentDetails: &streaming.StreamContentDetails{
        IsReusable: true,
    },
}, "snippet", "cdn", "status")

// Get OBS configuration
fmt.Printf("Server: %s\n", stream.RTMPSUrl())
fmt.Printf("Stream Key: %s\n", stream.StreamKey())

// Check stream status
if stream.IsActive() && stream.IsHealthy() {
    fmt.Println("Stream is ready to go live!")
}
```
