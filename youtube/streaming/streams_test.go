package streaming

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Its-donkey/yougopher/youtube/core"
)

func TestGetStreams(t *testing.T) {
	t.Run("success with IDs", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/liveStreams" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("id") != "stream123" {
				t.Errorf("unexpected id: %s", r.URL.Query().Get("id"))
			}
			if r.URL.Query().Get("part") != "snippet,cdn,status" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}

			resp := LiveStreamListResponse{
				Kind: "youtube#liveStreamListResponse",
				Items: []*LiveStream{
					{
						ID:   "stream123",
						Kind: "youtube#liveStream",
						Snippet: &StreamSnippet{
							Title:       "Test Stream",
							Description: "A test stream",
						},
						CDN: &StreamCDN{
							IngestionType: "rtmp",
							Resolution:    "1080p",
							FrameRate:     "30fps",
							IngestionInfo: &IngestionInfo{
								StreamName:       "stream-key-123",
								IngestionAddress: "rtmp://a.rtmp.youtube.com/live2",
							},
						},
						Status: &StreamStatus{
							StreamStatus: StreamStatusActive,
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		resp, err := GetStreams(context.Background(), client, &GetStreamsParams{
			IDs: []string{"stream123"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(resp.Items))
		}
		if resp.Items[0].ID != "stream123" {
			t.Errorf("unexpected ID: %s", resp.Items[0].ID)
		}
		if resp.Items[0].Snippet.Title != "Test Stream" {
			t.Errorf("unexpected title: %s", resp.Items[0].Snippet.Title)
		}
		if resp.Items[0].CDN.IngestionInfo.StreamName != "stream-key-123" {
			t.Errorf("unexpected stream key: %s", resp.Items[0].CDN.IngestionInfo.StreamName)
		}
	})

	t.Run("nil params", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetStreams(context.Background(), client, nil)
		if err == nil {
			t.Fatal("expected error for nil params")
		}
	})

	t.Run("no filter provided", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetStreams(context.Background(), client, &GetStreamsParams{})
		if err == nil {
			t.Fatal("expected error for no filter")
		}
	})

	t.Run("with mine", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("mine") != "true" {
				t.Errorf("unexpected mine: %s", r.URL.Query().Get("mine"))
			}

			resp := LiveStreamListResponse{Items: []*LiveStream{{ID: "stream123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetStreams(context.Background(), client, &GetStreamsParams{
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

			resp := LiveStreamListResponse{Items: []*LiveStream{{ID: "stream123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetStreams(context.Background(), client, &GetStreamsParams{
			IDs:   []string{"stream123"},
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

			resp := LiveStreamListResponse{Items: []*LiveStream{{ID: "stream123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetStreams(context.Background(), client, &GetStreamsParams{
			IDs:        []string{"stream123"},
			MaxResults: 10,
			PageToken:  "nextPage123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetStream(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := LiveStreamListResponse{
				Items: []*LiveStream{{ID: "stream123", Snippet: &StreamSnippet{Title: "Test"}}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		stream, err := GetStream(context.Background(), client, "stream123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if stream.ID != "stream123" {
			t.Errorf("unexpected ID: %s", stream.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := LiveStreamListResponse{Items: []*LiveStream{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetStream(context.Background(), client, "nonexistent")
		if err == nil {
			t.Fatal("expected error for not found stream")
		}
		notFoundErr, ok := err.(*core.NotFoundError)
		if !ok {
			t.Fatalf("expected NotFoundError, got %T", err)
		}
		if notFoundErr.ResourceType != "stream" {
			t.Errorf("unexpected resource type: %s", notFoundErr.ResourceType)
		}
	})

	t.Run("empty stream ID", func(t *testing.T) {
		client := core.NewClient()
		_, err := GetStream(context.Background(), client, "")
		if err == nil {
			t.Fatal("expected error for empty stream ID")
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "snippet" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}
			resp := LiveStreamListResponse{Items: []*LiveStream{{ID: "stream123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetStream(context.Background(), client, "stream123", "snippet")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetMyStreams(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("mine") != "true" {
				t.Errorf("expected mine=true, got %s", r.URL.Query().Get("mine"))
			}

			resp := LiveStreamListResponse{
				Items: []*LiveStream{{ID: "myStream", Snippet: &StreamSnippet{Title: "My Stream"}}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		resp, err := GetMyStreams(context.Background(), client)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(resp.Items))
		}
		if resp.Items[0].ID != "myStream" {
			t.Errorf("unexpected ID: %s", resp.Items[0].ID)
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "contentDetails" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}
			resp := LiveStreamListResponse{Items: []*LiveStream{{ID: "stream123"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetMyStreams(context.Background(), client, "contentDetails")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestInsertStream(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/liveStreams" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			var body LiveStream
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if body.Snippet.Title != "New Stream" {
				t.Errorf("unexpected title: %s", body.Snippet.Title)
			}

			resp := LiveStream{
				ID: "newStream123",
				Snippet: &StreamSnippet{
					Title: "New Stream",
				},
				CDN: &StreamCDN{
					IngestionType: "rtmp",
					Resolution:    "1080p",
					FrameRate:     "30fps",
					IngestionInfo: &IngestionInfo{
						StreamName:       "generated-key",
						IngestionAddress: "rtmp://a.rtmp.youtube.com/live2",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		stream, err := InsertStream(context.Background(), client, &LiveStream{
			Snippet: &StreamSnippet{
				Title: "New Stream",
			},
			CDN: &StreamCDN{
				IngestionType: "rtmp",
				Resolution:    "1080p",
				FrameRate:     "30fps",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if stream.ID != "newStream123" {
			t.Errorf("unexpected ID: %s", stream.ID)
		}
		if stream.CDN.IngestionInfo.StreamName != "generated-key" {
			t.Errorf("unexpected stream key: %s", stream.CDN.IngestionInfo.StreamName)
		}
	})

	t.Run("nil stream", func(t *testing.T) {
		client := core.NewClient()
		_, err := InsertStream(context.Background(), client, nil)
		if err == nil {
			t.Fatal("expected error for nil stream")
		}
	})

	t.Run("missing title", func(t *testing.T) {
		client := core.NewClient()
		_, err := InsertStream(context.Background(), client, &LiveStream{
			Snippet: &StreamSnippet{},
			CDN:     &StreamCDN{},
		})
		if err == nil {
			t.Fatal("expected error for missing title")
		}
	})

	t.Run("nil snippet", func(t *testing.T) {
		client := core.NewClient()
		_, err := InsertStream(context.Background(), client, &LiveStream{
			CDN: &StreamCDN{},
		})
		if err == nil {
			t.Fatal("expected error for nil snippet")
		}
	})

	t.Run("nil CDN", func(t *testing.T) {
		client := core.NewClient()
		_, err := InsertStream(context.Background(), client, &LiveStream{
			Snippet: &StreamSnippet{Title: "Test"},
		})
		if err == nil {
			t.Fatal("expected error for nil CDN")
		}
	})

	t.Run("with custom parts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("part") != "snippet,cdn" {
				t.Errorf("unexpected part: %s", r.URL.Query().Get("part"))
			}

			resp := LiveStream{ID: "stream123"}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := InsertStream(context.Background(), client, &LiveStream{
			Snippet: &StreamSnippet{Title: "Test"},
			CDN:     &StreamCDN{},
		}, "snippet", "cdn")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestUpdateStream(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT, got %s", r.Method)
			}

			var body LiveStream
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if body.ID != "stream123" {
				t.Errorf("unexpected ID: %s", body.ID)
			}
			if body.Snippet.Title != "Updated Title" {
				t.Errorf("unexpected title: %s", body.Snippet.Title)
			}

			resp := LiveStream{
				ID:      "stream123",
				Snippet: &StreamSnippet{Title: "Updated Title"},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		stream, err := UpdateStream(context.Background(), client, &LiveStream{
			ID:      "stream123",
			Snippet: &StreamSnippet{Title: "Updated Title"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if stream.Snippet.Title != "Updated Title" {
			t.Errorf("unexpected title: %s", stream.Snippet.Title)
		}
	})

	t.Run("nil stream", func(t *testing.T) {
		client := core.NewClient()
		_, err := UpdateStream(context.Background(), client, nil)
		if err == nil {
			t.Fatal("expected error for nil stream")
		}
	})

	t.Run("missing ID", func(t *testing.T) {
		client := core.NewClient()
		_, err := UpdateStream(context.Background(), client, &LiveStream{
			Snippet: &StreamSnippet{Title: "Test"},
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

			resp := LiveStream{ID: "stream123"}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := UpdateStream(context.Background(), client, &LiveStream{
			ID:      "stream123",
			Snippet: &StreamSnippet{Title: "Test"},
		}, "snippet")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestDeleteStream(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			if r.URL.Query().Get("id") != "stream123" {
				t.Errorf("unexpected id: %s", r.URL.Query().Get("id"))
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		err := DeleteStream(context.Background(), client, "stream123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty stream ID", func(t *testing.T) {
		client := core.NewClient()
		err := DeleteStream(context.Background(), client, "")
		if err == nil {
			t.Fatal("expected error for empty stream ID")
		}
	})
}

func TestLiveStream_Methods(t *testing.T) {
	t.Run("IsActive", func(t *testing.T) {
		tests := []struct {
			name   string
			stream *LiveStream
			want   bool
		}{
			{"nil status", &LiveStream{}, false},
			{"not active", &LiveStream{Status: &StreamStatus{StreamStatus: StreamStatusCreated}}, false},
			{"active", &LiveStream{Status: &StreamStatus{StreamStatus: StreamStatusActive}}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.stream.IsActive(); got != tt.want {
					t.Errorf("IsActive() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("IsReady", func(t *testing.T) {
		tests := []struct {
			name   string
			stream *LiveStream
			want   bool
		}{
			{"nil status", &LiveStream{}, false},
			{"not ready", &LiveStream{Status: &StreamStatus{StreamStatus: StreamStatusCreated}}, false},
			{"ready", &LiveStream{Status: &StreamStatus{StreamStatus: StreamStatusReady}}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.stream.IsReady(); got != tt.want {
					t.Errorf("IsReady() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("IsHealthy", func(t *testing.T) {
		tests := []struct {
			name   string
			stream *LiveStream
			want   bool
		}{
			{"nil status", &LiveStream{}, false},
			{"nil health", &LiveStream{Status: &StreamStatus{}}, false},
			{"bad health", &LiveStream{Status: &StreamStatus{HealthStatus: &StreamHealthStatus{Status: StreamHealthBad}}}, false},
			{"no data", &LiveStream{Status: &StreamStatus{HealthStatus: &StreamHealthStatus{Status: StreamHealthNoData}}}, false},
			{"good health", &LiveStream{Status: &StreamStatus{HealthStatus: &StreamHealthStatus{Status: StreamHealthGood}}}, true},
			{"ok health", &LiveStream{Status: &StreamStatus{HealthStatus: &StreamHealthStatus{Status: StreamHealthOK}}}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.stream.IsHealthy(); got != tt.want {
					t.Errorf("IsHealthy() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("StreamKey", func(t *testing.T) {
		tests := []struct {
			name   string
			stream *LiveStream
			want   string
		}{
			{"nil CDN", &LiveStream{}, ""},
			{"nil ingestion", &LiveStream{CDN: &StreamCDN{}}, ""},
			{"has key", &LiveStream{CDN: &StreamCDN{IngestionInfo: &IngestionInfo{StreamName: "key123"}}}, "key123"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.stream.StreamKey(); got != tt.want {
					t.Errorf("StreamKey() = %q, want %q", got, tt.want)
				}
			})
		}
	})

	t.Run("RTMPUrl", func(t *testing.T) {
		tests := []struct {
			name   string
			stream *LiveStream
			want   string
		}{
			{"nil CDN", &LiveStream{}, ""},
			{"nil ingestion", &LiveStream{CDN: &StreamCDN{}}, ""},
			{"has URL", &LiveStream{CDN: &StreamCDN{IngestionInfo: &IngestionInfo{IngestionAddress: "rtmp://test"}}}, "rtmp://test"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.stream.RTMPUrl(); got != tt.want {
					t.Errorf("RTMPUrl() = %q, want %q", got, tt.want)
				}
			})
		}
	})

	t.Run("RTMPSUrl", func(t *testing.T) {
		tests := []struct {
			name   string
			stream *LiveStream
			want   string
		}{
			{"nil CDN", &LiveStream{}, ""},
			{"nil ingestion", &LiveStream{CDN: &StreamCDN{}}, ""},
			{"has URL", &LiveStream{CDN: &StreamCDN{IngestionInfo: &IngestionInfo{RtmpsIngestionAddress: "rtmps://test"}}}, "rtmps://test"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.stream.RTMPSUrl(); got != tt.want {
					t.Errorf("RTMPSUrl() = %q, want %q", got, tt.want)
				}
			})
		}
	})

	t.Run("HasConfigurationIssues", func(t *testing.T) {
		tests := []struct {
			name   string
			stream *LiveStream
			want   bool
		}{
			{"nil status", &LiveStream{}, false},
			{"nil health", &LiveStream{Status: &StreamStatus{}}, false},
			{"no issues", &LiveStream{Status: &StreamStatus{HealthStatus: &StreamHealthStatus{}}}, false},
			{"has issues", &LiveStream{Status: &StreamStatus{HealthStatus: &StreamHealthStatus{
				ConfigurationIssues: []*ConfigurationIssue{{Type: "test"}},
			}}}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.stream.HasConfigurationIssues(); got != tt.want {
					t.Errorf("HasConfigurationIssues() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}

func TestLiveStreamListResponse_JSON(t *testing.T) {
	jsonData := `{
		"kind": "youtube#liveStreamListResponse",
		"etag": "abc123",
		"nextPageToken": "page2",
		"pageInfo": {
			"totalResults": 100,
			"resultsPerPage": 50
		},
		"items": [
			{
				"id": "stream123",
				"snippet": {
					"publishedAt": "2024-01-15T10:30:00Z",
					"channelId": "channel123",
					"title": "Test Stream",
					"description": "A test stream",
					"isDefaultStream": true
				},
				"cdn": {
					"ingestionType": "rtmp",
					"resolution": "1080p",
					"frameRate": "30fps",
					"ingestionInfo": {
						"streamName": "key123",
						"ingestionAddress": "rtmp://a.rtmp.youtube.com/live2",
						"backupIngestionAddress": "rtmp://b.rtmp.youtube.com/live2",
						"rtmpsIngestionAddress": "rtmps://a.rtmps.youtube.com/live2",
						"rtmpsBackupIngestionAddress": "rtmps://b.rtmps.youtube.com/live2"
					}
				},
				"status": {
					"streamStatus": "active",
					"healthStatus": {
						"status": "good",
						"lastUpdateTimeSeconds": "1705314600"
					}
				},
				"contentDetails": {
					"closedCaptionsIngestionUrl": "https://captions.youtube.com/ingest",
					"isReusable": true
				}
			}
		]
	}`

	var resp LiveStreamListResponse
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

	stream := resp.Items[0]
	if stream.Snippet.Title != "Test Stream" {
		t.Errorf("Title = %q, want 'Test Stream'", stream.Snippet.Title)
	}
	if !stream.Snippet.IsDefaultStream {
		t.Error("IsDefaultStream = false, want true")
	}
	if stream.CDN.IngestionType != "rtmp" {
		t.Errorf("IngestionType = %q, want 'rtmp'", stream.CDN.IngestionType)
	}
	if stream.CDN.IngestionInfo.StreamName != "key123" {
		t.Errorf("StreamName = %q, want 'key123'", stream.CDN.IngestionInfo.StreamName)
	}
	if stream.Status.StreamStatus != StreamStatusActive {
		t.Errorf("StreamStatus = %q, want %q", stream.Status.StreamStatus, StreamStatusActive)
	}
	if !stream.IsActive() {
		t.Error("IsActive() = false, want true")
	}
	if !stream.IsHealthy() {
		t.Error("IsHealthy() = false, want true")
	}
	if stream.StreamKey() != "key123" {
		t.Errorf("StreamKey() = %q, want 'key123'", stream.StreamKey())
	}
}

func TestStreamStatusConstants(t *testing.T) {
	expectedStatuses := map[string]string{
		"active":   StreamStatusActive,
		"created":  StreamStatusCreated,
		"error":    StreamStatusError,
		"inactive": StreamStatusInactive,
		"ready":    StreamStatusReady,
	}

	for expected, actual := range expectedStatuses {
		if actual != expected {
			t.Errorf("StreamStatus constant %q = %q, want %q", expected, actual, expected)
		}
	}
}

func TestStreamHealthConstants(t *testing.T) {
	expectedHealth := map[string]string{
		"good":   StreamHealthGood,
		"ok":     StreamHealthOK,
		"bad":    StreamHealthBad,
		"noData": StreamHealthNoData,
	}

	for expected, actual := range expectedHealth {
		if actual != expected {
			t.Errorf("StreamHealth constant %q = %q, want %q", expected, actual, expected)
		}
	}
}

// =============================================================================
// Error Type Verification Tests (errors.As)
// =============================================================================

func TestGetStream_ErrorAs_NotFoundError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty items to trigger NotFoundError
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveStreamListResponse{
			Items: []*LiveStream{},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	_, err := GetStream(context.Background(), client, "nonexistent123")

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
		if notFoundErr.ResourceType != "stream" {
			t.Errorf("ResourceType = %q, want 'stream'", notFoundErr.ResourceType)
		}
		if notFoundErr.ResourceID != "nonexistent123" {
			t.Errorf("ResourceID = %q, want 'nonexistent123'", notFoundErr.ResourceID)
		}
	}
}

// =============================================================================
// Mock Call Count and Request Verification Tests
// =============================================================================

func TestInsertStream_VerifyAPICall(t *testing.T) {
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
		_ = json.NewEncoder(w).Encode(LiveStream{
			ID:   "stream123",
			Kind: "youtube#liveStream",
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	stream := &LiveStream{
		Snippet: &StreamSnippet{
			Title: "Test Stream",
		},
		CDN: &StreamCDN{
			IngestionType: "rtmp",
			Resolution:    "1080p",
			FrameRate:     "30fps",
		},
	}

	_, err := InsertStream(context.Background(), client, stream)
	if err != nil {
		t.Fatalf("InsertStream() error = %v", err)
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
	if lastPath != "/liveStreams" {
		t.Errorf("path = %s, want /liveStreams", lastPath)
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

	cdn, ok := lastBody["cdn"].(map[string]any)
	if !ok {
		t.Fatal("request body missing cdn")
	}
	if cdn["ingestionType"] != "rtmp" {
		t.Errorf("ingestionType = %v, want 'rtmp'", cdn["ingestionType"])
	}
}

func TestDeleteStream_VerifyAPICall(t *testing.T) {
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
	err := DeleteStream(context.Background(), client, "stream123")
	if err != nil {
		t.Fatalf("DeleteStream() error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if callCount != 1 {
		t.Errorf("API called %d times, want 1", callCount)
	}

	if lastMethod != http.MethodDelete {
		t.Errorf("method = %s, want DELETE", lastMethod)
	}

	if lastPath != "/liveStreams" {
		t.Errorf("path = %s, want /liveStreams", lastPath)
	}

	if lastQuery != "stream123" {
		t.Errorf("id query param = %s, want stream123", lastQuery)
	}
}

// =============================================================================
// Mutation Testing - Kill Surviving Mutants
// =============================================================================

// TestGetStreams_DefaultParts verifies that default parts are used when none provided.
// Kills mutant: `parts = DefaultStreamParts` → removed
func TestGetStreams_DefaultParts(t *testing.T) {
	var capturedParts string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedParts = r.URL.Query().Get("part")
		resp := LiveStreamListResponse{Items: []*LiveStream{{ID: "s1"}}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	// Call with no Parts specified - should use defaults
	_, err := GetStreams(context.Background(), client, &GetStreamsParams{
		IDs:   []string{"s1"},
		Parts: nil, // No parts - should use DefaultStreamParts
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// DefaultStreamParts is []string{"snippet", "cdn", "status"}
	expected := "snippet,cdn,status"
	if capturedParts != expected {
		t.Errorf("part = %q, want %q (default parts)", capturedParts, expected)
	}
}

// TestGetStream_DefaultParts verifies GetStream uses default parts when none provided.
// Kills mutant: `if len(parts) == 0` → `if len(parts) == -1` or `== 1`
func TestGetStream_DefaultParts(t *testing.T) {
	var capturedParts string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedParts = r.URL.Query().Get("part")
		resp := LiveStreamListResponse{Items: []*LiveStream{{ID: "s1"}}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	// Call with no parts - should use defaults
	_, err := GetStream(context.Background(), client, "s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "snippet,cdn,status"
	if capturedParts != expected {
		t.Errorf("part = %q, want %q (default parts)", capturedParts, expected)
	}
}

// TestGetMyStreams_DefaultParts verifies GetMyStreams uses default parts when none provided.
// Kills mutant: `if len(parts) == 0` → `if len(parts) == -1`
func TestGetMyStreams_DefaultParts(t *testing.T) {
	var capturedParts string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedParts = r.URL.Query().Get("part")
		resp := LiveStreamListResponse{Items: []*LiveStream{{ID: "s1"}}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	// Call with no parts
	_, err := GetMyStreams(context.Background(), client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "snippet,cdn,status"
	if capturedParts != expected {
		t.Errorf("part = %q, want %q (default parts)", capturedParts, expected)
	}
}

// TestInsertStream_DefaultParts verifies InsertStream uses default parts when none provided.
// Kills mutant: `if len(parts) == 0` → `if len(parts) == 1` or `== -1`
func TestInsertStream_DefaultParts(t *testing.T) {
	var capturedParts string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedParts = r.URL.Query().Get("part")
		resp := LiveStream{ID: "s1"}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	// Call with no parts
	_, err := InsertStream(context.Background(), client, &LiveStream{
		Snippet: &StreamSnippet{Title: "Test"},
		CDN:     &StreamCDN{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "snippet,cdn,status"
	if capturedParts != expected {
		t.Errorf("part = %q, want %q (default parts)", capturedParts, expected)
	}
}

// TestUpdateStream_DefaultParts verifies UpdateStream uses default parts when none provided.
// Kills mutant: `if len(parts) == 0` → `if len(parts) == -1`
func TestUpdateStream_DefaultParts(t *testing.T) {
	var capturedParts string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedParts = r.URL.Query().Get("part")
		resp := LiveStream{ID: "s1"}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	// Call with no parts
	_, err := UpdateStream(context.Background(), client, &LiveStream{
		ID: "s1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "snippet,cdn,status"
	if capturedParts != expected {
		t.Errorf("part = %q, want %q (default parts)", capturedParts, expected)
	}
}

// TestGetStreams_MaxResultsBoundary tests MaxResults=1 boundary.
// Kills mutant: `if params.MaxResults > 0` → `> 1`
func TestGetStreams_MaxResultsBoundary(t *testing.T) {
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
				resp := LiveStreamListResponse{Items: []*LiveStream{{ID: "s1"}}}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := core.NewClient(core.WithBaseURL(server.URL))
			_, err := GetStreams(context.Background(), client, &GetStreamsParams{
				IDs:        []string{"s1"},
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

// TestGetStreams_IDsSetWhenProvided tests that IDs are sent when non-empty.
// Kills mutant: `if len(params.IDs) > 0` → `>= 0` or `> -1`
func TestGetStreams_IDsSetWhenProvided(t *testing.T) {
	t.Run("IDs provided - should be in query", func(t *testing.T) {
		var capturedID string
		var hasID bool

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedID = r.URL.Query().Get("id")
			hasID = r.URL.Query().Has("id")
			resp := LiveStreamListResponse{Items: []*LiveStream{{ID: "s1"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetStreams(context.Background(), client, &GetStreamsParams{
			IDs: []string{"s1"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !hasID {
			t.Error("expected id to be in query when IDs provided")
		}
		if capturedID != "s1" {
			t.Errorf("id = %q, want 's1'", capturedID)
		}
	})

	t.Run("no IDs with Mine=true - should not have id in query", func(t *testing.T) {
		var hasID bool

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hasID = r.URL.Query().Has("id")
			resp := LiveStreamListResponse{Items: []*LiveStream{{ID: "s1"}}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		_, err := GetStreams(context.Background(), client, &GetStreamsParams{
			IDs:  []string{}, // Empty IDs
			Mine: true,       // But Mine is true so it's valid
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if hasID {
			t.Error("id should not be in query when IDs is empty")
		}
	})
}

// TestGetStreams_ErrorPropagation verifies errors from API are returned.
// Kills mutant: `return nil, err` → `_ = err`
func TestGetStreams_ErrorPropagation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    500,
				"message": "Internal Server Error",
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	result, err := GetStreams(context.Background(), client, &GetStreamsParams{
		IDs: []string{"s1"},
	})

	// The key assertion: we must get an error back, not nil
	if err == nil {
		t.Fatal("expected error from failed API call, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
}

// TestGetStream_ErrorPropagation verifies GetStream propagates errors.
// Kills mutant: `return nil, err` → `_ = err`
func TestGetStream_ErrorPropagation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    500,
				"message": "Internal Server Error",
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	result, err := GetStream(context.Background(), client, "s1")

	if err == nil {
		t.Fatal("expected error from failed API call, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
}

// TestInsertStream_ErrorPropagation verifies InsertStream propagates errors.
// Kills mutant: `return nil, err` → `_ = err`
func TestInsertStream_ErrorPropagation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    400,
				"message": "Bad Request",
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	result, err := InsertStream(context.Background(), client, &LiveStream{
		Snippet: &StreamSnippet{Title: "Test"},
		CDN:     &StreamCDN{},
	})

	if err == nil {
		t.Fatal("expected error from failed API call, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
}

// TestUpdateStream_ErrorPropagation verifies UpdateStream propagates errors.
// Kills mutant: `return nil, err` → `_ = err`
func TestUpdateStream_ErrorPropagation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    403,
				"message": "Forbidden",
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	result, err := UpdateStream(context.Background(), client, &LiveStream{
		ID: "s1",
	})

	if err == nil {
		t.Fatal("expected error from failed API call, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
}

// TestDeleteStream_ErrorPropagation verifies DeleteStream propagates errors.
func TestDeleteStream_ErrorPropagation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    404,
				"message": "Not Found",
			},
		})
	}))
	defer server.Close()

	client := core.NewClient(core.WithBaseURL(server.URL))
	err := DeleteStream(context.Background(), client, "s1")

	if err == nil {
		t.Fatal("expected error from failed API call, got nil")
	}
}

// TestInsertStream_EmptyTitle tests that empty title (with valid snippet) is rejected.
// Kills mutant: `stream.Snippet.Title == ""` → `false`
func TestInsertStream_EmptyTitle(t *testing.T) {
	client := core.NewClient()
	_, err := InsertStream(context.Background(), client, &LiveStream{
		Snippet: &StreamSnippet{
			Title:       "",         // Empty title
			Description: "Has desc", // But has other fields
		},
		CDN: &StreamCDN{},
	})
	if err == nil {
		t.Fatal("expected error for empty title, got nil")
	}
}

// TestGetStreams_ValidationError_EmptyIDsNoMine tests validation returns error.
// Kills mutant: `return nil, fmt.Errorf(...)` → `_ = fmt.Errorf`
func TestGetStreams_ValidationError_EmptyIDsNoMine(t *testing.T) {
	client := core.NewClient()
	// Empty IDs and Mine=false should return validation error
	result, err := GetStreams(context.Background(), client, &GetStreamsParams{
		IDs:  []string{},
		Mine: false,
	})

	if err == nil {
		t.Fatal("expected validation error for empty IDs and Mine=false")
	}
	if result != nil {
		t.Errorf("expected nil result on validation error, got %+v", result)
	}
}

// TestGetStream_ValidationError_EmptyID tests empty stream ID returns error.
// Kills mutant: `return nil, fmt.Errorf(...)` → `_ = fmt.Errorf`
func TestGetStream_ValidationError_EmptyID(t *testing.T) {
	client := core.NewClient()
	result, err := GetStream(context.Background(), client, "")

	if err == nil {
		t.Fatal("expected validation error for empty stream ID")
	}
	if result != nil {
		t.Errorf("expected nil result on validation error, got %+v", result)
	}
}

// TestInsertStream_ValidationErrors tests all validation paths return errors.
// Kills mutant: `return nil, fmt.Errorf(...)` → `_ = fmt.Errorf`
func TestInsertStream_ValidationErrors(t *testing.T) {
	client := core.NewClient()

	tests := []struct {
		name   string
		stream *LiveStream
	}{
		{"nil stream", nil},
		{"nil snippet", &LiveStream{CDN: &StreamCDN{}}},
		{"empty title", &LiveStream{Snippet: &StreamSnippet{Title: ""}, CDN: &StreamCDN{}}},
		{"nil CDN", &LiveStream{Snippet: &StreamSnippet{Title: "Test"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := InsertStream(context.Background(), client, tt.stream)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if result != nil {
				t.Errorf("expected nil result on validation error, got %+v", result)
			}
		})
	}
}

// TestUpdateStream_ValidationErrors tests all validation paths return errors.
// Kills mutant: `return nil, fmt.Errorf(...)` → `_ = fmt.Errorf`
func TestUpdateStream_ValidationErrors(t *testing.T) {
	client := core.NewClient()

	tests := []struct {
		name   string
		stream *LiveStream
	}{
		{"nil stream", nil},
		{"missing ID", &LiveStream{Snippet: &StreamSnippet{Title: "Test"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := UpdateStream(context.Background(), client, tt.stream)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if result != nil {
				t.Errorf("expected nil result on validation error, got %+v", result)
			}
		})
	}
}

// TestDeleteStream_ValidationError_EmptyID tests empty stream ID returns error.
// Kills mutant: `return fmt.Errorf(...)` → `_ = fmt.Errorf`
func TestDeleteStream_ValidationError_EmptyID(t *testing.T) {
	client := core.NewClient()
	err := DeleteStream(context.Background(), client, "")

	if err == nil {
		t.Fatal("expected validation error for empty stream ID")
	}
}
