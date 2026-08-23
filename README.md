# Barikoi APIs - Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/barikoi/barikoiapis-golang.svg)](https://pkg.go.dev/github.com/barikoi/barikoiapis-golang)
[![Release](https://img.shields.io/github/release/barikoi/barikoiapis-golang.svg?style=flat-square)](https://github.com/barikoi/barikoiapis-golang/releases/latest)
[![CI](https://github.com/barikoi/barikoiapis-golang/actions/workflows/ci.yml/badge.svg)](https://github.com/barikoi/barikoiapis-golang/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/barikoi/barikoiapis-golang/badge.svg)](https://codecov.io/gh/barikoi/barikoiapis-golang)
[![Go Report Card](https://goreportcard.com/badge/github.com/barikoi/barikoiapis-golang?style=flat-square)](https://goreportcard.com/report/github.com/barikoi/barikoiapis-golang)
[![Go Version](https://img.shields.io/github/go-mod/go-version/barikoi/barikoiapis-golang?style=flat-square)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](https://opensource.org/licenses/MIT)

## Description

**Barikoi APIs** is the official Go SDK for [Barikoi Location Services](https://barikoi.com). Built on auto-generated code from the official [OpenAPI specification](openapi/barikoi-api-spec.yaml), it provides a type-safe interface for accessing a wide range of location-based services including search, geocoding, reverse geocoding, routing, and more.

This SDK is the Go counterpart of the TypeScript/JavaScript SDK [`barikoiapis`](https://www.npmjs.com/package/barikoiapis) and mirrors its method names, defaults, validation rules, and error types — developers switching between the two get the same experience.

## Features

- **Auto-generated from the OpenAPI spec** — endpoints, parameters, and request bodies always in sync with the API
- **Built on Go's native `net/http`** — no third-party HTTP stack; the transport is the standard library's `*http.Client`, replaceable with your own via `WithHTTPClient`
- **100% type-safe** — full struct types everywhere; tolerant decoding (`FlexFloat`/`FlexString`) absorbs the live API's number/string inconsistencies
- **Built-in validation** — inputs validated with Zod-equivalent rules before any HTTP call, with clear field-level errors
- **Runtime API key management** — rotate keys without reinitializing the client
- **Context-aware** — every method takes a `context.Context` first; cancellation and deadlines propagate to the HTTP request
- **Configurable timeouts** — 30s default, changeable at construction and runtime
- **Custom error types** — `BarikoiError`, `ValidationError`, `TimeoutError`, distinguished with `errors.As`
- **Concurrent-safe** — share one client across goroutines

## Getting Started

### Get Barikoi API Key

To access Barikoi's API services, you need to:

1. Register on the [Barikoi Developer Dashboard](https://developer.barikoi.com/register)
2. Verify with your phone number
3. Claim your API key

Once registered, you'll be able to access the full suite of Barikoi API services. If you exceed the free usage limits, you'll need to subscribe to a paid plan.

### API Key Handling

The Barikoi API authenticates each request with an API key sent as a query parameter. Never hardcode or commit your key — always load it from the environment:

```go
apiKey := os.Getenv("BARIKOI_API_KEY")
```

## Installation

```bash
go get github.com/barikoi/barikoiapis-golang
```

```go
import barikoi "github.com/barikoi/barikoiapis-golang"
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	barikoi "github.com/barikoi/barikoiapis-golang"
)

func main() {
	c, err := barikoi.NewClient(os.Getenv("BARIKOI_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	result, err := c.Autocomplete(context.Background(), &barikoi.AutocompleteRequest{Q: "Dhaka"})
	if err != nil {
		log.Fatal(err)
	}
	for _, place := range result.Places {
		fmt.Println(place.Address)
	}
}
```

A runnable program covering all 11 endpoints lives in [`examples/basic`](examples/basic/main.go):

```bash
BARIKOI_API_KEY=... go run ./examples/basic
```

## API Reference

Per-method notes and type declarations below.

> `FlexFloat` and `FlexString` decode fields the live API returns inconsistently — as JSON numbers or strings — so a `float64`/`string` value always lands in your struct.

1. [Autocomplete](#autocomplete) - Search for places with autocomplete suggestions
2. [Reverse Geocoding](#reverse-geocoding) - Convert coordinates to addresses
3. [Nearby Places](#nearby-places) - Find places within a radius
4. [Geocode (Rupantor)](#geocode-rupantor) - Format and geocode addresses
5. [Search Place](#search-place) - Search for places with session management
6. [Place Details](#place-details) - Get detailed place information
7. [Route Overview](#route-overview) - Get route information between points
8. [Calculate Route](#calculate-route) - Detailed route with turn-by-turn instructions
9. [Route Optimization](#route-optimization) - Optimize routes through waypoints
10. [Snap to Road](#snap-to-road) - Find nearest point on road network
11. [Check Nearby](#check-nearby) - Verify proximity within a radius

---

### Autocomplete

Search for places with autocomplete suggestions. Returns matching places with addresses in `English` and `Bangla`, coordinates, and place details. Use for search boxes, address forms, and location pickers.

```go
result, err := c.Autocomplete(ctx, &barikoi.AutocompleteRequest{
	Q:      "Dhaka",
	Bangla: barikoi.BoolPtr(false), // omit Bangla fields (default: true)
})
places := result.Places
```

<details>
<summary>Type Definitions</summary>

```go
// Q is required. Bangla defaults to true when nil, matching the TypeScript SDK.
type AutocompleteRequest struct {
	Q      string
	Bangla *bool
}

type AutocompletePlace struct {
	ID        int64      `json:"id"`
	Longitude FlexFloat  `json:"longitude"`
	Latitude  FlexFloat  `json:"latitude"`
	Address   string     `json:"address"`
	AddressBn string     `json:"address_bn"`
	City      string     `json:"city"`
	CityBn    string     `json:"city_bn"`
	Area      string     `json:"area"`
	AreaBn    string     `json:"area_bn"`
	District  string     `json:"district"`
	PostCode  FlexString `json:"postCode"`
	PType     string     `json:"pType"`
	SubType   string     `json:"subType"`
	UCode     string     `json:"uCode"`
}

type AutocompleteResponse struct {
	Places []AutocompletePlace `json:"places"`
	Status int                 `json:"status"`
}
```
</details>

### Reverse Geocoding

Convert coordinates to human-readable addresses with administrative details (district, division, thana) in `English` and `Bangla`. Use for displaying user location, delivery addresses, and location tagging.

**IMPORTANT:** ⚠️ Enabling optional parameters consumes additional API credits. Request only essential parameters.

```go
result, err := c.ReverseGeocode(ctx, &barikoi.ReverseGeocodeRequest{
	Latitude: 23.8103,
	Longitude: 90.4125,
	District: true,
	Bangla:   true,
})

place := result.Place
```

<details>
<summary>Type Definitions</summary>

```go
// Latitude/Longitude are required. Each boolean flag opts in to extra
// response fields; CountryCode defaults to "BD" when empty.
type ReverseGeocodeRequest struct {
	Latitude     float64
	Longitude    float64
	CountryCode  string // two-letter ISO Alpha-2, e.g. "BD"
	Country      bool
	District     bool
	PostCode     bool
	SubDistrict  bool
	Union        bool
	Pauroshova   bool
	LocationType bool
	Division     bool
	Address      bool
	Area         bool
	Bangla       bool
	Thana        bool
}

type ReverseGeocodePlace struct {
	ID                   int64      `json:"id"`
	DistanceWithinMeters FlexFloat  `json:"distance_within_meters"`
	Address              string     `json:"address"`
	Area                 string     `json:"area"`
	City                 string     `json:"city"`
	PostCode             FlexString `json:"postCode"`
	AddressBn            string     `json:"address_bn"`
	AreaBn               string     `json:"area_bn"`
	CityBn               string     `json:"city_bn"`
	Country              string     `json:"country"`
	Division             string     `json:"division"`
	District             string     `json:"district"`
	SubDistrict          string     `json:"sub_district"`
	Union                string     `json:"union"`
	Pauroshova           string     `json:"pauroshova"`
	LocationType         string     `json:"location_type"`
	Thana                string     `json:"thana"`
	ThanaBn              string     `json:"thana_bn"`
	AddressComponents    struct {
		PlaceName string `json:"place_name"`
		House     string `json:"house"`
		Road      string `json:"road"`
	} `json:"address_components"`
	AreaComponents struct {
		Area    string `json:"area"`
		SubArea string `json:"sub_area"`
	} `json:"area_components"`
}

type ReverseGeocodeResponse struct {
	Place  ReverseGeocodePlace `json:"place"`
	Status int                 `json:"status"`
}
```
</details>

### Nearby Places

Find places within a specified radius. Returns nearby locations sorted by distance with names, addresses, and coordinates. Perfect for "nearby stores", POI discovery, restaurant finders, and ATM locators.

```go
result, err := c.Nearby(ctx, &barikoi.NearbyRequest{
	Latitude:  23.87188719,
	Longitude: 90.38305163,
	Radius:    1,  // kilometers (default: 0.5)
	Limit:     20, // default: 10
})

places := result.Places
```

<details>
<summary>Type Definitions</summary>

```go
// Latitude/Longitude are required. Radius is in kilometers [0.1, 100],
// defaulting to 0.5 when zero; Limit is [1, 100], defaulting to 10.
// Both are sent as path parameters.
type NearbyRequest struct {
	Latitude  float64
	Longitude float64
	Radius    float64
	Limit     int
}

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

type NearbyResponse struct {
	Places []NearbyPlace `json:"places"`
	Status int           `json:"status"`
}
```
</details>

### Geocode (Rupantor)

Validate and format addresses with completeness status and confidence score. Returns standardized address format. Use for checkout validation, delivery verification, and CRM data cleaning.

**Note:** Uses 2 API calls internally — one Rupantor request consumes two Geocode API credits.

```go
result, err := c.Geocode(ctx, &barikoi.GeocodeRequest{
	Q: "house 23, road 5, mirpur dhaka",
	Thana:    true, // sent as "yes"
	District: true,
})

status     := result.AddressStatus              // "complete" | "incomplete"
confidence := result.ConfidenceScorePercentage
fixed      := result.FixedAddress
```

<details>
<summary>Type Definitions</summary>

```go
// Q is required. The boolean flags request extra fields and are sent as
// "yes" only when true, matching the TypeScript SDK.
type GeocodeRequest struct {
	Q        string
	Thana    bool
	District bool
	Bangla   bool
}

// The API returns the address under "address" or "Address" depending on
// the query; both are accepted and Address is populated from either.
type GeocodedPlace struct {
	ID                int64      `json:"id"`
	UCode             string     `json:"uCode"`
	PlaceCode         string     `json:"place_code"`
	Address           string     `json:"address"`
	AddressTitle      string     `json:"Address"`
	AddressBn         string     `json:"address_bn"`
	BusinessName      string     `json:"business_name"`
	Area              string     `json:"area"`
	AreaBn            string     `json:"area_bn"`
	SubArea           string     `json:"sub_area"`
	SuperSubArea      string     `json:"super_sub_area"`
	City              string     `json:"city"`
	CityBn            string     `json:"city_bn"`
	District          string     `json:"district"`
	SubDistrict       string     `json:"sub_district"`
	Thana             string     `json:"thana"`
	PType             string     `json:"pType"`
	SubType           string     `json:"subType"`
	PostCode          FlexString `json:"postCode"`
	Postcode          FlexString `json:"postcode"`
	Longitude         FlexFloat  `json:"longitude"`
	Latitude          FlexFloat  `json:"latitude"`
	GeoLocation       []float64  `json:"geo_location"` // [longitude, latitude]
	PopularityRanking int        `json:"popularity_ranking"`
}

type GeocodeResponse struct {
	GivenAddress              string        `json:"given_address"`
	FixedAddress              string        `json:"fixed_address"`
	BanglaAddress             string        `json:"bangla_address"`
	AddressStatus             string        `json:"address_status"` // "complete" | "incomplete"
	GeocodedAddress           GeocodedPlace `json:"geocoded_address"`
	ConfidenceScorePercentage FlexFloat     `json:"confidence_score_percentage"`
	Status                    int           `json:"status"`
}
```
</details>

### Search Place

Search for places and get unique place codes with a session ID. Returns matching places with addresses. Use for business search, landmark lookup, and location selection.

**Note:** Each request generates a new session ID required for Place Details API.

```go
result, err := c.SearchPlace(ctx, &barikoi.SearchPlaceRequest{Q: "barikoi"})

sessionID := result.SessionID
places := result.Places
```

<details>
<summary>Type Definitions</summary>

```go
// Q is required.
type SearchPlaceRequest struct {
	Q string
}

type SearchPlaceResult struct {
	Address   string `json:"address"`
	PlaceCode string `json:"place_code"`
}

// SessionID must be passed to PlaceDetails for the same search session.
type SearchPlaceResponse struct {
	Places    []SearchPlaceResult `json:"places"`
	SessionID string              `json:"session_id"`
	Status    int                 `json:"status"`
}
```
</details>

### Place Details

Get detailed place information using a place code and session ID. Returns complete address and coordinates. Use after Search Place to fetch full location data.

**Note:** Requires place code and session ID from Search Place request. SessionID must be a UUID.

```go
details, err := c.PlaceDetails(ctx, &barikoi.PlaceDetailsRequest{
	PlaceCode: "BKOI2017",
	SessionID: sessionID,
})

place := details.Place
```

<details>
<summary>Type Definitions</summary>

```go
// PlaceCode comes from SearchPlace; SessionID comes from the same
// SearchPlace response and must be a UUID. Both are required.
type PlaceDetailsRequest struct {
	PlaceCode string
	SessionID string
}

type PlaceDetailsPlace struct {
	Address   string    `json:"address"`
	PlaceCode string    `json:"place_code"`
	Latitude  FlexFloat `json:"latitude"`
	Longitude FlexFloat `json:"longitude"`
}

type PlaceDetailsResponse struct {
	SessionID string            `json:"session_id"`
	Status    int               `json:"status"`
	Place     PlaceDetailsPlace `json:"place"`
}
```
</details>

### Route Overview

Get route information between geographical points. Returns route geometry, distance, duration, and waypoints in polyline or GeoJSON format. Use for displaying routes, calculating distances, and showing ETAs.

**Note:** Coordinates must be in `longitude,latitude` format.

```go
result, err := c.RouteOverview(ctx, &barikoi.RouteOverviewRequest{
	Coordinates: "90.4125,23.8103;90.4000,23.8000", // "lon,lat;lon,lat"
	Geometries:  "geojson",                          // default: "polyline"
})

route       := result.Routes[0]
distanceKm  := route.Distance / 1000 // meters
durationMin := route.Duration / 60   // seconds
```

<details>
<summary>Type Definitions</summary>

```go
// Coordinates is required, formatted "lon,lat;lon,lat" (at least two
// pairs). Geometries is one of "polyline" (default), "polyline6", or
// "geojson". Profile is "car" (default) or "foot".
type RouteOverviewRequest struct {
	Coordinates string
	Geometries  string
	Profile     string
}

// OSRM-style response.
type RouteOverviewResponse struct {
	Code      string     `json:"code"`
	Routes    []Route    `json:"routes"`
	Waypoints []Waypoint `json:"waypoints"`
}

type Route struct {
	Geometry   PolylineOrGeoJSON `json:"geometry"` // string, or raw GeoJSON when geometries=geojson
	Legs       []RouteLeg        `json:"legs"`
	Distance   float64           `json:"distance"` // meters
	Duration   float64           `json:"duration"` // seconds
	WeightName string            `json:"weight_name"`
	Weight     float64           `json:"weight"`
}

type RouteLeg struct {
	// Steps' schema is not documented by Barikoi; each step is returned raw.
	Steps    []map[string]any `json:"steps"`
	Distance float64          `json:"distance"`
	Duration float64          `json:"duration"`
	Summary  string           `json:"summary"`
	Weight   float64          `json:"weight"`
}

type Waypoint struct {
	Hint     string    `json:"hint"`
	Distance float64   `json:"distance"`
	Name     string    `json:"name"`
	Location []float64 `json:"location"` // [longitude, latitude]
}
```
</details>

### Calculate Route

Get detailed route information powered by GraphHopper routing engine. Returns comprehensive route data including path coordinates, distance, travel time, turn-by-turn instructions with street names, and elevation data (ascend/descend). Use for GPS navigation apps, route planning, delivery optimization, and mapping applications requiring detailed routing information.

```go
result, err := c.CalculateRoute(ctx, &barikoi.CalculateRouteRequest{
	Start:       barikoi.Coordinate{Latitude: 23.8103, Longitude: 90.4125},
	Destination: barikoi.Coordinate{Latitude: 23.8, Longitude: 90.4},
	Type:        "gh", // GraphHopper engine (default)
	Profile:     "car",
})

hints := &result.Hints
paths := result.Paths
```

<details>
<summary>Type Definitions</summary>

```go
type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Type is the routing engine, defaulting to "gh" (the only documented
// value). Profile is "car" (default), "bike", or "motorcycle" and is
// sent only when set.
type CalculateRouteRequest struct {
	Start       Coordinate
	Destination Coordinate
	Type        string
	Profile     string
}

// GraphHopper-format response.
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
	Details          json.RawMessage    `json:"details"`
	Ascend           float64            `json:"ascend"`
	Descend          float64            `json:"descend"`
	SnappedWaypoints RouteGeometry      `json:"snapped_waypoints"`
}

// RouteGeometry holds the geometry in either of the two forms the API
// returns: an encoded polyline string when points_encoded is true, or a
// GeoJSON LineString when it is false. Exactly one of Polyline and GeoJSON
// is set.
type RouteGeometry struct {
	Polyline string
	GeoJSON  *GeoJSONLineString
}

type RouteInstruction struct {
	Distance   float64 `json:"distance"`
	Heading    float64 `json:"heading"`
	Sign       int     `json:"sign"`
	Interval   []int   `json:"interval"`
	Text       string  `json:"text"`
	Time       float64 `json:"time"`        // milliseconds
	StreetName string  `json:"street_name"` // "" when the API reports null
}

type GeoJSONLineString struct {
	Type        string      `json:"type"` // "LineString"
	Coordinates [][]float64 `json:"coordinates"` // [longitude, latitude] pairs
}
```
</details>

### Route Optimization

Optimize a route from source to destination through 1–50 waypoints, powered by the GraphHopper routing engine. Returns the same response format as [Calculate Route](#calculate-route). Use for delivery route planning, multi-stop optimization, and field-force routing. Waypoints are visited in ascending `id` order.

```go
result, err := c.OptimizeRoute(ctx, &barikoi.OptimizeRouteRequest{
	Source:      "23.8103,90.4125", // "lat,lon"
	Destination: "23.7461,90.3742",
	GeoPoints: []barikoi.OptimizeRoutePoint{
		{ID: 1, Point: "23.7925,90.4078"},
		{ID: 2, Point: "23.7609,90.3805"},
	},
	Profile: "car", // default
})

paths := result.Paths // RoutingResponse, same as CalculateRoute
```

<details>
<summary>Type Definitions</summary>

```go
// Source/Destination and each waypoint are formatted "lat,lon".
// Profile is "car" (default), "bike", "foot", or "motorcycle".
type OptimizeRouteRequest struct {
	Source      string
	Destination string
	Profile     string
	GeoPoints   []OptimizeRoutePoint
}

// Waypoints are visited in ascending ID order; between 1 and 50 allowed.
type OptimizeRoutePoint struct {
	ID    int    `json:"id"`
	Point string `json:"point"` // "lat,lon"
}

// Response is RoutingResponse, shared with CalculateRoute — see its
// Type Definitions above.
```
</details>

### Snap to Road

Find the nearest road point to given coordinates. Returns snapped coordinates and distance to road. Use for vehicle tracking, GPS trace alignment, and ride-sharing location accuracy.

```go
result, err := c.SnapToRoad(ctx, &barikoi.SnapToRoadRequest{
	Point: "23.8103,90.4125", // Format: "latitude,longitude"
})

snapped   := result.Coordinates // [longitude, latitude]
distance  := result.Distance    // meters
```

<details>
<summary>Type Definitions</summary>

```go
// Point is required and formatted "lat,lon".
type SnapToRoadRequest struct {
	Point string
}

type SnapToRoadResponse struct {
	Coordinates []float64 `json:"coordinates"` // [longitude, latitude]
	Distance    float64  `json:"distance"`     // meters
	Type        string   `json:"type"`         // "Point"
}
```
</details>

### Check Nearby

Verify if a location is within a specified radius. Returns "Inside geo fence" or "Outside geo fence" status. Perfect for delivery notifications, driver alerts, proximity triggers, and employee check-in from devices in HR applications.

```go
result, err := c.CheckNearby(ctx, &barikoi.CheckNearbyRequest{
	CurrentLatitude:      23.8103,
	CurrentLongitude:     90.4125,
	DestinationLatitude:  23.8,
	DestinationLongitude: 90.4,
	Radius:               100, // meters [10, 1000]
})

isInside := result.Data != nil // nil when outside the geo fence
```

<details>
<summary>Type Definitions</summary>

```go
// All fields are required. Radius is in meters and must be within
// [10, 1000].
type CheckNearbyRequest struct {
	CurrentLatitude      float64
	CurrentLongitude     float64
	DestinationLatitude  float64
	DestinationLongitude float64
	Radius               int // meters
}

// Data is nil when the destination is outside the radius.
type CheckNearbyResponse struct {
	Message string            `json:"message"` // "Inside geo fence" | "Outside geo fence"
	Status  int               `json:"status"`
	Data    *CheckNearbyPlace `json:"data"`
}

type CheckNearbyPlace struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Radius    string `json:"radius"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	UserID    int64  `json:"user_id"`
}
```
</details>

### API Key Management

```go
// Set during initialization (from the environment — never hardcode)
c, err := barikoi.NewClient(os.Getenv("BARIKOI_API_KEY"))

// Update API key at runtime
c.SetAPIKey(os.Getenv("BARIKOI_API_KEY_V2"))

// Get current API key
key := c.GetAPIKey()
```

### Timeout Configuration

```go
// Set timeout during initialization (default: 30s)
c, err := barikoi.NewClient(os.Getenv("BARIKOI_API_KEY"),
	barikoi.WithTimeout(60*time.Second))

// Update at runtime
c.SetTimeout(90 * time.Second)
```

### Custom Base URL

```go
c, err := barikoi.NewClient(os.Getenv("BARIKOI_API_KEY"),
	barikoi.WithBaseURL("https://custom-endpoint.barikoi.xyz"))
```

Base URLs must use `https://`. For local development, opt in explicitly with `barikoi.WithAllowInsecure()`. A custom `*http.Client` can be supplied with `barikoi.WithHTTPClient`.

## Error Handling

Three error types, distinguished with `errors.As`:

- `*barikoi.BarikoiError` — any non-2xx API response (`Message`, `StatusCode`, `Code`, `Details`); predicates `IsAuthError()` (401/403), `IsRateLimitError()` (429), `IsServerError()` (5xx)
- `*barikoi.ValidationError` — client-side validation failure (`Field`, `Message`); no HTTP call was made
- `*barikoi.TimeoutError` — the request was cancelled or timed out

```go
_, err := c.Nearby(ctx, req)
var apiErr *barikoi.BarikoiError
if errors.As(err, &apiErr) && apiErr.IsRateLimitError() {
	// back off and retry
}
var valErr *barikoi.ValidationError
if errors.As(err, &valErr) {
	log.Printf("bad field: %s", valErr.Field)
}
```

## Development

```bash
make check             # gofmt + go vet + unit tests (no key needed)
make test-integration  # live API tests; needs BARIKOI_API_KEY (or .env)
make codegen           # regenerate gen/ from openapi/barikoi-api-spec.yaml
```

The Makefile runs everything in Docker (`golang:1.22`) when no local Go toolchain is installed. Copy `env.example` to `.env` for integration tests. See [CONTRIBUTING.md](CONTRIBUTING.md) for architecture and workflow.

---

## Documentation

- [API Documentation](https://docs.barikoi.com/api) - Official Barikoi API docs
- [OpenAPI Specification](openapi/barikoi-api-spec.yaml) - Source of truth
- [TypeScript SDK](https://www.npmjs.com/package/barikoiapis) - The `barikoiapis` npm package this SDK mirrors

## Support Resources

- [Barikoi Website](https://www.barikoi.com/)
- [Developer Portal](https://developer.barikoi.com/)
- [Documentation](https://docs.barikoi.com/docs/maps-api)
- [Support Email](mailto:hello@barikoi.com)

## License

This library is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
