package streaming

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Its-donkey/yougopher/youtube/core"
)

func TestGetBroadcasts(t *testing.T) {
	t.Run("success with IDs", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/liveBroadcasts" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("id") != "broadcast123" {
				t.Errorf("unexpected id: %s", r.URL.Query().Get("id"))
			}
			if r.URL.Query().Get("part") != "snippet,status" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}

			resp := LiveBroadcastListResponse{
				Kind: "youtube#liveBroadcastListResponse",
				Items: []*LiveBroadcast{
					{
						ID:   "broadcast123",
						Kind: "youtube#liveBroadcast",
						Snippet: &BroadcastSnippet{
							Title:      "Test Broadcast",
							LiveChatID: "chat123",
						},
						Status: &BroadcastStatus{
							LifeCycleStatus: BroadcastStatusLive,
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		resp, err := GetBroadcasts(context.Background(), client, &GetBroadcastsParams{
			IDs: []string{"broadcast123"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(resp.Items))
		}
		if resp.Items[0].ID != "broadcast123" {
			t.Errorf("unexpected ID: %s", resp.Items[0].ID)
		}
		if resp.Items[0].Snippet.Title != "Test Broadcast" {
			t.Errorf("unexpected title: %s", resp.Items[0].Snippet.Title)
		}
		if resp.Items[0].Snippet.LiveChatID != "chat123" {
			t.Errorf("unexpected live chat ID: %s", resp.Items[0].Snippet.LiveChatID)
		}
	})

	t.Run("nil params", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetBroadcasts(context.Background(), client, nil)
		if err == nil {
			t.Fatal("expected error for nil params")
		}
	})

	t.Run("no filter provided", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetBroadcasts(context.Background(), client, &GetBroadcastsParams{})
		if err == nil {
			t.Fatal("expected error for no filter")
		}
	})

	t.Run("with mine", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("mine") != "true" {
				t.Errorf("unexpected mine: %s", r.URL.Query().Get("mine"))
			}

			resp := LiveBroadcastListResponse{Items: []*LiveBroadcast{{ID: "broadcast123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetBroadcasts(context.Background(), client, &GetBroadcastsParams{
			Mine: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with broadcast status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("broadcastStatus") != "active" {
				t.Errorf("unexpected broadcastStatus: %s", r.URL.Query().Get("broadcastStatus"))
			}

			resp := LiveBroadcastListResponse{Items: []*LiveBroadcast{{ID: "broadcast123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetBroadcasts(context.Background(), client, &GetBroadcastsParams{
			Mine:            true,
			BroadcastStatus: "active",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with broadcast type", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("broadcastType") != "event" {
				t.Errorf("unexpected broadcastType: %s", r.URL.Query().Get("broadcastType"))
			}

			resp := LiveBroadcastListResponse{Items: []*LiveBroadcast{{ID: "broadcast123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetBroadcasts(context.Background(), client, &GetBroadcastsParams{
			Mine:          true,
			BroadcastType: "event",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "snippet,contentDetails" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}

			resp := LiveBroadcastListResponse{Items: []*LiveBroadcast{{ID: "broadcast123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetBroadcasts(context.Background(), client, &GetBroadcastsParams{
			IDs:   []string{"broadcast123"},
			Parts: []string{"snippet", "contentDetails"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with pagination", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("maxResults") != "10" {
				t.Errorf("unexpected maxResults: %s", r.URL.Query().Get("maxResults"))
			}
			if r.URL.Query().Get("pageToken") != "nextPage123" {
				t.Errorf("unexpected pageToken: %s", r.URL.Query().Get("pageToken"))
			}

			resp := LiveBroadcastListResponse{Items: []*LiveBroadcast{{ID: "broadcast123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetBroadcasts(context.Background(), client, &GetBroadcastsParams{
			IDs:        []string{"broadcast123"},
			MaxResults: 10,
			PageToken:  "nextPage123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetBroadcast(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := LiveBroadcastListResponse{
				Items: []*LiveBroadcast{{ID: "broadcast123", Snippet: &BroadcastSnippet{Title: "Test"}}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		broadcast, err := GetBroadcast(context.Background(), client, "broadcast123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if broadcast.ID != "broadcast123" {
			t.Errorf("unexpected ID: %s", broadcast.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := LiveBroadcastListResponse{Items: []*LiveBroadcast{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetBroadcast(context.Background(), client, "nonexistent")
		if err == nil {
			t.Fatal("expected error for not found broadcast")
		}
		notFoundErr, ok := err.(*core.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T", err)
		}
		if notFoundErr.ResourceType != "broadcast" {
			t.Errorf("unexpected resource type: %s", notFoundErr.ResourceType)
		}
	})

	t.Run("empty broadcast ID", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetBroadcast(context.Background(), client, "")
		if err == nil {
			t.Fatal("expected error for empty broadcast ID")
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "snippet" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}
			resp := LiveBroadcastListResponse{Items: []*LiveBroadcast{{ID: "broadcast123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetBroadcast(context.Background(), client, "broadcast123", "snippet")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetMyActiveBroadcast(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("mine") != "true" {
				t.Errorf("expected mine=true, got %s", r.URL.Query().Get("mine"))
			}
			if r.URL.Query().Get("broadcastStatus") != "active" {
				t.Errorf("expected broadcastStatus=active, got %s", r.URL.Query().Get("broadcastStatus"))
			}

			resp := LiveBroadcastListResponse{
				Items: []*LiveBroadcast{{ID: "activeBroadcast", Snippet: &BroadcastSnippet{Title: "Active Stream"}}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		broadcast, err := GetMyActiveBroadcast(context.Background(), client)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if broadcast.ID != "activeBroadcast" {
			t.Errorf("unexpected ID: %s", broadcast.ID)
		}
	})

	t.Run("no active broadcast", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := LiveBroadcastListResponse{Items: []*LiveBroadcast{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetMyActiveBroadcast(context.Background(), client)
		if err == nil {
			t.Fatal("expected error for no active broadcast")
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "contentDetails" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}
			resp := LiveBroadcastListResponse{Items: []*LiveBroadcast{{ID: "broadcast123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetMyActiveBroadcast(context.Background(), client, "contentDetails")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetBroadcastLiveChatID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := LiveBroadcastListResponse{
				Items: []*LiveBroadcast{{
					ID: "broadcast123",
					Snippet: &BroadcastSnippet{
						LiveChatID: "chat123",
					},
				}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		chatID, err := GetBroadcastLiveChatID(context.Background(), client, "broadcast123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if chatID != "chat123" {
			t.Errorf("unexpected chat ID: %s", chatID)
		}
	})

	t.Run("no snippet", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := LiveBroadcastListResponse{Items: []*LiveBroadcast{{ID: "broadcast123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetBroadcastLiveChatID(context.Background(), client, "broadcast123")
		if err == nil {
			t.Fatal("expected error for no snippet")
		}
	})

	t.Run("no live chat", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := LiveBroadcastListResponse{
				Items: []*LiveBroadcast{{
					ID:      "broadcast123",
					Snippet: &BroadcastSnippet{Title: "No Chat"},
				}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetBroadcastLiveChatID(context.Background(), client, "broadcast123")
		if err == nil {
			t.Fatal("expected error for no live chat")
		}
	})
}

func TestLiveBroadcast_Methods(t *testing.T) {
	t.Run("IsLive", func(t *testing.T) {
		tests := []struct {
			name      string
			broadcast *LiveBroadcast
			want      bool
		}{
			{"nil status", &LiveBroadcast{}, false},
			{"not live", &LiveBroadcast{Status: &BroadcastStatus{LifeCycleStatus: BroadcastStatusReady}}, false},
			{"live", &LiveBroadcast{Status: &BroadcastStatus{LifeCycleStatus: BroadcastStatusLive}}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.broadcast.IsLive(); got != tt.want {
					t.Errorf("IsLive() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("IsComplete", func(t *testing.T) {
		tests := []struct {
			name      string
			broadcast *LiveBroadcast
			want      bool
		}{
			{"nil status", &LiveBroadcast{}, false},
			{"not complete", &LiveBroadcast{Status: &BroadcastStatus{LifeCycleStatus: BroadcastStatusLive}}, false},
			{"complete", &LiveBroadcast{Status: &BroadcastStatus{LifeCycleStatus: BroadcastStatusComplete}}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.broadcast.IsComplete(); got != tt.want {
					t.Errorf("IsComplete() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("IsUpcoming", func(t *testing.T) {
		tests := []struct {
			name      string
			broadcast *LiveBroadcast
			want      bool
		}{
			{"nil status", &LiveBroadcast{}, false},
			{"live", &LiveBroadcast{Status: &BroadcastStatus{LifeCycleStatus: BroadcastStatusLive}}, false},
			{"created", &LiveBroadcast{Status: &BroadcastStatus{LifeCycleStatus: BroadcastStatusCreated}}, true},
			{"ready", &LiveBroadcast{Status: &BroadcastStatus{LifeCycleStatus: BroadcastStatusReady}}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.broadcast.IsUpcoming(); got != tt.want {
					t.Errorf("IsUpcoming() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("LiveChatID", func(t *testing.T) {
		tests := []struct {
			name      string
			broadcast *LiveBroadcast
			want      string
		}{
			{"nil snippet", &LiveBroadcast{}, ""},
			{"no chat ID", &LiveBroadcast{Snippet: &BroadcastSnippet{}}, ""},
			{"has chat ID", &LiveBroadcast{Snippet: &BroadcastSnippet{LiveChatID: "chat123"}}, "chat123"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.broadcast.LiveChatID(); got != tt.want {
					t.Errorf("LiveChatID() = %q, want %q", got, tt.want)
				}
			})
		}
	})
}

func TestLiveBroadcastListResponse_JSON(t *testing.T) {
	jsonData := `{
		"kind": "youtube#liveBroadcastListResponse",
		"etag": "abc123",
		"nextPageToken": "page2",
		"pageInfo": {
			"totalResults": 100,
			"resultsPerPage": 50
		},
		"items": [
			{
				"id": "broadcast123",
				"snippet": {
					"publishedAt": "2024-01-15T10:30:00Z",
					"channelId": "channel123",
					"title": "Test Broadcast",
					"description": "A test broadcast",
					"liveChatId": "chat123"
				},
				"status": {
					"lifeCycleStatus": "live",
					"privacyStatus": "public",
					"recordingStatus": "recording"
				}
			}
		]
	}`

	var resp LiveBroadcastListResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if resp.NextPageToken != "page2" {
		t.Errorf("NextPageToken = %q, want 'page2'", resp.NextPageToken)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(resp.Items))
	}

	broadcast := resp.Items[0]
	if broadcast.Snippet.Title != "Test Broadcast" {
		t.Errorf("Title = %q, want 'Test Broadcast'", broadcast.Snippet.Title)
	}
	if broadcast.Snippet.LiveChatID != "chat123" {
		t.Errorf("LiveChatID = %q, want 'chat123'", broadcast.Snippet.LiveChatID)
	}
	if broadcast.Status.LifeCycleStatus != BroadcastStatusLive {
		t.Errorf("LifeCycleStatus = %q, want %q", broadcast.Status.LifeCycleStatus, BroadcastStatusLive)
	}
	if !broadcast.IsLive() {
		t.Error("IsLive() = false, want true")
	}
}

func TestBroadcastStatusConstants(t *testing.T) {
	expectedStatuses := map[string]string{
		"complete":     BroadcastStatusComplete,
		"created":      BroadcastStatusCreated,
		"live":         BroadcastStatusLive,
		"liveStarting": BroadcastStatusLiveStarting,
		"ready":        BroadcastStatusReady,
		"revoked":      BroadcastStatusRevoked,
		"testStarting": BroadcastStatusTestStarting,
		"testing":      BroadcastStatusTesting,
	}

	for expected, actual := range expectedStatuses {
		if actual != expected {
			t.Errorf("BroadcastStatus constant %q = %q, want %q", expected, actual, expected)
		}
	}
}

func TestInsertBroadcast(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/liveBroadcasts" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			var body LiveBroadcast
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if body.Snippet.Title != "New Broadcast" {
				t.Errorf("unexpected title: %s", body.Snippet.Title)
			}

			resp := LiveBroadcast{
				ID: "newBroadcast123",
				Snippet: &BroadcastSnippet{
					Title:      "New Broadcast",
					LiveChatID: "chat123",
				},
				Status: &BroadcastStatus{
					LifeCycleStatus: BroadcastStatusCreated,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		broadcast, err := InsertBroadcast(context.Background(), client, &LiveBroadcast{
			Snippet: &BroadcastSnippet{
				Title: "New Broadcast",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if broadcast.ID != "newBroadcast123" {
			t.Errorf("unexpected ID: %s", broadcast.ID)
		}
		if broadcast.Snippet.LiveChatID != "chat123" {
			t.Errorf("unexpected live chat ID: %s", broadcast.Snippet.LiveChatID)
		}
	})

	t.Run("nil broadcast", func(t *testing.T) {
		client := core.NewClient()
		_, err := InsertBroadcast(context.Background(), client, nil)
		if err == nil {
			t.Fatal("expected error for nil broadcast")
		}
	})

	t.Run("missing title", func(t *testing.T) {
		client := core.NewClient()
		_, err := InsertBroadcast(context.Background(), client, &LiveBroadcast{
			Snippet: &BroadcastSnippet{},
		})
		if err == nil {
			t.Fatal("expected error for missing title")
		}
	})

	t.Run("nil snippet", func(t *testing.T) {
		client := core.NewClient()
		_, err := InsertBroadcast(context.Background(), client, &LiveBroadcast{})
		if err == nil {
			t.Fatal("expected error for nil snippet")
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "snippet,contentDetails" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}

			resp := LiveBroadcast{ID: "broadcast123"}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := InsertBroadcast(context.Background(), client, &LiveBroadcast{
			Snippet: &BroadcastSnippet{Title: "Test"},
		}, "snippet", "contentDetails")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestUpdateBroadcast(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT, got %s", r.Method)
			}

			var body LiveBroadcast
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if body.ID != "broadcast123" {
				t.Errorf("unexpected ID: %s", body.ID)
			}
			if body.Snippet.Title != "Updated Title" {
				t.Errorf("unexpected title: %s", body.Snippet.Title)
			}

			resp := LiveBroadcast{
				ID:      "broadcast123",
				Snippet: &BroadcastSnippet{Title: "Updated Title"},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		broadcast, err := UpdateBroadcast(context.Background(), client, &LiveBroadcast{
			ID:      "broadcast123",
			Snippet: &BroadcastSnippet{Title: "Updated Title"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if broadcast.Snippet.Title != "Updated Title" {
			t.Errorf("unexpected title: %s", broadcast.Snippet.Title)
		}
	})

	t.Run("nil broadcast", func(t *testing.T) {
		client := core.NewClient()
		_, err := UpdateBroadcast(context.Background(), client, nil)
		if err == nil {
			t.Fatal("expected error for nil broadcast")
		}
	})

	t.Run("missing ID", func(t *testing.T) {
		client := core.NewClient()
		_, err := UpdateBroadcast(context.Background(), client, &LiveBroadcast{
			Snippet: &BroadcastSnippet{Title: "Test"},
		})
		if err == nil {
			t.Fatal("expected error for missing ID")
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "snippet" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}

			resp := LiveBroadcast{ID: "broadcast123"}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := UpdateBroadcast(context.Background(), client, &LiveBroadcast{
			ID:      "broadcast123",
			Snippet: &BroadcastSnippet{Title: "Test"},
		}, "snippet")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestDeleteBroadcast(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			if r.URL.Query().Get("id") != "broadcast123" {
				t.Errorf("unexpected id: %s", r.URL.Query().Get("id"))
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		err := DeleteBroadcast(context.Background(), client, "broadcast123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty broadcast ID", func(t *testing.T) {
		client := core.NewClient()
		err := DeleteBroadcast(context.Background(), client, "")
		if err == nil {
			t.Fatal("expected error for empty broadcast ID")
		}
	})
}

func TestBindBroadcast(t *testing.T) {
	t.Run("success with stream", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/liveBroadcasts/bind" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("id") != "broadcast123" {
				t.Errorf("unexpected id: %s", r.URL.Query().Get("id"))
			}
			if r.URL.Query().Get("streamId") != "stream456" {
				t.Errorf("unexpected streamId: %s", r.URL.Query().Get("streamId"))
			}

			resp := LiveBroadcast{
				ID: "broadcast123",
				ContentDetails: &BroadcastContentDetails{
					BoundStreamID: "stream456",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		broadcast, err := BindBroadcast(context.Background(), client, &BindBroadcastParams{
			BroadcastID: "broadcast123",
			StreamID:    "stream456",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if broadcast.ContentDetails.BoundStreamID != "stream456" {
			t.Errorf("unexpected bound stream ID: %s", broadcast.ContentDetails.BoundStreamID)
		}
	})

	t.Run("unbind (no stream)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("streamId") != "" {
				t.Errorf("streamId should be empty for unbind, got: %s", r.URL.Query().Get("streamId"))
			}

			resp := LiveBroadcast{ID: "broadcast123"}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := BindBroadcast(context.Background(), client, &BindBroadcastParams{
			BroadcastID: "broadcast123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("nil params", func(t *testing.T) {
		client := core.NewClient()
		_, err := BindBroadcast(context.Background(), client, nil)
		if err == nil {
			t.Fatal("expected error for nil params")
		}
	})

	t.Run("missing broadcast ID", func(t *testing.T) {
		client := core.NewClient()
		_, err := BindBroadcast(context.Background(), client, &BindBroadcastParams{
			StreamID: "stream123",
		})
		if err == nil {
			t.Fatal("expected error for missing broadcast ID")
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "contentDetails" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}

			resp := LiveBroadcast{ID: "broadcast123"}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := BindBroadcast(context.Background(), client, &BindBroadcastParams{
			BroadcastID: "broadcast123",
			StreamID:    "stream456",
			Parts:       []string{"contentDetails"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestTransitionBroadcast(t *testing.T) {
	t.Run("success testing", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/liveBroadcasts/transition" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("id") != "broadcast123" {
				t.Errorf("unexpected id: %s", r.URL.Query().Get("id"))
			}
			if r.URL.Query().Get("broadcastStatus") != "testing" {
				t.Errorf("unexpected status: %s", r.URL.Query().Get("broadcastStatus"))
			}

			resp := LiveBroadcast{
				ID: "broadcast123",
				Status: &BroadcastStatus{
					LifeCycleStatus: BroadcastStatusTesting,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		broadcast, err := TransitionBroadcast(context.Background(), client, "broadcast123", TransitionTesting)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if broadcast.Status.LifeCycleStatus != BroadcastStatusTesting {
			t.Errorf("unexpected status: %s", broadcast.Status.LifeCycleStatus)
		}
	})

	t.Run("success live", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("broadcastStatus") != "live" {
				t.Errorf("unexpected status: %s", r.URL.Query().Get("broadcastStatus"))
			}

			resp := LiveBroadcast{
				ID:     "broadcast123",
				Status: &BroadcastStatus{LifeCycleStatus: BroadcastStatusLive},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		broadcast, err := TransitionBroadcast(context.Background(), client, "broadcast123", TransitionLive)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !broadcast.IsLive() {
			t.Error("broadcast should be live")
		}
	})

	t.Run("success complete", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("broadcastStatus") != "complete" {
				t.Errorf("unexpected status: %s", r.URL.Query().Get("broadcastStatus"))
			}

			resp := LiveBroadcast{
				ID:     "broadcast123",
				Status: &BroadcastStatus{LifeCycleStatus: BroadcastStatusComplete},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		broadcast, err := TransitionBroadcast(context.Background(), client, "broadcast123", TransitionComplete)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !broadcast.IsComplete() {
			t.Error("broadcast should be complete")
		}
	})

	t.Run("empty broadcast ID", func(t *testing.T) {
		client := core.NewClient()
		_, err := TransitionBroadcast(context.Background(), client, "", TransitionLive)
		if err == nil {
			t.Fatal("expected error for empty broadcast ID")
		}
	})

	t.Run("empty status", func(t *testing.T) {
		client := core.NewClient()
		_, err := TransitionBroadcast(context.Background(), client, "broadcast123", "")
		if err == nil {
			t.Fatal("expected error for empty status")
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		client := core.NewClient()
		_, err := TransitionBroadcast(context.Background(), client, "broadcast123", "invalid")
		if err == nil {
			t.Fatal("expected error for invalid status")
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "status" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}

			resp := LiveBroadcast{ID: "broadcast123"}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := TransitionBroadcast(context.Background(), client, "broadcast123", TransitionLive, "status")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestLiveBroadcast_NewMethods(t *testing.T) {
	t.Run("IsTesting", func(t *testing.T) {
		tests := []struct {
			name      string
			broadcast *LiveBroadcast
			want      bool
		}{
			{"nil status", &LiveBroadcast{}, false},
			{"not testing", &LiveBroadcast{Status: &BroadcastStatus{LifeCycleStatus: BroadcastStatusLive}}, false},
			{"testing", &LiveBroadcast{Status: &BroadcastStatus{LifeCycleStatus: BroadcastStatusTesting}}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.broadcast.IsTesting(); got != tt.want {
					t.Errorf("IsTesting() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("BoundStreamID", func(t *testing.T) {
		tests := []struct {
			name      string
			broadcast *LiveBroadcast
			want      string
		}{
			{"nil contentDetails", &LiveBroadcast{}, ""},
			{"no bound stream", &LiveBroadcast{ContentDetails: &BroadcastContentDetails{}}, ""},
			{"has bound stream", &LiveBroadcast{ContentDetails: &BroadcastContentDetails{BoundStreamID: "stream123"}}, "stream123"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.broadcast.BoundStreamID(); got != tt.want {
					t.Errorf("BoundStreamID() = %q, want %q", got, tt.want)
				}
			})
		}
	})

	t.Run("HasBoundStream", func(t *testing.T) {
		tests := []struct {
			name      string
			broadcast *LiveBroadcast
			want      bool
		}{
			{"nil contentDetails", &LiveBroadcast{}, false},
			{"no bound stream", &LiveBroadcast{ContentDetails: &BroadcastContentDetails{}}, false},
			{"has bound stream", &LiveBroadcast{ContentDetails: &BroadcastContentDetails{BoundStreamID: "stream123"}}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.broadcast.HasBoundStream(); got != tt.want {
					t.Errorf("HasBoundStream() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("TotalChatCount", func(t *testing.T) {
		tests := []struct {
			name      string
			broadcast *LiveBroadcast
			want      uint64
		}{
			{"nil statistics", &LiveBroadcast{}, 0},
			{"zero count", &LiveBroadcast{Statistics: &BroadcastStatistics{}}, 0},
			{"has count", &LiveBroadcast{Statistics: &BroadcastStatistics{TotalChatCount: 12345}}, 12345},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.broadcast.TotalChatCount(); got != tt.want {
					t.Errorf("TotalChatCount() = %d, want %d", got, tt.want)
				}
			})
		}
	})

	t.Run("HasCuepointSchedule", func(t *testing.T) {
		tests := []struct {
			name      string
			broadcast *LiveBroadcast
			want      bool
		}{
			{"nil monetizationDetails", &LiveBroadcast{}, false},
			{"nil cuepointSchedule", &LiveBroadcast{MonetizationDetails: &BroadcastMonetizationDetails{}}, false},
			{"disabled", &LiveBroadcast{MonetizationDetails: &BroadcastMonetizationDetails{CuepointSchedule: &CuepointSchedule{Enabled: false}}}, false},
			{"enabled", &LiveBroadcast{MonetizationDetails: &BroadcastMonetizationDetails{CuepointSchedule: &CuepointSchedule{Enabled: true}}}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.broadcast.HasCuepointSchedule(); got != tt.want {
					t.Errorf("HasCuepointSchedule() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("CuepointRepeatInterval", func(t *testing.T) {
		tests := []struct {
			name      string
			broadcast *LiveBroadcast
			want      int
		}{
			{"nil monetizationDetails", &LiveBroadcast{}, 0},
			{"nil cuepointSchedule", &LiveBroadcast{MonetizationDetails: &BroadcastMonetizationDetails{}}, 0},
			{"zero interval", &LiveBroadcast{MonetizationDetails: &BroadcastMonetizationDetails{CuepointSchedule: &CuepointSchedule{}}}, 0},
			{"has interval", &LiveBroadcast{MonetizationDetails: &BroadcastMonetizationDetails{CuepointSchedule: &CuepointSchedule{RepeatIntervalSecs: 300}}}, 300},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.broadcast.CuepointRepeatInterval(); got != tt.want {
					t.Errorf("CuepointRepeatInterval() = %d, want %d", got, tt.want)
				}
			})
		}
	})
}

func TestTransitionConstants(t *testing.T) {
	expectedTransitions := map[string]string{
		"testing":  TransitionTesting,
		"live":     TransitionLive,
		"complete": TransitionComplete,
	}

	for expected, actual := range expectedTransitions {
		if actual != expected {
			t.Errorf("Transition constant %q = %q, want %q", expected, actual, expected)
		}
	}
}

func TestInsertCuepoint(t *testing.T) {
	t.Run("success immediate", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/liveBroadcasts/cuepoint" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("id") != "broadcast123" {
				t.Errorf("unexpected id: %s", r.URL.Query().Get("id"))
			}

			var body CuepointRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if body.CueType != CueTypeAd {
				t.Errorf("unexpected cueType: %s", body.CueType)
			}

			resp := Cuepoint{
				ID:      "cuepoint123",
				CueType: CueTypeAd,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		cuepoint, err := InsertCuepoint(context.Background(), client, &InsertCuepointParams{
			BroadcastID:           "broadcast123",
			InsertionOffsetTimeMs: CuepointInsertImmediate,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cuepoint.ID != "cuepoint123" {
			t.Errorf("unexpected ID: %s", cuepoint.ID)
		}
	})

	t.Run("success with duration", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body CuepointRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if body.DurationSecs != 60 {
				t.Errorf("unexpected duration: %d", body.DurationSecs)
			}

			resp := Cuepoint{
				ID:           "cuepoint123",
				CueType:      CueTypeAd,
				DurationSecs: 60,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		cuepoint, err := InsertCuepoint(context.Background(), client, &InsertCuepointParams{
			BroadcastID:  "broadcast123",
			DurationSecs: 60,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cuepoint.DurationSecs != 60 {
			t.Errorf("unexpected duration: %d", cuepoint.DurationSecs)
		}
	})

	t.Run("success with walltime", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body CuepointRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if body.WalltimeMs != 1705318200000 {
				t.Errorf("unexpected walltime: %d", body.WalltimeMs)
			}

			resp := Cuepoint{
				ID:         "cuepoint123",
				CueType:    CueTypeAd,
				WalltimeMs: 1705318200000,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := InsertCuepoint(context.Background(), client, &InsertCuepointParams{
			BroadcastID: "broadcast123",
			WalltimeMs:  1705318200000,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("nil params", func(t *testing.T) {
		client := core.NewClient()
		_, err := InsertCuepoint(context.Background(), client, nil)
		if err == nil {
			t.Fatal("expected error for nil params")
		}
	})

	t.Run("missing broadcast ID", func(t *testing.T) {
		client := core.NewClient()
		_, err := InsertCuepoint(context.Background(), client, &InsertCuepointParams{})
		if err == nil {
			t.Fatal("expected error for missing broadcast ID")
		}
	})
}

func TestInsertImmediateCuepoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body CuepointRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.InsertionOffsetTimeMs != CuepointInsertImmediate {
			t.Errorf("unexpected insertionOffsetTimeMs: %d", body.InsertionOffsetTimeMs)
		}

		resp := Cuepoint{ID: "cuepoint123", CueType: CueTypeAd}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	cuepoint, err := InsertImmediateCuepoint(context.Background(), client, "broadcast123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cuepoint.ID != "cuepoint123" {
		t.Errorf("unexpected ID: %s", cuepoint.ID)
	}
}

func TestCuepointConstants(t *testing.T) {
	if CueTypeAd != "cueTypeAd" {
		t.Errorf("CueTypeAd = %q, want 'cueTypeAd'", CueTypeAd)
	}
	if CuepointInsertImmediate != -1 {
		t.Errorf("CuepointInsertImmediate = %d, want -1", CuepointInsertImmediate)
	}
}

// =============================================================================
// Error Type Verification Tests (errors.As)
// =============================================================================

func TestGetBroadcast_ErrorAs_NotFoundError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty items to trigger NotFoundError
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveBroadcastListResponse{
			Items: []*LiveBroadcast{},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	_, err := GetBroadcast(context.Background(), client, "nonexistent123")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Verify using errors.As
	var notFoundErr *core.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("errors.As(err, *NotFoundError) = false, want true; err = %v (%T)", err, err)
	}

	// Verify error fields
	if notFoundErr != nil {
		if notFoundErr.ResourceType != "broadcast" {
			t.Errorf("ResourceType = %q, want 'broadcast'", notFoundErr.ResourceType)
		}
		if notFoundErr.ResourceID != "nonexistent123" {
			t.Errorf("ResourceID = %q, want 'nonexistent123'", notFoundErr.ResourceID)
		}
	}
}

func TestGetMyActiveBroadcast_ReturnsError_WhenNoBroadcast(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveBroadcastListResponse{
			Items: []*LiveBroadcast{},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	_, err := GetMyActiveBroadcast(context.Background(), client)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// GetMyActiveBroadcast returns a plain error (not NotFoundError) for no active broadcast
	expectedMsg := "no active broadcast found"
	if err.Error() != expectedMsg {
		t.Errorf("error message = %q, want %q", err.Error(), expectedMsg)
	}
}

// =============================================================================
// Mock Call Count and Request Verification Tests
// =============================================================================

func TestInsertBroadcast_VerifyAPICall(t *testing.T) {
	var (
		callCount   int
		lastMethod  string
		lastPath    string
		lastBody    map[string]any
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
		_ = json.NewEncoder(w).Encode(LiveBroadcast{
			ID:   "broadcast123",
			Kind: "youtube#liveBroadcast",
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	scheduledStart := time.Now().Add(1 * time.Hour)
	broadcast := &LiveBroadcast{
		Snippet: &BroadcastSnippet{
			Title:              "Test Stream",
			ScheduledStartTime: &scheduledStart,
		},
		Status: &BroadcastStatus{
			PrivacyStatus: "private",
		},
	}

	_, err := InsertBroadcast(context.Background(), client, broadcast)
	if err != nil {
		t.Fatalf("InsertBroadcast() error = %v", err)
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
	if lastPath != "/liveBroadcasts" {
		t.Errorf("path = %s, want /liveBroadcasts", lastPath)
	}

	// Verify Content-Type header
	if ct := lastHeaders.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", ct)
	}

	// Verify request body
	snippet, ok := lastBody["snippet"].(map[string]any)
	if !ok {
		t.Fatal("request body missing snippet")
	}
	if snippet["title"] != "Test Stream" {
		t.Errorf("title = %v, want 'Test Stream'", snippet["title"])
	}
}

func TestDeleteBroadcast_VerifyAPICall(t *testing.T) {
	var (
		callCount  int
		lastMethod string
		lastPath   string
		lastQuery  string
		mu         sync.Mutex
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		lastMethod = r.Method
		lastPath = r.URL.Path
		lastQuery = r.URL.Query().Get("id")
		mu.Unlock()

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	err := DeleteBroadcast(context.Background(), client, "broadcast123")
	if err != nil {
		t.Fatalf("DeleteBroadcast() error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if callCount != 1 {
		t.Errorf("API called %d times, want 1", callCount)
	}

	if lastMethod != http.MethodDelete {
		t.Errorf("method = %s, want DELETE", lastMethod)
	}

	if lastPath != "/liveBroadcasts" {
		t.Errorf("path = %s, want /liveBroadcasts", lastPath)
	}

	if lastQuery != "broadcast123" {
		t.Errorf("id query param = %s, want broadcast123", lastQuery)
	}
}

// =============================================================================
// Integration Test: Complete Broadcast Workflow
// =============================================================================

func TestBroadcastWorkflow_CreateBindTransition(t *testing.T) {
	// Track the workflow stages
	var (
		insertCalled     bool
		bindCalled       bool
		transitionCalled bool
		deleteCalled     bool
		mu               sync.Mutex
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/liveBroadcasts" && r.URL.Query().Get("streamId") == "":
			// Insert broadcast
			insertCalled = true
			_ = json.NewEncoder(w).Encode(LiveBroadcast{
				ID:   "broadcast123",
				Kind: "youtube#liveBroadcast",
				Status: &BroadcastStatus{
					LifeCycleStatus: BroadcastStatusCreated,
				},
			})

		case r.Method == http.MethodPost && r.URL.Path == "/liveBroadcasts/bind":
			// Bind stream to broadcast
			bindCalled = true
			_ = json.NewEncoder(w).Encode(LiveBroadcast{
				ID: "broadcast123",
				ContentDetails: &BroadcastContentDetails{
					BoundStreamID: r.URL.Query().Get("streamId"),
				},
			})

		case r.Method == http.MethodPost && r.URL.Path == "/liveBroadcasts/transition":
			transitionCalled = true
			_ = json.NewEncoder(w).Encode(LiveBroadcast{
				ID: "broadcast123",
				Status: &BroadcastStatus{
					LifeCycleStatus: r.URL.Query().Get("broadcastStatus"),
				},
			})

		case r.Method == http.MethodDelete && r.URL.Path == "/liveBroadcasts":
			deleteCalled = true
			w.WriteHeader(http.StatusNoContent)

		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	ctx := context.Background()

	// Step 1: Create broadcast
	scheduledTime := time.Now().Add(1 * time.Hour)
	broadcast, err := InsertBroadcast(ctx, client, &LiveBroadcast{
		Snippet: &BroadcastSnippet{
			Title:              "Integration Test Stream",
			ScheduledStartTime: &scheduledTime,
		},
		Status: &BroadcastStatus{
			PrivacyStatus: "private",
		},
	})
	if err != nil {
		t.Fatalf("InsertBroadcast() error = %v", err)
	}
	if broadcast.ID != "broadcast123" {
		t.Errorf("broadcast ID = %s, want broadcast123", broadcast.ID)
	}

	// Step 2: Bind stream to broadcast
	bound, err := BindBroadcast(ctx, client, &BindBroadcastParams{
		BroadcastID: "broadcast123",
		StreamID:    "stream123",
	})
	if err != nil {
		t.Fatalf("BindBroadcast() error = %v", err)
	}
	if bound.ContentDetails == nil || bound.ContentDetails.BoundStreamID != "stream123" {
		t.Error("stream not bound correctly")
	}

	// Step 3: Transition to testing
	testing, err := TransitionBroadcast(ctx, client, "broadcast123", TransitionTesting)
	if err != nil {
		t.Fatalf("TransitionBroadcast(testing) error = %v", err)
	}
	if testing.Status == nil || testing.Status.LifeCycleStatus != TransitionTesting {
		t.Error("broadcast not transitioned to testing")
	}

	// Step 4: Delete broadcast (cleanup)
	err = DeleteBroadcast(ctx, client, "broadcast123")
	if err != nil {
		t.Fatalf("DeleteBroadcast() error = %v", err)
	}

	// Verify all stages were called
	mu.Lock()
	defer mu.Unlock()

	if !insertCalled {
		t.Error("insert was not called")
	}
	if !bindCalled {
		t.Error("bind was not called")
	}
	if !transitionCalled {
		t.Error("transition was not called")
	}
	if !deleteCalled {
		t.Error("delete was not called")
	}
}

// =============================================================================
// Mutation Testing - Kill Surviving Mutants
// =============================================================================

// TestGetBroadcasts_DefaultParts verifies default parts are used when none provided.
// Kills mutant: `parts = DefaultBroadcastParts` → removed
func TestGetBroadcasts_DefaultParts(t *testing.T) {
	var capturedParts string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedParts = r.URL.Query().Get("part")
		resp := LiveBroadcastListResponse{Items: []*LiveBroadcast{{ID: "b1"}}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	_, err := GetBroadcasts(context.Background(), client, &GetBroadcastsParams{
		IDs:   []string{"b1"},
		Parts: nil, // No parts - should use DefaultBroadcastParts
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// DefaultBroadcastParts is []string{"snippet", "status"}
	expected := "snippet,status"
	if capturedParts != expected {
		t.Errorf("part = %q, want %q (default parts)", capturedParts, expected)
	}
}

// TestGetBroadcast_DefaultParts verifies GetBroadcast uses default parts when none provided.
// Kills mutant: `if len(parts) == 0` → `if len(parts) == -1` or `== 1`
func TestGetBroadcast_DefaultParts(t *testing.T) {
	var capturedParts string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedParts = r.URL.Query().Get("part")
		resp := LiveBroadcastListResponse{Items: []*LiveBroadcast{{ID: "b1"}}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	// Call with no parts - should use defaults
	_, err := GetBroadcast(context.Background(), client, "b1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "snippet,status"
	if capturedParts != expected {
		t.Errorf("part = %q, want %q (default parts)", capturedParts, expected)
	}
}

// TestInsertBroadcast_DefaultParts verifies InsertBroadcast uses default parts when none provided.
// Kills mutant: `if len(parts) == 0` → `if len(parts) == 1`
func TestInsertBroadcast_DefaultParts(t *testing.T) {
	var capturedParts string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedParts = r.URL.Query().Get("part")
		resp := LiveBroadcast{ID: "b1"}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	scheduledTime := time.Now().Add(1 * time.Hour)
	_, err := InsertBroadcast(context.Background(), client, &LiveBroadcast{
		Snippet: &BroadcastSnippet{
			Title:              "Test",
			ScheduledStartTime: &scheduledTime,
		},
		Status: &BroadcastStatus{PrivacyStatus: "private"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "snippet,status"
	if capturedParts != expected {
		t.Errorf("part = %q, want %q (default parts)", capturedParts, expected)
	}
}

// TestUpdateBroadcast_DefaultParts verifies UpdateBroadcast uses default parts when none provided.
// Kills mutant: `if len(parts) == 0` → `if len(parts) == -1`
func TestUpdateBroadcast_DefaultParts(t *testing.T) {
	var capturedParts string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedParts = r.URL.Query().Get("part")
		resp := LiveBroadcast{ID: "b1"}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	_, err := UpdateBroadcast(context.Background(), client, &LiveBroadcast{
		ID: "b1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "snippet,status"
	if capturedParts != expected {
		t.Errorf("part = %q, want %q (default parts)", capturedParts, expected)
	}
}

// TestGetBroadcasts_MaxResultsBoundary tests MaxResults=1 boundary.
// Kills mutant: `if params.MaxResults > 0` → `> 1`
func TestGetBroadcasts_MaxResultsBoundary(t *testing.T) {
	tests := []struct {
		name        string
		maxResults  int
		shouldHave  bool
		expectedVal string
	}{
		{"maxResults=0 not sent", 0, false, ""},
		{"maxResults=1 sent", 1, true, "1"},
		{"maxResults=5 sent", 5, true, "5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedMaxResults string
			var hasMaxResults bool

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedMaxResults = r.URL.Query().Get("maxResults")
				hasMaxResults = r.URL.Query().Has("maxResults")
				resp := LiveBroadcastListResponse{Items: []*LiveBroadcast{{ID: "b1"}}}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := core.NewClient(core.WithBaseURL(server.URL))
			_, err := GetBroadcasts(context.Background(), client, &GetBroadcastsParams{
				IDs:        []string{"b1"},
				MaxResults: tt.maxResults,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if hasMaxResults != tt.shouldHave {
				t.Errorf("maxResults presence = %v, want %v", hasMaxResults, tt.shouldHave)
			}
			if tt.shouldHave && capturedMaxResults != tt.expectedVal {
				t.Errorf("maxResults = %q, want %q", capturedMaxResults, tt.expectedVal)
			}
		})
	}
}

// TestInsertCuepoint_DurationBoundary tests DurationSecs=1 boundary.
// Kills mutant: `if params.DurationSecs > 0` → `> 1`
func TestInsertCuepoint_DurationBoundary(t *testing.T) {
	tests := []struct {
		name             string
		durationSecs     int
		expectInBody     bool
		expectedDuration int
	}{
		{"duration=0 default used", 0, false, 0},
		{"duration=1 sent", 1, true, 1},
		{"duration=60 sent", 60, true, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedBody CuepointRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewDecoder(r.Body).Decode(&capturedBody)
				resp := Cuepoint{ID: "cue1"}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := core.NewClient(core.WithBaseURL(server.URL))
			_, err := InsertCuepoint(context.Background(), client, &InsertCuepointParams{
				BroadcastID:  "b1",
				DurationSecs: tt.durationSecs,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectInBody && capturedBody.DurationSecs != tt.expectedDuration {
				t.Errorf("DurationSecs = %d, want %d", capturedBody.DurationSecs, tt.expectedDuration)
			}
		})
	}
}

// TestInsertCuepoint_WalltimeBoundary tests WalltimeMs=1 boundary.
// Kills mutant: `if params.WalltimeMs > 0` → `> 1`
func TestInsertCuepoint_WalltimeBoundary(t *testing.T) {
	tests := []struct {
		name           string
		walltimeMs     int64
		expectInBody   bool
		expectedWallMs int64
	}{
		{"walltimeMs=0 not sent", 0, false, 0},
		{"walltimeMs=1 sent", 1, true, 1},
		{"walltimeMs=1000 sent", 1000, true, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedBody CuepointRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewDecoder(r.Body).Decode(&capturedBody)
				resp := Cuepoint{ID: "cue1"}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := core.NewClient(core.WithBaseURL(server.URL))
			_, err := InsertCuepoint(context.Background(), client, &InsertCuepointParams{
				BroadcastID: "b1",
				WalltimeMs:  tt.walltimeMs,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectInBody && capturedBody.WalltimeMs != tt.expectedWallMs {
				t.Errorf("WalltimeMs = %d, want %d", capturedBody.WalltimeMs, tt.expectedWallMs)
			}
		})
	}
}

// TestInsertCuepoint_InsertionOffsetBoundary tests InsertionOffsetTimeMs boundary.
// Kills mutant: `if params.InsertionOffsetTimeMs != 0` → `!= 1`
func TestInsertCuepoint_InsertionOffsetBoundary(t *testing.T) {
	tests := []struct {
		name         string
		offsetMs     int64
		walltimeMs   int64 // Walltime takes precedence
		expectOffset bool
		expectedVal  int64
	}{
		{"offset=0 not sent", 0, 0, false, 0},
		{"offset=1 sent (when walltime=0)", 1, 0, true, 1},
		{"offset=-1 (immediate) sent", CuepointInsertImmediate, 0, true, CuepointInsertImmediate},
		{"offset ignored when walltime set", 100, 500, false, 0}, // Walltime takes precedence
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedBody CuepointRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewDecoder(r.Body).Decode(&capturedBody)
				resp := Cuepoint{ID: "cue1"}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := core.NewClient(core.WithBaseURL(server.URL))
			_, err := InsertCuepoint(context.Background(), client, &InsertCuepointParams{
				BroadcastID:           "b1",
				InsertionOffsetTimeMs: tt.offsetMs,
				WalltimeMs:            tt.walltimeMs,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectOffset && capturedBody.InsertionOffsetTimeMs != tt.expectedVal {
				t.Errorf("InsertionOffsetTimeMs = %d, want %d", capturedBody.InsertionOffsetTimeMs, tt.expectedVal)
			}
			if !tt.expectOffset && tt.walltimeMs > 0 && capturedBody.WalltimeMs != tt.walltimeMs {
				t.Errorf("WalltimeMs = %d, want %d", capturedBody.WalltimeMs, tt.walltimeMs)
			}
		})
	}
}

// TestGetBroadcasts_ErrorPropagation verifies errors from API are returned.
// Kills mutant: `return nil, err` → `_ = err`
func TestGetBroadcasts_ErrorPropagation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": 500, "message": "Internal Server Error"},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	result, err := GetBroadcasts(context.Background(), client, &GetBroadcastsParams{
		IDs: []string{"b1"},
	})

	if err == nil {
		t.Fatal("expected error from failed API call, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
}

// TestGetBroadcast_ErrorPropagation verifies GetBroadcast propagates errors.
// Kills mutant: `return nil, err` → `_ = err`
func TestGetBroadcast_ErrorPropagation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": 500, "message": "Internal Server Error"},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	result, err := GetBroadcast(context.Background(), client, "b1")

	if err == nil {
		t.Fatal("expected error from failed API call, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
}

// TestInsertBroadcast_ErrorPropagation verifies InsertBroadcast propagates errors.
// Kills mutant: `return nil, err` → `_ = err`
func TestInsertBroadcast_ErrorPropagation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": 400, "message": "Bad Request"},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	scheduledTime := time.Now().Add(1 * time.Hour)
	result, err := InsertBroadcast(context.Background(), client, &LiveBroadcast{
		Snippet: &BroadcastSnippet{Title: "Test", ScheduledStartTime: &scheduledTime},
		Status:  &BroadcastStatus{PrivacyStatus: "private"},
	})

	if err == nil {
		t.Fatal("expected error from failed API call, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
}

// TestUpdateBroadcast_ErrorPropagation verifies UpdateBroadcast propagates errors.
// Kills mutant: `return nil, err` → `_ = err`
func TestUpdateBroadcast_ErrorPropagation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": 403, "message": "Forbidden"},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	result, err := UpdateBroadcast(context.Background(), client, &LiveBroadcast{
		ID: "b1",
	})

	if err == nil {
		t.Fatal("expected error from failed API call, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
}

// TestDeleteBroadcast_ErrorPropagation verifies DeleteBroadcast propagates errors.
func TestDeleteBroadcast_ErrorPropagation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": 404, "message": "Not Found"},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	err := DeleteBroadcast(context.Background(), client, "b1")

	if err == nil {
		t.Fatal("expected error from failed API call, got nil")
	}
}

// TestBindBroadcast_ErrorPropagation verifies BindBroadcast propagates errors.
// Kills mutant: `return nil, err` → `_ = err`
func TestBindBroadcast_ErrorPropagation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": 400, "message": "Bad Request"},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	result, err := BindBroadcast(context.Background(), client, &BindBroadcastParams{
		BroadcastID: "b1",
		StreamID:    "s1",
	})

	if err == nil {
		t.Fatal("expected error from failed API call, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
}

// TestTransitionBroadcast_ErrorPropagation verifies TransitionBroadcast propagates errors.
// Kills mutant: `return nil, err` → `_ = err`
func TestTransitionBroadcast_ErrorPropagation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": 409, "message": "Conflict"},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	result, err := TransitionBroadcast(context.Background(), client, "b1", TransitionLive)

	if err == nil {
		t.Fatal("expected error from failed API call, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
}

// TestInsertCuepoint_ErrorPropagation verifies InsertCuepoint propagates errors.
// Kills mutant: `return nil, err` → `_ = err`
func TestInsertCuepoint_ErrorPropagation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": 400, "message": "Bad Request"},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	result, err := InsertCuepoint(context.Background(), client, &InsertCuepointParams{
		BroadcastID: "b1",
	})

	if err == nil {
		t.Fatal("expected error from failed API call, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
}

// TestGetBroadcasts_ValidationErrors tests validation paths return errors.
// Kills mutant: `return nil, fmt.Errorf(...)` → `_ = fmt.Errorf`
func TestGetBroadcasts_ValidationErrors(t *testing.T) {
	client := core.NewClient()

	t.Run("nil params", func(t *testing.T) {
		result, err := GetBroadcasts(context.Background(), client, nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
		if result != nil {
			t.Errorf("expected nil result on validation error, got %+v", result)
		}
	})

	t.Run("empty IDs no mine", func(t *testing.T) {
		result, err := GetBroadcasts(context.Background(), client, &GetBroadcastsParams{
			IDs:  []string{},
			Mine: false,
		})
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
		if result != nil {
			t.Errorf("expected nil result on validation error, got %+v", result)
		}
	})
}

// TestGetBroadcast_ValidationError_EmptyID tests empty broadcast ID returns error.
// Kills mutant: `return nil, fmt.Errorf(...)` → `_ = fmt.Errorf`
func TestGetBroadcast_ValidationError_EmptyID(t *testing.T) {
	client := core.NewClient()
	result, err := GetBroadcast(context.Background(), client, "")

	if err == nil {
		t.Fatal("expected validation error for empty broadcast ID")
	}
	if result != nil {
		t.Errorf("expected nil result on validation error, got %+v", result)
	}
}

// TestInsertBroadcast_ValidationErrors tests all validation paths return errors.
// Kills mutant: `return nil, fmt.Errorf(...)` → `_ = fmt.Errorf`
func TestInsertBroadcast_ValidationErrors(t *testing.T) {
	client := core.NewClient()

	tests := []struct {
		name      string
		broadcast *LiveBroadcast
	}{
		{"nil broadcast", nil},
		{"nil snippet", &LiveBroadcast{Status: &BroadcastStatus{}}},
		{"empty title", &LiveBroadcast{Snippet: &BroadcastSnippet{Title: ""}, Status: &BroadcastStatus{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := InsertBroadcast(context.Background(), client, tt.broadcast)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if result != nil {
				t.Errorf("expected nil result on validation error, got %+v", result)
			}
		})
	}
}

// TestUpdateBroadcast_ValidationErrors tests all validation paths return errors.
// Kills mutant: `return nil, fmt.Errorf(...)` → `_ = fmt.Errorf`
func TestUpdateBroadcast_ValidationErrors(t *testing.T) {
	client := core.NewClient()

	tests := []struct {
		name      string
		broadcast *LiveBroadcast
	}{
		{"nil broadcast", nil},
		{"missing ID", &LiveBroadcast{Snippet: &BroadcastSnippet{Title: "Test"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := UpdateBroadcast(context.Background(), client, tt.broadcast)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if result != nil {
				t.Errorf("expected nil result on validation error, got %+v", result)
			}
		})
	}
}

// TestDeleteBroadcast_ValidationError_EmptyID tests empty broadcast ID returns error.
// Kills mutant: `return fmt.Errorf(...)` → `_ = fmt.Errorf`
func TestDeleteBroadcast_ValidationError_EmptyID(t *testing.T) {
	client := core.NewClient()
	err := DeleteBroadcast(context.Background(), client, "")

	if err == nil {
		t.Fatal("expected validation error for empty broadcast ID")
	}
}

// TestBindBroadcast_ValidationError_EmptyBroadcastID tests empty broadcast ID returns error.
// Kills mutant: `return nil, fmt.Errorf(...)` → `_ = fmt.Errorf`
func TestBindBroadcast_ValidationError_EmptyBroadcastID(t *testing.T) {
	client := core.NewClient()
	result, err := BindBroadcast(context.Background(), client, &BindBroadcastParams{
		BroadcastID: "",
		StreamID:    "s1",
	})

	if err == nil {
		t.Fatal("expected validation error for empty broadcast ID")
	}
	if result != nil {
		t.Errorf("expected nil result on validation error, got %+v", result)
	}
}

// TestTransitionBroadcast_ValidationErrors tests validation paths return errors.
// Kills mutant: `return nil, fmt.Errorf(...)` → `_ = fmt.Errorf`
func TestTransitionBroadcast_ValidationErrors(t *testing.T) {
	client := core.NewClient()

	t.Run("empty broadcast ID", func(t *testing.T) {
		result, err := TransitionBroadcast(context.Background(), client, "", TransitionLive)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
		if result != nil {
			t.Errorf("expected nil result on validation error, got %+v", result)
		}
	})

	t.Run("empty status", func(t *testing.T) {
		result, err := TransitionBroadcast(context.Background(), client, "b1", "")
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
		if result != nil {
			t.Errorf("expected nil result on validation error, got %+v", result)
		}
	})
}

// TestInsertCuepoint_ValidationError_EmptyBroadcastID tests empty broadcast ID returns error.
// Kills mutant: `return nil, fmt.Errorf(...)` → `_ = fmt.Errorf`
func TestInsertCuepoint_ValidationError_EmptyBroadcastID(t *testing.T) {
	client := core.NewClient()
	result, err := InsertCuepoint(context.Background(), client, &InsertCuepointParams{
		BroadcastID: "",
	})

	if err == nil {
		t.Fatal("expected validation error for empty broadcast ID")
	}
	if result != nil {
		t.Errorf("expected nil result on validation error, got %+v", result)
	}
}
