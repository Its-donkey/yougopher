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

func TestSearch(t *testing.T) {
	t.Run("success with query", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/search" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("q") != "golang tutorial" {
				t.Errorf("unexpected q: %s", r.URL.Query().Get("q"))
			}

			resp := SearchListResponse{
				Items: []*SearchResult{{
					ID: &SearchResultID{
						Kind:    "youtube#video",
						VideoID: "video123",
					},
					Snippet: &SearchResultSnippet{
						Title: "Golang Tutorial",
					},
				}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		resp, err := Search(context.Background(), client, &SearchParams{
			Query: "golang tutorial",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(resp.Items))
		}
		if resp.Items[0].ID.VideoID != "video123" {
			t.Errorf("unexpected video ID: %s", resp.Items[0].ID.VideoID)
		}
	})

	t.Run("nil params", func(t *testing.T) {
		client := core.NewClient()
		_, err := Search(context.Background(), client, nil)
		if err == nil {
			t.Fatal("expected error for nil params")
		}
	})

	t.Run("with all parameters", func(t *testing.T) {
		pubAfter := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		pubBefore := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if q.Get("type") != "video" {
				t.Errorf("unexpected type: %s", q.Get("type"))
			}
			if q.Get("channelId") != "channel123" {
				t.Errorf("unexpected channelId: %s", q.Get("channelId"))
			}
			if q.Get("eventType") != "live" {
				t.Errorf("unexpected eventType: %s", q.Get("eventType"))
			}
			if q.Get("order") != "date" {
				t.Errorf("unexpected order: %s", q.Get("order"))
			}
			if q.Get("regionCode") != "US" {
				t.Errorf("unexpected regionCode: %s", q.Get("regionCode"))
			}
			if q.Get("relevanceLanguage") != "en" {
				t.Errorf("unexpected relevanceLanguage: %s", q.Get("relevanceLanguage"))
			}
			if q.Get("safeSearch") != "moderate" {
				t.Errorf("unexpected safeSearch: %s", q.Get("safeSearch"))
			}
			if q.Get("videoCategoryId") != "10" {
				t.Errorf("unexpected videoCategoryId: %s", q.Get("videoCategoryId"))
			}
			if q.Get("videoDefinition") != "high" {
				t.Errorf("unexpected videoDefinition: %s", q.Get("videoDefinition"))
			}
			if q.Get("videoDuration") != "medium" {
				t.Errorf("unexpected videoDuration: %s", q.Get("videoDuration"))
			}
			if q.Get("videoType") != "episode" {
				t.Errorf("unexpected videoType: %s", q.Get("videoType"))
			}
			if q.Get("maxResults") != "25" {
				t.Errorf("unexpected maxResults: %s", q.Get("maxResults"))
			}
			if q.Get("pageToken") != "nextPage" {
				t.Errorf("unexpected pageToken: %s", q.Get("pageToken"))
			}
			if q.Get("publishedAfter") == "" {
				t.Error("publishedAfter should be set")
			}
			if q.Get("publishedBefore") == "" {
				t.Error("publishedBefore should be set")
			}

			resp := SearchListResponse{Items: []*SearchResult{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := Search(context.Background(), client, &SearchParams{
			Query:             "test",
			Type:              SearchTypeVideo,
			ChannelID:         "channel123",
			EventType:         SearchEventTypeLive,
			Order:             SearchOrderDate,
			PublishedAfter:    &pubAfter,
			PublishedBefore:   &pubBefore,
			RegionCode:        "US",
			RelevanceLanguage: "en",
			SafeSearch:        "moderate",
			VideoCategoryID:   "10",
			VideoDefinition:   "high",
			VideoDuration:     "medium",
			VideoType:         "episode",
			MaxResults:        25,
			PageToken:         "nextPage",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "id" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}
			resp := SearchListResponse{Items: []*SearchResult{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := Search(context.Background(), client, &SearchParams{
			Query: "test",
			Parts: []string{"id"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestSearchVideos(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("type") != "video" {
			t.Errorf("unexpected type: %s", r.URL.Query().Get("type"))
		}
		if r.URL.Query().Get("q") != "test query" {
			t.Errorf("unexpected q: %s", r.URL.Query().Get("q"))
		}
		if r.URL.Query().Get("maxResults") != "10" {
			t.Errorf("unexpected maxResults: %s", r.URL.Query().Get("maxResults"))
		}
		resp := SearchListResponse{Items: []*SearchResult{{ID: &SearchResultID{VideoID: "v123"}}}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	resp, err := SearchVideos(context.Background(), client, "test query", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Items) != 1 || resp.Items[0].ID.VideoID != "v123" {
		t.Error("unexpected response")
	}
}

func TestSearchLiveStreams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("type") != "video" {
			t.Errorf("unexpected type: %s", r.URL.Query().Get("type"))
		}
		if r.URL.Query().Get("eventType") != "live" {
			t.Errorf("unexpected eventType: %s", r.URL.Query().Get("eventType"))
		}
		resp := SearchListResponse{Items: []*SearchResult{}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	_, err := SearchLiveStreams(context.Background(), client, "gaming", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSearchChannels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("type") != "channel" {
			t.Errorf("unexpected type: %s", r.URL.Query().Get("type"))
		}
		resp := SearchListResponse{Items: []*SearchResult{{ID: &SearchResultID{ChannelID: "c123"}}}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	resp, err := SearchChannels(context.Background(), client, "tech channel", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Items) != 1 || resp.Items[0].ID.ChannelID != "c123" {
		t.Error("unexpected response")
	}
}

func TestSearchResult_Methods(t *testing.T) {
	t.Run("IsVideo", func(t *testing.T) {
		tests := []struct {
			name   string
			result *SearchResult
			want   bool
		}{
			{"nil ID", &SearchResult{}, false},
			{"video", &SearchResult{ID: &SearchResultID{VideoID: "v123"}}, true},
			{"channel", &SearchResult{ID: &SearchResultID{ChannelID: "c123"}}, false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.result.IsVideo(); got != tt.want {
					t.Errorf("IsVideo() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("IsChannel", func(t *testing.T) {
		tests := []struct {
			name   string
			result *SearchResult
			want   bool
		}{
			{"nil ID", &SearchResult{}, false},
			{"channel", &SearchResult{ID: &SearchResultID{ChannelID: "c123"}}, true},
			{"video", &SearchResult{ID: &SearchResultID{VideoID: "v123"}}, false},
			{"both", &SearchResult{ID: &SearchResultID{VideoID: "v", ChannelID: "c"}}, false}, // video takes precedence
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.result.IsChannel(); got != tt.want {
					t.Errorf("IsChannel() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("IsPlaylist", func(t *testing.T) {
		tests := []struct {
			name   string
			result *SearchResult
			want   bool
		}{
			{"nil ID", &SearchResult{}, false},
			{"playlist", &SearchResult{ID: &SearchResultID{PlaylistID: "p123"}}, true},
			{"video", &SearchResult{ID: &SearchResultID{VideoID: "v123"}}, false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.result.IsPlaylist(); got != tt.want {
					t.Errorf("IsPlaylist() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("IsLive", func(t *testing.T) {
		tests := []struct {
			name   string
			result *SearchResult
			want   bool
		}{
			{"nil snippet", &SearchResult{}, false},
			{"live", &SearchResult{Snippet: &SearchResultSnippet{LiveBroadcastContent: "live"}}, true},
			{"upcoming", &SearchResult{Snippet: &SearchResultSnippet{LiveBroadcastContent: "upcoming"}}, false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.result.IsLive(); got != tt.want {
					t.Errorf("IsLive() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("IsUpcoming", func(t *testing.T) {
		tests := []struct {
			name   string
			result *SearchResult
			want   bool
		}{
			{"nil snippet", &SearchResult{}, false},
			{"upcoming", &SearchResult{Snippet: &SearchResultSnippet{LiveBroadcastContent: "upcoming"}}, true},
			{"live", &SearchResult{Snippet: &SearchResultSnippet{LiveBroadcastContent: "live"}}, false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.result.IsUpcoming(); got != tt.want {
					t.Errorf("IsUpcoming() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("ResourceID", func(t *testing.T) {
		tests := []struct {
			name   string
			result *SearchResult
			want   string
		}{
			{"nil ID", &SearchResult{}, ""},
			{"video", &SearchResult{ID: &SearchResultID{VideoID: "v123"}}, "v123"},
			{"channel", &SearchResult{ID: &SearchResultID{ChannelID: "c123"}}, "c123"},
			{"playlist", &SearchResult{ID: &SearchResultID{PlaylistID: "p123"}}, "p123"},
			{"video priority", &SearchResult{ID: &SearchResultID{VideoID: "v", ChannelID: "c"}}, "v"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.result.ResourceID(); got != tt.want {
					t.Errorf("ResourceID() = %q, want %q", got, tt.want)
				}
			})
		}
	})
}

func TestSearchConstants(t *testing.T) {
	// Verify type constants
	if SearchTypeVideo != "video" {
		t.Errorf("SearchTypeVideo = %q, want 'video'", SearchTypeVideo)
	}
	if SearchTypeChannel != "channel" {
		t.Errorf("SearchTypeChannel = %q, want 'channel'", SearchTypeChannel)
	}
	if SearchTypePlaylist != "playlist" {
		t.Errorf("SearchTypePlaylist = %q, want 'playlist'", SearchTypePlaylist)
	}

	// Verify event type constants
	if SearchEventTypeLive != "live" {
		t.Errorf("SearchEventTypeLive = %q, want 'live'", SearchEventTypeLive)
	}
	if SearchEventTypeUpcoming != "upcoming" {
		t.Errorf("SearchEventTypeUpcoming = %q, want 'upcoming'", SearchEventTypeUpcoming)
	}

	// Verify order constants
	if SearchOrderRelevance != "relevance" {
		t.Errorf("SearchOrderRelevance = %q, want 'relevance'", SearchOrderRelevance)
	}
	if SearchOrderDate != "date" {
		t.Errorf("SearchOrderDate = %q, want 'date'", SearchOrderDate)
	}
}

func TestSearchListResponse_JSON(t *testing.T) {
	jsonData := `{
		"kind": "youtube#searchListResponse",
		"etag": "abc123",
		"nextPageToken": "page2",
		"regionCode": "US",
		"pageInfo": {"totalResults": 100, "resultsPerPage": 25},
		"items": [
			{
				"id": {"kind": "youtube#video", "videoId": "video123"},
				"snippet": {
					"title": "Test Video",
					"channelId": "channel123",
					"liveBroadcastContent": "live"
				}
			}
		]
	}`

	var resp SearchListResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if resp.NextPageToken != "page2" {
		t.Errorf("NextPageToken = %q, want 'page2'", resp.NextPageToken)
	}
	if resp.RegionCode != "US" {
		t.Errorf("RegionCode = %q, want 'US'", resp.RegionCode)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(resp.Items))
	}
	if !resp.Items[0].IsVideo() {
		t.Error("expected IsVideo() = true")
	}
	if !resp.Items[0].IsLive() {
		t.Error("expected IsLive() = true")
	}
}
