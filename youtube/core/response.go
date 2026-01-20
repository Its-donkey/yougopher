package core

import (
	"encoding/json"
	"time"
)

// Response wraps a YouTube API response with pagination info.
type Response[T any] struct {
	// Items contains the response data.
	Items []T `json:"items,omitempty"`

	// Kind identifies the resource type (e.g., "youtube#videoListResponse").
	Kind string `json:"kind,omitempty"`

	// ETag for caching and conditional requests.
	ETag string `json:"etag,omitempty"`

	// PageInfo contains paging details.
	PageInfo *PageInfo `json:"pageInfo,omitempty"`

	// NextPageToken for fetching the next page of results.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// PrevPageToken for fetching the previous page of results.
	PrevPageToken string `json:"prevPageToken,omitempty"`

	// PollingIntervalMillis is returned by liveChatMessages.list
	// indicating how long to wait before polling again.
	PollingIntervalMillis int `json:"pollingIntervalMillis,omitempty"`
}

// PageInfo contains pagination metadata.
type PageInfo struct {
	TotalResults   int `json:"totalResults,omitempty"`
	ResultsPerPage int `json:"resultsPerPage,omitempty"`
}

// PollingInterval returns the recommended polling interval as a Duration.
// Returns 0 if no polling interval was specified.
func (r *Response[T]) PollingInterval() time.Duration {
	if r.PollingIntervalMillis > 0 {
		return time.Duration(r.PollingIntervalMillis) * time.Millisecond
	}
	return 0
}

// HasNextPage returns true if there are more pages available.
func (r *Response[T]) HasNextPage() bool {
	return r.NextPageToken != ""
}

// ErrorResponse represents a YouTube API error response body.
type ErrorResponse struct {
	Error *ErrorBody `json:"error,omitempty"`
}

// ErrorBody contains the error details from the API.
type ErrorBody struct {
	Code    int           `json:"code,omitempty"`
	Message string        `json:"message,omitempty"`
	Status  string        `json:"status,omitempty"`
	Errors  []ErrorItem   `json:"errors,omitempty"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorItem represents a single error in the errors array.
type ErrorItem struct {
	Message      string `json:"message,omitempty"`
	Domain       string `json:"domain,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Location     string `json:"location,omitempty"`
	LocationType string `json:"locationType,omitempty"`
}

// ToAPIError converts an ErrorResponse to an APIError.
func (e *ErrorResponse) ToAPIError() *APIError {
	if e.Error == nil {
		return &APIError{Message: "unknown error"}
	}

	apiErr := &APIError{
		StatusCode: e.Error.Code,
		Message:    e.Error.Message,
		Details:    e.Error.Details,
	}

	// Extract the primary error code from the errors array
	if len(e.Error.Errors) > 0 {
		apiErr.Code = e.Error.Errors[0].Reason
	}

	return apiErr
}

// Thumbnail represents a video/channel thumbnail image.
type Thumbnail struct {
	URL    string `json:"url,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// Thumbnails contains different sized thumbnails.
type Thumbnails struct {
	Default  *Thumbnail `json:"default,omitempty"`
	Medium   *Thumbnail `json:"medium,omitempty"`
	High     *Thumbnail `json:"high,omitempty"`
	Standard *Thumbnail `json:"standard,omitempty"`
	Maxres   *Thumbnail `json:"maxres,omitempty"`
}

// Localized contains localized text.
type Localized struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// RawJSON holds raw JSON that can be unmarshaled later.
type RawJSON json.RawMessage

// MarshalJSON implements json.Marshaler.
func (r RawJSON) MarshalJSON() ([]byte, error) {
	if r == nil {
		return []byte("null"), nil
	}
	return r, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *RawJSON) UnmarshalJSON(data []byte) error {
	if r == nil {
		return nil
	}
	*r = append((*r)[0:0], data...)
	return nil
}
