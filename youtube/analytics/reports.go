package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// YouTube Analytics API endpoint.
const DefaultAnalyticsURL = "https://youtubeanalytics.googleapis.com/v2/reports"

// Common metrics for analytics queries.
const (
	MetricViews                    = "views"
	MetricEstimatedMinutesWatched  = "estimatedMinutesWatched"
	MetricAverageViewDuration      = "averageViewDuration"
	MetricSubscribersGained        = "subscribersGained"
	MetricSubscribersLost          = "subscribersLost"
	MetricLikes                    = "likes"
	MetricDislikes                 = "dislikes"
	MetricComments                 = "comments"
	MetricShares                   = "shares"
	MetricAnnotationClickRate      = "annotationClickThroughRate"
	MetricAnnotationCloseRate      = "annotationCloseRate"
	MetricAverageViewPercentage    = "averageViewPercentage"
	MetricEstimatedRevenue         = "estimatedRevenue"
	MetricEstimatedAdRevenue       = "estimatedAdRevenue"
	MetricGrossRevenue             = "grossRevenue"
	MetricCPM                      = "cpm"
	MetricPlaybackBasedCPM         = "playbackBasedCpm"
	MetricAdImpressions            = "adImpressions"
	MetricMonetizedPlaybacks       = "monetizedPlaybacks"
	MetricCardClickRate            = "cardClickRate"
	MetricCardTeaserClickRate      = "cardTeaserClickRate"
	MetricRedViews                 = "redViews"
	MetricRedWatchedMinutes        = "redWatchedMinutes"
)

// Common dimensions for analytics queries.
const (
	DimensionDay                = "day"
	DimensionMonth              = "month"
	DimensionVideo              = "video"
	DimensionChannel            = "channel"
	DimensionCountry            = "country"
	DimensionProvince           = "province"
	DimensionCity               = "city"
	DimensionDeviceType         = "deviceType"
	DimensionOperatingSystem    = "operatingSystem"
	DimensionAgeGroup           = "ageGroup"
	DimensionGender             = "gender"
	DimensionSharingService     = "sharingService"
	DimensionPlaybackLocationType = "playbackLocationType"
	DimensionSubscribedStatus   = "subscribedStatus"
	DimensionLiveOrOnDemand     = "liveOrOnDemand"
	DimensionTrafficSourceType  = "trafficSourceType"
	DimensionTrafficSourceDetail = "trafficSourceDetail"
)

// Filter operators for analytics queries.
const (
	FilterEquals   = "=="
	FilterNotEqual = "!="
	FilterContains = "=@"
	FilterGreater  = ">"
	FilterLess     = "<"
)

// QueryParams contains parameters for an analytics query.
type QueryParams struct {
	// IDs specifies the channel or content owner.
	// Format: "channel==MINE" or "channel==UC1234" or "contentOwner==XYZ"
	IDs string

	// StartDate is the start of the date range (YYYY-MM-DD format).
	StartDate string

	// EndDate is the end of the date range (YYYY-MM-DD format).
	EndDate string

	// Metrics is a comma-separated list of metrics to retrieve.
	// Example: "views,estimatedMinutesWatched,averageViewDuration"
	Metrics string

	// Dimensions is a comma-separated list of dimensions to group by (optional).
	// Example: "day,video"
	Dimensions string

	// Filters is a semicolon-separated list of dimension filters (optional).
	// Example: "video==dQw4w9WgXcQ;country==US"
	Filters string

	// Sort is a comma-separated list of dimensions/metrics to sort by (optional).
	// Prefix with "-" for descending order.
	// Example: "-views,day"
	Sort string

	// MaxResults limits the number of rows returned (optional).
	MaxResults int

	// StartIndex is the 1-based index of the first row to retrieve (optional).
	StartIndex int

	// IncludeHistoricalChannelData includes data from before the channel
	// was linked to the content owner (optional).
	IncludeHistoricalChannelData bool

	// Currency specifies the currency for revenue metrics (optional).
	// Example: "USD", "EUR"
	Currency string
}

