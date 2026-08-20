package client

import (
	"errors"
	"fmt"
)

// BarikoiError is returned for any non-2xx HTTP response from the Barikoi API.
// Message is taken from the server's JSON body when present.
type BarikoiError struct {
	Message    string
	StatusCode int
	Code       string
	Details    any
}

func (e *BarikoiError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("barikoi: %s (status %d, code %s)", e.Message, e.StatusCode, e.Code)
	}
	return fmt.Sprintf("barikoi: %s (status %d)", e.Message, e.StatusCode)
}

// IsAuthError reports whether the API rejected the API key (HTTP 401 or 403).
func (e *BarikoiError) IsAuthError() bool {
	return e.StatusCode == 401 || e.StatusCode == 403
}

// IsRateLimitError reports whether the API rate limit was exceeded (HTTP 429).
func (e *BarikoiError) IsRateLimitError() bool {
	return e.StatusCode == 429
}

// IsServerError reports whether the Barikoi server itself failed (HTTP 5xx).
func (e *BarikoiError) IsServerError() bool {
	return e.StatusCode >= 500
}

// ValidationError is returned when a request fails client-side validation,
// before any HTTP call is made.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("barikoi: validation error on field %q: %s", e.Field, e.Message)
}

// TimeoutError is returned when a request is cancelled or times out.
type TimeoutError struct {
	Message string
}

func (e *TimeoutError) Error() string { return e.Message }

// Sentinel errors returned by the client.
var (
	// ErrMissingAPIKey is returned by NewClient when the API key is empty.
	ErrMissingAPIKey = errors.New("barikoi: API key is required")
	// ErrInvalidLatitude is returned when a latitude is outside [-90, 90].
	ErrInvalidLatitude = &ValidationError{Field: "latitude", Message: "must be between -90 and 90"}
	// ErrInvalidLongitude is returned when a longitude is outside [-180, 180].
	ErrInvalidLongitude = &ValidationError{Field: "longitude", Message: "must be between -180 and 180"}
)
