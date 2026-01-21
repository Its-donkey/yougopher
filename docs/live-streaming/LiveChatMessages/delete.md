---
layout: default
title: DeleteMessage
description: Delete a message from live chat
---

Deletes a message from a live chat.

**Quota Cost:** 50 units

## DeleteMessage (via Poller)

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

err := poller.DeleteMessage(ctx, "message-id")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Message deleted")
```

## DeleteMessage (via ChatBotClient)

```go
bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
if err != nil {
    log.Fatal(err)
}

err = bot.Delete(ctx, "message-id")
```

## Auto-Moderation Example

Delete messages containing banned words:

```go
bannedWords := []string{"spam", "scam", "buy followers"}

poller.OnMessage(func(msg *streaming.LiveChatMessage) {
    text := strings.ToLower(msg.Message())

    for _, word := range bannedWords {
        if strings.Contains(text, word) {
            err := poller.DeleteMessage(ctx, msg.ID)
            if err != nil {
                log.Printf("Failed to delete: %v", err)
            } else {
                log.Printf("Deleted message from %s", msg.AuthorDetails.DisplayName)
            }
            break
        }
    }
})
```

## Who Can Delete

| User | Own Messages | Others' Messages |
|------|--------------|------------------|
| Broadcast owner | Yes | Yes |
| Moderator | Yes | Yes |
| Regular viewer | Yes | No |

## Handling Deleted Messages

When a message is deleted, all listeners receive an event:

```go
poller.OnDelete(func(messageID string) {
    log.Printf("Message %s was deleted", messageID)
})
```

## Common Errors

| Error | Description |
|-------|-------------|
| `NotFoundError` | Message doesn't exist |
| `ForbiddenError` | No permission to delete |
| `liveChatEnded` | Chat has ended |
