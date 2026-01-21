---
layout: default
title: LiveStreams
description: Video stream management for YouTube live streaming
---

A `liveStream` resource contains information about the video stream that you are transmitting to YouTube.

The stream provides the content that will be broadcast to YouTube users. Once created, a `liveStream` resource can be bound to one or more `liveBroadcast` resources.

## Methods

| Method | HTTP Request | Description |
|--------|--------------|-------------|
| [list](list) | `GET /liveStreams` | Returns a list of video streams that match the API request parameters. |
| [insert](insert) | `POST /liveStreams` | Creates a video stream. |
| [update](update) | `PUT /liveStreams` | Updates a video stream. |
| [delete](delete) | `DELETE /liveStreams` | Deletes a video stream. |

## Resource Representation

```json
{
  "kind": "youtube#liveStream",
  "etag": string,
  "id": string,
  "snippet": {
    "publishedAt": datetime,
    "channelId": string,
    "title": string,
    "description": string,
    "isDefaultStream": boolean
  },
  "cdn": {
    "ingestionType": string,
    "ingestionInfo": {
      "streamName": string,
      "ingestionAddress": string,
      "backupIngestionAddress": string,
      "rtmpsIngestionAddress": string,
      "rtmpsBackupIngestionAddress": string
    },
    "resolution": string,
    "frameRate": string
  },
  "status": {
    "streamStatus": string,
    "healthStatus": {
      "status": string,
      "lastUpdateTimeSeconds": unsigned long,
      "configurationIssues": [
        {
          "type": string,
          "severity": string,
          "reason": string,
          "description": string
        }
      ]
    }
  },
  "contentDetails": {
    "closedCaptionsIngestionUrl": string,
    "isReusable": boolean
  }
}
```

## Properties

| Property | Type | Description |
|----------|------|-------------|
| `kind` | string | Identifies the resource type. Value: `youtube#liveStream`. |
| `etag` | string | The ETag of the resource. |
| `id` | string | The ID that YouTube assigns to uniquely identify the stream. |

### snippet

| Property | Type | Description |
|----------|------|-------------|
| `snippet.publishedAt` | datetime | The date and time the stream was created. |
| `snippet.channelId` | string | The ID of the channel that owns the stream. |
| `snippet.title` | string | The stream's title. |
| `snippet.description` | string | The stream's description. |
| `snippet.isDefaultStream` | boolean | Whether this is the channel's default stream. |

### cdn

| Property | Type | Description |
|----------|------|-------------|
| `cdn.ingestionType` | string | The method for sending the video stream. Values: `rtmp`, `dash`, `hls`. |
| `cdn.ingestionInfo.streamName` | string | The stream key used in streaming software (e.g., OBS). |
| `cdn.ingestionInfo.ingestionAddress` | string | The primary RTMP ingest URL. |
| `cdn.ingestionInfo.backupIngestionAddress` | string | The backup RTMP ingest URL. |
| `cdn.ingestionInfo.rtmpsIngestionAddress` | string | The primary RTMPS (secure) ingest URL. |
| `cdn.ingestionInfo.rtmpsBackupIngestionAddress` | string | The backup RTMPS ingest URL. |
| `cdn.resolution` | string | The resolution. Values: `240p`, `360p`, `480p`, `720p`, `1080p`, `1440p`, `2160p`, `variable`. |
| `cdn.frameRate` | string | The frame rate. Values: `30fps`, `60fps`, `variable`. |

### status

| Property | Type | Description |
|----------|------|-------------|
| `status.streamStatus` | string | The stream's status. Values: `active`, `created`, `error`, `inactive`, `ready`. |
| `status.healthStatus.status` | string | Health status. Values: `good`, `ok`, `bad`, `noData`. |
| `status.healthStatus.lastUpdateTimeSeconds` | long | When the health status was last updated. |
| `status.healthStatus.configurationIssues` | list | List of configuration issues. |

### contentDetails

| Property | Type | Description |
|----------|------|-------------|
| `contentDetails.closedCaptionsIngestionUrl` | string | The URL to send closed captions. |
| `contentDetails.isReusable` | boolean | Whether the stream can be reused for multiple broadcasts. |

---

## Yougopher Type

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

### Helper Methods

```go
stream.IsActive()              // Returns true if receiving video data
stream.IsReady()               // Returns true if ready to go live
stream.IsHealthy()             // Returns true if health is good or ok
stream.StreamKey()             // Returns the stream key for OBS
stream.RTMPUrl()               // Returns the primary RTMP URL
stream.RTMPSUrl()              // Returns the secure RTMPS URL
stream.HasConfigurationIssues() // Returns true if there are issues
```

### Status Values

| Value | Description |
|-------|-------------|
| `created` | Stream was created but never used. |
| `ready` | Stream is ready to receive data. |
| `active` | Stream is actively receiving video. |
| `inactive` | Stream was active but stopped receiving. |
| `error` | Stream encountered an error. |

### Health Values

| Value | Description |
|-------|-------------|
| `good` | Stream is healthy. |
| `ok` | Stream has minor issues but is working. |
| `bad` | Stream has problems affecting quality. |
| `noData` | No health data available yet. |
