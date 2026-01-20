// Package data provides access to the YouTube Data API v3.
//
// # Videos
//
// Retrieve video information:
//
//	videos, err := data.GetVideos(ctx, client, &data.GetVideosParams{
//		IDs:   []string{"video-id"},
//		Parts: []string{"snippet", "liveStreamingDetails"},
//	})
//
// # Channels
//
// Retrieve channel information:
//
//	channels, err := data.GetChannels(ctx, client, &data.GetChannelsParams{
//		IDs:   []string{"channel-id"},
//		Parts: []string{"snippet", "statistics"},
//	})
//
// # Search
//
// Search for videos, channels, or playlists:
//
//	results, err := data.Search(ctx, client, &data.SearchParams{
//		Query:     "golang tutorial",
//		Type:      "video",
//		EventType: "live", // Find live streams
//	})
//
// Note: Search costs 100 quota units per call.
//
// # LiveChatID
//
// Get the live chat ID from a video or broadcast:
//
//	video, err := data.GetVideo(ctx, client, videoID)
//	liveChatID := video.LiveStreamingDetails.ActiveLiveChatID
package data
