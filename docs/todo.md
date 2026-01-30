---
layout: default
title: TODO
description: Outstanding work for Yougopher.
---

Track outstanding work for Yougopher.

## Phase 1: Repository & CI Setup ✅
- [x] go.mod
- [x] GitHub Actions workflows
- [x] Mutation testing with mutagoph (diff-based, incremental reports)
- [x] dependabot.yml
- [x] README.md
- [x] CONTRIBUTING.md
- [x] CHANGELOG.md
- [x] LICENSE
- [x] Directory skeleton with doc.go files

## Phase 2: Core Infrastructure ✅
- [x] youtube/core/client.go - Base HTTP client
- [x] youtube/core/errors.go - Error types
- [x] youtube/core/quota.go - Quota tracking
- [x] youtube/core/response.go - Response types
- [x] Tests (94% coverage)
- [x] Documentation (docs/core.md)

## Phase 3: Authentication ✅
- [x] youtube/auth/auth.go - OAuth flows
- [x] youtube/auth/token.go - Token management
- [x] Tests (90.6% coverage)
- [x] Documentation (docs/auth.md)

## Phase 4: Live Chat Poller ✅
- [x] youtube/streaming/poller.go - Polling loop
- [x] youtube/streaming/types.go - Message types
- [x] Tests (91.2% coverage)
- [x] Documentation (doc.go, docs/streaming.md)

## Phase 5: ChatBotClient ✅
- [x] youtube/streaming/chat.go - High-level client
- [x] Moderation methods in chat.go (Ban, Timeout, Unban, Delete)
- [x] Tests (91.2% coverage)
- [x] Documentation (doc.go, docs/streaming.md)

## Phase 6: Supporting Data API ✅
- [x] youtube/data/videos.go - Video retrieval, GetLiveChatID
- [x] youtube/data/channels.go - Channel retrieval, GetMyChannel
- [x] youtube/streaming/broadcasts.go - Broadcast retrieval
- [x] Tests (93.4% data, 91.0% streaming)
- [x] Documentation (doc.go)

## Phase 7: Remaining Data API ✅
- [x] youtube/data/playlists.go - Playlists, PlaylistItems
- [x] youtube/data/search.go - Search (100 quota/call)
- [x] youtube/data/comments.go - CommentThreads, Comments
- [x] youtube/data/subscriptions.go - Subscriptions
- [x] Tests (95.8% coverage)
- [x] Documentation (doc.go)

## Phase 8: Cache & Middleware ✅
- [x] youtube/core/cache.go - In-memory cache with TTL
- [x] youtube/core/middleware.go - Logging, retry, metrics, rate limiting
- [x] Tests (93.1% coverage)
- [x] Documentation (doc.go, docs/core.md)

## Phase 9: Advanced Auth & Analytics ✅
- [x] youtube/auth/device.go - Device Code Flow for TVs/CLI
- [x] youtube/auth/service_account.go - Service Account JWT auth
- [x] youtube/analytics/reports.go - Analytics API queries
- [x] Tests (90.1% coverage)
- [x] Documentation (doc.go, docs/auth.md, docs/analytics.md)

## Phase 10: Live Streaming Control ✅
- [x] youtube/streaming/streams.go - Stream CRUD operations
- [x] youtube/streaming/broadcasts.go - Broadcast mutations & transitions
- [x] youtube/streaming/controller.go - StreamController high-level workflow
- [x] youtube/core/errors.go - Streaming error types
- [x] Tests (90.8% streaming, 93.2% core)
- [x] Documentation (doc.go, docs/streaming.md)

---

## Decisions Needed

### sponsors vs members API
The `sponsors` resource in the Live Streaming API is deprecated in favor of `members` in the Data API v3.

Options:
1. **Implement members only** - Use Data API `members.list` for channel membership info
2. **Implement both** - Support legacy `sponsors` for backwards compatibility
3. **Skip entirely** - Membership info already comes through live chat events (newSponsorEvent, memberMilestoneChatEvent)

Reference:
- [sponsors (deprecated)](https://developers.google.com/youtube/v3/live/docs/sponsors)
- [members](https://developers.google.com/youtube/v3/docs/members)
