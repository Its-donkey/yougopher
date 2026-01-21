---
layout: default
title: LiveChatModerators
description: Manage moderators for YouTube live chat
---

A `LiveChatModerator` represents a moderator for a YouTube live chat. Moderators can delete messages and ban users.

## Functions

| Function | Quota | Description |
|----------|-------|-------------|
| [ListModerators](list) | 50 | List chat moderators |
| [AddModerator](insert) | 50 | Add a moderator |
| [RemoveModerator](delete) | 50 | Remove a moderator |

## Type Definition

```go
type LiveChatModerator struct {
    Kind    string            `json:"kind,omitempty"`
    ETag    string            `json:"etag,omitempty"`
    ID      string            `json:"id,omitempty"`
    Snippet *ModeratorSnippet `json:"snippet,omitempty"`
}
```

## Moderator Permissions

| Action | Owner | Moderator | Viewer |
|--------|-------|-----------|--------|
| Delete any message | Yes | Yes | No |
| Delete own message | Yes | Yes | Yes |
| Ban/timeout users | Yes | Yes | No |
| Add moderators | Yes | No | No |
| Remove moderators | Yes | No | No |
| List moderators | Yes | No | No |
| Change chat mode | Yes | No | No |

## Notes

- Moderators are specific to a single live chat, not the channel
- The same user can be a moderator in multiple chats
- The broadcast owner is not included in the moderator list
- Removing a moderator doesn't ban them

## Quick Example

```go
import "github.com/Its-donkey/yougopher/youtube/streaming"

poller := streaming.NewLiveChatPoller(client, liveChatID)

// Add a moderator
mod, err := poller.AddModerator(ctx, "channel-id")

// List moderators
resp, err := poller.ListModerators(ctx, &streaming.ListModeratorsParams{
    MaxResults: 50,
})

// Remove a moderator
err := poller.RemoveModerator(ctx, mod.ID)
```
