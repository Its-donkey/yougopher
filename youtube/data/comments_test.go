package data

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Its-donkey/yougopher/youtube/core"
)

func TestGetCommentThreads(t *testing.T) {
	t.Run("success with videoId", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/commentThreads" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("videoId") != "video123" {
				t.Errorf("unexpected videoId: %s", r.URL.Query().Get("videoId"))
			}

			resp := CommentThreadListResponse{
				Items: []*CommentThread{{
					ID: "thread123",
					Snippet: &CommentThreadSnippet{
						VideoID:         "video123",
						TotalReplyCount: 5,
						TopLevelComment: &Comment{
							ID: "comment123",
							Snippet: &CommentSnippet{
								TextDisplay:       "Great video!",
								AuthorDisplayName: "Test User",
							},
						},
					},
				}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		resp, err := GetCommentThreads(context.Background(), client, &GetCommentThreadsParams{
			VideoID: "video123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(resp.Items))
		}
		if resp.Items[0].ReplyCount() != 5 {
			t.Errorf("unexpected reply count: %d", resp.Items[0].ReplyCount())
		}
	})

	t.Run("nil params", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetCommentThreads(context.Background(), client, nil)
		if err == nil {
			t.Fatal("expected error for nil params")
		}
	})

	t.Run("no filter provided", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetCommentThreads(context.Background(), client, &GetCommentThreadsParams{})
		if err == nil {
			t.Fatal("expected error for no filter")
		}
	})

	t.Run("with IDs", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("id") != "thread1,thread2" {
				t.Errorf("unexpected id: %s", r.URL.Query().Get("id"))
			}
			resp := CommentThreadListResponse{Items: []*CommentThread{{ID: "thread1"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetCommentThreads(context.Background(), client, &GetCommentThreadsParams{
			IDs: []string{"thread1", "thread2"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with channelId", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("channelId") != "channel123" {
				t.Errorf("unexpected channelId: %s", r.URL.Query().Get("channelId"))
			}
			resp := CommentThreadListResponse{Items: []*CommentThread{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetCommentThreads(context.Background(), client, &GetCommentThreadsParams{
			ChannelID: "channel123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with allThreadsRelatedToChannelId", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("allThreadsRelatedToChannelId") != "channel123" {
				t.Errorf("unexpected allThreadsRelatedToChannelId: %s", r.URL.Query().Get("allThreadsRelatedToChannelId"))
			}
			resp := CommentThreadListResponse{Items: []*CommentThread{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetCommentThreads(context.Background(), client, &GetCommentThreadsParams{
			AllThreadsRelatedToChannelID: "channel123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with all parameters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if q.Get("moderationStatus") != "published" {
				t.Errorf("unexpected moderationStatus: %s", q.Get("moderationStatus"))
			}
			if q.Get("order") != "time" {
				t.Errorf("unexpected order: %s", q.Get("order"))
			}
			if q.Get("searchTerms") != "awesome" {
				t.Errorf("unexpected searchTerms: %s", q.Get("searchTerms"))
			}
			if q.Get("maxResults") != "25" {
				t.Errorf("unexpected maxResults: %s", q.Get("maxResults"))
			}
			if q.Get("pageToken") != "nextPage" {
				t.Errorf("unexpected pageToken: %s", q.Get("pageToken"))
			}
			resp := CommentThreadListResponse{Items: []*CommentThread{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetCommentThreads(context.Background(), client, &GetCommentThreadsParams{
			VideoID:          "video123",
			ModerationStatus: ModerationStatusPublished,
			Order:            CommentOrderTime,
			SearchTerms:      "awesome",
			MaxResults:       25,
			PageToken:        "nextPage",
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
			resp := CommentThreadListResponse{Items: []*CommentThread{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetCommentThreads(context.Background(), client, &GetCommentThreadsParams{
			VideoID: "video123",
			Parts:   []string{"snippet"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetVideoComments(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("videoId") != "video123" {
				t.Errorf("unexpected videoId: %s", r.URL.Query().Get("videoId"))
			}
			if r.URL.Query().Get("order") != "relevance" {
				t.Errorf("unexpected order: %s", r.URL.Query().Get("order"))
			}
			resp := CommentThreadListResponse{Items: []*CommentThread{{ID: "thread1"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		resp, err := GetVideoComments(context.Background(), client, "video123", 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Items) != 1 {
			t.Errorf("expected 1 item, got %d", len(resp.Items))
		}
	})

	t.Run("empty video ID", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetVideoComments(context.Background(), client, "", 10)
		if err == nil {
			t.Fatal("expected error for empty video ID")
		}
	})
}

func TestGetComments(t *testing.T) {
	t.Run("success with IDs", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/comments" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("id") != "comment1,comment2" {
				t.Errorf("unexpected id: %s", r.URL.Query().Get("id"))
			}

			resp := CommentListResponse{
				Items: []*Comment{{
					ID: "comment1",
					Snippet: &CommentSnippet{
						TextDisplay: "Test comment",
					},
				}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		resp, err := GetComments(context.Background(), client, &GetCommentsParams{
			IDs: []string{"comment1", "comment2"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(resp.Items))
		}
	})

	t.Run("nil params", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetComments(context.Background(), client, nil)
		if err == nil {
			t.Fatal("expected error for nil params")
		}
	})

	t.Run("no filter provided", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetComments(context.Background(), client, &GetCommentsParams{})
		if err == nil {
			t.Fatal("expected error for no filter")
		}
	})

	t.Run("with parentId", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("parentId") != "parent123" {
				t.Errorf("unexpected parentId: %s", r.URL.Query().Get("parentId"))
			}
			resp := CommentListResponse{Items: []*Comment{{ID: "reply1"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetComments(context.Background(), client, &GetCommentsParams{
			ParentID: "parent123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with pagination", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("maxResults") != "50" {
				t.Errorf("unexpected maxResults: %s", r.URL.Query().Get("maxResults"))
			}
			if r.URL.Query().Get("pageToken") != "nextPage" {
				t.Errorf("unexpected pageToken: %s", r.URL.Query().Get("pageToken"))
			}
			resp := CommentListResponse{Items: []*Comment{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetComments(context.Background(), client, &GetCommentsParams{
			IDs:        []string{"comment1"},
			MaxResults: 50,
			PageToken:  "nextPage",
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
			resp := CommentListResponse{Items: []*Comment{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetComments(context.Background(), client, &GetCommentsParams{
			IDs:   []string{"comment1"},
			Parts: []string{"id"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetCommentReplies(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("parentId") != "parent123" {
				t.Errorf("unexpected parentId: %s", r.URL.Query().Get("parentId"))
			}
			resp := CommentListResponse{Items: []*Comment{{ID: "reply1"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		resp, err := GetCommentReplies(context.Background(), client, "parent123", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Items) != 1 {
			t.Errorf("expected 1 item, got %d", len(resp.Items))
		}
	})

	t.Run("empty parent ID", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetCommentReplies(context.Background(), client, "", 10)
		if err == nil {
			t.Fatal("expected error for empty parent ID")
		}
	})
}

func TestCommentThread_Methods(t *testing.T) {
	t.Run("TopLevelComment", func(t *testing.T) {
		tests := []struct {
			name   string
			thread *CommentThread
			want   bool
		}{
			{"nil snippet", &CommentThread{}, false},
			{"nil comment", &CommentThread{Snippet: &CommentThreadSnippet{}}, false},
			{"has comment", &CommentThread{
				Snippet: &CommentThreadSnippet{
					TopLevelComment: &Comment{ID: "comment123"},
				},
			}, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := tt.thread.TopLevelComment()
				hasComment := got != nil
				if hasComment != tt.want {
					t.Errorf("TopLevelComment() returned comment = %v, want %v", hasComment, tt.want)
				}
			})
		}
	})

	t.Run("ReplyCount", func(t *testing.T) {
		tests := []struct {
			name   string
			thread *CommentThread
			want   int
		}{
			{"nil snippet", &CommentThread{}, 0},
			{"zero replies", &CommentThread{Snippet: &CommentThreadSnippet{TotalReplyCount: 0}}, 0},
			{"has replies", &CommentThread{Snippet: &CommentThreadSnippet{TotalReplyCount: 10}}, 10},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.thread.ReplyCount(); got != tt.want {
					t.Errorf("ReplyCount() = %d, want %d", got, tt.want)
				}
			})
		}
	})
}

func TestComment_Methods(t *testing.T) {
	t.Run("AuthorID", func(t *testing.T) {
		tests := []struct {
			name    string
			comment *Comment
			want    string
		}{
			{"nil snippet", &Comment{}, ""},
			{"nil authorChannelId", &Comment{Snippet: &CommentSnippet{}}, ""},
			{"has authorChannelId", &Comment{
				Snippet: &CommentSnippet{
					AuthorChannelID: &AuthorChannelID{Value: "channel123"},
				},
			}, "channel123"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.comment.AuthorID(); got != tt.want {
					t.Errorf("AuthorID() = %q, want %q", got, tt.want)
				}
			})
		}
	})

	t.Run("IsReply", func(t *testing.T) {
		tests := []struct {
			name    string
			comment *Comment
			want    bool
		}{
			{"nil snippet", &Comment{}, false},
			{"no parentId", &Comment{Snippet: &CommentSnippet{}}, false},
			{"is reply", &Comment{Snippet: &CommentSnippet{ParentID: "parent123"}}, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.comment.IsReply(); got != tt.want {
					t.Errorf("IsReply() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}

func TestCommentConstants(t *testing.T) {
	// Moderation status constants
	if ModerationStatusPublished != "published" {
		t.Errorf("ModerationStatusPublished = %q, want 'published'", ModerationStatusPublished)
	}
	if ModerationStatusHeldForReview != "heldForReview" {
		t.Errorf("ModerationStatusHeldForReview = %q, want 'heldForReview'", ModerationStatusHeldForReview)
	}

	// Order constants
	if CommentOrderRelevance != "relevance" {
		t.Errorf("CommentOrderRelevance = %q, want 'relevance'", CommentOrderRelevance)
	}
	if CommentOrderTime != "time" {
		t.Errorf("CommentOrderTime = %q, want 'time'", CommentOrderTime)
	}
}

func TestCommentThreadListResponse_JSON(t *testing.T) {
	jsonData := `{
		"kind": "youtube#commentThreadListResponse",
		"etag": "abc123",
		"nextPageToken": "page2",
		"pageInfo": {"totalResults": 100, "resultsPerPage": 20},
		"items": [
			{
				"id": "thread123",
				"snippet": {
					"videoId": "video123",
					"topLevelComment": {
						"id": "comment123",
						"snippet": {
							"textDisplay": "Great video!",
							"authorDisplayName": "Test User",
							"likeCount": 42
						}
					},
					"canReply": true,
					"totalReplyCount": 5,
					"isPublic": true
				}
			}
		]
	}`

	var resp CommentThreadListResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if resp.NextPageToken != "page2" {
		t.Errorf("NextPageToken = %q, want 'page2'", resp.NextPageToken)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(resp.Items))
	}

	thread := resp.Items[0]
	if thread.Snippet.TotalReplyCount != 5 {
		t.Errorf("TotalReplyCount = %d, want 5", thread.Snippet.TotalReplyCount)
	}
	if thread.TopLevelComment().Snippet.TextDisplay != "Great video!" {
		t.Errorf("TextDisplay = %q, want 'Great video!'", thread.TopLevelComment().Snippet.TextDisplay)
	}
	if thread.TopLevelComment().Snippet.LikeCount != 42 {
		t.Errorf("LikeCount = %d, want 42", thread.TopLevelComment().Snippet.LikeCount)
	}
}
