---
layout: default
title: YouTube API Coverage
description: Mapping of YouTube API capabilities to Yougopher implementation.
---

Mapping of all YouTube API capabilities to Yougopher implementation status.

---

## YouTube Data API v3

| Resource | Method | Covered | Yougopher Location |
|----------|--------|---------|-------------------|
| **activities** | list | - | - |
| | insert | - | - |
| **captions** | list | - | - |
| | insert | - | - |
| | update | - | - |
| | delete | - | - |
| | download | - | - |
| **channelBanners** | insert | - | - |
| **channelSections** | list | - | - |
| | insert | - | - |
| | update | - | - |
| | delete | - | - |
| **channels** | list | ✅ | `data/channels.go` |
| | update | - | - |
| **commentThreads** | list | ✅ | `data/comments.go` |
| | insert | - | - |
| **comments** | list | ✅ | `data/comments.go` |
| | insert | - | - |
| | update | - | - |
| | delete | - | - |
| | setModerationStatus | - | - |
| | markAsSpam | - | - |
| **i18nLanguages** | list | - | - |
| **i18nRegions** | list | - | - |
| **members** | list | - | - |
| **membershipsLevels** | list | - | - |
| **playlistImages** | list | - | - |
| | insert | - | - |
| | update | - | - |
| | delete | - | - |
| **playlistItems** | list | ✅ | `data/playlists.go` |
| | insert | - | - |
| | update | - | - |
| | delete | - | - |
| **playlists** | list | ✅ | `data/playlists.go` |
| | insert | - | - |
| | update | - | - |
| | delete | - | - |
| **search** | list | ✅ | `data/search.go` |
| **subscriptions** | list | ✅ | `data/subscriptions.go` |
| | insert | - | - |
| | delete | - | - |
| **thumbnails** | set | - | - |
| **videoAbuseReportReasons** | list | - | - |
| **videoCategories** | list | - | - |
| **videos** | list | ✅ | `data/videos.go` |
| | insert | - | - |
| | update | - | - |
| | delete | - | - |
| | rate | - | - |
| | getRating | - | - |
| | reportAbuse | - | - |
| **watermarks** | set | - | - |
| | unset | - | - |

---

## YouTube Live Streaming API

| Resource | Method | Covered | Yougopher Location |
|----------|--------|---------|-------------------|
| **liveBroadcasts** | list | ✅ | `streaming/broadcasts.go` |
| | insert | ✅ | `streaming/broadcasts.go` |
| | update | ✅ | `streaming/broadcasts.go` |
| | delete | ✅ | `streaming/broadcasts.go` |
| | bind | ✅ | `streaming/broadcasts.go` |
| | transition | ✅ | `streaming/broadcasts.go` |
| | cuepoint | ✅ | `streaming/broadcasts.go` |
| **liveStreams** | list | ✅ | `streaming/streams.go` |
| | insert | ✅ | `streaming/streams.go` |
| | update | ✅ | `streaming/streams.go` |
| | delete | ✅ | `streaming/streams.go` |
| **liveChatMessages** | list | ✅ | `streaming/poller.go` |
| | streamList | ✅ | `streaming/sse.go` |
| | insert | ✅ | `streaming/poller.go`, `streaming/chat.go` |
| | delete | ✅ | `streaming/poller.go`, `streaming/chat.go` |
| | transition | ✅ | `streaming/poller.go` |
| **liveChatBans** | insert | ✅ | `streaming/poller.go` |
| | delete | ✅ | `streaming/poller.go` |
| **liveChatModerators** | list | ✅ | `streaming/poller.go` |
| | insert | ✅ | `streaming/poller.go` |
| | delete | ✅ | `streaming/poller.go` |
| **superChatEvents** | list | ✅ | `streaming/poller.go` |
| **sponsors** | list | - | Deprecated (see members in Data API) |

---

## YouTube Analytics API

| Resource | Method | Covered | Yougopher Location |
|----------|--------|---------|-------------------|
| **reports** | query | ✅ | `analytics/reports.go` |
| **groups** | list | - | - |
| | insert | - | - |
| | update | - | - |
| | delete | - | - |
| **groupItems** | list | - | - |
| | insert | - | - |
| | delete | - | - |

### Supported Metrics

| Metric | Covered |
|--------|---------|
| views | ✅ |
| estimatedMinutesWatched | ✅ |
| averageViewDuration | ✅ |
| subscribersGained | ✅ |
| subscribersLost | ✅ |
| likes | ✅ |
| dislikes | ✅ |
| comments | ✅ |
| shares | ✅ |
| estimatedRevenue | ✅ |
| estimatedAdRevenue | ✅ |
| grossRevenue | ✅ |
| cpm | ✅ |
| monetizedPlaybacks | ✅ |

### Supported Dimensions

| Dimension | Covered |
|-----------|---------|
| day | ✅ |
| month | ✅ |
| video | ✅ |
| channel | ✅ |
| country | ✅ |
| province | ✅ |
| city | ✅ |
| deviceType | ✅ |
| operatingSystem | ✅ |
| ageGroup | ✅ |
| gender | ✅ |
| subscribedStatus | ✅ |
| liveOrOnDemand | ✅ |
| trafficSourceType | ✅ |

---

## Coverage Summary

| API | Resources | Methods Covered | Methods Total | Coverage |
|-----|-----------|-----------------|---------------|----------|
| Data API v3 | 18 | 10 | 50+ | 20% |
| Live Streaming API | 7 | 22 | 23 | 96% |
| Analytics API | 3 | 1 | 8 | 12% |
| **Total** | 28 | 33 | 81+ | ~41% |

### Focus Areas (High Coverage)

| Area | Coverage | Notes |
|------|----------|-------|
| Broadcasts | 100% | Full CRUD + bind + transition + cuepoint |
| Streams | 100% | Full CRUD |
| Live Chat Messages | 100% | Polling, SSE streaming, send, delete, transition |
| Live Chat Moderation | 100% | Ban/unban, list/add/remove mods |
| Super Chat Events | 100% | Historical Super Chat/Sticker list |
| Analytics Queries | 100% | All metrics/dimensions supported |

### Not Covered (By Design)

| Resource | Reason |
|----------|--------|
| sponsors.list | Deprecated in favor of Data API members resource |
| captions | Outside chat bot scope |
| watermarks | Outside chat bot scope |
| thumbnails | Outside chat bot scope |
| video upload | Outside chat bot scope |
| channelBanners | Outside chat bot scope |
| Reporting API | Bulk reports not needed for real-time |

---

## References

- [YouTube Data API v3 Reference](https://developers.google.com/youtube/v3/docs)
- [YouTube Live Streaming API Reference](https://developers.google.com/youtube/v3/live/docs)
- [YouTube Analytics API Reference](https://developers.google.com/youtube/analytics/reference)
- [YouTube Analytics Dimensions](https://developers.google.com/youtube/analytics/dimensions)
- [YouTube Analytics Metrics](https://developers.google.com/youtube/analytics/metrics)
