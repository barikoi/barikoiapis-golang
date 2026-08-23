package barikoi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	barikoi "github.com/barikoi/barikoiapis-golang"
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
		if got := queryParam(r, "geometries"); got != "polyline" {
			t.Errorf("geometries = %q, want %q", got, "polyline")
		}
		if got := queryParam(r, "profile"); got != "foot" {
			t.Errorf("profile = %q, want %q", got, "foot")
		}
		if got := queryParam(r, "api_key"); got != "test-key" {
			t.Errorf("api_key = %q, want %q", got, "test-key")
		}
		writeJSON(t, w, http.StatusOK, respBody)
	})

	resp, err := c.RouteOverview(context.Background(), &barikoi.RouteOverviewRequest{
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

func TestRouteOverviewDefaults(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := queryParam(r, "geometries"); got != "polyline" {
			t.Errorf("geometries = %q, want %q (default)", got, "polyline")
		}
		if got := queryParam(r, "profile"); got != "car" {
			t.Errorf("profile = %q, want %q (default)", got, "car")
		}
		writeJSON(t, w, http.StatusOK, `{"code":"Ok","routes":[],"waypoints":[]}`)
	})
	if _, err := c.RouteOverview(context.Background(), &barikoi.RouteOverviewRequest{
		Coordinates: "90.4125,23.8103;90.3742,23.7461",
	}); err != nil {
		t.Fatalf("RouteOverview: %v", err)
	}
}

