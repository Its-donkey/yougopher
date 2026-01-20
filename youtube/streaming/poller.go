package streaming

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Its-donkey/yougopher/youtube/core"
)

// Lifecycle states for the poller.
const (
	stateStopped  int32 = 0
	stateStarting int32 = 1
	stateRunning  int32 = 2
	stateStopping int32 = 3
)

// Default polling configuration.
const (
	DefaultMinPollInterval = 1 * time.Second
	DefaultMaxPollInterval = 30 * time.Second
)

// Errors returned by the poller.
var (
	ErrAlreadyRunning = core.ErrAlreadyRunning
	ErrNotRunning     = core.ErrNotRunning
)

// Handler wrapper types for pointer identity.
type (
	messageHandler      struct{ fn func(*LiveChatMessage) }
	deleteHandler       struct{ fn func(string) }
	banHandler          struct{ fn func(*UserBannedDetails) }
	errorHandler        struct{ fn func(error) }
	connectHandler      struct{ fn func() }
	disconnectHandler   struct{ fn func() }
	pollCompleteHandler struct{ fn func(int, time.Duration) }
)

// LiveChatPoller provides low-level HTTP polling for YouTube Live Chat.
type LiveChatPoller struct {
	client     *core.Client
	liveChatID string

	// Polling state
	mu              sync.RWMutex
	pageToken       string
	pollInterval    time.Duration
	minPollInterval time.Duration
	maxPollInterval time.Duration

	// Composable handlers (wrapper pointers for identity-based unsubscribe)
	handlerMu            sync.RWMutex
	messageHandlers      []*messageHandler
	deleteHandlers       []*deleteHandler
	banHandlers          []*banHandler
	errorHandlers        []*errorHandler
	connectHandlers      []*connectHandler
	disconnectHandlers   []*disconnectHandler
	pollCompleteHandlers []*pollCompleteHandler

	// Lifecycle (context-based cancellation)
	lifecycleMu sync.Mutex // Protects Start/Stop atomicity
	state       atomic.Int32
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	backoff     *core.BackoffConfig

	// Options
	profileImageSize string // Default, medium, high
}

// PollerOption configures a LiveChatPoller.
type PollerOption func(*LiveChatPoller)

