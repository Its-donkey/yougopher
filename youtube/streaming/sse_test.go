package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Its-donkey/yougopher/youtube/core"
)

func TestNewLiveChatStream(t *testing.T) {
	client := core.NewClient()
	stream := NewLiveChatStream(client, "chat123")

	if stream.LiveChatID() != "chat123" {
		t.Errorf("LiveChatID() = %q, want 'chat123'", stream.LiveChatID())
	}
	if stream.maxResults != 500 {
		t.Errorf("maxResults = %d, want 500", stream.maxResults)
	}
	if stream.profileImageSize != 88 {
		t.Errorf("profileImageSize = %d, want 88", stream.profileImageSize)
	}
}

func TestNewLiveChatStream_WithOptions(t *testing.T) {
	client := core.NewClient()
	backoff := core.NewBackoffConfig(core.WithBaseDelay(2 * time.Second))

	stream := NewLiveChatStream(client, "chat123",
		WithStreamParts("id", "snippet"),
		WithStreamHL("en"),
		WithStreamMaxResults(1000),
		WithStreamProfileImageSize(240),
		WithStreamReconnectDelay(2*time.Second),
		WithStreamMaxReconnectDelay(60*time.Second),
		WithStreamBackoff(backoff),
		WithStreamAccessToken("test-token"),
	)

	if len(stream.parts) != 2 || stream.parts[0] != "id" || stream.parts[1] != "snippet" {
		t.Errorf("parts = %v, want ['id', 'snippet']", stream.parts)
	}
	if stream.hl != "en" {
		t.Errorf("hl = %q, want 'en'", stream.hl)
	}
	if stream.maxResults != 1000 {
		t.Errorf("maxResults = %d, want 1000", stream.maxResults)
	}
	if stream.profileImageSize != 240 {
		t.Errorf("profileImageSize = %d, want 240", stream.profileImageSize)
	}
	if stream.reconnectDelay != 2*time.Second {
		t.Errorf("reconnectDelay = %v, want 2s", stream.reconnectDelay)
	}
	if stream.maxReconnectDelay != 60*time.Second {
		t.Errorf("maxReconnectDelay = %v, want 60s", stream.maxReconnectDelay)
	}
	if stream.accessToken != "test-token" {
		t.Errorf("accessToken = %q, want 'test-token'", stream.accessToken)
	}
}

func TestLiveChatStream_MaxResultsBounds(t *testing.T) {
	client := core.NewClient()

	// Test below minimum (200)
	stream := NewLiveChatStream(client, "chat123", WithStreamMaxResults(100))
	if stream.maxResults != 500 { // Should keep default
		t.Errorf("maxResults = %d, want 500 (unchanged for invalid value)", stream.maxResults)
	}

	// Test above maximum (2000)
	stream = NewLiveChatStream(client, "chat123", WithStreamMaxResults(3000))
	if stream.maxResults != 500 { // Should keep default
		t.Errorf("maxResults = %d, want 500 (unchanged for invalid value)", stream.maxResults)
	}

	// Test valid value
	stream = NewLiveChatStream(client, "chat123", WithStreamMaxResults(1500))
	if stream.maxResults != 1500 {
		t.Errorf("maxResults = %d, want 1500", stream.maxResults)
	}
}

func TestLiveChatStream_ProfileImageSizeBounds(t *testing.T) {
	client := core.NewClient()

	// Test below minimum (16)
	stream := NewLiveChatStream(client, "chat123", WithStreamProfileImageSize(10))
	if stream.profileImageSize != 88 { // Should keep default
		t.Errorf("profileImageSize = %d, want 88 (unchanged for invalid value)", stream.profileImageSize)
	}

	// Test above maximum (720)
	stream = NewLiveChatStream(client, "chat123", WithStreamProfileImageSize(1000))
	if stream.profileImageSize != 88 { // Should keep default
		t.Errorf("profileImageSize = %d, want 88 (unchanged for invalid value)", stream.profileImageSize)
	}

	// Test valid value
	stream = NewLiveChatStream(client, "chat123", WithStreamProfileImageSize(500))
	if stream.profileImageSize != 500 {
		t.Errorf("profileImageSize = %d, want 500", stream.profileImageSize)
	}
}

