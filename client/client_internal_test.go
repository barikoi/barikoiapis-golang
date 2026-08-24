package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestDoNilOut covers the out == nil branch of do, unreachable from the
// public API (every method passes a response type).
func TestDoNilOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":200}`))
	}))
	defer srv.Close()

	c, err := NewClient("k", WithAllowInsecure(), WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := c.do(context.Background(), func(ctx context.Context) (*http.Response, error) {
		return http.Get(srv.URL)
	}, nil); err != nil {
		t.Fatalf("do(nil): %v", err)
	}
}