// NewLiveChatPoller creates a new live chat poller.
func NewLiveChatPoller(client *core.Client, liveChatID string, opts ...PollerOption) *LiveChatPoller {
	p := &LiveChatPoller{
		client:           client,
		liveChatID:       liveChatID,
		minPollInterval:  DefaultMinPollInterval,
		maxPollInterval:  DefaultMaxPollInterval,
		backoff:          core.NewBackoffConfig(),
		profileImageSize: "default",
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// WithMinPollInterval sets the minimum time between polls.
func WithMinPollInterval(d time.Duration) PollerOption {
	return func(p *LiveChatPoller) { p.minPollInterval = d }
}

// WithMaxPollInterval sets the maximum time between polls.
func WithMaxPollInterval(d time.Duration) PollerOption {
	return func(p *LiveChatPoller) { p.maxPollInterval = d }
}

// WithBackoff sets a custom backoff configuration for retries.
// If cfg is nil, the default backoff configuration is retained.
func WithBackoff(cfg *core.BackoffConfig) PollerOption {
	return func(p *LiveChatPoller) {
		if cfg != nil {
			p.backoff = cfg
		}
	}
}

// Profile image size constants.
const (
	ProfileImageDefault = "default" // 88px
	ProfileImageMedium  = "medium"  // 240px
	ProfileImageHigh    = "high"    // 800px
)

// WithProfileImageSize sets the profile image size to request.
// Valid options: ProfileImageDefault (88px), ProfileImageMedium (240px), ProfileImageHigh (800px).
// Invalid values default to "default".
func WithProfileImageSize(size string) PollerOption {
	return func(p *LiveChatPoller) {
		switch size {
		case ProfileImageDefault, ProfileImageMedium, ProfileImageHigh:
			p.profileImageSize = size
		default:
			p.profileImageSize = ProfileImageDefault
		}
	}
}

// LiveChatID returns the live chat ID being polled.
func (p *LiveChatPoller) LiveChatID() string {
	return p.liveChatID
}

// IsRunning returns true if the poller is currently running.
func (p *LiveChatPoller) IsRunning() bool {
	return p.state.Load() == stateRunning
}

// Start begins polling for chat messages.
// The poller will run until Stop is called or the context is cancelled.
func (p *LiveChatPoller) Start(ctx context.Context) error {
	if p.liveChatID == "" {
		return errors.New("liveChatID cannot be empty")
	}

	p.lifecycleMu.Lock()
	defer p.lifecycleMu.Unlock()

	// Atomically transition from stopped → starting
	if !p.state.CompareAndSwap(stateStopped, stateStarting) {
		return ErrAlreadyRunning
	}

	// Create cancellable context
	pollCtx, cancel := context.WithCancel(ctx)
	p.cancel = cancel

	p.wg.Add(1)
	go p.pollLoop(pollCtx)

	// Transition to running (under lock, so Stop can't race)
	p.state.Store(stateRunning)
	return nil
}

// Stop stops the poller and waits for it to fully shut down.
// Safe to call multiple times (idempotent).
func (p *LiveChatPoller) Stop() {
	p.lifecycleMu.Lock()

	// Check if already stopped or stopping
	state := p.state.Load()
	if state == stateStopped || state == stateStopping {
		p.lifecycleMu.Unlock()
		return
	}

	// Transition to stopping
	p.state.Store(stateStopping)

	// Cancel context to signal all goroutines
	if p.cancel != nil {
		p.cancel()
	}

	p.lifecycleMu.Unlock()

	// Wait for all goroutines to finish (outside lock to avoid deadlock)
	p.wg.Wait()

	// Note: pollLoop's defer also sets stateStopped, but we set it here too
	// as a safety net in case Stop() is called before pollLoop starts.
	p.state.Store(stateStopped)
}

// pollLoop is the main polling goroutine.
func (p *LiveChatPoller) pollLoop(ctx context.Context) {
	defer p.wg.Done()
	defer p.state.Store(stateStopped) // Ensure state is stopped on exit

	// Notify connect handlers
	p.dispatchConnect()

	var attempt int

	for {
		select {
		case <-ctx.Done():
			p.dispatchDisconnect()
			return
		default:
		}

		// Perform poll
		messages, nextPoll, err := p.poll(ctx)

		if err != nil {
			// Check for chat ended
			var chatEnded *core.ChatEndedError
			if errors.As(err, &chatEnded) {
				p.dispatchError(err)
				p.dispatchDisconnect()
				return
			}

			// Check for context cancellation
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				p.dispatchDisconnect()
				return
			}

			// Dispatch error and apply backoff
			p.dispatchError(err)
			backoffDelay := p.backoff.Delay(attempt)
			attempt++

			select {
			case <-ctx.Done():
				p.dispatchDisconnect()
				return
			case <-time.After(backoffDelay):
				continue
			}
		}

		// Reset attempt counter on success
		attempt = 0

		// Dispatch messages
		p.dispatchMessages(messages)

		// Notify poll complete
		p.dispatchPollComplete(len(messages), nextPoll)

		// Wait for next poll interval
		select {
		case <-ctx.Done():
			p.dispatchDisconnect()
			return
		case <-time.After(nextPoll):
		}
	}
}

// poll performs a single poll request.
func (p *LiveChatPoller) poll(ctx context.Context) ([]*LiveChatMessage, time.Duration, error) {
	p.mu.RLock()
	pageToken := p.pageToken
	p.mu.RUnlock()

	query := url.Values{
		"liveChatId":       {p.liveChatID},
		"part":             {"id,snippet,authorDetails"},
		"profileImageSize": {p.profileImageSize},
	}
	if pageToken != "" {
		query.Set("pageToken", pageToken)
	}

	var resp LiveChatMessageListResponse
	err := p.client.Get(ctx, "liveChat/messages", query, "liveChatMessages.list", &resp)
	if err != nil {
		return nil, p.minPollInterval, err
	}

	// Check if chat has ended
	if resp.IsChatEnded() {
		return nil, 0, &core.ChatEndedError{LiveChatID: p.liveChatID}
	}

	// Update page token
	p.mu.Lock()
	p.pageToken = resp.NextPageToken
	p.mu.Unlock()

	// Calculate poll interval
	pollInterval := resp.PollingInterval()
	pollInterval = max(pollInterval, p.minPollInterval)
	pollInterval = min(pollInterval, p.maxPollInterval)

	p.mu.Lock()
	p.pollInterval = pollInterval
	p.mu.Unlock()

	return resp.Items, pollInterval, nil
}

// OnMessage registers a handler for chat messages.
// Returns an unsubscribe function that is safe to call multiple times.
func (p *LiveChatPoller) OnMessage(fn func(*LiveChatMessage)) func() {
	p.handlerMu.Lock()
	defer p.handlerMu.Unlock()

	h := &messageHandler{fn: fn}
	p.messageHandlers = append(p.messageHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			p.handlerMu.Lock()
			defer p.handlerMu.Unlock()
			for i, handler := range p.messageHandlers {
				if handler == h {
					p.messageHandlers = slices.Delete(p.messageHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnDelete registers a handler for message deletions.
// Returns an unsubscribe function.
func (p *LiveChatPoller) OnDelete(fn func(string)) func() {
	p.handlerMu.Lock()
	defer p.handlerMu.Unlock()

	h := &deleteHandler{fn: fn}
	p.deleteHandlers = append(p.deleteHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			p.handlerMu.Lock()
			defer p.handlerMu.Unlock()
			for i, handler := range p.deleteHandlers {
				if handler == h {
					p.deleteHandlers = slices.Delete(p.deleteHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnBan registers a handler for user bans.
// Returns an unsubscribe function.
func (p *LiveChatPoller) OnBan(fn func(*UserBannedDetails)) func() {
	p.handlerMu.Lock()
	defer p.handlerMu.Unlock()

	h := &banHandler{fn: fn}
	p.banHandlers = append(p.banHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			p.handlerMu.Lock()
			defer p.handlerMu.Unlock()
			for i, handler := range p.banHandlers {
				if handler == h {
					p.banHandlers = slices.Delete(p.banHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnError registers a handler for polling errors.
// Returns an unsubscribe function.
func (p *LiveChatPoller) OnError(fn func(error)) func() {
	p.handlerMu.Lock()
	defer p.handlerMu.Unlock()

	h := &errorHandler{fn: fn}
	p.errorHandlers = append(p.errorHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			p.handlerMu.Lock()
			defer p.handlerMu.Unlock()
			for i, handler := range p.errorHandlers {
				if handler == h {
					p.errorHandlers = slices.Delete(p.errorHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnConnect registers a handler called when polling starts.
// Returns an unsubscribe function.
func (p *LiveChatPoller) OnConnect(fn func()) func() {
	p.handlerMu.Lock()
	defer p.handlerMu.Unlock()

	h := &connectHandler{fn: fn}
	p.connectHandlers = append(p.connectHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			p.handlerMu.Lock()
			defer p.handlerMu.Unlock()
			for i, handler := range p.connectHandlers {
				if handler == h {
					p.connectHandlers = slices.Delete(p.connectHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnDisconnect registers a handler called when polling stops.
// Returns an unsubscribe function.
func (p *LiveChatPoller) OnDisconnect(fn func()) func() {
	p.handlerMu.Lock()
	defer p.handlerMu.Unlock()

	h := &disconnectHandler{fn: fn}
	p.disconnectHandlers = append(p.disconnectHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			p.handlerMu.Lock()
			defer p.handlerMu.Unlock()
			for i, handler := range p.disconnectHandlers {
				if handler == h {
					p.disconnectHandlers = slices.Delete(p.disconnectHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnPollComplete registers a handler called after each successful poll.
// The handler receives the message count and next poll interval.
// Returns an unsubscribe function.
func (p *LiveChatPoller) OnPollComplete(fn func(int, time.Duration)) func() {
	p.handlerMu.Lock()
	defer p.handlerMu.Unlock()

	h := &pollCompleteHandler{fn: fn}
	p.pollCompleteHandlers = append(p.pollCompleteHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			p.handlerMu.Lock()
			defer p.handlerMu.Unlock()
			for i, handler := range p.pollCompleteHandlers {
				if handler == h {
					p.pollCompleteHandlers = slices.Delete(p.pollCompleteHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// dispatchMessages sends messages to all handlers.
func (p *LiveChatPoller) dispatchMessages(messages []*LiveChatMessage) {
	// Snapshot handlers under read lock
	p.handlerMu.RLock()
	msgHandlers := make([]*messageHandler, len(p.messageHandlers))
	copy(msgHandlers, p.messageHandlers)
	delHandlers := make([]*deleteHandler, len(p.deleteHandlers))
	copy(delHandlers, p.deleteHandlers)
	banHandlers := make([]*banHandler, len(p.banHandlers))
	copy(banHandlers, p.banHandlers)
	p.handlerMu.RUnlock()

	for _, msg := range messages {
		// Handle message deletion events
		if msg.Type() == MessageTypeMessageDeleted && msg.Snippet != nil && msg.Snippet.MessageDeletedDetails != nil {
			deletedID := msg.Snippet.MessageDeletedDetails.DeletedMessageID
			for _, h := range delHandlers {
				p.safeCall(func() { h.fn(deletedID) })
			}
			continue
		}

		// Handle user ban events
		if msg.Type() == MessageTypeUserBanned && msg.Snippet != nil && msg.Snippet.UserBannedDetails != nil {
			for _, h := range banHandlers {
				p.safeCall(func() { h.fn(msg.Snippet.UserBannedDetails) })
			}
			continue
		}

		// Regular message
		for _, h := range msgHandlers {
			p.safeCall(func() { h.fn(msg) })
		}
	}
}

// dispatchError sends an error to all error handlers.
func (p *LiveChatPoller) dispatchError(err error) {
	p.handlerMu.RLock()
	handlers := make([]*errorHandler, len(p.errorHandlers))
	copy(handlers, p.errorHandlers)
	p.handlerMu.RUnlock()

	for _, h := range handlers {
		p.safeCall(func() { h.fn(err) })
	}
}

// dispatchConnect notifies all connect handlers.
func (p *LiveChatPoller) dispatchConnect() {
	p.handlerMu.RLock()
	handlers := make([]*connectHandler, len(p.connectHandlers))
	copy(handlers, p.connectHandlers)
	p.handlerMu.RUnlock()

	for _, h := range handlers {
		p.safeCall(func() { h.fn() })
	}
}

// dispatchDisconnect notifies all disconnect handlers.
func (p *LiveChatPoller) dispatchDisconnect() {
	p.handlerMu.RLock()
	handlers := make([]*disconnectHandler, len(p.disconnectHandlers))
	copy(handlers, p.disconnectHandlers)
	p.handlerMu.RUnlock()

	for _, h := range handlers {
		p.safeCall(func() { h.fn() })
	}
}

// dispatchPollComplete notifies all poll complete handlers.
func (p *LiveChatPoller) dispatchPollComplete(count int, interval time.Duration) {
	p.handlerMu.RLock()
	handlers := make([]*pollCompleteHandler, len(p.pollCompleteHandlers))
	copy(handlers, p.pollCompleteHandlers)
	p.handlerMu.RUnlock()

	for _, h := range handlers {
		p.safeCall(func() { h.fn(count, interval) })
	}
}

// safeCall executes a handler function with panic recovery.
func (p *LiveChatPoller) safeCall(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("handler panic: %v", r)
			// Dispatch error but don't recurse if error handler panics
			p.handlerMu.RLock()
			handlers := make([]*errorHandler, len(p.errorHandlers))
			copy(handlers, p.errorHandlers)
			p.handlerMu.RUnlock()

			for _, h := range handlers {
				func() {
					defer func() { _ = recover() }() // Silently ignore panic in error handler
					h.fn(err)
				}()
			}
		}
	}()
	fn()
}

// SendMessage sends a text message to the live chat.
func (p *LiveChatPoller) SendMessage(ctx context.Context, text string) (*LiveChatMessage, error) {
	if text == "" {
		return nil, errors.New("text cannot be empty")
	}

	req := &InsertMessageRequest{
		Snippet: &InsertMessageSnippet{
			LiveChatID: p.liveChatID,
			Type:       MessageTypeText,
			TextMessageDetails: &TextMessageDetails{
				MessageText: text,
			},
		},
	}

	query := url.Values{"part": {"snippet"}}

	var resp LiveChatMessage
	err := p.client.Post(ctx, "liveChat/messages", query, req, "liveChatMessages.insert", &resp)
	if err != nil {
		return nil, fmt.Errorf("sending message: %w", err)
	}

	return &resp, nil
}

// DeleteMessage deletes a message from the live chat.
func (p *LiveChatPoller) DeleteMessage(ctx context.Context, messageID string) error {
	if messageID == "" {
		return errors.New("messageID cannot be empty")
	}

	query := url.Values{"id": {messageID}}

	err := p.client.Delete(ctx, "liveChat/messages", query, "liveChatMessages.delete")
	if err != nil {
		return fmt.Errorf("deleting message: %w", err)
	}

	return nil
}

// BanUser permanently bans a user from the live chat.
func (p *LiveChatPoller) BanUser(ctx context.Context, channelID string) (*LiveChatBan, error) {
	if channelID == "" {
		return nil, errors.New("channelID cannot be empty")
	}
	return p.banUserInternal(ctx, channelID, BanTypePermanent, 0)
}

// TimeoutUser temporarily bans a user from the live chat.
func (p *LiveChatPoller) TimeoutUser(ctx context.Context, channelID string, seconds int64) (*LiveChatBan, error) {
	if channelID == "" {
		return nil, errors.New("channelID cannot be empty")
	}
	if seconds <= 0 {
		return nil, errors.New("seconds must be positive")
	}
	return p.banUserInternal(ctx, channelID, BanTypeTemporary, seconds)
}

// banUserInternal handles both permanent and temporary bans.
func (p *LiveChatPoller) banUserInternal(ctx context.Context, channelID, banType string, seconds int64) (*LiveChatBan, error) {
	req := &InsertBanRequest{
		Snippet: &InsertBanSnippet{
			LiveChatID: p.liveChatID,
			Type:       banType,
			BannedUserDetails: &BannedUserDetails{
				ChannelID: channelID,
			},
		},
	}

	if banType == BanTypeTemporary {
		req.Snippet.BanDurationSeconds = seconds
	}

	query := url.Values{"part": {"snippet"}}

	var resp LiveChatBan
	err := p.client.Post(ctx, "liveChat/bans", query, req, "liveChatBans.insert", &resp)
	if err != nil {
		return nil, fmt.Errorf("banning user: %w", err)
	}

	return &resp, nil
}

// UnbanUser removes a ban from the live chat.
func (p *LiveChatPoller) UnbanUser(ctx context.Context, banID string) error {
	if banID == "" {
		return errors.New("banID cannot be empty")
	}

	query := url.Values{"id": {banID}}

	err := p.client.Delete(ctx, "liveChat/bans", query, "liveChatBans.delete")
	if err != nil {
		return fmt.Errorf("unbanning user: %w", err)
	}

	return nil
}

// AddModerator adds a moderator to the live chat.
func (p *LiveChatPoller) AddModerator(ctx context.Context, channelID string) (*LiveChatModerator, error) {
	if channelID == "" {
		return nil, errors.New("channelID cannot be empty")
	}

	req := &InsertModeratorRequest{
		Snippet: &InsertModeratorSnippet{
			LiveChatID: p.liveChatID,
			ModeratorDetails: &ModeratorDetails{
				ChannelID: channelID,
			},
		},
	}

	query := url.Values{"part": {"snippet"}}

	var resp LiveChatModerator
	err := p.client.Post(ctx, "liveChat/moderators", query, req, "liveChatModerators.insert", &resp)
	if err != nil {
		return nil, fmt.Errorf("adding moderator: %w", err)
	}

	return &resp, nil
}

// RemoveModerator removes a moderator from the live chat.
func (p *LiveChatPoller) RemoveModerator(ctx context.Context, moderatorID string) error {
	if moderatorID == "" {
		return errors.New("moderatorID cannot be empty")
	}

	query := url.Values{"id": {moderatorID}}

	err := p.client.Delete(ctx, "liveChat/moderators", query, "liveChatModerators.delete")
	if err != nil {
		return fmt.Errorf("removing moderator: %w", err)
	}

	return nil
}

// PollInterval returns the current poll interval.
func (p *LiveChatPoller) PollInterval() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.pollInterval
}

// PageToken returns the current page token.
func (p *LiveChatPoller) PageToken() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.pageToken
}

// SetPageToken sets the page token (useful for resuming polling).
func (p *LiveChatPoller) SetPageToken(token string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pageToken = token
}

// ResetPageToken clears the page token.
// Call this before restarting a stopped poller to start fresh.
func (p *LiveChatPoller) ResetPageToken() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pageToken = ""
}

// Reset clears all polling state, preparing the poller for reuse.
// This clears the page token and poll interval.
// Returns ErrAlreadyRunning if the poller is currently running.
func (p *LiveChatPoller) Reset() error {
	if p.IsRunning() {
		return ErrAlreadyRunning
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pageToken = ""
	p.pollInterval = 0
	return nil
}
