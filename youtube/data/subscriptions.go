package data

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Its-donkey/yougopher/youtube/core"
)

// Subscription represents a subscription resource.
type Subscription struct {
	// Kind is the resource type (youtube#subscription).
	Kind string `json:"kind,omitempty"`

	// ETag is the entity tag.
	ETag string `json:"etag,omitempty"`

	// ID is the subscription's unique identifier.
	ID string `json:"id,omitempty"`

	// Snippet contains basic details about the subscription.
	Snippet *SubscriptionSnippet `json:"snippet,omitempty"`

	// ContentDetails contains information about the subscription content.
	ContentDetails *SubscriptionContentDetails `json:"contentDetails,omitempty"`

	// SubscriberSnippet contains details about the subscriber.
	SubscriberSnippet *SubscriberSnippet `json:"subscriberSnippet,omitempty"`
}

// SubscriptionSnippet contains basic details about a subscription.
type SubscriptionSnippet struct {
	// PublishedAt is when the subscription was created.
	PublishedAt time.Time `json:"publishedAt,omitempty"`

	// ChannelID is the ID of the subscriber's channel.
	ChannelID string `json:"channelId,omitempty"`

	// Title is the title of the subscribed channel.
	Title string `json:"title,omitempty"`

	// Description is the description of the subscribed channel.
	Description string `json:"description,omitempty"`

	// ResourceID identifies the subscribed resource.
	ResourceID *SubscriptionResourceID `json:"resourceId,omitempty"`

	// Thumbnails contains thumbnail images.
	Thumbnails *ThumbnailDetails `json:"thumbnails,omitempty"`
}

// SubscriptionResourceID identifies the subscribed resource.
type SubscriptionResourceID struct {
	// Kind is the resource type (e.g., "youtube#channel").
	Kind string `json:"kind,omitempty"`

	// ChannelID is the ID of the subscribed channel.
	ChannelID string `json:"channelId,omitempty"`
}

// SubscriptionContentDetails contains subscription content information.
type SubscriptionContentDetails struct {
	// TotalItemCount is the number of items in the subscription feed.
	TotalItemCount int `json:"totalItemCount,omitempty"`

	// NewItemCount is the number of new items since last read.
	NewItemCount int `json:"newItemCount,omitempty"`

	// ActivityType specifies which activities are included.
	ActivityType string `json:"activityType,omitempty"`
}

// SubscriberSnippet contains details about the subscriber.
type SubscriberSnippet struct {
	// Title is the subscriber's channel title.
	Title string `json:"title,omitempty"`

	// Description is the subscriber's channel description.
	Description string `json:"description,omitempty"`

	// ChannelID is the subscriber's channel ID.
	ChannelID string `json:"channelId,omitempty"`

	// Thumbnails contains the subscriber's thumbnail images.
	Thumbnails *ThumbnailDetails `json:"thumbnails,omitempty"`
}

// SubscriptionListResponse is the response from subscriptions.list.
type SubscriptionListResponse struct {
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

	// Items contains the subscription resources.
	Items []*Subscription `json:"items,omitempty"`
}

// Subscription order constants.
const (
	SubscriptionOrderAlphabetical = "alphabetical"
	SubscriptionOrderRelevance    = "relevance"
	SubscriptionOrderUnread       = "unread"
)

// GetSubscriptionsParams contains parameters for subscriptions.list.
type GetSubscriptionsParams struct {
	// IDs is a list of subscription IDs to retrieve.
	IDs []string

	// ChannelID retrieves subscriptions for the specified channel.
	ChannelID string

	// Mine retrieves the authenticated user's subscriptions.
	Mine bool

	// MyRecentSubscribers retrieves recent subscribers (requires OAuth).
	MyRecentSubscribers bool

	// MySubscribers retrieves all subscribers (requires OAuth).
	MySubscribers bool

	// ForChannelID filters subscriptions to specific channels.
	ForChannelID string

	// Order specifies the sort order.
	// Values: "alphabetical", "relevance", "unread"
	Order string

	// Parts specifies which parts to include in the response.
	// Common values: "snippet", "contentDetails", "subscriberSnippet"
	Parts []string

	// MaxResults is the maximum number of items to return (1-50).
	MaxResults int

	// PageToken is the token for pagination.
	PageToken string
}

