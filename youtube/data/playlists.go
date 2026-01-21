package data

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Its-donkey/yougopher/youtube/core"
)

// Playlist represents a YouTube playlist resource.
type Playlist struct {
	// Kind is the resource type (youtube#playlist).
	Kind string `json:"kind,omitempty"`

	// ETag is the entity tag.
	ETag string `json:"etag,omitempty"`

	// ID is the playlist's unique identifier.
	ID string `json:"id,omitempty"`

	// Snippet contains basic details about the playlist.
	Snippet *PlaylistSnippet `json:"snippet,omitempty"`

	// Status contains the playlist's privacy status.
	Status *PlaylistStatus `json:"status,omitempty"`

	// ContentDetails contains information about the playlist content.
	ContentDetails *PlaylistContentDetails `json:"contentDetails,omitempty"`
}

// PlaylistSnippet contains basic details about a playlist.
type PlaylistSnippet struct {
	// PublishedAt is when the playlist was created.
	PublishedAt time.Time `json:"publishedAt,omitempty"`

	// ChannelID is the ID of the channel that created the playlist.
	ChannelID string `json:"channelId,omitempty"`

	// Title is the playlist's title.
	Title string `json:"title,omitempty"`

	// Description is the playlist's description.
	Description string `json:"description,omitempty"`

	// Thumbnails contains thumbnail images for the playlist.
	Thumbnails *ThumbnailDetails `json:"thumbnails,omitempty"`

	// ChannelTitle is the title of the channel.
	ChannelTitle string `json:"channelTitle,omitempty"`

	// DefaultLanguage is the default language of the playlist.
	DefaultLanguage string `json:"defaultLanguage,omitempty"`
}

// PlaylistStatus contains the playlist's privacy status.
type PlaylistStatus struct {
	// PrivacyStatus is the playlist's privacy status.
	// Values: "private", "public", "unlisted"
	PrivacyStatus string `json:"privacyStatus,omitempty"`
}

// PlaylistContentDetails contains playlist content information.
type PlaylistContentDetails struct {
	// ItemCount is the number of items in the playlist.
	ItemCount int `json:"itemCount,omitempty"`
}

// PlaylistListResponse is the response from playlists.list.
type PlaylistListResponse struct {
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

	// Items contains the playlist resources.
	Items []*Playlist `json:"items,omitempty"`
}

// GetPlaylistsParams contains parameters for playlists.list.
type GetPlaylistsParams struct {
	// IDs is a list of playlist IDs to retrieve.
	IDs []string

	// ChannelID retrieves playlists for the specified channel.
	ChannelID string

	// Mine retrieves the authenticated user's playlists.
	Mine bool

	// Parts specifies which parts to include in the response.
	// Common values: "snippet", "status", "contentDetails"
	Parts []string

	// MaxResults is the maximum number of items to return (1-50).
	MaxResults int

	// PageToken is the token for pagination.
	PageToken string
}

// DefaultPlaylistParts are the default parts to request for playlists.
var DefaultPlaylistParts = []string{"snippet", "contentDetails"}

