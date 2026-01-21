---
layout: default
title: LiveChatMessages.insert
description: Sends a message to a live chat
---

Sends a message to a live chat.

## Request

### HTTP Request

```
POST https://www.googleapis.com/youtube/v3/liveChat/messages
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `part` | Yes | string | Must include `snippet`. |

### Authorization

Requires OAuth 2.0 authorization with the following scope:

- `https://www.googleapis.com/auth/youtube.force-ssl`

### Request Body

```json
{
  "snippet": {
    "liveChatId": "string",
    "type": "textMessageEvent",
    "textMessageDetails": {
      "messageText": "string"
    }
  }
}
```

### Required Fields

| Field | Description |
|-------|-------------|
| `snippet.liveChatId` | The live chat ID to send the message to. |
| `snippet.type` | Must be `textMessageEvent`. |
| `snippet.textMessageDetails.messageText` | The message text (1-200 characters). |

## Response

If successful, this method returns the created liveChatMessage resource.

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `liveChatIdRequired` | The liveChatId is required. |
| 400 | `messageTextRequired` | The message text is required. |
| 400 | `messageTextTooLong` | Message exceeds 200 characters. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | The user cannot send messages to this chat. |
| 403 | `liveChatDisabled` | Live chat is disabled. |
| 403 | `liveChatEnded` | The live chat has ended. |
| 403 | `rateLimitExceeded` | Sending messages too quickly. |
| 404 | `liveChatNotFound` | The live chat does not exist. |

## Quota Cost

This method consumes **50 quota units**.

---

## Yougopher Implementation

### SendMessage (via Poller)

Send a message using the LiveChatPoller.

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

msg, err := poller.SendMessage(ctx, "Hello, chat!")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Message sent: %s\n", msg.ID)
```

### SendMessage (via ChatBotClient)

The high-level ChatBotClient also provides message sending.

```go
bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
if err != nil {
    log.Fatal(err)
}

err = bot.SendMessage(ctx, "Hello from the bot!")
```

### Message Limits

- **Length**: 1-200 characters
- **Rate limit**: Don't send more than 3 messages per second
- **Slow mode**: If slow mode is enabled, respect the delay

### Error Handling

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
        }
    }
}
```

### Best Practices

1. **Don't spam**: Limit message frequency.
2. **Check chat state**: Verify chat is active before sending.
3. **Handle rate limits**: Implement backoff when rate limited.
4. **Respect slow mode**: Check if slow mode is enabled and wait accordingly.