// DefaultSubscriptionParts are the default parts to request for subscriptions.
var DefaultSubscriptionParts = []string{"snippet", "contentDetails"}

// GetSubscriptions retrieves subscription information.
// Quota cost: 1 unit per call.
func GetSubscriptions(ctx context.Context, client *core.Client, params *GetSubscriptionsParams) (*SubscriptionListResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}

	// Validate that at least one filter is provided
	if len(params.IDs) == 0 && params.ChannelID == "" && !params.Mine && !params.MyRecentSubscribers && !params.MySubscribers {
		return nil, fmt.Errorf("at least one of IDs, ChannelID, Mine, MyRecentSubscribers, or MySubscribers is required")
	}

	parts := params.Parts
	if len(parts) == 0 {
		parts = DefaultSubscriptionParts
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
	if params.MyRecentSubscribers {
		query.Set("myRecentSubscribers", "true")
	}
	if params.MySubscribers {
		query.Set("mySubscribers", "true")
	}
	if params.ForChannelID != "" {
		query.Set("forChannelId", params.ForChannelID)
	}
	if params.Order != "" {
		query.Set("order", params.Order)
	}
	if params.MaxResults > 0 {
		query.Set("maxResults", fmt.Sprintf("%d", params.MaxResults))
	}
	if params.PageToken != "" {
		query.Set("pageToken", params.PageToken)
	}

	var resp SubscriptionListResponse
	err := client.Get(ctx, "subscriptions", query, "subscriptions.list", &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetMySubscriptions retrieves the authenticated user's subscriptions.
// Requires OAuth authentication.
// Quota cost: 1 unit per call.
func GetMySubscriptions(ctx context.Context, client *core.Client, maxResults int) (*SubscriptionListResponse, error) {
	return GetSubscriptions(ctx, client, &GetSubscriptionsParams{
		Mine:       true,
		MaxResults: maxResults,
		Order:      SubscriptionOrderAlphabetical,
	})
}

// GetChannelSubscriptions retrieves a channel's public subscriptions.
// Quota cost: 1 unit per call.
func GetChannelSubscriptions(ctx context.Context, client *core.Client, channelID string, maxResults int) (*SubscriptionListResponse, error) {
	if channelID == "" {
		return nil, fmt.Errorf("channel ID cannot be empty")
	}

	return GetSubscriptions(ctx, client, &GetSubscriptionsParams{
		ChannelID:  channelID,
		MaxResults: maxResults,
	})
}

// IsSubscribedTo checks if the authenticated user is subscribed to a channel.
// Requires OAuth authentication.
// Quota cost: 1 unit.
func IsSubscribedTo(ctx context.Context, client *core.Client, channelID string) (bool, error) {
	if channelID == "" {
		return false, fmt.Errorf("channel ID cannot be empty")
	}

	resp, err := GetSubscriptions(ctx, client, &GetSubscriptionsParams{
		Mine:         true,
		ForChannelID: channelID,
		MaxResults:   1,
	})
	if err != nil {
		return false, err
	}

	return len(resp.Items) > 0, nil
}

// SubscribedChannelID returns the ID of the subscribed channel.
// Returns empty string if not available.
func (s *Subscription) SubscribedChannelID() string {
	if s.Snippet == nil || s.Snippet.ResourceID == nil {
		return ""
	}
	return s.Snippet.ResourceID.ChannelID
}

// HasNewContent returns true if there are new items since last read.
func (s *Subscription) HasNewContent() bool {
	return s.ContentDetails != nil && s.ContentDetails.NewItemCount > 0
}
