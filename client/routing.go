package client

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/barikoi/barikoiapis-golang/gen"
)

// Coordinate is a latitude/longitude pair used in routing requests.
type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// RouteOverviewRequest describes a basic route request. Coordinates is
// required and formatted as "lon,lat;lon,lat" (at least two pairs).
// Geometries is one of "polyline" (default), "polyline6", or "geojson".
// Profile is "car" (default) or "foot".
type RouteOverviewRequest struct {
	Coordinates string
	Geometries  string
	Profile     string
}

// RouteLeg is one leg of a route.
type RouteLeg struct {
	// Steps' schema is not documented by Barikoi; each step is returned raw.
	Steps    []map[string]any `json:"steps"`
	Distance float64          `json:"distance"`
	Duration float64          `json:"duration"`
	Summary  string           `json:"summary"`
	Weight   float64          `json:"weight"`
}

// Route is a single route between the requested waypoints.
type Route struct {
	Geometry   PolylineOrGeoJSON `json:"geometry"`
	Legs       []RouteLeg        `json:"legs"`
	Distance   float64           `json:"distance"` // meters
	Duration   float64           `json:"duration"` // seconds
	WeightName string            `json:"weight_name"`
	Weight     float64           `json:"weight"`
}

// Waypoint is a snapped input point of a route.
type Waypoint struct {
	Hint     string    `json:"hint"`
	Distance float64   `json:"distance"`
	Name     string    `json:"name"`
	Location []float64 `json:"location"` // [longitude, latitude]
}

// RouteOverviewResponse is the OSRM-style response of RouteOverview.
type RouteOverviewResponse struct {
	Code      string     `json:"code"`
	Routes    []Route    `json:"routes"`
	Waypoints []Waypoint `json:"waypoints"`
}

