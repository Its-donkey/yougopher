package data

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Its-donkey/yougopher/youtube/core"
)

// SearchResult represents a search result resource.
type SearchResult struct {
	// Kind is the resource type (youtube#searchResult).
	Kind string `json:"kind,omitempty"`

	// ETag is the entity tag.
	ETag string `json:"etag,omitempty"`

	// ID identifies the result resource.
	ID *SearchResultID `json:"id,omitempty"`

	// Snippet contains basic details about the result.
	Snippet *SearchResultSnippet `json:"snippet,omitempty"`
}

// SearchResultID identifies the resource found by the search.
type SearchResultID struct {
	// Kind is the resource type (e.g., "youtube#video", "youtube#channel", "youtube#playlist").
	Kind string `json:"kind,omitempty"`

	// VideoID is the video ID if this is a video result.
	VideoID string `json:"videoId,omitempty"`

	// ChannelID is the channel ID if this is a channel result.
	ChannelID string `json:"channelId,omitempty"`

	// PlaylistID is the playlist ID if this is a playlist result.
	PlaylistID string `json:"playlistId,omitempty"`
}

// SearchResultSnippet contains basic details about a search result.
type SearchResultSnippet struct {
	// PublishedAt is when the resource was created.
	PublishedAt time.Time `json:"publishedAt,omitempty"`

	// ChannelID is the ID of the channel that created the resource.
	ChannelID string `json:"channelId,omitempty"`

	// Title is the resource's title.
	Title string `json:"title,omitempty"`

	// Description is the resource's description.
	Description string `json:"description,omitempty"`

	// Thumbnails contains thumbnail images.
	Thumbnails *ThumbnailDetails `json:"thumbnails,omitempty"`

	// ChannelTitle is the title of the channel.
	ChannelTitle string `json:"channelTitle,omitempty"`

	// LiveBroadcastContent indicates if the video is live.
	// Values: "live", "upcoming", "none"
	LiveBroadcastContent string `json:"liveBroadcastContent,omitempty"`
}

// SearchListResponse is the response from search.list.
type SearchListResponse struct {
	// Kind is the resource type.
	Kind string `json:"kind,omitempty"`

	// ETag is the entity tag.
	ETag string `json:"etag,omitempty"`

	// NextPageToken is the token for the next page.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// PrevPageToken is the token for the previous page.
	PrevPageToken string `json:"prevPageToken,omitempty"`

	// RegionCode is the region code used for the search.
	RegionCode string `json:"regionCode,omitempty"`

	// PageInfo contains paging information.
	PageInfo *PageInfo `json:"pageInfo,omitempty"`

	// Items contains the search result resources.
	Items []*SearchResult `json:"items,omitempty"`
}

// Search type constants.
const (
	SearchTypeVideo    = "video"
	SearchTypeChannel  = "channel"
	SearchTypePlaylist = "playlist"
)

// Search event type constants (for live streams).
const (
	SearchEventTypeCompleted = "completed"
	SearchEventTypeLive      = "live"
	SearchEventTypeUpcoming  = "upcoming"
)

// Search order constants.
const (
	SearchOrderDate       = "date"
	SearchOrderRating     = "rating"
	SearchOrderRelevance  = "relevance"
	SearchOrderTitle      = "title"
	SearchOrderVideoCount = "videoCount"
	SearchOrderViewCount  = "viewCount"
)

// SearchParams contains parameters for search.list.
type SearchParams struct {
	// Query is the search query string.
	Query string

	// Type filters results by type.
	// Values: "video", "channel", "playlist" (can be comma-separated)
	Type string

	// ChannelID filters results to a specific channel.
	ChannelID string

	// EventType filters live stream results.
	// Values: "completed", "live", "upcoming"
	EventType string

	// Order specifies the sort order.
	// Values: "date", "rating", "relevance", "title", "videoCount", "viewCount"
	Order string

	// PublishedAfter filters results published after this time.
	PublishedAfter *time.Time

	// PublishedBefore filters results published before this time.
	PublishedBefore *time.Time

	// RegionCode filters results to a specific region.
	RegionCode string

	// RelevanceLanguage filters results by language relevance.
	RelevanceLanguage string

	// SafeSearch filters based on content safety.
	// Values: "moderate", "none", "strict"
	SafeSearch string

	// VideoCategoryID filters videos by category.
	VideoCategoryID string

	// VideoDefinition filters videos by definition.
	// Values: "any", "high", "standard"
	VideoDefinition string

	// VideoDuration filters videos by duration.
	// Values: "any", "long", "medium", "short"
	VideoDuration string

	// VideoType filters videos by type.
	// Values: "any", "episode", "movie"
	VideoType string

	// Parts specifies which parts to include in the response.
	// Common values: "snippet"
	Parts []string

	// MaxResults is the maximum number of items to return (1-50).
	MaxResults int

	// PageToken is the token for pagination.
	PageToken string
}

