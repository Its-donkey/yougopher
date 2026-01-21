---
layout: default
title: LiveChatModerators.insert
description: Adds a moderator to a live chat
---

Adds a moderator to a live chat.

## Request

### HTTP Request

```
POST https://www.googleapis.com/youtube/v3/liveChat/moderators
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `part` | Yes | string | Must include `snippet`. |

### Authorization

Requires OAuth 2.0 authorization with the following scope:

- `https://www.googleapis.com/auth/youtube.force-ssl`

The authenticated user must be:
- The broadcast owner

### Request Body

```json
{
  "snippet": {
    "liveChatId": "string",
    "moderatorDetails": {
      "channelId": "string"
    }
  }
}
```

### Required Fields

| Field | Description |
|-------|-------------|
| `snippet.liveChatId` | The live chat ID. |
| `snippet.moderatorDetails.channelId` | The channel ID of the user to make moderator. |

## Response

If successful, this method returns a liveChatModerator resource:

```json
{
  "kind": "youtube#liveChatModerator",
  "etag": "string",
  "id": "string",
  "snippet": {
    "liveChatId": "string",
    "moderatorDetails": {
      "channelId": "string",
      "channelUrl": "string",
      "displayName": "string",
      "profileImageUrl": "string"
    }
  }
}
```

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `liveChatIdRequired` | The liveChatId is required. |
| 400 | `channelIdRequired` | The moderator's channel ID is required. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | Only the broadcast owner can add moderators. |
| 403 | `liveChatEnded` | The live chat has ended. |
| 403 | `alreadyModerator` | The user is already a moderator. |
| 404 | `liveChatNotFound` | The live chat does not exist. |
| 404 | `channelNotFound` | The user's channel does not exist. |

## Quota Cost

This method consumes **50 quota units**.

---

## Yougopher Implementation

### AddModerator

Add a moderator to the live chat.

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

mod, err := poller.AddModerator(ctx, "channel-id-to-promote")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Added moderator: %s\n", mod.Snippet.ModeratorDetails.DisplayName)
fmt.Printf("Moderator ID: %s\n", mod.ID) // Save this for removal
```

### Add Moderator via ChatBotClient

```go
bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
if err != nil {
    log.Fatal(err)
}

err = bot.AddModerator(ctx, "channel-id")
```

### Add Multiple Moderators

```go
trustedUsers := []string{
    "channel-id-1",
    "channel-id-2",
    "channel-id-3",
}

for _, channelID := range trustedUsers {
    mod, err := poller.AddModerator(ctx, channelID)
    if err != nil {
        log.Printf("Failed to add %s as moderator: %v", channelID, err)
        continue
    }
    fmt.Printf("Added %s as moderator\n", mod.Snippet.ModeratorDetails.DisplayName)
}
```

### Track Moderators for Removal

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

### Notes

- Only the broadcast owner can add moderators.
- Moderators are specific to a single live chat.
- The same user can be a moderator in multiple chats.
- Save the moderator ID if you want to remove them later.
