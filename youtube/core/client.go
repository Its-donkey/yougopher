package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the base URL for YouTube Data API v3.
	DefaultBaseURL = "https://www.googleapis.com/youtube/v3"

	// DefaultTimeout is the default HTTP request timeout.
	DefaultTimeout = 30 * time.Second

	// DefaultUserAgent is the default User-Agent header.
	DefaultUserAgent = "Yougopher/1.0"
)

// Client is an HTTP client for the YouTube API.
type Client struct {
	httpClient   *http.Client
	baseURL      string
	userAgent    string
	quotaTracker *QuotaTracker
	accessToken  string
	apiKey       string
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// NewClient creates a new YouTube API client.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: DefaultTimeout},
		baseURL:    DefaultBaseURL,
		userAgent:  DefaultUserAgent,
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

// WithBaseURL sets a custom base URL (useful for testing).
func WithBaseURL(url string) ClientOption {
	return func(c *Client) { c.baseURL = strings.TrimSuffix(url, "/") }
}

// WithUserAgent sets a custom User-Agent header.
func WithUserAgent(ua string) ClientOption {
	return func(c *Client) { c.userAgent = ua }
}

// WithQuotaTracker sets a quota tracker for the client.
func WithQuotaTracker(qt *QuotaTracker) ClientOption {
	return func(c *Client) { c.quotaTracker = qt }
}

// WithAccessToken sets the OAuth access token for authentication.
func WithAccessToken(token string) ClientOption {
	return func(c *Client) { c.accessToken = token }
}

// WithAPIKey sets the API key for authentication.
func WithAPIKey(key string) ClientOption {
	return func(c *Client) { c.apiKey = key }
}

// SetAccessToken updates the access token (for token refresh).
func (c *Client) SetAccessToken(token string) {
	c.accessToken = token
}

// QuotaTracker returns the client's quota tracker, if any.
func (c *Client) QuotaTracker() *QuotaTracker {
	return c.quotaTracker
}

// Request represents an API request.
type Request struct {
	Method    string
	Path      string
	Query     url.Values
	Body      any
	Operation string // For quota tracking (e.g., "videos.list")
}

// Do executes an HTTP request and decodes the response.
func (c *Client) Do(ctx context.Context, req *Request, result any) error {
	httpReq, err := c.newRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Track quota usage
	if c.quotaTracker != nil && req.Operation != "" {
		c.quotaTracker.Add(req.Operation, 1)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		return c.handleErrorResponse(resp.StatusCode, body)
	}

	// Decode successful response
	if result != nil && len(body) > 0 {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, query url.Values, operation string, result any) error {
	return c.Do(ctx, &Request{
		Method:    http.MethodGet,
		Path:      path,
		Query:     query,
		Operation: operation,
	}, result)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, query url.Values, body any, operation string, result any) error {
	return c.Do(ctx, &Request{
		Method:    http.MethodPost,
		Path:      path,
		Query:     query,
		Body:      body,
		Operation: operation,
	}, result)
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, path string, query url.Values, body any, operation string, result any) error {
	return c.Do(ctx, &Request{
		Method:    http.MethodPut,
		Path:      path,
		Query:     query,
		Body:      body,
		Operation: operation,
	}, result)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string, query url.Values, operation string) error {
	return c.Do(ctx, &Request{
		Method:    http.MethodDelete,
		Path:      path,
		Query:     query,
		Operation: operation,
	}, nil)
}

// newRequest creates an HTTP request.
func (c *Client) newRequest(ctx context.Context, req *Request) (*http.Request, error) {
	// Build URL
	u, err := url.Parse(c.baseURL + "/" + strings.TrimPrefix(req.Path, "/"))
	if err != nil {
		return nil, err
	}

	// Add query parameters
	query := u.Query()
	for k, v := range req.Query {
		for _, vv := range v {
			query.Add(k, vv)
		}
	}

	// Add API key if set and no access token
	if c.apiKey != "" && c.accessToken == "" {
		query.Set("key", c.apiKey)
	}

	u.RawQuery = query.Encode()

	// Prepare body
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("encoding body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, u.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	// Set headers
	httpReq.Header.Set("User-Agent", c.userAgent)

	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	if c.accessToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	return httpReq, nil
}

// handleErrorResponse parses an error response from the API.
func (c *Client) handleErrorResponse(statusCode int, body []byte) error {
	// Try to parse as YouTube API error
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != nil {
		apiErr := errResp.ToAPIError()
		apiErr.StatusCode = statusCode

		// Check for specific error types
		if apiErr.IsQuotaExceeded() {
			return &QuotaError{
				Used:    c.quotaUsed(),
				Limit:   c.quotaLimit(),
				ResetAt: nextPacificMidnight(),
			}
		}

		// Check for rate limiting
		if statusCode == 429 || apiErr.Code == "rateLimitExceeded" {
			return &RateLimitError{RetryAfter: 1 * time.Second}
		}

		return apiErr
	}

	// Fallback to generic error
	return &APIError{
		StatusCode: statusCode,
		Message:    string(body),
	}
}

// quotaUsed returns current quota usage or 0 if not tracked.
func (c *Client) quotaUsed() int {
	if c.quotaTracker != nil {
		return c.quotaTracker.Used()
	}
	return 0
}

// quotaLimit returns quota limit or default if not tracked.
func (c *Client) quotaLimit() int {
	if c.quotaTracker != nil {
		return c.quotaTracker.Limit()
	}
	return DefaultDailyQuota
}