// GetPlaylists retrieves playlist information.
// Quota cost: 1 unit per call.
func GetPlaylists(ctx context.Context, client *core.Client, params *GetPlaylistsParams) (*PlaylistListResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}

	// Validate that at least one filter is provided
	if len(params.IDs) == 0 && params.ChannelID == "" && !params.Mine {
		return nil, fmt.Errorf("at least one of IDs, ChannelID, or Mine is required")
	}

	parts := params.Parts
	if len(parts) == 0 {
		parts = DefaultPlaylistParts
	}

	query := url.Values{}
	query.Set("part", strings.Join(parts, ","))

	if len(params.IDs) > 0 {
		query.Set("id", strings.Join(params.IDs, ","))
	}
	if params.ChannelID != "" {
		query.Set("channelId", params.ChannelID)
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

	var resp PlaylistListResponse
	err := client.Get(ctx, "playlists", query, "playlists.list", &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetPlaylist retrieves a single playlist by ID.
// Quota cost: 1 unit.
func GetPlaylist(ctx context.Context, client *core.Client, playlistID string, parts ...string) (*Playlist, error) {
	if playlistID == "" {
		return nil, fmt.Errorf("playlist ID cannot be empty")
	}

	if len(parts) == 0 {
		parts = DefaultPlaylistParts
	}

	resp, err := GetPlaylists(ctx, client, &GetPlaylistsParams{
		IDs:   []string{playlistID},
		Parts: parts,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Items) == 0 {
		return nil, &core.NotFoundError{
			ResourceType: "playlist",
			ResourceID:   playlistID,
		}
	}

	return resp.Items[0], nil
}

// GetMyPlaylists retrieves the authenticated user's playlists.
// Requires OAuth authentication.
// Quota cost: 1 unit per call.
func GetMyPlaylists(ctx context.Context, client *core.Client, params *GetPlaylistsParams) (*PlaylistListResponse, error) {
	if params == nil {
		params = &GetPlaylistsParams{}
	}
	params.Mine = true
	return GetPlaylists(ctx, client, params)
}

// PlaylistItem represents an item in a playlist.
type PlaylistItem struct {
	// Kind is the resource type (youtube#playlistItem).
	Kind string `json:"kind,omitempty"`

	// ETag is the entity tag.
	ETag string `json:"etag,omitempty"`

	// ID is the playlist item's unique identifier.
	ID string `json:"id,omitempty"`

	// Snippet contains basic details about the playlist item.
	Snippet *PlaylistItemSnippet `json:"snippet,omitempty"`

	// ContentDetails contains information about the content.
	ContentDetails *PlaylistItemContentDetails `json:"contentDetails,omitempty"`

	// Status contains the playlist item's status.
	Status *PlaylistItemStatus `json:"status,omitempty"`
}

// PlaylistItemSnippet contains basic details about a playlist item.
type PlaylistItemSnippet struct {
	// PublishedAt is when the item was added to the playlist.
	PublishedAt time.Time `json:"publishedAt,omitempty"`

	// ChannelID is the ID of the channel that added the item.
	ChannelID string `json:"channelId,omitempty"`

	// Title is the item's title.
	Title string `json:"title,omitempty"`

	// Description is the item's description.
	Description string `json:"description,omitempty"`

	// Thumbnails contains thumbnail images.
	Thumbnails *ThumbnailDetails `json:"thumbnails,omitempty"`

	// ChannelTitle is the title of the channel.
	ChannelTitle string `json:"channelTitle,omitempty"`

	// PlaylistID is the ID of the playlist.
	PlaylistID string `json:"playlistId,omitempty"`

	// Position is the item's position in the playlist (0-indexed).
	Position int `json:"position,omitempty"`

	// ResourceID identifies the resource (usually a video).
	ResourceID *ResourceID `json:"resourceId,omitempty"`

	// VideoOwnerChannelID is the ID of the channel that uploaded the video.
	VideoOwnerChannelID string `json:"videoOwnerChannelId,omitempty"`

	// VideoOwnerChannelTitle is the title of the channel that uploaded the video.
	VideoOwnerChannelTitle string `json:"videoOwnerChannelTitle,omitempty"`
}

// ResourceID identifies a resource in a playlist item.
type ResourceID struct {
	// Kind is the resource type (e.g., "youtube#video").
	Kind string `json:"kind,omitempty"`

	// VideoID is the video ID if this is a video resource.
	VideoID string `json:"videoId,omitempty"`
}

// PlaylistItemContentDetails contains content information.
type PlaylistItemContentDetails struct {
	// VideoID is the ID of the video.
	VideoID string `json:"videoId,omitempty"`

	// VideoPublishedAt is when the video was published.
	VideoPublishedAt *time.Time `json:"videoPublishedAt,omitempty"`
}

// PlaylistItemStatus contains the playlist item's status.
type PlaylistItemStatus struct {
	// PrivacyStatus is the item's privacy status.
	PrivacyStatus string `json:"privacyStatus,omitempty"`
}

// PlaylistItemListResponse is the response from playlistItems.list.
type PlaylistItemListResponse struct {
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

	// Items contains the playlist item resources.
	Items []*PlaylistItem `json:"items,omitempty"`
}

// GetPlaylistItemsParams contains parameters for playlistItems.list.
type GetPlaylistItemsParams struct {
	// PlaylistID is the playlist to retrieve items from (required).
	PlaylistID string

	// IDs is a list of playlist item IDs to retrieve.
	IDs []string

	// Parts specifies which parts to include in the response.
	// Common values: "snippet", "contentDetails", "status"
	Parts []string

	// MaxResults is the maximum number of items to return (1-50).
	MaxResults int

	// PageToken is the token for pagination.
	PageToken string
}

// DefaultPlaylistItemParts are the default parts to request for playlist items.
var DefaultPlaylistItemParts = []string{"snippet", "contentDetails"}

// GetPlaylistItems retrieves items from a playlist.
// Quota cost: 1 unit per call.
func GetPlaylistItems(ctx context.Context, client *core.Client, params *GetPlaylistItemsParams) (*PlaylistItemListResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}

	// Validate that at least one filter is provided
	if params.PlaylistID == "" && len(params.IDs) == 0 {
		return nil, fmt.Errorf("at least one of PlaylistID or IDs is required")
	}

	parts := params.Parts
	if len(parts) == 0 {
		parts = DefaultPlaylistItemParts
	}

	query := url.Values{}
	query.Set("part", strings.Join(parts, ","))

	if params.PlaylistID != "" {
		query.Set("playlistId", params.PlaylistID)
	}
	if len(params.IDs) > 0 {
		query.Set("id", strings.Join(params.IDs, ","))
	}
	if params.MaxResults > 0 {
		query.Set("maxResults", fmt.Sprintf("%d", params.MaxResults))
	}
	if params.PageToken != "" {
		query.Set("pageToken", params.PageToken)
	}

	var resp PlaylistItemListResponse
	err := client.Get(ctx, "playlistItems", query, "playlistItems.list", &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// VideoID returns the video ID for this playlist item.
// Returns empty string if not available.
func (p *PlaylistItem) VideoID() string {
	if p.ContentDetails != nil && p.ContentDetails.VideoID != "" {
		return p.ContentDetails.VideoID
	}
	if p.Snippet != nil && p.Snippet.ResourceID != nil {
		return p.Snippet.ResourceID.VideoID
	}
	return ""
}
