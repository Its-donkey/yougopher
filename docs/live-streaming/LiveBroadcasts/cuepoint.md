---
layout: default
title: LiveBroadcasts.cuepoint
description: Inserts a cuepoint (ad break) into a live broadcast
---

Inserts a cuepoint (ad break) into a live broadcast.

## Request

### HTTP Request

```
POST https://www.googleapis.com/youtube/v3/liveBroadcasts/cuepoint
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `id` | Yes | string | The ID of the broadcast to insert the cuepoint into. |

### Authorization

Requires OAuth 2.0 authorization with the following scope:

- `https://www.googleapis.com/auth/youtube.force-ssl`

### Request Body

```json
{
  "cueType": "cueTypeAd",
  "durationSecs": integer,
  "insertionOffsetTimeMs": long,
  "walltimeMs": long
}
```

### Fields

| Field | Required | Description |
|-------|----------|-------------|
| `cueType` | Yes | Must be `"cueTypeAd"` for ad breaks. |
| `durationSecs` | No | Ad break duration in seconds. Default: 30. Maximum: 180. |
| `insertionOffsetTimeMs` | No | When to insert relative to broadcast start (ms). Use -1 for immediate. |
| `walltimeMs` | No | Wall clock time for insertion (Unix timestamp in ms). Takes precedence over `insertionOffsetTimeMs`. |

## Response

If successful, this method returns a cuepoint resource:

```json
{
  "kind": "youtube#cuepoint",
  "etag": "string",
  "id": "string",
  "insertionOffsetTimeMs": long,
  "walltimeMs": long,
  "durationSecs": integer,
  "cueType": "cueTypeAd"
}
```

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `idRequired` | The broadcast ID is required. |
| 400 | `invalidCuepointDuration` | Duration must be between 1 and 180 seconds. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | The user cannot insert cuepoints in this broadcast. |
| 403 | `broadcastNotLive` | Cuepoints can only be inserted into live broadcasts. |
| 403 | `monetizationNotEnabled` | The channel does not have monetization enabled. |
| 404 | `liveBroadcastNotFound` | The broadcast does not exist. |

## Quota Cost

This method consumes **50 quota units**.

---

## Yougopher Implementation

### InsertCuepoint

Insert a cuepoint with full control over timing.

```go
cuepoint, err := streaming.InsertCuepoint(ctx, client, &streaming.InsertCuepointParams{
    BroadcastID:  "broadcast-id",
    DurationSecs: 60, // 60 second ad break
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Cuepoint inserted: %s\n", cuepoint.ID)
```

### InsertImmediateCuepoint

Convenience function for immediate ad break with default 30-second duration.

```go
cuepoint, err := streaming.InsertImmediateCuepoint(ctx, client, "broadcast-id")
if err != nil {
    log.Fatal(err)
}
```

### Scheduled Cuepoint

Insert a cuepoint at a specific time relative to broadcast start.

```go
cuepoint, err := streaming.InsertCuepoint(ctx, client, &streaming.InsertCuepointParams{
    BroadcastID:           "broadcast-id",
    DurationSecs:          45,
    InsertionOffsetTimeMs: 3600000, // 1 hour into the broadcast
})
```

### Wall Clock Time

Insert a cuepoint at a specific wall clock time.

```go
insertTime := time.Now().Add(5 * time.Minute)

cuepoint, err := streaming.InsertCuepoint(ctx, client, &streaming.InsertCuepointParams{
    BroadcastID:  "broadcast-id",
    DurationSecs: 30,
    WalltimeMs:   insertTime.UnixMilli(),
})
```

### Cuepoint Constants

```go
streaming.CuepointInsertImmediate // -1 - Insert immediately
```

### Requirements

- **Broadcast must be live**: Cuepoints cannot be inserted before going live.
- **Monetization enabled**: The channel must have monetization enabled for ads to play.
- **Not all viewers see ads**: Ad-free subscribers (YouTube Premium) won't see the break.

### Best Practices

1. **Announce breaks**: Tell viewers an ad break is coming.
2. **30-90 second duration**: Standard ad break length.
3. **Avoid frequent breaks**: YouTube recommends no more than 4-5 breaks per hour.
4. **Natural pause points**: Insert breaks during natural transitions in content.
