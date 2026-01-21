---
layout: default
title: AddModerator
description: Add a moderator to live chat
---

Adds a moderator to a live chat.

**Quota Cost:** 50 units

## AddModerator

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

mod, err := poller.AddModerator(ctx, "channel-id-to-promote")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Added moderator: %s\n", mod.Snippet.ModeratorDetails.DisplayName)
fmt.Printf("Moderator ID: %s\n", mod.ID) // Save this for removal
```

## Via ChatBotClient

```go
bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
if err != nil {
    log.Fatal(err)
}

err = bot.AddModerator(ctx, "channel-id")
```

## Add Multiple Moderators

```go
trustedUsers := []string{
    "channel-id-1",
    "channel-id-2",
    "channel-id-3",
}

for _, channelID := range trustedUsers {
    mod, err := poller.AddModerator(ctx, channelID)
    if err != nil {
        log.Printf("Failed to add %s: %v", channelID, err)
        continue
    }
    fmt.Printf("Added %s as moderator\n", mod.Snippet.ModeratorDetails.DisplayName)
}
```

## Track Moderators for Removal

```go
type ModeratorManager struct {
    moderators map[string]string // channelID -> moderatorID
    mu         sync.RWMutex
}

func (m *ModeratorManager) Add(ctx context.Context, poller *streaming.LiveChatPoller, channelID string) error {
    mod, err := poller.AddModerator(ctx, channelID)
    if err != nil {
        return err
    }

    m.mu.Lock()
    m.moderators[channelID] = mod.ID
    m.mu.Unlock()
    return nil
}
```

## Notes

- Only the broadcast owner can add moderators
- Moderators are specific to a single live chat
- Save the moderator ID if you want to remove them later

## Common Errors

| Error | Description |
|-------|-------------|
| `ForbiddenError` | Not the broadcast owner |
| `alreadyModerator` | User is already a moderator |
| `channelNotFound` | User's channel doesn't exist |
| `liveChatEnded` | Chat has ended |