// Report represents an analytics report response.
type Report struct {
	// Kind is the resource type (youtubeAnalytics#resultTable).
	Kind string `json:"kind,omitempty"`

	// ColumnHeaders describes each column in the response.
	ColumnHeaders []ColumnHeader `json:"columnHeaders,omitempty"`

	// RawRows contains the data rows.
	// Each row is an array of values corresponding to the column headers.
	RawRows [][]any `json:"rows,omitempty"`
}

// ColumnHeader describes a column in the analytics report.
type ColumnHeader struct {
	// Name is the column name (dimension or metric name).
	Name string `json:"name"`

	// ColumnType is "DIMENSION" or "METRIC".
	ColumnType string `json:"columnType"`

	// DataType is the data type (STRING, INTEGER, FLOAT, etc.).
	DataType string `json:"dataType"`
}

// ReportRow represents a single row of analytics data with named fields.
type ReportRow struct {
	// Values maps column names to their values.
	Values map[string]any
}

// GetString returns a string value for the given column name.
func (r *ReportRow) GetString(name string) string {
	if v, ok := r.Values[name]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetInt returns an int64 value for the given column name.
func (r *ReportRow) GetInt(name string) int64 {
	if v, ok := r.Values[name]; ok {
		switch n := v.(type) {
		case float64:
			return int64(n)
		case int64:
			return n
		case int:
			return int64(n)
		}
	}
	return 0
}

// GetFloat returns a float64 value for the given column name.
func (r *ReportRow) GetFloat(name string) float64 {
	if v, ok := r.Values[name]; ok {
		switch n := v.(type) {
		case float64:
			return n
		case int64:
			return float64(n)
		case int:
			return float64(n)
		}
	}
	return 0
}

// Rows returns the report data as a slice of ReportRow for easier access.
func (r *Report) Rows() []ReportRow {
	if r == nil || len(r.ColumnHeaders) == 0 || len(r.RawRows) == 0 {
		return nil
	}

	rows := make([]ReportRow, len(r.RawRows))
	for i, row := range r.RawRows {
		rows[i] = ReportRow{Values: make(map[string]any)}
		for j, header := range r.ColumnHeaders {
			if j < len(row) {
				rows[i].Values[header.Name] = row[j]
			}
		}
	}
	return rows
}

// TotalViews returns the sum of the views metric across all rows.
func (r *Report) TotalViews() int64 {
	return r.sumMetric(MetricViews)
}

// TotalMinutesWatched returns the sum of estimated minutes watched.
func (r *Report) TotalMinutesWatched() float64 {
	return r.sumMetricFloat(MetricEstimatedMinutesWatched)
}

// sumMetric sums an integer metric across all rows.
func (r *Report) sumMetric(metric string) int64 {
	idx := r.metricIndex(metric)
	if idx < 0 {
		return 0
	}

	var total int64
	for _, row := range r.RawRows {
		if idx < len(row) {
			if v, ok := row[idx].(float64); ok {
				total += int64(v)
			}
		}
	}
	return total
}

// sumMetricFloat sums a float metric across all rows.
func (r *Report) sumMetricFloat(metric string) float64 {
	idx := r.metricIndex(metric)
	if idx < 0 {
		return 0
	}

	var total float64
	for _, row := range r.RawRows {
		if idx < len(row) {
			if v, ok := row[idx].(float64); ok {
				total += v
			}
		}
	}
	return total
}

// metricIndex returns the column index for a metric, or -1 if not found.
func (r *Report) metricIndex(metric string) int {
	for i, header := range r.ColumnHeaders {
		if header.Name == metric {
			return i
		}
	}
	return -1
}

// Client handles YouTube Analytics API requests.
type Client struct {
	httpClient   *http.Client
	analyticsURL string
	accessToken  string

	// TokenProvider is a function that returns a valid access token.
	// If set, it takes precedence over the static accessToken.
	tokenProvider func(context.Context) (string, error)
}

// ClientOption configures an analytics Client.
type ClientOption func(*Client)

// NewClient creates a new Analytics API client.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		analyticsURL: DefaultAnalyticsURL,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) { c.httpClient = hc }
}

