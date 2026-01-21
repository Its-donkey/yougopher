package streaming

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Its-donkey/yougopher/youtube/core"
)

// LiveStream represents a YouTube live stream resource.
// A stream is the video feed that is sent to YouTube for encoding and distribution.
type LiveStream struct {
	// Kind is the resource type (youtube#liveStream).
	Kind string `json:"kind,omitempty"`

	// ETag is the entity tag.
	ETag string `json:"etag,omitempty"`

	// ID is the stream's unique identifier.
	ID string `json:"id,omitempty"`

	// Snippet contains basic details about the stream.
	Snippet *StreamSnippet `json:"snippet,omitempty"`

	// CDN contains CDN settings for the stream.
	CDN *StreamCDN `json:"cdn,omitempty"`

	// Status contains the stream's status information.
	Status *StreamStatus `json:"status,omitempty"`

	// ContentDetails contains stream-specific settings.
	ContentDetails *StreamContentDetails `json:"contentDetails,omitempty"`
}

// StreamSnippet contains basic details about a stream.
type StreamSnippet struct {
	// PublishedAt is when the stream was created.
	PublishedAt time.Time `json:"publishedAt,omitempty"`

	// ChannelID is the ID of the channel that owns the stream.
	ChannelID string `json:"channelId,omitempty"`

	// Title is the stream's title.
	Title string `json:"title,omitempty"`

	// Description is the stream's description.
	Description string `json:"description,omitempty"`

	// IsDefaultStream indicates if this is the channel's default stream.
	IsDefaultStream bool `json:"isDefaultStream,omitempty"`
}

// StreamCDN contains CDN settings for ingest and distribution.
type StreamCDN struct {
	// IngestionType is the method for sending the video stream.
	// Values: "rtmp", "dash", "hls"
	IngestionType string `json:"ingestionType,omitempty"`

	// IngestionInfo contains the stream key and ingest URLs.
	IngestionInfo *IngestionInfo `json:"ingestionInfo,omitempty"`

	// Resolution is the resolution of the inbound video data.
	// Values: "240p", "360p", "480p", "720p", "1080p", "1440p", "2160p", "variable"
	Resolution string `json:"resolution,omitempty"`

	// FrameRate is the frame rate of the inbound video data.
	// Values: "30fps", "60fps", "variable"
	FrameRate string `json:"frameRate,omitempty"`
}

// IngestionInfo contains the stream key and ingest URLs.
type IngestionInfo struct {
	// StreamName is the stream key (used in OBS/streaming software).
	StreamName string `json:"streamName,omitempty"`

	// IngestionAddress is the primary RTMP ingest URL.
	IngestionAddress string `json:"ingestionAddress,omitempty"`

	// BackupIngestionAddress is the backup RTMP ingest URL.
	BackupIngestionAddress string `json:"backupIngestionAddress,omitempty"`

	// RtmpsIngestionAddress is the primary RTMPS (secure) ingest URL.
	RtmpsIngestionAddress string `json:"rtmpsIngestionAddress,omitempty"`

	// RtmpsBackupIngestionAddress is the backup RTMPS (secure) ingest URL.
	RtmpsBackupIngestionAddress string `json:"rtmpsBackupIngestionAddress,omitempty"`
}

// StreamStatus contains the stream's status information.
type StreamStatus struct {
	// StreamStatus is the stream's current status.
	// Values: "active", "created", "error", "inactive", "ready"
	StreamStatus string `json:"streamStatus,omitempty"`

	// HealthStatus contains health information about the stream.
	HealthStatus *StreamHealthStatus `json:"healthStatus,omitempty"`
}

// StreamHealthStatus contains health information about the stream.
type StreamHealthStatus struct {
	// Status is the overall health status.
	// Values: "good", "ok", "bad", "noData"
	Status string `json:"status,omitempty"`

	// LastUpdateTimeSeconds is when the health status was last updated.
	LastUpdateTimeSeconds int64 `json:"lastUpdateTimeSeconds,omitempty,string"`

	// ConfigurationIssues contains any configuration issues.
	ConfigurationIssues []*ConfigurationIssue `json:"configurationIssues,omitempty"`
}

