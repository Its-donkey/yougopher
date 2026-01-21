package streaming

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Its-donkey/yougopher/youtube/core"
)

// controllerMockTokenProvider implements TokenProvider for testing.
type controllerMockTokenProvider struct {
	token string
	err   error
	calls int32
}

func (m *controllerMockTokenProvider) AccessToken(ctx context.Context) (string, error) {
	atomic.AddInt32(&m.calls, 1)
	if m.err != nil {
		return "", m.err
	}
	return m.token, nil
}

func TestNewStreamController(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := core.NewClient()
		provider := &controllerMockTokenProvider{token: "test-token"}

		controller, err := NewStreamController(client, provider)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if controller == nil {
			t.Fatal("controller should not be nil")
		}
	})

	t.Run("nil client", func(t *testing.T) {
		provider := &controllerMockTokenProvider{token: "test-token"}

		_, err := NewStreamController(nil, provider)
		if err == nil {
			t.Fatal("expected error for nil client")
		}
	})

	t.Run("nil token provider", func(t *testing.T) {
		client := core.NewClient()

		controller, err := NewStreamController(client, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if controller == nil {
			t.Fatal("controller should not be nil")
		}
	})
}

func TestStreamController_CreateBroadcastWithStream(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++

			switch {
			case r.URL.Path == "/liveBroadcasts" && r.Method == http.MethodPost:
				// Insert broadcast
				resp := LiveBroadcast{
					ID:      "broadcast123",
					Snippet: &BroadcastSnippet{Title: "Test Broadcast", LiveChatID: "chat123"},
					Status:  &BroadcastStatus{LifeCycleStatus: BroadcastStatusCreated},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)

			case r.URL.Path == "/liveStreams" && r.Method == http.MethodPost:
				// Insert stream
				resp := LiveStream{
					ID:      "stream456",
					Snippet: &StreamSnippet{Title: "Test Stream"},
					CDN: &StreamCDN{
						IngestionInfo: &IngestionInfo{
							StreamName:       "stream-key-123",
							IngestionAddress: "rtmp://ingest.youtube.com/live2",
						},
					},
					Status: &StreamStatus{StreamStatus: StreamStatusCreated},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)

			case r.URL.Path == "/liveBroadcasts/bind":
				// Bind
				resp := LiveBroadcast{
					ID:             "broadcast123",
					Snippet:        &BroadcastSnippet{Title: "Test Broadcast"},
					ContentDetails: &BroadcastContentDetails{BoundStreamID: "stream456"},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)

			default:
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		provider := &controllerMockTokenProvider{token: "test-token"}
		controller, _ := NewStreamController(client, provider)

		result, err := controller.CreateBroadcastWithStream(context.Background(), &CreateBroadcastParams{
			Title:         "Test Broadcast",
			Description:   "A test broadcast",
			PrivacyStatus: "unlisted",
			Resolution:    "1080p",
			FrameRate:     "30fps",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Broadcast.ID != "broadcast123" {
			t.Errorf("unexpected broadcast ID: %s", result.Broadcast.ID)
		}
		if result.Stream.ID != "stream456" {
			t.Errorf("unexpected stream ID: %s", result.Stream.ID)
		}
		if result.Stream.CDN.IngestionInfo.StreamName != "stream-key-123" {
			t.Errorf("unexpected stream key: %s", result.Stream.CDN.IngestionInfo.StreamName)
		}

		if callCount != 3 {
			t.Errorf("expected 3 API calls, got %d", callCount)
		}
	})

	t.Run("nil params", func(t *testing.T) {
		client := core.NewClient()
		controller, _ := NewStreamController(client, nil)

		_, err := controller.CreateBroadcastWithStream(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error for nil params")
		}
	})

	t.Run("missing title", func(t *testing.T) {
		client := core.NewClient()
		controller, _ := NewStreamController(client, nil)

		_, err := controller.CreateBroadcastWithStream(context.Background(), &CreateBroadcastParams{})
		if err == nil {
			t.Fatal("expected error for missing title")
		}
	})

	t.Run("with defaults", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/liveBroadcasts" && r.Method == http.MethodPost:
				var body LiveBroadcast
				_ = json.NewDecoder(r.Body).Decode(&body)
				// Check defaults were applied
				if body.Status.PrivacyStatus != "unlisted" {
					t.Errorf("expected unlisted privacy, got %s", body.Status.PrivacyStatus)
				}

				resp := LiveBroadcast{ID: "broadcast123"}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)

			case r.URL.Path == "/liveStreams" && r.Method == http.MethodPost:
				var body LiveStream
				_ = json.NewDecoder(r.Body).Decode(&body)
				// Check defaults were applied
				if body.CDN.Resolution != "1080p" {
					t.Errorf("expected 1080p resolution, got %s", body.CDN.Resolution)
				}
				if body.CDN.FrameRate != "30fps" {
					t.Errorf("expected 30fps frame rate, got %s", body.CDN.FrameRate)
				}
				if body.CDN.IngestionType != "rtmp" {
					t.Errorf("expected rtmp ingestion, got %s", body.CDN.IngestionType)
				}

				resp := LiveStream{ID: "stream456", CDN: &StreamCDN{IngestionInfo: &IngestionInfo{}}}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)

			case r.URL.Path == "/liveBroadcasts/bind":
				resp := LiveBroadcast{ID: "broadcast123"}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			}
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		controller, _ := NewStreamController(client, nil)

		_, err := controller.CreateBroadcastWithStream(context.Background(), &CreateBroadcastParams{
			Title: "Test",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with scheduled start time", func(t *testing.T) {
		scheduledTime := time.Now().Add(24 * time.Hour)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/liveBroadcasts" && r.Method == http.MethodPost:
				var body LiveBroadcast
				_ = json.NewDecoder(r.Body).Decode(&body)
				if body.Snippet.ScheduledStartTime == nil {
					t.Error("expected scheduled start time")
				}

				resp := LiveBroadcast{ID: "broadcast123"}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)

			case r.URL.Path == "/liveStreams" && r.Method == http.MethodPost:
				resp := LiveStream{ID: "stream456", CDN: &StreamCDN{IngestionInfo: &IngestionInfo{}}}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)

			case r.URL.Path == "/liveBroadcasts/bind":
				resp := LiveBroadcast{ID: "broadcast123"}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			}
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		controller, _ := NewStreamController(client, nil)

		_, err := controller.CreateBroadcastWithStream(context.Background(), &CreateBroadcastParams{
			Title:              "Scheduled Test",
			ScheduledStartTime: &scheduledTime,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestStreamController_StartTesting(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/liveBroadcasts/transition" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("broadcastStatus") != "testing" {
				t.Errorf("unexpected status: %s", r.URL.Query().Get("broadcastStatus"))
			}

			resp := LiveBroadcast{
				ID:     "broadcast123",
				Status: &BroadcastStatus{LifeCycleStatus: BroadcastStatusTesting},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		controller, _ := NewStreamController(client, nil)

		broadcast, err := controller.StartTesting(context.Background(), "broadcast123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !broadcast.IsTesting() {
			t.Error("broadcast should be in testing state")
		}
	})

	t.Run("empty broadcast ID", func(t *testing.T) {
		client := core.NewClient()
		controller, _ := NewStreamController(client, nil)

		_, err := controller.StartTesting(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty broadcast ID")
		}
	})
}

