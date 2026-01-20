package streaming

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Its-donkey/yougopher/youtube/core"
)

func TestNewLiveChatPoller(t *testing.T) {
	client := core.NewClient()
	poller := NewLiveChatPoller(client, "chat123")

	if poller.LiveChatID() != "chat123" {
		t.Errorf("LiveChatID() = %q, want 'chat123'", poller.LiveChatID())
	}
	if poller.minPollInterval != DefaultMinPollInterval {
		t.Errorf("minPollInterval = %v, want %v", poller.minPollInterval, DefaultMinPollInterval)
	}
	if poller.maxPollInterval != DefaultMaxPollInterval {
		t.Errorf("maxPollInterval = %v, want %v", poller.maxPollInterval, DefaultMaxPollInterval)
	}
}

func TestNewLiveChatPoller_WithOptions(t *testing.T) {
	client := core.NewClient()
	backoff := core.NewBackoffConfig(core.WithBaseDelay(2 * time.Second))

	poller := NewLiveChatPoller(client, "chat123",
		WithMinPollInterval(500*time.Millisecond),
		WithMaxPollInterval(10*time.Second),
		WithBackoff(backoff),
		WithProfileImageSize("high"),
	)

	if poller.minPollInterval != 500*time.Millisecond {
		t.Errorf("minPollInterval = %v, want 500ms", poller.minPollInterval)
	}
	if poller.maxPollInterval != 10*time.Second {
		t.Errorf("maxPollInterval = %v, want 10s", poller.maxPollInterval)
	}
	if poller.profileImageSize != "high" {
		t.Errorf("profileImageSize = %q, want 'high'", poller.profileImageSize)
	}
}

func TestLiveChatPoller_StartStop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			NextPageToken:         "token1",
			PollingIntervalMillis: 1000,
			Items:                 []*LiveChatMessage{},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123",
		WithMinPollInterval(10*time.Millisecond),
	)

	ctx := context.Background()

	// Should not be running initially
	if poller.IsRunning() {
		t.Error("IsRunning() = true before Start()")
	}

	// Start polling
	err := poller.Start(ctx)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Should be running now
	if !poller.IsRunning() {
		t.Error("IsRunning() = false after Start()")
	}

	// Starting again should fail
	err = poller.Start(ctx)
	if err != ErrAlreadyRunning {
		t.Errorf("Start() again error = %v, want ErrAlreadyRunning", err)
	}

	// Stop polling
	poller.Stop()

	// Should not be running after stop
	if poller.IsRunning() {
		t.Error("IsRunning() = true after Stop()")
	}

	// Stop again should be safe (idempotent)
	poller.Stop()

	// Should be able to start again
	err = poller.Start(ctx)
	if err != nil {
		t.Fatalf("Start() after Stop() error = %v", err)
	}

	poller.Stop()
}

