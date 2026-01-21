package data

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Its-donkey/yougopher/youtube/core"
)

func TestGetPlaylists(t *testing.T) {
	t.Run("success with IDs", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/playlists" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("id") != "playlist123" {
				t.Errorf("unexpected id: %s", r.URL.Query().Get("id"))
			}

			resp := PlaylistListResponse{
				Items: []*Playlist{{
					ID:      "playlist123",
					Snippet: &PlaylistSnippet{Title: "Test Playlist"},
				}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		resp, err := GetPlaylists(context.Background(), client, &GetPlaylistsParams{
			IDs: []string{"playlist123"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(resp.Items))
		}
		if resp.Items[0].ID != "playlist123" {
			t.Errorf("unexpected ID: %s", resp.Items[0].ID)
		}
	})

	t.Run("nil params", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetPlaylists(context.Background(), client, nil)
		if err == nil {
			t.Fatal("expected error for nil params")
		}
	})

	t.Run("no filter provided", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetPlaylists(context.Background(), client, &GetPlaylistsParams{})
		if err == nil {
			t.Fatal("expected error for no filter")
		}
	})

	t.Run("with channelId", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("channelId") != "channel123" {
				t.Errorf("unexpected channelId: %s", r.URL.Query().Get("channelId"))
			}
			resp := PlaylistListResponse{Items: []*Playlist{{ID: "playlist123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetPlaylists(context.Background(), client, &GetPlaylistsParams{
			ChannelID: "channel123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with mine", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("mine") != "true" {
				t.Errorf("unexpected mine: %s", r.URL.Query().Get("mine"))
			}
			resp := PlaylistListResponse{Items: []*Playlist{{ID: "playlist123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetPlaylists(context.Background(), client, &GetPlaylistsParams{
			Mine: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with custom parts and pagination", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "snippet,status" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}
			if r.URL.Query().Get("maxResults") != "10" {
				t.Errorf("unexpected maxResults: %s", r.URL.Query().Get("maxResults"))
			}
			if r.URL.Query().Get("pageToken") != "nextPage" {
				t.Errorf("unexpected pageToken: %s", r.URL.Query().Get("pageToken"))
			}
			resp := PlaylistListResponse{Items: []*Playlist{{ID: "playlist123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetPlaylists(context.Background(), client, &GetPlaylistsParams{
			IDs:        []string{"playlist123"},
			Parts:      []string{"snippet", "status"},
			MaxResults: 10,
			PageToken:  "nextPage",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetPlaylist(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := PlaylistListResponse{
				Items: []*Playlist{{ID: "playlist123", Snippet: &PlaylistSnippet{Title: "Test"}}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		playlist, err := GetPlaylist(context.Background(), client, "playlist123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if playlist.ID != "playlist123" {
			t.Errorf("unexpected ID: %s", playlist.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := PlaylistListResponse{Items: []*Playlist{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetPlaylist(context.Background(), client, "nonexistent")
		if err == nil {
			t.Fatal("expected error for not found playlist")
		}
		notFoundErr, ok := err.(*core.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T", err)
		}
		if notFoundErr.ResourceType != "playlist" {
			t.Errorf("unexpected resource type: %s", notFoundErr.ResourceType)
		}
	})

	t.Run("empty playlist ID", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetPlaylist(context.Background(), client, "")
		if err == nil {
			t.Fatal("expected error for empty playlist ID")
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "snippet" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}
			resp := PlaylistListResponse{Items: []*Playlist{{ID: "playlist123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetPlaylist(context.Background(), client, "playlist123", "snippet")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetMyPlaylists(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("mine") != "true" {
				t.Errorf("expected mine=true")
			}
			resp := PlaylistListResponse{Items: []*Playlist{{ID: "myPlaylist"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		resp, err := GetMyPlaylists(context.Background(), client, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Items) != 1 || resp.Items[0].ID != "myPlaylist" {
			t.Errorf("unexpected response")
		}
	})

	t.Run("with params", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("maxResults") != "5" {
				t.Errorf("unexpected maxResults: %s", r.URL.Query().Get("maxResults"))
			}
			resp := PlaylistListResponse{Items: []*Playlist{{ID: "myPlaylist"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetMyPlaylists(context.Background(), client, &GetPlaylistsParams{MaxResults: 5})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetPlaylistItems(t *testing.T) {
	t.Run("success with playlistId", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/playlistItems" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("playlistId") != "playlist123" {
				t.Errorf("unexpected playlistId: %s", r.URL.Query().Get("playlistId"))
			}

			resp := PlaylistItemListResponse{
				Items: []*PlaylistItem{{
					ID:      "item123",
					Snippet: &PlaylistItemSnippet{Title: "Video Title"},
					ContentDetails: &PlaylistItemContentDetails{
						VideoID: "video123",
					},
				}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		resp, err := GetPlaylistItems(context.Background(), client, &GetPlaylistItemsParams{
			PlaylistID: "playlist123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(resp.Items))
		}
		if resp.Items[0].VideoID() != "video123" {
			t.Errorf("unexpected video ID: %s", resp.Items[0].VideoID())
		}
	})

	t.Run("nil params", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetPlaylistItems(context.Background(), client, nil)
		if err == nil {
			t.Fatal("expected error for nil params")
		}
	})

	t.Run("no filter provided", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetPlaylistItems(context.Background(), client, &GetPlaylistItemsParams{})
		if err == nil {
			t.Fatal("expected error for no filter")
		}
	})

	t.Run("with IDs", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("id") != "item1,item2" {
				t.Errorf("unexpected id: %s", r.URL.Query().Get("id"))
			}
			resp := PlaylistItemListResponse{Items: []*PlaylistItem{{ID: "item1"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetPlaylistItems(context.Background(), client, &GetPlaylistItemsParams{
			IDs: []string{"item1", "item2"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with pagination", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("maxResults") != "25" {
				t.Errorf("unexpected maxResults: %s", r.URL.Query().Get("maxResults"))
			}
			if r.URL.Query().Get("pageToken") != "nextPage" {
				t.Errorf("unexpected pageToken: %s", r.URL.Query().Get("pageToken"))
			}
			resp := PlaylistItemListResponse{Items: []*PlaylistItem{{ID: "item1"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetPlaylistItems(context.Background(), client, &GetPlaylistItemsParams{
			PlaylistID: "playlist123",
			MaxResults: 25,
			PageToken:  "nextPage",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "snippet" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}
			resp := PlaylistItemListResponse{Items: []*PlaylistItem{{ID: "item1"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetPlaylistItems(context.Background(), client, &GetPlaylistItemsParams{
			PlaylistID: "playlist123",
			Parts:      []string{"snippet"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestPlaylistItem_VideoID(t *testing.T) {
	tests := []struct {
		name string
		item *PlaylistItem
		want string
	}{
		{"from contentDetails", &PlaylistItem{
			ContentDetails: &PlaylistItemContentDetails{VideoID: "video123"},
		}, "video123"},
		{"from resourceId", &PlaylistItem{
			Snippet: &PlaylistItemSnippet{ResourceID: &ResourceID{VideoID: "video456"}},
		}, "video456"},
		{"contentDetails priority", &PlaylistItem{
			ContentDetails: &PlaylistItemContentDetails{VideoID: "video123"},
			Snippet:        &PlaylistItemSnippet{ResourceID: &ResourceID{VideoID: "video456"}},
		}, "video123"},
		{"empty", &PlaylistItem{}, ""},
		{"nil resourceId", &PlaylistItem{Snippet: &PlaylistItemSnippet{}}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.VideoID(); got != tt.want {
				t.Errorf("VideoID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPlaylistListResponse_JSON(t *testing.T) {
	jsonData := `{
		"kind": "youtube#playlistListResponse",
		"etag": "abc123",
		"nextPageToken": "page2",
		"pageInfo": {"totalResults": 10, "resultsPerPage": 5},
		"items": [
			{
				"id": "playlist123",
				"snippet": {
					"title": "Test Playlist",
					"description": "A test playlist",
					"channelId": "channel123"
				},
				"contentDetails": {
					"itemCount": 25
				}
			}
		]
	}`

	var resp PlaylistListResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if resp.NextPageToken != "page2" {
		t.Errorf("NextPageToken = %q, want 'page2'", resp.NextPageToken)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(resp.Items))
	}
	if resp.Items[0].Snippet.Title != "Test Playlist" {
		t.Errorf("Title = %q, want 'Test Playlist'", resp.Items[0].Snippet.Title)
	}
	if resp.Items[0].ContentDetails.ItemCount != 25 {
		t.Errorf("ItemCount = %d, want 25", resp.Items[0].ContentDetails.ItemCount)
	}
}
