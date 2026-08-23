package barikoi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	barikoi "github.com/barikoi/barikoiapis-golang"
)

// Facade test: exercises the public import path
// (github.com/barikoi/barikoiapis-golang) end to end.
func TestFacadeEndToEnd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("api_key") != "k" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"message": "Invalid API key", "status": 401})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"places": []map[string]any{{"id": 1, "address": "Mirpur, Dhaka", "longitude": 90.36, "latitude": 23.82}},
			"status": 200,
		})
	}))
	defer srv.Close()

	c, err := barikoi.NewClient("k", barikoi.WithBaseURL(srv.URL), barikoi.WithAllowInsecure())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	resp, err := c.Autocomplete(context.Background(), &barikoi.AutocompleteRequest{Q: "mirpur"})
	if err != nil {
		t.Fatalf("Autocomplete: %v", err)
	}
	if len(resp.Places) != 1 || resp.Places[0].Latitude != 23.82 {
		t.Errorf("places = %+v", resp.Places)
	}

	// Error types flow through the facade undamaged.
	_, err = c.Autocomplete(context.Background(), &barikoi.AutocompleteRequest{Q: ""})
	var valErr *barikoi.ValidationError
	if !errors.As(err, &valErr) || valErr.Field != "q" {
		t.Errorf("got %v, want *barikoi.ValidationError{Field: q}", err)
	}

	if _, err := barikoi.NewClient(""); !errors.Is(err, barikoi.ErrMissingAPIKey) {
		t.Errorf("got %v, want ErrMissingAPIKey", err)
	}

	_ = barikoi.BoolPtr(false) // exported helper
}
