package data

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Its-donkey/yougopher/youtube/core"
)

// Channel represents a YouTube channel resource.
type Channel struct {
	// Kind is the resource type (youtube#channel).
	Kind string `json:"kind,omitempty"`

	// ETag is the entity tag.
	ETag string `json:"etag,omitempty"`

	// ID is the channel's unique identifier.
	ID string `json:"id,omitempty"`

	// Snippet contains basic details about the channel.
	Snippet *ChannelSnippet `json:"snippet,omitempty"`

	// Statistics contains channel statistics.
	Statistics *ChannelStatistics `json:"statistics,omitempty"`

	// ContentDetails contains information about the channel's content.
	ContentDetails *ChannelContentDetails `json:"contentDetails,omitempty"`

	// BrandingSettings contains branding information.
	BrandingSettings *ChannelBrandingSettings `json:"brandingSettings,omitempty"`
}

// ChannelSnippet contains basic details about a channel.
type ChannelSnippet struct {
	// Title is the channel's title.
	Title string `json:"title,omitempty"`

	// Description is the channel's description.
	Description string `json:"description,omitempty"`

	// CustomURL is the channel's custom URL (e.g., @channelname).
	CustomURL string `json:"customUrl,omitempty"`

	// PublishedAt is when the channel was created.
	PublishedAt time.Time `json:"publishedAt,omitempty"`

	// Thumbnails contains the channel's avatar images.
	Thumbnails *ThumbnailDetails `json:"thumbnails,omitempty"`

	// DefaultLanguage is the default language of the channel's content.
	DefaultLanguage string `json:"defaultLanguage,omitempty"`

	// Country is the country with which the channel is associated.
	Country string `json:"country,omitempty"`
}

// ChannelStatistics contains channel statistics.
type ChannelStatistics struct {
	// ViewCount is the total view count across all videos.
	ViewCount string `json:"viewCount,omitempty"`

	// SubscriberCount is the subscriber count (may be hidden).
	SubscriberCount string `json:"subscriberCount,omitempty"`

	// HiddenSubscriberCount indicates if the subscriber count is hidden.
	HiddenSubscriberCount bool `json:"hiddenSubscriberCount,omitempty"`

	// VideoCount is the number of public videos.
	VideoCount string `json:"videoCount,omitempty"`
}

// ChannelContentDetails contains channel content information.
type ChannelContentDetails struct {
	// RelatedPlaylists contains IDs of related playlists.
	RelatedPlaylists *RelatedPlaylists `json:"relatedPlaylists,omitempty"`
}

// RelatedPlaylists contains IDs of related playlists.
type RelatedPlaylists struct {
	// Likes is the playlist of liked videos.
	Likes string `json:"likes,omitempty"`

	// Uploads is the playlist of uploaded videos.
	Uploads string `json:"uploads,omitempty"`
}

// ChannelBrandingSettings contains branding information.
type ChannelBrandingSettings struct {
	// Channel contains channel-level branding.
	Channel *ChannelSettings `json:"channel,omitempty"`

	// Image contains image URLs.
	Image *ImageSettings `json:"image,omitempty"`
}

// ChannelSettings contains channel-level branding settings.
type ChannelSettings struct {
	// Title is the channel title.
	Title string `json:"title,omitempty"`

	// Description is the channel description.
	Description string `json:"description,omitempty"`

	// Keywords are the channel's keyword tags.
	Keywords string `json:"keywords,omitempty"`

	// UnsubscribedTrailer is the video ID of the trailer for non-subscribers.
	UnsubscribedTrailer string `json:"unsubscribedTrailer,omitempty"`
}

// ImageSettings contains channel image URLs.
type ImageSettings struct {
	// BannerExternalURL is the URL of the banner image.
	BannerExternalURL string `json:"bannerExternalUrl,omitempty"`
}

// ChannelListResponse is the response from channels.list.
type ChannelListResponse struct {
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

	// Items contains the channel resources.
	Items []*Channel `json:"items,omitempty"`
}

// GetChannelsParams contains parameters for channels.list.
type GetChannelsParams struct {
	// IDs is a list of channel IDs to retrieve.
	IDs []string

	// ForUsername retrieves the channel for the specified username.
	ForUsername string

	// Mine retrieves the authenticated user's channel.
	Mine bool

	// Parts specifies which parts to include in the response.
	// Common values: "snippet", "statistics", "contentDetails", "brandingSettings"
	Parts []string

	// MaxResults is the maximum number of items to return (1-50).
	MaxResults int

	// PageToken is the token for pagination.
	PageToken string
}

// DefaultChannelParts are the default parts to request for channels.
var DefaultChannelParts = []string{"snippet", "statistics"}

// GetChannels retrieves channel information.
// Quota cost: 1 unit per call.
func GetChannels(ctx context.Context, client *core.Client, params *GetChannelsParams) (*ChannelListResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}

	// Validate that at least one filter is provided
	if len(params.IDs) == 0 && params.ForUsername == "" && !params.Mine {
		return nil, fmt.Errorf("at least one of IDs, ForUsername, or Mine is required")
	}

	parts := params.Parts
	if len(parts) == 0 {
		parts = DefaultChannelParts
	}

	query := url.Values{}
	query.Set("part", strings.Join(parts, ","))

	if len(params.IDs) > 0 {
		query.Set("id", strings.Join(params.IDs, ","))
	}
	if params.ForUsername != "" {
		query.Set("forUsername", params.ForUsername)
	}
	if params.Mine {
		query.Set("mine", "true")
	}
	if params.MaxResults > 0 {
		query.Set("maxResults", fmt.Sprintf("%d", params.MaxResults))
	}
	if params.PageToken != "" {
		query.Set("pageToken", params.PageToken)
	}

	var resp ChannelListResponse
	err := client.Get(ctx, "channels", query, "channels.list", &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetChannel retrieves a single channel by ID.
// This is a convenience wrapper around GetChannels.
// Quota cost: 1 unit.
func GetChannel(ctx context.Context, client *core.Client, channelID string, parts ...string) (*Channel, error) {
	if channelID == "" {
		return nil, fmt.Errorf("channel ID cannot be empty")
	}

	if len(parts) == 0 {
		parts = DefaultChannelParts
	}

	resp, err := GetChannels(ctx, client, &GetChannelsParams{
		IDs:   []string{channelID},
		Parts: parts,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Items) == 0 {
		return nil, &core.NotFoundError{
			ResourceType: "channel",
			ResourceID:   channelID,
		}
	}

	return resp.Items[0], nil
}

// GetMyChannel retrieves the authenticated user's channel.
// Requires OAuth authentication.
// Quota cost: 1 unit.
func GetMyChannel(ctx context.Context, client *core.Client, parts ...string) (*Channel, error) {
	if len(parts) == 0 {
		parts = DefaultChannelParts
	}

	resp, err := GetChannels(ctx, client, &GetChannelsParams{
		Mine:  true,
		Parts: parts,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no channel found for authenticated user")
	}

	return resp.Items[0], nil
}

// UploadsPlaylistID returns the uploads playlist ID for this channel.
// Returns empty string if not available.
func (c *Channel) UploadsPlaylistID() string {
	if c.ContentDetails == nil || c.ContentDetails.RelatedPlaylists == nil {
		return ""
	}
	return c.ContentDetails.RelatedPlaylists.Uploads
}
