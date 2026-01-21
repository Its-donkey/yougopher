---
layout: default
title: Data API
description: Access YouTube Data API v3 for videos, channels, playlists, comments, and search.
---

## Overview

The `youtube/data` package provides access to the YouTube Data API v3 resources. This package implements the core data resources needed for YouTube integrations:

| Resource | Functions | Quota Cost |
|----------|-----------|------------|
| Videos | `GetVideos`, `GetVideo`, `GetLiveChatID` | 1 unit |
| Channels | `GetChannels`, `GetChannel`, `GetMyChannel` | 1 unit |
| Playlists | `GetPlaylists`, `GetPlaylist`, `GetMyPlaylists` | 1 unit |
| PlaylistItems | `GetPlaylistItems` | 1 unit |
| Search | `Search`, `SearchVideos`, `SearchLiveStreams`, `SearchChannels` | **100 units** |
| CommentThreads | `GetCommentThreads`, `GetVideoComments` | 1 unit |
| Comments | `GetComments`, `GetCommentReplies` | 1 unit |
| Subscriptions | `GetSubscriptions`, `GetMySubscriptions`, `GetChannelSubscriptions`, `IsSubscribedTo` | 1 unit |

## Videos

Retrieve video information including live streaming details.

```go
// Get a single video
video, err := data.GetVideo(ctx, client, "dQw4w9WgXcQ")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Title: %s\n", video.Snippet.Title)
fmt.Printf("Views: %s\n", video.Statistics.ViewCount)

// Check if live
if video.IsLive() {
    fmt.Println("Currently streaming!")
    liveChatID := video.LiveStreamingDetails.ActiveLiveChatID
}

// Get multiple videos
resp, err := data.GetVideos(ctx, client, &data.GetVideosParams{
    IDs:   []string{"video1", "video2", "video3"},
    Parts: []string{"snippet", "statistics", "liveStreamingDetails"},
})

// Get live chat ID directly
liveChatID, err := data.GetLiveChatID(ctx, client, "video-id")
```

### Video Helper Methods

| Method | Description |
|--------|-------------|
| `IsLive()` | Returns true if currently streaming |
| `IsUpcoming()` | Returns true if scheduled but not started |
| `HasActiveLiveChat()` | Returns true if live chat is available |

## Channels

Retrieve channel information.

```go
// Get a single channel
channel, err := data.GetChannel(ctx, client, "UC_x5XG1OV2P6uZZ5FSM9Ttw")
fmt.Printf("Channel: %s\n", channel.Snippet.Title)
fmt.Printf("Subscribers: %s\n", channel.Statistics.SubscriberCount)

// Get the authenticated user's channel
myChannel, err := data.GetMyChannel(ctx, client)

// Get uploads playlist ID
uploadsPlaylistID := channel.UploadsPlaylistID()
```

## Playlists

Retrieve playlist and playlist item information.

```go
// Get a playlist
playlist, err := data.GetPlaylist(ctx, client, "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf")
fmt.Printf("Playlist: %s (%d items)\n",
    playlist.Snippet.Title,
    playlist.ContentDetails.ItemCount)

// Get my playlists
myPlaylists, err := data.GetMyPlaylists(ctx, client, &data.GetPlaylistsParams{
    MaxResults: 50,
})

// Get playlist items
items, err := data.GetPlaylistItems(ctx, client, &data.GetPlaylistItemsParams{
    PlaylistID: "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
    MaxResults: 50,
})

for _, item := range items.Items {
    fmt.Printf("- %s (video: %s)\n", item.Snippet.Title, item.VideoID())
}
```

## Search

Search for videos, channels, and playlists.

**WARNING: Search costs 100 quota units per call! Use sparingly.**

```go
// Search for videos
results, err := data.SearchVideos(ctx, client, "golang tutorial", 10)

// Search for live streams
liveStreams, err := data.SearchLiveStreams(ctx, client, "gaming", 10)

// Search for channels
channels, err := data.SearchChannels(ctx, client, "programming", 5)

// Advanced search with filters
results, err := data.Search(ctx, client, &data.SearchParams{
    Query:         "music",
    Type:          data.SearchTypeVideo,
    Order:         data.SearchOrderViewCount,
    VideoDuration: "medium",  // 4-20 minutes
    MaxResults:    25,
})

// Process results
for _, result := range results.Items {
    if result.IsVideo() {
        fmt.Printf("Video: %s (ID: %s)\n", result.Snippet.Title, result.ID.VideoID)
    } else if result.IsChannel() {
        fmt.Printf("Channel: %s\n", result.Snippet.Title)
    }

    if result.IsLive() {
        fmt.Println("  -> Currently live!")
    }
}
```

### Search Constants

