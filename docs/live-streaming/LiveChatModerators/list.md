---
layout: default
title: ListModerators
description: List moderators for a live chat
---

Lists the moderators for a live chat.

**Quota Cost:** 50 units

## ListModerators

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

## Pagination

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

## Check if User is Moderator

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

## Notes

- Only the broadcast owner can list moderators
- The broadcast owner is not included in the moderator list
- Moderators are added per-chat, not per-channel

## Common Errors

| Error | Description |
|-------|-------------|
| `ForbiddenError` | Not the broadcast owner |
| `liveChatEnded` | Chat has ended |
| `liveChatNotFound` | Chat doesn't exist |