// WithAccessToken sets a static access token.
func WithAccessToken(token string) ClientOption {
	return func(c *Client) { c.accessToken = token }
}

// WithTokenProvider sets a function to retrieve access tokens dynamically.
// This is useful for integration with auth.AuthClient.
func WithTokenProvider(provider func(context.Context) (string, error)) ClientOption {
	return func(c *Client) { c.tokenProvider = provider }
}

// WithAnalyticsURL sets a custom analytics API URL (for testing).
func WithAnalyticsURL(url string) ClientOption {
	return func(c *Client) { c.analyticsURL = url }
}

// Query executes an analytics query and returns the report.
func (c *Client) Query(ctx context.Context, params *QueryParams) (*Report, error) {
	if params == nil {
		return nil, fmt.Errorf("query params cannot be nil")
	}
	if params.IDs == "" {
		return nil, fmt.Errorf("ids parameter is required")
	}
	if params.StartDate == "" {
		return nil, fmt.Errorf("startDate parameter is required")
	}
	if params.EndDate == "" {
		return nil, fmt.Errorf("endDate parameter is required")
	}
	if params.Metrics == "" {
		return nil, fmt.Errorf("metrics parameter is required")
	}

	// Build query parameters
	query := url.Values{
		"ids":       {params.IDs},
		"startDate": {params.StartDate},
		"endDate":   {params.EndDate},
		"metrics":   {params.Metrics},
	}

	if params.Dimensions != "" {
		query.Set("dimensions", params.Dimensions)
	}
	if params.Filters != "" {
		query.Set("filters", params.Filters)
	}
	if params.Sort != "" {
		query.Set("sort", params.Sort)
	}
	if params.MaxResults > 0 {
		query.Set("maxResults", fmt.Sprintf("%d", params.MaxResults))
	}
	if params.StartIndex > 0 {
		query.Set("startIndex", fmt.Sprintf("%d", params.StartIndex))
	}
	if params.IncludeHistoricalChannelData {
		query.Set("includeHistoricalChannelData", "true")
	}
	if params.Currency != "" {
		query.Set("currency", params.Currency)
	}

	// Get access token
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting access token: %w", err)
	}
	if accessToken == "" {
		return nil, fmt.Errorf("no access token available")
	}

	// Build request
	reqURL := c.analyticsURL + "?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	body, err := c.readResponseBody(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp.StatusCode, body)
	}

	// Parse response
	var report Report
	if err := json.Unmarshal(body, &report); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &report, nil
}

// QueryChannelViews is a convenience method to get view statistics for a channel.
func (c *Client) QueryChannelViews(ctx context.Context, startDate, endDate string) (*Report, error) {
	return c.Query(ctx, &QueryParams{
		IDs:       "channel==MINE",
		StartDate: startDate,
		EndDate:   endDate,
		Metrics:   "views,estimatedMinutesWatched,averageViewDuration,subscribersGained,subscribersLost",
	})
}

// QueryDailyViews gets daily view breakdown for the channel.
func (c *Client) QueryDailyViews(ctx context.Context, startDate, endDate string) (*Report, error) {
	return c.Query(ctx, &QueryParams{
		IDs:        "channel==MINE",
		StartDate:  startDate,
		EndDate:    endDate,
		Metrics:    "views,estimatedMinutesWatched",
		Dimensions: "day",
		Sort:       "day",
	})
}

// QueryTopVideos gets the top videos by views.
func (c *Client) QueryTopVideos(ctx context.Context, startDate, endDate string, maxResults int) (*Report, error) {
	return c.Query(ctx, &QueryParams{
		IDs:        "channel==MINE",
		StartDate:  startDate,
		EndDate:    endDate,
		Metrics:    "views,estimatedMinutesWatched,likes,comments",
		Dimensions: "video",
		Sort:       "-views",
		MaxResults: maxResults,
	})
}

