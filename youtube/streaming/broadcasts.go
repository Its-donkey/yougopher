package streaming

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Its-donkey/yougopher/youtube/core"
)

// LiveBroadcast represents a YouTube live broadcast resource.
type LiveBroadcast struct {
	// Kind is the resource type (youtube#liveBroadcast).
	Kind string `json:"kind,omitempty"`

	// ETag is the entity tag.
	ETag string `json:"etag,omitempty"`

	// ID is the broadcast's unique identifier.
	ID string `json:"id,omitempty"`

	// Snippet contains basic details about the broadcast.
	Snippet *BroadcastSnippet `json:"snippet,omitempty"`

	// Status contains the broadcast's status information.
	Status *BroadcastStatus `json:"status,omitempty"`

	// ContentDetails contains broadcast-specific settings.
	ContentDetails *BroadcastContentDetails `json:"contentDetails,omitempty"`

	// Statistics contains broadcast statistics.
	Statistics *BroadcastStatistics `json:"statistics,omitempty"`

	// MonetizationDetails contains monetization settings.
	MonetizationDetails *BroadcastMonetizationDetails `json:"monetizationDetails,omitempty"`
}

// BroadcastSnippet contains basic details about a broadcast.
type BroadcastSnippet struct {
	// PublishedAt is when the broadcast was created.
	PublishedAt time.Time `json:"publishedAt,omitempty"`

	// ChannelID is the ID of the channel that created the broadcast.
	ChannelID string `json:"channelId,omitempty"`

	// Title is the broadcast's title.
	Title string `json:"title,omitempty"`

	// Description is the broadcast's description.
	Description string `json:"description,omitempty"`

	// Thumbnails contains thumbnail images for the broadcast.
	Thumbnails *BroadcastThumbnails `json:"thumbnails,omitempty"`

	// ScheduledStartTime is when the broadcast is scheduled to start.
	ScheduledStartTime *time.Time `json:"scheduledStartTime,omitempty"`

	// ScheduledEndTime is when the broadcast is scheduled to end.
	ScheduledEndTime *time.Time `json:"scheduledEndTime,omitempty"`

	// ActualStartTime is when the broadcast actually started.
	ActualStartTime *time.Time `json:"actualStartTime,omitempty"`

	// ActualEndTime is when the broadcast actually ended.
	ActualEndTime *time.Time `json:"actualEndTime,omitempty"`

	// LiveChatID is the ID of the live chat for this broadcast.
	// This is the key field for connecting a chat bot.
	LiveChatID string `json:"liveChatId,omitempty"`

	// IsDefaultBroadcast indicates if this is the default broadcast.
	IsDefaultBroadcast bool `json:"isDefaultBroadcast,omitempty"`
}

// BroadcastThumbnails contains thumbnail images at different sizes.
type BroadcastThumbnails struct {
	Default  *BroadcastThumbnail `json:"default,omitempty"`
	Medium   *BroadcastThumbnail `json:"medium,omitempty"`
	High     *BroadcastThumbnail `json:"high,omitempty"`
	Standard *BroadcastThumbnail `json:"standard,omitempty"`
	Maxres   *BroadcastThumbnail `json:"maxres,omitempty"`
}

