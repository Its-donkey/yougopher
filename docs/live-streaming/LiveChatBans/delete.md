---
layout: default
title: UnbanUser
description: Remove a ban from live chat
---

Removes a ban from a live chat (unbans a user).

**Quota Cost:** 50 units

## UnbanUser

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

err := poller.UnbanUser(ctx, "ban-id")
if err != nil {
    log.Fatal(err)
}

fmt.Println("User unbanned")
```

## Via ChatBotClient

```go
bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
if err != nil {
    log.Fatal(err)
}

err = bot.Unban(ctx, "ban-id")
```

## Getting the Ban ID

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

## Tracking Bans

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

## Notes

- Unbanning a timed-out user removes the timeout immediately
- You need the ban ID, not the user's channel ID, to unban
- The ban ID is only available from the original ban response
- If you don't have the ban ID, you cannot unban via the API

## Common Errors

| Error | Description |
|-------|-------------|
| `NotFoundError` | Ban doesn't exist |
| `ForbiddenError` | No permission to unban |
| `liveChatEnded` | Chat has ended |