func TestRouteOverviewGeojsonGeometry(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{"code":"Ok","routes":[{"geometry":{"type":"LineString","coordinates":[[90.3,23.8]]},"distance":1,"duration":1}],"waypoints":[]}`)
	})
	resp, err := c.RouteOverview(context.Background(), &barikoi.RouteOverviewRequest{
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
	c, err := barikoi.NewClient("k")
	if err != nil {
		t.Fatal(err)
	}
	var ve *barikoi.ValidationError
	_, err = c.RouteOverview(context.Background(), &barikoi.RouteOverviewRequest{})
	if !errors.As(err, &ve) || ve.Field != "coordinates" {
		t.Fatalf("empty: got %v, want *ValidationError{Field: coordinates}", err)
	}
	_, err = c.RouteOverview(context.Background(), &barikoi.RouteOverviewRequest{Coordinates: "90.3,23.8"})
	if !errors.As(err, &ve) || ve.Field != "coordinates" {
		t.Fatalf("single pair: got %v, want *ValidationError{Field: coordinates}", err)
	}
	_, err = c.RouteOverview(context.Background(), &barikoi.RouteOverviewRequest{
		Coordinates: "90.3,23.8;90.4,23.9",
		Geometries:  "wkt",
	})
	if !errors.As(err, &ve) || ve.Field != "geometries" {
		t.Fatalf("bad geometries: got %v, want *ValidationError{Field: geometries}", err)
	}
	_, err = c.RouteOverview(context.Background(), &barikoi.RouteOverviewRequest{
		Coordinates: "90.3,23.8;90.4,23.9",
		Profile:     "bike",
	})
	if !errors.As(err, &ve) || ve.Field != "profile" {
		t.Fatalf("bad profile: got %v, want *ValidationError{Field: profile}", err)
	}
	_, err = c.RouteOverview(context.Background(), &barikoi.RouteOverviewRequest{
		Coordinates: "90.3,23.8;90.4,95",
	})
	if !errors.Is(err, barikoi.ErrInvalidLatitude) {
		t.Fatalf("out-of-bounds pair: got %v, want ErrInvalidLatitude", err)
	}
}

func TestCalculateRouteSuccess(t *testing.T) {
	// Live GraphHopper-format response: paths with GeoJSON points
	// (points_encoded=false) and instructions.
	const respBody = `{
		"hints": {"visited_nodes.sum": 144, "visited_nodes.average": 144},
		"info": {"copyrights": ["GraphHopper", "OpenStreetMap contributors"], "took": 10, "road_data_timestamp": "1970-01-01T00:00:00Z"},
		"paths": [{
			"distance": 11851.357,
			"weight": 2671.174236,
			"time": 1604549,
			"transfers": 0,
			"points_encoded": false,
			"bbox": [90.374253, 23.744679, 90.413263, 23.814402],
			"points": {"type": "LineString", "coordinates": [[90.412494, 23.810399], [90.413045, 23.810428]]},
			"instructions": [
				{"distance": 0, "heading": 310.05, "sign": 5, "interval": [0, 0], "text": "Continue onto west", "time": 0, "street_name": null}
			],
			"legs": [],
			"details": [],
			"ascend": 0,
			"descend": 0,
			"snapped_waypoints": {"type": "LineString", "coordinates": [[90.412494, 23.810399], [90.374253, 23.746129]]}
		}]
	}`
	var gotBody struct {
		Data struct {
			Start       barikoi.Coordinate `json:"start"`
			Destination barikoi.Coordinate `json:"destination"`
		} `json:"data"`
	}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got, want := r.URL.Path, "/v2/api/routing"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		if got := queryParam(r, "type"); got != "gh" {
			t.Errorf("type = %q, want %q (default)", got, "gh")
		}
		if got := queryParam(r, "profile"); got != "car" {
			t.Errorf("profile = %q, want %q", got, "car")
		}
		if got := queryParam(r, "api_key"); got != "test-key" {
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

	resp, err := c.CalculateRoute(context.Background(), &barikoi.CalculateRouteRequest{
		Start:       barikoi.Coordinate{Latitude: 23.791645065364126, Longitude: 90.36558776260725},
		Destination: barikoi.Coordinate{Latitude: 23.784715477921843, Longitude: 90.3676300089066},
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
	if len(resp.Paths) != 1 {
		t.Fatalf("len(Paths) = %d, want 1", len(resp.Paths))
	}
	p := resp.Paths[0]
	if p.Distance != 11851.357 || p.Time != 1604549 {
		t.Errorf("path = %.3f m, %.0f ms", p.Distance, p.Time)
	}
	if len(p.BBox) != 4 {
		t.Errorf("bbox = %v, want 4 values", p.BBox)
	}
	if p.PointsEncoded {
		t.Error("points_encoded = true, want false")
	}
	if p.Points.GeoJSON == nil || p.Points.GeoJSON.Type != "LineString" ||
		len(p.Points.GeoJSON.Coordinates) != 2 || p.Points.GeoJSON.Coordinates[0][0] != 90.412494 {
		t.Errorf("points = %+v, want GeoJSON LineString", p.Points.GeoJSON)
	}
	if p.Points.Polyline != "" {
		t.Errorf("points.Polyline = %q, want empty for GeoJSON geometry", p.Points.Polyline)
	}
	if len(p.Instructions) != 1 || p.Instructions[0].Text != "Continue onto west" {
		t.Errorf("instructions = %+v", p.Instructions)
	}
	if p.Instructions[0].StreetName != "" {
		t.Errorf("street_name = %q, want empty for JSON null", p.Instructions[0].StreetName)
	}
	if p.SnappedWaypoints.GeoJSON == nil || len(p.SnappedWaypoints.GeoJSON.Coordinates) != 2 {
		t.Errorf("snapped_waypoints = %+v, want GeoJSON LineString", p.SnappedWaypoints.GeoJSON)
	}
	if resp.Hints.VisitedNodesSum != 144 {
		t.Errorf("hints.visited_nodes.sum = %v, want 144", resp.Hints.VisitedNodesSum)
	}
	if len(resp.Info.Copyrights) != 2 {
		t.Errorf("info.copyrights = %v, want 2 entries", resp.Info.Copyrights)
	}
}

func TestCalculateRouteValidation(t *testing.T) {
	c, err := barikoi.NewClient("k")
	if err != nil {
		t.Fatal(err)
	}
	req := &barikoi.CalculateRouteRequest{
		Start:       barikoi.Coordinate{Latitude: 100, Longitude: 90},
		Destination: barikoi.Coordinate{Latitude: 23, Longitude: 90},
	}
	if _, err := c.CalculateRoute(context.Background(), req); !errors.Is(err, barikoi.ErrInvalidLatitude) {
		t.Errorf("bad start lat: got %v, want ErrInvalidLatitude", err)
	}
	req.Start = barikoi.Coordinate{Latitude: 23, Longitude: 200}
	if _, err := c.CalculateRoute(context.Background(), req); !errors.Is(err, barikoi.ErrInvalidLongitude) {
		t.Errorf("bad start lon: got %v, want ErrInvalidLongitude", err)
	}
	req.Start = barikoi.Coordinate{Latitude: 23, Longitude: 90}
	req.Destination = barikoi.Coordinate{Latitude: -91, Longitude: 90}
	if _, err := c.CalculateRoute(context.Background(), req); !errors.Is(err, barikoi.ErrInvalidLatitude) {
		t.Errorf("bad destination lat: got %v, want ErrInvalidLatitude", err)
	}
	req.Destination = barikoi.Coordinate{Latitude: 23, Longitude: 90}
	var ve *barikoi.ValidationError
	if _, err := c.CalculateRoute(context.Background(), &barikoi.CalculateRouteRequest{
		Start:       req.Start,
		Destination: req.Destination,
		Type:        "vh",
	}); !errors.As(err, &ve) || ve.Field != "type" {
		t.Errorf("type vh: got %v, want *ValidationError{Field: type}", err)
	}
	if _, err := c.CalculateRoute(context.Background(), &barikoi.CalculateRouteRequest{
		Start:       req.Start,
		Destination: req.Destination,
		Profile:     "foot",
	}); !errors.As(err, &ve) || ve.Field != "profile" {
		t.Errorf("profile foot: got %v, want *ValidationError{Field: profile}", err)
	}
}

func TestOptimizeRouteSuccess(t *testing.T) {
	// Optimized routes come back with encoded polyline points
	// (points_encoded=true) — the same GraphHopper envelope as
	// CalculateRoute, but with string geometries.
	const respBody = `{
		"hints": {"visited_nodes.sum": 50, "visited_nodes.average": 10},
		"info": {"copyrights": ["GraphHopper"], "took": 0},
		"paths": [{
			"distance": 1846.266,
			"weight": 147.083567,
			"time": 147071,
			"transfers": 0,
			"points_encoded": true,
			"bbox": [0.1, 0.2, 0.3, 0.4],
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
		APIKey      string                       `json:"api_key"`
		Source      string                       `json:"source"`
		Destination string                       `json:"destination"`
		Profile     string                       `json:"profile"`
		GeoPoints   []barikoi.OptimizeRoutePoint `json:"geo_points"`
	}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got, want := r.URL.Path, "/v2/api/route/optimized"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		if got := queryParam(r, "api_key"); got != "test-key" {
			t.Errorf("api_key query parameter = %q, want %q", got, "test-key")
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		writeJSON(t, w, http.StatusOK, respBody)
	})

	resp, err := c.OptimizeRoute(context.Background(), &barikoi.OptimizeRouteRequest{
		Source:      "23.746086,90.37368",
		Destination: "23.746214,90.371654",
		GeoPoints: []barikoi.OptimizeRoutePoint{
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
	if gotBody.Profile != "car" {
		t.Errorf("body profile = %q, want %q (default)", gotBody.Profile, "car")
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
	p := resp.Paths[0]
	if !p.PointsEncoded {
		t.Error("points_encoded = false, want true")
	}
	if p.Points.Polyline == "" || p.Points.GeoJSON != nil {
		t.Errorf("points = %+v, want encoded polyline string", p.Points)
	}
	if p.SnappedWaypoints.Polyline == "" {
		t.Error("snapped_waypoints = want encoded polyline string")
	}
	if p.Details == nil {
		t.Error("details (raw) = nil, want raw JSON (object form)")
	}
	if len(p.Instructions) != 1 || p.Instructions[0].Text != "Waypoint 1" {
		t.Errorf("instructions = %+v", p.Instructions)
	}
	if resp.Hints.VisitedNodesSum != 50 {
		t.Errorf("hints.visited_nodes.sum = %v, want 50", resp.Hints.VisitedNodesSum)
	}
}

func TestOptimizeRouteValidation(t *testing.T) {
	c, err := barikoi.NewClient("k")
	if err != nil {
		t.Fatal(err)
	}
	var ve *barikoi.ValidationError
	_, err = c.OptimizeRoute(context.Background(), &barikoi.OptimizeRouteRequest{Destination: "23.7,90.3"})
	if !errors.As(err, &ve) || ve.Field != "source" {
		t.Fatalf("got %v, want *ValidationError{Field: source}", err)
	}
	_, err = c.OptimizeRoute(context.Background(), &barikoi.OptimizeRouteRequest{Source: "23.7,90.3"})
	if !errors.As(err, &ve) || ve.Field != "destination" {
		t.Fatalf("got %v, want *ValidationError{Field: destination}", err)
	}
	// Source/destination must be "lat,lon" with valid bounds.
	_, err = c.OptimizeRoute(context.Background(), &barikoi.OptimizeRouteRequest{
		Source:      "23.7",
		Destination: "23.7,90.3",
	})
	if !errors.As(err, &ve) || ve.Field != "source" {
		t.Fatalf("malformed source: got %v, want *ValidationError{Field: source}", err)
	}
	_, err = c.OptimizeRoute(context.Background(), &barikoi.OptimizeRouteRequest{
		Source:      "23.7,90.3",
		Destination: "95,90.3",
	})
	if !errors.Is(err, barikoi.ErrInvalidLatitude) {
		t.Fatalf("out-of-bounds destination: got %v, want ErrInvalidLatitude", err)
	}
	// geo_points must contain 1 to 50 entries, each a valid point.
	_, err = c.OptimizeRoute(context.Background(), &barikoi.OptimizeRouteRequest{
		Source:      "23.7,90.3",
		Destination: "23.8,90.4",
	})
	if !errors.As(err, &ve) || ve.Field != "geo_points" {
		t.Fatalf("no geo_points: got %v, want *ValidationError{Field: geo_points}", err)
	}
	tooMany := make([]barikoi.OptimizeRoutePoint, 51)
	for i := range tooMany {
		tooMany[i] = barikoi.OptimizeRoutePoint{ID: i, Point: "23.7,90.3"}
	}
	_, err = c.OptimizeRoute(context.Background(), &barikoi.OptimizeRouteRequest{
		Source:      "23.7,90.3",
		Destination: "23.8,90.4",
		GeoPoints:   tooMany,
	})
	if !errors.As(err, &ve) || ve.Field != "geo_points" {
		t.Fatalf("51 geo_points: got %v, want *ValidationError{Field: geo_points}", err)
	}
	_, err = c.OptimizeRoute(context.Background(), &barikoi.OptimizeRouteRequest{
		Source:      "23.7,90.3",
		Destination: "23.8,90.4",
		GeoPoints:   []barikoi.OptimizeRoutePoint{{ID: 1, Point: "not-a-point"}},
	})
	if !errors.As(err, &ve) || ve.Field != "geo_points" {
		t.Fatalf("bad geo_point: got %v, want *ValidationError{Field: geo_points}", err)
	}
	_, err = c.OptimizeRoute(context.Background(), &barikoi.OptimizeRouteRequest{
		Source:      "23.7,90.3",
		Destination: "23.8,90.4",
		Profile:     "boat",
		GeoPoints:   []barikoi.OptimizeRoutePoint{{ID: 1, Point: "23.7,90.3"}},
	})
	if !errors.As(err, &ve) || ve.Field != "profile" {
		t.Fatalf("bad profile: got %v, want *ValidationError{Field: profile}", err)
	}
}

func TestSnapToRoadSuccess(t *testing.T) {
	const respBody = `{
		"coordinates": [90.384425, 23.726761],
		"distance": 15.5,
		"type": "Point"
	}`
	const point = "23.7267599142696,90.38436119310136"
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/v2/api/routing/nearest"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		if got := queryParam(r, "point"); got != point {
			t.Errorf("point = %q, want %q", got, point)
		}
		if got := queryParam(r, "api_key"); got != "test-key" {
			t.Errorf("api_key = %q, want %q", got, "test-key")
		}
		writeJSON(t, w, http.StatusOK, respBody)
	})

	resp, err := c.SnapToRoad(context.Background(), &barikoi.SnapToRoadRequest{Point: point})
	if err != nil {
		t.Fatalf("SnapToRoad: %v", err)
	}
	if resp.Distance != 15.5 {
		t.Errorf("Distance = %v, want 15.5", resp.Distance)
	}
	if len(resp.Coordinates) != 2 || resp.Coordinates[0] != 90.384425 {
		t.Errorf("coordinates = %+v", resp.Coordinates)
	}
	if resp.Type != "Point" {
		t.Errorf("type = %q, want Point", resp.Type)
	}
}

func TestSnapToRoadValidation(t *testing.T) {
	c, err := barikoi.NewClient("k")
	if err != nil {
		t.Fatal(err)
	}
	var ve *barikoi.ValidationError
	_, err = c.SnapToRoad(context.Background(), &barikoi.SnapToRoadRequest{})
	if !errors.As(err, &ve) || ve.Field != "point" {
		t.Fatalf("empty: got %v, want *ValidationError{Field: point}", err)
	}
	_, err = c.SnapToRoad(context.Background(), &barikoi.SnapToRoadRequest{Point: "90.38,23.72,extra"})
	if !errors.As(err, &ve) || ve.Field != "point" {
		t.Fatalf("malformed: got %v, want *ValidationError{Field: point}", err)
	}
	_, err = c.SnapToRoad(context.Background(), &barikoi.SnapToRoadRequest{Point: "95,90.3"})
	if !errors.Is(err, barikoi.ErrInvalidLatitude) {
		t.Fatalf("out-of-bounds: got %v, want ErrInvalidLatitude", err)
	}
}
