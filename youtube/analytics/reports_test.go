package analytics

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client.analyticsURL != DefaultAnalyticsURL {
		t.Errorf("expected analyticsURL %q, got %q", DefaultAnalyticsURL, client.analyticsURL)
	}
	if client.httpClient == nil {
		t.Error("expected httpClient to be set")
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	customHTTP := &http.Client{Timeout: 60 * time.Second}
	customURL := "https://custom.example.com/analytics"
	token := "test-token"
	providerCalled := false

	client := NewClient(
		WithHTTPClient(customHTTP),
		WithAnalyticsURL(customURL),
		WithAccessToken(token),
		WithTokenProvider(func(ctx context.Context) (string, error) {
			providerCalled = true
			return "provider-token", nil
		}),
	)

	if client.httpClient != customHTTP {
		t.Error("expected custom HTTP client to be set")
	}
	if client.analyticsURL != customURL {
		t.Errorf("expected analyticsURL %q, got %q", customURL, client.analyticsURL)
	}
	if client.accessToken != token {
		t.Errorf("expected accessToken %q, got %q", token, client.accessToken)
	}

	// Token provider should take precedence
	tok, err := client.getAccessToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !providerCalled {
		t.Error("expected token provider to be called")
	}
	if tok != "provider-token" {
		t.Errorf("expected token %q, got %q", "provider-token", tok)
	}
}