// ConfigurationIssue represents a stream configuration problem.
type ConfigurationIssue struct {
	// Type is the type of issue.
	Type string `json:"type,omitempty"`

	// Severity is the severity of the issue.
	// Values: "info", "warning", "error"
	Severity string `json:"severity,omitempty"`

	// Reason is the reason for the issue.
	Reason string `json:"reason,omitempty"`

	// Description is a description of the issue.
	Description string `json:"description,omitempty"`
}

// StreamContentDetails contains stream-specific settings.
type StreamContentDetails struct {
	// ClosedCaptionsIngestionURL is the URL to send closed captions.
	ClosedCaptionsIngestionURL string `json:"closedCaptionsIngestionUrl,omitempty"`

	// IsReusable indicates if the stream can be reused for multiple broadcasts.
	IsReusable bool `json:"isReusable,omitempty"`
}

// Stream status constants.
const (
	StreamStatusActive   = "active"
	StreamStatusCreated  = "created"
	StreamStatusError    = "error"
	StreamStatusInactive = "inactive"
	StreamStatusReady    = "ready"
)

// Stream health status constants.
const (
	StreamHealthGood   = "good"
	StreamHealthOK     = "ok"
	StreamHealthBad    = "bad"
	StreamHealthNoData = "noData"
)

// LiveStreamListResponse is the response from liveStreams.list.
type LiveStreamListResponse struct {
	// Kind is the resource type.
	Kind string `json:"kind,omitempty"`

	// ETag is the entity tag.
	ETag string `json:"etag,omitempty"`

	// NextPageToken is the token for the next page.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// PrevPageToken is the token for the previous page.
	PrevPageToken string `json:"prevPageToken,omitempty"`

	// PageInfo contains paging information.
	PageInfo *StreamPageInfo `json:"pageInfo,omitempty"`

	// Items contains the stream resources.
	Items []*LiveStream `json:"items,omitempty"`
}

// StreamPageInfo contains paging information.
type StreamPageInfo struct {
	// TotalResults is the total number of results.
	TotalResults int `json:"totalResults,omitempty"`

	// ResultsPerPage is the number of results per page.
	ResultsPerPage int `json:"resultsPerPage,omitempty"`
}

// GetStreamsParams contains parameters for liveStreams.list.
type GetStreamsParams struct {
	// IDs is a list of stream IDs to retrieve.
	IDs []string

	// Mine retrieves the authenticated user's streams.
	Mine bool

	// Parts specifies which parts to include in the response.
	// Common values: "snippet", "cdn", "status", "contentDetails"
	Parts []string

	// MaxResults is the maximum number of items to return (1-50).
	MaxResults int

	// PageToken is the token for pagination.
	PageToken string
}

// DefaultStreamParts are the default parts to request for streams.
var DefaultStreamParts = []string{"snippet", "cdn", "status"}

