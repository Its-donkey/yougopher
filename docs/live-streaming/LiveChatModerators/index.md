---
layout: default
title: LiveChatModerators
description: Manage moderators for YouTube live chat
---

A `liveChatModerator` resource represents a moderator for a YouTube live chat.

Moderators can delete messages and ban users from the chat. Only the broadcast owner can add or remove moderators.

## Methods

| Method | HTTP Request | Description |
|--------|--------------|-------------|
| [list](list) | `GET /liveChat/moderators` | Lists the moderators for a live chat. |
| [insert](insert) | `POST /liveChat/moderators` | Adds a moderator to a live chat. |
| [delete](delete) | `DELETE /liveChat/moderators` | Removes a moderator from a live chat. |

## Resource Representation

```json
{
  "kind": "youtube#liveChatModerator",
  "etag": string,
  "id": string,
  "snippet": {
    "liveChatId": string,
    "moderatorDetails": {
      "channelId": string,
      "channelUrl": string,
      "displayName": string,
      "profileImageUrl": string
    }
  }
}
```

## Properties

| Property | Type | Description |
|----------|------|-------------|
| `kind` | string | Identifies the resource type. Value: `youtube#liveChatModerator`. |
| `etag` | string | The ETag of the resource. |
| `id` | string | The ID that YouTube assigns to uniquely identify the moderator. |

### snippet

| Property | Type | Description |
|----------|------|-------------|
| `snippet.liveChatId` | string | The ID of the live chat where the user is a moderator. |
| `snippet.moderatorDetails.channelId` | string | The moderator's channel ID. |
| `snippet.moderatorDetails.channelUrl` | string | The URL of the moderator's channel. |
| `snippet.moderatorDetails.displayName` | string | The moderator's display name. |
| `snippet.moderatorDetails.profileImageUrl` | string | The URL of the moderator's profile image. |

---

## Yougopher Type

```go
type LiveChatModerator struct {
    Kind    string              `json:"kind,omitempty"`
    ETag    string              `json:"etag,omitempty"`
    ID      string              `json:"id,omitempty"`
    Snippet *ModeratorSnippet   `json:"snippet,omitempty"`
}
```

### Moderator Permissions

| Action | Broadcast Owner | Moderator | Regular Viewer |
|--------|-----------------|-----------|----------------|
| Delete any message | Yes | Yes | No |
| Delete own message | Yes | Yes | Yes |
| Ban/timeout users | Yes | Yes | No |
| Add moderators | Yes | No | No |
| Remove moderators | Yes | No | No |
| List moderators | Yes | No | No |
| Change chat mode | Yes | No | No |

### Notes

- Moderators are specific to a single live chat, not the channel.
- The same user can be a moderator in multiple chats.
- The broadcast owner is not included in the moderator list.
- Removing a moderator does not ban them; they can still chat normally.
