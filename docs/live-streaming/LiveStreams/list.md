---
layout: default
title: LiveStreams.list
description: Returns a list of video streams that match the API request parameters
---

Returns a list of video streams that match the API request parameters.

## Request

### HTTP Request

```
GET https://www.googleapis.com/youtube/v3/liveStreams
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `part` | Yes | string | Comma-separated list of resource parts. Valid values: `id`, `snippet`, `cdn`, `contentDetails`, `status`. |
| `id` | No* | string | Comma-separated list of stream IDs to retrieve. |
| `mine` | No* | boolean | Set to `true` to retrieve the authenticated user's streams. |
| `maxResults` | No | integer | Maximum number of items to return (1-50). Default: 5. |
| `pageToken` | No | string | Token for pagination. |

*At least one of `id` or `mine` is required.

### Authorization

Requires OAuth 2.0 authorization with one of the following scopes:

- `https://www.googleapis.com/auth/youtube.readonly`
- `https://www.googleapis.com/auth/youtube`
- `https://www.googleapis.com/auth/youtube.force-ssl`

### Request Body

Do not provide a request body when calling this method.

## Response

```json
{
  "kind": "youtube#liveStreamListResponse",
  "etag": "string",
  "nextPageToken": "string",
  "prevPageToken": "string",
  "pageInfo": {
    "totalResults": integer,
    "resultsPerPage": integer
  },
  "items": [
    liveStream Resource
  ]
}
```

### liveStream Resource

| Property | Type | Description |
|----------|------|-------------|
| `id` | string | The stream's unique identifier. |
| `snippet.publishedAt` | datetime | When the stream was created. |
| `snippet.channelId` | string | ID of the channel that owns the stream. |
| `snippet.title` | string | The stream's title. |
| `snippet.description` | string | The stream's description. |
| `snippet.isDefaultStream` | boolean | Whether this is the default stream. |
| `cdn.ingestionType` | string | Ingest type: `rtmp`, `dash`, `hls`. |
| `cdn.resolution` | string | Resolution: `240p`, `360p`, `480p`, `720p`, `1080p`, `1440p`, `2160p`, `variable`. |
| `cdn.frameRate` | string | Frame rate: `30fps`, `60fps`, `variable`. |
| `cdn.ingestionInfo.streamName` | string | The stream key for OBS/streaming software. |
| `cdn.ingestionInfo.ingestionAddress` | string | Primary RTMP ingest URL. |
| `cdn.ingestionInfo.rtmpsIngestionAddress` | string | Primary RTMPS (secure) ingest URL. |
| `status.streamStatus` | string | Status: `active`, `created`, `error`, `inactive`, `ready`. |
| `status.healthStatus.status` | string | Health: `good`, `ok`, `bad`, `noData`. |

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `badRequest` | Required parameter is missing or invalid. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | Access to the stream is forbidden. |
| 404 | `notFound` | The specified stream does not exist. |

## Quota Cost

This method consumes **5 quota units**.

---

## Yougopher Implementation

### GetStreams

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

### GetStream

Retrieve a single stream by ID.

```go
stream, err := streaming.GetStream(ctx, client, "stream-id", "snippet", "cdn", "status")
if err != nil {
    var notFound *core.NotFoundError
    if errors.As(err, &notFound) {
        log.Println("Stream not found")
    }
}
```

### GetMyStreams

Retrieve all streams owned by the authenticated user.

```go
resp, err := streaming.GetMyStreams(ctx, client, "snippet", "cdn", "status")
for _, stream := range resp.Items {
    fmt.Printf("Stream: %s\n", stream.Snippet.Title)
    fmt.Printf("  Key: %s\n", stream.StreamKey())
    fmt.Printf("  RTMP: %s\n", stream.RTMPUrl())
}
```

### Helper Methods

```go
// Get streaming credentials
stream.StreamKey()   // Stream key for OBS
stream.RTMPUrl()     // Primary RTMP URL
stream.RTMPSUrl()    // Secure RTMPS URL

// Check stream status
stream.IsActive()    // Receiving video data
stream.IsReady()     // Ready to go live
stream.IsHealthy()   // Health is good or ok

// Check for issues
stream.HasConfigurationIssues() // Configuration problems exist
```

### Stream Status Values

| Status | Description |
|--------|-------------|
| `created` | Stream was created but never used. |
| `ready` | Stream is ready to receive data. |
| `active` | Stream is actively receiving video. |
| `inactive` | Stream was active but stopped receiving. |
| `error` | Stream encountered an error. |

### Health Status Values

| Status | Description |
|--------|-------------|
| `good` | Stream is healthy. |
| `ok` | Stream has minor issues but is working. |
| `bad` | Stream has problems affecting quality. |
| `noData` | No health data available yet. |
