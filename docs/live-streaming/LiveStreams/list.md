---
layout: default
title: GetStreams
description: Retrieve video streams
---

Retrieve video streams matching filter parameters.

**Quota Cost:** 5 units

## GetStreams

Retrieve multiple streams with filtering options.

```go
resp, err := streaming.GetStreams(ctx, client, &streaming.GetStreamsParams{
    Mine:  true,
    Parts: []string{"snippet", "cdn", "status"},
})
if err != nil {
    log.Fatal(err)
}

for _, stream := range resp.Items {
    fmt.Printf("Stream: %s - %s\n", stream.Snippet.Title, stream.Status.StreamStatus)
}
```

### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `Mine` | bool | Retrieve the authenticated user's streams |
| `IDs` | []string | Specific stream IDs to retrieve |
| `Parts` | []string | Resource parts: `snippet`, `cdn`, `status`, `contentDetails` |
| `MaxResults` | int | Maximum results (1-50, default 5) |
| `PageToken` | string | Pagination token |

## GetStream

Retrieve a single stream by ID.

```go
stream, err := streaming.GetStream(ctx, client, "stream-id", "snippet", "cdn", "status")
if err != nil {
    var notFound *core.NotFoundError
    if errors.As(err, &notFound) {
        log.Println("Stream not found")
        return
    }
    log.Fatal(err)
}

fmt.Printf("Title: %s\n", stream.Snippet.Title)
fmt.Printf("Status: %s\n", stream.Status.StreamStatus)
```

## GetMyStreams

Retrieve all streams owned by the authenticated user.

```go
resp, err := streaming.GetMyStreams(ctx, client, "snippet", "cdn", "status")
if err != nil {
    log.Fatal(err)
}

for _, stream := range resp.Items {
    fmt.Printf("Stream: %s\n", stream.Snippet.Title)
    fmt.Printf("  Key: %s\n", stream.StreamKey())
    fmt.Printf("  RTMP: %s\n", stream.RTMPUrl())
}
```

## Get Streaming Credentials

```go
stream, _ := streaming.GetStream(ctx, client, streamID, "cdn")

fmt.Println("=== OBS Configuration ===")
fmt.Printf("Server: %s\n", stream.RTMPSUrl())  // Use RTMPS for security
fmt.Printf("Stream Key: %s\n", stream.StreamKey())
```

## Check Stream Health

```go
stream, _ := streaming.GetStream(ctx, client, streamID, "status")

if stream.IsActive() {
    if stream.IsHealthy() {
        fmt.Println("Stream is healthy")
    } else {
        fmt.Println("Stream has issues")
        for _, issue := range stream.Status.HealthStatus.ConfigurationIssues {
            fmt.Printf("  - %s: %s\n", issue.Severity, issue.Description)
        }
    }
}
```

## Common Errors

| Error | Description |
|-------|-------------|
| `NotFoundError` | Stream doesn't exist |
| `AuthError` | Missing or invalid OAuth token |
| `ForbiddenError` | No access to this stream |
