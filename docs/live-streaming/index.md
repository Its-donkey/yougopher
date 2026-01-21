---
layout: default
title: Live Streaming
description: YouTube Live Streaming API with Yougopher
---

The `youtube/streaming` package provides a complete implementation for YouTube Live Streaming.

## Resources

| Resource | Description |
|----------|-------------|
| [LiveBroadcasts](LiveBroadcasts/) | Manage broadcasts - create, update, delete, bind streams, go live |
| [LiveStreams](LiveStreams/) | Manage video streams - create, configure, get stream keys |
| [LiveChatMessages](LiveChatMessages/) | Read and send chat messages with polling or SSE |
| [LiveChatBans](LiveChatBans/) | Ban and timeout users in live chat |
| [LiveChatModerators](LiveChatModerators/) | Add and remove chat moderators |
| [SuperChatEvents](SuperChatEvents/) | Access Super Chat and Super Sticker data |

## Key Types

| Type | Description |
|------|-------------|
| `LiveBroadcast` | A live streaming event |
| `LiveStream` | Video stream with RTMP credentials |
| `LiveChatMessage` | Chat message with author details |
| `LiveChatPoller` | Automatic chat polling client |
| `LiveChatStream` | SSE streaming chat client |
| `ChatBotClient` | High-level chat bot |
| `StreamController` | High-level broadcast manager |

## Quota Costs

| Operation | Cost |
|-----------|------|
| List broadcasts/streams/chat | 5 |
| Insert/update/delete broadcast | 50 |
| Insert/update/delete stream | 50 |
| Bind/transition broadcast | 50 |
| Send/delete chat message | 50 |
| Ban/unban user | 50 |
| Add/remove moderator | 50 |
