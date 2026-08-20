package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestRouteOverviewSuccess(t *testing.T) {
	const respBody = `{
		"code": "OK",
		"routes": [{
			"geometry": "bqCmhpfPdMaDqEqOuEqO_mFcBueAjZonBlGisJs_auhgManAeJaaBzGsZbQvjsAwjwAqt@pBiBtjAvZzhAga@fAmXbQoi@mmAvALyCfE",
			"legs": [{"steps": [], "distance": 31580.5, "duration": 5011.5, "summary": "s", "weight": 5011.5}],
			"distance": 31580.5,
			"duration": 5011.5,
			"weight_name": "routability",
			"weight": 5011.5
		}],
		"waypoints": [{
			"hint": "1M4MgNbODIAAAAAAUQAAAAAAAABWAgAAAAAAAFZN10EAAAAA6SFHQwAAAABRAAAAAAAAAFYCAADlAAAA6tliBeZObQG10mIF1E9tAQAATwsayG8z",
			"distance": 189.684241,
			"name": null,
			"location": [90.364394, 23.940838]
		}]
	}`
	const coords = "90.362548828125,23.94107556246209;90.31585693359375,24.134221690669204"
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/v2/api/route/"+coords; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		q := r.URL.Query()
		if got := q.Get("geometries"); got != "polyline" {
			t.Errorf("geometries = %q, want %q", got, "polyline")
		}
		if got := q.Get("profile"); got != "foot" {
			t.Errorf("profile = %q, want %q", got, "foot")
		}
		if got := q.Get("api_key"); got != "test-key" {
			t.Errorf("api_key = %q, want %q", got, "test-key")
		}
		writeJSON(t, w, http.StatusOK, respBody)
	})

	resp, err := c.RouteOverview(context.Background(), &RouteOverviewRequest{
		Coordinates: coords,
		Geometries:  "polyline",
		Profile:     "foot",
	})
	if err != nil {
		t.Fatalf("RouteOverview: %v", err)
	}
	if resp.Code != "OK" {
		t.Errorf("Code = %q, want \"OK\"", resp.Code)
	}
	if len(resp.Routes) != 1 || resp.Routes[0].Distance != 31580.5 {
		t.Errorf("routes = %+v, want one route with distance 31580.5", resp.Routes)
	}
	if len(resp.Waypoints) != 1 || len(resp.Waypoints[0].Location) != 2 {
		t.Fatalf("waypoints = %+v, want one with [lon, lat]", resp.Waypoints)
	}
	if resp.Waypoints[0].Location[0] != 90.364394 {
		t.Errorf("waypoint lon = %v, want 90.364394", resp.Waypoints[0].Location[0])
	}
}