// newSSEServer creates a test server that responds with SSE events.
func newSSEServer(events []string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE not supported", http.StatusInternalServerError)
			return
		}

		for _, event := range events {
			_, _ = fmt.Fprintf(w, "data: %s\n\n", event)
			flusher.Flush()
		}
	}))
}

func TestLiveChatStream_StartStop(t *testing.T) {
	resp := LiveChatMessageListResponse{
		NextPageToken:         "token1",
		PollingIntervalMillis: 1000,
		Items:                 []*LiveChatMessage{},
	}
	data, _ := json.Marshal(resp)

	server := newSSEServer([]string{string(data)})
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	stream := NewLiveChatStream(client, "chat123",
		WithStreamAccessToken("test-token"),
		WithStreamBaseURL(server.URL),
	)

	ctx := context.Background()

	// Should not be running initially
	if stream.IsRunning() {
		t.Error("IsRunning() = true before Start()")
	}

	// Start streaming
	err := stream.Start(ctx)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Should be running now
	if !stream.IsRunning() {
		t.Error("IsRunning() = false after Start()")
	}

	// Starting again should fail
	err = stream.Start(ctx)
	if err != ErrAlreadyRunning {
		t.Errorf("Start() again error = %v, want ErrAlreadyRunning", err)
	}

	// Stop streaming
	stream.Stop()

	// Should not be running after stop
	if stream.IsRunning() {
		t.Error("IsRunning() = true after Stop()")
	}

	// Stop again should be safe (idempotent)
	stream.Stop()
}

func TestLiveChatStream_EmptyLiveChatID(t *testing.T) {
	client := core.NewClient()
	stream := NewLiveChatStream(client, "")

	err := stream.Start(context.Background())
	if err == nil {
		t.Error("Start() with empty liveChatID should fail")
	}
	if err.Error() != "liveChatID cannot be empty" {
		t.Errorf("error = %q, want 'liveChatID cannot be empty'", err.Error())
	}
}

