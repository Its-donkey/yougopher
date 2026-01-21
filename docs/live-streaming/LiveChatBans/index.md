---
layout: default
title: LiveChatBans
description: Ban and timeout users in YouTube live chat
---

A `liveChatBan` resource represents a ban that prevents a YouTube user from posting messages in a live chat.

Bans can be either permanent or temporary (timeout). Moderators and the broadcast owner can ban users from the chat.

## Methods

| Method | HTTP Request | Description |
|--------|--------------|-------------|
| [insert](insert) | `POST /liveChat/bans` | Bans a user from a live chat (permanent or timeout). |
| [delete](delete) | `DELETE /liveChat/bans` | Removes a ban (unbans a user). |

## Resource Representation

```json
{
  "kind": "youtube#liveChatBan",
  "etag": string,
  "id": string,
  "snippet": {
    "liveChatId": string,
    "type": string,
    "banDurationSeconds": unsigned long,
    "bannedUserDetails": {
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
| `kind` | string | Identifies the resource type. Value: `youtube#liveChatBan`. |
| `etag` | string | The ETag of the resource. |
| `id` | string | The ID that YouTube assigns to uniquely identify the ban. |

### snippet

| Property | Type | Description |
|----------|------|-------------|
| `snippet.liveChatId` | string | The ID of the live chat where the ban applies. |
| `snippet.type` | string | The type of ban: `permanent` or `temporary`. |
| `snippet.banDurationSeconds` | long | The duration of a temporary ban in seconds. |
| `snippet.bannedUserDetails.channelId` | string | The banned user's channel ID. |
| `snippet.bannedUserDetails.channelUrl` | string | The URL of the banned user's channel. |
| `snippet.bannedUserDetails.displayName` | string | The banned user's display name. |
| `snippet.bannedUserDetails.profileImageUrl` | string | The URL of the banned user's profile image. |

## Ban Types

| Type | Description |
|------|-------------|
| `permanent` | User is permanently banned from the chat. |
| `temporary` | User is temporarily banned (timeout). Requires `banDurationSeconds`. |

---

## Yougopher Type

```go
type LiveChatBan struct {
    Kind    string          `json:"kind,omitempty"`
    ETag    string          `json:"etag,omitempty"`
    ID      string          `json:"id,omitempty"`
    Snippet *BanSnippet     `json:"snippet,omitempty"`
}
```

### Constants

```go
streaming.BanTypePermanent // "permanent"
streaming.BanTypeTemporary // "temporary"
```

### Common Timeout Durations

| Duration | Seconds | Use Case |
|----------|---------|----------|
| 1 minute | 60 | Minor warning |
| 5 minutes | 300 | Standard timeout |
| 10 minutes | 600 | Moderate offense |
| 1 hour | 3600 | Serious offense |
| 24 hours | 86400 | Severe offense |

### Who Can Ban

| User | Can Ban |
|------|---------|
| Broadcast owner | Yes |
| Moderator | Yes |
| Regular viewer | No |
