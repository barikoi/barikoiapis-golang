package barikoi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	barikoi "github.com/barikoi/barikoiapis-golang"
)

func TestNewClientMissingAPIKey(t *testing.T) {
	if _, err := barikoi.NewClient(""); !errors.Is(err, barikoi.ErrMissingAPIKey) {
		t.Fatalf("got %v, want ErrMissingAPIKey", err)
	}
}

func TestNewClientDefaults(t *testing.T) {
	c, err := barikoi.NewClient("k")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if got := c.GetAPIKey(); got != "k" {
		t.Errorf("GetAPIKey() = %q, want %q", got, "k")
	}
	if got := c.GetTimeout(); got != barikoi.DefaultTimeout {
		t.Errorf("GetTimeout() = %v, want %v", got, barikoi.DefaultTimeout)
	}
}

func TestBaseURLTrailingSlashStripped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/prefix/v2/api/search/autocomplete/place"; got != want {
			t.Errorf("path = %q, want %q (trailing slash stripped, path joined)", got, want)
		}
		writeJSON(t, w, http.StatusOK, `{"places": [], "status": 200}`)
	}))
	defer srv.Close()

	c, err := barikoi.NewClient("test-key",
		barikoi.WithBaseURL(srv.URL+"/prefix/"),
		barikoi.WithAllowInsecure())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.Autocomplete(context.Background(), &barikoi.AutocompleteRequest{Q: "x"}); err != nil {
		t.Fatalf("Autocomplete: %v", err)
	}
}

func TestBaseURLValidation(t *testing.T) {
	cases := []struct {
		name    string
		url     string
		insec   bool
		wantErr bool
	}{
		{"https ok", "https://barikoi.example", false, false},
		{"trailing slash stripped", "https://barikoi.example/", false, false},
		{"http rejected", "http://localhost:3000", false, true},
		{"http allowed with option", "http://localhost:3000", true, false},
		{"garbage rejected", "not-a-url", false, true},
		{"unparseable rejected", "://not-a-url", false, true},
		{"ftp rejected", "ftp://example.com", false, true},
	}
	for _, tc := range cases {
		var opts []barikoi.Option
		if tc.insec {
			opts = append(opts, barikoi.WithAllowInsecure())
		}
		_, err := barikoi.NewClient("k", append(opts, barikoi.WithBaseURL(tc.url))...)
		if gotErr := err != nil; gotErr != tc.wantErr {
			t.Errorf("%s: got err %v, wantErr %v", tc.name, err, tc.wantErr)
		}
	}
}

func TestCustomHTTPClientUsed(t *testing.T) {
	c := offlineClient(t, http.StatusOK, `{"places": [], "status": 200}`)
	if _, err := c.Autocomplete(context.Background(), &barikoi.AutocompleteRequest{Q: "x"}); err != nil {
		t.Fatalf("Autocomplete: %v", err)
	}
}

func TestSetAPIKeyRotation(t *testing.T) {
	var gotKey string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotKey = queryParam(r, "api_key")
		writeJSON(t, w, http.StatusOK, `{"places": [], "status": 200}`)
	})
	c.SetAPIKey("rotated-key")
	if _, err := c.Autocomplete(context.Background(), &barikoi.AutocompleteRequest{Q: "x"}); err != nil {
		t.Fatalf("Autocomplete: %v", err)
	}
	if gotKey != "rotated-key" {
		t.Errorf("api_key = %q, want %q", gotKey, "rotated-key")
	}
}

func TestTimeoutManagement(t *testing.T) {
	c, err := barikoi.NewClient("k")
	if err != nil {
		t.Fatal(err)
	}
	c.SetTimeout(10 * time.Second)
	if got := c.GetTimeout(); got != 10*time.Second {
		t.Errorf("GetTimeout() = %v, want 10s", got)
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
		_, err := c.Autocomplete(context.Background(), &barikoi.AutocompleteRequest{Q: "x"})
		if err == nil {
			t.Fatalf("status %d: want error, got nil", tc.status)
		}
		var be *barikoi.BarikoiError
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

	_, err := c.Autocomplete(ctx, &barikoi.AutocompleteRequest{Q: "x"})
	if err == nil {
		t.Fatal("want error, got nil")
	}
	var te *barikoi.TimeoutError
	if !errors.As(err, &te) {
		t.Fatalf("got %T (%v), want *TimeoutError", err, err)
	}
	if te.Message == "" {
		t.Error("TimeoutError.Message is empty")
	}
}

func TestClientLevelTimeoutError(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		writeJSON(t, w, http.StatusOK, `{"places": [], "status": 200}`)
	})
	c.SetTimeout(50 * time.Millisecond)

	_, err := c.Autocomplete(context.Background(), &barikoi.AutocompleteRequest{Q: "x"})
	var te *barikoi.TimeoutError
	if !errors.As(err, &te) {
		t.Fatalf("got %T (%v), want *TimeoutError", err, err)
	}
}

func TestAPIKeySentAsQueryParameterOnPost(t *testing.T) {
	var gotKey string
	var gotBody map[string]any
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotKey = queryParam(r, "api_key")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		writeJSON(t, w, http.StatusOK, `{"paths": [], "info": {"took": 1}}`)
	})
	_, err := c.CalculateRoute(context.Background(), &barikoi.CalculateRouteRequest{
		Start:       barikoi.Coordinate{Latitude: 23.79, Longitude: 90.36},
		Destination: barikoi.Coordinate{Latitude: 23.78, Longitude: 90.37},
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
