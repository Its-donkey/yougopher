package core

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient()

	if c.baseURL != DefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, DefaultBaseURL)
	}
	if c.userAgent != DefaultUserAgent {
		t.Errorf("userAgent = %q, want %q", c.userAgent, DefaultUserAgent)
	}
	if c.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	customHTTP := &http.Client{Timeout: 60 * time.Second}
	qt := NewQuotaTracker(5000)

	c := NewClient(
		WithHTTPClient(customHTTP),
		WithBaseURL("https://example.com/api/"),
		WithUserAgent("TestAgent/1.0"),
		WithQuotaTracker(qt),
		WithAccessToken("test-token"),
		WithAPIKey("test-api-key"),
	)

	if c.httpClient != customHTTP {
		t.Error("httpClient not set correctly")
	}
	if c.baseURL != "https://example.com/api" { // Trailing slash removed
		t.Errorf("baseURL = %q, want %q", c.baseURL, "https://example.com/api")
	}
	if c.userAgent != "TestAgent/1.0" {
		t.Errorf("userAgent = %q, want %q", c.userAgent, "TestAgent/1.0")
	}
	if c.quotaTracker != qt {
		t.Error("quotaTracker not set correctly")
	}
	if c.accessToken != "test-token" {
		t.Errorf("accessToken = %q, want %q", c.accessToken, "test-token")
	}
	if c.apiKey != "test-api-key" {
		t.Errorf("apiKey = %q, want %q", c.apiKey, "test-api-key")
	}
}

func TestClient_SetAccessToken(t *testing.T) {
	c := NewClient()
	c.SetAccessToken("new-token")

	if c.accessToken != "new-token" {
		t.Errorf("accessToken = %q, want %q", c.accessToken, "new-token")
	}
}

func TestClient_QuotaTracker(t *testing.T) {
	qt := NewQuotaTracker(10000)
	c := NewClient(WithQuotaTracker(qt))

	if c.QuotaTracker() != qt {
		t.Error("QuotaTracker() should return the configured tracker")
	}

	c2 := NewClient()
	if c2.QuotaTracker() != nil {
		t.Error("QuotaTracker() should return nil when not configured")
	}
}

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/videos" {
			t.Errorf("Path = %q, want /videos", r.URL.Path)
		}
		if r.URL.Query().Get("id") != "abc123" {
			t.Errorf("Query id = %q, want abc123", r.URL.Query().Get("id"))
		}
		if r.Header.Get("User-Agent") != "TestAgent" {
			t.Errorf("User-Agent = %q, want TestAgent", r.Header.Get("User-Agent"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"kind":  "youtube#videoListResponse",
			"items": []map[string]string{{"id": "abc123"}},
		})
	}))
	defer server.Close()

	c := NewClient(
		WithBaseURL(server.URL),
		WithUserAgent("TestAgent"),
	)

	var result map[string]any
	err := c.Get(context.Background(), "/videos", url.Values{"id": {"abc123"}}, "videos.list", &result)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if result["kind"] != "youtube#videoListResponse" {
		t.Errorf("result[kind] = %v, want youtube#videoListResponse", result["kind"])
	}
}

func TestClient_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}

		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["message"] != "Hello" {
			t.Errorf("body[message] = %q, want Hello", body["message"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := NewClient(WithBaseURL(server.URL))

	var result map[string]string
	err := c.Post(context.Background(), "/messages", nil, map[string]string{"message": "Hello"}, "liveChatMessages.insert", &result)
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("result[status] = %v, want ok", result["status"])
	}
}

func TestClient_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Method = %q, want PUT", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"updated": "true"})
	}))
	defer server.Close()

	c := NewClient(WithBaseURL(server.URL))

	var result map[string]string
	err := c.Put(context.Background(), "/resource", nil, map[string]string{"field": "value"}, "videos.update", &result)
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
}

func TestClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Method = %q, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(WithBaseURL(server.URL))

	err := c.Delete(context.Background(), "/resource", url.Values{"id": {"123"}}, "videos.delete")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestClient_AuthorizationHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer my-access-token" {
			t.Errorf("Authorization = %q, want Bearer my-access-token", auth)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(
		WithBaseURL(server.URL),
		WithAccessToken("my-access-token"),
	)

	err := c.Get(context.Background(), "/test", nil, "", nil)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
}

func TestClient_APIKeyInQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key != "my-api-key" {
			t.Errorf("Query key = %q, want my-api-key", key)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(
		WithBaseURL(server.URL),
		WithAPIKey("my-api-key"),
	)

	err := c.Get(context.Background(), "/test", nil, "", nil)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
}

func TestClient_APIKeyNotAddedWithAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key != "" {
			t.Errorf("Query key should be empty when access token is set, got %q", key)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(
		WithBaseURL(server.URL),
		WithAPIKey("my-api-key"),
		WithAccessToken("my-access-token"),
	)

	err := c.Get(context.Background(), "/test", nil, "", nil)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
}

func TestClient_QuotaTracking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	qt := NewQuotaTracker(10000)
	c := NewClient(
		WithBaseURL(server.URL),
		WithQuotaTracker(qt),
	)

	// videos.list costs 1 quota
	_ = c.Get(context.Background(), "/videos", nil, "videos.list", nil)
	if qt.Used() != 1 {
		t.Errorf("Used() = %d, want 1", qt.Used())
	}

	// search.list costs 100 quota
	_ = c.Get(context.Background(), "/search", nil, "search.list", nil)
	if qt.Used() != 101 {
		t.Errorf("Used() = %d, want 101", qt.Used())
	}
}

func TestClient_ErrorResponse_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    400,
				"message": "Invalid parameter",
				"errors": []map[string]string{
					{"reason": "invalidParameter", "message": "Invalid parameter", "domain": "youtube.parameter"},
				},
			},
		})
	}))
	defer server.Close()

	c := NewClient(WithBaseURL(server.URL))

	err := c.Get(context.Background(), "/test", nil, "", nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
	if apiErr.Code != "invalidParameter" {
		t.Errorf("Code = %q, want invalidParameter", apiErr.Code)
	}
}

func TestClient_ErrorResponse_QuotaExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    403,
				"message": "Quota exceeded",
				"errors": []map[string]string{
					{"reason": "quotaExceeded", "message": "Quota exceeded", "domain": "youtube.quota"},
				},
			},
		})
	}))
	defer server.Close()

	c := NewClient(WithBaseURL(server.URL))

	err := c.Get(context.Background(), "/test", nil, "", nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	quotaErr, ok := err.(*QuotaError)
	if !ok {
		t.Fatalf("Expected *QuotaError, got %T: %v", err, err)
	}
	if quotaErr.ResetAt.IsZero() {
		t.Error("ResetAt should not be zero")
	}
}

func TestClient_ErrorResponse_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    429,
				"message": "Rate limit exceeded",
				"errors": []map[string]string{
					{"reason": "rateLimitExceeded", "message": "Rate limit exceeded"},
				},
			},
		})
	}))
	defer server.Close()

	c := NewClient(WithBaseURL(server.URL))

	err := c.Get(context.Background(), "/test", nil, "", nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	rateLimitErr, ok := err.(*RateLimitError)
	if !ok {
		t.Fatalf("Expected *RateLimitError, got %T: %v", err, err)
	}
	if rateLimitErr.RetryAfter == 0 {
		t.Error("RetryAfter should not be zero")
	}
}

func TestClient_ErrorResponse_GenericError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	c := NewClient(WithBaseURL(server.URL))

	err := c.Get(context.Background(), "/test", nil, "", nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
	}
	if apiErr.Message != "Internal Server Error" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Internal Server Error")
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(WithBaseURL(server.URL))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := c.Get(ctx, "/test", nil, "", nil)
	if err == nil {
		t.Fatal("Expected error due to cancelled context")
	}
}

func TestClient_quotaUsed_NilTracker(t *testing.T) {
	c := NewClient()
	if c.quotaUsed() != 0 {
		t.Errorf("quotaUsed() = %d, want 0", c.quotaUsed())
	}
}

func TestClient_quotaLimit_NilTracker(t *testing.T) {
	c := NewClient()
	if c.quotaLimit() != DefaultDailyQuota {
		t.Errorf("quotaLimit() = %d, want %d", c.quotaLimit(), DefaultDailyQuota)
	}
}

func TestClient_quotaUsed_WithTracker(t *testing.T) {
	qt := NewQuotaTracker(10000)
	qt.AddCost(500)

	c := NewClient(WithQuotaTracker(qt))
	if c.quotaUsed() != 500 {
		t.Errorf("quotaUsed() = %d, want 500", c.quotaUsed())
	}
}

func TestClient_quotaLimit_WithTracker(t *testing.T) {
	qt := NewQuotaTracker(5000)
	c := NewClient(WithQuotaTracker(qt))

	if c.quotaLimit() != 5000 {
		t.Errorf("quotaLimit() = %d, want 5000", c.quotaLimit())
	}
}
