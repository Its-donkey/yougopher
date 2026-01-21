package data

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Its-donkey/yougopher/youtube/core"
)

// Video represents a YouTube video resource.
type Video struct {
	// Kind is the resource type (youtube#video).
	Kind string `json:"kind,omitempty"`

	// ETag is the entity tag.
	ETag string `json:"etag,omitempty"`

	// ID is the video's unique identifier.
	ID string `json:"id,omitempty"`

	// Snippet contains basic details about the video.
	Snippet *VideoSnippet `json:"snippet,omitempty"`

	// LiveStreamingDetails contains live streaming metadata.
	LiveStreamingDetails *LiveStreamingDetails `json:"liveStreamingDetails,omitempty"`

	// ContentDetails contains information about the video content.
	ContentDetails *VideoContentDetails `json:"contentDetails,omitempty"`

	// Statistics contains video statistics.
	Statistics *VideoStatistics `json:"statistics,omitempty"`

	// Status contains the video's upload status.
	Status *VideoStatus `json:"status,omitempty"`
}

// VideoSnippet contains basic details about a video.
type VideoSnippet struct {
	// PublishedAt is the date and time the video was published.
	PublishedAt time.Time `json:"publishedAt,omitempty"`

	// ChannelID is the ID of the channel that uploaded the video.
	ChannelID string `json:"channelId,omitempty"`

	// Title is the video's title.
	Title string `json:"title,omitempty"`

	// Description is the video's description.
	Description string `json:"description,omitempty"`

	// Thumbnails contains thumbnail images for the video.
	Thumbnails *ThumbnailDetails `json:"thumbnails,omitempty"`

	// ChannelTitle is the title of the channel.
	ChannelTitle string `json:"channelTitle,omitempty"`

	// Tags contains keyword tags associated with the video.
	Tags []string `json:"tags,omitempty"`

	// CategoryID is the video category.
	CategoryID string `json:"categoryId,omitempty"`

	// LiveBroadcastContent indicates if the video is live.
	// Values: "live", "upcoming", "none"
	LiveBroadcastContent string `json:"liveBroadcastContent,omitempty"`
}

// LiveStreamingDetails contains live streaming information.
type LiveStreamingDetails struct {
	// ActualStartTime is when the broadcast actually started.
	ActualStartTime *time.Time `json:"actualStartTime,omitempty"`

	// ActualEndTime is when the broadcast actually ended.
	ActualEndTime *time.Time `json:"actualEndTime,omitempty"`

	// ScheduledStartTime is when the broadcast is scheduled to start.
	ScheduledStartTime *time.Time `json:"scheduledStartTime,omitempty"`

	// ScheduledEndTime is when the broadcast is scheduled to end.
	ScheduledEndTime *time.Time `json:"scheduledEndTime,omitempty"`

	// ConcurrentViewers is the current number of viewers (live only).
	ConcurrentViewers string `json:"concurrentViewers,omitempty"`

	// ActiveLiveChatID is the live chat ID for this broadcast.
	// Use this to connect to the live chat.
	ActiveLiveChatID string `json:"activeLiveChatId,omitempty"`
}

// VideoContentDetails contains information about the video content.
type VideoContentDetails struct {
	// Duration is the video's duration in ISO 8601 format.
	Duration string `json:"duration,omitempty"`

	// Dimension indicates whether the video is 2D or 3D.
	Dimension string `json:"dimension,omitempty"`

	// Definition indicates whether the video is SD or HD.
	Definition string `json:"definition,omitempty"`

	// Caption indicates whether captions are available.
	Caption string `json:"caption,omitempty"`

	// LicensedContent indicates if the content is licensed.
	LicensedContent bool `json:"licensedContent,omitempty"`
}

// VideoStatistics contains video statistics.
type VideoStatistics struct {
	// ViewCount is the number of views.
	ViewCount string `json:"viewCount,omitempty"`

	// LikeCount is the number of likes.
	LikeCount string `json:"likeCount,omitempty"`

	// CommentCount is the number of comments.
	CommentCount string `json:"commentCount,omitempty"`
}

// VideoStatus contains the video's upload status.
type VideoStatus struct {
	// UploadStatus is the upload status.
	UploadStatus string `json:"uploadStatus,omitempty"`

	// PrivacyStatus is the privacy status.
	PrivacyStatus string `json:"privacyStatus,omitempty"`

	// Embeddable indicates if the video can be embedded.
	Embeddable bool `json:"embeddable,omitempty"`

	// PublicStatsViewable indicates if statistics are public.
	PublicStatsViewable bool `json:"publicStatsViewable,omitempty"`
}