// QueryCountryBreakdown gets views broken down by country.
func (c *Client) QueryCountryBreakdown(ctx context.Context, startDate, endDate string) (*Report, error) {
	return c.Query(ctx, &QueryParams{
		IDs:        "channel==MINE",
		StartDate:  startDate,
		EndDate:    endDate,
		Metrics:    "views,estimatedMinutesWatched",
		Dimensions: "country",
		Sort:       "-views",
	})
}

// QueryDeviceBreakdown gets views broken down by device type.
func (c *Client) QueryDeviceBreakdown(ctx context.Context, startDate, endDate string) (*Report, error) {
	return c.Query(ctx, &QueryParams{
		IDs:        "channel==MINE",
		StartDate:  startDate,
		EndDate:    endDate,
		Metrics:    "views,estimatedMinutesWatched",
		Dimensions: "deviceType",
		Sort:       "-views",
	})
}

// QueryRevenueReport gets revenue metrics for the channel.
// Requires the youtube.readonly scope and channel monetization.
func (c *Client) QueryRevenueReport(ctx context.Context, startDate, endDate string) (*Report, error) {
	return c.Query(ctx, &QueryParams{
		IDs:        "channel==MINE",
		StartDate:  startDate,
		EndDate:    endDate,
		Metrics:    "estimatedRevenue,estimatedAdRevenue,monetizedPlaybacks,cpm",
		Dimensions: "day",
		Sort:       "day",
	})
}

// getAccessToken retrieves the access token from provider or static value.
func (c *Client) getAccessToken(ctx context.Context) (string, error) {
	if c.tokenProvider != nil {
		return c.tokenProvider(ctx)
	}
	return c.accessToken, nil
}

// readResponseBody reads and limits response body size.
func (c *Client) readResponseBody(body io.Reader) ([]byte, error) {
	const maxResponseSize = 10 * 1024 * 1024 // 10MB for analytics reports
	limitedReader := io.LimitReader(body, maxResponseSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if len(data) > maxResponseSize {
		return nil, fmt.Errorf("response body exceeds maximum size of %d bytes", maxResponseSize)
	}
	return data, nil
}

// parseError parses an error response from the Analytics API.
func (c *Client) parseError(statusCode int, body []byte) error {
	var errResp struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
			Errors  []struct {
				Domain  string `json:"domain"`
				Reason  string `json:"reason"`
				Message string `json:"message"`
			} `json:"errors"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
		reason := ""
		if len(errResp.Error.Errors) > 0 {
			reason = errResp.Error.Errors[0].Reason
		}
		return &AnalyticsError{
			StatusCode: statusCode,
			Code:       errResp.Error.Status,
			Reason:     reason,
			Message:    errResp.Error.Message,
		}
	}

	return fmt.Errorf("analytics API error (status %d): %s", statusCode, string(body))
}

// AnalyticsError represents an error from the YouTube Analytics API.
type AnalyticsError struct {
	StatusCode int
	Code       string // e.g., "PERMISSION_DENIED", "INVALID_ARGUMENT"
	Reason     string // e.g., "forbidden", "invalidParameter"
	Message    string
}

// Error implements the error interface.
func (e *AnalyticsError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("analytics error: %s (%s) - %s", e.Code, e.Reason, e.Message)
	}
	return fmt.Sprintf("analytics error: %s - %s", e.Code, e.Message)
}

// IsPermissionDenied returns true if the error is a permission denied error.
func (e *AnalyticsError) IsPermissionDenied() bool {
	return e.Code == "PERMISSION_DENIED" || e.Reason == "forbidden"
}

// IsInvalidArgument returns true if the error is an invalid argument error.
func (e *AnalyticsError) IsInvalidArgument() bool {
	return e.Code == "INVALID_ARGUMENT" || e.Reason == "invalidParameter"
}

// IsQuotaExceeded returns true if the error is a quota exceeded error.
func (e *AnalyticsError) IsQuotaExceeded() bool {
	return e.Reason == "quotaExceeded" || strings.Contains(e.Message, "quota")
}

// Query is a package-level convenience function for one-off queries.
func Query(ctx context.Context, client *Client, params *QueryParams) (*Report, error) {
	return client.Query(ctx, params)
}