// RouteOverview returns the basic route between coordinates via
// GET /v2/api/route/{coordinates}.
func (c *Client) RouteOverview(ctx context.Context, req *RouteOverviewRequest) (*RouteOverviewResponse, error) {
	if err := validateCoordinatesString(req.Coordinates); err != nil {
		return nil, err
	}
	geometries := req.Geometries
	if geometries == "" {
		geometries = "polyline" // default, matching the TypeScript SDK
	}
	if err := validateEnum("geometries", geometries, "polyline", "polyline6", "geojson"); err != nil {
		return nil, err
	}
	profile := req.Profile
	if profile == "" {
		profile = "car" // default, matching the TypeScript SDK
	}
	if err := validateEnum("profile", profile, "car", "foot"); err != nil {
		return nil, err
	}
	params := &gen.RouteOverviewParams{
		ApiKey:     c.apiKeyParam(),
		Geometries: (*gen.RouteOverviewParamsGeometries)(&geometries),
		Profile:    (*gen.RouteOverviewParamsProfile)(&profile),
	}

	var resp RouteOverviewResponse
	err := c.do(ctx, func(ctx context.Context) (*http.Response, error) {
		return c.gen.RouteOverview(ctx, gen.Coordinates(req.Coordinates), params)
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CalculateRouteRequest describes a detailed routing request with
// turn-by-turn instructions. Type is the routing engine and defaults to "gh"
// (the only value the API documents); Profile is "car" (default), "bike", or
// "motorcycle" and is sent only when set.
type CalculateRouteRequest struct {
	Start       Coordinate
	Destination Coordinate
	Type        string
	Profile     string
}

// RouteInstruction is one navigation instruction of a route path
// (GraphHopper format).
type RouteInstruction struct {
	Distance   float64 `json:"distance"`
	Heading    float64 `json:"heading"`
	Sign       int     `json:"sign"`
	Interval   []int   `json:"interval"`
	Text       string  `json:"text"`
	Time       float64 `json:"time"`        // milliseconds
	StreetName string  `json:"street_name"` // "" when the API reports null
}

// GeoJSONLineString is the GeoJSON geometry returned for route points when
// points_encoded is false.
type GeoJSONLineString struct {
	Type        string      `json:"type"`        // "LineString"
	Coordinates [][]float64 `json:"coordinates"` // [longitude, latitude] pairs
}

// RouteGeometry holds a route geometry in either of the two forms the API
// returns: an encoded polyline string when points_encoded is true, or a
// GeoJSON LineString object when it is false. Exactly one of Polyline and
// GeoJSON is set.
type RouteGeometry struct {
	Polyline string
	GeoJSON  *GeoJSONLineString
}

func (g *RouteGeometry) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		g.Polyline = s
		g.GeoJSON = nil
		return nil
	}
	return json.Unmarshal(data, &g.GeoJSON)
}

// RoutePath is one route path (GraphHopper format): distance in meters, time
// in milliseconds.
type RoutePath struct {
	Distance         float64            `json:"distance"` // meters
	Weight           float64            `json:"weight"`
	Time             float64            `json:"time"` // milliseconds
	Transfers        int                `json:"transfers"`
	PointsEncoded    bool               `json:"points_encoded"`
	BBox             []float64          `json:"bbox"` // [minLon, minLat, maxLon, maxLat]
	Points           RouteGeometry      `json:"points"`
	Instructions     []RouteInstruction `json:"instructions"`
	Legs             []map[string]any   `json:"legs"`
	Details          json.RawMessage    `json:"details"` // object or array depending on endpoint
	Ascend           float64            `json:"ascend"`
	Descend          float64            `json:"descend"`
	SnappedWaypoints RouteGeometry      `json:"snapped_waypoints"`
}

// RoutingResponse is the GraphHopper-format response shared by
// CalculateRoute and OptimizeRoute.
type RoutingResponse struct {
	Hints struct {
		VisitedNodesSum     float64 `json:"visited_nodes.sum"`
		VisitedNodesAverage float64 `json:"visited_nodes.average"`
	} `json:"hints"`
	Info struct {
		Copyrights        []string `json:"copyrights"`
		Took              float64  `json:"took"`
		RoadDataTimestamp string   `json:"road_data_timestamp"`
	} `json:"info"`
	Paths []RoutePath `json:"paths"`
}

// CalculateRoute returns a detailed route with turn-by-turn instructions via
// POST /v2/api/routing.
func (c *Client) CalculateRoute(ctx context.Context, req *CalculateRouteRequest) (*RoutingResponse, error) {
	if err := validateCoords(req.Start.Latitude, req.Start.Longitude); err != nil {
		return nil, err
	}
	if err := validateCoords(req.Destination.Latitude, req.Destination.Longitude); err != nil {
		return nil, err
	}
	routeType := req.Type
	if routeType == "" {
		routeType = "gh" // default and only documented value
	}
	if err := validateEnum("type", routeType, "gh"); err != nil {
		return nil, err
	}
	params := gen.CalculateRouteParams{
		ApiKey: c.apiKeyParam(),
		Type:   gen.CalculateRouteParamsType(routeType),
	}
	if req.Profile != "" {
		if err := validateEnum("profile", req.Profile, "car", "bike", "motorcycle"); err != nil {
			return nil, err
		}
		params.Profile = (*gen.CalculateRouteParamsProfile)(&req.Profile)
	}
	body := gen.CalculateRouteJSONRequestBody{}
	body.Data.Start.Latitude = req.Start.Latitude
	body.Data.Start.Longitude = req.Start.Longitude
	body.Data.Destination.Latitude = req.Destination.Latitude
	body.Data.Destination.Longitude = req.Destination.Longitude

	var resp RoutingResponse
	err := c.do(ctx, func(ctx context.Context) (*http.Response, error) {
		return c.gen.CalculateRoute(ctx, &params, body)
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// OptimizeRoutePoint is a waypoint with its sort ID. Waypoints are visited in
// ascending ID order; between 1 and 50 are allowed.
type OptimizeRoutePoint struct {
	ID    int    `json:"id"`
	Point string `json:"point"` // "lat,lon"
}

// OptimizeRouteRequest describes a route optimization from Source to
// Destination through GeoPoints, each point formatted as "lat,lon".
// Profile is "car" (default), "bike", "foot", or "motorcycle".
type OptimizeRouteRequest struct {
	Source      string
	Destination string
	Profile     string
	GeoPoints   []OptimizeRoutePoint
}

// OptimizeRoute optimizes a route through 1 to 50 waypoints via
// POST /v2/api/route/optimized. The API key is sent both as the api_key
// query parameter (as on every request) and in the JSON body, as the
// endpoint's documentation shows it in both places.
func (c *Client) OptimizeRoute(ctx context.Context, req *OptimizeRouteRequest) (*RoutingResponse, error) {
	if err := validatePointString("source", req.Source); err != nil {
		return nil, err
	}
	if err := validatePointString("destination", req.Destination); err != nil {
		return nil, err
	}
	if len(req.GeoPoints) < 1 || len(req.GeoPoints) > 50 {
		return nil, &ValidationError{Field: "geo_points", Message: "must contain between 1 and 50 points"}
	}
	for _, gp := range req.GeoPoints {
		if err := validatePointString("geo_points", gp.Point); err != nil {
			return nil, err
		}
	}
	profile := req.Profile
	if profile == "" {
		profile = "car" // default, matching the OpenAPI spec
	}
	if err := validateEnum("profile", profile, "car", "bike", "foot", "motorcycle"); err != nil {
		return nil, err
	}
	body := gen.OptimizeRouteJSONRequestBody{
		ApiKey:      c.GetAPIKey(),
		Source:      req.Source,
		Destination: req.Destination,
		Profile:     (*gen.RouteOptimizationBodyProfile)(&profile),
	}
	// The spec declares geo_points as an inline array of {id, point}; the
	// anonymous struct below is tag-identical to the generated field's type.
	geoPoints := make([]struct {
		Id    int    `json:"id"`
		Point string `json:"point"`
	}, len(req.GeoPoints))
	for i, gp := range req.GeoPoints {
		geoPoints[i].Id = gp.ID
		geoPoints[i].Point = gp.Point
	}
	body.GeoPoints = geoPoints

	var resp RoutingResponse
	err := c.do(ctx, func(ctx context.Context) (*http.Response, error) {
		// The spec declares optimizeRoute's api_key only in the request body;
		// the TypeScript SDK's generator injects the global apiKey security
		// scheme as a query parameter on every operation, so do the same here.
		return c.gen.OptimizeRoute(ctx, body, c.withAPIKeyQuery())
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SnapToRoadRequest holds the point to snap to the nearest road.
// Point is required and formatted as "lat,lon".
type SnapToRoadRequest struct {
	Point string
}

// SnapToRoadResponse is the response of SnapToRoad: the nearest road point
// ([lon, lat]) and its distance from the input point in meters.
type SnapToRoadResponse struct {
	Coordinates []float64 `json:"coordinates"`
	Distance    float64   `json:"distance"`
	Type        string    `json:"type"`
}

// SnapToRoad finds the nearest point on the road network to a single
// coordinate via GET /v2/api/routing/nearest.
func (c *Client) SnapToRoad(ctx context.Context, req *SnapToRoadRequest) (*SnapToRoadResponse, error) {
	if err := validatePointString("point", req.Point); err != nil {
		return nil, err
	}
	params := &gen.SnapToRoadParams{ApiKey: c.apiKeyParam(), Point: gen.Point(req.Point)}

	var resp SnapToRoadResponse
	err := c.do(ctx, func(ctx context.Context) (*http.Response, error) { return c.gen.SnapToRoad(ctx, params) }, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
