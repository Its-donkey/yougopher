---
layout: default
title: LiveChatMessages.delete
description: Deletes a message from a live chat
---

Deletes a message from a live chat.

## Request

### HTTP Request

```
DELETE https://www.googleapis.com/youtube/v3/liveChat/messages
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `id` | Yes | string | The ID of the message to delete. |

### Authorization

Requires OAuth 2.0 authorization with the following scope:

- `https://www.googleapis.com/auth/youtube.force-ssl`

The authenticated user must be:
- The message author, OR
- The broadcast owner, OR
- A chat moderator

### Request Body

Do not provide a request body when calling this method.

## Response

If successful, this method returns an empty response body with HTTP status code 204.

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `idRequired` | The message ID is required. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | The user cannot delete this message. |
| 403 | `liveChatEnded` | The live chat has ended. |
| 404 | `liveChatMessageNotFound` | The message does not exist. |

## Quota Cost

This method consumes **50 quota units**.

---

## Yougopher Implementation

### DeleteMessage (via Poller)

Delete a message using the LiveChatPoller.

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

err := poller.DeleteMessage(ctx, "message-id")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Message deleted")
```

### DeleteMessage (via ChatBotClient)

```go
bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
if err != nil {
    log.Fatal(err)
}

err = bot.Delete(ctx, "message-id")
```

### Auto-Moderation Example

Delete messages containing banned words:

```go
bannedWords := []string{"spam", "scam", "buy followers"}

poller.OnMessage(func(msg *streaming.LiveChatMessage) {
    text := strings.ToLower(msg.Snippet.DisplayMessage)

    for _, word := range bannedWords {
        if strings.Contains(text, word) {
            err := poller.DeleteMessage(ctx, msg.ID)
            if err != nil {
                log.Printf("Failed to delete message: %v", err)
            } else {
                log.Printf("Deleted message from %s", msg.AuthorDetails.DisplayName)
            }
            break
        }
    }
})
```

### Who Can Delete

| User | Own Messages | Others' Messages |
|------|--------------|------------------|
| Broadcast owner | Yes | Yes |
| Moderator | Yes | Yes |
| Regular viewer | Yes | No |

### Handling Deleted Messages

When a message is deleted, an event is sent to all listeners:

```go
poller.OnDelete(func(messageID string) {
    log.Printf("Message %s was deleted", messageID)
})
```
