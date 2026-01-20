# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project structure
- GitHub Actions CI/CD workflows
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