func TestLiveChatStream_OnConnect(t *testing.T) {
	var connectCalled atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		// Just hang to simulate a long-lived connection
		<-r.Context().Done()
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	stream := NewLiveChatStream(client, "chat123",
		WithStreamAccessToken("test-token"),
		WithStreamBaseURL(server.URL),
	)

	stream.OnConnect(func() {
		connectCalled.Store(true)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = stream.Start(ctx)

	// Wait for connect handler to be called
	time.Sleep(50 * time.Millisecond)

	if !connectCalled.Load() {
		t.Error("OnConnect handler was not called")
	}

	stream.Stop()
}

func TestLiveChatStream_OnDisconnect(t *testing.T) {
	var disconnectCalled atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		// Return immediately to trigger disconnect
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	stream := NewLiveChatStream(client, "chat123",
		WithStreamAccessToken("test-token"),
		WithStreamReconnectDelay(50*time.Millisecond),
		WithStreamBaseURL(server.URL),
	)

	stream.OnDisconnect(func() {
		disconnectCalled.Store(true)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = stream.Start(ctx)

	// Wait for stream to process and context to cancel
	<-ctx.Done()
	stream.Stop()

	if !disconnectCalled.Load() {
		t.Error("OnDisconnect handler was not called")
	}
}

func TestLiveChatStream_OnMessage(t *testing.T) {
	var receivedMessages []*LiveChatMessage
	var mu sync.Mutex

	resp := LiveChatMessageListResponse{
		NextPageToken:         "token1",
		PollingIntervalMillis: 100,
		Items: []*LiveChatMessage{
			{
				ID: "msg1",
				Snippet: &MessageSnippet{
					Type:           MessageTypeText,
					DisplayMessage: "Hello from SSE",
				},
			},
			{
				ID: "msg2",
				Snippet: &MessageSnippet{
					Type:           MessageTypeText,
					DisplayMessage: "Another message",
				},
			},
		},
	}
	data, _ := json.Marshal(resp)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE not supported", http.StatusInternalServerError)
			return
		}
		_, _ = fmt.Fprintf(w, "data: %s\n\n", string(data))
		flusher.Flush()
		// Wait for context to be cancelled
		<-r.Context().Done()
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	stream := NewLiveChatStream(client, "chat123",
		WithStreamAccessToken("test-token"),
		WithStreamBaseURL(server.URL),
	)

	stream.OnMessage(func(msg *LiveChatMessage) {
		mu.Lock()
		receivedMessages = append(receivedMessages, msg)
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = stream.Start(ctx)

	// Wait for messages to be received
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	count := len(receivedMessages)
	mu.Unlock()

	if count != 2 {
		t.Errorf("received %d messages, want 2", count)
	}

	stream.Stop()
}

func TestLiveChatStream_OnResponse(t *testing.T) {
	var receivedResponse *LiveChatMessageListResponse
	var mu sync.Mutex

	resp := LiveChatMessageListResponse{
		NextPageToken:         "token123",
		PollingIntervalMillis: 1000,
		Items: []*LiveChatMessage{
			{ID: "msg1"},
		},
	}
	data, _ := json.Marshal(resp)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}
		_, _ = fmt.Fprintf(w, "data: %s\n\n", string(data))
		flusher.Flush()
		<-r.Context().Done()
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	stream := NewLiveChatStream(client, "chat123",
		WithStreamAccessToken("test-token"),
		WithStreamBaseURL(server.URL),
	)

	stream.OnResponse(func(r *LiveChatMessageListResponse) {
		mu.Lock()
		receivedResponse = r
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = stream.Start(ctx)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if receivedResponse == nil {
		t.Error("OnResponse handler was not called")
	} else if receivedResponse.NextPageToken != "token123" {
		t.Errorf("NextPageToken = %q, want 'token123'", receivedResponse.NextPageToken)
	}

	stream.Stop()
}

func TestLiveChatStream_OnError(t *testing.T) {
	var receivedError error
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		errResp := core.ErrorResponse{
			Error: &core.ErrorBody{
				Code:    403,
				Message: "Access denied",
				Errors: []core.ErrorItem{
					{Reason: "forbidden", Domain: "youtube"},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(errResp)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	stream := NewLiveChatStream(client, "chat123",
		WithStreamAccessToken("test-token"),
		WithStreamReconnectDelay(50*time.Millisecond),
		WithStreamBaseURL(server.URL),
	)

	stream.OnError(func(err error) {
		mu.Lock()
		receivedError = err
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = stream.Start(ctx)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	err := receivedError
	mu.Unlock()

	if err == nil {
		t.Error("OnError handler was not called")
	}

	stream.Stop()
}

func TestLiveChatStream_OnDelete(t *testing.T) {
	var deletedIDs []string
	var mu sync.Mutex

	resp := LiveChatMessageListResponse{
		NextPageToken: "token1",
		Items: []*LiveChatMessage{
			{
				ID: "delete-event-1",
				Snippet: &MessageSnippet{
					Type: MessageTypeMessageDeleted,
					MessageDeletedDetails: &MessageDeletedDetails{
						DeletedMessageID: "msg-to-delete",
					},
				},
			},
		},
	}
	data, _ := json.Marshal(resp)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}
		_, _ = fmt.Fprintf(w, "data: %s\n\n", string(data))
		flusher.Flush()
		<-r.Context().Done()
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	stream := NewLiveChatStream(client, "chat123",
		WithStreamAccessToken("test-token"),
		WithStreamBaseURL(server.URL),
	)

	stream.OnDelete(func(id string) {
		mu.Lock()
		deletedIDs = append(deletedIDs, id)
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = stream.Start(ctx)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	count := len(deletedIDs)
	mu.Unlock()

	if count != 1 {
		t.Errorf("OnDelete called %d times, want 1", count)
	}

	stream.Stop()
}

func TestLiveChatStream_OnBan(t *testing.T) {
	var bannedUsers []*UserBannedDetails
	var mu sync.Mutex

	resp := LiveChatMessageListResponse{
		NextPageToken: "token1",
		Items: []*LiveChatMessage{
			{
				ID: "ban-event-1",
				Snippet: &MessageSnippet{
					Type: MessageTypeUserBanned,
					UserBannedDetails: &UserBannedDetails{
						BannedUserDetails: &BannedUserDetails{
							ChannelID:   "banned-channel",
							DisplayName: "Banned User",
						},
						BanType: BanTypePermanent,
					},
				},
			},
		},
	}
	data, _ := json.Marshal(resp)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}
		_, _ = fmt.Fprintf(w, "data: %s\n\n", string(data))
		flusher.Flush()
		<-r.Context().Done()
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	stream := NewLiveChatStream(client, "chat123",
		WithStreamAccessToken("test-token"),
		WithStreamBaseURL(server.URL),
	)

	stream.OnBan(func(details *UserBannedDetails) {
		mu.Lock()
		bannedUsers = append(bannedUsers, details)
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = stream.Start(ctx)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	count := len(bannedUsers)
	mu.Unlock()

	if count != 1 {
		t.Errorf("OnBan called %d times, want 1", count)
	}

	stream.Stop()
}

func TestLiveChatStream_PageToken(t *testing.T) {
	client := core.NewClient()
	stream := NewLiveChatStream(client, "chat123")

	// Initial state
	if token := stream.PageToken(); token != "" {
		t.Errorf("initial PageToken() = %q, want ''", token)
	}

	// Set page token
	stream.SetPageToken("token123")
	if token := stream.PageToken(); token != "token123" {
		t.Errorf("PageToken() = %q, want 'token123'", token)
	}

	// Reset page token
	stream.ResetPageToken()
	if token := stream.PageToken(); token != "" {
		t.Errorf("PageToken() after reset = %q, want ''", token)
	}
}

func TestLiveChatStream_Reset(t *testing.T) {
	client := core.NewClient()
	stream := NewLiveChatStream(client, "chat123")

	stream.SetPageToken("token123")

	err := stream.Reset()
	if err != nil {
		t.Errorf("Reset() error = %v", err)
	}

	if token := stream.PageToken(); token != "" {
		t.Errorf("PageToken() after Reset() = %q, want ''", token)
	}
}

func TestLiveChatStream_Unsubscribe(t *testing.T) {
	client := core.NewClient()
	stream := NewLiveChatStream(client, "chat123")

	var count int
	unsub := stream.OnMessage(func(msg *LiveChatMessage) {
		count++
	})

	// Unsubscribe
	unsub()

	// Verify handler count decreased
	stream.handlerMu.RLock()
	handlerCount := len(stream.messageHandlers)
	stream.handlerMu.RUnlock()

	if handlerCount != 0 {
		t.Errorf("handler count after unsubscribe = %d, want 0", handlerCount)
	}

	// Unsubscribe again should be safe
	unsub()
}

func TestLiveChatStream_AccessToken(t *testing.T) {
	client := core.NewClient()
	stream := NewLiveChatStream(client, "chat123")

	// Set via option
	stream2 := NewLiveChatStream(client, "chat456", WithStreamAccessToken("token-from-option"))
	if stream2.getAccessToken() != "token-from-option" {
		t.Errorf("getAccessToken() = %q, want 'token-from-option'", stream2.getAccessToken())
	}

	// Set via method
	stream.SetAccessToken("new-token")
	if stream.getAccessToken() != "new-token" {
		t.Errorf("getAccessToken() = %q, want 'new-token'", stream.getAccessToken())
	}
}

func TestLiveChatStream_RetryDirective(t *testing.T) {
	var mu sync.Mutex
	retryUpdated := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}
		// Send retry directive
		_, _ = fmt.Fprintf(w, "retry: 5000\n\n")
		flusher.Flush()
		<-r.Context().Done()
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	stream := NewLiveChatStream(client, "chat123",
		WithStreamAccessToken("test-token"),
		WithStreamBaseURL(server.URL),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = stream.Start(ctx)

	time.Sleep(100 * time.Millisecond)

	stream.mu.RLock()
	if stream.reconnectDelay == 5*time.Second {
		mu.Lock()
		retryUpdated = true
		mu.Unlock()
	}
	stream.mu.RUnlock()

	stream.Stop()

	mu.Lock()
	updated := retryUpdated
	mu.Unlock()

	if !updated {
		t.Error("retry directive was not processed")
	}
}

func TestParseUint(t *testing.T) {
	tests := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{"123", 123, false},
		{"0", 0, false},
		{"5000", 5000, false},
		{"abc", 0, true},
		{"12a3", 0, true},
		{"-5", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseUint(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseUint(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseUint(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestWithStreamHTTPClient(t *testing.T) {
	client := core.NewClient()

	t.Run("custom HTTP client", func(t *testing.T) {
		customHTTP := &http.Client{Timeout: 30 * time.Second}
		stream := NewLiveChatStream(client, "chat123",
			WithStreamHTTPClient(customHTTP),
		)
		if stream.httpClient != customHTTP {
			t.Error("custom HTTP client was not set")
		}
	})

	t.Run("nil HTTP client keeps default", func(t *testing.T) {
		stream := NewLiveChatStream(client, "chat123",
			WithStreamHTTPClient(nil),
		)
		if stream.httpClient == nil {
			t.Error("HTTP client should not be nil")
		}
	})
}

func TestLiveChatStream_QuotaMethods(t *testing.T) {
	t.Run("quotaUsed without tracker", func(t *testing.T) {
		client := core.NewClient()
		stream := NewLiveChatStream(client, "chat123")

		used := stream.quotaUsed()
		if used != 0 {
			t.Errorf("quotaUsed() = %d, want 0", used)
		}
	})

	t.Run("quotaLimit without tracker", func(t *testing.T) {
		client := core.NewClient()
		stream := NewLiveChatStream(client, "chat123")

		limit := stream.quotaLimit()
		if limit != core.DefaultDailyQuota {
			t.Errorf("quotaLimit() = %d, want %d", limit, core.DefaultDailyQuota)
		}
	})

	t.Run("with quota tracker", func(t *testing.T) {
		qt := core.NewQuotaTracker(core.DefaultDailyQuota)
		client := core.NewClient(core.WithQuotaTracker(qt))
		stream := NewLiveChatStream(client, "chat123")

		// Initial values
		if stream.quotaUsed() != 0 {
			t.Errorf("quotaUsed() = %d, want 0", stream.quotaUsed())
		}
		if stream.quotaLimit() != core.DefaultDailyQuota {
			t.Errorf("quotaLimit() = %d, want %d", stream.quotaLimit(), core.DefaultDailyQuota)
		}
	})
}

func TestLiveChatStream_HandlerUnsubscribe(t *testing.T) {
	client := core.NewClient()

	t.Run("OnDelete unsubscribe", func(t *testing.T) {
		stream := NewLiveChatStream(client, "chat123")

		unsub := stream.OnDelete(func(id string) {})

		stream.handlerMu.RLock()
		count := len(stream.deleteHandlers)
		stream.handlerMu.RUnlock()
		if count != 1 {
			t.Errorf("delete handler count = %d, want 1", count)
		}

		unsub()
		unsub() // Double unsubscribe should be safe

		stream.handlerMu.RLock()
		count = len(stream.deleteHandlers)
		stream.handlerMu.RUnlock()
		if count != 0 {
			t.Errorf("delete handler count after unsub = %d, want 0", count)
		}
	})

	t.Run("OnBan unsubscribe", func(t *testing.T) {
		stream := NewLiveChatStream(client, "chat123")

		unsub := stream.OnBan(func(details *UserBannedDetails) {})

		stream.handlerMu.RLock()
		count := len(stream.banHandlers)
		stream.handlerMu.RUnlock()
		if count != 1 {
			t.Errorf("ban handler count = %d, want 1", count)
		}

		unsub()

		stream.handlerMu.RLock()
		count = len(stream.banHandlers)
		stream.handlerMu.RUnlock()
		if count != 0 {
			t.Errorf("ban handler count after unsub = %d, want 0", count)
		}
	})

	t.Run("OnError unsubscribe", func(t *testing.T) {
		stream := NewLiveChatStream(client, "chat123")

		unsub := stream.OnError(func(err error) {})

		stream.handlerMu.RLock()
		count := len(stream.errorHandlers)
		stream.handlerMu.RUnlock()
		if count != 1 {
			t.Errorf("error handler count = %d, want 1", count)
		}

		unsub()

		stream.handlerMu.RLock()
		count = len(stream.errorHandlers)
		stream.handlerMu.RUnlock()
		if count != 0 {
			t.Errorf("error handler count after unsub = %d, want 0", count)
		}
	})

	t.Run("OnConnect unsubscribe", func(t *testing.T) {
		stream := NewLiveChatStream(client, "chat123")

		unsub := stream.OnConnect(func() {})

		stream.handlerMu.RLock()
		count := len(stream.connectHandlers)
		stream.handlerMu.RUnlock()
		if count != 1 {
			t.Errorf("connect handler count = %d, want 1", count)
		}

		unsub()

		stream.handlerMu.RLock()
		count = len(stream.connectHandlers)
		stream.handlerMu.RUnlock()
		if count != 0 {
			t.Errorf("connect handler count after unsub = %d, want 0", count)
		}
	})

	t.Run("OnDisconnect unsubscribe", func(t *testing.T) {
		stream := NewLiveChatStream(client, "chat123")

		unsub := stream.OnDisconnect(func() {})

		stream.handlerMu.RLock()
		count := len(stream.disconnectHandlers)
		stream.handlerMu.RUnlock()
		if count != 1 {
			t.Errorf("disconnect handler count = %d, want 1", count)
		}

		unsub()

		stream.handlerMu.RLock()
		count = len(stream.disconnectHandlers)
		stream.handlerMu.RUnlock()
		if count != 0 {
			t.Errorf("disconnect handler count after unsub = %d, want 0", count)
		}
	})

	t.Run("OnResponse unsubscribe", func(t *testing.T) {
		stream := NewLiveChatStream(client, "chat123")

		unsub := stream.OnResponse(func(resp *LiveChatMessageListResponse) {})

		stream.handlerMu.RLock()
		count := len(stream.responseHandlers)
		stream.handlerMu.RUnlock()
		if count != 1 {
			t.Errorf("response handler count = %d, want 1", count)
		}

		unsub()

		stream.handlerMu.RLock()
		count = len(stream.responseHandlers)
		stream.handlerMu.RUnlock()
		if count != 0 {
			t.Errorf("response handler count after unsub = %d, want 0", count)
		}
	})
}

func TestLiveChatStream_ResetClearsState(t *testing.T) {
	client := core.NewClient()
	stream := NewLiveChatStream(client, "chat123")

	// Set some state
	stream.SetPageToken("token123")
	stream.SetAccessToken("access123")

	// Reset
	_ = stream.Reset()

	if stream.PageToken() != "" {
		t.Errorf("PageToken() after reset = %q, want empty", stream.PageToken())
	}
}
