package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// SearchPlaceRequest holds a place search query. Q is required.
type SearchPlaceRequest struct {
	Q string
}

// SearchPlaceResult is one search hit: an address and its unique place code.
type SearchPlaceResult struct {
	Address   string `json:"address"`
	PlaceCode string `json:"place_code"`
}

// SearchPlaceResponse is the response of SearchPlace. SessionID must be
// passed to PlaceDetails for the same search session.
type SearchPlaceResponse struct {
	Places    []SearchPlaceResult `json:"places"`
	SessionID string              `json:"session_id"`
	Status    int                 `json:"status"`
}

// SearchPlace searches for places matching a query via
// GET /api/v2/search-place, returning place codes and a session ID for use
// with PlaceDetails.
func (c *Client) SearchPlace(ctx context.Context, req *SearchPlaceRequest) (*SearchPlaceResponse, error) {
	if err := requireString("q", req.Q); err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("q", req.Q)

	var resp SearchPlaceResponse
	if err := c.doGet(ctx, "/api/v2/search-place", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// PlaceDetailsRequest fetches a place by code. PlaceCode comes from
// SearchPlace; SessionID comes from the same SearchPlace response. Both are
// required.
type PlaceDetailsRequest struct {
	PlaceCode string
	SessionID string
}

// PlaceDetailsPlace is the place returned by PlaceDetails.
type PlaceDetailsPlace struct {
	Address   string    `json:"address"`
	PlaceCode string    `json:"place_code"`
	Latitude  FlexFloat `json:"latitude"`
	Longitude FlexFloat `json:"longitude"`
}

// PlaceDetailsResponse is the response of PlaceDetails.
type PlaceDetailsResponse struct {
	SessionID string            `json:"session_id"`
	Status    int               `json:"status"`
	Place     PlaceDetailsPlace `json:"place"`
}

// PlaceDetails returns the address and coordinates of a place via
// GET /api/v2/places.
func (c *Client) PlaceDetails(ctx context.Context, req *PlaceDetailsRequest) (*PlaceDetailsResponse, error) {
	if err := requireString("place_code", req.PlaceCode); err != nil {
		return nil, err
	}
	if err := requireString("session_id", req.SessionID); err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("place_code", req.PlaceCode)
	q.Set("session_id", req.SessionID)

	var resp PlaceDetailsResponse
	if err := c.doGet(ctx, "/api/v2/places", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// NearbyRequest searches for places near a point. Radius is in kilometers
// and must be within [0.1, 100]; Limit must be within [1, 100]. Both are
// sent as path parameters.
type NearbyRequest struct {
	Latitude  float64
	Longitude float64
	Radius    float64
	Limit     int
}

// NearbyPlace is one nearby place. Field tags match the live /v2 nearby
// response: id arrives as a numeric string, type/sub_type/place_code are
// snake_case (or lowercase) keys.
type NearbyPlace struct {
	ID               FlexString `json:"id"`
	Name             string     `json:"name"`
	DistanceInMeters FlexFloat  `json:"distance_in_meters"`
	Longitude        FlexFloat  `json:"longitude"`
	Latitude         FlexFloat  `json:"latitude"`
	PType            string     `json:"type"`
	Address          string     `json:"address"`
	Area             string     `json:"area"`
	City             string     `json:"city"`
	PostCode         FlexString `json:"postcode"`
	SubType          string     `json:"sub_type"`
	PlaceCode        string     `json:"place_code"`
}

// NearbyResponse is the response of Nearby.
type NearbyResponse struct {
	Places []NearbyPlace `json:"places"`
	Status int           `json:"status"`
}

// Nearby finds places within Radius kilometers of a point via
// GET /v2/api/search/nearby/{radius}/{limit}.
func (c *Client) Nearby(ctx context.Context, req *NearbyRequest) (*NearbyResponse, error) {
	if err := validateCoords(req.Latitude, req.Longitude); err != nil {
		return nil, err
	}
	if req.Radius < 0.1 || req.Radius > 100 {
		return nil, &ValidationError{Field: "radius", Message: "must be between 0.1 and 100 (kilometers)"}
	}
	if req.Limit < 1 || req.Limit > 100 {
		return nil, &ValidationError{Field: "limit", Message: "must be between 1 and 100"}
	}
	path := fmt.Sprintf("/v2/api/search/nearby/%s/%d", formatFloat(req.Radius), req.Limit)
	q := url.Values{}
	q.Set("latitude", formatFloat(req.Latitude))
	q.Set("longitude", formatFloat(req.Longitude))

	var resp NearbyResponse
	if err := c.doGet(ctx, path, q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CheckNearbyRequest checks whether the destination point is within Radius
// meters of the current point. Radius must be within [10, 1000].
type CheckNearbyRequest struct {
	CurrentLatitude      float64
	CurrentLongitude     float64
	DestinationLatitude  float64
	DestinationLongitude float64
	Radius               int // meters
}

// CheckNearbyPlace is the nearby geofence point reported by CheckNearby.
type CheckNearbyPlace struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Radius    string `json:"radius"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	UserID    int64  `json:"user_id"`
}

// CheckNearbyResponse is the response of CheckNearby. Data is nil when the
// destination is outside the radius.
type CheckNearbyResponse struct {
	Message string            `json:"message"`
	Status  int               `json:"status"`
	Data    *CheckNearbyPlace `json:"data"`
}

// CheckNearby reports whether the destination is within Radius meters of the
// current point via GET /v2/api/check/nearby.
func (c *Client) CheckNearby(ctx context.Context, req *CheckNearbyRequest) (*CheckNearbyResponse, error) {
	if err := validateCoords(req.CurrentLatitude, req.CurrentLongitude); err != nil {
		return nil, err
	}
	if err := validateCoords(req.DestinationLatitude, req.DestinationLongitude); err != nil {
		return nil, err
	}
	if req.Radius < 10 || req.Radius > 1000 {
		return nil, &ValidationError{Field: "radius", Message: "must be between 10 and 1000 (meters)"}
	}
	q := url.Values{}
	q.Set("current_latitude", formatFloat(req.CurrentLatitude))
	q.Set("current_longitude", formatFloat(req.CurrentLongitude))
	q.Set("destination_latitude", formatFloat(req.DestinationLatitude))
	q.Set("destination_longitude", formatFloat(req.DestinationLongitude))
	q.Set("radius", strconv.Itoa(req.Radius))

	var resp CheckNearbyResponse
	if err := c.doGet(ctx, "/v2/api/check/nearby", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
