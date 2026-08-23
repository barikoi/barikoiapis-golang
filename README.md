# Barikoi APIs - Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/barikoi/barikoiapis-golang.svg)](https://pkg.go.dev/github.com/barikoi/barikoiapis-golang)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Description

**Barikoi APIs** is the official Go SDK for [Barikoi Location Services](https://barikoi.com). It provides a type-safe interface for location-based services including search, geocoding, reverse geocoding, routing, and more.

This SDK is the Go counterpart of the TypeScript/JavaScript SDK [`barikoiapis`](https://www.npmjs.com/package/barikoiapis) and mirrors its method names, defaults, validation rules, and error types — developers switching between the two get the same experience. The HTTP client is [generated from the OpenAPI specification](openapi/barikoi-api-spec.yaml), the single source of truth for endpoints.

## Features

- **Generated from the OpenAPI spec** — endpoints, parameters, and request bodies always in sync with the API
- **100% type-safe** — full struct types everywhere; tolerant decoding absorbs the live API's number/string inconsistencies
- **Built-in validation** — inputs validated before any HTTP call, with clear field-level errors
- **Runtime API key management** — rotate keys without reinitializing the client
- **Context-aware** — every method takes a `context.Context` first
- **Configurable timeouts** — 30s default, changeable at construction and runtime
- **Custom error types** — `BarikoiError`, `ValidationError`, `TimeoutError`, distinguished with `errors.As`
- **Concurrent-safe** — share one client across goroutines

## Getting Started

### Get Barikoi API Key

1. Register on the [Barikoi Developer Dashboard](https://developer.barikoi.com/register)
2. Verify with your phone number
3. Claim your API key

If you exceed the free usage limits, you'll need to subscribe to a paid plan.

### API Key Handling

The API authenticates each request with an API key sent as a query parameter. Never commit your key; load it from the environment:

```go
apiKey := os.Getenv("BARIKOI_API_KEY")
```

## Installation

```bash
go get github.com/barikoi/barikoiapis-golang
```

```go
import "github.com/barikoi/barikoiapis-golang"
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	barikoi "github.com/barikoi/barikoiapis-golang"
)

func main() {
	c, err := barikoi.NewClient("YOUR_BARIKOI_API_KEY")
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

Full type documentation is on [pkg.go.dev](https://pkg.go.dev/github.com/barikoi/barikoiapis-golang). Per-method notes below.

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

Place suggestions for a partial query, with addresses in English and Bangla, coordinates, and place codes. Use for search boxes and location pickers.

```go
result, err := c.Autocomplete(ctx, &barikoi.AutocompleteRequest{
	Q:      "Dhaka",
	Bangla: barikoi.BoolPtr(true), // default true
})
```

### Reverse Geocoding

Convert coordinates to a human-readable address with optional administrative fields (district, division, thana, Bangla text, ...).

**Note:** every enabled flag may consume additional API credits — request only what you need.

```go
result, err := c.ReverseGeocode(ctx, &barikoi.ReverseGeocodeRequest{
	Latitude: 23.8103, Longitude: 90.4125,
	District: true, Bangla: true,
})
addr := result.Place.Address
```

### Nearby Places

Places within a radius of a point. Radius is kilometers, both values are path parameters and default to `0.5`/`10`.

```go
result, err := c.Nearby(ctx, &barikoi.NearbyRequest{
	Latitude: 23.87188719, Longitude: 90.38305163,
	Radius: 0.5, Limit: 10,
})
```

### Geocode (Rupantor)

Format and geocode an address string: returns the fixed address, the matched place, and a confidence score. One Rupantor request consumes two Geocode API credits.

```go
result, err := c.Geocode(ctx, &barikoi.GeocodeRequest{
	Q: "shewrapara mirpur dhaka", Thana: true,
})
place := result.GeocodedAddress
```

### Search Place

Search places by query; each hit carries a place code, and the response carries a session ID required by [Place Details](#place-details).

```go
result, err := c.SearchPlace(ctx, &barikoi.SearchPlaceRequest{Q: "barikoi"})
```

### Place Details

Address and coordinates for a place code, using the session ID from the same [Search Place](#search-place) response.

```go
result, err := c.PlaceDetails(ctx, &barikoi.PlaceDetailsRequest{
	PlaceCode: search.Places[0].PlaceCode,
	SessionID: search.SessionID,
})
```

### Route Overview

Basic route between two or more points: geometry (polyline by default, or GeoJSON), distance in meters, duration in seconds, waypoints.

```go
result, err := c.RouteOverview(ctx, &barikoi.RouteOverviewRequest{
	Coordinates: "90.4125,23.8103;90.3742,23.7461", // "lon,lat;lon,lat"
	Geometries:  "polyline", Profile: "car",
})
km := result.Routes[0].Distance / 1000
```

### Calculate Route

Detailed route with turn-by-turn instructions (GraphHopper format): distance in meters, time in milliseconds, encoded polyline or GeoJSON geometry.

```go
result, err := c.CalculateRoute(ctx, &barikoi.CalculateRouteRequest{
	Start:       barikoi.Coordinate{Latitude: 23.8103, Longitude: 90.4125},
	Destination: barikoi.Coordinate{Latitude: 23.7461, Longitude: 90.3742},
	Type: "gh", Profile: "car",
})
first := result.Paths[0].Instructions[0].Text
```

### Route Optimization

Route from source to destination through 1–50 waypoints, sorted by ascending `id`. Returns the same GraphHopper-format response as Calculate Route.

```go
result, err := c.OptimizeRoute(ctx, &barikoi.OptimizeRouteRequest{
	Source: "23.8103,90.4125", Destination: "23.7461,90.3742",
	GeoPoints: []barikoi.OptimizeRoutePoint{
		{ID: 1, Point: "23.7925,90.4078"},
		{ID: 2, Point: "23.7609,90.3805"},
	},
})
```

### Snap to Road

Nearest point on the road network to a coordinate. `Coordinates` is `[longitude, latitude]`; `Distance` is meters.

```go
result, err := c.SnapToRoad(ctx, &barikoi.SnapToRoadRequest{Point: "23.8065,90.3613"})
```

### Check Nearby

Whether the destination point is within 10–1000 meters of the current point. `Data` is nil when outside the radius.

```go
result, err := c.CheckNearby(ctx, &barikoi.CheckNearbyRequest{
	CurrentLatitude: 23.76241, CurrentLongitude: 90.37864,
	DestinationLatitude: 23.76245, DestinationLongitude: 90.37852,
	Radius: 50,
})
inside := result.Data != nil
```

---

### API Key Management

```go
c.SetAPIKey("new-api-key") // rotate at runtime
key := c.GetAPIKey()
```

### Timeout Configuration

```go
c, _ := barikoi.NewClient("YOUR_KEY", barikoi.WithTimeout(60*time.Second))
c.SetTimeout(90 * time.Second) // runtime
```

### Custom Base URL

```go
c, _ := barikoi.NewClient("YOUR_KEY",
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

The Makefile runs everything in Docker (`golang:1.22`) when no local Go
toolchain is installed. Copy `env.example` to `.env` for integration tests.
See [CONTRIBUTING.md](CONTRIBUTING.md) for architecture and workflow.

---

## Documentation

- [pkg.go.dev reference](https://pkg.go.dev/github.com/barikoi/barikoiapis-golang)
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