func TestLiveChatPoller_OnMessage(t *testing.T) {
	var receivedMessages []*LiveChatMessage
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			NextPageToken:         "token1",
			PollingIntervalMillis: 100,
			Items: []*LiveChatMessage{
				{
					ID: "msg1",
					Snippet: &MessageSnippet{
						Type:           MessageTypeText,
						DisplayMessage: "Hello",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123",
		WithMinPollInterval(10*time.Millisecond),
	)

	unsub := poller.OnMessage(func(msg *LiveChatMessage) {
		mu.Lock()
		receivedMessages = append(receivedMessages, msg)
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = poller.Start(ctx)

	// Wait for some messages
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	count := len(receivedMessages)
	mu.Unlock()

	if count == 0 {
		t.Error("No messages received")
	}

	// Unsubscribe
	unsub()

	// Unsubscribe again should be safe
	unsub()

	poller.Stop()
}

func TestLiveChatPoller_OnConnect_OnDisconnect(t *testing.T) {
	var connectCalled atomic.Bool
	var disconnectCalled atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 1000,
			Items:                 []*LiveChatMessage{},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123",
		WithMinPollInterval(10*time.Millisecond),
	)

	poller.OnConnect(func() {
		connectCalled.Store(true)
	})

	poller.OnDisconnect(func() {
		disconnectCalled.Store(true)
	})

	ctx := context.Background()
	_ = poller.Start(ctx)

	// Wait for connect handler
	time.Sleep(50 * time.Millisecond)

	if !connectCalled.Load() {
		t.Error("OnConnect handler not called")
	}

	poller.Stop()

	// Wait for disconnect handler
	time.Sleep(50 * time.Millisecond)

	if !disconnectCalled.Load() {
		t.Error("OnDisconnect handler not called")
	}
}

func TestLiveChatPoller_OnError(t *testing.T) {
	var errorReceived atomic.Bool

	// Server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": {"message": "test error"}}`))
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123",
		WithMinPollInterval(10*time.Millisecond),
		WithBackoff(core.NewBackoffConfig(core.WithBaseDelay(10*time.Millisecond))),
	)

	poller.OnError(func(err error) {
		errorReceived.Store(true)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = poller.Start(ctx)

	// Wait for error
	time.Sleep(50 * time.Millisecond)

	if !errorReceived.Load() {
		t.Error("OnError handler not called")
	}

	poller.Stop()
}

func TestLiveChatPoller_OnPollComplete(t *testing.T) {
	var pollCount atomic.Int32
	var lastMessageCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 100,
			Items: []*LiveChatMessage{
				{ID: "msg1", Snippet: &MessageSnippet{Type: MessageTypeText}},
				{ID: "msg2", Snippet: &MessageSnippet{Type: MessageTypeText}},
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123",
		WithMinPollInterval(10*time.Millisecond),
	)

	poller.OnPollComplete(func(count int, interval time.Duration) {
		pollCount.Add(1)
		lastMessageCount.Store(int32(count))
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = poller.Start(ctx)

	time.Sleep(50 * time.Millisecond)

	if pollCount.Load() == 0 {
		t.Error("OnPollComplete handler not called")
	}
	if lastMessageCount.Load() != 2 {
		t.Errorf("lastMessageCount = %d, want 2", lastMessageCount.Load())
	}

	poller.Stop()
}

func TestLiveChatPoller_OnDelete(t *testing.T) {
	var deletedID string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 1000,
			Items: []*LiveChatMessage{
				{
					ID: "delete-event",
					Snippet: &MessageSnippet{
						Type: MessageTypeMessageDeleted,
						MessageDeletedDetails: &MessageDeletedDetails{
							DeletedMessageID: "deleted123",
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123",
		WithMinPollInterval(10*time.Millisecond),
	)

	poller.OnDelete(func(id string) {
		mu.Lock()
		deletedID = id
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = poller.Start(ctx)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	got := deletedID
	mu.Unlock()

	if got != "deleted123" {
		t.Errorf("deletedID = %q, want 'deleted123'", got)
	}

	poller.Stop()
}

func TestLiveChatPoller_OnBan(t *testing.T) {
	var bannedChannelID string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 1000,
			Items: []*LiveChatMessage{
				{
					ID: "ban-event",
					Snippet: &MessageSnippet{
						Type: MessageTypeUserBanned,
						UserBannedDetails: &UserBannedDetails{
							BanType: BanTypePermanent,
							BannedUserDetails: &BannedUserDetails{
								ChannelID: "banned123",
							},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123",
		WithMinPollInterval(10*time.Millisecond),
	)

	poller.OnBan(func(details *UserBannedDetails) {
		mu.Lock()
		bannedChannelID = details.BannedUserDetails.ChannelID
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = poller.Start(ctx)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	got := bannedChannelID
	mu.Unlock()

	if got != "banned123" {
		t.Errorf("bannedChannelID = %q, want 'banned123'", got)
	}

	poller.Stop()
}

func TestLiveChatPoller_HandlerUnsubscribe(t *testing.T) {
	client := core.NewClient()
	poller := NewLiveChatPoller(client, "chat123")

	// Register multiple handlers
	unsub1 := poller.OnMessage(func(msg *LiveChatMessage) {})
	unsub2 := poller.OnMessage(func(msg *LiveChatMessage) {})

	poller.handlerMu.RLock()
	count := len(poller.messageHandlers)
	poller.handlerMu.RUnlock()

	if count != 2 {
		t.Errorf("handler count = %d, want 2", count)
	}

	// Unsubscribe first
	unsub1()

	poller.handlerMu.RLock()
	count = len(poller.messageHandlers)
	poller.handlerMu.RUnlock()

	if count != 1 {
		t.Errorf("handler count after unsub = %d, want 1", count)
	}

	// Unsubscribe again (idempotent)
	unsub1()

	poller.handlerMu.RLock()
	count = len(poller.messageHandlers)
	poller.handlerMu.RUnlock()

	if count != 1 {
		t.Errorf("handler count after double unsub = %d, want 1", count)
	}

	// Unsubscribe second
	unsub2()

	poller.handlerMu.RLock()
	count = len(poller.messageHandlers)
	poller.handlerMu.RUnlock()

	if count != 0 {
		t.Errorf("handler count after all unsub = %d, want 0", count)
	}
}

func TestLiveChatPoller_PanicRecovery(t *testing.T) {
	var errorReceived atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 1000,
			Items: []*LiveChatMessage{
				{ID: "msg1", Snippet: &MessageSnippet{Type: MessageTypeText}},
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123",
		WithMinPollInterval(10*time.Millisecond),
	)

	// Handler that panics
	poller.OnMessage(func(msg *LiveChatMessage) {
		panic("test panic")
	})

	// Error handler to catch the panic
	poller.OnError(func(err error) {
		errorReceived.Store(true)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = poller.Start(ctx)

	time.Sleep(50 * time.Millisecond)

	// Poller should still be running despite panic
	if !poller.IsRunning() {
		t.Error("Poller stopped after handler panic")
	}

	if !errorReceived.Load() {
		t.Error("Error handler not called for panic")
	}

	poller.Stop()
}

func TestLiveChatPoller_ContextCancellation(t *testing.T) {
	var disconnectCalled atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 5000,
			Items:                 []*LiveChatMessage{},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123",
		WithMinPollInterval(10*time.Millisecond),
	)

	poller.OnDisconnect(func() {
		disconnectCalled.Store(true)
	})

	ctx, cancel := context.WithCancel(context.Background())
	_ = poller.Start(ctx)

	time.Sleep(50 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for poller to stop
	time.Sleep(100 * time.Millisecond)

	if !disconnectCalled.Load() {
		t.Error("OnDisconnect not called after context cancellation")
	}
}

func TestLiveChatPoller_PageToken(t *testing.T) {
	client := core.NewClient()
	poller := NewLiveChatPoller(client, "chat123")

	// Initially empty
	if poller.PageToken() != "" {
		t.Errorf("PageToken() = %q, want empty", poller.PageToken())
	}

	// Set page token
	poller.SetPageToken("token123")

	if poller.PageToken() != "token123" {
		t.Errorf("PageToken() = %q, want 'token123'", poller.PageToken())
	}
}

func TestLiveChatPoller_PollInterval(t *testing.T) {
	client := core.NewClient()
	poller := NewLiveChatPoller(client, "chat123")

	// Initially zero
	if poller.PollInterval() != 0 {
		t.Errorf("PollInterval() = %v, want 0", poller.PollInterval())
	}
}

func TestLiveChatPoller_SendMessage(t *testing.T) {
	var receivedBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/liveChat/messages" && r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&receivedBody)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(LiveChatMessage{
				ID: "sent123",
				Snippet: &MessageSnippet{
					Type:           MessageTypeText,
					DisplayMessage: "Hello!",
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123")

	msg, err := poller.SendMessage(context.Background(), "Hello!")
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	if msg.ID != "sent123" {
		t.Errorf("msg.ID = %q, want 'sent123'", msg.ID)
	}

	// Verify request body
	snippet, ok := receivedBody["snippet"].(map[string]any)
	if !ok {
		t.Fatal("snippet not in request body")
	}
	if snippet["liveChatId"] != "chat123" {
		t.Errorf("liveChatId = %q, want 'chat123'", snippet["liveChatId"])
	}
}

func TestLiveChatPoller_DeleteMessage(t *testing.T) {
	var deletedID string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/liveChat/messages" && r.Method == http.MethodDelete {
			deletedID = r.URL.Query().Get("id")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123")

	err := poller.DeleteMessage(context.Background(), "msg123")
	if err != nil {
		t.Fatalf("DeleteMessage() error = %v", err)
	}

	if deletedID != "msg123" {
		t.Errorf("deletedID = %q, want 'msg123'", deletedID)
	}
}

func TestLiveChatPoller_BanUser(t *testing.T) {
	var receivedBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/liveChat/bans" && r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&receivedBody)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(LiveChatBan{
				ID: "ban123",
				Snippet: &BanSnippet{
					BanType: BanTypePermanent,
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123")

	ban, err := poller.BanUser(context.Background(), "baduser123")
	if err != nil {
		t.Fatalf("BanUser() error = %v", err)
	}

	if ban.ID != "ban123" {
		t.Errorf("ban.ID = %q, want 'ban123'", ban.ID)
	}

	snippet, ok := receivedBody["snippet"].(map[string]any)
	if !ok {
		t.Fatal("snippet not in request body")
	}
	if snippet["type"] != BanTypePermanent {
		t.Errorf("type = %q, want %q", snippet["type"], BanTypePermanent)
	}
}

func TestLiveChatPoller_TimeoutUser(t *testing.T) {
	var receivedBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/liveChat/bans" && r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&receivedBody)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(LiveChatBan{
				ID: "ban123",
				Snippet: &BanSnippet{
					BanType:            BanTypeTemporary,
					BanDurationSeconds: 300,
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123")

	ban, err := poller.TimeoutUser(context.Background(), "baduser123", 300)
	if err != nil {
		t.Fatalf("TimeoutUser() error = %v", err)
	}

	if ban.ID != "ban123" {
		t.Errorf("ban.ID = %q, want 'ban123'", ban.ID)
	}

	snippet, ok := receivedBody["snippet"].(map[string]any)
	if !ok {
		t.Fatal("snippet not in request body")
	}
	if snippet["type"] != BanTypeTemporary {
		t.Errorf("type = %q, want %q", snippet["type"], BanTypeTemporary)
	}
	if snippet["banDurationSeconds"] != float64(300) {
		t.Errorf("banDurationSeconds = %v, want 300", snippet["banDurationSeconds"])
	}
}

func TestLiveChatPoller_UnbanUser(t *testing.T) {
	var unbannedID string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/liveChat/bans" && r.Method == http.MethodDelete {
			unbannedID = r.URL.Query().Get("id")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123")

	err := poller.UnbanUser(context.Background(), "ban123")
	if err != nil {
		t.Fatalf("UnbanUser() error = %v", err)
	}

	if unbannedID != "ban123" {
		t.Errorf("unbannedID = %q, want 'ban123'", unbannedID)
	}
}

func TestLiveChatPoller_AddModerator(t *testing.T) {
	var receivedBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/liveChat/moderators" && r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&receivedBody)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(LiveChatModerator{
				ID: "mod123",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123")

	mod, err := poller.AddModerator(context.Background(), "user123")
	if err != nil {
		t.Fatalf("AddModerator() error = %v", err)
	}

	if mod.ID != "mod123" {
		t.Errorf("mod.ID = %q, want 'mod123'", mod.ID)
	}
}

func TestLiveChatPoller_RemoveModerator(t *testing.T) {
	var removedID string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/liveChat/moderators" && r.Method == http.MethodDelete {
			removedID = r.URL.Query().Get("id")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123")

	err := poller.RemoveModerator(context.Background(), "mod123")
	if err != nil {
		t.Fatalf("RemoveModerator() error = %v", err)
	}

	if removedID != "mod123" {
		t.Errorf("removedID = %q, want 'mod123'", removedID)
	}
}

func TestLiveChatPoller_ChatEnded(t *testing.T) {
	var disconnectCalled atomic.Bool
	var errorReceived atomic.Bool

	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			OfflineAt:             &now,
			PollingIntervalMillis: 1000,
			Items:                 []*LiveChatMessage{},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123",
		WithMinPollInterval(10*time.Millisecond),
	)

	poller.OnDisconnect(func() {
		disconnectCalled.Store(true)
	})

	poller.OnError(func(err error) {
		errorReceived.Store(true)
	})

	ctx := context.Background()
	_ = poller.Start(ctx)

	// Wait for chat ended to be detected
	time.Sleep(100 * time.Millisecond)

	if !disconnectCalled.Load() {
		t.Error("OnDisconnect not called when chat ended")
	}
	if !errorReceived.Load() {
		t.Error("OnError not called with ChatEndedError")
	}
}

func TestLiveChatPoller_PollIntervalBounds(t *testing.T) {
	tests := []struct {
		name        string
		apiInterval int
		minInterval time.Duration
		maxInterval time.Duration
		want        time.Duration
	}{
		{
			name:        "below minimum",
			apiInterval: 100,
			minInterval: 1 * time.Second,
			maxInterval: 30 * time.Second,
			want:        1 * time.Second,
		},
		{
			name:        "above maximum",
			apiInterval: 60000,
			minInterval: 1 * time.Second,
			maxInterval: 30 * time.Second,
			want:        30 * time.Second,
		},
		{
			name:        "within bounds",
			apiInterval: 5000,
			minInterval: 1 * time.Second,
			maxInterval: 30 * time.Second,
			want:        5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
					PollingIntervalMillis: tt.apiInterval,
					Items:                 []*LiveChatMessage{},
				})
			}))
			defer server.Close()

			client := core.NewClient(core.WithBaseURL(server.URL))
			poller := NewLiveChatPoller(client, "chat123",
				WithMinPollInterval(tt.minInterval),
				WithMaxPollInterval(tt.maxInterval),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			_ = poller.Start(ctx)

			time.Sleep(50 * time.Millisecond)

			got := poller.PollInterval()
			if got != tt.want {
				t.Errorf("PollInterval() = %v, want %v", got, tt.want)
			}

			poller.Stop()
		})
	}
}

func TestLiveChatPoller_AllHandlerTypes(t *testing.T) {
	client := core.NewClient()
	poller := NewLiveChatPoller(client, "chat123")

	// Register all handler types
	unsubs := []func(){
		poller.OnMessage(func(msg *LiveChatMessage) {}),
		poller.OnDelete(func(id string) {}),
		poller.OnBan(func(details *UserBannedDetails) {}),
		poller.OnError(func(err error) {}),
		poller.OnConnect(func() {}),
		poller.OnDisconnect(func() {}),
		poller.OnPollComplete(func(count int, interval time.Duration) {}),
	}

	// Verify all handlers registered
	poller.handlerMu.RLock()
	if len(poller.messageHandlers) != 1 {
		t.Error("messageHandlers not registered")
	}
	if len(poller.deleteHandlers) != 1 {
		t.Error("deleteHandlers not registered")
	}
	if len(poller.banHandlers) != 1 {
		t.Error("banHandlers not registered")
	}
	if len(poller.errorHandlers) != 1 {
		t.Error("errorHandlers not registered")
	}
	if len(poller.connectHandlers) != 1 {
		t.Error("connectHandlers not registered")
	}
	if len(poller.disconnectHandlers) != 1 {
		t.Error("disconnectHandlers not registered")
	}
	if len(poller.pollCompleteHandlers) != 1 {
		t.Error("pollCompleteHandlers not registered")
	}
	poller.handlerMu.RUnlock()

	// Unsubscribe all
	for _, unsub := range unsubs {
		unsub()
	}

	// Verify all handlers removed
	poller.handlerMu.RLock()
	if len(poller.messageHandlers) != 0 {
		t.Error("messageHandlers not unsubscribed")
	}
	if len(poller.deleteHandlers) != 0 {
		t.Error("deleteHandlers not unsubscribed")
	}
	if len(poller.banHandlers) != 0 {
		t.Error("banHandlers not unsubscribed")
	}
	if len(poller.errorHandlers) != 0 {
		t.Error("errorHandlers not unsubscribed")
	}
	if len(poller.connectHandlers) != 0 {
		t.Error("connectHandlers not unsubscribed")
	}
	if len(poller.disconnectHandlers) != 0 {
		t.Error("disconnectHandlers not unsubscribed")
	}
	if len(poller.pollCompleteHandlers) != 0 {
		t.Error("pollCompleteHandlers not unsubscribed")
	}
	poller.handlerMu.RUnlock()
}