// BroadcastThumbnail represents a single thumbnail image.
type BroadcastThumbnail struct {
	URL    string `json:"url,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// BroadcastStatus contains the broadcast's status information.
type BroadcastStatus struct {
	// LifeCycleStatus is the broadcast's lifecycle status.
	// Values: "complete", "created", "live", "liveStarting", "ready", "revoked", "testStarting", "testing"
	LifeCycleStatus string `json:"lifeCycleStatus,omitempty"`

	// PrivacyStatus is the broadcast's privacy status.
	// Values: "private", "public", "unlisted"
	PrivacyStatus string `json:"privacyStatus,omitempty"`

	// RecordingStatus indicates if the broadcast is being recorded.
	// Values: "notRecording", "recorded", "recording"
	RecordingStatus string `json:"recordingStatus,omitempty"`

	// MadeForKids indicates if the broadcast is made for kids.
	MadeForKids bool `json:"madeForKids,omitempty"`

	// SelfDeclaredMadeForKids indicates the creator's self-declaration.
	SelfDeclaredMadeForKids bool `json:"selfDeclaredMadeForKids,omitempty"`
}

// BroadcastContentDetails contains broadcast-specific settings.
type BroadcastContentDetails struct {
	// BoundStreamID is the ID of the bound live stream.
	BoundStreamID string `json:"boundStreamId,omitempty"`

	// BoundStreamLastUpdateTimeMs is when the binding was last updated.
	BoundStreamLastUpdateTimeMs string `json:"boundStreamLastUpdateTimeMs,omitempty"`

	// MonitorStream contains monitor stream settings.
	MonitorStream *MonitorStreamInfo `json:"monitorStream,omitempty"`

	// EnableEmbed indicates if the broadcast can be embedded.
	EnableEmbed bool `json:"enableEmbed,omitempty"`

	// EnableDvr indicates if DVR is enabled.
	EnableDvr bool `json:"enableDvr,omitempty"`

	// EnableContentEncryption indicates if content encryption is enabled.
	EnableContentEncryption bool `json:"enableContentEncryption,omitempty"`

	// StartWithSlate indicates if the broadcast starts with a slate.
	StartWithSlate bool `json:"startWithSlate,omitempty"`

	// RecordFromStart indicates if recording starts automatically.
	RecordFromStart bool `json:"recordFromStart,omitempty"`

	// EnableClosedCaptions indicates if closed captions are enabled.
	EnableClosedCaptions bool `json:"enableClosedCaptions,omitempty"`

	// ClosedCaptionsType is the closed caption type.
	ClosedCaptionsType string `json:"closedCaptionsType,omitempty"`

	// EnableLowLatency indicates if low latency mode is enabled.
	EnableLowLatency bool `json:"enableLowLatency,omitempty"`

	// LatencyPreference is the latency preference.
	LatencyPreference string `json:"latencyPreference,omitempty"`

	// Projection is the broadcast projection (rectangular or 360).
	Projection string `json:"projection,omitempty"`

	// EnableAutoStart indicates if auto-start is enabled.
	EnableAutoStart bool `json:"enableAutoStart,omitempty"`

	// EnableAutoStop indicates if auto-stop is enabled.
	EnableAutoStop bool `json:"enableAutoStop,omitempty"`
}

// MonitorStreamInfo contains monitor stream settings.
type MonitorStreamInfo struct {
	// EnableMonitorStream indicates if the monitor stream is enabled.
	EnableMonitorStream bool `json:"enableMonitorStream,omitempty"`

	// BroadcastStreamDelayMs is the delay between the monitor stream and broadcast.
	BroadcastStreamDelayMs int `json:"broadcastStreamDelayMs,omitempty"`

	// EmbedHTML is the HTML code to embed the monitor stream.
	EmbedHTML string `json:"embedHtml,omitempty"`
}

// BroadcastStatistics contains broadcast statistics.
// Note: The totalChatCount field is deprecated by YouTube but still returned.
type BroadcastStatistics struct {
	// TotalChatCount is the total number of live chat messages.
	// Deprecated: This field is deprecated by the YouTube API.
	TotalChatCount uint64 `json:"totalChatCount,omitempty,string"`
}

// BroadcastMonetizationDetails contains monetization settings for the broadcast.
type BroadcastMonetizationDetails struct {
	// CuepointSchedule contains settings for automatic ad cuepoint scheduling.
	CuepointSchedule *CuepointSchedule `json:"cuepointSchedule,omitempty"`
}

// CuepointSchedule contains settings for automatic ad cuepoint scheduling.
type CuepointSchedule struct {
	// Enabled indicates if automatic cuepoint scheduling is enabled.
	Enabled bool `json:"enabled,omitempty"`

	// PauseAdsUntil is the time until which ads are paused.
	// Format: RFC 3339 datetime.
	PauseAdsUntil string `json:"pauseAdsUntil,omitempty"`

	// ScheduleStrategy is the strategy for scheduling cuepoints.
	// Values: "CONCURRENT", "NON_CONCURRENT"
	ScheduleStrategy string `json:"scheduleStrategy,omitempty"`

	// RepeatIntervalSecs is the interval between scheduled cuepoints in seconds.
	RepeatIntervalSecs int `json:"repeatIntervalSecs,omitempty"`
}

// Broadcast lifecycle status constants.
const (
	BroadcastStatusComplete     = "complete"
	BroadcastStatusCreated      = "created"
	BroadcastStatusLive         = "live"
	BroadcastStatusLiveStarting = "liveStarting"
	BroadcastStatusReady        = "ready"
	BroadcastStatusRevoked      = "revoked"
	BroadcastStatusTestStarting = "testStarting"
	BroadcastStatusTesting      = "testing"
)

// LiveBroadcastListResponse is the response from liveBroadcasts.list.
type LiveBroadcastListResponse struct {
	// Kind is the resource type.
	Kind string `json:"kind,omitempty"`

	// ETag is the entity tag.
	ETag string `json:"etag,omitempty"`

	// NextPageToken is the token for the next page.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// PrevPageToken is the token for the previous page.
	PrevPageToken string `json:"prevPageToken,omitempty"`

	// PageInfo contains paging information.
	PageInfo *BroadcastPageInfo `json:"pageInfo,omitempty"`

	// Items contains the broadcast resources.
	Items []*LiveBroadcast `json:"items,omitempty"`
}

// BroadcastPageInfo contains paging information.
type BroadcastPageInfo struct {
	// TotalResults is the total number of results.
	TotalResults int `json:"totalResults,omitempty"`

	// ResultsPerPage is the number of results per page.
	ResultsPerPage int `json:"resultsPerPage,omitempty"`
}

// GetBroadcastsParams contains parameters for liveBroadcasts.list.
type GetBroadcastsParams struct {
	// IDs is a list of broadcast IDs to retrieve.
	IDs []string

	// Mine retrieves the authenticated user's broadcasts.
	Mine bool

	// BroadcastStatus filters by broadcast status.
	// Values: "active", "all", "completed", "upcoming"
	BroadcastStatus string

	// BroadcastType filters by broadcast type.
	// Values: "all", "event", "persistent"
	BroadcastType string

	// Parts specifies which parts to include in the response.
	// Common values: "snippet", "status", "contentDetails"
	Parts []string

	// MaxResults is the maximum number of items to return (1-50).
	MaxResults int

	// PageToken is the token for pagination.
	PageToken string
}

// DefaultBroadcastParts are the default parts to request for broadcasts.
var DefaultBroadcastParts = []string{"snippet", "status"}

// GetBroadcasts retrieves live broadcast information.
// Requires OAuth authentication.
// Quota cost: 5 units per call.
func GetBroadcasts(ctx context.Context, client *core.Client, params *GetBroadcastsParams) (*LiveBroadcastListResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}

	// Validate that at least one filter is provided
	if len(params.IDs) == 0 && !params.Mine {
		return nil, fmt.Errorf("at least one of IDs or Mine is required")
	}

	parts := params.Parts
	if len(parts) == 0 {
		parts = DefaultBroadcastParts
	}

	query := url.Values{}
	query.Set("part", strings.Join(parts, ","))

	if len(params.IDs) > 0 {
		query.Set("id", strings.Join(params.IDs, ","))
	}
	if params.Mine {
		query.Set("mine", "true")
	}
	if params.BroadcastStatus != "" {
		query.Set("broadcastStatus", params.BroadcastStatus)
	}
	if params.BroadcastType != "" {
		query.Set("broadcastType", params.BroadcastType)
	}
	if params.MaxResults > 0 {
		query.Set("maxResults", fmt.Sprintf("%d", params.MaxResults))
	}
	if params.PageToken != "" {
		query.Set("pageToken", params.PageToken)
	}

	var resp LiveBroadcastListResponse
	err := client.Get(ctx, "liveBroadcasts", query, "liveBroadcasts.list", &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetBroadcast retrieves a single broadcast by ID.
// This is a convenience wrapper around GetBroadcasts.
// Quota cost: 5 units.
func GetBroadcast(ctx context.Context, client *core.Client, broadcastID string, parts ...string) (*LiveBroadcast, error) {
	if broadcastID == "" {
		return nil, fmt.Errorf("broadcast ID cannot be empty")
	}

	if len(parts) == 0 {
		parts = DefaultBroadcastParts
	}

	resp, err := GetBroadcasts(ctx, client, &GetBroadcastsParams{
		IDs:   []string{broadcastID},
		Parts: parts,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Items) == 0 {
		return nil, &core.NotFoundError{
			ResourceType: "broadcast",
			ResourceID:   broadcastID,
		}
	}

	return resp.Items[0], nil
}

// GetMyActiveBroadcast retrieves the authenticated user's currently active broadcast.
// Requires OAuth authentication.
// Quota cost: 5 units.
func GetMyActiveBroadcast(ctx context.Context, client *core.Client, parts ...string) (*LiveBroadcast, error) {
	if len(parts) == 0 {
		parts = DefaultBroadcastParts
	}

	resp, err := GetBroadcasts(ctx, client, &GetBroadcastsParams{
		Mine:            true,
		BroadcastStatus: "active",
		Parts:           parts,
		MaxResults:      1,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no active broadcast found")
	}

	return resp.Items[0], nil
}

// GetBroadcastLiveChatID retrieves the live chat ID for a broadcast.
// Requires OAuth authentication.
// Quota cost: 5 units.
func GetBroadcastLiveChatID(ctx context.Context, client *core.Client, broadcastID string) (string, error) {
	broadcast, err := GetBroadcast(ctx, client, broadcastID, "snippet")
	if err != nil {
		return "", err
	}

	if broadcast.Snippet == nil {
		return "", fmt.Errorf("broadcast %s has no snippet", broadcastID)
	}

	if broadcast.Snippet.LiveChatID == "" {
		return "", fmt.Errorf("broadcast %s has no live chat", broadcastID)
	}

	return broadcast.Snippet.LiveChatID, nil
}

// IsLive returns true if the broadcast is currently live.
func (b *LiveBroadcast) IsLive() bool {
	if b.Status == nil {
		return false
	}
	return b.Status.LifeCycleStatus == BroadcastStatusLive
}

// IsComplete returns true if the broadcast has ended.
func (b *LiveBroadcast) IsComplete() bool {
	if b.Status == nil {
		return false
	}
	return b.Status.LifeCycleStatus == BroadcastStatusComplete
}

// IsUpcoming returns true if the broadcast hasn't started yet.
func (b *LiveBroadcast) IsUpcoming() bool {
	if b.Status == nil {
		return false
	}
	switch b.Status.LifeCycleStatus {
	case BroadcastStatusCreated, BroadcastStatusReady:
		return true
	default:
		return false
	}
}

// LiveChatID returns the live chat ID for this broadcast.
// Returns empty string if not available.
func (b *LiveBroadcast) LiveChatID() string {
	if b.Snippet == nil {
		return ""
	}
	return b.Snippet.LiveChatID
}

// Transition status constants for liveBroadcasts.transition.
const (
	// TransitionTesting transitions to testing state (preview).
	TransitionTesting = "testing"

	// TransitionLive transitions to live state (public broadcast).
	TransitionLive = "live"

	// TransitionComplete transitions to complete state (end broadcast).
	TransitionComplete = "complete"
)

// BindBroadcastParams contains parameters for liveBroadcasts.bind.
type BindBroadcastParams struct {
	// BroadcastID is the ID of the broadcast to bind.
	BroadcastID string

	// StreamID is the ID of the stream to bind (optional - omit to unbind).
	StreamID string

	// Parts specifies which parts to include in the response.
	Parts []string
}

// InsertBroadcast creates a new live broadcast.
// Requires OAuth authentication with youtube.force-ssl scope.
// Quota cost: 50 units.
func InsertBroadcast(ctx context.Context, client *core.Client, broadcast *LiveBroadcast, parts ...string) (*LiveBroadcast, error) {
	if broadcast == nil {
		return nil, fmt.Errorf("broadcast cannot be nil")
	}
	if broadcast.Snippet == nil || broadcast.Snippet.Title == "" {
		return nil, fmt.Errorf("broadcast title is required")
	}

	if len(parts) == 0 {
		parts = DefaultBroadcastParts
	}

	query := url.Values{}
	query.Set("part", strings.Join(parts, ","))

	var resp LiveBroadcast
	err := client.Post(ctx, "liveBroadcasts", query, broadcast, "liveBroadcasts.insert", &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// UpdateBroadcast updates an existing live broadcast.
// Requires OAuth authentication with youtube.force-ssl scope.
// The broadcast must include the ID field.
// Quota cost: 50 units.
func UpdateBroadcast(ctx context.Context, client *core.Client, broadcast *LiveBroadcast, parts ...string) (*LiveBroadcast, error) {
	if broadcast == nil {
		return nil, fmt.Errorf("broadcast cannot be nil")
	}
	if broadcast.ID == "" {
		return nil, fmt.Errorf("broadcast ID is required for update")
	}

	if len(parts) == 0 {
		parts = DefaultBroadcastParts
	}

	query := url.Values{}
	query.Set("part", strings.Join(parts, ","))

	var resp LiveBroadcast
	err := client.Put(ctx, "liveBroadcasts", query, broadcast, "liveBroadcasts.update", &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// DeleteBroadcast deletes a live broadcast.
// Requires OAuth authentication with youtube.force-ssl scope.
// Quota cost: 50 units.
func DeleteBroadcast(ctx context.Context, client *core.Client, broadcastID string) error {
	if broadcastID == "" {
		return fmt.Errorf("broadcast ID cannot be empty")
	}

	query := url.Values{}
	query.Set("id", broadcastID)

	return client.Delete(ctx, "liveBroadcasts", query, "liveBroadcasts.delete")
}

// BindBroadcast binds a stream to a broadcast.
// This associates the video stream with the broadcast, allowing the broadcast
// to receive video from the stream.
// Requires OAuth authentication with youtube.force-ssl scope.
// Quota cost: 50 units.
func BindBroadcast(ctx context.Context, client *core.Client, params *BindBroadcastParams) (*LiveBroadcast, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.BroadcastID == "" {
		return nil, fmt.Errorf("broadcast ID is required")
	}

	parts := params.Parts
	if len(parts) == 0 {
		parts = DefaultBroadcastParts
	}

	query := url.Values{}
	query.Set("id", params.BroadcastID)
	query.Set("part", strings.Join(parts, ","))

	// StreamID is optional - omit to unbind
	if params.StreamID != "" {
		query.Set("streamId", params.StreamID)
	}

	var resp LiveBroadcast
	err := client.Post(ctx, "liveBroadcasts/bind", query, nil, "liveBroadcasts.bind", &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// TransitionBroadcast transitions a broadcast to a new lifecycle status.
// Valid transitions:
//   - testing: Start testing the broadcast (requires bound and active stream)
//   - live: Go live (transition from testing state)
//   - complete: End the broadcast
//
// Requires OAuth authentication with youtube.force-ssl scope.
// Quota cost: 50 units.
func TransitionBroadcast(ctx context.Context, client *core.Client, broadcastID, status string, parts ...string) (*LiveBroadcast, error) {
	if broadcastID == "" {
		return nil, fmt.Errorf("broadcast ID cannot be empty")
	}
	if status == "" {
		return nil, fmt.Errorf("transition status cannot be empty")
	}

	// Validate status
	switch status {
	case TransitionTesting, TransitionLive, TransitionComplete:
		// valid
	default:
		return nil, fmt.Errorf("invalid transition status: %s (must be testing, live, or complete)", status)
	}

	if len(parts) == 0 {
		parts = DefaultBroadcastParts
	}

	query := url.Values{}
	query.Set("id", broadcastID)
	query.Set("broadcastStatus", status)
	query.Set("part", strings.Join(parts, ","))

	var resp LiveBroadcast
	err := client.Post(ctx, "liveBroadcasts/transition", query, nil, "liveBroadcasts.transition", &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// IsTesting returns true if the broadcast is in testing state.
func (b *LiveBroadcast) IsTesting() bool {
	if b.Status == nil {
		return false
	}
	return b.Status.LifeCycleStatus == BroadcastStatusTesting
}

// BoundStreamID returns the ID of the bound stream, if any.
func (b *LiveBroadcast) BoundStreamID() string {
	if b.ContentDetails == nil {
		return ""
	}
	return b.ContentDetails.BoundStreamID
}

// HasBoundStream returns true if a stream is bound to this broadcast.
func (b *LiveBroadcast) HasBoundStream() bool {
	return b.BoundStreamID() != ""
}

// TotalChatCount returns the total number of chat messages in the broadcast.
// Returns 0 if statistics are not available.
// Deprecated: This field is deprecated by the YouTube API.
func (b *LiveBroadcast) TotalChatCount() uint64 {
	if b.Statistics == nil {
		return 0
	}
	return b.Statistics.TotalChatCount
}

// HasCuepointSchedule returns true if automatic ad scheduling is enabled.
func (b *LiveBroadcast) HasCuepointSchedule() bool {
	if b.MonetizationDetails == nil || b.MonetizationDetails.CuepointSchedule == nil {
		return false
	}
	return b.MonetizationDetails.CuepointSchedule.Enabled
}

// CuepointRepeatInterval returns the interval between scheduled ad cuepoints in seconds.
// Returns 0 if not configured.
func (b *LiveBroadcast) CuepointRepeatInterval() int {
	if b.MonetizationDetails == nil || b.MonetizationDetails.CuepointSchedule == nil {
		return 0
	}
	return b.MonetizationDetails.CuepointSchedule.RepeatIntervalSecs
}

// InsertCuepointParams contains parameters for inserting a cuepoint.
type InsertCuepointParams struct {
	// BroadcastID is the ID of the broadcast (required).
	BroadcastID string

	// DurationSecs is the ad break duration in seconds (default 30, max 180).
	// If zero, defaults to 30 seconds.
	DurationSecs int

	// InsertionOffsetTimeMs specifies when to insert the cuepoint relative to
	// the broadcast start. Use -1 or CuepointInsertImmediate for immediate insertion.
	// If zero, InsertionOffsetTimeMs is not sent (defaults to immediate).
	InsertionOffsetTimeMs int64

	// WalltimeMs specifies the wall clock time (Unix milliseconds) when the
	// cuepoint should be inserted. Alternative to InsertionOffsetTimeMs.
	// If both are specified, WalltimeMs takes precedence.
	WalltimeMs int64
}

// CuepointInsertImmediate is the value for immediate cuepoint insertion.
const CuepointInsertImmediate int64 = -1

// InsertCuepoint inserts an ad break cuepoint into a live broadcast.
// This triggers mid-roll ads for viewers who have ad-supported viewing.
//
// Note: Cuepoints can only be inserted into broadcasts that are currently live.
// The broadcast must have monetization enabled for ads to play.
//
// Requires OAuth authentication with youtube.force-ssl scope.
// Quota cost: 50 units.
func InsertCuepoint(ctx context.Context, client *core.Client, params *InsertCuepointParams) (*Cuepoint, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.BroadcastID == "" {
		return nil, fmt.Errorf("broadcast ID is required")
	}

	query := url.Values{}
	query.Set("id", params.BroadcastID)

	req := &CuepointRequest{
		CueType: CueTypeAd,
	}

	// Set duration (default 30 if not specified)
	if params.DurationSecs > 0 {
		req.DurationSecs = params.DurationSecs
	}

	// Set insertion time
	if params.WalltimeMs > 0 {
		req.WalltimeMs = params.WalltimeMs
	} else if params.InsertionOffsetTimeMs != 0 {
		req.InsertionOffsetTimeMs = params.InsertionOffsetTimeMs
	}

	var resp Cuepoint
	err := client.Post(ctx, "liveBroadcasts/cuepoint", query, req, "liveBroadcasts.cuepoint", &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// InsertImmediateCuepoint is a convenience function to insert an immediate ad break.
// Uses default 30 second duration. For custom duration, use InsertCuepoint.
func InsertImmediateCuepoint(ctx context.Context, client *core.Client, broadcastID string) (*Cuepoint, error) {
	return InsertCuepoint(ctx, client, &InsertCuepointParams{
		BroadcastID:           broadcastID,
		InsertionOffsetTimeMs: CuepointInsertImmediate,
	})
}