func TestRouteOverviewGeojsonGeometry(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{"code":"Ok","routes":[{"geometry":{"type":"LineString","coordinates":[[90.3,23.8]]},"distance":1,"duration":1}],"waypoints":[]}`)
	})
	resp, err := c.RouteOverview(context.Background(), &RouteOverviewRequest{
		Coordinates: "90.3,23.8;90.4,23.9",
		Geometries:  "geojson",
	})
	if err != nil {
		t.Fatalf("RouteOverview: %v", err)
	}
	if want := `{"type":"LineString","coordinates":[[90.3,23.8]]}`; string(resp.Routes[0].Geometry) != want {
		t.Errorf("Geometry = %q, want %q", resp.Routes[0].Geometry, want)
	}
}

func TestRouteOverviewValidation(t *testing.T) {
	c, err := NewClient("k")
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.RouteOverview(context.Background(), &RouteOverviewRequest{})
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Field != "coordinates" {
		t.Fatalf("got %v, want *ValidationError{Field: coordinates}", err)
	}
}

func TestCalculateRouteSuccess(t *testing.T) {
	const respBody = `{
		"trip": {
			"locations": [{"type": "break", "lat": 23.791645, "lon": 90.365587, "original_index": 0}],
			"legs": [{
				"maneuvers": [{
					"type": 1,
					"instruction": "Drive north.",
					"verbal_pre_transition_instruction": "Drive north. Then Bear left.",
					"time": 1.011,
					"length": 0.006,
					"street_names": ["60 Feet Kamal Soroni Road"]
				}],
				"summary": {"has_toll": false, "time": 97.969, "length": 0.58, "cost": 257.358},
				"shape": "wcklA_hnjkD"
			}],
			"summary": {"has_toll": false, "time": 97.969, "length": 0.58, "cost": 257.358},
			"status_message": "Found route between points",
			"status": 0,
			"units": "miles",
			"language": "en-US"
		},
		"id": "test_route"
	}`
	var gotBody struct {
		Data struct {
			Start       Coordinate `json:"start"`
			Destination Coordinate `json:"destination"`
		} `json:"data"`
	}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got, want := r.URL.Path, "/v2/api/routing"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		q := r.URL.Query()
		if got := q.Get("type"); got != "vh" {
			t.Errorf("type = %q, want %q", got, "vh")
		}
		if got := q.Get("profile"); got != "car" {
			t.Errorf("profile = %q, want %q", got, "car")
		}
		if got := q.Get("api_key"); got != "test-key" {
			t.Errorf("api_key = %q, want %q", got, "test-key")
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		writeJSON(t, w, http.StatusOK, respBody)
	})

	resp, err := c.CalculateRoute(context.Background(), &CalculateRouteRequest{
		Start:       Coordinate{Latitude: 23.791645065364126, Longitude: 90.36558776260725},
		Destination: Coordinate{Latitude: 23.784715477921843, Longitude: 90.3676300089066},
		Type:        "vh",
		Profile:     "car",
	})
	if err != nil {
		t.Fatalf("CalculateRoute: %v", err)
	}
	if gotBody.Data.Start.Latitude != 23.791645065364126 || gotBody.Data.Start.Longitude != 90.36558776260725 {
		t.Errorf("body start = %+v", gotBody.Data.Start)
	}
	if gotBody.Data.Destination.Latitude != 23.784715477921843 {
		t.Errorf("body destination = %+v", gotBody.Data.Destination)
	}
	if resp.ID != "test_route" {
		t.Errorf("ID = %q, want %q", resp.ID, "test_route")
	}
	if len(resp.Trip.Legs) != 1 || resp.Trip.Legs[0].Maneuvers[0].Instruction != "Drive north." {
		t.Errorf("trip legs = %+v", resp.Trip.Legs)
	}
	if resp.Trip.Summary.Length != 0.58 {
		t.Errorf("summary length = %v, want 0.58", resp.Trip.Summary.Length)
	}
}

func TestCalculateRouteValidation(t *testing.T) {
	c, err := NewClient("k")
	if err != nil {
		t.Fatal(err)
	}
	req := &CalculateRouteRequest{
		Start:       Coordinate{Latitude: 100, Longitude: 90},
		Destination: Coordinate{Latitude: 23, Longitude: 90},
	}
	if _, err := c.CalculateRoute(context.Background(), req); !errors.Is(err, ErrInvalidLatitude) {
		t.Errorf("bad start lat: got %v, want ErrInvalidLatitude", err)
	}
	req.Start = Coordinate{Latitude: 23, Longitude: 200}
	if _, err := c.CalculateRoute(context.Background(), req); !errors.Is(err, ErrInvalidLongitude) {
		t.Errorf("bad start lon: got %v, want ErrInvalidLongitude", err)
	}
	req.Start = Coordinate{Latitude: 23, Longitude: 90}
	req.Destination = Coordinate{Latitude: -91, Longitude: 90}
	if _, err := c.CalculateRoute(context.Background(), req); !errors.Is(err, ErrInvalidLatitude) {
		t.Errorf("bad destination lat: got %v, want ErrInvalidLatitude", err)
	}
}

func TestOptimizeRouteSuccess(t *testing.T) {
	const respBody = `{
		"hints": {"visited_nodes.sum": 50, "visited_nodes.average": 10},
		"info": {"copyrights": ["GraphHopper"], "took": 0},
		"paths": [{
			"distance": 1846.266,
			"weight": 147.083567,
			"time": 147071,
			"transfers": 0,
			"points_encoded": true,
			"bbox": [0.1],
			"points": "c{|oCacrfPh@hA~B~ErCcBrGwDBw@rOkGDNyKtEiChAmIxE~@fB_AgB_L~GK[d@Y",
			"instructions": [{"distance": 0, "sign": 5, "interval": [0], "text": "Waypoint 1", "time": 0, "street_name": null}],
			"legs": [],
			"details": {},
			"ascend": 0,
			"descend": 0,
			"snapped_waypoints": "c{|oCacrfP??h@hArGzB|AtBeMC"
		}]
	}`
	var gotBody struct {
		APIKey      string               `json:"api_key"`
		Source      string               `json:"source"`
		Destination string               `json:"destination"`
		Profile     string               `json:"profile"`
		GeoPoints   []OptimizeRoutePoint `json:"geo_points"`
	}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got, want := r.URL.Path, "/v2/api/route/optimized"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		if got := r.URL.Query().Get("api_key"); got != "test-key" {
			t.Errorf("api_key query parameter = %q, want %q", got, "test-key")
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		writeJSON(t, w, http.StatusOK, respBody)
	})

	resp, err := c.OptimizeRoute(context.Background(), &OptimizeRouteRequest{
		Source:      "23.746086,90.37368",
		Destination: "23.746214,90.371654",
		Profile:     "car",
		GeoPoints: []OptimizeRoutePoint{
			{ID: 1, Point: "23.746086,90.37368"},
			{ID: 2, Point: "23.746214,90.371654"},
		},
	})
	if err != nil {
		t.Fatalf("OptimizeRoute: %v", err)
	}
	if gotBody.APIKey != "test-key" {
		t.Errorf("body api_key = %q, want %q", gotBody.APIKey, "test-key")
	}
	if gotBody.Source != "23.746086,90.37368" || gotBody.Destination != "23.746214,90.371654" {
		t.Errorf("body source/destination = %q/%q", gotBody.Source, gotBody.Destination)
	}
	if len(gotBody.GeoPoints) != 2 || gotBody.GeoPoints[1].ID != 2 {
		t.Errorf("body geo_points = %+v", gotBody.GeoPoints)
	}
	if len(resp.Paths) != 1 || resp.Paths[0].Distance != 1846.266 {
		t.Errorf("paths = %+v", resp.Paths)
	}
	if resp.Hints.VisitedNodesSum != 50 {
		t.Errorf("hints.visited_nodes.sum = %d, want 50", resp.Hints.VisitedNodesSum)
	}
}

func TestOptimizeRouteValidation(t *testing.T) {
	c, err := NewClient("k")
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.OptimizeRoute(context.Background(), &OptimizeRouteRequest{Destination: "23.7,90.3"})
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Field != "source" {
		t.Fatalf("got %v, want *ValidationError{Field: source}", err)
	}
	_, err = c.OptimizeRoute(context.Background(), &OptimizeRouteRequest{Source: "23.7,90.3"})
	if !errors.As(err, &ve) || ve.Field != "destination" {
		t.Fatalf("got %v, want *ValidationError{Field: destination}", err)
	}
}

func TestSnapToRoadSuccess(t *testing.T) {
	const respBody = `{
		"geometry": {
			"coordinates": [[90.384425, 23.726761], [90.384427, 23.726622]],
			"type": "LineString"
		},
		"distance": 15.5,
		"status": 200
	}`
	const points = "90.38436119310136,23.7267599142696;90.38438265469962,23.726622279057658"
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/v2/api/routing/matching"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		if got := r.URL.Query().Get("points"); got != points {
			t.Errorf("points = %q, want %q", got, points)
		}
		if got := r.URL.Query().Get("api_key"); got != "test-key" {
			t.Errorf("api_key = %q, want %q", got, "test-key")
		}
		writeJSON(t, w, http.StatusOK, respBody)
	})

	resp, err := c.SnapToRoad(context.Background(), &SnapToRoadRequest{Points: points})
	if err != nil {
		t.Fatalf("SnapToRoad: %v", err)
	}
	if resp.Distance != 15.5 {
		t.Errorf("Distance = %v, want 15.5", resp.Distance)
	}
	if len(resp.Geometry.Coordinates) != 2 || resp.Geometry.Coordinates[0][0] != 90.384425 {
		t.Errorf("coordinates = %+v", resp.Geometry.Coordinates)
	}
	if resp.Geometry.Type != "LineString" {
		t.Errorf("type = %q, want LineString", resp.Geometry.Type)
	}
}

func TestSnapToRoadValidation(t *testing.T) {
	c, err := NewClient("k")
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.SnapToRoad(context.Background(), &SnapToRoadRequest{})
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Field != "points" {
		t.Fatalf("got %v, want *ValidationError{Field: points}", err)
	}
}
