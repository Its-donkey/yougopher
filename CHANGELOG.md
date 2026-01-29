# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Fixed

## [0.2.2] - 2026-01-29

### Added
- Initial project structure
- Package skeleton for core, auth, streaming, data, analytics
- Core: HTTP client with OAuth and API key authentication
- Core: Error types (APIError, QuotaError, RateLimitError, AuthError, ChatEndedError)
- Core: Configurable exponential backoff with test-friendly jitter injection
- Core: Thread-safe quota tracker with automatic Pacific midnight reset
- Core: Generic response types with pagination support
- Auth: OAuth 2.0 authorization code flow with token exchange
- Auth: Automatic token refresh with configurable early refresh window
- Auth: Thread-safe token management with expiry tracking
- Auth: YouTube API scopes (live chat, moderation, read-only, upload)
- Streaming: LiveChatPoller with HTTP polling loop
- Streaming: Dynamic poll interval from API response with configurable bounds
- Streaming: pageToken pagination for continuous message retrieval
- Streaming: All message type parsing (text, SuperChat, SuperSticker, membership, etc.)
- Streaming: Composable handler pattern with idempotent unsubscribe
- Streaming: Panic recovery in handlers
- Streaming: Auto-retry on transient errors with exponential backoff
- Streaming: Moderation actions (ban, timeout, unban, add/remove moderator)
- Streaming: Send and delete message support
- Streaming: ChatBotClient high-level wrapper with semantic event handlers
- Streaming: Semantic event types (ChatMessage, SuperChatEvent, MembershipEvent, etc.)
- Streaming: Author struct with role flags (IsModerator, IsOwner, IsMember, IsVerified)
- Streaming: TokenProvider interface for flexible auth integration
- Streaming: BanEvent with typed BanType (permanent/temporary) and duration
- Streaming: LiveBroadcast retrieval (GetBroadcasts, GetBroadcast, GetMyActiveBroadcast)
- Streaming: GetBroadcastLiveChatID for retrieving live chat ID from broadcasts
- Data: Video resource (GetVideos, GetVideo, GetLiveChatID)
- Data: Video helper methods (IsLive, IsUpcoming, HasActiveLiveChat)
- Data: Channel resource (GetChannels, GetChannel, GetMyChannel)
- Data: Channel helper methods (UploadsPlaylistID)
- Data: Playlist resource (GetPlaylists, GetPlaylist, GetMyPlaylists)
- Data: PlaylistItem resource (GetPlaylistItems, VideoID helper)
- Data: Search resource with 100 quota unit warning (Search, SearchVideos, SearchLiveStreams, SearchChannels)
- Data: Search helper methods (IsVideo, IsChannel, IsPlaylist, IsLive, IsUpcoming, ResourceID)
- Data: CommentThread resource (GetCommentThreads, GetVideoComments)
- Data: Comment resource (GetComments, GetCommentReplies)
- Data: Comment helper methods (TopLevelComment, ReplyCount, AuthorID, IsReply)
- Data: Subscription resource (GetSubscriptions, GetMySubscriptions, GetChannelSubscriptions, IsSubscribedTo)
- Data: Subscription helper methods (SubscribedChannelID, HasNewContent)
- Core: NotFoundError for resource-not-found handling
- Core: Cache with TTL support (Set, Get, SetWithTTL, GetOrSet, Cleanup, Stats)
- Core: Middleware framework with MiddlewareChain for composing middlewares
- Core: LoggingMiddleware for request/response logging with timing
- Core: RetryMiddleware with exponential backoff and customizable retry logic
- Core: MetricsMiddleware for tracking request counts and durations
- Core: RateLimitingMiddleware for controlling requests per second
- Core: CachingMiddleware for cache key generation
- Auth: Device Code Flow for limited-input devices (TVs, consoles, CLI apps)
- Auth: DeviceClient with RequestDeviceCode, PollForToken, PollForTokenAsync
- Auth: DeviceAuthError with typed error detection (IsAuthorizationPending, IsExpired, etc.)
- Auth: Service Account JWT authentication for server-to-server communication
- Auth: ServiceAccountClient with FetchToken, AccessToken, auto-refresh
- Auth: NewServiceAccountClientFromJSON for loading Google Cloud credentials
- Auth: Domain-wide delegation support via WithSubject option
- Analytics: YouTube Analytics API client for channel statistics
- Analytics: Query method with full parameter support (metrics, dimensions, filters, sort)
- Analytics: Convenience methods (QueryChannelViews, QueryDailyViews, QueryTopVideos)
- Analytics: QueryCountryBreakdown and QueryDeviceBreakdown for demographics
- Analytics: QueryRevenueReport for monetization metrics
- Analytics: Report with typed accessors (GetString, GetInt, GetFloat)
- Analytics: Aggregate methods (TotalViews, TotalMinutesWatched)
- Analytics: AnalyticsError with typed detection (IsPermissionDenied, IsQuotaExceeded)
- Streaming: Broadcast management (InsertBroadcast, UpdateBroadcast, DeleteBroadcast)
- Streaming: Broadcast state transitions (TransitionBroadcast with testing/live/complete states)
- Streaming: Broadcast stream binding (BindBroadcast)
- Streaming: LiveStream CRUD (GetStreams, GetStream, InsertStream, UpdateStream, DeleteStream)
- Streaming: Stream helper methods (StreamKey, RTMPUrl, RTMPSUrl, IsActive, IsHealthy)
- Streaming: StreamController high-level workflow helper
- Streaming: CreateBroadcastWithStream for one-call setup
- Streaming: Lifecycle methods (StartTesting, GoLive, EndBroadcast, GetStreamHealth)
- Core: InvalidTransitionError for broadcast state transition errors
- Core: StreamNotHealthyError for stream health issues
- Core: StreamNotBoundError for unbound broadcast errors
- CI: Mutation testing with mutagoph (diff-based incremental testing)
- CI: Mutation report merging and GitHub Actions summary
- Tests: Mutation-killing tests for streaming package
