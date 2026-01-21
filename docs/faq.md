---
layout: default
title: FAQ
description: Frequently asked questions about Yougopher
---

## General

### What is Yougopher?

Yougopher is a Go library for interacting with the YouTube Data API v3 and YouTube Live Streaming API. It provides a clean, idiomatic Go interface for building YouTube integrations.

### What Go version is required?

Yougopher requires Go 1.21 or later.

### Is Yougopher officially supported by Google/YouTube?

No, Yougopher is a community-developed library. It uses the official YouTube API but is not affiliated with or endorsed by Google.

---

## Authentication

### What authentication methods are supported?

Yougopher supports:
- **OAuth 2.0 Web Flow** - For web applications
- **OAuth 2.0 Device Flow** - For CLI applications and devices without browsers
- **Service Accounts** - For server-to-server communication

### How do I get API credentials?

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a project and enable the YouTube Data API v3
3. Create OAuth 2.0 credentials
4. Download the credentials JSON file

### Why do I need OAuth? Can I use an API key?

Most YouTube API endpoints require OAuth authentication to access user-specific data. API keys only work for public data and have limited functionality. Yougopher focuses on authenticated operations.

### How do I refresh expired tokens?

The OAuth2 client automatically refreshes tokens when they expire. Just ensure you're using `config.Client(ctx, token)` which wraps the HTTP client with auto-refresh capability.

### Can I use service accounts for live streaming?

Service accounts work for some operations but have limitations with live streaming. For full live chat functionality, OAuth 2.0 with user consent is recommended.

---

## API Usage

### What's the API quota limit?

The YouTube Data API has a default quota of 10,000 units per day. Different operations cost different amounts:
- Read operations: 1-5 units
- Write operations: 50 units
- Video uploads: 1,600 units

### How do I check my quota usage?

You can view quota usage in the [Google Cloud Console](https://console.cloud.google.com/apis/api/youtube.googleapis.com/quotas) under APIs & Services → YouTube Data API v3 → Quotas.

### Can I increase my quota?

Yes, you can request a quota increase through the Google Cloud Console. You'll need to explain your use case and may need to comply with additional requirements.

### How do I handle rate limiting?

Yougopher returns errors when rate limited. Implement exponential backoff:

```go
backoff := time.Second
for retries := 0; retries < 3; retries++ {
    err := makeAPICall()
    if err == nil {
        break
    }
    if isRateLimitError(err) {
        time.Sleep(backoff)
        backoff *= 2
    }
}
```

---

## Live Streaming

### How do I get the live chat ID?

The live chat ID is in the broadcast's snippet:

```go
broadcasts, _ := streaming.GetBroadcasts(ctx, client, &streaming.GetBroadcastsParams{
    BroadcastStatus: "active",
    Part:            []string{"snippet"},
})
chatID := broadcasts.Items[0].Snippet.LiveChatID
```

### What's the difference between polling and SSE streaming?

- **Polling**: Repeatedly calls the API at intervals. Simpler but higher latency and quota usage.
- **SSE Streaming**: Maintains a persistent connection for real-time updates. Lower latency and more efficient.

### Why are my chat messages delayed?

YouTube's API has inherent latency. For lowest latency:
1. Use SSE streaming instead of polling
2. Reduce your polling interval (but watch quota usage)
3. Consider that YouTube itself adds some delay for moderation

### Can I read chat from someone else's stream?

Yes, you can read public chat messages from any live stream. However, you cannot send messages or moderate unless you have appropriate permissions.

### How do I handle Super Chats?

Super Chats appear as special message types in the chat:

```go
poller.OnMessage(func(msg *streaming.LiveChatMessage) {
    if msg.Type() == streaming.MessageTypeSuperChat {
        amount := msg.Snippet.SuperChatDetails.AmountDisplayString
        // Handle Super Chat
    }
})
```

---

## Troubleshooting

### "Access Not Configured" Error

The YouTube Data API isn't enabled for your project. Enable it in the Google Cloud Console.

### "Invalid Credentials" Error

- Check your client ID and secret are correct
- Ensure the OAuth consent screen is configured
- Verify the token hasn't been revoked

### "Quota Exceeded" Error

You've hit your daily quota limit. Either wait for reset (midnight Pacific Time) or request a quota increase.

### "Live Chat Not Found" Error

- The broadcast may have ended
- The chat ID may be invalid
- Chat may be disabled for that broadcast

### Messages not appearing in chat

- Check that the message was sent successfully (no error returned)
- There may be a delay before messages appear
- YouTube may filter certain messages

---

## Contributing

### How can I contribute to Yougopher?

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Submit a pull request

### Where do I report bugs?

Open an issue on the [GitHub repository](https://github.com/Its-donkey/yougopher/issues).

### Is there a roadmap?

Check the `.TODO.md` file in the repository for planned features.
