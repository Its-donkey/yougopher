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

// mockTokenProvider implements TokenProvider for testing.
type mockTokenProvider struct {
	token string
	err   error
}

func (m *mockTokenProvider) AccessToken(ctx context.Context) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.token, nil
}

func TestNewChatBotClient(t *testing.T) {
	client := core.NewClient()
	bot, _ := NewChatBotClient(client, nil, "chat123")

	if bot.client != client {
		t.Error("client not set correctly")
	}
	if bot.liveChatID != "chat123" {
		t.Errorf("liveChatID = %q, want 'chat123'", bot.liveChatID)
	}
}

func TestNewChatBotClient_WithTokenProvider(t *testing.T) {
	client := core.NewClient()
	tp := &mockTokenProvider{token: "test-token"}
	bot, _ := NewChatBotClient(client, tp, "chat123")

	if bot.tokenProvider != tp {
		t.Error("tokenProvider not set correctly")
	}
}

func TestNewChatBotClient_WithPoller(t *testing.T) {
	client := core.NewClient()
	poller := NewLiveChatPoller(client, "chat123")
	bot, _ := NewChatBotClient(client, nil, "chat123", WithPoller(poller))

	if bot.poller != poller {
		t.Error("poller not set correctly")
	}
}

func TestNewChatBotClient_Validation(t *testing.T) {
	t.Run("nil client", func(t *testing.T) {
		_, err := NewChatBotClient(nil, nil, "chat123")
		if err == nil {
			t.Error("expected error for nil client")
		}
	})

	t.Run("empty liveChatID", func(t *testing.T) {
		client := core.NewClient()
		_, err := NewChatBotClient(client, nil, "")
		if err == nil {
			t.Error("expected error for empty liveChatID")
		}
	})
}

func TestChatBotClient_ConnectClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 1000,
			Items:                 []*LiveChatMessage{},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	tp := &mockTokenProvider{token: "test-token"}
	bot, _ := NewChatBotClient(client, tp, "chat123")

	ctx := context.Background()

	// Should not be connected initially
	if bot.IsConnected() {
		t.Error("IsConnected() = true before Connect()")
	}

	// Connect
	err := bot.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	// Should be connected now
	time.Sleep(50 * time.Millisecond)
	if !bot.IsConnected() {
		t.Error("IsConnected() = false after Connect()")
	}

	// Close
	err = bot.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Should not be connected after close
	if bot.IsConnected() {
		t.Error("IsConnected() = true after Close()")
	}
}

func TestChatBotClient_Connect_WithoutTokenProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 1000,
			Items:                 []*LiveChatMessage{},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	bot, _ := NewChatBotClient(client, nil, "chat123")

	ctx := context.Background()

	err := bot.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	_ = bot.Close()
}

