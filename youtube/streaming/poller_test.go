package streaming

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func TestLiveChatPoller_ListModerators(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/liveChat/moderators" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("liveChatId") != "chat123" {
				t.Errorf("unexpected liveChatId: %s", r.URL.Query().Get("liveChatId"))
			}

			resp := LiveChatModeratorListResponse{
				Items: []*LiveChatModerator{
					{
						ID: "mod123",
						Snippet: &ModeratorSnippet{
							LiveChatID: "chat123",
							ModeratorDetails: &ModeratorDetails{
								ChannelID:   "channel456",
								DisplayName: "ModUser",
							},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		poller := NewLiveChatPoller(client, "chat123")

		resp, err := poller.ListModerators(context.Background(), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Items) != 1 {
			t.Fatalf("expected 1 moderator, got %d", len(resp.Items))
		}
		if resp.Items[0].Snippet.ModeratorDetails.DisplayName != "ModUser" {
			t.Errorf("unexpected display name: %s", resp.Items[0].Snippet.ModeratorDetails.DisplayName)
		}
	})

	t.Run("with params", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("maxResults") != "50" {
				t.Errorf("unexpected maxResults: %s", r.URL.Query().Get("maxResults"))
			}
			if r.URL.Query().Get("pageToken") != "token123" {
				t.Errorf("unexpected pageToken: %s", r.URL.Query().Get("pageToken"))
			}

			resp := LiveChatModeratorListResponse{Items: []*LiveChatModerator{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		poller := NewLiveChatPoller(client, "chat123")

		_, err := poller.ListModerators(context.Background(), &ListModeratorsParams{
			MaxResults: 50,
			PageToken:  "token123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestLiveChatPoller_TransitionChatMode(t *testing.T) {
	t.Run("success subscribers only", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/liveChat/messages/transition" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			var body TransitionChatModeRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if body.Snippet.LiveChatID != "chat123" {
				t.Errorf("unexpected liveChatId: %s", body.Snippet.LiveChatID)
			}
			if body.Snippet.Type != ChatModeSubscribersOnly {
				t.Errorf("unexpected type: %s", body.Snippet.Type)
			}

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		poller := NewLiveChatPoller(client, "chat123")

		err := poller.TransitionChatMode(context.Background(), ChatModeSubscribersOnly)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("slow mode with delay", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body TransitionChatModeRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if body.Snippet.Type != ChatModeSlowMode {
				t.Errorf("unexpected type: %s", body.Snippet.Type)
			}
			if body.Snippet.SlowModeDelayMs != 5000 {
				t.Errorf("unexpected slowModeDelayMs: %d", body.Snippet.SlowModeDelayMs)
			}

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		poller := NewLiveChatPoller(client, "chat123")

		err := poller.TransitionChatModeWithDelay(context.Background(), ChatModeSlowMode, 5000)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty mode", func(t *testing.T) {
		client := core.NewClient()
		poller := NewLiveChatPoller(client, "chat123")

		err := poller.TransitionChatMode(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty mode")
		}
	})
}

func TestListSuperChatEvents(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/superChatEvents" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			resp := SuperChatEventResourceListResponse{
				Items: []*SuperChatEventResource{
					{
						ID: "superchat123",
						Snippet: &SuperChatEventResourceSnippet{
							ChannelID:     "channel123",
							DisplayString: "$5.00",
							AmountMicros:  5000000,
							Currency:      "USD",
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))

		resp, err := ListSuperChatEvents(context.Background(), client, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Items) != 1 {
			t.Fatalf("expected 1 event, got %d", len(resp.Items))
		}
		if resp.Items[0].Snippet.DisplayString != "$5.00" {
			t.Errorf("unexpected display string: %s", resp.Items[0].Snippet.DisplayString)
		}
	})

	t.Run("with params", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("hl") != "en" {
				t.Errorf("unexpected hl: %s", r.URL.Query().Get("hl"))
			}
			if r.URL.Query().Get("maxResults") != "25" {
				t.Errorf("unexpected maxResults: %s", r.URL.Query().Get("maxResults"))
			}
			if r.URL.Query().Get("pageToken") != "token456" {
				t.Errorf("unexpected pageToken: %s", r.URL.Query().Get("pageToken"))
			}

			resp := SuperChatEventResourceListResponse{Items: []*SuperChatEventResource{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))

		_, err := ListSuperChatEvents(context.Background(), client, &ListSuperChatEventsParams{
			HL:         "en",
			MaxResults: 25,
			PageToken:  "token456",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestChatModeConstants(t *testing.T) {
	expectedModes := map[string]string{
		"subscribersOnlyMode": ChatModeSubscribersOnly,
		"membersOnlyMode":     ChatModeMembersOnly,
		"slowMode":            ChatModeSlowMode,
		"normal":              ChatModeNormal,
	}

	for expected, actual := range expectedModes {
		if actual != expected {
			t.Errorf("ChatMode constant %q = %q, want %q", expected, actual, expected)
		}
	}
}

func TestLiveChatPoller_ResetPageToken(t *testing.T) {
	client := core.NewClient()
	poller := NewLiveChatPoller(client, "chat123")

	// Set a page token
	poller.SetPageToken("token123")
	if poller.PageToken() != "token123" {
		t.Errorf("PageToken() = %q, want 'token123'", poller.PageToken())
	}

	// Reset it
	poller.ResetPageToken()

	if poller.PageToken() != "" {
		t.Errorf("PageToken() after reset = %q, want empty", poller.PageToken())
	}
}

func TestLiveChatPoller_Reset(t *testing.T) {
	t.Run("success when not running", func(t *testing.T) {
		client := core.NewClient()
		poller := NewLiveChatPoller(client, "chat123")

		// Set some state
		poller.SetPageToken("token123")
		poller.mu.Lock()
		poller.pollInterval = 5 * time.Second
		poller.mu.Unlock()

		// Reset
		err := poller.Reset()
		if err != nil {
			t.Fatalf("Reset() error = %v", err)
		}

		if poller.PageToken() != "" {
			t.Errorf("PageToken() after reset = %q, want empty", poller.PageToken())
		}
		if poller.PollInterval() != 0 {
			t.Errorf("PollInterval() after reset = %v, want 0", poller.PollInterval())
		}
	})

	t.Run("error when running", func(t *testing.T) {
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

		ctx := context.Background()
		_ = poller.Start(ctx)

		// Try to reset while running
		err := poller.Reset()
		if err != ErrAlreadyRunning {
			t.Errorf("Reset() while running error = %v, want ErrAlreadyRunning", err)
		}

		poller.Stop()
	})
}

func TestLiveChatPoller_SendMessage_Error(t *testing.T) {
	t.Run("empty message", func(t *testing.T) {
		client := core.NewClient()
		poller := NewLiveChatPoller(client, "chat123")

		_, err := poller.SendMessage(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty message")
		}
	})

	t.Run("API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error": {"message": "forbidden"}}`))
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		poller := NewLiveChatPoller(client, "chat123")

		_, err := poller.SendMessage(context.Background(), "test")
		if err == nil {
			t.Fatal("expected error for API error")
		}
	})
}

func TestLiveChatPoller_DeleteMessage_Error(t *testing.T) {
	t.Run("empty message ID", func(t *testing.T) {
		client := core.NewClient()
		poller := NewLiveChatPoller(client, "chat123")

		err := poller.DeleteMessage(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty message ID")
		}
	})
}

func TestLiveChatPoller_BanUser_Error(t *testing.T) {
	t.Run("empty channel ID", func(t *testing.T) {
		client := core.NewClient()
		poller := NewLiveChatPoller(client, "chat123")

		_, err := poller.BanUser(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty channel ID")
		}
	})
}

func TestLiveChatPoller_TimeoutUser_Error(t *testing.T) {
	t.Run("empty channel ID", func(t *testing.T) {
		client := core.NewClient()
		poller := NewLiveChatPoller(client, "chat123")

		_, err := poller.TimeoutUser(context.Background(), "", 300)
		if err == nil {
			t.Fatal("expected error for empty channel ID")
		}
	})

	t.Run("invalid duration", func(t *testing.T) {
		client := core.NewClient()
		poller := NewLiveChatPoller(client, "chat123")

		_, err := poller.TimeoutUser(context.Background(), "user123", 0)
		if err == nil {
			t.Fatal("expected error for zero duration")
		}
	})
}

func TestLiveChatPoller_UnbanUser_Error(t *testing.T) {
	t.Run("empty ban ID", func(t *testing.T) {
		client := core.NewClient()
		poller := NewLiveChatPoller(client, "chat123")

		err := poller.UnbanUser(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty ban ID")
		}
	})
}

func TestLiveChatPoller_AddModerator_Error(t *testing.T) {
	t.Run("empty channel ID", func(t *testing.T) {
		client := core.NewClient()
		poller := NewLiveChatPoller(client, "chat123")

		_, err := poller.AddModerator(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty channel ID")
		}
	})
}

func TestLiveChatPoller_RemoveModerator_Error(t *testing.T) {
	t.Run("empty moderator ID", func(t *testing.T) {
		client := core.NewClient()
		poller := NewLiveChatPoller(client, "chat123")

		err := poller.RemoveModerator(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty moderator ID")
		}
	})
}

func TestWithProfileImageSize(t *testing.T) {
	client := core.NewClient()

	t.Run("valid size", func(t *testing.T) {
		poller := NewLiveChatPoller(client, "chat123",
			WithProfileImageSize("medium"),
		)
		if poller.profileImageSize != "medium" {
			t.Errorf("profileImageSize = %q, want 'medium'", poller.profileImageSize)
		}
	})

	t.Run("empty size keeps default", func(t *testing.T) {
		poller := NewLiveChatPoller(client, "chat123",
			WithProfileImageSize(""),
		)
		// Empty string option keeps the default value
		if poller.profileImageSize != "default" {
			t.Errorf("profileImageSize = %q, want 'default'", poller.profileImageSize)
		}
	})
}

// =============================================================================
// Boundary Condition Tests
// =============================================================================

func TestLiveChatPoller_TimeoutUser_BoundaryValues(t *testing.T) {
	tests := []struct {
		name      string
		duration  int64
		wantError bool
	}{
		{"zero duration", 0, true},
		{"negative duration", -1, true},
		{"minimum valid (1 second)", 1, false},
		{"typical value", 300, false},
		{"large value", 86400, false}, // 24 hours
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestReceived bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestReceived = true
				w.Header().Set("Content-Type", "application/json")
				// banDurationSeconds needs to be a string in JSON due to ,string tag
				response := fmt.Sprintf(`{
					"id": "ban123",
					"snippet": {
						"liveChatId": "chat123",
						"type": "temporary",
						"banDurationSeconds": "%d"
					}
				}`, tt.duration)
				_, _ = w.Write([]byte(response))
			}))
			defer server.Close()

			client := core.NewClient(core.WithBaseURL(server.URL))
			poller := NewLiveChatPoller(client, "chat123")

			_, err := poller.TimeoutUser(context.Background(), "user123", tt.duration)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if requestReceived {
					t.Error("API request should not be made for invalid input")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestLiveChatPoller_PollIntervalBoundaries(t *testing.T) {
	tests := []struct {
		name             string
		apiIntervalMs    int
		minInterval      time.Duration
		maxInterval      time.Duration
		expectedApproxMs int
	}{
		{
			name:             "API returns below minimum",
			apiIntervalMs:    100,
			minInterval:      1 * time.Second,
			maxInterval:      30 * time.Second,
			expectedApproxMs: 1000, // clamped to min
		},
		{
			name:             "API returns above maximum",
			apiIntervalMs:    60000,
			minInterval:      1 * time.Second,
			maxInterval:      30 * time.Second,
			expectedApproxMs: 30000, // clamped to max
		},
		{
			name:             "API returns within bounds",
			apiIntervalMs:    5000,
			minInterval:      1 * time.Second,
			maxInterval:      30 * time.Second,
			expectedApproxMs: 5000, // used as-is
		},
		{
			name:             "API returns exactly at minimum",
			apiIntervalMs:    1000,
			minInterval:      1 * time.Second,
			maxInterval:      30 * time.Second,
			expectedApproxMs: 1000,
		},
		{
			name:             "API returns exactly at maximum",
			apiIntervalMs:    30000,
			minInterval:      1 * time.Second,
			maxInterval:      30 * time.Second,
			expectedApproxMs: 30000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
					PollingIntervalMillis: tt.apiIntervalMs,
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

			interval := poller.PollInterval()
			expectedMin := time.Duration(tt.expectedApproxMs-100) * time.Millisecond
			expectedMax := time.Duration(tt.expectedApproxMs+100) * time.Millisecond

			if interval < expectedMin || interval > expectedMax {
				t.Errorf("PollInterval() = %v, want approximately %d ms", interval, tt.expectedApproxMs)
			}

			poller.Stop()
		})
	}
}

// =============================================================================
// Error Type Verification Tests (errors.Is / errors.As)
// =============================================================================

func TestLiveChatPoller_ErrorIs_ErrAlreadyRunning(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 1000,
			Items:                 []*LiveChatMessage{},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123")

	ctx := context.Background()
	_ = poller.Start(ctx)
	defer poller.Stop()

	err := poller.Start(ctx)

	// Verify using errors.Is
	if !errors.Is(err, ErrAlreadyRunning) {
		t.Errorf("errors.Is(err, ErrAlreadyRunning) = false, want true; err = %v", err)
	}

	// Also verify it matches core.ErrAlreadyRunning (same error)
	if !errors.Is(err, core.ErrAlreadyRunning) {
		t.Errorf("errors.Is(err, core.ErrAlreadyRunning) = false, want true; err = %v", err)
	}
}

func TestLiveChatPoller_ErrorIs_ErrNotRunning(t *testing.T) {
	client := core.NewClient()
	poller := NewLiveChatPoller(client, "chat123")

	err := poller.Reset()

	// Reset should succeed when not running
	if err != nil {
		t.Errorf("Reset() when not running should succeed, got err = %v", err)
	}

	// Start the poller
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 1000,
			Items:                 []*LiveChatMessage{},
		})
	}))
	defer server.Close()

	client = core.NewClient(core.WithBaseURL(server.URL))
	poller = NewLiveChatPoller(client, "chat123")

	ctx := context.Background()
	_ = poller.Start(ctx)

	err = poller.Reset()

	if !errors.Is(err, ErrAlreadyRunning) {
		t.Errorf("errors.Is(err, ErrAlreadyRunning) = false for Reset() while running; err = %v", err)
	}

	poller.Stop()
}

// =============================================================================
// Deadlock Detection Tests
// =============================================================================

func TestLiveChatPoller_ConcurrentStartStop_NoDeadlock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 100,
			Items:                 []*LiveChatMessage{},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123",
		WithMinPollInterval(10*time.Millisecond),
	)

	done := make(chan struct{})
	go func() {
		defer close(done)
		// Test rapid start/stop cycles sequentially to avoid WaitGroup reuse
		for i := 0; i < 20; i++ {
			_ = poller.Start(context.Background())
			time.Sleep(5 * time.Millisecond) // Allow state to settle
			poller.Stop()
		}
	}()

	select {
	case <-done:
		// Success - no deadlock
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock detected: rapid Start/Stop cycles did not complete within 5 seconds")
	}
}

func TestLiveChatPoller_ConcurrentStopCalls_NoDeadlock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 100,
			Items:                 []*LiveChatMessage{},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123",
		WithMinPollInterval(10*time.Millisecond),
	)

	_ = poller.Start(context.Background())
	time.Sleep(20 * time.Millisecond) // Ensure it's running

	done := make(chan struct{})
	go func() {
		defer close(done)
		var wg sync.WaitGroup
		// Multiple concurrent Stop calls should be safe (idempotent)
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				poller.Stop()
			}()
		}
		wg.Wait()
	}()

	select {
	case <-done:
		// Success - no deadlock
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock detected: concurrent Stop calls did not complete within 5 seconds")
	}
}

func TestLiveChatPoller_ConcurrentHandlerRegistration_NoDeadlock(t *testing.T) {
	client := core.NewClient()
	poller := NewLiveChatPoller(client, "chat123")

	done := make(chan struct{})
	go func() {
		defer close(done)
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(6)
			go func() {
				defer wg.Done()
				unsub := poller.OnMessage(func(m *LiveChatMessage) {})
				unsub()
			}()
			go func() {
				defer wg.Done()
				unsub := poller.OnDelete(func(id string) {})
				unsub()
			}()
			go func() {
				defer wg.Done()
				unsub := poller.OnBan(func(b *UserBannedDetails) {})
				unsub()
			}()
			go func() {
				defer wg.Done()
				unsub := poller.OnError(func(e error) {})
				unsub()
			}()
			go func() {
				defer wg.Done()
				unsub := poller.OnConnect(func() {})
				unsub()
			}()
			go func() {
				defer wg.Done()
				unsub := poller.OnDisconnect(func() {})
				unsub()
			}()
		}
		wg.Wait()
	}()

	select {
	case <-done:
		// Success - no deadlock
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock detected: concurrent handler registration did not complete within 5 seconds")
	}
}

// =============================================================================
// Mock Call Count Verification Tests
// =============================================================================

func TestLiveChatPoller_SendMessage_VerifyAPICall(t *testing.T) {
	var (
		callCount   int
		lastMethod  string
		lastPath    string
		lastBody    map[string]interface{}
		lastHeaders http.Header
		mu          sync.Mutex
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		lastMethod = r.Method
		lastPath = r.URL.Path
		lastHeaders = r.Header.Clone()
		_ = json.NewDecoder(r.Body).Decode(&lastBody)
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "msg123",
			"snippet": map[string]interface{}{
				"liveChatId":     "chat123",
				"type":           "textMessageEvent",
				"displayMessage": "Hello World",
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123")

	_, err := poller.SendMessage(context.Background(), "Hello World")
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	// Verify call count
	if callCount != 1 {
		t.Errorf("API called %d times, want 1", callCount)
	}

	// Verify method
	if lastMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", lastMethod)
	}

	// Verify path
	if lastPath != "/liveChat/messages" {
		t.Errorf("path = %s, want /liveChat/messages", lastPath)
	}

	// Verify Content-Type header
	if ct := lastHeaders.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", ct)
	}

	// Verify request body structure
	snippet, ok := lastBody["snippet"].(map[string]interface{})
	if !ok {
		t.Fatal("request body missing snippet")
	}
	if snippet["liveChatId"] != "chat123" {
		t.Errorf("liveChatId = %v, want chat123", snippet["liveChatId"])
	}
	if snippet["textMessageDetails"].(map[string]interface{})["messageText"] != "Hello World" {
		t.Errorf("messageText = %v, want Hello World", snippet["textMessageDetails"])
	}
}

func TestLiveChatPoller_BanUser_VerifyAPICall(t *testing.T) {
	var (
		callCount int
		lastBody  map[string]interface{}
		mu        sync.Mutex
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		_ = json.NewDecoder(r.Body).Decode(&lastBody)
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "ban123",
			"snippet": map[string]interface{}{
				"liveChatId": "chat123",
				"type":       "permanent",
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123")

	_, err := poller.BanUser(context.Background(), "baduser123")
	if err != nil {
		t.Fatalf("BanUser() error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if callCount != 1 {
		t.Errorf("API called %d times, want 1", callCount)
	}

	snippet, ok := lastBody["snippet"].(map[string]interface{})
	if !ok {
		t.Fatal("request body missing snippet")
	}
	if snippet["liveChatId"] != "chat123" {
		t.Errorf("liveChatId = %v, want chat123", snippet["liveChatId"])
	}
	if snippet["type"] != "permanent" {
		t.Errorf("type = %v, want permanent", snippet["type"])
	}
	bannedUserDetails, ok := snippet["bannedUserDetails"].(map[string]interface{})
	if !ok {
		t.Fatal("request body missing bannedUserDetails")
	}
	if bannedUserDetails["channelId"] != "baduser123" {
		t.Errorf("channelId = %v, want baduser123", bannedUserDetails["channelId"])
	}
}

// =============================================================================
// String Length Limit Tests
// =============================================================================

func TestLiveChatPoller_SendMessage_StringLengthLimits(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		wantError bool
	}{
		{"empty message", "", true},
		{"single character", "a", false},
		{"typical message", "Hello, how are you?", false},
		{"200 characters", string(make([]byte, 200)), false},    // YouTube allows up to 200 chars
		{"very long message", string(make([]byte, 500)), false}, // API will reject but we send it
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestReceived bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestReceived = true
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"id": "msg123",
					"snippet": map[string]interface{}{
						"liveChatId":     "chat123",
						"displayMessage": tt.message,
					},
				})
			}))
			defer server.Close()

			client := core.NewClient(core.WithBaseURL(server.URL))
			poller := NewLiveChatPoller(client, "chat123")

			_, err := poller.SendMessage(context.Background(), tt.message)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if requestReceived {
					t.Error("API request should not be made for invalid input")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// =============================================================================
// Contention Tests (100+ goroutines)
// =============================================================================

func TestLiveChatPoller_HighContention_HandlerDispatch(t *testing.T) {
	var messageCount atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 50,
			Items: []*LiveChatMessage{
				{ID: "msg1", Snippet: &MessageSnippet{Type: MessageTypeText, DisplayMessage: "test"}},
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123",
		WithMinPollInterval(10*time.Millisecond),
	)

	// Register 100 concurrent handlers
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unsub := poller.OnMessage(func(m *LiveChatMessage) {
				messageCount.Add(1)
			})
			// Keep handler registered for the test duration
			time.Sleep(100 * time.Millisecond)
			unsub()
		}()
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()
		_ = poller.Start(ctx)
		<-ctx.Done()
		poller.Stop()
		wg.Wait()
	}()

	select {
	case <-done:
		// Success - no deadlock under contention
		if messageCount.Load() == 0 {
			t.Error("expected messages to be dispatched to handlers")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock detected under high contention")
	}
}

func TestLiveChatPoller_HighContention_ConcurrentOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/liveChat/messages":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
					PollingIntervalMillis: 100,
					Items:                 []*LiveChatMessage{},
				})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": "msg123"})
			}
		default:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": "result123"})
		}
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	poller := NewLiveChatPoller(client, "chat123",
		WithMinPollInterval(10*time.Millisecond),
	)

	_ = poller.Start(context.Background())
	defer poller.Stop()

	done := make(chan struct{})
	go func() {
		defer close(done)
		var wg sync.WaitGroup

		// 100 concurrent SendMessage calls
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = poller.SendMessage(context.Background(), "test message")
			}()
		}

		// Concurrent state reads
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = poller.IsRunning()
				_ = poller.PageToken()
				_ = poller.PollInterval()
			}()
		}

		wg.Wait()
	}()

	select {
	case <-done:
		// Success
	case <-time.After(10 * time.Second):
		t.Fatal("deadlock detected with 100+ concurrent operations")
	}
}

// =============================================================================
// Retry Logic Tests
// =============================================================================

func TestBackoffConfig_ExponentialDelays(t *testing.T) {
	// Use deterministic random for testing
	backoff := core.NewBackoffConfig(
		core.WithBaseDelay(100*time.Millisecond),
		core.WithMaxDelay(5*time.Second),
		core.WithMultiplier(2.0),
		core.WithJitter(0), // No jitter for deterministic testing
	)

	expectedDelays := []time.Duration{
		100 * time.Millisecond,  // attempt 0: 100ms * 2^0
		200 * time.Millisecond,  // attempt 1: 100ms * 2^1
		400 * time.Millisecond,  // attempt 2: 100ms * 2^2
		800 * time.Millisecond,  // attempt 3: 100ms * 2^3
		1600 * time.Millisecond, // attempt 4: 100ms * 2^4
		3200 * time.Millisecond, // attempt 5: 100ms * 2^5
		5 * time.Second,         // attempt 6: capped at max
		5 * time.Second,         // attempt 7: capped at max
	}

	for i, expected := range expectedDelays {
		actual := backoff.Delay(i)
		if actual != expected {
			t.Errorf("Delay(%d) = %v, want %v", i, actual, expected)
		}
	}
}

func TestBackoffConfig_JitterRange(t *testing.T) {
	backoff := core.NewBackoffConfig(
		core.WithBaseDelay(1*time.Second),
		core.WithJitter(0.2), // ±20%
	)

	// Run multiple times to test jitter distribution
	minDelay := time.Duration(1<<63 - 1)
	maxDelay := time.Duration(0)

	for i := 0; i < 100; i++ {
		delay := backoff.Delay(0)
		if delay < minDelay {
			minDelay = delay
		}
		if delay > maxDelay {
			maxDelay = delay
		}
	}

	// With 20% jitter, delays should be in range [800ms, 1200ms]
	expectedMin := 800 * time.Millisecond
	expectedMax := 1200 * time.Millisecond

	if minDelay < expectedMin-100*time.Millisecond {
		t.Errorf("min delay %v is too low (expected ~%v)", minDelay, expectedMin)
	}
	if maxDelay > expectedMax+100*time.Millisecond {
		t.Errorf("max delay %v is too high (expected ~%v)", maxDelay, expectedMax)
	}
}

// =============================================================================
// Benchmarks (Performance Tests)
// =============================================================================

func BenchmarkLiveChatPoller_HandlerRegistration(b *testing.B) {
	client := core.NewClient()
	poller := NewLiveChatPoller(client, "chat123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		unsub := poller.OnMessage(func(m *LiveChatMessage) {})
		unsub()
	}
}

func BenchmarkLiveChatMessage_Clone(b *testing.B) {
	msg := &LiveChatMessage{
		ID:   "msg123",
		Kind: "youtube#liveChatMessage",
		Snippet: &MessageSnippet{
			Type:           MessageTypeText,
			LiveChatID:     "chat123",
			DisplayMessage: "Hello, this is a test message!",
		},
		AuthorDetails: &AuthorDetails{
			ChannelID:       "channel123",
			DisplayName:     "Test User",
			ProfileImageURL: "https://example.com/image.jpg",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.Clone()
	}
}

func BenchmarkLiveChatPoller_StateReads(b *testing.B) {
	client := core.NewClient()
	poller := NewLiveChatPoller(client, "chat123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = poller.IsRunning()
		_ = poller.PageToken()
		_ = poller.PollInterval()
		_ = poller.LiveChatID()
	}
}
