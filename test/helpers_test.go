package barikoi_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	barikoi "github.com/barikoi/barikoiapis-golang"
)

// Shared helpers for the test suite. All tests are black-box: they exercise
// the SDK only through the public facade (package barikoi), mirroring the
// TypeScript SDK's test/ directory that imports from the built package.

// testClient returns a Client pointed at a throwaway test server. The server
// speaks plain http, so the client is created with WithAllowInsecure.
func testClient(t *testing.T, h http.HandlerFunc) *barikoi.Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c, err := barikoi.NewClient("test-key", barikoi.WithBaseURL(srv.URL), barikoi.WithAllowInsecure())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

// offlineClient returns a Client whose custom transport responds with the
// given canned body, proving WithHTTPClient wiring without any network.
func offlineClient(t *testing.T, status int, body string) *barikoi.Client {
	t.Helper()
	hc := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Request:    r,
		}, nil
	})}
	c, err := barikoi.NewClient("test-key", barikoi.WithHTTPClient(hc))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func writeJSON(t *testing.T, w http.ResponseWriter, status int, body string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write([]byte(body)); err != nil {
		t.Errorf("writing response: %v", err)
	}
}

// queryParam returns the named query parameter of the request.
func queryParam(r *http.Request, name string) string { return r.URL.Query().Get(name) }