// GetStreams retrieves live stream information.
// Requires OAuth authentication.
// Quota cost: 5 units per call.
func GetStreams(ctx context.Context, client *core.Client, params *GetStreamsParams) (*LiveStreamListResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}

	// Validate that at least one filter is provided
	if len(params.IDs) == 0 && !params.Mine {
		return nil, fmt.Errorf("at least one of IDs or Mine is required")
	}

	parts := params.Parts
	if len(parts) == 0 {
		parts = DefaultStreamParts
	}

	query := url.Values{}
	query.Set("part", strings.Join(parts, ","))

	if len(params.IDs) > 0 {
		query.Set("id", strings.Join(params.IDs, ","))
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

	var resp LiveStreamListResponse
	err := client.Get(ctx, "liveStreams", query, "liveStreams.list", &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetStream retrieves a single stream by ID.
// This is a convenience wrapper around GetStreams.
// Quota cost: 5 units.
func GetStream(ctx context.Context, client *core.Client, streamID string, parts ...string) (*LiveStream, error) {
	if streamID == "" {
		return nil, fmt.Errorf("stream ID cannot be empty")
	}

	if len(parts) == 0 {
		parts = DefaultStreamParts
	}

	resp, err := GetStreams(ctx, client, &GetStreamsParams{
		IDs:   []string{streamID},
		Parts: parts,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Items) == 0 {
		return nil, &core.NotFoundError{
			ResourceType: "stream",
			ResourceID:   streamID,
		}
	}

	return resp.Items[0], nil
}

// GetMyStreams retrieves the authenticated user's streams.
// Requires OAuth authentication.
// Quota cost: 5 units.
func GetMyStreams(ctx context.Context, client *core.Client, parts ...string) (*LiveStreamListResponse, error) {
	if len(parts) == 0 {
		parts = DefaultStreamParts
	}

	return GetStreams(ctx, client, &GetStreamsParams{
		Mine:  true,
		Parts: parts,
	})
}

// InsertStream creates a new live stream.
// Requires OAuth authentication with youtube.force-ssl scope.
// Quota cost: 50 units.
func InsertStream(ctx context.Context, client *core.Client, stream *LiveStream, parts ...string) (*LiveStream, error) {
	if stream == nil {
		return nil, fmt.Errorf("stream cannot be nil")
	}
	if stream.Snippet == nil || stream.Snippet.Title == "" {
		return nil, fmt.Errorf("stream title is required")
	}
	if stream.CDN == nil {
		return nil, fmt.Errorf("stream CDN configuration is required")
	}

	if len(parts) == 0 {
		parts = DefaultStreamParts
	}

	query := url.Values{}
	query.Set("part", strings.Join(parts, ","))

	var resp LiveStream
	err := client.Post(ctx, "liveStreams", query, stream, "liveStreams.insert", &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// UpdateStream updates an existing live stream.
// Requires OAuth authentication with youtube.force-ssl scope.
// The stream must include the ID field.
// Quota cost: 50 units.
func UpdateStream(ctx context.Context, client *core.Client, stream *LiveStream, parts ...string) (*LiveStream, error) {
	if stream == nil {
		return nil, fmt.Errorf("stream cannot be nil")
	}
	if stream.ID == "" {
		return nil, fmt.Errorf("stream ID is required for update")
	}

	if len(parts) == 0 {
		parts = DefaultStreamParts
	}

	query := url.Values{}
	query.Set("part", strings.Join(parts, ","))

	var resp LiveStream
	err := client.Put(ctx, "liveStreams", query, stream, "liveStreams.update", &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// DeleteStream deletes a live stream.
// Requires OAuth authentication with youtube.force-ssl scope.
// Quota cost: 50 units.
func DeleteStream(ctx context.Context, client *core.Client, streamID string) error {
	if streamID == "" {
		return fmt.Errorf("stream ID cannot be empty")
	}

	query := url.Values{}
	query.Set("id", streamID)

	return client.Delete(ctx, "liveStreams", query, "liveStreams.delete")
}

// IsActive returns true if the stream is actively receiving data.
func (s *LiveStream) IsActive() bool {
	if s.Status == nil {
		return false
	}
	return s.Status.StreamStatus == StreamStatusActive
}

// IsReady returns true if the stream is ready to go live.
func (s *LiveStream) IsReady() bool {
	if s.Status == nil {
		return false
	}
	return s.Status.StreamStatus == StreamStatusReady
}

// IsHealthy returns true if the stream health is good or ok.
func (s *LiveStream) IsHealthy() bool {
	if s.Status == nil || s.Status.HealthStatus == nil {
		return false
	}
	status := s.Status.HealthStatus.Status
	return status == StreamHealthGood || status == StreamHealthOK
}

// StreamKey returns the stream key for OBS/streaming software.
// Returns empty string if not available.
func (s *LiveStream) StreamKey() string {
	if s.CDN == nil || s.CDN.IngestionInfo == nil {
		return ""
	}
	return s.CDN.IngestionInfo.StreamName
}

// RTMPUrl returns the primary RTMP ingest URL.
// Returns empty string if not available.
func (s *LiveStream) RTMPUrl() string {
	if s.CDN == nil || s.CDN.IngestionInfo == nil {
		return ""
	}
	return s.CDN.IngestionInfo.IngestionAddress
}

// RTMPSUrl returns the primary RTMPS (secure) ingest URL.
// Returns empty string if not available.
func (s *LiveStream) RTMPSUrl() string {
	if s.CDN == nil || s.CDN.IngestionInfo == nil {
		return ""
	}
	return s.CDN.IngestionInfo.RtmpsIngestionAddress
}

// HasConfigurationIssues returns true if there are any configuration issues.
func (s *LiveStream) HasConfigurationIssues() bool {
	if s.Status == nil || s.Status.HealthStatus == nil {
		return false
	}
	return len(s.Status.HealthStatus.ConfigurationIssues) > 0
}
