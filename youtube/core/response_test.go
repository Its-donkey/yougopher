package core

import (
	"encoding/json"
	"testing"
	"time"
)

func TestResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"kind": "youtube#videoListResponse",
		"etag": "abc123",
		"nextPageToken": "CAUQAA",
		"prevPageToken": "CAUQAQ",
		"pageInfo": {
			"totalResults": 100,
			"resultsPerPage": 25
		},
		"items": [
			{"id": "video1"},
			{"id": "video2"}
		]
	}`

	type Item struct {
		ID string `json:"id"`
	}

	var resp Response[Item]
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if resp.Kind != "youtube#videoListResponse" {
		t.Errorf("Kind = %q, want %q", resp.Kind, "youtube#videoListResponse")
	}
	if resp.ETag != "abc123" {
		t.Errorf("ETag = %q, want %q", resp.ETag, "abc123")
	}
	if resp.NextPageToken != "CAUQAA" {
		t.Errorf("NextPageToken = %q, want %q", resp.NextPageToken, "CAUQAA")
	}
	if resp.PrevPageToken != "CAUQAQ" {
		t.Errorf("PrevPageToken = %q, want %q", resp.PrevPageToken, "CAUQAQ")
	}
	if resp.PageInfo.TotalResults != 100 {
		t.Errorf("PageInfo.TotalResults = %d, want 100", resp.PageInfo.TotalResults)
	}
	if resp.PageInfo.ResultsPerPage != 25 {
		t.Errorf("PageInfo.ResultsPerPage = %d, want 25", resp.PageInfo.ResultsPerPage)
	}
	if len(resp.Items) != 2 {
		t.Errorf("len(Items) = %d, want 2", len(resp.Items))
	}
	if resp.Items[0].ID != "video1" {
		t.Errorf("Items[0].ID = %q, want %q", resp.Items[0].ID, "video1")
	}
}

func TestResponse_EmptyItems(t *testing.T) {
	jsonData := `{
		"kind": "youtube#searchListResponse",
		"pageInfo": {
			"totalResults": 0,
			"resultsPerPage": 0
		},
		"items": []
	}`

	type Item struct {
		ID string `json:"id"`
	}

	var resp Response[Item]
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if len(resp.Items) != 0 {
		t.Errorf("len(Items) = %d, want 0", len(resp.Items))
	}
}

func TestErrorResponse_ToAPIError(t *testing.T) {
	errResp := ErrorResponse{
		Error: &ErrorBody{
			Code:    403,
			Message: "Quota exceeded",
			Errors: []ErrorItem{
				{
					Message: "The request cannot be completed because you have exceeded your quota.",
					Domain:  "youtube.quota",
					Reason:  "quotaExceeded",
				},
			},
		},
	}

	apiErr := errResp.ToAPIError()

	if apiErr.StatusCode != 403 {
		t.Errorf("StatusCode = %d, want 403", apiErr.StatusCode)
	}
	if apiErr.Message != "Quota exceeded" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Quota exceeded")
	}
	if apiErr.Code != "quotaExceeded" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "quotaExceeded")
	}
}

func TestErrorResponse_ToAPIError_NoErrors(t *testing.T) {
	errResp := ErrorResponse{
		Error: &ErrorBody{
			Code:    400,
			Message: "Bad Request",
			Errors:  []ErrorItem{},
		},
	}

	apiErr := errResp.ToAPIError()

	if apiErr.Code != "" {
		t.Errorf("Code = %q, want empty", apiErr.Code)
	}
}

func TestErrorResponse_ToAPIError_NilError(t *testing.T) {
	errResp := ErrorResponse{
		Error: nil,
	}

	apiErr := errResp.ToAPIError()

	if apiErr.StatusCode != 0 {
		t.Errorf("StatusCode = %d, want 0", apiErr.StatusCode)
	}
	if apiErr.Message != "unknown error" {
		t.Errorf("Message = %q, want 'unknown error'", apiErr.Message)
	}
}

func TestThumbnails_Unmarshal(t *testing.T) {
	jsonData := `{
		"default": {
			"url": "https://i.ytimg.com/vi/abc/default.jpg",
			"width": 120,
			"height": 90
		},
		"medium": {
			"url": "https://i.ytimg.com/vi/abc/mqdefault.jpg",
			"width": 320,
			"height": 180
		},
		"high": {
			"url": "https://i.ytimg.com/vi/abc/hqdefault.jpg",
			"width": 480,
			"height": 360
		},
		"standard": {
			"url": "https://i.ytimg.com/vi/abc/sddefault.jpg",
			"width": 640,
			"height": 480
		},
		"maxres": {
			"url": "https://i.ytimg.com/vi/abc/maxresdefault.jpg",
			"width": 1280,
			"height": 720
		}
	}`

	var thumbs Thumbnails
	if err := json.Unmarshal([]byte(jsonData), &thumbs); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if thumbs.Default.URL != "https://i.ytimg.com/vi/abc/default.jpg" {
		t.Errorf("Default.URL = %q, want correct URL", thumbs.Default.URL)
	}
	if thumbs.Default.Width != 120 {
		t.Errorf("Default.Width = %d, want 120", thumbs.Default.Width)
	}
	if thumbs.Default.Height != 90 {
		t.Errorf("Default.Height = %d, want 90", thumbs.Default.Height)
	}

	if thumbs.Medium.URL != "https://i.ytimg.com/vi/abc/mqdefault.jpg" {
		t.Errorf("Medium.URL incorrect")
	}
	if thumbs.High.URL != "https://i.ytimg.com/vi/abc/hqdefault.jpg" {
		t.Errorf("High.URL incorrect")
	}
	if thumbs.Standard.URL != "https://i.ytimg.com/vi/abc/sddefault.jpg" {
		t.Errorf("Standard.URL incorrect")
	}
	if thumbs.Maxres.URL != "https://i.ytimg.com/vi/abc/maxresdefault.jpg" {
		t.Errorf("Maxres.URL incorrect")
	}
	if thumbs.Maxres.Width != 1280 || thumbs.Maxres.Height != 720 {
		t.Errorf("Maxres dimensions incorrect")
	}
}

func TestPageInfo_Unmarshal(t *testing.T) {
	jsonData := `{"totalResults": 500, "resultsPerPage": 50}`

	var info PageInfo
	if err := json.Unmarshal([]byte(jsonData), &info); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if info.TotalResults != 500 {
		t.Errorf("TotalResults = %d, want 500", info.TotalResults)
	}
	if info.ResultsPerPage != 50 {
		t.Errorf("ResultsPerPage = %d, want 50", info.ResultsPerPage)
	}
}

func TestResponse_PollingInterval(t *testing.T) {
	type Item struct{}

	tests := []struct {
		name     string
		millis   int
		expected time.Duration
	}{
		{"positive", 5000, 5000 * time.Millisecond},
		{"zero", 0, 0},
		{"negative", -100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := Response[Item]{PollingIntervalMillis: tt.millis}
			got := resp.PollingInterval()
			if got != tt.expected {
				t.Errorf("PollingInterval() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestResponse_HasNextPage(t *testing.T) {
	type Item struct{}

	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{"with token", "CAUQAA", true},
		{"empty token", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := Response[Item]{NextPageToken: tt.token}
			got := resp.HasNextPage()
			if got != tt.expected {
				t.Errorf("HasNextPage() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRawJSON_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    RawJSON
		expected string
	}{
		{"nil", nil, "null"},
		{"object", RawJSON(`{"key":"value"}`), `{"key":"value"}`},
		{"array", RawJSON(`[1,2,3]`), `[1,2,3]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(got) != tt.expected {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestRawJSON_UnmarshalJSON(t *testing.T) {
	var raw RawJSON
	data := []byte(`{"test":"data"}`)

	if err := raw.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	if string(raw) != string(data) {
		t.Errorf("UnmarshalJSON() = %s, want %s", raw, data)
	}
}

func TestRawJSON_UnmarshalJSON_Nil(t *testing.T) {
	var raw *RawJSON
	data := []byte(`{"test":"data"}`)

	// Should not panic on nil receiver
	err := raw.UnmarshalJSON(data)
	if err != nil {
		t.Errorf("UnmarshalJSON() on nil should return nil, got %v", err)
	}
}