func TestChatBotClient_OnMessage(t *testing.T) {
	var receivedMessages []*ChatMessage
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 100,
			Items: []*LiveChatMessage{
				{
					ID: "msg1",
					Snippet: &MessageSnippet{
						Type:           MessageTypeText,
						DisplayMessage: "Hello world",
						PublishedAt:    time.Now(),
					},
					AuthorDetails: &AuthorDetails{
						ChannelID:   "channel1",
						DisplayName: "Test User",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	bot, _ := NewChatBotClient(client, nil, "chat123")

	bot.OnMessage(func(msg *ChatMessage) {
		mu.Lock()
		receivedMessages = append(receivedMessages, msg)
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = bot.Connect(ctx)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	count := len(receivedMessages)
	mu.Unlock()

	if count == 0 {
		t.Error("No messages received")
	}

	mu.Lock()
	if count > 0 && receivedMessages[0].Message != "Hello world" {
		t.Errorf("Message = %q, want 'Hello world'", receivedMessages[0].Message)
	}
	if count > 0 && receivedMessages[0].Author.DisplayName != "Test User" {
		t.Errorf("Author.DisplayName = %q, want 'Test User'", receivedMessages[0].Author.DisplayName)
	}
	mu.Unlock()

	_ = bot.Close()
}

func TestChatBotClient_OnSuperChat(t *testing.T) {
	var receivedEvents []*SuperChatEvent
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 100,
			Items: []*LiveChatMessage{
				{
					ID: "sc1",
					Snippet: &MessageSnippet{
						Type: MessageTypeSuperChat,
						SuperChatDetails: &SuperChatDetails{
							AmountMicros:        5000000,
							Currency:            "USD",
							AmountDisplayString: "$5.00",
							UserComment:         "Great stream!",
							Tier:                3,
						},
					},
					AuthorDetails: &AuthorDetails{
						ChannelID:   "channel1",
						DisplayName: "Donor",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	bot, _ := NewChatBotClient(client, nil, "chat123")

	bot.OnSuperChat(func(event *SuperChatEvent) {
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = bot.Connect(ctx)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	count := len(receivedEvents)
	mu.Unlock()

	if count == 0 {
		t.Error("No SuperChat events received")
	}

	mu.Lock()
	if count > 0 {
		event := receivedEvents[0]
		if event.Amount != "$5.00" {
			t.Errorf("Amount = %q, want '$5.00'", event.Amount)
		}
		if event.Message != "Great stream!" {
			t.Errorf("Message = %q, want 'Great stream!'", event.Message)
		}
		if event.Tier != 3 {
			t.Errorf("Tier = %d, want 3", event.Tier)
		}
	}
	mu.Unlock()

	_ = bot.Close()
}

func TestChatBotClient_OnSuperSticker(t *testing.T) {
	var receivedEvents []*SuperStickerEvent
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 100,
			Items: []*LiveChatMessage{
				{
					ID: "ss1",
					Snippet: &MessageSnippet{
						Type: MessageTypeSuperSticker,
						SuperStickerDetails: &SuperStickerDetails{
							SuperStickerID:      "sticker1",
							AmountMicros:        2000000,
							Currency:            "USD",
							AmountDisplayString: "$2.00",
							Tier:                1,
							SuperStickerMetadata: &SuperStickerMetadata{
								AltText: "Heart sticker",
							},
						},
					},
					AuthorDetails: &AuthorDetails{
						ChannelID:   "channel1",
						DisplayName: "Sticker Fan",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	bot, _ := NewChatBotClient(client, nil, "chat123")

	bot.OnSuperSticker(func(event *SuperStickerEvent) {
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = bot.Connect(ctx)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	count := len(receivedEvents)
	mu.Unlock()

	if count == 0 {
		t.Error("No SuperSticker events received")
	}

	mu.Lock()
	if count > 0 {
		event := receivedEvents[0]
		if event.StickerID != "sticker1" {
			t.Errorf("StickerID = %q, want 'sticker1'", event.StickerID)
		}
		if event.AltText != "Heart sticker" {
			t.Errorf("AltText = %q, want 'Heart sticker'", event.AltText)
		}
	}
	mu.Unlock()

	_ = bot.Close()
}

func TestChatBotClient_OnMembership(t *testing.T) {
	var receivedEvents []*MembershipEvent
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 100,
			Items: []*LiveChatMessage{
				{
					ID: "member1",
					Snippet: &MessageSnippet{
						Type: MessageTypeMembership,
						NewSponsorDetails: &NewSponsorDetails{
							MemberLevelName: "Gold Member",
							IsUpgrade:       false,
						},
					},
					AuthorDetails: &AuthorDetails{
						ChannelID:   "channel1",
						DisplayName: "New Member",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	bot, _ := NewChatBotClient(client, nil, "chat123")

	bot.OnMembership(func(event *MembershipEvent) {
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = bot.Connect(ctx)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	count := len(receivedEvents)
	mu.Unlock()

	if count == 0 {
		t.Error("No Membership events received")
	}

	mu.Lock()
	if count > 0 {
		event := receivedEvents[0]
		if event.LevelName != "Gold Member" {
			t.Errorf("LevelName = %q, want 'Gold Member'", event.LevelName)
		}
	}
	mu.Unlock()

	_ = bot.Close()
}

func TestChatBotClient_OnMemberMilestone(t *testing.T) {
	var receivedEvents []*MemberMilestoneEvent
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 100,
			Items: []*LiveChatMessage{
				{
					ID: "milestone1",
					Snippet: &MessageSnippet{
						Type: MessageTypeMemberMilestone,
						MemberMilestoneChatDetails: &MemberMilestoneChatDetails{
							MemberLevelName: "Gold Member",
							MemberMonth:     12,
							UserComment:     "One year!",
						},
					},
					AuthorDetails: &AuthorDetails{
						ChannelID:   "channel1",
						DisplayName: "Loyal Member",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	bot, _ := NewChatBotClient(client, nil, "chat123")

	bot.OnMemberMilestone(func(event *MemberMilestoneEvent) {
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = bot.Connect(ctx)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	count := len(receivedEvents)
	mu.Unlock()

	if count == 0 {
		t.Error("No MemberMilestone events received")
	}

	mu.Lock()
	if count > 0 {
		event := receivedEvents[0]
		if event.Months != 12 {
			t.Errorf("Months = %d, want 12", event.Months)
		}
		if event.Message != "One year!" {
			t.Errorf("Message = %q, want 'One year!'", event.Message)
		}
	}
	mu.Unlock()

	_ = bot.Close()
}

func TestChatBotClient_OnGiftMembership(t *testing.T) {
	var receivedEvents []*GiftMembershipEvent
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 100,
			Items: []*LiveChatMessage{
				{
					ID: "gift1",
					Snippet: &MessageSnippet{
						Type: MessageTypeMembershipGifting,
						MembershipGiftingDetails: &MembershipGiftingDetails{
							MemberLevelName:      "Gold Member",
							GiftMembershipsCount: 5,
						},
					},
					AuthorDetails: &AuthorDetails{
						ChannelID:   "channel1",
						DisplayName: "Generous User",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	bot, _ := NewChatBotClient(client, nil, "chat123")

	bot.OnGiftMembership(func(event *GiftMembershipEvent) {
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = bot.Connect(ctx)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	count := len(receivedEvents)
	mu.Unlock()

	if count == 0 {
		t.Error("No GiftMembership events received")
	}

	mu.Lock()
	if count > 0 {
		event := receivedEvents[0]
		if event.Count != 5 {
			t.Errorf("Count = %d, want 5", event.Count)
		}
		if event.LevelName != "Gold Member" {
			t.Errorf("LevelName = %q, want 'Gold Member'", event.LevelName)
		}
	}
	mu.Unlock()

	_ = bot.Close()
}

func TestChatBotClient_OnGiftMembershipReceived(t *testing.T) {
	var receivedEvents []*GiftMembershipReceivedEvent
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 100,
			Items: []*LiveChatMessage{
				{
					ID: "gift-received-1",
					Snippet: &MessageSnippet{
						Type: MessageTypeGiftMembershipReceived,
						GiftMembershipReceivedDetails: &GiftMembershipReceivedDetails{
							MemberLevelName:                      "Gold Member",
							GifterChannelID:                      "gifter123",
							AssociatedMembershipGiftingMessageID: "giftmsg1",
						},
					},
					AuthorDetails: &AuthorDetails{
						ChannelID:   "channel1",
						DisplayName: "Recipient User",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	bot, _ := NewChatBotClient(client, nil, "chat123")

	bot.OnGiftMembershipReceived(func(event *GiftMembershipReceivedEvent) {
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = bot.Connect(ctx)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	count := len(receivedEvents)
	mu.Unlock()

	if count == 0 {
		t.Error("No GiftMembershipReceived events received")
	}

	mu.Lock()
	if count > 0 {
		event := receivedEvents[0]
		if event.LevelName != "Gold Member" {
			t.Errorf("LevelName = %q, want 'Gold Member'", event.LevelName)
		}
		if event.GifterChannelID != "gifter123" {
			t.Errorf("GifterChannelID = %q, want 'gifter123'", event.GifterChannelID)
		}
		if event.AssociatedGiftingMessageID != "giftmsg1" {
			t.Errorf("AssociatedGiftingMessageID = %q, want 'giftmsg1'", event.AssociatedGiftingMessageID)
		}
		if event.Author.DisplayName != "Recipient User" {
			t.Errorf("Author.DisplayName = %q, want 'Recipient User'", event.Author.DisplayName)
		}
	}
	mu.Unlock()

	_ = bot.Close()
}

func TestChatBotClient_OnMessageDeleted(t *testing.T) {
	var deletedID string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 100,
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
	bot, _ := NewChatBotClient(client, nil, "chat123")

	bot.OnMessageDeleted(func(id string) {
		mu.Lock()
		deletedID = id
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = bot.Connect(ctx)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	got := deletedID
	mu.Unlock()

	if got != "deleted123" {
		t.Errorf("deletedID = %q, want 'deleted123'", got)
	}

	_ = bot.Close()
}

func TestChatBotClient_OnUserBanned(t *testing.T) {
	var bannedEvent *BanEvent
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 100,
			Items: []*LiveChatMessage{
				{
					ID: "ban-event",
					Snippet: &MessageSnippet{
						Type: MessageTypeUserBanned,
						UserBannedDetails: &UserBannedDetails{
							BanType:            BanTypeTemporary,
							BanDurationSeconds: 300,
							BannedUserDetails: &BannedUserDetails{
								ChannelID:   "banned123",
								DisplayName: "Bad User",
							},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	bot, _ := NewChatBotClient(client, nil, "chat123")

	bot.OnUserBanned(func(event *BanEvent) {
		mu.Lock()
		bannedEvent = event
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = bot.Connect(ctx)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	got := bannedEvent
	mu.Unlock()

	if got == nil {
		t.Fatal("No BanEvent received")
	}
	if got.BanType != BanTypeTemporary {
		t.Errorf("BanType = %q, want %q", got.BanType, BanTypeTemporary)
	}
	if got.Duration != 300*time.Second {
		t.Errorf("Duration = %v, want 300s", got.Duration)
	}
	if got.BannedUser.ChannelID != "banned123" {
		t.Errorf("BannedUser.ChannelID = %q, want 'banned123'", got.BannedUser.ChannelID)
	}

	_ = bot.Close()
}

func TestChatBotClient_OnConnectDisconnect(t *testing.T) {
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
	bot, _ := NewChatBotClient(client, nil, "chat123")

	bot.OnConnect(func() {
		connectCalled.Store(true)
	})

	bot.OnDisconnect(func() {
		disconnectCalled.Store(true)
	})

	ctx := context.Background()
	_ = bot.Connect(ctx)

	time.Sleep(50 * time.Millisecond)

	if !connectCalled.Load() {
		t.Error("OnConnect handler not called")
	}

	_ = bot.Close()

	// Give more time for the disconnect handler to be called
	time.Sleep(200 * time.Millisecond)

	if !disconnectCalled.Load() {
		t.Error("OnDisconnect handler not called")
	}
}

func TestChatBotClient_OnError(t *testing.T) {
	var errorReceived atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": {"message": "test error"}}`))
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	bot, _ := NewChatBotClient(client, nil, "chat123")

	bot.OnError(func(err error) {
		errorReceived.Store(true)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = bot.Connect(ctx)

	time.Sleep(100 * time.Millisecond)

	if !errorReceived.Load() {
		t.Error("OnError handler not called")
	}

	_ = bot.Close()
}

func TestChatBotClient_Say(t *testing.T) {
	var receivedBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/liveChat/messages" && r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&receivedBody)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(LiveChatMessage{ID: "sent123"})
			return
		}
		// Return empty poll response
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 5000,
			Items:                 []*LiveChatMessage{},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	bot, _ := NewChatBotClient(client, nil, "chat123")

	ctx := context.Background()
	_ = bot.Connect(ctx)

	time.Sleep(50 * time.Millisecond)

	err := bot.Say(ctx, "Hello!")
	if err != nil {
		t.Fatalf("Say() error = %v", err)
	}

	snippet, ok := receivedBody["snippet"].(map[string]any)
	if !ok {
		t.Fatal("snippet not in request body")
	}
	if snippet["liveChatId"] != "chat123" {
		t.Errorf("liveChatId = %q, want 'chat123'", snippet["liveChatId"])
	}

	_ = bot.Close()
}

func TestChatBotClient_Say_NotConnected(t *testing.T) {
	client := core.NewClient()
	bot, _ := NewChatBotClient(client, nil, "chat123")

	err := bot.Say(context.Background(), "Hello!")
	if err != ErrNotRunning {
		t.Errorf("Say() error = %v, want ErrNotRunning", err)
	}
}

func TestChatBotClient_Delete_NotConnected(t *testing.T) {
	client := core.NewClient()
	bot, _ := NewChatBotClient(client, nil, "chat123")

	err := bot.Delete(context.Background(), "msg123")
	if err != ErrNotRunning {
		t.Errorf("Delete() error = %v, want ErrNotRunning", err)
	}
}

func TestChatBotClient_Ban_NotConnected(t *testing.T) {
	client := core.NewClient()
	bot, _ := NewChatBotClient(client, nil, "chat123")

	err := bot.Ban(context.Background(), "channel123")
	if err != ErrNotRunning {
		t.Errorf("Ban() error = %v, want ErrNotRunning", err)
	}
}

func TestChatBotClient_Timeout_NotConnected(t *testing.T) {
	client := core.NewClient()
	bot, _ := NewChatBotClient(client, nil, "chat123")

	err := bot.Timeout(context.Background(), "channel123", 300)
	if err != ErrNotRunning {
		t.Errorf("Timeout() error = %v, want ErrNotRunning", err)
	}
}

func TestChatBotClient_Unban_NotConnected(t *testing.T) {
	client := core.NewClient()
	bot, _ := NewChatBotClient(client, nil, "chat123")

	err := bot.Unban(context.Background(), "ban123")
	if err != ErrNotRunning {
		t.Errorf("Unban() error = %v, want ErrNotRunning", err)
	}
}

func TestChatBotClient_AddModerator_NotConnected(t *testing.T) {
	client := core.NewClient()
	bot, _ := NewChatBotClient(client, nil, "chat123")

	err := bot.AddModerator(context.Background(), "channel123")
	if err != ErrNotRunning {
		t.Errorf("AddModerator() error = %v, want ErrNotRunning", err)
	}
}

func TestChatBotClient_RemoveModerator_NotConnected(t *testing.T) {
	client := core.NewClient()
	bot, _ := NewChatBotClient(client, nil, "chat123")

	err := bot.RemoveModerator(context.Background(), "mod123")
	if err != ErrNotRunning {
		t.Errorf("RemoveModerator() error = %v, want ErrNotRunning", err)
	}
}

func TestChatBotClient_HandlerUnsubscribe(t *testing.T) {
	client := core.NewClient()
	bot, _ := NewChatBotClient(client, nil, "chat123")

	// Register all handler types
	unsubs := []func(){
		bot.OnMessage(func(msg *ChatMessage) {}),
		bot.OnSuperChat(func(event *SuperChatEvent) {}),
		bot.OnSuperSticker(func(event *SuperStickerEvent) {}),
		bot.OnMembership(func(event *MembershipEvent) {}),
		bot.OnMemberMilestone(func(event *MemberMilestoneEvent) {}),
		bot.OnGiftMembership(func(event *GiftMembershipEvent) {}),
		bot.OnGiftMembershipReceived(func(event *GiftMembershipReceivedEvent) {}),
		bot.OnMessageDeleted(func(id string) {}),
		bot.OnUserBanned(func(event *BanEvent) {}),
		bot.OnConnect(func() {}),
		bot.OnDisconnect(func() {}),
		bot.OnError(func(err error) {}),
	}

	// Verify all handlers registered
	bot.mu.RLock()
	if len(bot.messageHandlers) != 1 {
		t.Error("messageHandlers not registered")
	}
	if len(bot.superChatHandlers) != 1 {
		t.Error("superChatHandlers not registered")
	}
	if len(bot.giftMembershipReceivedHandlers) != 1 {
		t.Error("giftMembershipReceivedHandlers not registered")
	}
	if len(bot.connectHandlers) != 1 {
		t.Error("connectHandlers not registered")
	}
	bot.mu.RUnlock()

	// Unsubscribe all
	for _, unsub := range unsubs {
		unsub()
	}

	// Verify all handlers removed
	bot.mu.RLock()
	if len(bot.messageHandlers) != 0 {
		t.Error("messageHandlers not unsubscribed")
	}
	if len(bot.superChatHandlers) != 0 {
		t.Error("superChatHandlers not unsubscribed")
	}
	if len(bot.giftMembershipReceivedHandlers) != 0 {
		t.Error("giftMembershipReceivedHandlers not unsubscribed")
	}
	if len(bot.connectHandlers) != 0 {
		t.Error("connectHandlers not unsubscribed")
	}
	bot.mu.RUnlock()
}

func TestChatBotClient_HandleMessageNilSnippet(t *testing.T) {
	client := core.NewClient()
	bot, _ := NewChatBotClient(client, nil, "chat123")

	var called bool
	bot.OnMessage(func(msg *ChatMessage) {
		called = true
	})

	// This should not panic
	bot.handleMessage(&LiveChatMessage{
		ID:      "msg1",
		Snippet: nil,
	})

	if called {
		t.Error("Handler called for message with nil snippet")
	}
}

func TestParseAuthor(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		author := parseAuthor(nil)
		if author != nil {
			t.Errorf("parseAuthor(nil) = %v, want nil", author)
		}
	})

	t.Run("full author", func(t *testing.T) {
		ad := &AuthorDetails{
			ChannelID:       "channel123",
			DisplayName:     "Test User",
			ProfileImageURL: "https://example.com/image.jpg",
			IsChatModerator: true,
			IsChatOwner:     false,
			IsChatSponsor:   true,
			IsVerified:      true,
		}

		author := parseAuthor(ad)

		if author.ChannelID != "channel123" {
			t.Errorf("ChannelID = %q, want 'channel123'", author.ChannelID)
		}
		if author.DisplayName != "Test User" {
			t.Errorf("DisplayName = %q, want 'Test User'", author.DisplayName)
		}
		if !author.IsModerator {
			t.Error("IsModerator = false, want true")
		}
		if author.IsOwner {
			t.Error("IsOwner = true, want false")
		}
		if !author.IsMember {
			t.Error("IsMember = false, want true")
		}
		if !author.IsVerified {
			t.Error("IsVerified = false, want true")
		}
	})
}

func TestChatBotClient_Close_Idempotent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveChatMessageListResponse{
			PollingIntervalMillis: 1000,
			Items:                 []*LiveChatMessage{},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	bot, _ := NewChatBotClient(client, nil, "chat123")

	ctx := context.Background()
	_ = bot.Connect(ctx)

	// Close multiple times should be safe
	_ = bot.Close()
	_ = bot.Close()
	_ = bot.Close()
}
