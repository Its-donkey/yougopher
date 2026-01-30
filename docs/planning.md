---
layout: default
title: Implementation Plan
description: Yougopher development roadmap and architecture.
---

A YouTube API toolkit in Go focused on **live chat bot functionality**.

## Progress Tracking

| Phase | Status | Coverage | Docs |
|-------|--------|----------|------|
| 1. Repository & CI Setup | ✅ Complete | N/A | ✅ |
| 2. Core Infrastructure | ✅ Complete | 94% | ✅ |
| 3. Authentication | ✅ Complete | 90.6% | ✅ |
| 4. Live Chat Poller | ✅ Complete | 91.2% | ✅ |
| 5. ChatBotClient | ✅ Complete | 91.2% | ✅ |
| 6. Supporting Data API | ✅ Complete | 93.4% | ✅ |
| 7. Remaining Data API | ✅ Complete | 95.8% | ✅ |
| 8. Cache & Middleware | ✅ Complete | 93.1% | ✅ |
| 9. Advanced Auth & Analytics | ✅ Complete | 90.1% | ✅ |
| 10. Live Streaming Control | ✅ Complete | 90%+ | ✅ |

Legend: ⬜ Not Started | 🟡 In Progress | ✅ Complete

---

## API Stability & Versioning Policy

### Semantic Versioning
- **v0.x.x** - Development phase, API may change between minor versions
- **v1.0.0+** - Stable API, follows semantic versioning strictly

### Stability Tiers

| Tier | Marker | Guarantee |
|------|--------|-----------|
| **Stable** | (default) | Breaking changes only in major versions |
| **Beta** | `// Beta:` | May change in minor versions with deprecation notice |
| **Experimental** | `// Experimental:` | May change or be removed in any release |
| **Internal** | `internal/` package | No stability guarantee, not for external use |

### Breaking Change Policy (post-v1.0.0)
**These ARE breaking changes:**
- Removing or renaming exported types, functions, methods, or fields
- Changing function/method signatures
- Changing behavior that users depend on
- Adding required parameters

**These are NOT breaking changes:**
- Adding new exported types, functions, methods, or fields
- Adding optional parameters via functional options
- Bug fixes that correct clearly wrong behavior
- Performance improvements

### Deprecation Process
1. Add `// Deprecated:` comment with alternative
2. Keep deprecated API working for at least 2 minor versions
3. Remove in next major version

```go
// Deprecated: Use NewChatBotClientWithConfig instead.
// Will be removed in v2.0.0.
func NewChatBotClient(liveChatID string) *ChatBotClient
```

### Version Compatibility
- Support **latest 2 Go versions** (currently 1.23, 1.24)
- YouTube API changes tracked via automated monitoring

---

## Primary Use Cases

1. **Live chat bot** - Monitor and interact with YouTube live chat
2. **Chat monitoring** - Read-only logging, moderation alerts

---

## Project Structure

```
yougopher/
├── go.mod
├── youtube/
│   ├── core/                     # Core infrastructure
│   │   ├── client.go             # Base HTTP client, options pattern
│   │   ├── errors.go             # Error types (APIError, QuotaError)
│   │   ├── quota.go              # YouTube quota tracking
│   │   ├── response.go           # Generic response types
│   │   ├── cache.go              # Caching layer with TTL
│   │   └── middleware.go         # Request/response middleware
│   │
│   ├── auth/                     # Authentication (separate for modularity)
│   │   ├── auth.go               # AuthClient, OAuth flows
│   │   ├── token.go              # Token struct, refresh, validation
│   │   ├── service_account.go    # Service account JWT auth
│   │   └── device.go             # Device code flow for TVs
│   │
│   ├── data/                     # YouTube Data API v3
│   │   ├── videos.go             # Videos resource
│   │   ├── channels.go           # Channels resource
│   │   ├── playlists.go          # Playlists resource
│   │   ├── playlist_items.go     # PlaylistItems resource
│   │   ├── search.go             # Search resource (100 quota/call)
│   │   ├── comments.go           # Comments, CommentThreads
│   │   └── subscriptions.go      # Subscriptions resource
│   │
│   ├── streaming/                # YouTube Live Streaming API
│   │   ├── broadcasts.go         # LiveBroadcasts resource
│   │   ├── streams.go            # LiveStreams resource
│   │   ├── chat.go               # ChatBotClient (high-level)
│   │   ├── poller.go             # LiveChatPoller (low-level)
│   │   ├── types.go              # Chat message types
│   │   ├── moderator.go          # Moderation actions
│   │   ├── controller.go         # StreamController (high-level)
│   │   └── superchat.go          # SuperChatEvents resource
│   │
│   └── analytics/                # YouTube Analytics API
│       ├── reports.go            # Analytics queries
│       └── groups.go             # Groups, GroupItems
│
└── docs/
    └── examples/
        ├── chat-bot/             # Basic chat bot
        └── moderation-bot/       # Auto-moderation
```

---

## Live Chat Client Architecture

### Two-Layer Design

```
┌─────────────────────────────────────────────────────────┐
│  ChatBotClient (High-Level) - streaming/chat.go         │
│  - Semantic event handlers (OnSuperChat, OnMembership)  │
│  - Convenience methods (Say, Ban, Timeout)              │
│  - Automatic token handling from auth.AuthClient        │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│  LiveChatPoller (Low-Level) - streaming/poller.go       │
│  - HTTP polling loop                                    │
│  - Dynamic poll interval from API                       │
│  - pageToken management                                 │
│  - Raw message handlers                                 │
│  - Auto-reconnect on errors                             │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│  YouTube Live Chat API                                  │
│  - liveChatMessages.list (polling)                      │
│  - liveChatMessages.insert (send)                       │
│  - liveChatBans, liveChatModerators                     │
└─────────────────────────────────────────────────────────┘
```

---

## Error Types

```go
// APIError for YouTube API errors
type APIError struct {
    StatusCode int
    Code       string  // e.g., "quotaExceeded", "forbidden"
    Message    string
    Details    []ErrorDetail
}

// QuotaError when daily quota exceeded
type QuotaError struct {
    Used      int
    Limit     int
    ResetAt   time.Time  // Pacific midnight
}

// RateLimitError for per-second rate limits
type RateLimitError struct {
    RetryAfter time.Duration
}

// AuthError for authentication failures
type AuthError struct {
    Code    string  // "invalid_grant", "expired_token", etc.
    Message string
}

// ChatEndedError when live chat has ended
type ChatEndedError struct {
    LiveChatID string
}
```

---

## Quota Costs

```go
var QuotaCosts = map[string]int{
    "liveChatMessages.list":    5,   // Polling
    "liveChatMessages.insert": 50,   // Send message
    "liveChatMessages.delete": 50,   // Delete message
    "liveChatBans.insert":     50,   // Ban user
    "liveChatBans.delete":     50,   // Unban user
    "liveChatModerators.insert": 50, // Add mod
    "liveChatModerators.delete": 50, // Remove mod
    "videos.list":              1,
    "channels.list":            1,
    "search.list":            100,   // Expensive!
}
```

---

## Testing Strategy

- **Pattern:** Table-driven tests with `*_test.go` files alongside source
- **Mocking:** HTTP responses via `httptest.Server`
- **Coverage target:** 90% (tracked in TODO.md, reported in CI step summary)
- **Integration tests:** Build tag `// +build integration`, run separately
