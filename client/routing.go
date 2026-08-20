package client

import (
	"context"
	"net/url"
)

// Coordinate is a latitude/longitude pair used in routing requests.
type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// RouteOverviewRequest describes a basic route request. Coordinates is
// required and formatted as "lon,lat;lon,lat". Geometries is one of
// "polyline" (default), "polyline6", or "geojson". Profile is "car"
// (default) or "foot".
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
	if err := requireString("coordinates", req.Coordinates); err != nil {
		return nil, err
	}
	q := url.Values{}
	if req.Geometries != "" {
		q.Set("geometries", req.Geometries)
	}
	if req.Profile != "" {
		q.Set("profile", req.Profile)
	}

	var resp RouteOverviewResponse
	if err := c.doGet(ctx, "/v2/api/route/"+req.Coordinates, q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CalculateRouteRequest describes a detailed routing request with
// turn-by-turn maneuvers. Type is the routing algorithm (e.g. "vh", required
// by the API). Profile is "car" (default), "bike", or "motorcycle".
type CalculateRouteRequest struct {
	Start       Coordinate
	Destination Coordinate
	Type        string
	Profile     string
}

// RouteManeuver is a single turn-by-turn instruction.
type RouteManeuver struct {
	Type                                int      `json:"type"`
	Instruction                         string   `json:"instruction"`
	VerbalSuccinctTransitionInstruction string   `json:"verbal_succinct_transition_instruction"`
	VerbalPreTransitionInstruction      string   `json:"verbal_pre_transition_instruction"`
	VerbalPostTransitionInstruction     string   `json:"verbal_post_transition_instruction"`
	Time                                float64  `json:"time"`
	Length                              float64  `json:"length"`
	Cost                                float64  `json:"cost"`
	BeginShapeIndex                     int      `json:"begin_shape_index"`
	EndShapeIndex                       int      `json:"end_shape_index"`
	VerbalMultiCue                      bool     `json:"verbal_multi_cue"`
	TravelMode                          string   `json:"travel_mode"`
	TravelType                          string   `json:"travel_type"`
	StreetNames                         []string `json:"street_names"`
}

// RouteTripLocation is a snapped start/destination point of a trip.
type RouteTripLocation struct {
	Type          string  `json:"type"`
	Lat           float64 `json:"lat"`
	Lon           float64 `json:"lon"`
	OriginalIndex int     `json:"original_index"`
}

// RouteTripSummary summarizes a trip or leg: time in seconds, length in the
// units reported by the API (miles by default).
type RouteTripSummary struct {
	HasTimeRestrictions bool    `json:"has_time_restrictions"`
	HasToll             bool    `json:"has_toll"`
	HasHighway          bool    `json:"has_highway"`
	HasFerry            bool    `json:"has_ferry"`
	MinLat              float64 `json:"min_lat"`
	MinLon              float64 `json:"min_lon"`
	MaxLat              float64 `json:"max_lat"`
	MaxLon              float64 `json:"max_lon"`
	Time                float64 `json:"time"`
	Length              float64 `json:"length"`
	Cost                float64 `json:"cost"`
}

// RouteTripLeg is one leg of a trip with its maneuvers.
type RouteTripLeg struct {
	Maneuvers []RouteManeuver  `json:"maneuvers"`
	Summary   RouteTripSummary `json:"summary"`
	Shape     string           `json:"shape"` // encoded polyline
}

// RouteTrip is the full trip returned by CalculateRoute.
type RouteTrip struct {
	Locations     []RouteTripLocation `json:"locations"`
	Legs          []RouteTripLeg      `json:"legs"`
	Summary       RouteTripSummary    `json:"summary"`
	StatusMessage string              `json:"status_message"`
	Status        int                 `json:"status"`
	Units         string              `json:"units"`
	Language      string              `json:"language"`
}

// CalculateRouteResponse is the response of CalculateRoute.
type CalculateRouteResponse struct {
	Trip RouteTrip `json:"trip"`
	ID   string    `json:"id"`
}

// CalculateRoute returns a detailed route with turn-by-turn maneuvers via
// POST /v2/api/routing.
func (c *Client) CalculateRoute(ctx context.Context, req *CalculateRouteRequest) (*CalculateRouteResponse, error) {
	if err := validateCoords(req.Start.Latitude, req.Start.Longitude); err != nil {
		return nil, err
	}
	if err := validateCoords(req.Destination.Latitude, req.Destination.Longitude); err != nil {
		return nil, err
	}
	q := url.Values{}
	if req.Type != "" {
		q.Set("type", req.Type)
	}
	if req.Profile != "" {
		q.Set("profile", req.Profile)
	}
	body := struct {
		Data struct {
			Start       Coordinate `json:"start"`
			Destination Coordinate `json:"destination"`
		} `json:"data"`
	}{}
	body.Data.Start = req.Start
	body.Data.Destination = req.Destination

	var resp CalculateRouteResponse
	if err := c.doPostJSON(ctx, "/v2/api/routing", q, &body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// OptimizeRoutePoint is a waypoint with its sort ID. Waypoints are visited in
// ascending ID order; up to 50 are allowed.
type OptimizeRoutePoint struct {
	ID    int    `json:"id"`
	Point string `json:"point"` // "lat,lon"
}

// OptimizeRouteRequest describes a route optimization from Source to
// Destination through GeoPoints, each point formatted as "lat,lon".
// Profile is "car" (default), "bike", or "motorcycle".
type OptimizeRouteRequest struct {
	Source      string
	Destination string
	Profile     string
	GeoPoints   []OptimizeRoutePoint
}

// OptimizedPathInstruction is one navigation instruction of an optimized path.
type OptimizedPathInstruction struct {
	Distance   float64 `json:"distance"`
	Sign       int     `json:"sign"`
	Interval   []int   `json:"interval"`
	Text       string  `json:"text"`
	Time       float64 `json:"time"`
	StreetName *string `json:"street_name"`
}

// OptimizedPath is one optimized path (GraphHopper-style).
type OptimizedPath struct {
	Distance         float64                    `json:"distance"` // meters
	Weight           float64                    `json:"weight"`
	Time             float64                    `json:"time"` // milliseconds
	Transfers        int                        `json:"transfers"`
	PointsEncoded    bool                       `json:"points_encoded"`
	BBox             []float64                  `json:"bbox"`
	Points           string                     `json:"points"` // encoded polyline
	Instructions     []OptimizedPathInstruction `json:"instructions"`
	Ascend           float64                    `json:"ascend"`
	Descend          float64                    `json:"descend"`
	SnappedWaypoints string                     `json:"snapped_waypoints"`
}

// OptimizeRouteResponse is the response of OptimizeRoute.
type OptimizeRouteResponse struct {
	Hints struct {
		VisitedNodesSum     int     `json:"visited_nodes.sum"`
		VisitedNodesAverage float64 `json:"visited_nodes.average"`
	} `json:"hints"`
	Info struct {
		Copyrights []string `json:"copyrights"`
		Took       float64  `json:"took"`
	} `json:"info"`
	Paths []OptimizedPath `json:"paths"`
}

// OptimizeRoute optimizes a route through up to 50 waypoints via
// POST /v2/api/route/optimized. The API key is sent both as the api_key
// query parameter (as on every request) and in the JSON body, as the
// endpoint's documentation shows it in both places.
func (c *Client) OptimizeRoute(ctx context.Context, req *OptimizeRouteRequest) (*OptimizeRouteResponse, error) {
	if err := requireString("source", req.Source); err != nil {
		return nil, err
	}
	if err := requireString("destination", req.Destination); err != nil {
		return nil, err
	}
	body := struct {
		APIKey      string               `json:"api_key"`
		Source      string               `json:"source"`
		Destination string               `json:"destination"`
		Profile     string               `json:"profile,omitempty"`
		GeoPoints   []OptimizeRoutePoint `json:"geo_points,omitempty"`
	}{
		APIKey:      c.GetAPIKey(),
		Source:      req.Source,
		Destination: req.Destination,
		Profile:     req.Profile,
		GeoPoints:   req.GeoPoints,
	}

	var resp OptimizeRouteResponse
	if err := c.doPostJSON(ctx, "/v2/api/route/optimized", nil, &body, &resp); err != nil {
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
	if err := requireString("point", req.Point); err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("point", req.Point)

	var resp SnapToRoadResponse
	if err := c.doGet(ctx, "/v2/api/routing/nearest", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
