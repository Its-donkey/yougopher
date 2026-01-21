package streaming

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Its-donkey/yougopher/youtube/core"
)

// SSE endpoint for live chat streaming.
const (
	// SSEBaseURL is the base URL for YouTube Live Streaming API SSE endpoint.
	SSEBaseURL = "https://www.googleapis.com/youtube/v3"

	// DefaultSSEReconnectDelay is the default delay before reconnecting after disconnect.
	DefaultSSEReconnectDelay = 1 * time.Second

	// DefaultSSEMaxReconnectDelay is the maximum backoff delay for reconnection attempts.
	DefaultSSEMaxReconnectDelay = 30 * time.Second
)

// sseMessageHandler handles incoming messages from SSE stream.
type sseMessageHandler struct{ fn func(*LiveChatMessage) }
type sseDeleteHandler struct{ fn func(string) }
type sseBanHandler struct{ fn func(*UserBannedDetails) }
type sseErrorHandler struct{ fn func(error) }
type sseConnectHandler struct{ fn func() }
type sseDisconnectHandler struct{ fn func() }
type sseResponseHandler struct{ fn func(*LiveChatMessageListResponse) }

// LiveChatStream provides Server-Sent Events (SSE) streaming for YouTube Live Chat.
// It establishes a long-lived HTTP connection to receive chat messages in real-time
// with lower latency than polling.
type LiveChatStream struct {
	client     *core.Client
	httpClient *http.Client
	liveChatID string
	baseURL    string

	// Configuration
	mu                  sync.RWMutex
	parts               []string
	hl                  string
	maxResults          int
	profileImageSize    int
	pageToken           string
	reconnectDelay      time.Duration
	maxReconnectDelay   time.Duration

	// Handlers
	handlerMu           sync.RWMutex
	messageHandlers     []*sseMessageHandler
	deleteHandlers      []*sseDeleteHandler
	banHandlers         []*sseBanHandler
	errorHandlers       []*sseErrorHandler
	connectHandlers     []*sseConnectHandler
	disconnectHandlers  []*sseDisconnectHandler
	responseHandlers    []*sseResponseHandler

	// Lifecycle
	lifecycleMu sync.Mutex
	state       atomic.Int32
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	backoff     *core.BackoffConfig

	// Authentication
	accessToken string
}

// StreamOption configures a LiveChatStream.
type StreamOption func(*LiveChatStream)

