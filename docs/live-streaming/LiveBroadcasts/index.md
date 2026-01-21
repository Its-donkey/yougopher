---
layout: default
title: LiveBroadcasts
description: Live broadcast management for YouTube streaming
---

A `liveBroadcast` resource represents an event that will be streamed, via live video, on YouTube.

## Methods

| Method | HTTP Request | Description |
|--------|--------------|-------------|
| [list](list) | `GET /liveBroadcasts` | Returns a list of YouTube broadcasts that match the API request parameters. |
| [insert](insert) | `POST /liveBroadcasts` | Creates a broadcast. |
| [update](update) | `PUT /liveBroadcasts` | Updates a broadcast. |
| [delete](delete) | `DELETE /liveBroadcasts` | Deletes a broadcast. |
| [bind](bind) | `POST /liveBroadcasts/bind` | Binds a YouTube broadcast to a stream or removes an existing binding. |
| [transition](transition) | `POST /liveBroadcasts/transition` | Changes the status of a YouTube live broadcast. |
| [cuepoint](cuepoint) | `POST /liveBroadcasts/cuepoint` | Inserts a cuepoint (ad break) into a live broadcast. |

## Resource Representation

```json
{
  "kind": "youtube#liveBroadcast",
  "etag": string,
  "id": string,
  "snippet": {
    "publishedAt": datetime,
    "channelId": string,
    "title": string,
    "description": string,
    "thumbnails": {
      (key): {
        "url": string,
        "width": unsigned integer,
        "height": unsigned integer
      }
    },
    "scheduledStartTime": datetime,
    "scheduledEndTime": datetime,
    "actualStartTime": datetime,
    "actualEndTime": datetime,
    "liveChatId": string,
    "isDefaultBroadcast": boolean
  },
  "status": {
    "lifeCycleStatus": string,
    "privacyStatus": string,
    "recordingStatus": string,
    "madeForKids": boolean,
    "selfDeclaredMadeForKids": boolean
  },
  "contentDetails": {
    "boundStreamId": string,
    "boundStreamLastUpdateTimeMs": string,
    "monitorStream": {
      "enableMonitorStream": boolean,
      "broadcastStreamDelayMs": unsigned integer,
      "embedHtml": string
    },
    "enableEmbed": boolean,
    "enableDvr": boolean,
    "enableContentEncryption": boolean,
    "startWithSlate": boolean,
    "recordFromStart": boolean,
    "enableClosedCaptions": boolean,
    "closedCaptionsType": string,
    "enableLowLatency": boolean,
    "latencyPreference": string,
    "projection": string,
    "enableAutoStart": boolean,
    "enableAutoStop": boolean
  }
}
```

## Properties

| Property | Type | Description |
|----------|------|-------------|
| `kind` | string | Identifies the resource type. Value: `youtube#liveBroadcast`. |
| `etag` | string | The ETag of the resource. |
| `id` | string | The ID that YouTube assigns to uniquely identify the broadcast. |

### snippet

| Property | Type | Description |
|----------|------|-------------|
| `snippet.publishedAt` | datetime | The date and time the broadcast was added to YouTube's live broadcast schedule. |
| `snippet.channelId` | string | The ID of the channel that owns the broadcast. |
| `snippet.title` | string | The broadcast's title. Max 100 characters. |
| `snippet.description` | string | The broadcast's description. Max 5000 characters. |
| `snippet.thumbnails` | object | Thumbnail images associated with the broadcast. |
| `snippet.scheduledStartTime` | datetime | The date and time that the broadcast is scheduled to start. |
| `snippet.scheduledEndTime` | datetime | The date and time that the broadcast is scheduled to end. |
| `snippet.actualStartTime` | datetime | The date and time that the broadcast actually started. |
| `snippet.actualEndTime` | datetime | The date and time that the broadcast actually ended. |
| `snippet.liveChatId` | string | The ID for the live chat associated with this broadcast. |
| `snippet.isDefaultBroadcast` | boolean | Indicates whether this is the channel's default broadcast. |

### status

| Property | Type | Description |
|----------|------|-------------|
| `status.lifeCycleStatus` | string | The broadcast's status. Values: `complete`, `created`, `live`, `liveStarting`, `ready`, `revoked`, `testStarting`, `testing`. |
| `status.privacyStatus` | string | The broadcast's privacy status. Values: `private`, `public`, `unlisted`. |
| `status.recordingStatus` | string | The recording status. Values: `notRecording`, `recorded`, `recording`. |
| `status.madeForKids` | boolean | Whether the broadcast is designated as made for kids. |
| `status.selfDeclaredMadeForKids` | boolean | The creator's declaration of whether the broadcast is made for kids. |

### contentDetails

| Property | Type | Description |
|----------|------|-------------|
| `contentDetails.boundStreamId` | string | The ID of the stream bound to the broadcast. |
| `contentDetails.boundStreamLastUpdateTimeMs` | string | The time the binding was last updated. |
| `contentDetails.monitorStream.enableMonitorStream` | boolean | Whether the monitor stream is enabled. |
| `contentDetails.monitorStream.broadcastStreamDelayMs` | integer | The delay in ms between the monitor stream and broadcast. |
| `contentDetails.monitorStream.embedHtml` | string | HTML code to embed the monitor stream player. |
| `contentDetails.enableEmbed` | boolean | Whether the broadcast can be embedded in a player. |
| `contentDetails.enableDvr` | boolean | Whether viewers can access DVR controls. |
| `contentDetails.enableContentEncryption` | boolean | Whether content encryption is enabled. |
| `contentDetails.startWithSlate` | boolean | Whether the broadcast begins with a slate. |
| `contentDetails.recordFromStart` | boolean | Whether to automatically start recording. |
| `contentDetails.enableClosedCaptions` | boolean | Whether closed captions are enabled. |
| `contentDetails.closedCaptionsType` | string | The closed captions type. |
| `contentDetails.enableLowLatency` | boolean | Whether low latency mode is enabled. |
| `contentDetails.latencyPreference` | string | Latency preference: `normal`, `low`, `ultraLow`. |
| `contentDetails.projection` | string | Broadcast projection: `rectangular`, `360`. |
| `contentDetails.enableAutoStart` | boolean | Whether to auto-start when the stream is active. |
| `contentDetails.enableAutoStop` | boolean | Whether to auto-stop when the stream ends. |

---

## Yougopher Type

```go
type LiveBroadcast struct {
    Kind           string                   `json:"kind,omitempty"`
    ETag           string                   `json:"etag,omitempty"`
    ID             string                   `json:"id,omitempty"`
    Snippet        *BroadcastSnippet        `json:"snippet,omitempty"`
    Status         *BroadcastStatus         `json:"status,omitempty"`
    ContentDetails *BroadcastContentDetails `json:"contentDetails,omitempty"`
}
```

### Helper Methods

```go
broadcast.IsLive()         // Returns true if currently live
broadcast.IsComplete()     // Returns true if broadcast has ended
broadcast.IsUpcoming()     // Returns true if not started yet
broadcast.IsTesting()      // Returns true if in testing state
broadcast.LiveChatID()     // Returns the live chat ID
broadcast.BoundStreamID()  // Returns the bound stream ID
broadcast.HasBoundStream() // Returns true if a stream is bound
```
