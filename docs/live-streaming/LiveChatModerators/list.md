---
layout: default
title: LiveChatModerators.list
description: Lists the moderators for a live chat
---

Lists the moderators for a live chat.

## Request

### HTTP Request

```
GET https://www.googleapis.com/youtube/v3/liveChat/moderators
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `liveChatId` | Yes | string | The ID of the live chat. |
| `part` | Yes | string | Must include `id` and/or `snippet`. |
| `maxResults` | No | integer | Maximum number of items to return (1-50). Default: 5. |
| `pageToken` | No | string | Token for pagination. |

### Authorization

Requires OAuth 2.0 authorization with the following scope:

- `https://www.googleapis.com/auth/youtube.force-ssl`

The authenticated user must be:
- The broadcast owner

### Request Body

Do not provide a request body when calling this method.

## Response

```json
{
  "kind": "youtube#liveChatModeratorListResponse",
  "etag": "string",
  "nextPageToken": "string",
  "prevPageToken": "string",
  "pageInfo": {
    "totalResults": integer,
    "resultsPerPage": integer
  },
  "items": [
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
  ]
}
```

### Properties

| Property | Type | Description |
|----------|------|-------------|
| `id` | string | The moderator resource ID (use for removal). |
| `snippet.liveChatId` | string | The live chat ID. |
| `snippet.moderatorDetails.channelId` | string | Moderator's channel ID. |
| `snippet.moderatorDetails.displayName` | string | Moderator's display name. |
| `snippet.moderatorDetails.profileImageUrl` | string | Moderator's profile image. |

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `liveChatIdRequired` | The liveChatId is required. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | Only the broadcast owner can list moderators. |
| 403 | `liveChatEnded` | The live chat has ended. |
| 404 | `liveChatNotFound` | The live chat does not exist. |

## Quota Cost

This method consumes **50 quota units**.

---

## Yougopher Implementation

### ListModerators

List all moderators for a live chat.

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

resp, err := poller.ListModerators(ctx, &streaming.ListModeratorsParams{
    MaxResults: 50,
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Total moderators: %d\n", resp.PageInfo.TotalResults)
for _, mod := range resp.Items {
    fmt.Printf("- %s (%s)\n",
        mod.Snippet.ModeratorDetails.DisplayName,
        mod.Snippet.ModeratorDetails.ChannelID)
}
```

### Pagination

```go
var allModerators []*streaming.LiveChatModerator
pageToken := ""

for {
    resp, err := poller.ListModerators(ctx, &streaming.ListModeratorsParams{
        MaxResults: 50,
        PageToken:  pageToken,
    })
    if err != nil {
        log.Fatal(err)
    }

    allModerators = append(allModerators, resp.Items...)

    if resp.NextPageToken == "" {
        break
    }
    pageToken = resp.NextPageToken
}

fmt.Printf("Found %d moderators\n", len(allModerators))
```

### Check if User is Moderator

```go
func isModerator(ctx context.Context, poller *streaming.LiveChatPoller, channelID string) (bool, error) {
    pageToken := ""

    for {
        resp, err := poller.ListModerators(ctx, &streaming.ListModeratorsParams{
            MaxResults: 50,
            PageToken:  pageToken,
        })
        if err != nil {
            return false, err
        }

        for _, mod := range resp.Items {
            if mod.Snippet.ModeratorDetails.ChannelID == channelID {
                return true, nil
            }
        }

        if resp.NextPageToken == "" {
            break
        }
        pageToken = resp.NextPageToken
    }

    return false, nil
}
```

### Notes

- Only the broadcast owner can list moderators.
- The broadcast owner is not included in the moderator list.
- Moderators are added per-chat, not per-channel.