func TestStreamController_GoLive(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
		controller, _ := NewStreamController(client, nil)

		broadcast, err := controller.GoLive(context.Background(), "broadcast123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !broadcast.IsLive() {
			t.Error("broadcast should be live")
		}
	})

	t.Run("empty broadcast ID", func(t *testing.T) {
		client := core.NewClient()
		controller, _ := NewStreamController(client, nil)

		_, err := controller.GoLive(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty broadcast ID")
		}
	})
}

func TestStreamController_EndBroadcast(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
		controller, _ := NewStreamController(client, nil)

		broadcast, err := controller.EndBroadcast(context.Background(), "broadcast123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !broadcast.IsComplete() {
			t.Error("broadcast should be complete")
		}
	})

	t.Run("empty broadcast ID", func(t *testing.T) {
		client := core.NewClient()
		controller, _ := NewStreamController(client, nil)

		_, err := controller.EndBroadcast(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty broadcast ID")
		}
	})
}

func TestStreamController_GetStreamHealth(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := LiveStreamListResponse{
				Items: []*LiveStream{{
					ID: "stream123",
					Status: &StreamStatus{
						StreamStatus: StreamStatusActive,
						HealthStatus: &StreamHealthStatus{
							Status: StreamHealthGood,
						},
					},
				}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		controller, _ := NewStreamController(client, nil)

		health, err := controller.GetStreamHealth(context.Background(), "stream123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if health.Status != StreamHealthGood {
			t.Errorf("unexpected health status: %s", health.Status)
		}
	})

	t.Run("empty stream ID", func(t *testing.T) {
		client := core.NewClient()
		controller, _ := NewStreamController(client, nil)

		_, err := controller.GetStreamHealth(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty stream ID")
		}
	})

	t.Run("no health status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := LiveStreamListResponse{
				Items: []*LiveStream{{
					ID:     "stream123",
					Status: &StreamStatus{StreamStatus: StreamStatusCreated},
				}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		controller, _ := NewStreamController(client, nil)

		_, err := controller.GetStreamHealth(context.Background(), "stream123")
		if err == nil {
			t.Fatal("expected error for no health status")
		}
	})
}

