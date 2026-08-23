// Package client is the hand-written layer of the Barikoi Go SDK. It wraps
// the generated API client (github.com/barikoi/barikoiapis-golang/gen) with
// simplified request types, client-side validation, API key injection,
// timeouts, and the SDK's error types — the equivalent of the TypeScript
// SDK's src/lib/client.ts.
package client

import (
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

	"github.com/barikoi/barikoiapis-golang/gen"
)

// Defaults used by NewClient.
const (
	DefaultBaseURL = "https://barikoi.xyz"
	DefaultTimeout = 30 * time.Second
)

// Client talks to the Barikoi Location APIs. Create one with NewClient; it is
// safe for concurrent use. All API methods hang off *Client.
type Client struct {
	mu      sync.RWMutex
	apiKey  string
	baseURL *url.URL
	gen     *gen.Client
	// httpClient is retained for test introspection; gen issues requests
	// through it.
	httpClient *http.Client
	timeout    time.Duration
}

type config struct {
	baseURL       string
	timeout       time.Duration
	httpClient    *http.Client
	allowInsecure bool
}

// Option configures a Client in NewClient.
type Option func(*config)

// WithBaseURL sets the API base URL (default "https://barikoi.xyz"). The URL
// must use https; http is refused to keep the API key out of cleartext
// traffic. Use WithAllowInsecure to permit http for local development.
func WithBaseURL(rawURL string) Option {
	return func(c *config) { c.baseURL = rawURL }
}

// WithTimeout sets the per-request timeout (default 30s).
func WithTimeout(d time.Duration) Option {
	return func(c *config) { c.timeout = d }
}

// WithHTTPClient sets a custom *http.Client for outgoing requests. The
// per-request timeout from WithTimeout (or the 30s default) still applies on
// top of the client's own Timeout, whichever is shorter.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *config) { c.httpClient = hc }
}

// WithAllowInsecure permits http:// base URLs, which the client refuses by
// default. Use only for local development and testing.
func WithAllowInsecure() Option {
	return func(c *config) { c.allowInsecure = true }
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
		cfg.httpClient = &http.Client{}
	}
	if cfg.timeout <= 0 {
		cfg.timeout = DefaultTimeout
	}
	base, err := parseBaseURL(cfg.baseURL, cfg.allowInsecure)
	if err != nil {
		return nil, err
	}
	gc, err := gen.NewClient(base.String(), gen.WithHTTPClient(cfg.httpClient))
	if err != nil {
		return nil, fmt.Errorf("barikoi: creating API client: %w", err)
	}
	return &Client{apiKey: apiKey, baseURL: base, gen: gc, httpClient: cfg.httpClient, timeout: cfg.timeout}, nil
}

// parseBaseURL validates the base URL the way the TypeScript SDK does: it
// must be an absolute http(s) URL with no fragment or query, https unless
// insecure is allowed, and it is normalized without a trailing slash.
func parseBaseURL(rawURL string, allowInsecure bool) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" || u.Fragment != "" || u.RawQuery != "" {
		return nil, fmt.Errorf("barikoi: invalid base URL %q", rawURL)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("barikoi: invalid base URL %q: scheme must be http or https", rawURL)
	}
	if u.Scheme == "http" && !allowInsecure {
		return nil, fmt.Errorf("barikoi: base URL must use https:// (got http://); pass WithAllowInsecure for local development")
	}
	u.Path = strings.TrimSuffix(u.Path, "/")
	return u, nil
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

// SetTimeout replaces the per-request timeout for subsequent requests.
func (c *Client) SetTimeout(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timeout = d
}

// GetTimeout returns the current per-request timeout.
func (c *Client) GetTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.timeout
}

// apiKeyParam returns the API key for the request being built.
func (c *Client) apiKeyParam() gen.ApiKey {
	return gen.ApiKey(c.GetAPIKey())
}

// withAPIKeyQuery returns a request editor that adds the api_key query
// parameter, for operations whose spec parameters omit it (only optimizeRoute
// today: it carries the key in its request body instead).
func (c *Client) withAPIKeyQuery() gen.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		q := req.URL.Query()
		q.Set("api_key", c.GetAPIKey())
		req.URL.RawQuery = q.Encode()
		return nil
	}
}

// do runs a generated API call with the per-request timeout applied, then
// decodes the JSON response into out. Non-2xx responses become *BarikoiError,
// cancellations and timeouts become *TimeoutError.
func (c *Client) do(ctx context.Context, call func(ctx context.Context) (*http.Response, error), out any) error {
	if timeout := c.GetTimeout(); timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	resp, err := call(ctx)
	if err != nil {
		return wrapTransportError(ctx, err)
	}
	defer func() { _ = resp.Body.Close() }()

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
		return fmt.Errorf("barikoi: decoding response: %w", err)
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