// NewLiveChatStream creates a new SSE-based live chat stream.
func NewLiveChatStream(client *core.Client, liveChatID string, opts ...StreamOption) *LiveChatStream {
	s := &LiveChatStream{
		client:            client,
		httpClient:        &http.Client{},
		liveChatID:        liveChatID,
		baseURL:           SSEBaseURL,
		parts:             []string{"id", "snippet", "authorDetails"},
		maxResults:        500,
		profileImageSize:  88,
		reconnectDelay:    DefaultSSEReconnectDelay,
		maxReconnectDelay: DefaultSSEMaxReconnectDelay,
		backoff:           core.NewBackoffConfig(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// WithStreamHTTPClient sets a custom HTTP client for the stream.
func WithStreamHTTPClient(hc *http.Client) StreamOption {
	return func(s *LiveChatStream) {
		if hc != nil {
			s.httpClient = hc
		}
	}
}

// WithStreamParts sets the resource parts to include in responses.
// Valid parts: "id", "snippet", "authorDetails".
func WithStreamParts(parts ...string) StreamOption {
	return func(s *LiveChatStream) {
		if len(parts) > 0 {
			s.parts = parts
		}
	}
}

// WithStreamHL sets the language code for localized currency display.
func WithStreamHL(hl string) StreamOption {
	return func(s *LiveChatStream) { s.hl = hl }
}

// WithStreamMaxResults sets the maximum messages per response (200-2000, default 500).
func WithStreamMaxResults(n int) StreamOption {
	return func(s *LiveChatStream) {
		if n >= 200 && n <= 2000 {
			s.maxResults = n
		}
	}
}

// WithStreamProfileImageSize sets the profile image dimensions (16-720, default 88).
func WithStreamProfileImageSize(size int) StreamOption {
	return func(s *LiveChatStream) {
		if size >= 16 && size <= 720 {
			s.profileImageSize = size
		}
	}
}

// WithStreamReconnectDelay sets the initial reconnect delay.
func WithStreamReconnectDelay(d time.Duration) StreamOption {
	return func(s *LiveChatStream) { s.reconnectDelay = d }
}

// WithStreamMaxReconnectDelay sets the maximum backoff delay for reconnection.
func WithStreamMaxReconnectDelay(d time.Duration) StreamOption {
	return func(s *LiveChatStream) { s.maxReconnectDelay = d }
}

// WithStreamBackoff sets a custom backoff configuration.
func WithStreamBackoff(cfg *core.BackoffConfig) StreamOption {
	return func(s *LiveChatStream) {
		if cfg != nil {
			s.backoff = cfg
		}
	}
}

// WithStreamAccessToken sets the OAuth access token for authentication.
func WithStreamAccessToken(token string) StreamOption {
	return func(s *LiveChatStream) { s.accessToken = token }
}

// WithStreamBaseURL sets a custom base URL (useful for testing).
func WithStreamBaseURL(url string) StreamOption {
	return func(s *LiveChatStream) {
		if url != "" {
			s.baseURL = strings.TrimSuffix(url, "/")
		}
	}
}

// LiveChatID returns the live chat ID being streamed.
func (s *LiveChatStream) LiveChatID() string {
	return s.liveChatID
}

// IsRunning returns true if the stream is currently connected.
func (s *LiveChatStream) IsRunning() bool {
	return s.state.Load() == stateRunning
}

// Start begins the SSE stream connection.
// The stream will run until Stop is called or the context is cancelled.
func (s *LiveChatStream) Start(ctx context.Context) error {
	if s.liveChatID == "" {
		return errors.New("liveChatID cannot be empty")
	}

	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	if !s.state.CompareAndSwap(stateStopped, stateStarting) {
		return ErrAlreadyRunning
	}

	streamCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	s.wg.Add(1)
	go s.streamLoop(streamCtx)

	s.state.Store(stateRunning)
	return nil
}

// Stop stops the stream and waits for it to fully shut down.
func (s *LiveChatStream) Stop() {
	s.lifecycleMu.Lock()

	state := s.state.Load()
	if state == stateStopped || state == stateStopping {
		s.lifecycleMu.Unlock()
		return
	}

	s.state.Store(stateStopping)

	if s.cancel != nil {
		s.cancel()
	}

	s.lifecycleMu.Unlock()

	s.wg.Wait()
	s.state.Store(stateStopped)
}

// streamLoop is the main streaming goroutine.
func (s *LiveChatStream) streamLoop(ctx context.Context) {
	defer s.wg.Done()
	defer s.state.Store(stateStopped)

	s.dispatchConnect()

	var attempt int

	for {
		select {
		case <-ctx.Done():
			s.dispatchDisconnect()
			return
		default:
		}

		err := s.connect(ctx)

		if err != nil {
			// Check for chat ended
			var chatEnded *core.ChatEndedError
			if errors.As(err, &chatEnded) {
				s.dispatchError(err)
				s.dispatchDisconnect()
				return
			}

			// Check for context cancellation
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				s.dispatchDisconnect()
				return
			}

			// Dispatch error and apply backoff
			s.dispatchError(err)
			backoffDelay := min(s.backoff.Delay(attempt), s.maxReconnectDelay)
			attempt++

			select {
			case <-ctx.Done():
				s.dispatchDisconnect()
				return
			case <-time.After(backoffDelay):
				continue
			}
		}

		// Reset attempt counter on successful connection
		attempt = 0
	}
}

// connect establishes the SSE connection and processes events.
func (s *LiveChatStream) connect(ctx context.Context) error {
	s.mu.RLock()
	pageToken := s.pageToken
	s.mu.RUnlock()

	req, err := s.buildRequest(ctx, pageToken)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connecting: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Track quota usage
	if qt := s.client.QuotaTracker(); qt != nil {
		qt.Add("liveChatMessages.streamList", 1)
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		return s.handleErrorResponse(resp)
	}

	// Process SSE events
	return s.processEvents(ctx, resp.Body)
}

// buildRequest creates the HTTP request for the SSE endpoint.
func (s *LiveChatStream) buildRequest(ctx context.Context, pageToken string) (*http.Request, error) {
	u, err := url.Parse(s.baseURL + "/liveChat/messages/stream")
	if err != nil {
		return nil, err
	}

	query := u.Query()
	query.Set("liveChatId", s.liveChatID)
	query.Set("part", strings.Join(s.parts, ","))
	query.Set("maxResults", fmt.Sprintf("%d", s.maxResults))
	query.Set("profileImageSize", fmt.Sprintf("%d", s.profileImageSize))

	if s.hl != "" {
		query.Set("hl", s.hl)
	}
	if pageToken != "" {
		query.Set("pageToken", pageToken)
	}

	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set SSE headers
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("User-Agent", "Yougopher/1.0")

	// Add authorization
	// Note: We access the token via the client's internal method
	// For now, we need to get the token from the core.Client
	// This requires adding a method to expose it or using a different approach
	if token := s.getAccessToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}

// getAccessToken retrieves the access token.
func (s *LiveChatStream) getAccessToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accessToken
}

// SetAccessToken sets the OAuth access token for authentication.
func (s *LiveChatStream) SetAccessToken(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessToken = token
}

// processEvents reads and processes SSE events from the stream.
func (s *LiveChatStream) processEvents(ctx context.Context, body io.Reader) error {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024) // 1MB max line size

	var eventData strings.Builder

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()

		// Empty line indicates end of event
		if line == "" {
			if eventData.Len() > 0 {
				if err := s.processEvent(eventData.String()); err != nil {
					// Log error but continue processing
					s.dispatchError(fmt.Errorf("processing event: %w", err))
				}
				eventData.Reset()
			}
			continue
		}

		// Parse SSE field
		if data, ok := strings.CutPrefix(line, "data:"); ok {
			eventData.WriteString(strings.TrimSpace(data))
		} else if retryStr, ok := strings.CutPrefix(line, "retry:"); ok {
			// Handle retry directive (reconnection delay in ms)
			retryStr = strings.TrimSpace(retryStr)
			if retryMs, err := parseUint(retryStr); err == nil {
				s.mu.Lock()
				s.reconnectDelay = time.Duration(retryMs) * time.Millisecond
				s.mu.Unlock()
			}
		}
		// Ignore other fields (event:, id:, comments starting with :)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading stream: %w", err)
	}

	// Stream ended normally
	return nil
}

