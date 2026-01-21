package data

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Its-donkey/yougopher/youtube/core"
)

func TestGetChannels(t *testing.T) {
	t.Run("success with IDs", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/channels" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("id") != "channel123" {
				t.Errorf("unexpected id: %s", r.URL.Query().Get("id"))
			}
			if r.URL.Query().Get("part") != "snippet,statistics" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}

			resp := ChannelListResponse{
				Kind: "youtube#channelListResponse",
				Items: []*Channel{
					{
						ID:   "channel123",
						Kind: "youtube#channel",
						Snippet: &ChannelSnippet{
							Title:       "Test Channel",
							Description: "A test channel",
						},
						Statistics: &ChannelStatistics{
							SubscriberCount: "1000",
							VideoCount:      "50",
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		resp, err := GetChannels(context.Background(), client, &GetChannelsParams{
			IDs: []string{"channel123"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(resp.Items))
		}
		if resp.Items[0].ID != "channel123" {
			t.Errorf("unexpected ID: %s", resp.Items[0].ID)
		}
		if resp.Items[0].Snippet.Title != "Test Channel" {
			t.Errorf("unexpected title: %s", resp.Items[0].Snippet.Title)
		}
	})

	t.Run("nil params", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetChannels(context.Background(), client, nil)
		if err == nil {
			t.Fatal("expected error for nil params")
		}
	})

	t.Run("no filter provided", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetChannels(context.Background(), client, &GetChannelsParams{})
		if err == nil {
			t.Fatal("expected error for no filter")
		}
	})

	t.Run("with forUsername", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("forUsername") != "testuser" {
				t.Errorf("unexpected forUsername: %s", r.URL.Query().Get("forUsername"))
			}

			resp := ChannelListResponse{Items: []*Channel{{ID: "channel123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetChannels(context.Background(), client, &GetChannelsParams{
			ForUsername: "testuser",
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

			resp := ChannelListResponse{Items: []*Channel{{ID: "channel123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetChannels(context.Background(), client, &GetChannelsParams{
			Mine: true,
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

			resp := ChannelListResponse{Items: []*Channel{{ID: "channel123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetChannels(context.Background(), client, &GetChannelsParams{
			IDs:   []string{"channel123"},
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

			resp := ChannelListResponse{Items: []*Channel{{ID: "channel123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetChannels(context.Background(), client, &GetChannelsParams{
			IDs:        []string{"channel123"},
			MaxResults: 10,
			PageToken:  "nextPage123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetChannel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := ChannelListResponse{
				Items: []*Channel{{ID: "channel123", Snippet: &ChannelSnippet{Title: "Test"}}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		channel, err := GetChannel(context.Background(), client, "channel123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if channel.ID != "channel123" {
			t.Errorf("unexpected ID: %s", channel.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := ChannelListResponse{Items: []*Channel{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetChannel(context.Background(), client, "nonexistent")
		if err == nil {
			t.Fatal("expected error for not found channel")
		}
		notFoundErr, ok := err.(*core.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T", err)
		}
		if notFoundErr.ResourceType != "channel" {
			t.Errorf("unexpected resource type: %s", notFoundErr.ResourceType)
		}
	})

	t.Run("empty channel ID", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetChannel(context.Background(), client, "")
		if err == nil {
			t.Fatal("expected error for empty channel ID")
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "snippet" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}
			resp := ChannelListResponse{Items: []*Channel{{ID: "channel123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetChannel(context.Background(), client, "channel123", "snippet")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetMyChannel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("mine") != "true" {
				t.Errorf("expected mine=true, got %s", r.URL.Query().Get("mine"))
			}

			resp := ChannelListResponse{
				Items: []*Channel{{ID: "myChannel", Snippet: &ChannelSnippet{Title: "My Channel"}}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		channel, err := GetMyChannel(context.Background(), client)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if channel.ID != "myChannel" {
			t.Errorf("unexpected ID: %s", channel.ID)
		}
	})

	t.Run("no channel found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := ChannelListResponse{Items: []*Channel{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetMyChannel(context.Background(), client)
		if err == nil {
			t.Fatal("expected error for no channel found")
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "contentDetails" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}
			resp := ChannelListResponse{Items: []*Channel{{ID: "myChannel"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetMyChannel(context.Background(), client, "contentDetails")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestChannel_UploadsPlaylistID(t *testing.T) {
	tests := []struct {
		name    string
		channel *Channel
		want    string
	}{
		{"nil contentDetails", &Channel{}, ""},
		{"nil relatedPlaylists", &Channel{ContentDetails: &ChannelContentDetails{}}, ""},
		{"has uploads", &Channel{
			ContentDetails: &ChannelContentDetails{
				RelatedPlaylists: &RelatedPlaylists{Uploads: "UU123"},
			},
		}, "UU123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.channel.UploadsPlaylistID(); got != tt.want {
				t.Errorf("UploadsPlaylistID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestChannelListResponse_JSON(t *testing.T) {
	jsonData := `{
		"kind": "youtube#channelListResponse",
		"etag": "abc123",
		"nextPageToken": "page2",
		"pageInfo": {
			"totalResults": 100,
			"resultsPerPage": 50
		},
		"items": [
			{
				"id": "channel123",
				"snippet": {
					"title": "Test Channel",
					"description": "A test channel",
					"customUrl": "@testchannel",
					"publishedAt": "2020-01-15T10:30:00Z"
				},
				"statistics": {
					"viewCount": "1000000",
					"subscriberCount": "50000",
					"hiddenSubscriberCount": false,
					"videoCount": "200"
				},
				"contentDetails": {
					"relatedPlaylists": {
						"likes": "LL",
						"uploads": "UU123"
					}
				}
			}
		]
	}`

	var resp ChannelListResponse
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

	channel := resp.Items[0]
	if channel.Snippet.Title != "Test Channel" {
		t.Errorf("Title = %q, want 'Test Channel'", channel.Snippet.Title)
	}
	if channel.Snippet.CustomURL != "@testchannel" {
		t.Errorf("CustomURL = %q, want '@testchannel'", channel.Snippet.CustomURL)
	}
	if channel.Statistics.SubscriberCount != "50000" {
		t.Errorf("SubscriberCount = %q, want '50000'", channel.Statistics.SubscriberCount)
	}
	if channel.UploadsPlaylistID() != "UU123" {
		t.Errorf("UploadsPlaylistID() = %q, want 'UU123'", channel.UploadsPlaylistID())
	}
}
