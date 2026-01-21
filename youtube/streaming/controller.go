package streaming

import (
	"context"
	"fmt"
	"time"

	"github.com/Its-donkey/yougopher/youtube/core"
)

// StreamController provides high-level broadcast and stream management.
// It orchestrates common streaming workflows like creating broadcasts,
// binding streams, and managing lifecycle transitions.
type StreamController struct {
	client        *core.Client
	tokenProvider TokenProvider
}

// StreamControllerOption configures a StreamController.
type StreamControllerOption func(*StreamController)

// NewStreamController creates a new stream controller.
// The client should be configured with access token or API key.
// The tokenProvider is used for automatic token refresh.
func NewStreamController(client *core.Client, tokenProvider TokenProvider, opts ...StreamControllerOption) (*StreamController, error) {
	if client == nil {
		return nil, fmt.Errorf("client cannot be nil")
	}

	c := &StreamController{
		client:        client,
		tokenProvider: tokenProvider,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// CreateBroadcastParams contains parameters for creating a broadcast with stream.
type CreateBroadcastParams struct {
	// Title is the broadcast title (required).
	Title string

	// Description is the broadcast description.
	Description string

	// ScheduledStartTime is when the broadcast is scheduled to start.
	// If nil, the broadcast can start immediately.
	ScheduledStartTime *time.Time

	// PrivacyStatus is the privacy setting.
	// Values: "private", "public", "unlisted" (default: "unlisted")
	PrivacyStatus string

	// StreamTitle is the title for the stream (default: same as broadcast title).
	StreamTitle string

	// Resolution is the stream resolution.
	// Values: "240p", "360p", "480p", "720p", "1080p", "1440p", "2160p", "variable"
	// Default: "1080p"
	Resolution string

	// FrameRate is the stream frame rate.
	// Values: "30fps", "60fps", "variable"
	// Default: "30fps"
	FrameRate string

	// IngestionType is the ingest method.
	// Values: "rtmp", "dash", "hls"
	// Default: "rtmp"
	IngestionType string

	// EnableDVR enables DVR for the broadcast.
	EnableDVR bool

	// EnableEmbed allows the broadcast to be embedded.
	EnableEmbed bool

	// EnableAutoStart starts broadcast when stream starts.
	EnableAutoStart bool

	// EnableAutoStop ends broadcast when stream stops.
	EnableAutoStop bool

	// RecordFromStart enables recording from the start.
	RecordFromStart bool

	// EnableLowLatency enables low latency mode.
	EnableLowLatency bool

	// LatencyPreference sets the latency preference.
	// Values: "normal", "low", "ultraLow"
	LatencyPreference string

	// MadeForKids indicates if the content is made for children.
	MadeForKids bool
}

// BroadcastWithStream contains a paired broadcast and stream.
type BroadcastWithStream struct {
	// Broadcast is the created broadcast.
	Broadcast *LiveBroadcast

	// Stream is the created and bound stream.
	Stream *LiveStream
}

// CreateBroadcastWithStream creates a broadcast and stream, binding them together.
// This is a convenience method for the common workflow of:
// 1. Creating a broadcast
// 2. Creating a stream
// 3. Binding the stream to the broadcast
//
// Quota cost: 150 units (50 insert broadcast + 50 insert stream + 50 bind)
func (c *StreamController) CreateBroadcastWithStream(ctx context.Context, params *CreateBroadcastParams) (*BroadcastWithStream, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	// Ensure access token is fresh
	if err := c.refreshToken(ctx); err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}

	// Set defaults
	privacyStatus := params.PrivacyStatus
	if privacyStatus == "" {
		privacyStatus = "unlisted"
	}
	resolution := params.Resolution
	if resolution == "" {
		resolution = "1080p"
	}
	frameRate := params.FrameRate
	if frameRate == "" {
		frameRate = "30fps"
	}
	ingestionType := params.IngestionType
	if ingestionType == "" {
		ingestionType = "rtmp"
	}
	streamTitle := params.StreamTitle
	if streamTitle == "" {
		streamTitle = params.Title
	}

	// Create broadcast
	broadcast := &LiveBroadcast{
		Snippet: &BroadcastSnippet{
			Title:              params.Title,
			Description:        params.Description,
			ScheduledStartTime: params.ScheduledStartTime,
		},
		Status: &BroadcastStatus{
			PrivacyStatus:           privacyStatus,
			SelfDeclaredMadeForKids: params.MadeForKids,
		},
		ContentDetails: &BroadcastContentDetails{
			EnableDvr:         params.EnableDVR,
			EnableEmbed:       params.EnableEmbed,
			EnableAutoStart:   params.EnableAutoStart,
			EnableAutoStop:    params.EnableAutoStop,
			RecordFromStart:   params.RecordFromStart,
			EnableLowLatency:  params.EnableLowLatency,
			LatencyPreference: params.LatencyPreference,
		},
	}

	createdBroadcast, err := InsertBroadcast(ctx, c.client, broadcast, "snippet", "status", "contentDetails")
	if err != nil {
		return nil, fmt.Errorf("creating broadcast: %w", err)
	}

	// Create stream
	stream := &LiveStream{
		Snippet: &StreamSnippet{
			Title: streamTitle,
		},
		CDN: &StreamCDN{
			IngestionType: ingestionType,
			Resolution:    resolution,
			FrameRate:     frameRate,
		},
	}

	createdStream, err := InsertStream(ctx, c.client, stream, "snippet", "cdn", "status")
	if err != nil {
		// Try to clean up the broadcast
		_ = DeleteBroadcast(ctx, c.client, createdBroadcast.ID)
		return nil, fmt.Errorf("creating stream: %w", err)
	}

	// Bind stream to broadcast
	boundBroadcast, err := BindBroadcast(ctx, c.client, &BindBroadcastParams{
		BroadcastID: createdBroadcast.ID,
		StreamID:    createdStream.ID,
		Parts:       []string{"snippet", "status", "contentDetails"},
	})
	if err != nil {
		// Try to clean up
		_ = DeleteStream(ctx, c.client, createdStream.ID)
		_ = DeleteBroadcast(ctx, c.client, createdBroadcast.ID)
		return nil, fmt.Errorf("binding stream to broadcast: %w", err)
	}

	return &BroadcastWithStream{
		Broadcast: boundBroadcast,
		Stream:    createdStream,
	}, nil
}

// StartTesting transitions a broadcast to testing state.
// The broadcast must have a bound stream that is receiving data.
// Quota cost: 50 units.
func (c *StreamController) StartTesting(ctx context.Context, broadcastID string) (*LiveBroadcast, error) {
	if broadcastID == "" {
		return nil, fmt.Errorf("broadcast ID cannot be empty")
	}

	if err := c.refreshToken(ctx); err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}

	return TransitionBroadcast(ctx, c.client, broadcastID, TransitionTesting, "snippet", "status", "contentDetails")
}