// DefaultSearchParts are the default parts to request for search.
var DefaultSearchParts = []string{"snippet"}

// Search performs a YouTube search.
// WARNING: Each call costs 100 quota units! Use sparingly.
// Quota cost: 100 units per call.
func Search(ctx context.Context, client *core.Client, params *SearchParams) (*SearchListResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}

	parts := params.Parts
	if len(parts) == 0 {
		parts = DefaultSearchParts
	}

	query := url.Values{}
	query.Set("part", strings.Join(parts, ","))

	if params.Query != "" {
		query.Set("q", params.Query)
	}
	if params.Type != "" {
		query.Set("type", params.Type)
	}
	if params.ChannelID != "" {
		query.Set("channelId", params.ChannelID)
	}
	if params.EventType != "" {
		query.Set("eventType", params.EventType)
	}
	if params.Order != "" {
		query.Set("order", params.Order)
	}
	if params.PublishedAfter != nil {
		query.Set("publishedAfter", params.PublishedAfter.Format(time.RFC3339))
	}
	if params.PublishedBefore != nil {
		query.Set("publishedBefore", params.PublishedBefore.Format(time.RFC3339))
	}
	if params.RegionCode != "" {
		query.Set("regionCode", params.RegionCode)
	}
	if params.RelevanceLanguage != "" {
		query.Set("relevanceLanguage", params.RelevanceLanguage)
	}
	if params.SafeSearch != "" {
		query.Set("safeSearch", params.SafeSearch)
	}
	if params.VideoCategoryID != "" {
		query.Set("videoCategoryId", params.VideoCategoryID)
	}
	if params.VideoDefinition != "" {
		query.Set("videoDefinition", params.VideoDefinition)
	}
	if params.VideoDuration != "" {
		query.Set("videoDuration", params.VideoDuration)
	}
	if params.VideoType != "" {
		query.Set("videoType", params.VideoType)
	}
	if params.MaxResults > 0 {
		query.Set("maxResults", fmt.Sprintf("%d", params.MaxResults))
	}
	if params.PageToken != "" {
		query.Set("pageToken", params.PageToken)
	}

	var resp SearchListResponse
	err := client.Get(ctx, "search", query, "search.list", &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// SearchVideos searches for videos.
// WARNING: Each call costs 100 quota units! Use sparingly.
// Quota cost: 100 units per call.
func SearchVideos(ctx context.Context, client *core.Client, query string, maxResults int) (*SearchListResponse, error) {
	return Search(ctx, client, &SearchParams{
		Query:      query,
		Type:       SearchTypeVideo,
		MaxResults: maxResults,
	})
}

// SearchLiveStreams searches for currently live streams.
// WARNING: Each call costs 100 quota units! Use sparingly.
// Quota cost: 100 units per call.
func SearchLiveStreams(ctx context.Context, client *core.Client, query string, maxResults int) (*SearchListResponse, error) {
	return Search(ctx, client, &SearchParams{
		Query:      query,
		Type:       SearchTypeVideo,
		EventType:  SearchEventTypeLive,
		MaxResults: maxResults,
	})
}

// SearchChannels searches for channels.
// WARNING: Each call costs 100 quota units! Use sparingly.
// Quota cost: 100 units per call.
func SearchChannels(ctx context.Context, client *core.Client, query string, maxResults int) (*SearchListResponse, error) {
	return Search(ctx, client, &SearchParams{
		Query:      query,
		Type:       SearchTypeChannel,
		MaxResults: maxResults,
	})
}

// IsVideo returns true if this result is a video.
func (s *SearchResult) IsVideo() bool {
	return s.ID != nil && s.ID.VideoID != ""
}

// IsChannel returns true if this result is a channel.
func (s *SearchResult) IsChannel() bool {
	return s.ID != nil && s.ID.ChannelID != "" && s.ID.VideoID == ""
}

// IsPlaylist returns true if this result is a playlist.
func (s *SearchResult) IsPlaylist() bool {
	return s.ID != nil && s.ID.PlaylistID != ""
}

// IsLive returns true if this result is a currently live video.
func (s *SearchResult) IsLive() bool {
	return s.Snippet != nil && s.Snippet.LiveBroadcastContent == "live"
}

// IsUpcoming returns true if this result is an upcoming broadcast.
func (s *SearchResult) IsUpcoming() bool {
	return s.Snippet != nil && s.Snippet.LiveBroadcastContent == "upcoming"
}

// ResourceID returns the appropriate ID for this result (video, channel, or playlist).
func (s *SearchResult) ResourceID() string {
	if s.ID == nil {
		return ""
	}
	if s.ID.VideoID != "" {
		return s.ID.VideoID
	}
	if s.ID.ChannelID != "" {
		return s.ID.ChannelID
	}
	return s.ID.PlaylistID
}
