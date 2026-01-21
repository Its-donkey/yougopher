package data

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Its-donkey/yougopher/youtube/core"
)

func TestGetVideos(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/videos" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("id") != "video123" {
				t.Errorf("unexpected id: %s", r.URL.Query().Get("id"))
			}
			if r.URL.Query().Get("part") != "snippet,liveStreamingDetails" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}

			resp := VideoListResponse{
				Kind: "youtube#videoListResponse",
				Items: []*Video{
					{
						ID:   "video123",
						Kind: "youtube#video",
						Snippet: &VideoSnippet{
							Title:                "Test Video",
							LiveBroadcastContent: "live",
						},
						LiveStreamingDetails: &LiveStreamingDetails{
							ActiveLiveChatID: "chat123",
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		resp, err := GetVideos(context.Background(), client, &GetVideosParams{
			IDs: []string{"video123"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(resp.Items))
		}
		if resp.Items[0].ID != "video123" {
			t.Errorf("unexpected ID: %s", resp.Items[0].ID)
		}
		if resp.Items[0].Snippet.Title != "Test Video" {
			t.Errorf("unexpected title: %s", resp.Items[0].Snippet.Title)
		}
		if resp.Items[0].LiveStreamingDetails.ActiveLiveChatID != "chat123" {
			t.Errorf("unexpected live chat ID: %s", resp.Items[0].LiveStreamingDetails.ActiveLiveChatID)
		}
	})

	t.Run("nil params", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetVideos(context.Background(), client, nil)
		if err == nil {
			t.Fatal("expected error for nil params")
		}
	})

	t.Run("no IDs", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetVideos(context.Background(), client, &GetVideosParams{})
		if err == nil {
			t.Fatal("expected error for no IDs")
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "snippet,statistics" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}

			resp := VideoListResponse{Items: []*Video{{ID: "video123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetVideos(context.Background(), client, &GetVideosParams{
			IDs:   []string{"video123"},
			Parts: []string{"snippet", "statistics"},
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

			resp := VideoListResponse{Items: []*Video{{ID: "video123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetVideos(context.Background(), client, &GetVideosParams{
			IDs:        []string{"video123"},
			MaxResults: 10,
			PageToken:  "nextPage123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetVideo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := VideoListResponse{
				Items: []*Video{{ID: "video123", Snippet: &VideoSnippet{Title: "Test"}}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		video, err := GetVideo(context.Background(), client, "video123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if video.ID != "video123" {
			t.Errorf("unexpected ID: %s", video.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := VideoListResponse{Items: []*Video{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetVideo(context.Background(), client, "nonexistent")
		if err == nil {
			t.Fatal("expected error for not found video")
		}
		notFoundErr, ok := err.(*core.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T", err)
		}
		if notFoundErr.ResourceType != "video" {
			t.Errorf("unexpected resource type: %s", notFoundErr.ResourceType)
		}
	})

	t.Run("empty video ID", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetVideo(context.Background(), client, "")
		if err == nil {
			t.Fatal("expected error for empty video ID")
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "snippet" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}
			resp := VideoListResponse{Items: []*Video{{ID: "video123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetVideo(context.Background(), client, "video123", "snippet")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetLiveChatID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := VideoListResponse{
				Items: []*Video{{
					ID: "video123",
					LiveStreamingDetails: &LiveStreamingDetails{
						ActiveLiveChatID: "chat123",
					},
				}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		chatID, err := GetLiveChatID(context.Background(), client, "video123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if chatID != "chat123" {
			t.Errorf("unexpected chat ID: %s", chatID)
		}
	})

	t.Run("no live streaming details", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := VideoListResponse{Items: []*Video{{ID: "video123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetLiveChatID(context.Background(), client, "video123")
		if err == nil {
			t.Fatal("expected error for no live streaming details")
		}
	})

	t.Run("no active live chat", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := VideoListResponse{
				Items: []*Video{{
					ID:                   "video123",
					LiveStreamingDetails: &LiveStreamingDetails{},
				}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetLiveChatID(context.Background(), client, "video123")
		if err == nil {
			t.Fatal("expected error for no active live chat")
		}
	})
}

func TestVideo_Methods(t *testing.T) {
	t.Run("IsLive", func(t *testing.T) {
		tests := []struct {
			name    string
			video   *Video
			want    bool
		}{
			{"nil snippet", &Video{}, false},
			{"not live", &Video{Snippet: &VideoSnippet{LiveBroadcastContent: "none"}}, false},
			{"live", &Video{Snippet: &VideoSnippet{LiveBroadcastContent: "live"}}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.video.IsLive(); got != tt.want {
					t.Errorf("IsLive() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("IsUpcoming", func(t *testing.T) {
		tests := []struct {
			name    string
			video   *Video
			want    bool
		}{
			{"nil snippet", &Video{}, false},
			{"not upcoming", &Video{Snippet: &VideoSnippet{LiveBroadcastContent: "live"}}, false},
			{"upcoming", &Video{Snippet: &VideoSnippet{LiveBroadcastContent: "upcoming"}}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.video.IsUpcoming(); got != tt.want {
					t.Errorf("IsUpcoming() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("HasActiveLiveChat", func(t *testing.T) {
		tests := []struct {
			name    string
			video   *Video
			want    bool
		}{
			{"nil details", &Video{}, false},
			{"no chat ID", &Video{LiveStreamingDetails: &LiveStreamingDetails{}}, false},
			{"has chat ID", &Video{LiveStreamingDetails: &LiveStreamingDetails{ActiveLiveChatID: "chat123"}}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.video.HasActiveLiveChat(); got != tt.want {
					t.Errorf("HasActiveLiveChat() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}

func TestVideoListResponse_JSON(t *testing.T) {
	jsonData := `{
		"kind": "youtube#videoListResponse",
		"etag": "abc123",
		"nextPageToken": "page2",
		"pageInfo": {
			"totalResults": 100,
			"resultsPerPage": 50
		},
		"items": [
			{
				"id": "video123",
				"snippet": {
					"publishedAt": "2024-01-15T10:30:00Z",
					"channelId": "channel123",
					"title": "Test Video",
					"description": "A test video",
					"liveBroadcastContent": "live"
				},
				"liveStreamingDetails": {
					"activeLiveChatId": "chat123",
					"concurrentViewers": "1000"
				}
			}
		]
	}`

	var resp VideoListResponse
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
	if resp.Items[0].Snippet.Title != "Test Video" {
		t.Errorf("Title = %q, want 'Test Video'", resp.Items[0].Snippet.Title)
	}
	if resp.Items[0].LiveStreamingDetails.ActiveLiveChatID != "chat123" {
		t.Errorf("ActiveLiveChatID = %q, want 'chat123'", resp.Items[0].LiveStreamingDetails.ActiveLiveChatID)
	}
}

func TestLiveStreamingDetails_JSON(t *testing.T) {
	startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	jsonData := `{
		"actualStartTime": "2024-01-15T10:00:00Z",
		"actualEndTime": "2024-01-15T12:00:00Z",
		"scheduledStartTime": "2024-01-15T10:00:00Z",
		"concurrentViewers": "5000",
		"activeLiveChatId": "chatABC"
	}`

	var details LiveStreamingDetails
	err := json.Unmarshal([]byte(jsonData), &details)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if details.ActualStartTime == nil || !details.ActualStartTime.Equal(startTime) {
		t.Errorf("ActualStartTime = %v, want %v", details.ActualStartTime, startTime)
	}
	if details.ActualEndTime == nil || !details.ActualEndTime.Equal(endTime) {
		t.Errorf("ActualEndTime = %v, want %v", details.ActualEndTime, endTime)
	}
	if details.ConcurrentViewers != "5000" {
		t.Errorf("ConcurrentViewers = %q, want '5000'", details.ConcurrentViewers)
	}
	if details.ActiveLiveChatID != "chatABC" {
		t.Errorf("ActiveLiveChatID = %q, want 'chatABC'", details.ActiveLiveChatID)
	}
}
