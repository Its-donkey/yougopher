---
layout: default
title: LiveChatMessages
description: Live chat message operations for YouTube streaming
---

A `liveChatMessage` resource represents a chat message in a YouTube live chat. The resource can represent different types of messages, including text messages, Super Chats, and membership events.

## Methods

| Method | HTTP Request | Description |
|--------|--------------|-------------|
| [list](list) | `GET /liveChat/messages` | Lists live chat messages (polling). |
| [streamList](stream-list) | `GET /liveChat/messages/stream` | SSE streaming for real-time messages. |
| [insert](insert) | `POST /liveChat/messages` | Sends a message to the live chat. |
| [delete](delete) | `DELETE /liveChat/messages` | Deletes a message from the live chat. |
| [transition](transition) | `POST /liveChat/messages/transition` | Changes the chat mode. |

## Resource Representation

```json
{
  "kind": "youtube#liveChatMessage",
  "etag": string,
  "id": string,
  "snippet": {
    "type": string,
    "liveChatId": string,
    "authorChannelId": string,
    "publishedAt": datetime,
    "hasDisplayContent": boolean,
    "displayMessage": string,
    "textMessageDetails": {
      "messageText": string
    },
    "messageDeletedDetails": {
      "deletedMessageId": string
    },
    "userBannedDetails": {
      "bannedUserDetails": {
        "channelId": string,
        "channelUrl": string,
        "displayName": string,
        "profileImageUrl": string
      },
      "banType": string,
      "banDurationSeconds": unsigned long
    },
    "superChatDetails": {
      "amountMicros": unsigned long,
      "currency": string,
      "amountDisplayString": string,
      "userComment": string,
      "tier": unsigned integer
    },
    "superStickerDetails": {
      "superStickerMetadata": {
        "stickerId": string,
        "altText": string,
        "language": string
      },
      "amountMicros": unsigned long,
      "currency": string,
      "amountDisplayString": string,
      "tier": unsigned integer
    },
    "newSponsorDetails": {
      "memberLevelName": string,
      "isUpgrade": boolean
    },
    "memberMilestoneChatDetails": {
      "userComment": string,
      "memberMonth": unsigned integer,
      "memberLevelName": string
    }
  },
  "authorDetails": {
    "channelId": string,
    "channelUrl": string,
    "displayName": string,
    "profileImageUrl": string,
    "isVerified": boolean,
    "isChatOwner": boolean,
    "isChatSponsor": boolean,
    "isChatModerator": boolean
  }
}
```

## Properties

| Property | Type | Description |
|----------|------|-------------|
| `kind` | string | Identifies the resource type. Value: `youtube#liveChatMessage`. |
| `etag` | string | The ETag of the resource. |
| `id` | string | The ID that YouTube assigns to uniquely identify the message. |

### snippet

| Property | Type | Description |
|----------|------|-------------|
| `snippet.type` | string | The type of message (see Message Types below). |
| `snippet.liveChatId` | string | The ID of the live chat this message belongs to. |
| `snippet.authorChannelId` | string | The channel ID of the message author. |
| `snippet.publishedAt` | datetime | The date and time the message was published. |
| `snippet.hasDisplayContent` | boolean | Whether the message has displayable content. |
| `snippet.displayMessage` | string | The message content to display to users. |

### authorDetails

| Property | Type | Description |
|----------|------|-------------|
| `authorDetails.channelId` | string | The author's channel ID. |
| `authorDetails.channelUrl` | string | The URL of the author's channel. |
| `authorDetails.displayName` | string | The author's display name. |
| `authorDetails.profileImageUrl` | string | The URL of the author's profile image. |
| `authorDetails.isVerified` | boolean | Whether the author is verified. |
| `authorDetails.isChatOwner` | boolean | Whether the author is the broadcast owner. |
| `authorDetails.isChatSponsor` | boolean | Whether the author is a channel member. |
| `authorDetails.isChatModerator` | boolean | Whether the author is a moderator. |

## Message Types

| Type | Description |
|------|-------------|
| `textMessageEvent` | Regular text message. |
| `superChatEvent` | Super Chat donation. |
| `superStickerEvent` | Super Sticker. |
| `newSponsorEvent` | New channel membership. |
| `memberMilestoneChatEvent` | Membership milestone celebration. |
| `membershipGiftingEvent` | Gifted memberships (sender). |
| `giftMembershipReceivedEvent` | Received gift (recipient). |
| `messageDeletedEvent` | Message was deleted. |
| `userBannedEvent` | User was banned or timed out. |
| `pollEvent` | Poll-related event. |

---

## Yougopher Type

```go
type LiveChatMessage struct {
    Kind          string                `json:"kind,omitempty"`
    ETag          string                `json:"etag,omitempty"`
    ID            string                `json:"id,omitempty"`
    Snippet       *LiveChatSnippet      `json:"snippet,omitempty"`
    AuthorDetails *LiveChatAuthorDetails `json:"authorDetails,omitempty"`
}
```

### Helper Methods

```go
msg.Type()              // Returns the message type constant
msg.IsFromOwner()       // Returns true if from broadcast owner
msg.IsFromModerator()   // Returns true if from a moderator
msg.IsFromMember()      // Returns true if from a channel member
```

### Polling vs SSE Streaming

| Feature | Polling (list) | SSE (streamList) |
|---------|---------------|------------------|
| Latency | Higher (poll interval) | Lower (~instant) |
| Connection | Repeated requests | Long-lived HTTP |
| Quota | 5 units per poll | 5 units per connection |
| Reconnection | Manual | Automatic with backoff |
| Use case | Custom polling logic | Real-time bots |