// GoLive transitions a broadcast from testing to live.
// The broadcast must be in testing state.
// Quota cost: 50 units.
func (c *StreamController) GoLive(ctx context.Context, broadcastID string) (*LiveBroadcast, error) {
	if broadcastID == "" {
		return nil, fmt.Errorf("broadcast ID cannot be empty")
	}

	if err := c.refreshToken(ctx); err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}

	return TransitionBroadcast(ctx, c.client, broadcastID, TransitionLive, "snippet", "status", "contentDetails")
}

// EndBroadcast transitions a broadcast to complete state.
// Quota cost: 50 units.
func (c *StreamController) EndBroadcast(ctx context.Context, broadcastID string) (*LiveBroadcast, error) {
	if broadcastID == "" {
		return nil, fmt.Errorf("broadcast ID cannot be empty")
	}

	if err := c.refreshToken(ctx); err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}

	return TransitionBroadcast(ctx, c.client, broadcastID, TransitionComplete, "snippet", "status", "contentDetails")
}

// GetStreamHealth retrieves the health status of a stream.
// Quota cost: 5 units.
func (c *StreamController) GetStreamHealth(ctx context.Context, streamID string) (*StreamHealthStatus, error) {
	if streamID == "" {
		return nil, fmt.Errorf("stream ID cannot be empty")
	}

	if err := c.refreshToken(ctx); err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}

	stream, err := GetStream(ctx, c.client, streamID, "status")
	if err != nil {
		return nil, err
	}

	if stream.Status == nil || stream.Status.HealthStatus == nil {
		return nil, fmt.Errorf("stream %s has no health status", streamID)
	}

	return stream.Status.HealthStatus, nil
}

// GetBroadcast retrieves a broadcast by ID.
// Quota cost: 5 units.
func (c *StreamController) GetBroadcast(ctx context.Context, broadcastID string) (*LiveBroadcast, error) {
	if broadcastID == "" {
		return nil, fmt.Errorf("broadcast ID cannot be empty")
	}

	if err := c.refreshToken(ctx); err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}

	return GetBroadcast(ctx, c.client, broadcastID, "snippet", "status", "contentDetails")
}

// GetStream retrieves a stream by ID.
// Quota cost: 5 units.
func (c *StreamController) GetStream(ctx context.Context, streamID string) (*LiveStream, error) {
	if streamID == "" {
		return nil, fmt.Errorf("stream ID cannot be empty")
	}

	if err := c.refreshToken(ctx); err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}

	return GetStream(ctx, c.client, streamID, "snippet", "cdn", "status", "contentDetails")
}

// DeleteBroadcast deletes a broadcast.
// Quota cost: 50 units.
func (c *StreamController) DeleteBroadcast(ctx context.Context, broadcastID string) error {
	if broadcastID == "" {
		return fmt.Errorf("broadcast ID cannot be empty")
	}

	if err := c.refreshToken(ctx); err != nil {
		return fmt.Errorf("refreshing token: %w", err)
	}

	return DeleteBroadcast(ctx, c.client, broadcastID)
}

// DeleteStream deletes a stream.
// Quota cost: 50 units.
func (c *StreamController) DeleteStream(ctx context.Context, streamID string) error {
	if streamID == "" {
		return fmt.Errorf("stream ID cannot be empty")
	}

	if err := c.refreshToken(ctx); err != nil {
		return fmt.Errorf("refreshing token: %w", err)
	}

	return DeleteStream(ctx, c.client, streamID)
}

// refreshToken refreshes the access token if a token provider is configured.
func (c *StreamController) refreshToken(ctx context.Context) error {
	if c.tokenProvider == nil {
		return nil
	}

	token, err := c.tokenProvider.AccessToken(ctx)
	if err != nil {
		return err
	}

	c.client.SetAccessToken(token)
	return nil
}
