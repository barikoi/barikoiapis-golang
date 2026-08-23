// Command basic exercises every Barikoi Go SDK API (11 endpoints) against
// the live service, with structured error handling.
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

	"github.com/barikoi/barikoiapis-golang"
)

// Common coordinates for the examples below: Barikoi HQ area, Dhaka.
const (
	lat = 23.806703092211507
	lon = 90.35722628659195
)

func main() {
	apiKey := os.Getenv("BARIKOI_API_KEY")
	if apiKey == "" {
		log.Fatal("BARIKOI_API_KEY environment variable is not set")
	}

	c, err := barikoi.NewClient(apiKey)
	if err != nil {
		log.Fatalf("creating client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// --- Geocoding ---

	// 1. Coordinates -> address.
	if place, err := c.ReverseGeocode(ctx, &barikoi.ReverseGeocodeRequest{
		Latitude:  lat,
		Longitude: lon,
		Area:      true,
	}); err != nil {
		handleErr("ReverseGeocode", err)
	} else {
		fmt.Printf("1. ReverseGeocode: %s\n", place.Place.Address)
	}

	// 2. Partial query -> suggestions.
	if suggestions, err := c.Autocomplete(ctx, &barikoi.AutocompleteRequest{Q: "barikoi"}); err != nil {
		handleErr("Autocomplete", err)
	} else {
		fmt.Printf("2. Autocomplete: %d suggestion(s)\n", len(suggestions.Places))
	}

	// 3. Address string -> formatted address and coordinates.
	if geo, err := c.Geocode(ctx, &barikoi.GeocodeRequest{Q: "barikoi office dhaka"}); err != nil {
		handleErr("Geocode", err)
	} else {
		fmt.Printf("3. Geocode: %s (%.4f%% confidence)\n",
			geo.GeocodedAddress.Address, geo.ConfidenceScorePercentage)
	}

	// --- Routing ---

	// 4. Basic route between two points.
	if route, err := c.RouteOverview(ctx, &barikoi.RouteOverviewRequest{
		Coordinates: "90.362548828125,23.94107556246209;90.31585693359375,24.134221690669204",
	}); err != nil {
		handleErr("RouteOverview", err)
	} else if len(route.Routes) > 0 {
		r := route.Routes[0]
		fmt.Printf("4. RouteOverview: %.1f km in about %.0f minutes\n", r.Distance/1000, r.Duration/60)
	}

	// 5. Detailed route with turn-by-turn instructions.
	if route, err := c.CalculateRoute(ctx, &barikoi.CalculateRouteRequest{
		Start:       barikoi.Coordinate{Latitude: 23.94107556246209, Longitude: 90.362548828125},
		Destination: barikoi.Coordinate{Latitude: 24.134221690669204, Longitude: 90.31585693359375},
	}); err != nil {
		handleErr("CalculateRoute", err)
	} else if len(route.Paths) > 0 && len(route.Paths[0].Instructions) > 0 {
		p := route.Paths[0]
		fmt.Printf("5. CalculateRoute: %.1f km, %d instruction(s), first: %s\n",
			p.Distance/1000, len(p.Instructions), p.Instructions[0].Text)
	}

	// 6. Route optimization through intermediate waypoints.
	if opt, err := c.OptimizeRoute(ctx, &barikoi.OptimizeRouteRequest{
		Source:      "23.94107556246209,90.362548828125",
		Destination: "24.134221690669204,90.31585693359375",
		GeoPoints: []barikoi.OptimizeRoutePoint{
			{ID: 1, Point: "23.95,90.38"},
			{ID: 2, Point: "24.05,90.33"},
		},
	}); err != nil {
		handleErr("OptimizeRoute", err)
	} else if len(opt.Paths) > 0 {
		fmt.Printf("6. OptimizeRoute: %.1f km optimized path\n", opt.Paths[0].Distance/1000)
	}

	// 7. Coordinate -> nearest point on the road network.
	if snap, err := c.SnapToRoad(ctx, &barikoi.SnapToRoadRequest{
		Point: "23.94107556246209,90.362548828125",
	}); err != nil {
		handleErr("SnapToRoad", err)
	} else {
		fmt.Printf("7. SnapToRoad: snapped to [%.6f, %.6f], %.1f m away\n",
			snap.Coordinates[0], snap.Coordinates[1], snap.Distance)
	}

	// --- Search ---

	// 8. Place search; the response feeds PlaceDetails below.
	var search *barikoi.SearchPlaceResponse
	if search, err = c.SearchPlace(ctx, &barikoi.SearchPlaceRequest{Q: "barikoi"}); err != nil {
		handleErr("SearchPlace", err)
	} else {
		fmt.Printf("8. SearchPlace: %d hit(s), session %s\n", len(search.Places), search.SessionID)
	}

	// 9. Details for the first place code from the search above.
	if search != nil && len(search.Places) > 0 {
		if details, err := c.PlaceDetails(ctx, &barikoi.PlaceDetailsRequest{
			PlaceCode: search.Places[0].PlaceCode,
			SessionID: search.SessionID,
		}); err != nil {
			handleErr("PlaceDetails", err)
		} else {
			fmt.Printf("9. PlaceDetails: %s (%v, %v)\n",
				details.Place.Address, details.Place.Latitude, details.Place.Longitude)
		}
	}

	// 10. Places within 2 km of a point.
	if nearby, err := c.Nearby(ctx, &barikoi.NearbyRequest{
		Latitude:  lat,
		Longitude: lon,
		Radius:    2,
		Limit:     5,
	}); err != nil {
		handleErr("Nearby", err)
	} else {
		fmt.Printf("10. Nearby: %d place(s)", len(nearby.Places))
		if len(nearby.Places) > 0 {
			fmt.Printf(", closest: %s", nearby.Places[0].Name)
		}
		fmt.Println()
	}

	// 11. Geofence check: is the destination within 500 m of the current point?
	if check, err := c.CheckNearby(ctx, &barikoi.CheckNearbyRequest{
		CurrentLatitude:      lat,
		CurrentLongitude:     lon,
		DestinationLatitude:  lat + 0.001, // ~111 m north
		DestinationLongitude: lon,
		Radius:               500,
	}); err != nil {
		handleErr("CheckNearby", err)
	} else if check.Data != nil {
		fmt.Printf("11. CheckNearby: within radius, near %s\n", check.Data.Name)
	} else {
		fmt.Printf("11. CheckNearby: %s\n", check.Message)
	}

	// A client-side validation failure, caught before any HTTP call.
	if _, err := c.ReverseGeocode(ctx, &barikoi.ReverseGeocodeRequest{Latitude: 120, Longitude: 90}); err != nil {
		handleErr("ReverseGeocode", err)
	}
}

// handleErr prints the error using the errors.As pattern the SDK is built
// around: BarikoiError for API failures, ValidationError for bad input,
// TimeoutError for cancelled or timed-out requests.
func handleErr(op string, err error) {
	var apiErr *barikoi.BarikoiError
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
	var valErr *barikoi.ValidationError
	if errors.As(err, &valErr) {
		log.Printf("%s: invalid input: field %q %s", op, valErr.Field, valErr.Message)
		return
	}
	var timeoutErr *barikoi.TimeoutError
	if errors.As(err, &timeoutErr) {
		log.Printf("%s: request timed out: %v", op, timeoutErr)
		return
	}
	log.Printf("%s: %v", op, err)
}
