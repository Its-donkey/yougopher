package streaming

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
