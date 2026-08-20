package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Defaults used by NewClient.
const (
	DefaultBaseURL = "https://barikoi.xyz"
	DefaultTimeout = 30 * time.Second
)

// Client talks to the Barikoi Location APIs. Create one with NewClient; it is
// safe for concurrent use. All API methods hang off *Client.
type Client struct {
	mu         sync.RWMutex
	apiKey     string
	baseURL    *url.URL
	httpClient *http.Client
}

type config struct {
	baseURL    string
	timeout    time.Duration
	httpClient *http.Client
}

// Option configures a Client in NewClient.
type Option func(*config)

// WithBaseURL sets the API base URL (default "https://barikoi.xyz").
func WithBaseURL(rawURL string) Option {
	return func(c *config) { c.baseURL = rawURL }
}

// WithTimeout sets the per-request timeout (default 30s). It overrides the
// timeout of an http.Client supplied via WithHTTPClient.
func WithTimeout(d time.Duration) Option {
	return func(c *config) { c.timeout = d }
}

// WithHTTPClient sets a custom *http.Client for outgoing requests. A custom
// client keeps its own Timeout unless WithTimeout is also given.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *config) { c.httpClient = hc }
}

// NewClient returns a Client that authenticates with the given API key,
// obtained from https://developer.barikoi.com. The key is sent as the
// api_key query parameter on every request, GET and POST alike.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}
	cfg := config{baseURL: DefaultBaseURL}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.httpClient == nil {
		cfg.httpClient = &http.Client{Timeout: DefaultTimeout}
	}
	if cfg.timeout != 0 {
		cfg.httpClient.Timeout = cfg.timeout
	}
	base, err := url.Parse(cfg.baseURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return nil, fmt.Errorf("barikoi: invalid base URL %q", cfg.baseURL)
	}
	base.Path = strings.TrimSuffix(base.Path, "/")
	return &Client{apiKey: apiKey, baseURL: base, httpClient: cfg.httpClient}, nil
}

// SetAPIKey replaces the API key used for subsequent requests.
func (c *Client) SetAPIKey(apiKey string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.apiKey = apiKey
}

// GetAPIKey returns the current API key.
func (c *Client) GetAPIKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.apiKey
}

// doGet performs a GET request and decodes the JSON response into out.
func (c *Client) doGet(ctx context.Context, path string, query url.Values, out any) error {
	return c.do(ctx, http.MethodGet, path, query, nil, "", out)
}

// doPostJSON sends body as a JSON POST and decodes the response into out.
func (c *Client) doPostJSON(ctx context.Context, path string, query url.Values, body, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("barikoi: encoding request body: %w", err)
	}
	return c.do(ctx, http.MethodPost, path, query, data, "application/json", out)
}

// doPostForm sends form as an application/x-www-form-urlencoded POST and
// decodes the response into out.
func (c *Client) doPostForm(ctx context.Context, path string, query, form url.Values, out any) error {
	return c.do(ctx, http.MethodPost, path, query, []byte(form.Encode()), "application/x-www-form-urlencoded", out)
}

// do performs an HTTP request against the API, adding the api_key query
// parameter, and decodes the JSON response into out.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body []byte, contentType string, out any) error {
	u := *c.baseURL
	u.Path += path
	if query == nil {
		query = url.Values{}
	}
	query.Set("api_key", c.GetAPIKey())
	u.RawQuery = query.Encode()

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return wrapTransportError(ctx, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return wrapTransportError(ctx, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return barikoiError(resp.StatusCode, data)
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("barikoi: decoding response from %s: %w", path, err)
	}
	return nil
}

// wrapTransportError maps cancellation, deadline, and network timeouts to
// *TimeoutError; other transport errors pass through unchanged.
func wrapTransportError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return &TimeoutError{Message: fmt.Sprintf("barikoi: request cancelled or timed out: %v", ctx.Err())}
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return &TimeoutError{Message: fmt.Sprintf("barikoi: request timed out: %v", err)}
	}
	return err
}

// barikoiError builds a *BarikoiError from a non-2xx response, preferring the
// message in the server's JSON body.
func barikoiError(statusCode int, data []byte) error {
	e := &BarikoiError{
		StatusCode: statusCode,
		Code:       codeForStatus(statusCode),
		Details:    string(data),
	}
	var body struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(data, &body) == nil && body.Message != "" {
		e.Message = body.Message
	} else {
		e.Message = fmt.Sprintf("request failed with status %d", statusCode)
	}
	return e
}

func codeForStatus(statusCode int) string {
	switch {
	case statusCode == http.StatusBadRequest:
		return "missing_parameter"
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
		return "no_registered_key"
	case statusCode == http.StatusPaymentRequired:
		return "payment_exception"
	case statusCode == http.StatusTooManyRequests:
		return "api_limit_exceeded"
	case statusCode >= 500:
		return "server_error"
	default:
		return "unknown_error"
	}
}

// validateCoords checks latitude/longitude bounds before any HTTP call.
func validateCoords(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return ErrInvalidLatitude
	}
	if lon < -180 || lon > 180 {
		return ErrInvalidLongitude
	}
	return nil
}

// requireString returns a *ValidationError if value is blank.
func requireString(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{Field: field, Message: "is required"}
	}
	return nil
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// setBoolParam sets name to "true" only when v is true.
func setBoolParam(q url.Values, name string, v bool) {
	if v {
		q.Set(name, "true")
	}
}

// FlexFloat unmarshals a JSON number or a numeric string. Barikoi returns
// coordinates in both forms depending on the endpoint.
type FlexFloat float64

func (f *FlexFloat) UnmarshalJSON(data []byte) error {
	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexFloat(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	n, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return fmt.Errorf("barikoi: %q is not a number: %w", s, err)
	}
	*f = FlexFloat(n)
	return nil
}

// FlexString unmarshals a JSON string or number. Barikoi returns post codes
// in both forms depending on the endpoint.
type FlexString string

func (s *FlexString) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = FlexString(str)
		return nil
	}
	var n float64
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*s = FlexString(strconv.FormatFloat(n, 'f', -1, 64))
	return nil
}

// PolylineOrGeoJSON holds a route geometry: an encoded polyline string, or
// the raw JSON geometry when geometries=geojson was requested.
type PolylineOrGeoJSON string

func (g *PolylineOrGeoJSON) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*g = PolylineOrGeoJSON(s)
		return nil
	}
	*g = PolylineOrGeoJSON(data)
	return nil
}
