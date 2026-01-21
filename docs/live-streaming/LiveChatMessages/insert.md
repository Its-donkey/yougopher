---
layout: default
title: SendMessage
description: Send a message to live chat
---

Sends a message to a live chat.

**Quota Cost:** 50 units

## SendMessage (via Poller)

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

msg, err := poller.SendMessage(ctx, "Hello, chat!")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Message sent: %s\n", msg.ID)
```

## SendMessage (via ChatBotClient)

```go
bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
if err != nil {
    log.Fatal(err)
}

err = bot.SendMessage(ctx, "Hello from the bot!")
```

## Limits

- **Length**: 1-200 characters
- **Rate limit**: ~3 messages per second
- **Slow mode**: Respect the delay if enabled

## Error Handling

```go
msg, err := poller.SendMessage(ctx, text)
if err != nil {
    var apiErr *core.APIError
    if errors.As(err, &apiErr) {
        switch {
        case apiErr.IsChatEnded():
            log.Println("Chat has ended")
        case apiErr.IsChatDisabled():
            log.Println("Chat is disabled")
        case apiErr.Code == "rateLimitExceeded":
            log.Println("Sending too fast, slow down")
        case apiErr.Code == "messageTextTooLong":
            log.Println("Message exceeds 200 characters")
        }
    }
}
```

## Best Practices

1. **Limit message frequency**: Implement rate limiting to avoid API throttling
2. **Check chat state**: Verify chat is active before sending
3. **Handle rate limits**: Implement exponential backoff when rate limited
4. **Respect slow mode**: Check if slow mode is enabled and wait accordingly

## Common Errors

| Error | Description |
|-------|-------------|
| `messageTextRequired` | Empty message |
| `messageTextTooLong` | Exceeds 200 characters |
| `rateLimitExceeded` | Sending too fast |
| `liveChatEnded` | Chat has ended |
| `liveChatDisabled` | Chat is disabled |