// ThumbnailDetails contains thumbnail images at different sizes.
type ThumbnailDetails struct {
	Default  *Thumbnail `json:"default,omitempty"`
	Medium   *Thumbnail `json:"medium,omitempty"`
	High     *Thumbnail `json:"high,omitempty"`
	Standard *Thumbnail `json:"standard,omitempty"`
	Maxres   *Thumbnail `json:"maxres,omitempty"`
}

// Thumbnail represents a single thumbnail image.
type Thumbnail struct {
	URL    string `json:"url,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// VideoListResponse is the response from videos.list.
type VideoListResponse struct {
	// Kind is the resource type.
	Kind string `json:"kind,omitempty"`

	// ETag is the entity tag.
	ETag string `json:"etag,omitempty"`

	// NextPageToken is the token for the next page.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// PrevPageToken is the token for the previous page.
	PrevPageToken string `json:"prevPageToken,omitempty"`

	// PageInfo contains paging information.
	PageInfo *PageInfo `json:"pageInfo,omitempty"`

	// Items contains the video resources.
	Items []*Video `json:"items,omitempty"`
}

// PageInfo contains paging information.
type PageInfo struct {
	// TotalResults is the total number of results.
	TotalResults int `json:"totalResults,omitempty"`

	// ResultsPerPage is the number of results per page.
	ResultsPerPage int `json:"resultsPerPage,omitempty"`
}

// GetVideosParams contains parameters for videos.list.
type GetVideosParams struct {
	// IDs is a list of video IDs to retrieve.
	IDs []string

	// Parts specifies which parts to include in the response.
	// Common values: "snippet", "liveStreamingDetails", "contentDetails", "statistics", "status"
	Parts []string

	// MaxResults is the maximum number of items to return (1-50).
	MaxResults int

	// PageToken is the token for pagination.
	PageToken string
}

// DefaultVideoParts are the default parts to request for videos.
var DefaultVideoParts = []string{"snippet", "liveStreamingDetails"}

// GetVideos retrieves video information.
// Quota cost: 1 unit per call.
func GetVideos(ctx context.Context, client *core.Client, params *GetVideosParams) (*VideoListResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if len(params.IDs) == 0 {
		return nil, fmt.Errorf("at least one video ID is required")
	}

	parts := params.Parts
	if len(parts) == 0 {
		parts = DefaultVideoParts
	}

	query := url.Values{}
	query.Set("part", strings.Join(parts, ","))
	query.Set("id", strings.Join(params.IDs, ","))

	if params.MaxResults > 0 {
		query.Set("maxResults", fmt.Sprintf("%d", params.MaxResults))
	}
	if params.PageToken != "" {
		query.Set("pageToken", params.PageToken)
	}

	var resp VideoListResponse
	err := client.Get(ctx, "videos", query, "videos.list", &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetVideo retrieves a single video by ID.
// This is a convenience wrapper around GetVideos.
// Quota cost: 1 unit.
func GetVideo(ctx context.Context, client *core.Client, videoID string, parts ...string) (*Video, error) {
	if videoID == "" {
		return nil, fmt.Errorf("video ID cannot be empty")
	}

	if len(parts) == 0 {
		parts = DefaultVideoParts
	}

	resp, err := GetVideos(ctx, client, &GetVideosParams{
		IDs:   []string{videoID},
		Parts: parts,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Items) == 0 {
		return nil, &core.NotFoundError{
			ResourceType: "video",
			ResourceID:   videoID,
		}
	}

	return resp.Items[0], nil
}

// GetLiveChatID retrieves the active live chat ID for a video.
// This is a convenience function for chat bot initialization.
// Quota cost: 1 unit.
func GetLiveChatID(ctx context.Context, client *core.Client, videoID string) (string, error) {
	video, err := GetVideo(ctx, client, videoID, "liveStreamingDetails")
	if err != nil {
		return "", err
	}

	if video.LiveStreamingDetails == nil {
		return "", fmt.Errorf("video %s has no live streaming details", videoID)
	}

	if video.LiveStreamingDetails.ActiveLiveChatID == "" {
		return "", fmt.Errorf("video %s has no active live chat", videoID)
	}

	return video.LiveStreamingDetails.ActiveLiveChatID, nil
}

// IsLive returns true if the video is currently live.
func (v *Video) IsLive() bool {
	if v.Snippet == nil {
		return false
	}
	return v.Snippet.LiveBroadcastContent == "live"
}

// IsUpcoming returns true if the video is an upcoming broadcast.
func (v *Video) IsUpcoming() bool {
	if v.Snippet == nil {
		return false
	}
	return v.Snippet.LiveBroadcastContent == "upcoming"
}

// HasActiveLiveChat returns true if the video has an active live chat.
func (v *Video) HasActiveLiveChat() bool {
	if v.LiveStreamingDetails == nil {
		return false
	}
	return v.LiveStreamingDetails.ActiveLiveChatID != ""
}