// parseUint parses a string as unsigned integer.
func parseUint(s string) (int64, error) {
	var result int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errors.New("invalid number")
		}
		result = result*10 + int64(c-'0')
	}
	return result, nil
}

// processEvent processes a single SSE event data.
func (s *LiveChatStream) processEvent(data string) error {
	var resp LiveChatMessageListResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	// Update page token for reconnection
	if resp.NextPageToken != "" {
		s.mu.Lock()
		s.pageToken = resp.NextPageToken
		s.mu.Unlock()
	}

	// Check if chat has ended
	if resp.IsChatEnded() {
		return &core.ChatEndedError{LiveChatID: s.liveChatID}
	}

	// Dispatch the full response to response handlers
	s.dispatchResponse(&resp)

	// Dispatch individual messages
	s.dispatchMessages(resp.Items)

	return nil
}

// handleErrorResponse parses an error response from the API.
func (s *LiveChatStream) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return fmt.Errorf("reading error response: %w", err)
	}

	// Try to parse as YouTube API error
	var errResp core.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != nil {
		apiErr := errResp.ToAPIError()
		apiErr.StatusCode = resp.StatusCode

		// Check for specific error types
		if apiErr.IsQuotaExceeded() {
			return &core.QuotaError{
				Used:    s.quotaUsed(),
				Limit:   s.quotaLimit(),
				ResetAt: time.Now().Add(24 * time.Hour), // Approximate
			}
		}

		if apiErr.IsRateLimited() {
			return &core.RateLimitError{
				RetryAfter: 1 * time.Second,
				Code:       apiErr.Code,
				Message:    apiErr.Message,
			}
		}

		// Check for chat ended/disabled errors
		if apiErr.IsChatEnded() || apiErr.IsChatDisabled() {
			return &core.ChatEndedError{LiveChatID: s.liveChatID}
		}

		return apiErr
	}

	return &core.APIError{
		StatusCode: resp.StatusCode,
		Message:    string(body),
	}
}

