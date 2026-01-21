---
layout: default
title: Live Streaming API Reference
description: API reference for YouTube Live Streaming API resources and methods
---

API reference documentation for the YouTube Live Streaming API, following the structure of [Google's official documentation](https://developers.google.com/youtube/v3/live/docs).

## Resources

| Resource | Description |
|----------|-------------|
| [liveBroadcasts](LiveBroadcasts/) | A broadcast represents the event being streamed to YouTube. |
| [liveStreams](LiveStreams/) | A stream is the video feed sent to YouTube for encoding. |
| [liveChatMessages](LiveChatMessages/) | Messages sent by viewers during a live broadcast. |
| [liveChatBans](LiveChatBans/) | Ban users from participating in live chat. |
| [liveChatModerators](LiveChatModerators/) | Moderators who can delete messages and ban users. |
| [superChatEvents](SuperChatEvents/) | Super Chat and Super Sticker monetary contributions. |

## Yougopher Implementation

Each resource documentation includes:
1. **Overview** - Resource representation, properties table
2. **Methods** - API reference with request/response format, parameters, errors
3. **Yougopher Examples** - Go code examples

### Key Types

| Yougopher Type | API Resource |
|---------------|--------------|
| `streaming.LiveBroadcast` | liveBroadcast |
| `streaming.LiveStream` | liveStream |
| `streaming.LiveChatMessage` | liveChatMessage |
| `streaming.LiveChatBan` | liveChatBan |
| `streaming.LiveChatModerator` | liveChatModerator |
| `streaming.LiveChatPoller` | liveChatMessages polling client |
| `streaming.LiveChatStream` | liveChatMessages SSE client |
| `streaming.ChatBotClient` | High-level chat bot |
| `streaming.StreamController` | High-level broadcast manager |

### Quick Start

```go
import (
    "github.com/Its-donkey/yougopher/youtube/core"
    "github.com/Its-donkey/yougopher/youtube/streaming"
)

// Create client
client := core.NewClient(core.WithAPIKey("your-api-key"))

// For authenticated requests
client := core.NewClient()
client.SetAccessToken("your-oauth-token")
```

## Related Documentation

- [Streaming Package](../streaming) - High-level overview and ChatBotClient documentation
- [Google Live Streaming API](https://developers.google.com/youtube/v3/live/docs) - Official API documentation
