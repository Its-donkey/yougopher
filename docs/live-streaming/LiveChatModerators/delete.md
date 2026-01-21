---
layout: default
title: LiveChatModerators.delete
description: Removes a moderator from a live chat
---

Removes a moderator from a live chat.

## Request

### HTTP Request

```
DELETE https://www.googleapis.com/youtube/v3/liveChat/moderators
```

### Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `id` | Yes | string | The ID of the moderator resource to remove (from liveChatModerators.list or insert response). |

### Authorization

Requires OAuth 2.0 authorization with the following scope:

- `https://www.googleapis.com/auth/youtube.force-ssl`

The authenticated user must be:
- The broadcast owner

### Request Body

Do not provide a request body when calling this method.

## Response

If successful, this method returns an empty response body with HTTP status code 204.

## Errors

| Status Code | Error | Description |
|-------------|-------|-------------|
| 400 | `idRequired` | The moderator ID is required. |
| 401 | `unauthorized` | The request is not authorized. |
| 403 | `forbidden` | Only the broadcast owner can remove moderators. |
| 403 | `liveChatEnded` | The live chat has ended. |
| 404 | `liveChatModeratorNotFound` | The moderator does not exist. |

## Quota Cost

This method consumes **50 quota units**.

---

## Yougopher Implementation

### RemoveModerator

Remove a moderator using the moderator resource ID.

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

err := poller.RemoveModerator(ctx, "moderator-id")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Moderator removed")
```

### Remove Moderator via ChatBotClient

```go
bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
if err != nil {
    log.Fatal(err)
}

err = bot.RemoveModerator(ctx, "moderator-id")
```

### Find and Remove by Channel ID

To remove a moderator when you only have their channel ID:

```go
func removeModerator(ctx context.Context, poller *streaming.LiveChatPoller, channelID string) error {
    // Find the moderator
    pageToken := ""
    for {
        resp, err := poller.ListModerators(ctx, &streaming.ListModeratorsParams{
            MaxResults: 50,
            PageToken:  pageToken,
        })
        if err != nil {
            return err
        }

        for _, mod := range resp.Items {
            if mod.Snippet.ModeratorDetails.ChannelID == channelID {
                // Found the moderator, remove them
                return poller.RemoveModerator(ctx, mod.ID)
            }
        }

        if resp.NextPageToken == "" {
            break
        }
        pageToken = resp.NextPageToken
    }

    return fmt.Errorf("moderator not found: %s", channelID)
}
```

### Remove All Moderators

```go
func removeAllModerators(ctx context.Context, poller *streaming.LiveChatPoller) error {
    pageToken := ""
    for {
        resp, err := poller.ListModerators(ctx, &streaming.ListModeratorsParams{
            MaxResults: 50,
            PageToken:  pageToken,
        })
        if err != nil {
            return err
        }

        for _, mod := range resp.Items {
            err := poller.RemoveModerator(ctx, mod.ID)
            if err != nil {
                log.Printf("Failed to remove %s: %v",
                    mod.Snippet.ModeratorDetails.DisplayName, err)
                continue
            }
            log.Printf("Removed moderator: %s", mod.Snippet.ModeratorDetails.DisplayName)
        }

        if resp.NextPageToken == "" {
            break
        }
        pageToken = resp.NextPageToken
    }
    return nil
}
```

### Notes

- The moderator ID is different from the user's channel ID.
- You need the moderator resource ID from `list` or `insert` response.
- If you only have the channel ID, use `list` to find the moderator ID first.
- Removing a moderator doesn't ban them; they can still chat normally.
