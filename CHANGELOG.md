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
