package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// testClient returns a Client pointed at a throwaway test server.
func testClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c, err := NewClient("test-key", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, body string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write([]byte(body)); err != nil {
		t.Errorf("writing response: %v", err)
	}
}

func TestNewClientMissingAPIKey(t *testing.T) {
	if _, err := NewClient(""); !errors.Is(err, ErrMissingAPIKey) {
		t.Fatalf("got %v, want ErrMissingAPIKey", err)
	}
}

func TestNewClientDefaults(t *testing.T) {
	c, err := NewClient("k")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if got := c.GetAPIKey(); got != "k" {
		t.Errorf("GetAPIKey() = %q, want %q", got, "k")
	}
	if got := c.baseURL.String(); got != DefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", got, DefaultBaseURL)
	}
	if got := c.httpClient.Timeout; got != DefaultTimeout {
		t.Errorf("timeout = %v, want %v", got, DefaultTimeout)
	}
}

func TestClientOptions(t *testing.T) {
	hc := &http.Client{}
	c, err := NewClient("k",
		WithBaseURL("http://example.com/prefix/"),
		WithHTTPClient(hc),
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if got := c.baseURL.String(); got != "http://example.com/prefix" {
		t.Errorf("baseURL = %q, want %q", got, "http://example.com/prefix")
	}
	if c.httpClient != hc {
		t.Error("httpClient was not the custom client")
	}
	if got := c.httpClient.Timeout; got != 5*time.Second {
		t.Errorf("timeout = %v, want 5s", got)
	}

	if _, err := NewClient("k", WithBaseURL("://not-a-url")); err == nil {
		t.Error("invalid base URL accepted, want error")
	}
}

func TestSetAPIKeyRotation(t *testing.T) {
	var gotKey string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.URL.Query().Get("api_key")
		writeJSON(t, w, http.StatusOK, `{"places": [], "status": 200}`)
	})
	c.SetAPIKey("rotated-key")
	if _, err := c.Autocomplete(context.Background(), &AutocompleteRequest{Q: "x"}); err != nil {
		t.Fatalf("Autocomplete: %v", err)
	}
	if gotKey != "rotated-key" {
		t.Errorf("api_key = %q, want %q", gotKey, "rotated-key")
	}
}

func TestBarikoiErrorMapping(t *testing.T) {
	cases := []struct {
		status      int
		body        string
		wantAuth    bool
		wantRate    bool
		wantServer  bool
		wantMessage string
	}{
		{http.StatusUnauthorized, `{"message": "Invalid or No Registered Key", "status": 401}`, true, false, false, "Invalid or No Registered Key"},
		{http.StatusForbidden, `{"message": "forbidden", "status": 403}`, true, false, false, "forbidden"},
		{http.StatusTooManyRequests, `{"message": "API limit exceeded", "status": 429}`, false, true, false, "API limit exceeded"},
		{http.StatusInternalServerError, `{"message": "internal error", "status": 500}`, false, false, true, "internal error"},
		{http.StatusBadGateway, `not json`, false, false, true, "request failed with status 502"},
	}
	for _, tc := range cases {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(t, w, tc.status, tc.body)
		})
		_, err := c.Autocomplete(context.Background(), &AutocompleteRequest{Q: "x"})
		if err == nil {
			t.Fatalf("status %d: want error, got nil", tc.status)
		}
		var be *BarikoiError
		if !errors.As(err, &be) {
			t.Fatalf("status %d: got %T, want *BarikoiError", tc.status, err)
		}
		if be.StatusCode != tc.status {
			t.Errorf("status %d: StatusCode = %d", tc.status, be.StatusCode)
		}
		if be.Message != tc.wantMessage {
			t.Errorf("status %d: Message = %q, want %q", tc.status, be.Message, tc.wantMessage)
		}
		if be.IsAuthError() != tc.wantAuth {
			t.Errorf("status %d: IsAuthError() = %v, want %v", tc.status, be.IsAuthError(), tc.wantAuth)
		}
		if be.IsRateLimitError() != tc.wantRate {
			t.Errorf("status %d: IsRateLimitError() = %v, want %v", tc.status, be.IsRateLimitError(), tc.wantRate)
		}
		if be.IsServerError() != tc.wantServer {
			t.Errorf("status %d: IsServerError() = %v, want %v", tc.status, be.IsServerError(), tc.wantServer)
		}
		if _, ok := be.Details.(string); !ok {
			t.Errorf("status %d: Details = %T, want string with raw body", tc.status, be.Details)
		}
	}
}

func TestTimeoutError(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		writeJSON(t, w, http.StatusOK, `{"places": [], "status": 200}`)
	})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.Autocomplete(ctx, &AutocompleteRequest{Q: "x"})
	if err == nil {
		t.Fatal("want error, got nil")
	}
	var te *TimeoutError
	if !errors.As(err, &te) {
		t.Fatalf("got %T (%v), want *TimeoutError", err, err)
	}
	if te.Message == "" {
		t.Error("TimeoutError.Message is empty")
	}
}

func TestAPIKeySentAsQueryParameterOnPost(t *testing.T) {
	var gotKey string
	var gotBody map[string]any
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.URL.Query().Get("api_key")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		writeJSON(t, w, http.StatusOK, `{"trip": {}, "id": "t"}`)
	})
	_, err := c.CalculateRoute(context.Background(), &CalculateRouteRequest{
		Start:       Coordinate{Latitude: 23.79, Longitude: 90.36},
		Destination: Coordinate{Latitude: 23.78, Longitude: 90.37},
		Type:        "vh",
	})
	if err != nil {
		t.Fatalf("CalculateRoute: %v", err)
	}
	if gotKey != "test-key" {
		t.Errorf("api_key query parameter = %q, want %q", gotKey, "test-key")
	}
	if gotBody["api_key"] != nil {
		t.Errorf("JSON body unexpectedly contains api_key: %v", gotBody)
	}
}
