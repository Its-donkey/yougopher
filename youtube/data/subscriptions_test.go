package data

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Its-donkey/yougopher/youtube/core"
)

func TestGetSubscriptions(t *testing.T) {
	t.Run("success with IDs", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/subscriptions" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("id") != "sub123" {
				t.Errorf("unexpected id: %s", r.URL.Query().Get("id"))
			}

			resp := SubscriptionListResponse{
				Items: []*Subscription{{
					ID: "sub123",
					Snippet: &SubscriptionSnippet{
						Title:      "Test Channel",
						ResourceID: &SubscriptionResourceID{ChannelID: "channel123"},
					},
				}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		resp, err := GetSubscriptions(context.Background(), client, &GetSubscriptionsParams{
			IDs: []string{"sub123"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(resp.Items))
		}
		if resp.Items[0].SubscribedChannelID() != "channel123" {
			t.Errorf("unexpected channel ID: %s", resp.Items[0].SubscribedChannelID())
		}
	})

	t.Run("nil params", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetSubscriptions(context.Background(), client, nil)
		if err == nil {
			t.Fatal("expected error for nil params")
		}
	})

	t.Run("no filter provided", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetSubscriptions(context.Background(), client, &GetSubscriptionsParams{})
		if err == nil {
			t.Fatal("expected error for no filter")
		}
	})

	t.Run("with channelId", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("channelId") != "channel123" {
				t.Errorf("unexpected channelId: %s", r.URL.Query().Get("channelId"))
			}
			resp := SubscriptionListResponse{Items: []*Subscription{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetSubscriptions(context.Background(), client, &GetSubscriptionsParams{
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
			resp := SubscriptionListResponse{Items: []*Subscription{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetSubscriptions(context.Background(), client, &GetSubscriptionsParams{
			Mine: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with myRecentSubscribers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("myRecentSubscribers") != "true" {
				t.Errorf("unexpected myRecentSubscribers: %s", r.URL.Query().Get("myRecentSubscribers"))
			}
			resp := SubscriptionListResponse{Items: []*Subscription{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetSubscriptions(context.Background(), client, &GetSubscriptionsParams{
			MyRecentSubscribers: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with mySubscribers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("mySubscribers") != "true" {
				t.Errorf("unexpected mySubscribers: %s", r.URL.Query().Get("mySubscribers"))
			}
			resp := SubscriptionListResponse{Items: []*Subscription{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetSubscriptions(context.Background(), client, &GetSubscriptionsParams{
			MySubscribers: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with all parameters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if q.Get("forChannelId") != "target123" {
				t.Errorf("unexpected forChannelId: %s", q.Get("forChannelId"))
			}
			if q.Get("order") != "alphabetical" {
				t.Errorf("unexpected order: %s", q.Get("order"))
			}
			if q.Get("maxResults") != "25" {
				t.Errorf("unexpected maxResults: %s", q.Get("maxResults"))
			}
			if q.Get("pageToken") != "nextPage" {
				t.Errorf("unexpected pageToken: %s", q.Get("pageToken"))
			}
			resp := SubscriptionListResponse{Items: []*Subscription{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetSubscriptions(context.Background(), client, &GetSubscriptionsParams{
			Mine:         true,
			ForChannelID: "target123",
			Order:        SubscriptionOrderAlphabetical,
			MaxResults:   25,
			PageToken:    "nextPage",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "snippet,subscriberSnippet" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}
			resp := SubscriptionListResponse{Items: []*Subscription{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetSubscriptions(context.Background(), client, &GetSubscriptionsParams{
			Mine:  true,
			Parts: []string{"snippet", "subscriberSnippet"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetMySubscriptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("mine") != "true" {
			t.Errorf("expected mine=true")
		}
		if r.URL.Query().Get("order") != "alphabetical" {
			t.Errorf("unexpected order: %s", r.URL.Query().Get("order"))
		}
		if r.URL.Query().Get("maxResults") != "10" {
			t.Errorf("unexpected maxResults: %s", r.URL.Query().Get("maxResults"))
		}
		resp := SubscriptionListResponse{Items: []*Subscription{{ID: "sub1"}}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	resp, err := GetMySubscriptions(context.Background(), client, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Items) != 1 || resp.Items[0].ID != "sub1" {
		t.Error("unexpected response")
	}
}

func TestGetChannelSubscriptions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("channelId") != "channel123" {
				t.Errorf("unexpected channelId: %s", r.URL.Query().Get("channelId"))
			}
			resp := SubscriptionListResponse{Items: []*Subscription{{ID: "sub1"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		resp, err := GetChannelSubscriptions(context.Background(), client, "channel123", 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Items) != 1 {
			t.Errorf("expected 1 item, got %d", len(resp.Items))
		}
	})

	t.Run("empty channel ID", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetChannelSubscriptions(context.Background(), client, "", 10)
		if err == nil {
			t.Fatal("expected error for empty channel ID")
		}
	})
}

func TestIsSubscribedTo(t *testing.T) {
	t.Run("subscribed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("mine") != "true" {
				t.Errorf("expected mine=true")
			}
			if r.URL.Query().Get("forChannelId") != "targetChannel" {
				t.Errorf("unexpected forChannelId: %s", r.URL.Query().Get("forChannelId"))
			}
			resp := SubscriptionListResponse{Items: []*Subscription{{ID: "sub1"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		subscribed, err := IsSubscribedTo(context.Background(), client, "targetChannel")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !subscribed {
			t.Error("expected subscribed = true")
		}
	})

	t.Run("not subscribed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := SubscriptionListResponse{Items: []*Subscription{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		subscribed, err := IsSubscribedTo(context.Background(), client, "targetChannel")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if subscribed {
			t.Error("expected subscribed = false")
		}
	})

	t.Run("empty channel ID", func(t *testing.T) {
		client := core.NewClient()
		_, err := IsSubscribedTo(context.Background(), client, "")
		if err == nil {
			t.Fatal("expected error for empty channel ID")
		}
	})
}

func TestSubscription_Methods(t *testing.T) {
	t.Run("SubscribedChannelID", func(t *testing.T) {
		tests := []struct {
			name         string
			subscription *Subscription
			want         string
		}{
			{"nil snippet", &Subscription{}, ""},
			{"nil resourceId", &Subscription{Snippet: &SubscriptionSnippet{}}, ""},
			{"has channelId", &Subscription{
				Snippet: &SubscriptionSnippet{
					ResourceID: &SubscriptionResourceID{ChannelID: "channel123"},
				},
			}, "channel123"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.subscription.SubscribedChannelID(); got != tt.want {
					t.Errorf("SubscribedChannelID() = %q, want %q", got, tt.want)
				}
			})
		}
	})

	t.Run("HasNewContent", func(t *testing.T) {
		tests := []struct {
			name         string
			subscription *Subscription
			want         bool
		}{
			{"nil contentDetails", &Subscription{}, false},
			{"no new content", &Subscription{ContentDetails: &SubscriptionContentDetails{NewItemCount: 0}}, false},
			{"has new content", &Subscription{ContentDetails: &SubscriptionContentDetails{NewItemCount: 5}}, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.subscription.HasNewContent(); got != tt.want {
					t.Errorf("HasNewContent() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}

func TestSubscriptionConstants(t *testing.T) {
	if SubscriptionOrderAlphabetical != "alphabetical" {
		t.Errorf("SubscriptionOrderAlphabetical = %q, want 'alphabetical'", SubscriptionOrderAlphabetical)
	}
	if SubscriptionOrderRelevance != "relevance" {
		t.Errorf("SubscriptionOrderRelevance = %q, want 'relevance'", SubscriptionOrderRelevance)
	}
	if SubscriptionOrderUnread != "unread" {
		t.Errorf("SubscriptionOrderUnread = %q, want 'unread'", SubscriptionOrderUnread)
	}
}

func TestSubscriptionListResponse_JSON(t *testing.T) {
	jsonData := `{
		"kind": "youtube#subscriptionListResponse",
		"etag": "abc123",
		"nextPageToken": "page2",
		"pageInfo": {"totalResults": 50, "resultsPerPage": 25},
		"items": [
			{
				"id": "sub123",
				"snippet": {
					"title": "Test Channel",
					"description": "A test channel",
					"resourceId": {
						"kind": "youtube#channel",
						"channelId": "channel123"
					}
				},
				"contentDetails": {
					"totalItemCount": 100,
					"newItemCount": 5,
					"activityType": "all"
				}
			}
		]
	}`

	var resp SubscriptionListResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if resp.NextPageToken != "page2" {
		t.Errorf("NextPageToken = %q, want 'page2'", resp.NextPageToken)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(resp.Items))
	}

	sub := resp.Items[0]
	if sub.Snippet.Title != "Test Channel" {
		t.Errorf("Title = %q, want 'Test Channel'", sub.Snippet.Title)
	}
	if sub.SubscribedChannelID() != "channel123" {
		t.Errorf("SubscribedChannelID() = %q, want 'channel123'", sub.SubscribedChannelID())
	}
	if !sub.HasNewContent() {
		t.Error("expected HasNewContent() = true")
	}
	if sub.ContentDetails.NewItemCount != 5 {
		t.Errorf("NewItemCount = %d, want 5", sub.ContentDetails.NewItemCount)
	}
}
