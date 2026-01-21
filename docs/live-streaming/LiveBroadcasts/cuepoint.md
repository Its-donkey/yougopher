---
layout: default
title: InsertCuepoint
description: Insert an ad break into a live broadcast
---

Inserts a cuepoint (ad break) into a live broadcast.

**Quota Cost:** 50 units

## InsertCuepoint

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

## InsertImmediateCuepoint

Convenience function for immediate ad break with default 30-second duration.

```go
cuepoint, err := streaming.InsertImmediateCuepoint(ctx, client, "broadcast-id")
if err != nil {
    log.Fatal(err)
}
```

## Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `BroadcastID` | string | required | The broadcast to insert into |
| `DurationSecs` | int | 30 | Ad break duration (1-180 seconds) |
| `InsertionOffsetTimeMs` | int64 | -1 | When to insert relative to broadcast start |
| `WalltimeMs` | int64 | 0 | Wall clock time for insertion |

## Scheduled Cuepoint

Insert a cuepoint at a specific time relative to broadcast start:

```go
cuepoint, err := streaming.InsertCuepoint(ctx, client, &streaming.InsertCuepointParams{
    BroadcastID:           "broadcast-id",
    DurationSecs:          45,
    InsertionOffsetTimeMs: 3600000, // 1 hour into the broadcast
})
```

## Wall Clock Time

Insert a cuepoint at a specific wall clock time:

```go
insertTime := time.Now().Add(5 * time.Minute)

cuepoint, err := streaming.InsertCuepoint(ctx, client, &streaming.InsertCuepointParams{
    BroadcastID:  "broadcast-id",
    DurationSecs: 30,
    WalltimeMs:   insertTime.UnixMilli(),
})
```

## Constants

```go
streaming.CuepointInsertImmediate // -1 - Insert immediately
```

## Requirements

- **Broadcast must be live**: Cuepoints cannot be inserted before going live
- **Monetization enabled**: Channel must have monetization enabled
- **Not all viewers see ads**: YouTube Premium subscribers don't see ads

## Common Errors

| Error | Description |
|-------|-------------|
| `broadcastNotLive` | Broadcast isn't live yet |
| `monetizationNotEnabled` | Channel can't run ads |
| `invalidCuepointDuration` | Duration must be 1-180 seconds |
