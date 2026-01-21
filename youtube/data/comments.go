package data

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Its-donkey/yougopher/youtube/core"
)

// CommentThread represents a comment thread resource.
type CommentThread struct {
	// Kind is the resource type (youtube#commentThread).
	Kind string `json:"kind,omitempty"`

	// ETag is the entity tag.
	ETag string `json:"etag,omitempty"`

	// ID is the comment thread's unique identifier.
	ID string `json:"id,omitempty"`

	// Snippet contains basic details about the comment thread.
	Snippet *CommentThreadSnippet `json:"snippet,omitempty"`

	// Replies contains the comment's replies.
	Replies *CommentThreadReplies `json:"replies,omitempty"`
}

// CommentThreadSnippet contains basic details about a comment thread.
type CommentThreadSnippet struct {
	// ChannelID is the ID of the channel associated with the comments.
	ChannelID string `json:"channelId,omitempty"`

	// VideoID is the ID of the video the comments are for.
	VideoID string `json:"videoId,omitempty"`

	// TopLevelComment is the top-level (parent) comment.
	TopLevelComment *Comment `json:"topLevelComment,omitempty"`

	// CanReply indicates if the viewer can reply to the thread.
	CanReply bool `json:"canReply,omitempty"`

	// TotalReplyCount is the total number of replies.
	TotalReplyCount int `json:"totalReplyCount,omitempty"`

	// IsPublic indicates if the thread is publicly visible.
	IsPublic bool `json:"isPublic,omitempty"`
}

// CommentThreadReplies contains replies to a comment thread.
type CommentThreadReplies struct {
	// Comments contains the reply comments.
	Comments []*Comment `json:"comments,omitempty"`
}

// Comment represents a single comment.
type Comment struct {
	// Kind is the resource type (youtube#comment).
	Kind string `json:"kind,omitempty"`

	// ETag is the entity tag.
	ETag string `json:"etag,omitempty"`

	// ID is the comment's unique identifier.
	ID string `json:"id,omitempty"`

	// Snippet contains the comment's details.
	Snippet *CommentSnippet `json:"snippet,omitempty"`
}