func TestStreamController_GetBroadcast(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := LiveBroadcastListResponse{
				Items: []*LiveBroadcast{{
					ID:      "broadcast123",
					Snippet: &BroadcastSnippet{Title: "Test"},
				}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		controller, _ := NewStreamController(client, nil)

		broadcast, err := controller.GetBroadcast(context.Background(), "broadcast123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if broadcast.ID != "broadcast123" {
			t.Errorf("unexpected ID: %s", broadcast.ID)
		}
	})

	t.Run("empty broadcast ID", func(t *testing.T) {
		client := core.NewClient()
		controller, _ := NewStreamController(client, nil)

		_, err := controller.GetBroadcast(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty broadcast ID")
		}
	})
}

func TestStreamController_GetStream(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := LiveStreamListResponse{
				Items: []*LiveStream{{
					ID:      "stream123",
					Snippet: &StreamSnippet{Title: "Test"},
				}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		controller, _ := NewStreamController(client, nil)

		stream, err := controller.GetStream(context.Background(), "stream123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if stream.ID != "stream123" {
			t.Errorf("unexpected ID: %s", stream.ID)
		}
	})

	t.Run("empty stream ID", func(t *testing.T) {
		client := core.NewClient()
		controller, _ := NewStreamController(client, nil)

		_, err := controller.GetStream(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty stream ID")
		}
	})
}

func TestStreamController_DeleteBroadcast(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		controller, _ := NewStreamController(client, nil)

		err := controller.DeleteBroadcast(context.Background(), "broadcast123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty broadcast ID", func(t *testing.T) {
		client := core.NewClient()
		controller, _ := NewStreamController(client, nil)

		err := controller.DeleteBroadcast(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty broadcast ID")
		}
	})
}

func TestStreamController_DeleteStream(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		controller, _ := NewStreamController(client, nil)

		err := controller.DeleteStream(context.Background(), "stream123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty stream ID", func(t *testing.T) {
		client := core.NewClient()
		controller, _ := NewStreamController(client, nil)

		err := controller.DeleteStream(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty stream ID")
		}
	})
}

func TestStreamController_TokenRefresh(t *testing.T) {
	t.Run("refreshes token on each call", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := LiveBroadcastListResponse{
				Items: []*LiveBroadcast{{ID: "broadcast123"}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := core.NewClient(core.WithBaseURL(server.URL))
		provider := &controllerMockTokenProvider{token: "test-token"}
		controller, _ := NewStreamController(client, provider)

		_, _ = controller.GetBroadcast(context.Background(), "broadcast123")
		_, _ = controller.GetBroadcast(context.Background(), "broadcast123")

		if provider.calls != 2 {
			t.Errorf("expected 2 token refresh calls, got %d", provider.calls)
		}
	})
}