// quotaUsed returns current quota usage.
func (s *LiveChatStream) quotaUsed() int {
	if qt := s.client.QuotaTracker(); qt != nil {
		return qt.Used()
	}
	return 0
}

// quotaLimit returns quota limit.
func (s *LiveChatStream) quotaLimit() int {
	if qt := s.client.QuotaTracker(); qt != nil {
		return qt.Limit()
	}
	return core.DefaultDailyQuota
}

// OnMessage registers a handler for chat messages.
func (s *LiveChatStream) OnMessage(fn func(*LiveChatMessage)) func() {
	s.handlerMu.Lock()
	defer s.handlerMu.Unlock()

	h := &sseMessageHandler{fn: fn}
	s.messageHandlers = append(s.messageHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			s.handlerMu.Lock()
			defer s.handlerMu.Unlock()
			for i, handler := range s.messageHandlers {
				if handler == h {
					s.messageHandlers = slices.Delete(s.messageHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnDelete registers a handler for message deletions.
func (s *LiveChatStream) OnDelete(fn func(string)) func() {
	s.handlerMu.Lock()
	defer s.handlerMu.Unlock()

	h := &sseDeleteHandler{fn: fn}
	s.deleteHandlers = append(s.deleteHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			s.handlerMu.Lock()
			defer s.handlerMu.Unlock()
			for i, handler := range s.deleteHandlers {
				if handler == h {
					s.deleteHandlers = slices.Delete(s.deleteHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnBan registers a handler for user bans.
func (s *LiveChatStream) OnBan(fn func(*UserBannedDetails)) func() {
	s.handlerMu.Lock()
	defer s.handlerMu.Unlock()

	h := &sseBanHandler{fn: fn}
	s.banHandlers = append(s.banHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			s.handlerMu.Lock()
			defer s.handlerMu.Unlock()
			for i, handler := range s.banHandlers {
				if handler == h {
					s.banHandlers = slices.Delete(s.banHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnError registers a handler for stream errors.
func (s *LiveChatStream) OnError(fn func(error)) func() {
	s.handlerMu.Lock()
	defer s.handlerMu.Unlock()

	h := &sseErrorHandler{fn: fn}
	s.errorHandlers = append(s.errorHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			s.handlerMu.Lock()
			defer s.handlerMu.Unlock()
			for i, handler := range s.errorHandlers {
				if handler == h {
					s.errorHandlers = slices.Delete(s.errorHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnConnect registers a handler called when the stream connects.
func (s *LiveChatStream) OnConnect(fn func()) func() {
	s.handlerMu.Lock()
	defer s.handlerMu.Unlock()

	h := &sseConnectHandler{fn: fn}
	s.connectHandlers = append(s.connectHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			s.handlerMu.Lock()
			defer s.handlerMu.Unlock()
			for i, handler := range s.connectHandlers {
				if handler == h {
					s.connectHandlers = slices.Delete(s.connectHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnDisconnect registers a handler called when the stream disconnects.
func (s *LiveChatStream) OnDisconnect(fn func()) func() {
	s.handlerMu.Lock()
	defer s.handlerMu.Unlock()

	h := &sseDisconnectHandler{fn: fn}
	s.disconnectHandlers = append(s.disconnectHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			s.handlerMu.Lock()
			defer s.handlerMu.Unlock()
			for i, handler := range s.disconnectHandlers {
				if handler == h {
					s.disconnectHandlers = slices.Delete(s.disconnectHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// OnResponse registers a handler for each full response received.
// This is useful for accessing metadata like offlineAt.
func (s *LiveChatStream) OnResponse(fn func(*LiveChatMessageListResponse)) func() {
	s.handlerMu.Lock()
	defer s.handlerMu.Unlock()

	h := &sseResponseHandler{fn: fn}
	s.responseHandlers = append(s.responseHandlers, h)

	var once sync.Once
	return func() {
		once.Do(func() {
			s.handlerMu.Lock()
			defer s.handlerMu.Unlock()
			for i, handler := range s.responseHandlers {
				if handler == h {
					s.responseHandlers = slices.Delete(s.responseHandlers, i, i+1)
					return
				}
			}
		})
	}
}

// dispatchMessages sends messages to all handlers.
func (s *LiveChatStream) dispatchMessages(messages []*LiveChatMessage) {
	s.handlerMu.RLock()
	msgHandlers := make([]*sseMessageHandler, len(s.messageHandlers))
	copy(msgHandlers, s.messageHandlers)
	delHandlers := make([]*sseDeleteHandler, len(s.deleteHandlers))
	copy(delHandlers, s.deleteHandlers)
	banHandlers := make([]*sseBanHandler, len(s.banHandlers))
	copy(banHandlers, s.banHandlers)
	s.handlerMu.RUnlock()

	for _, msg := range messages {
		// Handle message deletion events
		if msg.Type() == MessageTypeMessageDeleted && msg.Snippet != nil && msg.Snippet.MessageDeletedDetails != nil {
			deletedID := msg.Snippet.MessageDeletedDetails.DeletedMessageID
			for _, h := range delHandlers {
				s.safeCall(func() { h.fn(deletedID) })
			}
			continue
		}

		// Handle user ban events
		if msg.Type() == MessageTypeUserBanned && msg.Snippet != nil && msg.Snippet.UserBannedDetails != nil {
			for _, h := range banHandlers {
				s.safeCall(func() { h.fn(msg.Snippet.UserBannedDetails) })
			}
			continue
		}

		// Regular message
		for _, h := range msgHandlers {
			s.safeCall(func() { h.fn(msg) })
		}
	}
}

// dispatchResponse sends the full response to response handlers.
func (s *LiveChatStream) dispatchResponse(resp *LiveChatMessageListResponse) {
	s.handlerMu.RLock()
	handlers := make([]*sseResponseHandler, len(s.responseHandlers))
	copy(handlers, s.responseHandlers)
	s.handlerMu.RUnlock()

	for _, h := range handlers {
		s.safeCall(func() { h.fn(resp) })
	}
}

// dispatchError sends an error to all error handlers.
func (s *LiveChatStream) dispatchError(err error) {
	s.handlerMu.RLock()
	handlers := make([]*sseErrorHandler, len(s.errorHandlers))
	copy(handlers, s.errorHandlers)
	s.handlerMu.RUnlock()

	for _, h := range handlers {
		s.safeCall(func() { h.fn(err) })
	}
}

// dispatchConnect notifies all connect handlers.
func (s *LiveChatStream) dispatchConnect() {
	s.handlerMu.RLock()
	handlers := make([]*sseConnectHandler, len(s.connectHandlers))
	copy(handlers, s.connectHandlers)
	s.handlerMu.RUnlock()

	for _, h := range handlers {
		s.safeCall(func() { h.fn() })
	}
}

// dispatchDisconnect notifies all disconnect handlers.
func (s *LiveChatStream) dispatchDisconnect() {
	s.handlerMu.RLock()
	handlers := make([]*sseDisconnectHandler, len(s.disconnectHandlers))
	copy(handlers, s.disconnectHandlers)
	s.handlerMu.RUnlock()

	for _, h := range handlers {
		s.safeCall(func() { h.fn() })
	}
}

// safeCall executes a handler function with panic recovery.
func (s *LiveChatStream) safeCall(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("handler panic: %v", r)
			s.handlerMu.RLock()
			handlers := make([]*sseErrorHandler, len(s.errorHandlers))
			copy(handlers, s.errorHandlers)
			s.handlerMu.RUnlock()

			for _, h := range handlers {
				func() {
					defer func() { _ = recover() }()
					h.fn(err)
				}()
			}
		}
	}()
	fn()
}

// PageToken returns the current page token (for resumption).
func (s *LiveChatStream) PageToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pageToken
}

// SetPageToken sets the page token (useful for resuming streams).
func (s *LiveChatStream) SetPageToken(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pageToken = token
}

// ResetPageToken clears the page token.
func (s *LiveChatStream) ResetPageToken() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pageToken = ""
}

// Reset clears all streaming state, preparing the stream for reuse.
func (s *LiveChatStream) Reset() error {
	if s.IsRunning() {
		return ErrAlreadyRunning
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pageToken = ""
	return nil
}
