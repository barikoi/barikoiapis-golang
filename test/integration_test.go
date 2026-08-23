package barikoi_test

import (
	"bufio"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	barikoi "github.com/barikoi/barikoiapis-golang"
)

// Integration tests against the live Barikoi API. They are skipped unless
// BARIKOI_API_KEY is set (environment or .env file), mirroring the
// integrationSuite helper of the TypeScript SDK:
//
//	make test-integration
//
// Unit tests elsewhere in this directory run against httptest stubs and
// need neither a key nor network access.

// loadDotEnv sets env vars from a .env file in the repo root without
// overriding variables already present in the environment.
func loadDotEnv(t *testing.T) {
	t.Helper()
	f, err := os.Open("../.env")
	if err != nil {
		return // no .env, fine
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		v = strings.Trim(v, `"'`)
		if os.Getenv(k) == "" {
			_ = os.Setenv(k, v)
		}
	}
}

var integrationWarned = false

func integrationSuite(t *testing.T) {
	loadDotEnv(t)
	if os.Getenv("BARIKOI_API_KEY") == "" {
		if !integrationWarned {
			integrationWarned = true
			t.Log("\n[integration] BARIKOI_API_KEY not set — skipping live API tests.\n" +
				"Set it in your env or .env file (see env.example) to run the integration suite.")
		}
		t.Skip("BARIKOI_API_KEY not set")
	}
}

func integrationClient(t *testing.T) *barikoi.Client {
	t.Helper()
	integrationSuite(t)
	c, err := barikoi.NewClient(os.Getenv("BARIKOI_API_KEY"), barikoi.WithTimeout(60*time.Second))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

// Common coordinates: Barikoi HQ area, Dhaka.
const (
	itLat = 23.806703092211507
	itLon = 90.35722628659195
)

func TestIntegrationReverseGeocode(t *testing.T) {
	c := integrationClient(t)
	resp, err := c.ReverseGeocode(context.Background(), &barikoi.ReverseGeocodeRequest{Latitude: itLat, Longitude: itLon, Area: true})
	if err != nil {
		t.Fatalf("ReverseGeocode: %v", err)
	}
	if resp.Place.Address == "" {
		t.Error("place.address is empty")
	}
}

func TestIntegrationAutocomplete(t *testing.T) {
	c := integrationClient(t)
	resp, err := c.Autocomplete(context.Background(), &barikoi.AutocompleteRequest{Q: "barikoi"})
	if err != nil {
		t.Fatalf("Autocomplete: %v", err)
	}
	if len(resp.Places) == 0 {
		t.Error("no places returned")
	}
}

func TestIntegrationGeocode(t *testing.T) {
	c := integrationClient(t)
	resp, err := c.Geocode(context.Background(), &barikoi.GeocodeRequest{Q: "shewrapara mirpur dhaka", Thana: true, District: true})
	if err != nil {
		t.Fatalf("Geocode: %v", err)
	}
	if resp.GeocodedAddress.Address == "" {
		t.Error("geocoded_address.address is empty")
	}
}

func TestIntegrationSearchPlaceAndDetails(t *testing.T) {
	c := integrationClient(t)
	search, err := c.SearchPlace(context.Background(), &barikoi.SearchPlaceRequest{Q: "barikoi"})
	if err != nil {
		t.Fatalf("SearchPlace: %v", err)
	}
	if len(search.Places) == 0 || search.SessionID == "" {
		t.Fatalf("search returned %d places, session %q", len(search.Places), search.SessionID)
	}
	details, err := c.PlaceDetails(context.Background(), &barikoi.PlaceDetailsRequest{
		PlaceCode: search.Places[0].PlaceCode,
		SessionID: search.SessionID,
	})
	if err != nil {
		t.Fatalf("PlaceDetails: %v", err)
	}
	if details.Place.PlaceCode == "" || details.Place.Latitude == 0 {
		t.Errorf("place = %+v, want code and coordinates", details.Place)
	}
}

func TestIntegrationNearby(t *testing.T) {
	c := integrationClient(t)
	resp, err := c.Nearby(context.Background(), &barikoi.NearbyRequest{Latitude: 23.87188719, Longitude: 90.38305163})
	if err != nil {
		t.Fatalf("Nearby: %v", err)
	}
	if len(resp.Places) == 0 {
		t.Error("no nearby places returned")
	}
}

func TestIntegrationCheckNearby(t *testing.T) {
	c := integrationClient(t)
	resp, err := c.CheckNearby(context.Background(), &barikoi.CheckNearbyRequest{
		CurrentLatitude:      23.762412943322726,
		CurrentLongitude:     90.37864864706823,
		DestinationLatitude:  23.7624553867393,
		DestinationLongitude: 90.37852866512583,
		Radius:               50,
	})
	if err != nil {
		t.Fatalf("CheckNearby: %v", err)
	}
	if resp.Status != 200 {
		t.Errorf("status = %d, want 200", resp.Status)
	}
}

func TestIntegrationRouteOverview(t *testing.T) {
	c := integrationClient(t)
	resp, err := c.RouteOverview(context.Background(), &barikoi.RouteOverviewRequest{
		Coordinates: "90.4125,23.8103;90.3742,23.7461",
	})
	if err != nil {
		t.Fatalf("RouteOverview: %v", err)
	}
	if len(resp.Routes) == 0 || resp.Routes[0].Distance <= 0 {
		t.Errorf("routes = %+v, want one with distance", resp.Routes)
	}
}

func TestIntegrationCalculateRoute(t *testing.T) {
	c := integrationClient(t)
	resp, err := c.CalculateRoute(context.Background(), &barikoi.CalculateRouteRequest{
		Start:       barikoi.Coordinate{Latitude: 23.8103, Longitude: 90.4125},
		Destination: barikoi.Coordinate{Latitude: 23.7461, Longitude: 90.3742},
	})
	if err != nil {
		t.Fatalf("CalculateRoute: %v", err)
	}
	if len(resp.Paths) == 0 || len(resp.Paths[0].Instructions) == 0 {
		t.Errorf("paths = %+v, want instructions", resp.Paths)
	}
}

func TestIntegrationOptimizeRoute(t *testing.T) {
	c := integrationClient(t)
	resp, err := c.OptimizeRoute(context.Background(), &barikoi.OptimizeRouteRequest{
		Source:      "23.8103,90.4125",
		Destination: "23.7461,90.3742",
		GeoPoints: []barikoi.OptimizeRoutePoint{
			{ID: 1, Point: "23.7925,90.4078"},
			{ID: 2, Point: "23.7609,90.3805"},
		},
	})
	if err != nil {
		t.Fatalf("OptimizeRoute: %v", err)
	}
	if len(resp.Paths) == 0 || resp.Paths[0].Distance <= 0 {
		t.Errorf("paths = %+v, want distance", resp.Paths)
	}
}

func TestIntegrationSnapToRoad(t *testing.T) {
	c := integrationClient(t)
	resp, err := c.SnapToRoad(context.Background(), &barikoi.SnapToRoadRequest{Point: "23.8065,90.3613"})
	if err != nil {
		t.Fatalf("SnapToRoad: %v", err)
	}
	if len(resp.Coordinates) != 2 {
		t.Errorf("coordinates = %v, want [lon, lat]", resp.Coordinates)
	}
}