```go
// Type filters
data.SearchTypeVideo    // "video"
data.SearchTypeChannel  // "channel"
data.SearchTypePlaylist // "playlist"

// Event types (for live content)
data.SearchEventTypeLive      // "live"
data.SearchEventTypeUpcoming  // "upcoming"
data.SearchEventTypeCompleted // "completed"

// Sort order
data.SearchOrderDate       // "date"
data.SearchOrderRating     // "rating"
data.SearchOrderRelevance  // "relevance"
data.SearchOrderTitle      // "title"
data.SearchOrderVideoCount // "videoCount"
data.SearchOrderViewCount  // "viewCount"
```

## Comments

Retrieve comment threads and replies.

```go
// Get comments for a video
comments, err := data.GetVideoComments(ctx, client, "video-id", 20)

for _, thread := range comments.Items {
    topComment := thread.TopLevelComment()
    fmt.Printf("%s: %s\n",
        topComment.Snippet.AuthorDisplayName,
        topComment.Snippet.TextDisplay)

    if thread.ReplyCount() > 0 {
        fmt.Printf("  (%d replies)\n", thread.ReplyCount())
    }
}

// Get replies to a specific comment
replies, err := data.GetCommentReplies(ctx, client, "parent-comment-id", 10)

// Advanced comment thread query
threads, err := data.GetCommentThreads(ctx, client, &data.GetCommentThreadsParams{
    VideoID:    "video-id",
    Order:      data.CommentOrderTime,
    MaxResults: 100,
})
```

### Comment Helper Methods

| Method | Description |
|--------|-------------|
| `TopLevelComment()` | Returns the parent comment of a thread |
| `ReplyCount()` | Returns the number of replies |
| `AuthorID()` | Returns the comment author's channel ID |
| `IsReply()` | Returns true if this is a reply to another comment |

### Moderation Status Constants

```go
data.ModerationStatusHeldForReview // "heldForReview"
data.ModerationStatusLikelySpam    // "likelySpam"
data.ModerationStatusPublished     // "published"
data.ModerationStatusRejected      // "rejected"
```

## Subscriptions

Retrieve subscription information.

```go
// Get my subscriptions
subs, err := data.GetMySubscriptions(ctx, client, 50)

for _, sub := range subs.Items {
    fmt.Printf("Subscribed to: %s\n", sub.Snippet.Title)
    if sub.HasNewContent() {
        fmt.Printf("  -> %d new items!\n", sub.ContentDetails.NewItemCount)
    }
}

// Check if subscribed to a channel
subscribed, err := data.IsSubscribedTo(ctx, client, "channel-id")
if subscribed {
    fmt.Println("You are subscribed!")
}

// Get a channel's public subscriptions
channelSubs, err := data.GetChannelSubscriptions(ctx, client, "channel-id", 50)

// Advanced subscription query
resp, err := data.GetSubscriptions(ctx, client, &data.GetSubscriptionsParams{
    Mine:         true,
    ForChannelID: "target-channel-id",  // Filter to specific channel
    Order:        data.SubscriptionOrderAlphabetical,
    MaxResults:   50,
})
```

### Subscription Order Constants

```go
data.SubscriptionOrderAlphabetical // "alphabetical"
data.SubscriptionOrderRelevance    // "relevance"
data.SubscriptionOrderUnread       // "unread"
```

## Pagination

All list functions return responses with pagination support.

```go
var allVideos []*data.Video
pageToken := ""

for {
    resp, err := data.GetVideos(ctx, client, &data.GetVideosParams{
        IDs:       videoIDs,
        PageToken: pageToken,
    })
    if err != nil {
        return err
    }

    allVideos = append(allVideos, resp.Items...)

    if resp.NextPageToken == "" {
        break
    }
    pageToken = resp.NextPageToken
}
```

## Error Handling

The package uses typed errors for common scenarios:

```go
video, err := data.GetVideo(ctx, client, "invalid-id")
if err != nil {
    var notFound *core.NotFoundError
    if errors.As(err, &notFound) {
        fmt.Printf("%s not found: %s\n", notFound.ResourceType, notFound.ResourceID)
    }

    var apiErr *core.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API error %d: %s\n", apiErr.StatusCode, apiErr.Message)
    }
}
```

## Quota Management

YouTube API has a daily quota limit (default 10,000 units). Track usage:

```go
// Check quota before expensive operations
if client.QuotaTracker().Remaining() < 100 {
    log.Println("Warning: Low quota, skipping search")
    return
}

// Search costs 100 units!
results, err := data.Search(ctx, client, params)
```

| Operation | Quota Cost |
|-----------|------------|
| videos.list | 1 |
| channels.list | 1 |
| playlists.list | 1 |
| playlistItems.list | 1 |
| **search.list** | **100** |
| commentThreads.list | 1 |
| comments.list | 1 |
| subscriptions.list | 1 |