func TestClient_Query(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Authorization header, got %q", r.Header.Get("Authorization"))
		}

		// Verify query parameters
		q := r.URL.Query()
		if q.Get("ids") != "channel==MINE" {
			t.Errorf("expected ids channel==MINE, got %s", q.Get("ids"))
		}
		if q.Get("startDate") != "2025-01-01" {
			t.Errorf("expected startDate 2025-01-01, got %s", q.Get("startDate"))
		}
		if q.Get("endDate") != "2025-01-31" {
			t.Errorf("expected endDate 2025-01-31, got %s", q.Get("endDate"))
		}
		if q.Get("metrics") != "views,estimatedMinutesWatched" {
			t.Errorf("expected metrics views,estimatedMinutesWatched, got %s", q.Get("metrics"))
		}
		if q.Get("dimensions") != "day" {
			t.Errorf("expected dimensions day, got %s", q.Get("dimensions"))
		}

		resp := map[string]any{
			"kind": "youtubeAnalytics#resultTable",
			"columnHeaders": []map[string]string{
				{"name": "day", "columnType": "DIMENSION", "dataType": "STRING"},
				{"name": "views", "columnType": "METRIC", "dataType": "INTEGER"},
				{"name": "estimatedMinutesWatched", "columnType": "METRIC", "dataType": "FLOAT"},
			},
			"rows": [][]any{
				{"2025-01-01", float64(1000), 5000.5},
				{"2025-01-02", float64(1500), 7500.25},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(
		WithAnalyticsURL(server.URL),
		WithAccessToken("test-token"),
	)

	report, err := client.Query(context.Background(), &QueryParams{
		IDs:        "channel==MINE",
		StartDate:  "2025-01-01",
		EndDate:    "2025-01-31",
		Metrics:    "views,estimatedMinutesWatched",
		Dimensions: "day",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Kind != "youtubeAnalytics#resultTable" {
		t.Errorf("expected Kind %q, got %q", "youtubeAnalytics#resultTable", report.Kind)
	}
	if len(report.ColumnHeaders) != 3 {
		t.Errorf("expected 3 column headers, got %d", len(report.ColumnHeaders))
	}
	if len(report.RawRows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(report.RawRows))
	}
}

func TestClient_Query_AllParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("filters") != "video==abc123" {
			t.Errorf("expected filters video==abc123, got %s", q.Get("filters"))
		}
		if q.Get("sort") != "-views" {
			t.Errorf("expected sort -views, got %s", q.Get("sort"))
		}
		if q.Get("maxResults") != "10" {
			t.Errorf("expected maxResults 10, got %s", q.Get("maxResults"))
		}
		if q.Get("startIndex") != "5" {
			t.Errorf("expected startIndex 5, got %s", q.Get("startIndex"))
		}
		if q.Get("includeHistoricalChannelData") != "true" {
			t.Errorf("expected includeHistoricalChannelData true, got %s", q.Get("includeHistoricalChannelData"))
		}
		if q.Get("currency") != "USD" {
			t.Errorf("expected currency USD, got %s", q.Get("currency"))
		}

		resp := map[string]any{
			"kind":          "youtubeAnalytics#resultTable",
			"columnHeaders": []map[string]string{},
			"rows":          [][]any{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(
		WithAnalyticsURL(server.URL),
		WithAccessToken("test-token"),
	)

	_, err := client.Query(context.Background(), &QueryParams{
		IDs:                          "channel==MINE",
		StartDate:                    "2025-01-01",
		EndDate:                      "2025-01-31",
		Metrics:                      "views",
		Filters:                      "video==abc123",
		Sort:                         "-views",
		MaxResults:                   10,
		StartIndex:                   5,
		IncludeHistoricalChannelData: true,
		Currency:                     "USD",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Query_ValidationErrors(t *testing.T) {
	client := NewClient(WithAccessToken("test-token"))

	tests := []struct {
		name   string
		params *QueryParams
		errMsg string
	}{
		{
			name:   "nil params",
			params: nil,
			errMsg: "query params cannot be nil",
		},
		{
			name:   "missing IDs",
			params: &QueryParams{StartDate: "2025-01-01", EndDate: "2025-01-31", Metrics: "views"},
			errMsg: "IDs parameter is required",
		},
		{
			name:   "missing StartDate",
			params: &QueryParams{IDs: "channel==MINE", EndDate: "2025-01-31", Metrics: "views"},
			errMsg: "StartDate parameter is required",
		},
		{
			name:   "missing EndDate",
			params: &QueryParams{IDs: "channel==MINE", StartDate: "2025-01-01", Metrics: "views"},
			errMsg: "EndDate parameter is required",
		},
		{
			name:   "missing Metrics",
			params: &QueryParams{IDs: "channel==MINE", StartDate: "2025-01-01", EndDate: "2025-01-31"},
			errMsg: "Metrics parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.Query(context.Background(), tt.params)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tt.errMsg {
				t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

func TestClient_Query_NoToken(t *testing.T) {
	client := NewClient()

	_, err := client.Query(context.Background(), &QueryParams{
		IDs:       "channel==MINE",
		StartDate: "2025-01-01",
		EndDate:   "2025-01-31",
		Metrics:   "views",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "no access token available" {
		t.Errorf("expected error 'no access token available', got %q", err.Error())
	}
}

func TestClient_Query_TokenProviderError(t *testing.T) {
	client := NewClient(
		WithTokenProvider(func(ctx context.Context) (string, error) {
			return "", errors.New("token error")
		}),
	)

	_, err := client.Query(context.Background(), &QueryParams{
		IDs:       "channel==MINE",
		StartDate: "2025-01-01",
		EndDate:   "2025-01-31",
		Metrics:   "views",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_Query_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		resp := map[string]any{
			"error": map[string]any{
				"code":    403,
				"message": "The request is not properly authorized.",
				"status":  "PERMISSION_DENIED",
				"errors": []map[string]string{
					{"domain": "youtube.analytics", "reason": "forbidden", "message": "Forbidden"},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(
		WithAnalyticsURL(server.URL),
		WithAccessToken("test-token"),
	)

	_, err := client.Query(context.Background(), &QueryParams{
		IDs:       "channel==MINE",
		StartDate: "2025-01-01",
		EndDate:   "2025-01-31",
		Metrics:   "views",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var analyticsErr *AnalyticsError
	if !errors.As(err, &analyticsErr) {
		t.Fatalf("expected AnalyticsError, got %T", err)
	}
	if analyticsErr.StatusCode != http.StatusForbidden {
		t.Errorf("expected StatusCode %d, got %d", http.StatusForbidden, analyticsErr.StatusCode)
	}
	if analyticsErr.Code != "PERMISSION_DENIED" {
		t.Errorf("expected Code %q, got %q", "PERMISSION_DENIED", analyticsErr.Code)
	}
	if analyticsErr.Reason != "forbidden" {
		t.Errorf("expected Reason %q, got %q", "forbidden", analyticsErr.Reason)
	}
	if !analyticsErr.IsPermissionDenied() {
		t.Error("expected IsPermissionDenied() true")
	}
}

func TestClient_QueryChannelViews(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("ids") != "channel==MINE" {
			t.Errorf("expected ids channel==MINE, got %s", q.Get("ids"))
		}
		if q.Get("metrics") != "views,estimatedMinutesWatched,averageViewDuration,subscribersGained,subscribersLost" {
			t.Errorf("unexpected metrics: %s", q.Get("metrics"))
		}

		resp := map[string]any{
			"kind":          "youtubeAnalytics#resultTable",
			"columnHeaders": []map[string]string{},
			"rows":          [][]any{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(
		WithAnalyticsURL(server.URL),
		WithAccessToken("test-token"),
	)

	_, err := client.QueryChannelViews(context.Background(), "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_QueryDailyViews(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("dimensions") != "day" {
			t.Errorf("expected dimensions day, got %s", q.Get("dimensions"))
		}
		if q.Get("sort") != "day" {
			t.Errorf("expected sort day, got %s", q.Get("sort"))
		}

		resp := map[string]any{
			"kind":          "youtubeAnalytics#resultTable",
			"columnHeaders": []map[string]string{},
			"rows":          [][]any{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(
		WithAnalyticsURL(server.URL),
		WithAccessToken("test-token"),
	)

	_, err := client.QueryDailyViews(context.Background(), "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_QueryTopVideos(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("dimensions") != "video" {
			t.Errorf("expected dimensions video, got %s", q.Get("dimensions"))
		}
		if q.Get("sort") != "-views" {
			t.Errorf("expected sort -views, got %s", q.Get("sort"))
		}
		if q.Get("maxResults") != "10" {
			t.Errorf("expected maxResults 10, got %s", q.Get("maxResults"))
		}

		resp := map[string]any{
			"kind":          "youtubeAnalytics#resultTable",
			"columnHeaders": []map[string]string{},
			"rows":          [][]any{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(
		WithAnalyticsURL(server.URL),
		WithAccessToken("test-token"),
	)

	_, err := client.QueryTopVideos(context.Background(), "2025-01-01", "2025-01-31", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_QueryCountryBreakdown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("dimensions") != "country" {
			t.Errorf("expected dimensions country, got %s", q.Get("dimensions"))
		}

		resp := map[string]any{
			"kind":          "youtubeAnalytics#resultTable",
			"columnHeaders": []map[string]string{},
			"rows":          [][]any{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(
		WithAnalyticsURL(server.URL),
		WithAccessToken("test-token"),
	)

	_, err := client.QueryCountryBreakdown(context.Background(), "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_QueryDeviceBreakdown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("dimensions") != "deviceType" {
			t.Errorf("expected dimensions deviceType, got %s", q.Get("dimensions"))
		}

		resp := map[string]any{
			"kind":          "youtubeAnalytics#resultTable",
			"columnHeaders": []map[string]string{},
			"rows":          [][]any{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(
		WithAnalyticsURL(server.URL),
		WithAccessToken("test-token"),
	)

	_, err := client.QueryDeviceBreakdown(context.Background(), "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_QueryRevenueReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("metrics") != "estimatedRevenue,estimatedAdRevenue,monetizedPlaybacks,cpm" {
			t.Errorf("unexpected metrics: %s", q.Get("metrics"))
		}

		resp := map[string]any{
			"kind":          "youtubeAnalytics#resultTable",
			"columnHeaders": []map[string]string{},
			"rows":          [][]any{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(
		WithAnalyticsURL(server.URL),
		WithAccessToken("test-token"),
	)

	_, err := client.QueryRevenueReport(context.Background(), "2025-01-01", "2025-01-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReport_Rows(t *testing.T) {
	report := &Report{
		ColumnHeaders: []ColumnHeader{
			{Name: "day", ColumnType: "DIMENSION", DataType: "STRING"},
			{Name: "views", ColumnType: "METRIC", DataType: "INTEGER"},
			{Name: "estimatedMinutesWatched", ColumnType: "METRIC", DataType: "FLOAT"},
		},
		RawRows: [][]any{
			{"2025-01-01", float64(1000), 5000.5},
			{"2025-01-02", float64(1500), 7500.25},
		},
	}

	rows := report.Rows()
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	// Check first row
	row := rows[0]
	if row.GetString("day") != "2025-01-01" {
		t.Errorf("expected day %q, got %q", "2025-01-01", row.GetString("day"))
	}
	if row.GetInt("views") != 1000 {
		t.Errorf("expected views %d, got %d", 1000, row.GetInt("views"))
	}
	if row.GetFloat("estimatedMinutesWatched") != 5000.5 {
		t.Errorf("expected estimatedMinutesWatched %f, got %f", 5000.5, row.GetFloat("estimatedMinutesWatched"))
	}

	// Check second row
	row = rows[1]
	if row.GetInt("views") != 1500 {
		t.Errorf("expected views %d, got %d", 1500, row.GetInt("views"))
	}
}

func TestReport_Rows_Empty(t *testing.T) {
	// Nil report
	var report *Report
	if report.Rows() != nil {
		t.Error("expected nil rows for nil report")
	}

	// Empty column headers
	report = &Report{
		ColumnHeaders: nil,
		RawRows:       [][]any{{"2025-01-01", float64(1000)}},
	}
	if report.Rows() != nil {
		t.Error("expected nil rows for empty headers")
	}

	// Empty raw rows
	report = &Report{
		ColumnHeaders: []ColumnHeader{{Name: "day"}},
		RawRows:       nil,
	}
	if report.Rows() != nil {
		t.Error("expected nil rows for empty raw rows")
	}
}

func TestReportRow_TypeConversion(t *testing.T) {
	row := ReportRow{
		Values: map[string]any{
			"string_val":  "hello",
			"float_val":   float64(123.45),
			"int_val":     int64(100),
			"int_val2":    int(200),
			"missing_val": nil,
		},
	}

	// String conversion
	if row.GetString("string_val") != "hello" {
		t.Errorf("expected string %q, got %q", "hello", row.GetString("string_val"))
	}
	if row.GetString("missing") != "" {
		t.Errorf("expected empty string for missing key")
	}

	// Int conversion
	if row.GetInt("float_val") != 123 {
		t.Errorf("expected int %d, got %d", 123, row.GetInt("float_val"))
	}
	if row.GetInt("int_val") != 100 {
		t.Errorf("expected int %d, got %d", 100, row.GetInt("int_val"))
	}
	if row.GetInt("int_val2") != 200 {
		t.Errorf("expected int %d, got %d", 200, row.GetInt("int_val2"))
	}
	if row.GetInt("missing") != 0 {
		t.Errorf("expected 0 for missing key")
	}

	// Float conversion
	if row.GetFloat("float_val") != 123.45 {
		t.Errorf("expected float %f, got %f", 123.45, row.GetFloat("float_val"))
	}
	if row.GetFloat("int_val") != 100.0 {
		t.Errorf("expected float %f, got %f", 100.0, row.GetFloat("int_val"))
	}
	if row.GetFloat("int_val2") != 200.0 {
		t.Errorf("expected float %f, got %f", 200.0, row.GetFloat("int_val2"))
	}
	if row.GetFloat("missing") != 0 {
		t.Errorf("expected 0 for missing key")
	}
}

func TestReport_TotalViews(t *testing.T) {
	report := &Report{
		ColumnHeaders: []ColumnHeader{
			{Name: "day", ColumnType: "DIMENSION"},
			{Name: "views", ColumnType: "METRIC"},
		},
		RawRows: [][]any{
			{"2025-01-01", float64(1000)},
			{"2025-01-02", float64(1500)},
			{"2025-01-03", float64(2000)},
		},
	}

	total := report.TotalViews()
	if total != 4500 {
		t.Errorf("expected total views %d, got %d", 4500, total)
	}
}

func TestReport_TotalMinutesWatched(t *testing.T) {
	report := &Report{
		ColumnHeaders: []ColumnHeader{
			{Name: "day", ColumnType: "DIMENSION"},
			{Name: "estimatedMinutesWatched", ColumnType: "METRIC"},
		},
		RawRows: [][]any{
			{"2025-01-01", 1000.5},
			{"2025-01-02", 1500.25},
			{"2025-01-03", 2000.75},
		},
	}

	total := report.TotalMinutesWatched()
	expected := 4501.5
	if total != expected {
		t.Errorf("expected total minutes %f, got %f", expected, total)
	}
}

func TestReport_SumMetric_NotFound(t *testing.T) {
	report := &Report{
		ColumnHeaders: []ColumnHeader{
			{Name: "day", ColumnType: "DIMENSION"},
		},
		RawRows: [][]any{{"2025-01-01"}},
	}

	total := report.TotalViews()
	if total != 0 {
		t.Errorf("expected 0 for missing metric, got %d", total)
	}
}

func TestAnalyticsError(t *testing.T) {
	tests := []struct {
		name                 string
		err                  *AnalyticsError
		expectedMsg          string
		isPermissionDenied   bool
		isInvalidArgument    bool
		isQuotaExceeded      bool
	}{
		{
			name: "permission denied with reason",
			err: &AnalyticsError{
				StatusCode: 403,
				Code:       "PERMISSION_DENIED",
				Reason:     "forbidden",
				Message:    "Access denied",
			},
			expectedMsg:        "analytics error: PERMISSION_DENIED (forbidden) - Access denied",
			isPermissionDenied: true,
		},
		{
			name: "permission denied by code only",
			err: &AnalyticsError{
				StatusCode: 403,
				Code:       "PERMISSION_DENIED",
				Message:    "Access denied",
			},
			expectedMsg:        "analytics error: PERMISSION_DENIED - Access denied",
			isPermissionDenied: true,
		},
		{
			name: "invalid argument",
			err: &AnalyticsError{
				StatusCode: 400,
				Code:       "INVALID_ARGUMENT",
				Reason:     "invalidParameter",
				Message:    "Invalid date format",
			},
			expectedMsg:       "analytics error: INVALID_ARGUMENT (invalidParameter) - Invalid date format",
			isInvalidArgument: true,
		},
		{
			name: "quota exceeded",
			err: &AnalyticsError{
				StatusCode: 429,
				Code:       "RESOURCE_EXHAUSTED",
				Reason:     "quotaExceeded",
				Message:    "Quota exceeded",
			},
			expectedMsg:     "analytics error: RESOURCE_EXHAUSTED (quotaExceeded) - Quota exceeded",
			isQuotaExceeded: true,
		},
		{
			name: "quota in message",
			err: &AnalyticsError{
				StatusCode: 429,
				Code:       "ERROR",
				Message:    "Daily quota limit exceeded",
			},
			expectedMsg:     "analytics error: ERROR - Daily quota limit exceeded",
			isQuotaExceeded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expectedMsg {
				t.Errorf("expected Error() %q, got %q", tt.expectedMsg, tt.err.Error())
			}
			if tt.err.IsPermissionDenied() != tt.isPermissionDenied {
				t.Errorf("expected IsPermissionDenied() %v, got %v", tt.isPermissionDenied, tt.err.IsPermissionDenied())
			}
			if tt.err.IsInvalidArgument() != tt.isInvalidArgument {
				t.Errorf("expected IsInvalidArgument() %v, got %v", tt.isInvalidArgument, tt.err.IsInvalidArgument())
			}
			if tt.err.IsQuotaExceeded() != tt.isQuotaExceeded {
				t.Errorf("expected IsQuotaExceeded() %v, got %v", tt.isQuotaExceeded, tt.err.IsQuotaExceeded())
			}
		})
	}
}

func TestQuery_PackageFunction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"kind":          "youtubeAnalytics#resultTable",
			"columnHeaders": []map[string]string{},
			"rows":          [][]any{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(
		WithAnalyticsURL(server.URL),
		WithAccessToken("test-token"),
	)

	report, err := Query(context.Background(), client, &QueryParams{
		IDs:       "channel==MINE",
		StartDate: "2025-01-01",
		EndDate:   "2025-01-31",
		Metrics:   "views",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report == nil {
		t.Error("expected report, got nil")
	}
}
