---
layout: default
title: Live Streaming
description: YouTube Live Streaming API with Yougopher
---

The `youtube/streaming` package provides a complete implementation for YouTube Live Streaming.

## Resources

YouTube Live Streaming API endpoints available through Yougopher.

| Resource | Methods | Description |
|----------|---------|-------------|
| [LiveBroadcasts](LiveBroadcasts/) | list, insert, update, delete, bind, transition, cuepoint | Broadcast events that appear on your channel. Bind to streams, transition through states (testing → live → complete), insert ad breaks. |
| [LiveStreams](LiveStreams/) | list, insert, update, delete | Video streams containing RTMP/RTMPS ingestion URLs and stream keys for OBS or other encoders. |
| [LiveChatMessages](LiveChatMessages/) | list, insert, delete | Chat messages including text, Super Chats, Super Stickers, memberships, and polls. |
| [LiveChatBans](LiveChatBans/) | insert, delete | Temporary timeouts or permanent bans for chat users. |
| [LiveChatModerators](LiveChatModerators/) | list, insert, delete | Grant and revoke moderator privileges for chat users. |
| [SuperChatEvents](SuperChatEvents/) | list | Monetary contributions (Super Chats and Super Stickers) with amounts and messages. |

## Types

Go structs and clients exported by the `streaming` package.

| Type | Description |
|------|-------------|
| `LiveBroadcast` | Represents a broadcast event with snippet (title, description, scheduled time), status (privacy, life cycle), content details (bound stream, monitor stream), and statistics. |
| `LiveStream` | Contains CDN settings (ingestion type, resolution, frame rate), ingestion info (RTMP URLs, stream key), and health status. |
| `LiveChatMessage` | Message with author details, snippet (message text, type), and type-specific data (Super Chat amount, poll choices, etc.). |
| `LiveChatPoller` | Client that polls for new messages at the API-recommended interval. Supports callbacks for messages, Super Chats, bans, and deletions. |
| `LiveChatStream` | SSE-based client for real-time message streaming without polling. Lower latency than polling. |
| `ChatBotClient` | High-level client combining polling, message sending, and command handling for chat bots. |
| `StreamController` | High-level client for managing the full broadcast lifecycle: create, bind, transition, and monitor. |
