---
layout: default
title: RemoveModerator
description: Remove a moderator from live chat
---

Removes a moderator from a live chat.

**Quota Cost:** 50 units

## RemoveModerator

```go
poller := streaming.NewLiveChatPoller(client, liveChatID)

err := poller.RemoveModerator(ctx, "moderator-id")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Moderator removed")
```

## Via ChatBotClient

```go
bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
if err != nil {
    log.Fatal(err)
}

err = bot.RemoveModerator(ctx, "moderator-id")
```

## Find and Remove by Channel ID

To remove a moderator when you only have their channel ID:

```go
func removeModerator(ctx context.Context, poller *streaming.LiveChatPoller, channelID string) error {
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

## Remove All Moderators

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

## Notes

- The moderator ID is different from the user's channel ID
- You need the moderator resource ID from `list` or `insert` response
- Removing a moderator doesn't ban them

## Common Errors

| Error | Description |
|-------|-------------|
| `NotFoundError` | Moderator doesn't exist |
| `ForbiddenError` | Not the broadcast owner |
| `liveChatEnded` | Chat has ended |
