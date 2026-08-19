// Command basic demonstrates the Barikoi Go SDK: a reverse geocode, an
// autocomplete, and a route overview, with structured error handling.
//
// Set BARIKOI_API_KEY before running:
//
//	export BARIKOI_API_KEY=your-key
//	go run ./examples/basic
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	bk "github.com/barikoi/barikoiapis/go/client"
)

func main() {
	apiKey := os.Getenv("BARIKOI_API_KEY")
	if apiKey == "" {
		log.Fatal("BARIKOI_API_KEY environment variable is not set")
	}

	c, err := bk.NewClient(apiKey)
	if err != nil {
		log.Fatalf("creating client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Coordinates -> address.
	place, err := c.ReverseGeocode(ctx, &bk.ReverseGeocodeRequest{
		Latitude:  23.806703092211507,
		Longitude: 90.35722628659195,
		Area:      true,
	})
	if err != nil {
		log.Fatalf("ReverseGeocode: %v", err)
	}
	fmt.Printf("Address: %s\n", place.Place.Address)

	// Partial query -> suggestions.
	suggestions, err := c.Autocomplete(ctx, &bk.AutocompleteRequest{Q: "barikoi"})
	if err != nil {
		handleErr("Autocomplete", err)
	} else {
		fmt.Printf("Autocomplete returned %d suggestion(s)\n", len(suggestions.Places))
	}

	// Route between two points.
	route, err := c.RouteOverview(ctx, &bk.RouteOverviewRequest{
		Coordinates: "90.362548828125,23.94107556246209;90.31585693359375,24.134221690669204",
	})
	if err != nil {
		handleErr("RouteOverview", err)
	} else if len(route.Routes) > 0 {
		r := route.Routes[0]
		fmt.Printf("Route: %.1f km in about %.0f minutes\n", r.Distance/1000, r.Duration/60)
	}

	// A client-side validation failure, caught before any HTTP call.
	if _, err := c.ReverseGeocode(ctx, &bk.ReverseGeocodeRequest{Latitude: 120, Longitude: 90}); err != nil {
		handleErr("ReverseGeocode", err)
	}
}

// handleErr prints the error using the errors.As pattern the SDK is built
// around: BarikoiError for API failures, ValidationError for bad input,
// TimeoutError for cancelled or timed-out requests.
func handleErr(op string, err error) {
	var apiErr *bk.BarikoiError
	if errors.As(err, &apiErr) {
		switch {
		case apiErr.IsAuthError():
			log.Printf("%s: invalid API key (status %d)", op, apiErr.StatusCode)
		case apiErr.IsRateLimitError():
			log.Printf("%s: rate limit exceeded, back off and retry", op)
		case apiErr.IsServerError():
			log.Printf("%s: Barikoi server error (status %d)", op, apiErr.StatusCode)
		default:
			log.Printf("%s: API error: %v", op, apiErr)
		}
		return
	}
	var valErr *bk.ValidationError
	if errors.As(err, &valErr) {
		log.Printf("%s: invalid input: field %q %s", op, valErr.Field, valErr.Message)
		return
	}
	var timeoutErr *bk.TimeoutError
	if errors.As(err, &timeoutErr) {
		log.Printf("%s: request timed out: %v", op, timeoutErr)
		return
	}
	log.Printf("%s: %v", op, err)
}