// CommentSnippet contains the comment's details.
type CommentSnippet struct {
	// ChannelID is the ID of the channel associated with the comment.
	ChannelID string `json:"channelId,omitempty"`

	// VideoID is the ID of the video the comment is for.
	VideoID string `json:"videoId,omitempty"`

	// TextDisplay is the comment's text (rendered with links).
	TextDisplay string `json:"textDisplay,omitempty"`

	// TextOriginal is the comment's original text.
	TextOriginal string `json:"textOriginal,omitempty"`

	// ParentID is the ID of the parent comment (for replies).
	ParentID string `json:"parentId,omitempty"`

	// AuthorDisplayName is the author's display name.
	AuthorDisplayName string `json:"authorDisplayName,omitempty"`

	// AuthorProfileImageURL is the URL of the author's profile image.
	AuthorProfileImageURL string `json:"authorProfileImageUrl,omitempty"`

	// AuthorChannelURL is the URL of the author's channel.
	AuthorChannelURL string `json:"authorChannelUrl,omitempty"`

	// AuthorChannelID contains the author's channel ID.
	AuthorChannelID *AuthorChannelID `json:"authorChannelId,omitempty"`

	// CanRate indicates if the viewer can rate the comment.
	CanRate bool `json:"canRate,omitempty"`

	// ViewerRating is the viewer's rating of the comment.
	ViewerRating string `json:"viewerRating,omitempty"`

	// LikeCount is the number of likes on the comment.
	LikeCount int `json:"likeCount,omitempty"`

	// ModerationStatus is the moderation status of the comment.
	ModerationStatus string `json:"moderationStatus,omitempty"`

	// PublishedAt is when the comment was published.
	PublishedAt time.Time `json:"publishedAt,omitempty"`

	// UpdatedAt is when the comment was last updated.
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

// AuthorChannelID contains the channel ID of a comment author.
type AuthorChannelID struct {
	Value string `json:"value,omitempty"`
}

// CommentThreadListResponse is the response from commentThreads.list.
type CommentThreadListResponse struct {
	// Kind is the resource type.
	Kind string `json:"kind,omitempty"`

	// ETag is the entity tag.
	ETag string `json:"etag,omitempty"`

	// NextPageToken is the token for the next page.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// PageInfo contains paging information.
	PageInfo *PageInfo `json:"pageInfo,omitempty"`

	// Items contains the comment thread resources.
	Items []*CommentThread `json:"items,omitempty"`
}

// Comment moderation status constants.
const (
	ModerationStatusHeldForReview = "heldForReview"
	ModerationStatusLikelySpam    = "likelySpam"
	ModerationStatusPublished     = "published"
	ModerationStatusRejected      = "rejected"
)

// Comment thread order constants.
const (
	CommentOrderRelevance = "relevance"
	CommentOrderTime      = "time"
)

// GetCommentThreadsParams contains parameters for commentThreads.list.
type GetCommentThreadsParams struct {
	// IDs is a list of comment thread IDs to retrieve.
	IDs []string

	// VideoID retrieves comments for the specified video.
	VideoID string

	// ChannelID retrieves comments for the specified channel.
	ChannelID string

	// AllThreadsRelatedToChannelID retrieves all threads related to a channel.
	AllThreadsRelatedToChannelID string

	// ModerationStatus filters by moderation status.
	// Values: "heldForReview", "likelySpam", "published"
	ModerationStatus string

	// Order specifies the sort order.
	// Values: "relevance", "time"
	Order string

	// SearchTerms filters comments by search terms.
	SearchTerms string

	// Parts specifies which parts to include in the response.
	// Common values: "snippet", "replies"
	Parts []string

	// MaxResults is the maximum number of items to return (1-100).
	MaxResults int

	// PageToken is the token for pagination.
	PageToken string
}

// DefaultCommentThreadParts are the default parts to request for comment threads.
var DefaultCommentThreadParts = []string{"snippet", "replies"}

// GetCommentThreads retrieves comment threads.
// Quota cost: 1 unit per call.
func GetCommentThreads(ctx context.Context, client *core.Client, params *GetCommentThreadsParams) (*CommentThreadListResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}

	// Validate that at least one filter is provided
	if len(params.IDs) == 0 && params.VideoID == "" && params.ChannelID == "" && params.AllThreadsRelatedToChannelID == "" {
		return nil, fmt.Errorf("at least one of IDs, VideoID, ChannelID, or AllThreadsRelatedToChannelID is required")
	}

	parts := params.Parts
	if len(parts) == 0 {
		parts = DefaultCommentThreadParts
	}

	query := url.Values{}
	query.Set("part", strings.Join(parts, ","))

	if len(params.IDs) > 0 {
		query.Set("id", strings.Join(params.IDs, ","))
	}
	if params.VideoID != "" {
		query.Set("videoId", params.VideoID)
	}
	if params.ChannelID != "" {
		query.Set("channelId", params.ChannelID)
	}
	if params.AllThreadsRelatedToChannelID != "" {
		query.Set("allThreadsRelatedToChannelId", params.AllThreadsRelatedToChannelID)
	}
	if params.ModerationStatus != "" {
		query.Set("moderationStatus", params.ModerationStatus)
	}
	if params.Order != "" {
		query.Set("order", params.Order)
	}
	if params.SearchTerms != "" {
		query.Set("searchTerms", params.SearchTerms)
	}
	if params.MaxResults > 0 {
		query.Set("maxResults", fmt.Sprintf("%d", params.MaxResults))
	}
	if params.PageToken != "" {
		query.Set("pageToken", params.PageToken)
	}

	var resp CommentThreadListResponse
	err := client.Get(ctx, "commentThreads", query, "commentThreads.list", &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetVideoComments retrieves comment threads for a video.
// Quota cost: 1 unit per call.
func GetVideoComments(ctx context.Context, client *core.Client, videoID string, maxResults int) (*CommentThreadListResponse, error) {
	if videoID == "" {
		return nil, fmt.Errorf("video ID cannot be empty")
	}

	return GetCommentThreads(ctx, client, &GetCommentThreadsParams{
		VideoID:    videoID,
		MaxResults: maxResults,
		Order:      CommentOrderRelevance,
	})
}

// CommentListResponse is the response from comments.list.
type CommentListResponse struct {
	// Kind is the resource type.
	Kind string `json:"kind,omitempty"`

	// ETag is the entity tag.
	ETag string `json:"etag,omitempty"`

	// NextPageToken is the token for the next page.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// PageInfo contains paging information.
	PageInfo *PageInfo `json:"pageInfo,omitempty"`

	// Items contains the comment resources.
	Items []*Comment `json:"items,omitempty"`
}

// GetCommentsParams contains parameters for comments.list.
type GetCommentsParams struct {
	// IDs is a list of comment IDs to retrieve.
	IDs []string

	// ParentID retrieves replies to the specified comment.
	ParentID string

	// Parts specifies which parts to include in the response.
	// Common values: "snippet"
	Parts []string

	// MaxResults is the maximum number of items to return (1-100).
	MaxResults int

	// PageToken is the token for pagination.
	PageToken string
}

// DefaultCommentParts are the default parts to request for comments.
var DefaultCommentParts = []string{"snippet"}

// GetComments retrieves comments.
// Quota cost: 1 unit per call.
func GetComments(ctx context.Context, client *core.Client, params *GetCommentsParams) (*CommentListResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}

	// Validate that at least one filter is provided
	if len(params.IDs) == 0 && params.ParentID == "" {
		return nil, fmt.Errorf("at least one of IDs or ParentID is required")
	}

	parts := params.Parts
	if len(parts) == 0 {
		parts = DefaultCommentParts
	}

	query := url.Values{}
	query.Set("part", strings.Join(parts, ","))

	if len(params.IDs) > 0 {
		query.Set("id", strings.Join(params.IDs, ","))
	}
	if params.ParentID != "" {
		query.Set("parentId", params.ParentID)
	}
	if params.MaxResults > 0 {
		query.Set("maxResults", fmt.Sprintf("%d", params.MaxResults))
	}
	if params.PageToken != "" {
		query.Set("pageToken", params.PageToken)
	}

	var resp CommentListResponse
	err := client.Get(ctx, "comments", query, "comments.list", &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetCommentReplies retrieves replies to a comment.
// Quota cost: 1 unit per call.
func GetCommentReplies(ctx context.Context, client *core.Client, parentID string, maxResults int) (*CommentListResponse, error) {
	if parentID == "" {
		return nil, fmt.Errorf("parent ID cannot be empty")
	}

	return GetComments(ctx, client, &GetCommentsParams{
		ParentID:   parentID,
		MaxResults: maxResults,
	})
}

// TopLevelComment returns the top-level comment from a thread.
// Returns nil if not available.
func (ct *CommentThread) TopLevelComment() *Comment {
	if ct.Snippet == nil {
		return nil
	}
	return ct.Snippet.TopLevelComment
}

// ReplyCount returns the number of replies on this thread.
func (ct *CommentThread) ReplyCount() int {
	if ct.Snippet == nil {
		return 0
	}
	return ct.Snippet.TotalReplyCount
}

// AuthorID returns the channel ID of the comment author.
// Returns empty string if not available.
func (c *Comment) AuthorID() string {
	if c.Snippet == nil || c.Snippet.AuthorChannelID == nil {
		return ""
	}
	return c.Snippet.AuthorChannelID.Value
}

// IsReply returns true if this comment is a reply to another comment.
func (c *Comment) IsReply() bool {
	return c.Snippet != nil && c.Snippet.ParentID != ""
}
