---
layout: default
title: LiveChatBans.delete
description: Removes a ban from a live chat (unbans a user)
---

Removes a ban from a live chat (unbans a user).

## Request

### HTTP Request

```
DELETE https://www.googleapis.com/youtube/v3/liveChat/bans
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `id` | Yes | string | The ID of the ban to remove (from liveChatBans.insert response). |

### Authorization

Requires OAuth 2.0 authorization with the following scope:

- `https://www.googleapis.com/auth/youtube.force-ssl`

The authenticated user must be:
- The broadcast owner, OR
- A chat moderator

### Request Body

Do not provide a request body when calling this method.

## Response

If successful, this method returns an empty response body with HTTP status code 204.

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `idRequired` | The ban ID is required. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | The user cannot unban in this chat. |
| 403 | `liveChatEnded` | The live chat has ended. |
| 404 | `liveChatBanNotFound` | The ban does not exist. |

## Quota Cost

This method consumes **50 quota units**.

---

## Yougopher Implementation

### UnbanUser

Remove a ban using the ban ID.

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

err := poller.UnbanUser(ctx, "ban-id")
if err != nil {
    log.Fatal(err)
}

fmt.Println("User unbanned")
```

### Unban via ChatBotClient

```go
bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
if err != nil {
    log.Fatal(err)
}

err = bot.Unban(ctx, "ban-id")
```

### Getting the Ban ID

The ban ID is returned when you ban a user:

```go
// Ban the user
ban, err := poller.BanUser(ctx, channelID)
if err != nil {
    log.Fatal(err)
}

// Save the ban ID for later
banID := ban.ID

// Later, unban the user
err = poller.UnbanUser(ctx, banID)
```

### Tracking Bans

Keep track of bans for easy unbanning:

```go
type BanTracker struct {
    bans map[string]string // channelID -> banID
    mu   sync.RWMutex
}

func (t *BanTracker) Ban(ctx context.Context, poller *streaming.LiveChatPoller, channelID string) error {
    ban, err := poller.BanUser(ctx, channelID)
    if err != nil {
        return err
    }

    t.mu.Lock()
    t.bans[channelID] = ban.ID
    t.mu.Unlock()

    return nil
}

func (t *BanTracker) Unban(ctx context.Context, poller *streaming.LiveChatPoller, channelID string) error {
    t.mu.RLock()
    banID, ok := t.bans[channelID]
    t.mu.RUnlock()

    if !ok {
        return fmt.Errorf("no ban found for channel %s", channelID)
    }

    err := poller.UnbanUser(ctx, banID)
    if err != nil {
        return err
    }

    t.mu.Lock()
    delete(t.bans, channelID)
    t.mu.Unlock()

    return nil
}
```

### Notes

- Unbanning a timed-out user removes the timeout immediately.
- You need the ban ID, not the user's channel ID, to unban.
- The ban ID is only available from the original ban response.
- If you don't have the ban ID, you cannot unban via the API.
