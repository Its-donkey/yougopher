// Package data provides access to the YouTube Data API v3.
//
// This package implements the core data resources needed for YouTube
// integrations: videos, channels, playlists, search, comments, and
// subscriptions.
//
// # Videos
//
// Retrieve video information including live streaming details:
//
//	video, err := data.GetVideo(ctx, client, "video-id")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(video.Snippet.Title)
//
//	// Check if live
//	if video.IsLive() {
//		liveChatID := video.LiveStreamingDetails.ActiveLiveChatID
//	}
//
// Multiple videos:
//
//	resp, err := data.GetVideos(ctx, client, &data.GetVideosParams{
//		IDs:   []string{"video1", "video2"},
//		Parts: []string{"snippet", "liveStreamingDetails"},
//	})
//
// # Channels
//
// Retrieve channel information:
//
//	channel, err := data.GetChannel(ctx, client, "channel-id")
//	fmt.Printf("Channel: %s (%s subscribers)\n",
//		channel.Snippet.Title, channel.Statistics.SubscriberCount)
//
// Get the authenticated user's channel:
//
//	myChannel, err := data.GetMyChannel(ctx, client)
//
// # Playlists
//
// Retrieve playlists and playlist items:
//
//	playlist, err := data.GetPlaylist(ctx, client, "playlist-id")
//	fmt.Printf("Playlist: %s (%d items)\n",
//		playlist.Snippet.Title, playlist.ContentDetails.ItemCount)
//
//	items, err := data.GetPlaylistItems(ctx, client, &data.GetPlaylistItemsParams{
//		PlaylistID: "playlist-id",
//		MaxResults: 50,
//	})
//
// # Search
//
// Search for videos, channels, and playlists.
// WARNING: Search costs 100 quota units per call!
//
//	// Search for live streams
//	results, err := data.SearchLiveStreams(ctx, client, "gaming", 10)
//
//	// Search for videos
//	results, err := data.SearchVideos(ctx, client, "golang tutorial", 10)
//
//	// Advanced search
//	results, err := data.Search(ctx, client, &data.SearchParams{
//		Query:      "music",
//		Type:       data.SearchTypeVideo,
//		Order:      data.SearchOrderViewCount,
//		MaxResults: 25,
//	})
//
// # Comments
//
// Retrieve comment threads and replies:
//
//	// Get video comments
//	comments, err := data.GetVideoComments(ctx, client, "video-id", 20)
//	for _, thread := range comments.Items {
//		topComment := thread.TopLevelComment()
//		fmt.Printf("%s: %s\n",
//			topComment.Snippet.AuthorDisplayName,
//			topComment.Snippet.TextDisplay)
//	}
//
//	// Get replies
//	replies, err := data.GetCommentReplies(ctx, client, "parent-id", 10)
//
// # Subscriptions
//
// Retrieve subscription information:
//
//	// Get my subscriptions
//	subs, err := data.GetMySubscriptions(ctx, client, 50)
//
//	// Check if subscribed
//	subscribed, err := data.IsSubscribedTo(ctx, client, "channel-id")
//
// # LiveChatID
//
// Get the live chat ID from a video (for connecting a chat bot):
//
//	liveChatID, err := data.GetLiveChatID(ctx, client, videoID)
//	bot, err := streaming.NewChatBotClient(client, authClient, liveChatID)
//
// # Quota Costs
//
// Most endpoints cost 1 quota unit per call. The exception is search.list
// which costs 100 quota units per call - use sparingly!
//
//	| Operation           | Quota Cost |
//	|---------------------|------------|
//	| videos.list         | 1          |
//	| channels.list       | 1          |
//	| playlists.list      | 1          |
//	| playlistItems.list  | 1          |
//	| search.list         | 100        |
//	| commentThreads.list | 1          |
//	| comments.list       | 1          |
//	| subscriptions.list  | 1          |
package data
