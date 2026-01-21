---
layout: default
title: LiveBroadcasts.list
description: Returns a list of YouTube broadcasts that match the API request parameters
---

Returns a list of YouTube broadcasts that match the API request parameters.

## Request

### HTTP Request

```
GET https://www.googleapis.com/youtube/v3/liveBroadcasts
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `part` | Yes | string | Comma-separated list of resource parts to include. Valid values: `id`, `snippet`, `contentDetails`, `status`. |
| `id` | No* | string | Comma-separated list of broadcast IDs to retrieve. |
| `mine` | No* | boolean | Set to `true` to retrieve the authenticated user's broadcasts. |
| `broadcastStatus` | No | string | Filter by broadcast status. Values: `active`, `all`, `completed`, `upcoming`. |
| `broadcastType` | No | string | Filter by broadcast type. Values: `all`, `event`, `persistent`. |
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

If successful, this method returns a response body with the following structure:

```json
{
  "kind": "youtube#liveBroadcastListResponse",
  "etag": "string",
  "nextPageToken": "string",
  "prevPageToken": "string",
  "pageInfo": {
    "totalResults": integer,
    "resultsPerPage": integer
  },
  "items": [
    liveBroadcast Resource
  ]
}
```

### Properties

| Property | Type | Description |
|----------|------|-------------|
| `kind` | string | Identifies the resource type. Value: `"youtube#liveBroadcastListResponse"`. |
| `etag` | string | The ETag of the response. |
| `nextPageToken` | string | Token for the next page of results. |
| `prevPageToken` | string | Token for the previous page of results. |
| `pageInfo.totalResults` | integer | Total number of results. |
| `pageInfo.resultsPerPage` | integer | Number of results per page. |
| `items[]` | list | List of broadcast resources. |

### liveBroadcast Resource

| Property | Type | Description |
|----------|------|-------------|
| `id` | string | The broadcast's unique identifier. |
| `snippet.publishedAt` | datetime | When the broadcast was created. |
| `snippet.channelId` | string | ID of the channel that created the broadcast. |
| `snippet.title` | string | The broadcast's title. |
| `snippet.description` | string | The broadcast's description. |
| `snippet.scheduledStartTime` | datetime | Scheduled start time. |
| `snippet.scheduledEndTime` | datetime | Scheduled end time. |
| `snippet.actualStartTime` | datetime | Actual start time. |
| `snippet.actualEndTime` | datetime | Actual end time. |
| `snippet.liveChatId` | string | The live chat ID for this broadcast. |
| `status.lifeCycleStatus` | string | Lifecycle status: `complete`, `created`, `live`, `liveStarting`, `ready`, `revoked`, `testStarting`, `testing`. |
| `status.privacyStatus` | string | Privacy status: `private`, `public`, `unlisted`. |
| `status.recordingStatus` | string | Recording status: `notRecording`, `recorded`, `recording`. |
| `contentDetails.boundStreamId` | string | ID of the bound live stream. |
| `contentDetails.enableDvr` | boolean | Whether DVR is enabled. |
| `contentDetails.enableEmbed` | boolean | Whether embedding is allowed. |
| `contentDetails.enableAutoStart` | boolean | Whether auto-start is enabled. |
| `contentDetails.enableAutoStop` | boolean | Whether auto-stop is enabled. |

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `badRequest` | Required parameter is missing or invalid. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | Access to the broadcast is forbidden. |
| 404 | `notFound` | The specified broadcast does not exist. |

## Quota Cost

This method consumes **5 quota units**.

---

## Yougopher Implementation

### GetBroadcasts

Retrieve multiple broadcasts with filtering options.

```go
resp, err := streaming.GetBroadcasts(ctx, client, &streaming.GetBroadcastsParams{
    Mine:            true,
    BroadcastStatus: "active",
    Parts:           []string{"snippet", "status", "contentDetails"},
    MaxResults:      10,
})
```

### GetBroadcast

Retrieve a single broadcast by ID.

```go
broadcast, err := streaming.GetBroadcast(ctx, client, "broadcast-id", "snippet", "status")
if err != nil {
    var notFound *core.NotFoundError
    if errors.As(err, &notFound) {
        log.Println("Broadcast not found")
    }
}
```

### GetMyActiveBroadcast

Retrieve the authenticated user's currently active broadcast.

```go
broadcast, err := streaming.GetMyActiveBroadcast(ctx, client, "snippet", "status")
if broadcast.IsLive() {
    fmt.Println("Currently streaming!")
}
```

### GetBroadcastLiveChatID

Get the live chat ID for a broadcast.

```go
liveChatID, err := streaming.GetBroadcastLiveChatID(ctx, client, "broadcast-id")
```

### Helper Methods

```go
// Check broadcast status
broadcast.IsLive()      // Currently live
broadcast.IsComplete()  // Has ended
broadcast.IsUpcoming()  // Not started yet
broadcast.IsTesting()   // In testing state

// Get related IDs
broadcast.LiveChatID()    // Get live chat ID
broadcast.BoundStreamID() // Get bound stream ID
broadcast.HasBoundStream() // Check if stream is bound
```
